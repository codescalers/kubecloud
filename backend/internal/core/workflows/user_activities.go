package workflows

import (
	"context"
	"fmt"
	"kubecloud/internal/auth"
	"kubecloud/internal/billing"
	cfg "kubecloud/internal/config"
	"kubecloud/internal/core/generators"
	"kubecloud/internal/core/models"
	"kubecloud/internal/infrastructure/gridclient"
	"kubecloud/internal/infrastructure/kyc"
	"kubecloud/internal/infrastructure/mailservice"
	"kubecloud/internal/infrastructure/metrics"
	"sync"

	"slices"
	"strings"

	"kubecloud/internal/infrastructure/logger"

	"github.com/hashicorp/go-multierror"
	"github.com/vedhavyas/go-subkey"
	"github.com/xmonader/ewf"
)

func CreateUserStep(config cfg.Configuration, userRepo models.UserRepository) ewf.StepFn {
	return func(ctx context.Context, state ewf.State) error {
		emailVal, ok := state["email"]
		if !ok {
			return fmt.Errorf("missing 'email' in state")
		}
		email, ok := emailVal.(string)
		if !ok {
			return fmt.Errorf("'email' in state is not a string")
		}

		nameVal, ok := state["name"]
		if !ok {
			return fmt.Errorf("missing 'name' in state")
		}
		name, ok := nameVal.(string)
		if !ok {
			return fmt.Errorf("'name' in state is not a string")
		}

		passwordVal, ok := state["password"]
		if !ok {
			return fmt.Errorf("missing 'password' in state")
		}
		password, ok := passwordVal.(string)
		if !ok {
			return fmt.Errorf("'password' in state is not a string")
		}

		hashedPassword, err := auth.HashAndSaltPassword([]byte(password))
		if err != nil {
			return fmt.Errorf("hashing password failed: %w", err)
		}

		user := models.User{
			Username: name,
			Email:    email,
			Password: hashedPassword,
			Admin:    slices.Contains(config.Admins, email),
		}

		existingUser, err := userRepo.GetUserByEmail(email)
		if err != nil && err != models.ErrUserNotFound {
			return fmt.Errorf("failed to check existing user: %w", err)
		}

		if err == models.ErrUserNotFound {
			if err = userRepo.RegisterUser(&user); err != nil {
				return fmt.Errorf("user registration failed: %w", err)
			}
			// Store user_id in state for later use (e.g., in hooks/notifications)
			state["config"] = map[string]interface{}{
				"user_id": user.ID,
			}
			return nil
		}

		user.ID = existingUser.ID
		if updateErr := userRepo.UpdateUserByID(&user); updateErr != nil {
			return fmt.Errorf("failed to update user: %w", updateErr)
		}
		state["config"] = map[string]interface{}{
			"user_id": user.ID,
		}
		return nil
	}
}

func SendVerificationEmailStep(mailService mailservice.MailService, config cfg.Configuration) ewf.StepFn {
	return func(ctx context.Context, state ewf.State) error {
		emailVal, ok := state["email"]
		if !ok {
			return fmt.Errorf("missing 'email' in state")
		}
		email, ok := emailVal.(string)
		if !ok {
			return fmt.Errorf("'email' in state is not a string")
		}

		nameVal, ok := state["name"]
		if !ok {
			return fmt.Errorf("missing 'name' in state")
		}
		name, ok := nameVal.(string)
		if !ok {
			return fmt.Errorf("'name' in state is not a string")
		}

		code := generators.GenerateVerificationCode(config.VerificationCodeLength)
		err := mailService.SendSignUpMail(email, code, name)
		if err != nil {
			return fmt.Errorf("send mail failed: %w", err)
		}

		state["code"] = code
		return nil
	}
}

func UpdateCodeStep(userRepo models.UserRepository) ewf.StepFn {
	return func(ctx context.Context, state ewf.State) error {
		emailVal, ok := state["email"]
		if !ok {
			return fmt.Errorf("missing 'email' in state")
		}
		email, ok := emailVal.(string)
		if !ok {
			return fmt.Errorf("'email' in state is not a string")
		}

		codeVal, ok := state["code"]
		if !ok {
			return fmt.Errorf("missing 'code' in state")
		}
		code, ok := codeVal.(int)
		if !ok {
			return fmt.Errorf("'code' in state is not a int")
		}

		existingUser, err := userRepo.GetUserByEmail(email)
		if err != nil && err != models.ErrUserNotFound {
			return fmt.Errorf("failed to check existing user: %w", err)
		}

		existingUser.Code = code
		return userRepo.UpdateUserByID(&existingUser)
	}
}

func SetupTFChainStep(gridClient gridclient.GridClient, userRepo models.UserRepository, termsAndConditions cfg.TermsANDConditions) ewf.StepFn {
	return func(ctx context.Context, state ewf.State) error {
		userConfig, err := getConfig(state)
		if err != nil {
			return fmt.Errorf("failed to get config from state: %w", err)
		}

		existingUser, err := userRepo.GetUserByID(userConfig.UserID)
		if err != nil {
			return fmt.Errorf("failed to check existing user: %w", err)
		}

		if len(strings.TrimSpace(existingUser.Mnemonic)) > 0 {
			// Update config with mnemonic
			state["config"] = map[string]interface{}{
				"user_id":  userConfig.UserID,
				"mnemonic": existingUser.Mnemonic,
			}
			return nil
		}

		mnemonic, _, err := gridClient.SetupUserOnTFChain(termsAndConditions)
		if err != nil {
			return err
		}

		if err := userRepo.UpdateUserByID(&models.User{
			ID:       userConfig.UserID,
			Mnemonic: mnemonic,
		}); err != nil {
			return fmt.Errorf("failed to update user mnemonic: %w", err)
		}

		// Update config with mnemonic
		state["config"] = map[string]interface{}{
			"user_id":  userConfig.UserID,
			"mnemonic": mnemonic,
		}
		return nil
	}
}

func CreateStripeCustomerStep(userRepo models.UserRepository, stripeClient billing.StripeClient) ewf.StepFn {
	return func(ctx context.Context, state ewf.State) error {
		userConfig, err := getConfig(state)
		if err != nil {
			return fmt.Errorf("failed to get config from state: %w", err)
		}

		existingUser, err := userRepo.GetUserByID(userConfig.UserID)
		if err != nil {
			return fmt.Errorf("failed to check existing user: %w", err)
		}

		if len(strings.TrimSpace(existingUser.StripeCustomerID)) > 0 {
			return nil
		}

		emailVal, ok := state["email"]
		if !ok {
			return fmt.Errorf("missing 'email' in state")
		}
		email, ok := emailVal.(string)
		if !ok {
			return fmt.Errorf("'email' in state is not a string")
		}

		nameVal, ok := state["name"]
		if !ok {
			return fmt.Errorf("missing 'name' in state")
		}
		name, ok := nameVal.(string)
		if !ok {
			return fmt.Errorf("'name' in state is not a string")
		}

		customer, err := stripeClient.CreateCustomer(name, email)
		if err != nil {
			return err
		}

		if err := userRepo.UpdateUserByID(&models.User{
			ID:               userConfig.UserID,
			StripeCustomerID: customer.ID,
		}); err != nil {
			return fmt.Errorf("failed to update user stripe customer: %w", err)
		}

		return nil
	}
}

func CreateKYCSponsorship(kycClient *kyc.KYCClient, sponsorAddress string, sponsorKeyPair subkey.KeyPair, userRepo models.UserRepository) ewf.StepFn {
	return func(ctx context.Context, state ewf.State) error {
		log := logger.ForOperation("user_activities", "create_kyc_sponsorship")
		userConfig, err := getConfig(state)
		if err != nil {
			return fmt.Errorf("failed to get config from state: %w", err)
		}

		existingUser, err := userRepo.GetUserByID(userConfig.UserID)
		if err != nil {
			return fmt.Errorf("failed to check existing user: %w", err)
		}

		if existingUser.Sponsored && len(strings.TrimSpace(existingUser.AccountAddress)) > 0 {
			return nil
		}

		// Get mnemonic from config
		if userConfig.Mnemonic == "" {
			return fmt.Errorf("missing 'mnemonic' in config")
		}
		mnemonic := userConfig.Mnemonic

		// Set user.AccountAddress from mnemonic
		sponseeKeyPair, err := auth.KeyPairFromMnemonic(mnemonic)
		if err != nil {
			log.Error().Err(err).Msg("failed to create keypair for SS58 address")
			return err
		}

		sponseeAddress, err := auth.AccountAddressFromKeypair(sponseeKeyPair)
		if err != nil {
			log.Error().Err(err).Msg("failed to get SS58 address")
			return err
		}

		if err := kycClient.CreateSponsorship(ctx, sponsorAddress, sponsorKeyPair, sponseeAddress, sponseeKeyPair); err != nil {
			return fmt.Errorf("failed to create KYC sponsorship: %w", err)
		}

		if err := userRepo.UpdateUserByID(&models.User{
			ID:             userConfig.UserID,
			Sponsored:      true,
			AccountAddress: sponseeAddress,
		}); err != nil {
			return fmt.Errorf("failed to update user data: %w", err)
		}

		return nil
	}
}

func SendWelcomeEmailStep(mailService mailservice.MailService, metrics *metrics.Metrics) ewf.StepFn {
	return func(ctx context.Context, state ewf.State) error {
		metrics.IncrementUserRegistration()

		emailVal, ok := state["email"]
		if !ok {
			return fmt.Errorf("missing 'email' in state")
		}
		email, ok := emailVal.(string)
		if !ok {
			return fmt.Errorf("'email' in state is not a string")
		}

		nameVal, ok := state["name"]
		if !ok {
			return fmt.Errorf("missing 'name' in state")
		}
		name, ok := nameVal.(string)
		if !ok {
			return fmt.Errorf("'name' in state is not a string")
		}

		err := mailService.SendWelcomeEmail(email, name)
		if err != nil {
			return fmt.Errorf("send mail failed: %w", err)
		}
		return nil
	}
}

func CreatePaymentIntentStep(currency string, metrics *metrics.Metrics, stripeClient billing.StripeClient) ewf.StepFn {
	return func(ctx context.Context, state ewf.State) error {
		customerIDVal, ok := state["stripe_customer_id"]
		if !ok {
			return fmt.Errorf("missing 'stripe_customer_id' in state")
		}
		customerID, ok := customerIDVal.(string)
		if !ok {
			return fmt.Errorf("'stripe_customer_id' in state is not a string")
		}
		paymentMethodIDVal, ok := state["payment_method_id"]
		if !ok {
			return fmt.Errorf("missing 'payment_method_id' in state")
		}
		paymentMethodID, ok := paymentMethodIDVal.(string)
		if !ok {
			return fmt.Errorf("'payment_method_id' in state is not a string")
		}
		amountVal, ok := state["amount"]
		if !ok {
			return fmt.Errorf("missing 'amount' in state")
		}
		amount, ok := amountVal.(uint64)
		if !ok {
			return fmt.Errorf("'amount' in state is not a uint64")
		}

		intent, err := stripeClient.CreatePaymentIntent(customerID, paymentMethodID, currency, amount)
		if err != nil {
			metrics.IncrementStripePaymentFailure()
			return fmt.Errorf("error creating payment intent: %w", err)
		}

		metrics.IncrementStripePaymentSuccess()
		state["payment_intent_id"] = intent.ID
		return nil
	}
}

func UpdateCreditCardBalanceStep(userRepo models.UserRepository) ewf.StepFn {
	return func(ctx context.Context, state ewf.State) error {
		config, err := getConfig(state)
		if err != nil {
			return fmt.Errorf("failed to get config from state: %w", err)
		}

		amountVal, ok := state["amount"]
		if !ok {
			return fmt.Errorf("missing 'amount' in state")
		}
		amount, ok := amountVal.(uint64)
		if !ok {
			return fmt.Errorf("'amount' in state is not a uint64")
		}

		user, err := userRepo.GetUserByID(config.UserID)
		if err != nil {
			return fmt.Errorf("user not found: %w", err)
		}

		user.CreditCardBalance += amount
		if err := userRepo.UpdateUserByID(&user); err != nil {
			return fmt.Errorf("error updating user: %w", err)
		}

		// Update config with mnemonic
		state["config"] = map[string]interface{}{
			"user_id":  config.UserID,
			"mnemonic": user.Mnemonic,
		}
		netBalance := int64(user.CreditCardBalance) + int64(user.CreditedBalance) - int64(user.Debt)
		if netBalance < 0 {
			netBalance = 0
		}
		state["net_balance"] = uint64(netBalance)

		return nil
	}
}

// DrainUserBalanceStep transfers a user's balance to the system account
func DrainUserBalanceStep(userRepo models.UserRepository, gridClient gridclient.GridClient) ewf.StepFn {
	return func(ctx context.Context, state ewf.State) error {
		userIDVal, ok := state["target_user_id"]
		if !ok {
			return fmt.Errorf("missing 'target_user_id' in state")
		}
		userID, ok := userIDVal.(int)
		if !ok {
			return fmt.Errorf("'target_user_id' in state is not an int")
		}

		user, err := userRepo.GetUserByID(userID)
		if err != nil {
			return fmt.Errorf("failed to get user: %w", err)
		}

		state["target_username"] = user.Username
		balanceInTFT, err := gridClient.GetFreeBalanceTFT(user.Mnemonic)
		if err != nil {
			return fmt.Errorf("failed to get user balance: %w", err)
		}

		// Minimum balance threshold to keep (0.00001 TFT)
		const minBalanceThreshold uint64 = 1e5

		if balanceInTFT <= minBalanceThreshold {
			logger.GetLogger().Info().
				Int("user_id", userID).
				Uint64("balance", balanceInTFT).
				Msg("user balance below minimum threshold, nothing to drain")
			return nil
		}

		// Transfer the balance minus threshold
		transferAmount := balanceInTFT - minBalanceThreshold

		// Perform the transfer from user to system account
		err = gridClient.TransferTFTsToSystem(transferAmount, user.Mnemonic)
		if err != nil {
			return fmt.Errorf("failed to transfer balance: %w", err)
		}

		logger.GetLogger().Info().
			Int("user_id", userID).
			Uint64("amount_tft", transferAmount).
			Uint64("remaining_balance_tft", minBalanceThreshold).
			Msg("successfully drained user balance to system account")

		state["drained_amount_tft"] = transferAmount

		return nil
	}
}

func DrainAllUsersBalanceStep(userRepo models.UserRepository, ewfEngine *ewf.Engine, maxConcurrent int) ewf.StepFn {
	return func(ctx context.Context, state ewf.State) error {
		users, err := userRepo.ListAllUsers()
		if err != nil {
			return fmt.Errorf("failed to get all users: %w", err)
		}
		multiErr := &multierror.Error{}
		wg := sync.WaitGroup{}
		mu := sync.Mutex{}
		sem := make(chan struct{}, maxConcurrent)

		for _, user := range users {
			wg.Add(1)
			sem <- struct{}{}
			go func(user models.User) {
				defer wg.Done()
				defer func() { <-sem }()
				drainDisplayName := fmt.Sprintf("Drain %s balance", user.Username)
				wf, err := ewfEngine.NewWorkflow(WorkflowDrainUser, ewf.WithDisplayName(drainDisplayName))
				if err != nil {
					mu.Lock()
					multiErr = multierror.Append(multiErr, err)
					mu.Unlock()
					return
				}

				wf.State = map[string]interface{}{
					"target_user_id":        user.ID,
					"target_username":       user.Username,
					"suppress_notification": true,
				}

				if err = ewfEngine.Run(ctx, wf); err != nil {
					mu.Lock()
					multiErr = multierror.Append(multiErr, err)
					mu.Unlock()
				}
			}(user)
		}
		wg.Wait()
		if err := multiErr.ErrorOrNil(); err != nil {
			return fmt.Errorf("failed to drain all users balance: %w", err)
		}
		return nil
	}
}

func UpdateCreditedBalanceStep(userRepo models.UserRepository) ewf.StepFn {
	return func(ctx context.Context, state ewf.State) error {
		config, err := getConfig(state)
		if err != nil {
			return fmt.Errorf("failed to get config from state: %w", err)
		}

		amountVal, ok := state["amount"]
		if !ok {
			return fmt.Errorf("missing 'amount' in state")
		}
		amount, ok := amountVal.(uint64)
		if !ok {
			return fmt.Errorf("'amount' in state is not a uint64")
		}

		user, err := userRepo.GetUserByID(config.UserID)
		if err != nil {
			return fmt.Errorf("user is not found: %w", err)
		}

		user.CreditedBalance += amount
		if err := userRepo.UpdateUserByID(&user); err != nil {
			return fmt.Errorf("error updating user: %w", err)
		}

		netBalance := int64(user.CreditCardBalance) + int64(user.CreditedBalance) - int64(user.Debt)
		if netBalance < 0 {
			netBalance = 0
		}
		state["net_balance"] = uint64(netBalance)

		return nil
	}
}
