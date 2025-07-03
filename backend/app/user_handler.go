package app

import (
	"fmt"
	"kubecloud/internal"
	"kubecloud/models"
	"net/http"
	"strconv"
	"time"

	substrate "github.com/threefoldtech/tfchain/clients/tfchain-client-go"
	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/graphql"
	proxy "github.com/threefoldtech/tfgrid-sdk-go/grid-proxy/pkg/client"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/paymentmethod"
	"gorm.io/gorm"
)

// Handler struct holds configs for all handlers
type Handler struct {
	tokenManager    internal.TokenManager
	db              models.DB
	config          internal.Configuration
	mailService     internal.MailService
	proxyClient     proxy.Client
	substrateClient *substrate.Substrate
	graphqlClient   graphql.GraphQl
	firesquidClient graphql.GraphQl
	redis           *internal.RedisClient
	sseManager      *internal.SSEManager
}

// NewHandler create new handler
func NewHandler(tokenManager internal.TokenManager, db models.DB,
	config internal.Configuration, mailService internal.MailService,
	gridproxy proxy.Client, substrateClient *substrate.Substrate,
	graphqlClient graphql.GraphQl, firesquidClient graphql.GraphQl,
	redis *internal.RedisClient, sseManager *internal.SSEManager) *Handler {
	return &Handler{
		tokenManager:    tokenManager,
		db:              db,
		config:          config,
		mailService:     mailService,
		proxyClient:     gridproxy,
		substrateClient: substrateClient,
		graphqlClient:   graphqlClient,
		firesquidClient: firesquidClient,
		redis:           redis,
		sseManager:      sseManager,
	}
}

// RegisterInput struct for data needed when user register
type RegisterInput struct {
	Name            string `json:"name" binding:"required" validate:"min=3,max=64"`
	Email           string `json:"email" binding:"required,email"`
	Password        string `json:"password" binding:"required,min=8,max=64"`
	ConfirmPassword string `json:"confirm_password" binding:"required,eqfield=Password"`
}

// LoginInput struct for login handler
type LoginInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=3,max=64"`
}

// RefreshTokenInput struct when user refresh token
type RefreshTokenInput struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// EmailInput struct for user when forgetting password
type EmailInput struct {
	Email string `json:"email" binding:"required,email"`
}

// VerifyCodeInput struct takes verification code from user
type VerifyCodeInput struct {
	Email string `json:"email" binding:"required,email"`
	Code  int    `json:"code" binding:"required,numeric"`
}

// ChangePasswordInput struct for user to change password
type ChangePasswordInput struct {
	Email           string `json:"email" binding:"required,email"`
	Password        string `json:"password" binding:"required,min=8,max=64"`
	ConfirmPassword string `json:"confirm_password" binding:"required,eqfield=Password"`
}

// ChargeBalanceInput struct holds required data to charge users' balance
type ChargeBalanceInput struct {
	CardType     string `json:"card_type" binding:"required"`
	PaymentToken string `json:"payment_method_id" binding:"required"`
	Amount       uint64 `json:"amount" binding:"required"`
}

// RegisterResponse struct holds data returned when user registers
type RegisterResponse struct {
	Email   string `json:"email"`
	Timeout string `json:"timeout"`
}

// RefreshTokenResponse struct holds data returned when user refreshes token
type RefreshTokenResponse struct {
	AccessToken string `json:"access_token"`
}

// chargeBalanceResponse struct holds the response data after charging balance
type ChargeBalanceResponse struct {
	PaymentIntentID string  `json:"payment_intent_id"`
	NewBalance      float64 `json:"new_balance"`
}

// UserBalanceResponse struct holds the response data for user balance
type UserBalanceResponse struct {
	BalanceUSD float64 `json:"balance_usd"`
	DebtUSD    float64 `json:"debt_usd"`
}

// SSHKeyInput struct for adding SSH keys
type SSHKeyInput struct {
	Name      string `json:"name" binding:"required"`
	PublicKey string `json:"public_key" binding:"required"`
}

// ResendCodeInput struct for resending verification code
type ResendCodeInput struct {
	Email string `json:"email" binding:"required,email"`
	Type  string `json:"type" binding:"required,oneof=verify reset"` // "verify" for sign-up, "reset" for password reset
}

// @Summary Register a user
// @Description Registers a new user to the system
// @Tags users
// @ID register-user
// @Accept json
// @Produce json
// @Param body body RegisterInput true "Register Input"
// @Success 201 {object} RegisterResponse
// @Failure 400 {object} APIResponse "Invalid request format"
// @Failure 409 {object} APIResponse "User already registered"
// @Failure 500 {object} APIResponse
// @Router /user/register [post]
// RegisterHandler registers user to the system
func (h *Handler) RegisterHandler(c *gin.Context) {
	var request RegisterInput

	// check on request format
	if err := c.ShouldBindJSON(&request); err != nil {
		log.Error().Err(err).Send()
		Error(c, http.StatusBadRequest, "Invalid request format", err.Error())
		return
	}

	// password and confirm password should match
	if request.Password != request.ConfirmPassword {
		Error(c, http.StatusBadRequest, "Validation Error", "password and confirm password don't match")
		return
	}

	// check if user previously exists
	existingUser, getErr := h.db.GetUserByEmail(request.Email)
	if getErr != gorm.ErrRecordNotFound {
		if existingUser.Verified {
			Error(c, http.StatusConflict, "Conflict", "user already registered")
			return
		}
	}

	code := internal.GenerateRandomCode()
	fmt.Printf("code: %v\n", code)
	subject, body := h.mailService.SignUpMailContent(code, h.config.MailSender.Timeout, request.Name, h.config.Server.Host)

	err := h.mailService.SendMail(h.config.MailSender.Email, request.Email, subject, body)
	if err != nil {
		log.Error().Err(err).Msg("failed to send verification code")
		InternalServerError(c)
		return
	}

	// hash password
	hashedPassword, err := internal.HashAndSaltPassword([]byte(request.Password))
	if err != nil {
		log.Error().Err(err).Msg("error hashing password")
		InternalServerError(c)
		return
	}

	isAdmin := internal.Contains(h.config.Admins, request.Email)

	user := models.User{
		Username: request.Name,
		Email:    request.Email,
		Password: hashedPassword,
		Code:     code,
		Verified: existingUser.Verified,
		Admin:    isAdmin,
	}

	if getErr != gorm.ErrRecordNotFound {
		if !user.Verified {
			user.ID = existingUser.ID
			err = h.db.UpdateUserByID(&user)
			if err != nil {
				log.Error().Err(err).Send()
				InternalServerError(c)
				return
			}
		}
	}

	// create user model in db
	if getErr != nil {
		mnemonic, _, err := internal.SetupUserOnTFChain(h.substrateClient, h.config)
		if err != nil {
			log.Error().Err(err).Msg("failed to setup user on TFChain")
			InternalServerError(c)
			return
		}
		customer, err := internal.CreateStripeCustomer(request.Name, request.Email)
		if err != nil {
			log.Error().Err(err).Send()
			InternalServerError(c)
			return
		}

		user.Mnemonic = mnemonic
		user.StripeCustomerID = customer.ID

		err = h.db.RegisterUser(&user)
		if err != nil {
			log.Error().Err(err).Send()
			InternalServerError(c)
			return
		}
	}

	Success(c, http.StatusCreated, "Verification code sent successfully", RegisterResponse{
		Email:   request.Email,
		Timeout: fmt.Sprintf("%d seconds", h.config.MailSender.Timeout),
	})

}

// @Summary Verify registration code
// @Description Verifies the email using the registration code
// @Tags users
// @ID verify-register-code
// @Accept json
// @Produce json
// @Param body body VerifyCodeInput true "Verify Code Input"
// @Success 201 {object} internal.TokenPair
// @Failure 400 {object} APIResponse "Invalid request format or verification failed"
// @Failure 500 {object} APIResponse
// @Router /user/register/verify [post]
// VerifyRegisterCode verifies email when signing up
func (h *Handler) VerifyRegisterCode(c *gin.Context) {
	var request VerifyCodeInput

	if err := c.ShouldBindJSON(&request); err != nil {
		log.Error().Err(err).Send()
		Error(c, http.StatusBadRequest, "Invalid request format", err.Error())
		return
	}

	// get user by email
	user, err := h.db.GetUserByEmail(request.Email)
	if err != nil {
		log.Error().Err(err).Msg("failed to get user by email")
		Error(c, http.StatusBadRequest, "verification failed", "email or password is incorrect")
		return

	}

	if user.Verified {
		Error(c, http.StatusBadRequest, "verification failed", "user already registered")
		return
	}

	if user.Code != request.Code {
		Error(c, http.StatusBadRequest, "verification failed", "wrong code")
		return
	}

	if user.UpdatedAt.Add(time.Duration(h.config.MailSender.Timeout) * time.Second).Before(time.Now()) {
		Error(c, http.StatusBadRequest, "verification failed", "code has expired")
		return
	}

	err = h.db.UpdateUserVerification(user.ID, true)
	if err != nil {
		log.Error().Err(err).Send()
		InternalServerError(c)
		return

	}

	subject, body := h.mailService.WelcomeMailContent(user.Username, h.config.Server.Host)
	err = h.mailService.SendMail(h.config.MailSender.Email, request.Email, subject, body)
	if err != nil {
		log.Error().Err(err).Send()
		InternalServerError(c)
		return
	}

	// create token pairs
	tokenPair, err := h.tokenManager.CreateTokenPair(user.ID, user.Username, user.Admin)
	if err != nil {
		log.Error().Err(err).Msg("Failed to generate token pair")
		InternalServerError(c)
		return
	}
	Success(c, http.StatusCreated, "token pair generated", tokenPair)
}

// @Summary Resend verification code
// @Description Resends a verification code to the user's email for verification or password reset
// @Tags users
// @ID resend-code
// @Accept json
// @Produce json
// @Param body body ResendCodeInput true "Resend Code Input"
// @Success 200 {object} RegisterResponse
// @Failure 400 {object} APIResponse "Invalid request format or already verified"
// @Failure 500 {object} APIResponse
// @Router /user/resend_code [post]
func (h *Handler) ResendCodeHandler(c *gin.Context) {
	var request ResendCodeInput

	if err := c.ShouldBindJSON(&request); err != nil {
		log.Error().Err(err).Send()
		Error(c, http.StatusBadRequest, "Invalid request format", err.Error())
		return
	}

	user, err := h.db.GetUserByEmail(request.Email)
	if err != nil {
		log.Error().Err(err).Msg("failed to get user by email")
		Error(c, http.StatusBadRequest, "Invalid request", "email not found")
		return
	}

	if request.Type == "verify" && user.Verified {
		Error(c, http.StatusBadRequest, "Already verified", "user already verified")
		return
	}

	// Generate new verification code
	code := internal.GenerateRandomCode()
	fmt.Printf("code from resend: %v\n", code)
	user.Code = code
	user.UpdatedAt = time.Now()

	if err := h.db.UpdateUserByID(&user); err != nil {
		log.Error().Err(err).Msg("failed to update verification code")
		InternalServerError(c)
		return
	}

	// Prepare and send email based on type
	var subject, body string
	switch request.Type {
	case "verify":
		subject, body = h.mailService.SignUpMailContent(code, h.config.MailSender.Timeout, user.Username, h.config.Server.Host)
	case "reset":
		subject, body = h.mailService.ResetPasswordMailContent(code, h.config.MailSender.Timeout, user.Username, h.config.Server.Host)
	default:
		Error(c, http.StatusBadRequest, "Invalid type", "type must be either 'verify' or 'reset'")
		return
	}

	if err := h.mailService.SendMail(h.config.MailSender.Email, user.Email, subject, body); err != nil {
		log.Error().Err(err).Msg("failed to send verification code")
		InternalServerError(c)
		return
	}

	Success(c, http.StatusOK, "Verification code resent successfully", RegisterResponse{
		Email:   user.Email,
		Timeout: fmt.Sprintf("%d seconds", h.config.MailSender.Timeout),
	})
}

// @Summary Login user
// @Description Logs a user into the system
// @Tags users
// @ID login-user
// @Accept json
// @Produce json
// @Param body body LoginInput true "Login Input"
// @Success 201 {object} internal.TokenPair
// @Failure 400 {object} APIResponse "Invalid request format"
// @Failure 401 {object} APIResponse "Login failed"
// @Failure 500 {object} APIResponse
// @Router /user/login [post]
// LoginUserHandler logs user into the system
func (h *Handler) LoginUserHandler(c *gin.Context) {
	var request LoginInput

	// check on request format
	if err := c.ShouldBindJSON(&request); err != nil {
		Error(c, http.StatusBadRequest, "Invalid request format", err.Error())
		return
	}

	// get user by email
	user, err := h.db.GetUserByEmail(request.Email)
	if err != nil {
		log.Error().Err(err).Msg("failed to get user by email")
		Error(c, http.StatusBadRequest, "verification failed", "email or password is incorrect")
		return

	}

	// verify password
	match := internal.VerifyPassword(user.Password, request.Password)
	if !match {
		Error(c, http.StatusUnauthorized, "login failed", "email or password is incorrect")
		return
	}

	// create token pairs
	tokenPair, err := h.tokenManager.CreateTokenPair(user.ID, user.Username, user.Admin)
	if err != nil {
		log.Error().Err(err).Msg("Failed to generate token pair")
		InternalServerError(c)
		return
	}
	Success(c, http.StatusCreated, "token pair generated", tokenPair)
}

// @Summary Refresh access token
// @Description Refreshes the access token using a valid refresh token
// @Tags users
// @ID refresh-token
// @Accept json
// @Produce json
// @Param body body RefreshTokenInput true "Refresh Token Input"
// @Success 201 {object} RefreshTokenResponse
// @Failure 400 {object} APIResponse "Invalid request format"
// @Failure 401 {object} APIResponse "Invalid or expired refresh token"
// @Failure 500 {object} APIResponse
// @Router /user/refresh [post]
// RefreshTokenHandler handles token refresh requests
func (h *Handler) RefreshTokenHandler(c *gin.Context) {
	var request RefreshTokenInput

	if err := c.ShouldBindJSON(&request); err != nil {
		log.Error().Err(err).Send()
		Error(c, http.StatusBadRequest, "Invalid request format", err.Error())
		return
	}

	accessToken, err := h.tokenManager.AccessTokenFromRefresh(request.RefreshToken)
	if err != nil {
		log.Error().Err(err).Send()
		Error(c, http.StatusUnauthorized, "refresh token failed", "Invalid or expired refresh token")

		return
	}

	Success(c, http.StatusCreated, "access token refreshed successfully", RefreshTokenResponse{
		AccessToken: accessToken,
	})
}

// @Summary Forgot password
// @Description Sends a verification code to the user's email for password reset
// @Tags users
// @ID forgot-password
// @Accept json
// @Produce json
// @Param body body EmailInput true "Email Input"
// @Success 200 {object} RegisterResponse
// @Failure 400 {object} APIResponse "Invalid request format"
// @Failure 404 {object} APIResponse "User is not found"
// @Failure 500 {object} APIResponse
// @Router /user/forgot_password [post]
// ForgotPasswordHandler sends user verification code
func (h *Handler) ForgotPasswordHandler(c *gin.Context) {
	var request EmailInput

	if err := c.ShouldBindJSON(&request); err != nil {
		log.Error().Err(err).Send()
		Error(c, http.StatusBadRequest, "Invalid request format", err.Error())
		return
	}

	// get user by email
	user, err := h.db.GetUserByEmail(request.Email)
	if err != nil {
		log.Error().Err(err).Msg("failed to get user ")
		Error(c, http.StatusNotFound, "user lookup failed", "failed to get user")
		return

	}

	code := internal.GenerateRandomCode()
	subject, body := h.mailService.ResetPasswordMailContent(code, h.config.MailSender.Timeout, user.Username, h.config.Server.Host)
	err = h.mailService.SendMail(h.config.MailSender.Email, request.Email, subject, body)

	if err != nil {
		log.Error().Err(err).Msg("failed to send verification code")
		InternalServerError(c)
		return
	}

	err = h.db.UpdateUserByID(
		&models.User{
			ID:        user.ID,
			UpdatedAt: time.Now(),
			Code:      code,
		},
	)

	if err != nil {
		log.Error().Err(err).Msg("error updating user data")
		InternalServerError(c)
		return
	}

	Success(c, http.StatusOK, "Verification code sent", RegisterResponse{
		Email:   request.Email,
		Timeout: fmt.Sprintf("%d seconds", h.config.MailSender.Timeout),
	})

}

// @Summary Verify forgot password code
// @Description Verifies the code sent to the user's email for password reset
// @Tags users
// @ID verify-forgot-password-code
// @Accept json
// @Produce json
// @Param body body VerifyCodeInput true "Verify Code Input"
// @Success 201 {object} internal.TokenPair "Verification successful"
// @Failure 400 {object} APIResponse "Invalid request format or verification failed"
// @Failure 500 {object} APIResponse
// @Router /user/forgot_password/verify [post]
// VerifyForgetPasswordCodeHandler verifies code sent to user when forgetting password
func (h *Handler) VerifyForgetPasswordCodeHandler(c *gin.Context) {
	var request VerifyCodeInput

	if err := c.ShouldBindJSON(&request); err != nil {
		log.Error().Err(err).Send()
		Error(c, http.StatusBadRequest, "Invalid request format", err.Error())
		return
	}

	// get user by email
	user, err := h.db.GetUserByEmail(request.Email)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			Error(c, http.StatusBadRequest, "Invalid request format", err.Error())
			return

		}
		log.Error().Err(err).Msg("failed to get user by email")
		InternalServerError(c)
		return

	}

	if user.Code != request.Code {
		Error(c, http.StatusBadRequest, "Invalid code", "")
		return
	}

	if user.UpdatedAt.Add(time.Duration(h.config.MailSender.Timeout) * time.Second).Before(time.Now()) {
		Error(c, http.StatusBadRequest, "code expired", "verification code has expired")

		return
	}
	isAdmin := internal.Contains(h.config.Admins, request.Email)

	// create token pairs
	tokenPair, err := h.tokenManager.CreateTokenPair(user.ID, user.Username, isAdmin)
	if err != nil {
		log.Error().Err(err).Msg("Failed to generate token pair")
		InternalServerError(c)
		return
	}

	Success(c, http.StatusCreated, "Verification successful", tokenPair)
}

// @Summary Change password
// @Description Changes the user's password
// @Tags users
// @ID change-password
// @Accept json
// @Produce json
// @Param body body ChangePasswordInput true "Change Password Input"
// @Success 202 {object} APIResponse "Password updated successfully"
// @Failure 400 {object} APIResponse "Invalid request format or password mismatch"
// @Failure 404 {object} APIResponse "User is not found"
// @Failure 500 {object} APIResponse
// @Router /user/change_password [put]
// ChangePasswordHandler changes password of user
func (h *Handler) ChangePasswordHandler(c *gin.Context) {
	var request ChangePasswordInput

	if err := c.ShouldBindJSON(&request); err != nil {
		log.Error().Err(err).Send()
		Error(c, http.StatusBadRequest, "Invalid request format", err.Error())

		return
	}

	if request.Password != request.ConfirmPassword {
		Error(c, http.StatusBadRequest, "password mismatch", "password and confirm password don't match")
		return
	}

	// hash password
	hashedPassword, err := internal.HashAndSaltPassword([]byte(request.Password))
	if err != nil {
		log.Error().Err(err).Msg("error hashing password")
		InternalServerError(c)
		return
	}

	err = h.db.UpdatePassword(request.Email, hashedPassword)
	if err == gorm.ErrRecordNotFound {
		log.Error().Err(err).Msg("user is not found")
		Error(c, http.StatusNotFound, "user is not found", err.Error())

		return
	}

	if err != nil {
		log.Error().Err(err).Send()
		InternalServerError(c)
		return

	}

	Success(c, http.StatusAccepted, "password is updated successfully", nil)

}

// @Summary Charge user balance
// @Description Charges the user's balance using a payment method
// @Tags users
// @ID charge-balance
// @Accept json
// @Produce json
// @Param body body ChargeBalanceInput true "Charge Balance Input"
// @Success 201 {object} ChargeBalanceResponse "Balance is charged successfully"
// @Failure 400 {object} APIResponse "Invalid request format or amount"
// @Failure 404 {object} APIResponse "User is not found"
// @Failure 500 {object} APIResponse
// @Router /user/balance/charge [post]
// ChargeBalance charges the user's balance
func (h *Handler) ChargeBalance(c *gin.Context) {
	userID := c.GetInt("user_id")

	var request ChargeBalanceInput
	if err := c.ShouldBindJSON(&request); err != nil {
		log.Error().Err(err).Send()
		Error(c, http.StatusBadRequest, "Invalid request format", err.Error())
		return
	}

	if request.Amount <= 0 {
		Error(c, http.StatusBadRequest, "Amount must be greater than zero", "")
		return
	}

	user, err := h.db.GetUserByID(userID)
	if err != nil {
		log.Error().Err(err).Send()
		Error(c, http.StatusNotFound, "User is not found", "")
		return
	}

	paymentMethod, err := internal.CreatePaymentMethod(request.CardType, request.PaymentToken)
	if err != nil {
		log.Error().Err(err).Msg("error creating payment method")
		InternalServerError(c)
		return
	}

	_, err = paymentmethod.Attach(paymentMethod.ID, &stripe.PaymentMethodAttachParams{
		Customer: stripe.String(user.StripeCustomerID),
	})
	if err != nil {
		log.Error().Err(err).Msg("error attaching payment method to customer")
		InternalServerError(c)
		return
	}

	intent, err := internal.CreatePaymentIntent(user.StripeCustomerID, paymentMethod.ID, h.config.Currency, request.Amount)
	if err != nil {
		log.Error().Err(err).Msg("error creating payment intent")
		InternalServerError(c)
		return
	}

	user.CreditCardBalance += float64(request.Amount)

	err = h.db.UpdateUserByID(&user)
	if err != nil {
		log.Error().Err(err).Msg("error updating user data")
		InternalServerError(c)
		return
	}

	err = internal.TransferTFTs(h.substrateClient, request.Amount, user.Mnemonic, h.config.SystemAccount.Mnemonics)
	if err != nil {
		log.Error().Err(err).Send()
		InternalServerError(c)
		return
	}

	Success(c, http.StatusCreated, "Balance is charged successfully", ChargeBalanceResponse{
		PaymentIntentID: intent.ID,
		NewBalance:      user.CreditCardBalance,
	})

}

// @Summary Get user details
// @Description Retrieves all data of the user
// @Tags users
// @ID get-user
// @Produce json
// @Success 200 {object} models.User "User is retrieved successfully"
// @Failure 404 {object} APIResponse "User is not found"
// @Failure 500 {object} APIResponse
// @Router /user [get]
// GetUserHandler retrieves all data of the user
func (h *Handler) GetUserHandler(c *gin.Context) {
	userID := c.GetInt("user_id")

	user, err := h.db.GetUserByID(userID)
	if err != nil {
		log.Error().Err(err).Send()
		Error(c, http.StatusNotFound, "User is not found", "")
		return
	}

	Success(c, http.StatusOK, "User is retrieved successfully", gin.H{
		"user": user,
	})
}

// @Summary Get user balance
// @Description Retrieves the user's balance in USD
// @Tags users
// @ID get-user-balance
// @Produce json
// @Success 200 {object} UserBalanceResponse "Balance fetched successfully"
// @Failure 404 {object} APIResponse "User is not found"
// @Failure 500 {object} APIResponse
// @Router /user/balance [get]
// GetUserBalance returns user's balance in usd
func (h *Handler) GetUserBalance(c *gin.Context) {
	userID := c.GetInt("user_id")

	user, err := h.db.GetUserByID(userID)
	if err != nil {
		log.Error().Err(err).Send()
		Error(c, http.StatusNotFound, "User is not found", "")
		return
	}
	usdBalance, err := internal.GetUserBalanceUSD(h.substrateClient, user.Mnemonic, user.Debt)
	if err != nil {
		log.Error().Err(err).Send()
		InternalServerError(c)
	}
	Success(c, http.StatusOK, "Balance is fetched", UserBalanceResponse{
		BalanceUSD: usdBalance,
		DebtUSD:    user.Debt,
	})
}

// @Summary Redeem voucher
// @Description Redeems a voucher for the user
// @Tags users
// @ID redeem-voucher
// @Param voucher_code path string true "Voucher Code"
// @Produce json
// @Success 202 {object} APIResponse "Voucher redeemed successfully"
// @Failure 400 {object} APIResponse "Invalid voucher code or already redeemed"
// @Failure 404 {object} APIResponse "User or voucher are not found"
// @Failure 500 {object} APIResponse
// @Router /user/redeem/{voucher_code} [put]
// RedeemVoucherHandler redeems a voucher for the user
func (h *Handler) RedeemVoucherHandler(c *gin.Context) {
	voucherCodeParam := c.Param("voucher_code")
	if voucherCodeParam == "" {
		Error(c, http.StatusBadRequest, "Voucher Code is required", "")
		return
	}
	userID := c.GetInt("user_id")

	user, err := h.db.GetUserByID(userID)
	if err != nil {
		log.Error().Err(err).Send()
		Error(c, http.StatusNotFound, "User is not found", "")
		return
	}

	// check voucher exists
	voucher, err := h.db.GetVoucherByCode(voucherCodeParam)
	if err != nil {
		log.Error().Err(err).Send()
		Error(c, http.StatusNotFound, "Voucher is not found", "")
		return
	}

	// check voucher not redeemed
	if voucher.Redeemed {
		Error(c, http.StatusBadRequest, "Voucher is already redeemed", "")
		return
	}

	// check on expiration time of voucher
	if voucher.ExpiresAt.Before(time.Now()) {
		Error(c, http.StatusBadRequest, "Voucher is already expired", "")
		return

	}

	err = h.db.RedeemVoucher(voucher.Code)
	if err != nil {
		log.Error().Err(err).Send()
		InternalServerError(c)
		return
	}

	user.CreditedBalance += voucher.Value
	err = h.db.UpdateUserByID(&user)
	if err != nil {
		log.Error().Err(err).Send()
		InternalServerError(c)
		return
	}

	err = internal.TransferTFTs(h.substrateClient, uint64(voucher.Value), user.Mnemonic, h.config.SystemAccount.Mnemonics)
	if err != nil {
		log.Error().Err(err).Send()
		InternalServerError(c)
		return
	}

	Success(c, http.StatusOK, "Voucher is redeemed successfully", nil)
}

// @Summary List user SSH keys
// @Description Lists all SSH keys for the authenticated user
// @Tags users
// @ID list-ssh-keys
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {array} models.SSHKey
// @Failure 401 {object} APIResponse "Unauthorized"
// @Failure 500 {object} APIResponse
// @Router /user/ssh-keys [get]
// ListSSHKeysHandler lists all SSH keys for the authenticated user
func (h *Handler) ListSSHKeysHandler(c *gin.Context) {
	userID := c.GetInt("user_id")
	if userID == 0 {
		Error(c, http.StatusUnauthorized, "Unauthorized", "user not authenticated")
		return
	}

	sshKeys, err := h.db.ListUserSSHKeys(userID)
	if err != nil {
		log.Error().Err(err).Msg("failed to list SSH keys")
		InternalServerError(c)
		return
	}

	Success(c, http.StatusOK, "SSH keys retrieved successfully", sshKeys)
}

// @Summary Add SSH key
// @Description Adds a new SSH key for the authenticated user
// @Tags users
// @ID add-ssh-key
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body SSHKeyInput true "SSH Key Input"
// @Success 201 {object} models.SSHKey
// @Failure 400 {object} APIResponse "Invalid request format"
// @Failure 401 {object} APIResponse "Unauthorized"
// @Failure 500 {object} APIResponse
// @Router /user/ssh-keys [post]
// AddSSHKeyHandler adds a new SSH key for the authenticated user
func (h *Handler) AddSSHKeyHandler(c *gin.Context) {
	userID := c.GetInt("user_id")
	if userID == 0 {
		Error(c, http.StatusUnauthorized, "Unauthorized", "user not authenticated")
		return
	}

	var request SSHKeyInput
	if err := c.ShouldBindJSON(&request); err != nil {
		log.Error().Err(err).Send()
		Error(c, http.StatusBadRequest, "Invalid request format", err.Error())
		return
	}

	// Validate SSH key format
	if err := internal.ValidateSSH(request.PublicKey); err != nil {
		Error(c, http.StatusBadRequest, "Validation Error", "invalid SSH key format")
		return
	}

	sshKey := models.SSHKey{
		UserID:    userID,
		Name:      request.Name,
		PublicKey: request.PublicKey,
	}

	if err := h.db.CreateSSHKey(&sshKey); err != nil {
		log.Error().Err(err).Msg("failed to create SSH key")
		InternalServerError(c)
		return
	}

	Success(c, http.StatusCreated, "SSH key added successfully", sshKey)
}

// @Summary Delete SSH key
// @Description Deletes an SSH key for the authenticated user
// @Tags users
// @ID delete-ssh-key
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param ssh_key_id path int true "SSH Key ID"
// @Success 200 {object} APIResponse
// @Failure 400 {object} APIResponse "Invalid SSH key ID"
// @Failure 401 {object} APIResponse "Unauthorized"
// @Failure 404 {object} APIResponse "SSH key not found"
// @Failure 500 {object} APIResponse
// @Router /user/ssh-keys/{ssh_key_id} [delete]
// DeleteSSHKeyHandler deletes an SSH key for the authenticated user
func (h *Handler) DeleteSSHKeyHandler(c *gin.Context) {
	userID := c.GetInt("user_id")
	if userID == 0 {
		Error(c, http.StatusUnauthorized, "Unauthorized", "user not authenticated")
		return
	}

	sshKeyID := c.Param("ssh_key_id")
	if sshKeyID == "" {
		Error(c, http.StatusBadRequest, "Invalid request", "SSH key ID is required")
		return
	}

	// Convert sshKeyID to int
	var keyID int
	keyID, err := strconv.Atoi(sshKeyID)
	if err != nil {
		Error(c, http.StatusBadRequest, "Invalid request", "invalid SSH key ID format")
		return
	}

	if err := h.db.DeleteSSHKey(keyID, userID); err != nil {
		if err.Error() == fmt.Sprintf("no SSH key found with ID %d for user %d", keyID, userID) {
			Error(c, http.StatusNotFound, "Not Found", "SSH key not found")
			return
		}
		log.Error().Err(err).Msg("failed to delete SSH key")
		InternalServerError(c)
		return
	}

	Success(c, http.StatusOK, "SSH key deleted successfully", nil)
}
