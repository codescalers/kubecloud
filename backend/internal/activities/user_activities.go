package activities

import (
	"context"
	"fmt"
	"kubecloud/internal"
	"kubecloud/internal/metrics"
	"kubecloud/internal/notification"
	"kubecloud/models"
	"strings"

	"kubecloud/internal/logger"

	substrate "github.com/threefoldtech/tfchain/clients/tfchain-client-go"
	"github.com/vedhavyas/go-subkey"
	"github.com/xmonader/ewf"
	"gorm.io/gorm"
)

func CreateUserStep(config internal.Configuration, db models.DB) ewf.StepFn {
	return func(ctx context.Context, state ewf.State) error {
		email, err := getFromState[string](state, "email")
		if err != nil {
			return err
		}

		name, err := getFromState[string](state, "name")
		if err != nil {
			return err
		}

		password, err := getFromState[string](state, "password")
		if err != nil {
			return err
		}

		hashedPassword, err := internal.HashAndSaltPassword([]byte(password))
		if err != nil {
			return fmt.Errorf("hashing password failed: %w", err)
		}

		user := models.User{
			Username: name,
			Email:    email,
			Password: hashedPassword,
			Admin:    internal.Contains(config.Admins, email),
		}

		existingUser, err := db.GetUserByEmail(email)
		if err != nil && err != gorm.ErrRecordNotFound {
			return fmt.Errorf("failed to check existing user: %w", err)
		}

		if err == gorm.ErrRecordNotFound {
			if err = db.RegisterUser(&user); err != nil {
				return fmt.Errorf("user registration failed: %w", err)
			}
			return nil
		}

		user.ID = existingUser.ID
		if updateErr := db.UpdateUserByID(&user); updateErr != nil {
			return fmt.Errorf("failed to update user: %w", updateErr)
		}

		return nil
	}
}

func SendVerificationEmailStep(mailService internal.MailService, config internal.Configuration) ewf.StepFn {
	return func(ctx context.Context, state ewf.State) error {
		email, err := getFromState[string](state, "email")
		if err != nil {
			return err
		}

		name, err := getFromState[string](state, "name")
		if err != nil {
			return err
		}

		code := internal.GenerateRandomCode()
		subject, body := mailService.SignUpMailContent(code, config.MailSender.TimeoutMin, name, config.Server.Host)

		if err := mailService.SendMail(config.MailSender.Email, email, subject, body); err != nil {
			return fmt.Errorf("send mail failed: %w", err)
		}

		state["code"] = code
		return nil
	}
}

func UpdateCodeStep(db models.DB) ewf.StepFn {
	return func(ctx context.Context, state ewf.State) error {
		email, err := getFromState[string](state, "email")
		if err != nil {
			return err
		}

		code, err := getFromState[int](state, "code")
		if err != nil {
			return err
		}

		existingUser, err := db.GetUserByEmail(email)
		if err != nil && err != gorm.ErrRecordNotFound {
			return fmt.Errorf("failed to check existing user: %w", err)
		}

		existingUser.Code = code
		return db.UpdateUserByID(&existingUser)
	}
}

func SetupTFChainStep(client *substrate.Substrate, config internal.Configuration, notificationService *notification.NotificationService, db models.DB) ewf.StepFn {
	return func(ctx context.Context, state ewf.State) error {
		userID, err := getFromState[int](state, "user_id")
		if err != nil {
			return err
		}

		existingUser, err := db.GetUserByID(userID)
		if err != nil {
			return fmt.Errorf("failed to check existing user: %w", err)
		}

		if len(strings.TrimSpace(existingUser.Mnemonic)) > 0 {
			state["mnemonic"] = existingUser.Mnemonic
			return nil
		}

		mnemonic, _, err := internal.SetupUserOnTFChain(client, config)
		if err != nil {
			return err
		}

		if err := db.UpdateUserByID(&models.User{
			ID:       userID,
			Mnemonic: mnemonic,
		}); err != nil {
			return fmt.Errorf("failed to update user mnemonic: %w", err)
		}

		state["mnemonic"] = mnemonic
		return nil
	}
}

func CreateStripeCustomerStep(db models.DB) ewf.StepFn {
	return func(ctx context.Context, state ewf.State) error {
		userID, err := getFromState[int](state, "user_id")
		if err != nil {
			return err
		}

		existingUser, err := db.GetUserByID(userID)
		if err != nil {
			return fmt.Errorf("failed to check existing user: %w", err)
		}

		if len(strings.TrimSpace(existingUser.StripeCustomerID)) > 0 {
			return nil
		}

		email, err := getFromState[string](state, "email")
		if err != nil {
			return err
		}

		name, err := getFromState[string](state, "name")
		if err != nil {
			return err
		}

		customer, err := internal.CreateStripeCustomer(name, email)
		if err != nil {
			return err
		}

		if err := db.UpdateUserByID(&models.User{
			ID:               userID,
			StripeCustomerID: customer.ID,
		}); err != nil {
			return fmt.Errorf("failed to update user stripe customer: %w", err)
		}

		return nil
	}
}

func CreateKYCSponsorship(kycClient *internal.KYCClient, notificationService *notification.NotificationService, sponsorAddress string, sponsorKeyPair subkey.KeyPair, db models.DB) ewf.StepFn {
	return func(ctx context.Context, state ewf.State) error {
		userID, err := getFromState[int](state, "user_id")
		if err != nil {
			return err
		}

		existingUser, err := db.GetUserByID(userID)
		if err != nil {
			return fmt.Errorf("failed to check existing user: %w", err)
		}

		if existingUser.Sponsored && len(strings.TrimSpace(existingUser.AccountAddress)) > 0 {
			return nil
		}

		mnemonic, err := getFromState[string](state, "mnemonic")
		if err != nil {
			return err
		}

		// Set user.AccountAddress from mnemonic
		sponseeKeyPair, err := internal.KeyPairFromMnemonic(mnemonic)
		if err != nil {
			logger.GetLogger().Error().Err(err).Msg("failed to create keypair for SS58 address")
			return err
		}

		sponseeAddress, err := internal.AccountAddressFromKeypair(sponseeKeyPair)
		if err != nil {
			logger.GetLogger().Error().Err(err).Msg("failed to get SS58 address")
			return err
		}

		if err := kycClient.CreateSponsorship(ctx, sponsorAddress, sponsorKeyPair, sponseeAddress, sponseeKeyPair); err != nil {
			return fmt.Errorf("failed to create KYC sponsorship: %w", err)
		}

		if err := db.UpdateUserByID(&models.User{
			ID:             userID,
			Sponsored:      true,
			AccountAddress: sponseeAddress,
		}); err != nil {
			return fmt.Errorf("failed to update user data: %w", err)
		}

		return nil
	}
}

func SendWelcomeEmailStep(mailService internal.MailService, config internal.Configuration, metrics *metrics.Metrics) ewf.StepFn {
	return func(ctx context.Context, state ewf.State) error {
		metrics.IncrementUserRegistration()

		email, err := getFromState[string](state, "email")
		if err != nil {
			return err
		}

		name, err := getFromState[string](state, "name")
		if err != nil {
			return err
		}

		subject, body := mailService.WelcomeMailContent(name, config.Server.Host)
		if err := mailService.SendMail(config.MailSender.Email, email, subject, body); err != nil {
			return fmt.Errorf("send mail failed: %w", err)
		}
		return nil
	}
}

func CreatePaymentIntentStep(currency string, metrics *metrics.Metrics, notificationService *notification.NotificationService) ewf.StepFn {
	return func(ctx context.Context, state ewf.State) error {
		customerID, err := getFromState[string](state, "stripe_customer_id")
		if err != nil {
			return err
		}

		paymentMethodID, err := getFromState[string](state, "payment_method_id")
		if err != nil {
			return err
		}

		amount, err := getFromState[uint64](state, "amount")
		if err != nil {
			return err
		}

		intent, err := internal.CreatePaymentIntent(customerID, paymentMethodID, currency, amount)
		if err != nil {
			metrics.IncrementStripePaymentFailure()
			return fmt.Errorf("error creating payment intent: %w", err)
		}

		metrics.IncrementStripePaymentSuccess()
		state["payment_intent_id"] = intent.ID
		return nil
	}
}

func CreatePendingRecord(substrateClient *substrate.Substrate, db models.DB, systemMnemonic string) ewf.StepFn {
	return func(ctx context.Context, state ewf.State) error {
		amount, err := getFromState[uint64](state, "amount")
		if err != nil {
			return err
		}

		userID, err := getFromState[int](state, "user_id")
		if err != nil {
			return err
		}

		username, err := getFromState[string](state, "username")
		if err != nil {
			return err
		}

		transferMode, err := getFromState[string](state, "transfer_mode")
		if err != nil {
			return err
		}

		requestedTFTs, err := internal.FromUSDMillicentToTFT(substrateClient, amount)
		if err != nil {
			logger.GetLogger().Error().Err(err).Msg("error converting usd")
			return err
		}

		if err = db.CreatePendingRecord(&models.PendingRecord{
			UserID:       userID,
			Username:     username,
			TFTAmount:    requestedTFTs,
			TransferMode: transferMode,
		}); err != nil {
			logger.GetLogger().Error().Err(err).Send()
			return err
		}

		return nil
	}
}

func UpdateCreditCardBalanceStep(db models.DB) ewf.StepFn {
	return func(ctx context.Context, state ewf.State) error {
		userID, err := getFromState[int](state, "user_id")
		if err != nil {
			return err
		}

		amount, err := getFromState[uint64](state, "amount")
		if err != nil {
			return err
		}

		user, err := db.GetUserByID(userID)
		if err != nil {
			return fmt.Errorf("user not found: %w", err)
		}

		user.CreditCardBalance += amount
		if err := db.UpdateUserByID(&user); err != nil {
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

func UpdateCreditedBalanceStep(db models.DB) ewf.StepFn {
	return func(ctx context.Context, state ewf.State) error {
		userID, err := getFromState[int](state, "user_id")
		if err != nil {
			return err
		}

		amount, err := getFromState[uint64](state, "amount")
		if err != nil {
			return err
		}

		user, err := db.GetUserByID(userID)
		if err != nil {
			return fmt.Errorf("user is not found: %w", err)
		}

		user.CreditedBalance += amount
		if err := db.UpdateUserByID(&user); err != nil {
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
