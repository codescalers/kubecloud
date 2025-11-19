package workflows

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"kubecloud/internal/auth"
	"kubecloud/internal/billing"
	cfg "kubecloud/internal/config"
	"kubecloud/internal/core/generators"
	"kubecloud/internal/core/models"
	"kubecloud/internal/infrastructure/grid"
	"kubecloud/internal/infrastructure/kyc"
	mailservice "kubecloud/internal/infrastructure/mailservice"
	"kubecloud/internal/infrastructure/metrics"
	"net/http"
	"time"

	"slices"
	"strings"

	"kubecloud/internal/infrastructure/logger"

	"github.com/cosmos/go-bip39"
	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/deployer"
	"github.com/vedhavyas/go-subkey"
	"github.com/xmonader/ewf"
)

func FromUSDMilliCentToUSD(amountMillicent uint64) float64 {
	return float64(amountMillicent) / 1000
}

func FromUSDToUSDMillicent(amountUSD float64) uint64 {
	return uint64(amountUSD * 1000)
}

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
			return nil
		}

		user.ID = existingUser.ID
		if updateErr := userRepo.UpdateUserByID(&user); updateErr != nil {
			return fmt.Errorf("failed to update user: %w", updateErr)
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
		subject, body := mailService.SignUpMailContent(code, config.MailSender.TimeoutMin, name)

		if err := mailService.SendMailFromSystem(email, subject, body); err != nil {
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

// GenerateMnemonic generate mnemonic
func GenerateMnemonic() (string, error) {
	entropy, err := bip39.NewEntropy(128)
	if err != nil {
		return "", fmt.Errorf("failed to generate entropy: %w", err)
	}

	mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil {
		return "", fmt.Errorf("failed to generate mnemonic: %w", err)
	}

	if !bip39.IsMnemonicValid(mnemonic) {
		return "", fmt.Errorf("generated mnemonic is not valid")
	}

	return mnemonic, nil
}

// Activates user account with activation service
func activateAccount(substrateAccountID, network string) error {
	activationServiceURL := grid.ActivationServiceURLs[network]

	body := map[string]string{"substrateAccountID": substrateAccountID}
	jsonData, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal activation payload: %w", err)
	}

	resp, err := http.Post(activationServiceURL, "application/json", bytes.NewReader(jsonData))
	if err != nil {
		return fmt.Errorf("activation request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("activation failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}

// SetupUserOnTFChain performs all TFChain setup steps and returns mnemonic, identity, twin ID
func SetupUserOnTFChain(gridClient deployer.TFPluginClient, termsAndConditions cfg.TermsANDConditions, network string) (mnemonic string, twinID uint32, err error) {
	mnemonic, err = GenerateMnemonic()
	if err != nil {
		return "", 0, fmt.Errorf("generate mnemonic failed: %w", err)
	}

	identity, err := gridClient.SubstrateConn.NewIdentityFromSr25519Phrase(mnemonic)
	if err != nil {
		return "", 0, fmt.Errorf("identity creation failed: %w", err)
	}

	// Activate account with activation service
	if err := activateAccount(identity.Address(), network); err != nil {
		return "", 0, fmt.Errorf("activation failed: %w", err)
	}

	// Wait a few seconds for account activation to complete
	time.Sleep(7 * time.Second)

	if err := gridClient.SubstrateConn.AcceptTermsAndConditions(identity, termsAndConditions.DocumentLink, termsAndConditions.DocumentHash); err != nil {
		return "", 0, fmt.Errorf("accept terms failed: %w", err)
	}

	// Create Twin
	twinID, err = gridClient.SubstrateConn.CreateTwin(identity, "", []byte{})
	if err != nil {
		return "", 0, fmt.Errorf("create twin failed: %w", err)
	}

	log := logger.ForOperation("chain_account", "create_twin")
	log.Debug().
		Uint32("twin_id", twinID).
		Str("address", identity.Address()).
		Msg("Twin created successfully")
	return mnemonic, twinID, nil
}

func SetupTFChainStep(gridClient deployer.TFPluginClient, userRepo models.UserRepository, config cfg.Configuration) ewf.StepFn {
	return func(ctx context.Context, state ewf.State) error {
		userIDVal, ok := state["user_id"]
		if !ok {
			return fmt.Errorf("missing 'user_id' in state")
		}
		userID, ok := userIDVal.(int)
		if !ok {
			return fmt.Errorf("'user_id' in state is not an int")
		}

		existingUser, err := userRepo.GetUserByID(userID)
		if err != nil {
			return fmt.Errorf("failed to check existing user: %w", err)
		}

		if len(strings.TrimSpace(existingUser.Mnemonic)) > 0 {
			state["mnemonic"] = existingUser.Mnemonic
			return nil
		}

		mnemonic, _, err := SetupUserOnTFChain(gridClient, config.TermsANDConditions, config.SystemAccount.Network)
		if err != nil {
			return err
		}

		if err := userRepo.UpdateUserByID(&models.User{
			ID:       userID,
			Mnemonic: mnemonic,
		}); err != nil {
			return fmt.Errorf("failed to update user mnemonic: %w", err)
		}

		state["mnemonic"] = mnemonic
		return nil
	}
}

func CreateStripeCustomerStep(userRepo models.UserRepository, stripeClient billing.StripeClient) ewf.StepFn {
	return func(ctx context.Context, state ewf.State) error {
		userIDVal, ok := state["user_id"]
		if !ok {
			return fmt.Errorf("missing 'user_id' in state")
		}
		userID, ok := userIDVal.(int)
		if !ok {
			return fmt.Errorf("'user_id' in state is not an int")
		}

		existingUser, err := userRepo.GetUserByID(userID)
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
			ID:               userID,
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
		userIDVal, ok := state["user_id"]
		if !ok {
			return fmt.Errorf("missing 'user_id' in state")
		}
		userID, ok := userIDVal.(int)
		if !ok {
			return fmt.Errorf("'user_id' in state is not an int")
		}

		existingUser, err := userRepo.GetUserByID(userID)
		if err != nil {
			return fmt.Errorf("failed to check existing user: %w", err)
		}

		if existingUser.Sponsored && len(strings.TrimSpace(existingUser.AccountAddress)) > 0 {
			return nil
		}

		mnemonicVal, ok := state["mnemonic"]
		if !ok {
			return fmt.Errorf("missing 'mnemonic' in state")
		}
		mnemonic, ok := mnemonicVal.(string)
		if !ok {
			return fmt.Errorf("'mnemonic' in state is not a string")
		}

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
			ID:             userID,
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

		subject, body := mailService.WelcomeMailContent(name)
		if err := mailService.SendMailFromSystem(email, subject, body); err != nil {
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

func CreatePendingRecord(gridClient deployer.TFPluginClient, pendingRecordRepo models.PendingRecordRepository) ewf.StepFn {
	return func(ctx context.Context, state ewf.State) error {
		log := logger.ForOperation("user_activities", "create_pending_record")
		amountVal, ok := state["amount"]
		if !ok {
			return fmt.Errorf("missing 'amount' in state")
		}

		amount, ok := amountVal.(uint64)
		if !ok {
			return fmt.Errorf("'amount' in state is not a uint64")
		}

		userIDVal, ok := state["user_id"]
		if !ok {
			return fmt.Errorf("missing 'user_id' in state")
		}
		userID, ok := userIDVal.(int)
		if !ok {
			return fmt.Errorf("'user_id' in state is not an int")
		}

		usernameVal, ok := state["username"]
		if !ok {
			return fmt.Errorf("missing 'username' in state")
		}
		username, ok := usernameVal.(string)
		if !ok {
			return fmt.Errorf("'username' in state is not a string")
		}

		transferModeVal, ok := state["transfer_mode"]
		if !ok {
			return fmt.Errorf("missing 'transfer_mode' in state")
		}
		transferMode, ok := transferModeVal.(string)
		if !ok {
			return fmt.Errorf("'transfer_mode' in state is not a string")
		}

		requestedTFTs, err := gridClient.SubstrateConn.FromUSDMillicentToTFT(amount)
		if err != nil {
			log.Error().Err(err).Msg("error converting USD to TFT")
			return err
		}

		if err = pendingRecordRepo.CreatePendingRecord(&models.PendingRecord{
			UserID:       userID,
			Username:     username,
			TFTAmount:    requestedTFTs,
			TransferMode: transferMode,
		}); err != nil {
			log.Error().Err(err).Msg("failed to create pending record")
			return err
		}

		return nil
	}
}

func UpdateCreditCardBalanceStep(userRepo models.UserRepository) ewf.StepFn {
	return func(ctx context.Context, state ewf.State) error {
		userIDVal, ok := state["user_id"]
		if !ok {
			return fmt.Errorf("missing 'user_id' in state")
		}
		userID, ok := userIDVal.(int)
		if !ok {
			return fmt.Errorf("'user_id' in state is not an int")
		}

		amountVal, ok := state["amount"]
		if !ok {
			return fmt.Errorf("missing 'amount' in state")
		}
		amount, ok := amountVal.(uint64)
		if !ok {
			return fmt.Errorf("'amount' in state is not a uint64")
		}

		user, err := userRepo.GetUserByID(userID)
		if err != nil {
			return fmt.Errorf("user not found: %w", err)
		}

		user.CreditCardBalance += amount
		if err := userRepo.UpdateUserByID(&user); err != nil {
			return fmt.Errorf("error updating user: %w", err)
		}

		state["mnemonic"] = user.Mnemonic
		netBalance := int64(user.CreditCardBalance) + int64(user.CreditedBalance) - int64(user.Debt)
		if netBalance < 0 {
			netBalance = 0
		}
		state["net_balance"] = uint64(netBalance)

		return nil
	}
}

// DrainUserBalanceStep transfers a user's balance to the system account
func DrainUserBalanceStep(userRepo models.UserRepository, gridClient deployer.TFPluginClient, systemMnemonic string) ewf.StepFn {
	return func(ctx context.Context, state ewf.State) error {
		userIDVal, ok := state["user_id"]
		if !ok {
			return fmt.Errorf("missing 'user_id' in state")
		}
		userID, ok := userIDVal.(int)
		if !ok {
			return fmt.Errorf("'user_id' in state is not an int")
		}

		user, err := userRepo.GetUserByID(userID)
		if err != nil {
			return fmt.Errorf("failed to get user: %w", err)
		}

		userIdentity, err := gridClient.SubstrateConn.NewIdentityFromSr25519Phrase(user.Mnemonic)
		if err != nil {
			return fmt.Errorf("failed to get user Identity from mnemonic %v", err)
		}

		// Get user's current balance in TFT from on-chain
		balance, err := gridClient.SubstrateConn.GetBalance(userIdentity)
		if err != nil {
			return fmt.Errorf("failed to get user balance: %w", err)
		}

		balanceInTFT := balance.Free.Uint64()

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
		err = gridClient.SubstrateConn.TransferTFTsToSystem(transferAmount, user.Mnemonic, systemMnemonic)
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
func UpdateCreditedBalanceStep(userRepo models.UserRepository) ewf.StepFn {
	return func(ctx context.Context, state ewf.State) error {
		userIDVal, ok := state["user_id"]
		if !ok {
			return fmt.Errorf("missing 'user_id' in state")
		}
		userID, ok := userIDVal.(int)
		if !ok {
			return fmt.Errorf("'user_id' in state is not an int")
		}

		amountVal, ok := state["amount"]
		if !ok {
			return fmt.Errorf("missing 'amount' in state")
		}
		amount, ok := amountVal.(uint64)
		if !ok {
			return fmt.Errorf("'amount' in state is not a uint64")
		}

		user, err := userRepo.GetUserByID(userID)
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
