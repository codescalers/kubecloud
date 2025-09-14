package app

import (
	"context"
	"errors"
	"fmt"
	"kubecloud/internal"
	"kubecloud/internal/metrics"
	"kubecloud/internal/notification"
	"kubecloud/models"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/mattn/go-sqlite3"
	"github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/paymentmethod"
	substrate "github.com/threefoldtech/tfchain/clients/tfchain-client-go"
	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/graphql"
	proxy "github.com/threefoldtech/tfgrid-sdk-go/grid-proxy/pkg/client"
	"github.com/xmonader/ewf"

	"kubecloud/internal/constants"
	"kubecloud/internal/logger"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/vedhavyas/go-subkey"
)

// Handler struct holds configs for all handlers
type Handler struct {
	tokenManager        internal.TokenManager
	db                  models.DB
	config              internal.Configuration
	mailService         internal.MailService
	proxyClient         proxy.Client
	substrateClient     *substrate.Substrate
	graphqlClient       graphql.GraphQl
	firesquidClient     graphql.GraphQl
	redis               *internal.RedisClient
	sseManager          *internal.SSEManager
	ewfEngine           *ewf.Engine
	gridNet             string // Network name for the grid
	sshPublicKey        string // SSH public key loaded at startup
	systemIdentity      substrate.Identity
	kycClient           *internal.KYCClient
	sponsorKeyPair      subkey.KeyPair
	sponsorAddress      string
	metrics             *metrics.Metrics
	notificationService *notification.NotificationService
}

// NewHandler create new handler
func NewHandler(tokenManager internal.TokenManager, db models.DB,
	config internal.Configuration, mailService internal.MailService,
	gridproxy proxy.Client, substrateClient *substrate.Substrate,
	graphqlClient graphql.GraphQl, firesquidClient graphql.GraphQl,
	redis *internal.RedisClient, sseManager *internal.SSEManager, ewfEngine *ewf.Engine,
	gridNet string, sshPublicKey string, systemIdentity substrate.Identity,
	kycClient *internal.KYCClient, sponsorKeyPair subkey.KeyPair, sponsorAddress string,
	metrics *metrics.Metrics, notificationService *notification.NotificationService) *Handler {

	return &Handler{
		tokenManager:        tokenManager,
		db:                  db,
		config:              config,
		mailService:         mailService,
		proxyClient:         gridproxy,
		substrateClient:     substrateClient,
		graphqlClient:       graphqlClient,
		firesquidClient:     firesquidClient,
		redis:               redis,
		sseManager:          sseManager,
		ewfEngine:           ewfEngine,
		gridNet:             gridNet,
		sshPublicKey:        sshPublicKey,
		systemIdentity:      systemIdentity,
		kycClient:           kycClient,
		sponsorKeyPair:      sponsorKeyPair,
		sponsorAddress:      sponsorAddress,
		metrics:             metrics,
		notificationService: notificationService,
	}
}

// RegisterInput struct for data needed when user register
type RegisterInput struct {
	Name            string `json:"name" validate:"required,min=3,max=64"`
	Email           string `json:"email" validate:"required,email"`
	Password        string `json:"password" validate:"required,min=8,max=64"`
	ConfirmPassword string `json:"confirm_password" validate:"required,eqfield=Password"`
}

// LoginInput struct for login handler
type LoginInput struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=3,max=64"`
}

// RefreshTokenInput struct when user refresh token
type RefreshTokenInput struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// EmailInput struct for user when forgetting password
type EmailInput struct {
	Email string `json:"email" validate:"required,email"`
}

// VerifyCodeInput struct takes verification code from user
type VerifyCodeInput struct {
	Email string `json:"email" validate:"required,email"`
	Code  int    `json:"code" validate:"required"`
}

// ChangePasswordInput struct for user to change password
type ChangePasswordInput struct {
	Email           string `json:"email" validate:"required,email"`
	Password        string `json:"password" validate:"required,min=8,max=64"`
	ConfirmPassword string `json:"confirm_password" validate:"required,eqfield=Password"`
}

// ChargeBalanceInput struct holds required data to charge users' balance
type ChargeBalanceInput struct {
	CardType     string  `json:"card_type" validate:"required"`
	PaymentToken string  `json:"payment_method_id" validate:"required"`
	Amount       float64 `json:"amount" validate:"required,gt=0"`
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

// ChargeBalanceResponse holds the response for charging user balance
type ChargeBalanceResponse struct {
	WorkflowID string `json:"workflow_id"`
	Email      string `json:"email"`
}

// UserBalanceResponse struct holds the response data for user balance
type UserBalanceResponse struct {
	BalanceUSD        float64 `json:"balance_usd"`
	DebtUSD           float64 `json:"debt_usd"`
	PendingBalanceUSD float64 `json:"pending_balance_usd"`
}

// SSHKeyInput struct for adding SSH keys
type SSHKeyInput struct {
	Name      string `json:"name" validate:"required"`
	PublicKey string `json:"public_key" validate:"required"`
}

// RegisterUserResponse holds the response for user registration
type RegisterUserResponse struct {
	WorkflowID string `json:"workflow_id"`
	Email      string `json:"email"`
}

type VerifyRegisterUserResponse struct {
	WorkflowID string `json:"workflow_id"`
	Email      string `json:"email"`
	*internal.TokenPair
}

// RedeemVoucherResponse holds the response for redeeming a voucher
type RedeemVoucherResponse struct {
	WorkflowID  string  `json:"workflow_id"`
	VoucherCode string  `json:"voucher_code"`
	Amount      float64 `json:"amount"`
	Email       string  `json:"email"`
}

type GetUserResponse struct {
	models.User
	PendingBalanceUSD float64 `json:"pending_balance_usd"`
}

// RegisterHandler registers user to the system
// @Summary Register user (with KYC sponsorship)
// @Description Registers a new user, sets up blockchain account, and creates KYC sponsorship. Sends verification code to email.
// @Tags users
// @ID register-user
// @Accept json
// @Produce json
// @Param body body RegisterInput true "Register Input"
// @Success 201 {object} RegisterUserResponse "workflow_id: string, email: string"
// @Failure 400 {object} APIResponse "Invalid request format"
// @Failure 409 {object} APIResponse "User is already registered"
// @Failure 500 {object} APIResponse "Internal server error"
// @Router /user/register [post]
func (h *Handler) RegisterHandler(c *gin.Context) {
	var request RegisterInput

	// check on request format
	if err := c.ShouldBindJSON(&request); err != nil {
		logger.GetLogger().Error().Err(err).Send()
		Error(c, http.StatusBadRequest, "Invalid request format", err.Error())
		return
	}

	if err := internal.ValidateStruct(request); err != nil {
		Error(c, http.StatusBadRequest, "Validation failed", err.Error())
		return
	}

	// check if user previously exists
	existingUser, getErr := h.db.GetUserByEmail(request.Email)
	if getErr != nil && getErr != gorm.ErrRecordNotFound {
		InternalServerError(c)
		return
	}

	if getErr != gorm.ErrRecordNotFound {
		if isUserRegistered(existingUser) {
			Error(c, http.StatusConflict, "Conflict", "User is already registered")
			return
		}
	}

	wf, err := h.ewfEngine.NewWorkflow(constants.WorkflowUserRegistration)
	if err != nil {
		logger.GetLogger().Error().Err(err).Msg("failed to start registration workflow")
		InternalServerError(c)
		return
	}

	wf.State = ewf.State{
		"name":     request.Name,
		"email":    request.Email,
		"password": request.Password,
	}

	h.ewfEngine.RunAsync(context.Background(), wf)

	Success(c, http.StatusCreated, "Registration in progress. You can check its status using the workflow id.", RegisterUserResponse{
		WorkflowID: wf.UUID,
		Email:      request.Email,
	})
}

// @Summary Verify registration code
// @Description Verifies the email using the registration code
// @Tags users
// @ID verify-register-code
// @Accept json
// @Produce json
// @Param request body VerifyCodeInput true "Verification details"
// @Success 201 {object} VerifyRegisterUserResponse
// @Failure 400 {object} APIResponse "Invalid request"
// @Failure 409 {object} APIResponse "User is already registered"
// @Failure 500 {object} APIResponse "Internal server error"
// @Router /user/register/verify [post]
func (h *Handler) VerifyRegisterCode(c *gin.Context) {
	var request VerifyCodeInput

	if err := c.ShouldBindJSON(&request); err != nil {
		logger.GetLogger().Error().Err(err).Send()
		Error(c, http.StatusBadRequest, "Invalid request format", err.Error())
		return
	}

	if err := internal.ValidateStruct(request); err != nil {
		Error(c, http.StatusBadRequest, "Validation failed", err.Error())
		return
	}

	// get user by email
	user, err := h.db.GetUserByEmail(request.Email)
	if err != nil {
		logger.GetLogger().Error().Err(err).Msg("failed to get user by email")
		Error(c, http.StatusBadRequest, "verification failed", "Make sure you have registered before")
		return
	}

	// check if user is already registered (all required fields are set)
	if isUserRegistered(user) {
		Error(c, http.StatusConflict, "verification failed", "User is already registered")
		return
	}

	// check verification if user is not verified
	if !user.Verified {
		if user.Code != request.Code {
			Error(c, http.StatusBadRequest, "verification failed", "Invalid verification code")
			return
		}

		if user.UpdatedAt.Add(time.Duration(h.config.MailSender.TimeoutMin) * time.Minute).Before(time.Now()) {
			Error(c, http.StatusBadRequest, "verification failed", "code has expired")
			return
		}

		if err := h.db.UpdateUserByID(&models.User{
			ID:       user.ID,
			Verified: true,
		}); err != nil {
			logger.GetLogger().Error().Err(err).Msg("failed to update user data")
			InternalServerError(c)
			return
		}
		payload := notification.CommonPayload{
			Message: "User email is verified successfully",
			Subject: "User email verified",
		}
		notification := models.NewNotification(user.ID, "user_registration", notification.MergePayload(payload, map[string]string{}), models.WithNoPersist(), models.WithChannels(notification.ChannelUI), models.WithSeverity(models.NotificationSeveritySuccess))
		err = h.notificationService.Send(context.Background(), notification)
		if err != nil {
			logger.GetLogger().Error().Err(err).Msg("failed to send user registration notification")
		}
	}

	wf, err := h.ewfEngine.NewWorkflow(constants.WorkflowUserVerification)
	if err != nil {
		logger.GetLogger().Error().Err(err).Msg("failed to start user verification workflow")
		InternalServerError(c)
		return
	}

	if err = h.ewfEngine.Store().SaveWorkflow(c.Request.Context(), wf); err != nil {
		logger.GetLogger().Error().Err(err).Msg("failed to save user verification workflow")
		InternalServerError(c)
		return
	}

	// Start the user-verification workflow
	wf.State = ewf.State{
		"email":   user.Email,
		"name":    user.Username,
		"user_id": user.ID,
	}

	h.ewfEngine.RunAsync(context.Background(), wf)

	tokenPair, err := h.tokenManager.CreateTokenPair(user.ID, user.Username, user.Admin)
	if err != nil {
		logger.GetLogger().Error().Err(err).Msg("Failed to generate token pair")
		InternalServerError(c)
		return
	}

	Success(c, http.StatusCreated, "Verification is in progress", VerifyRegisterUserResponse{
		WorkflowID: wf.UUID,
		Email:      user.Email,
		TokenPair:  tokenPair,
	})
}

// @Summary Login user (KYC verification checked)
// @Description Logs a user in. Checks KYC verification status and updates user sponsorship status if needed. Login is not blocked by KYC errors.
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

	if err := internal.ValidateStruct(request); err != nil {
		Error(c, http.StatusBadRequest, "Validation failed", err.Error())
		return
	}

	// get user by email
	user, err := h.db.GetUserByEmail(request.Email)
	if err != nil {
		logger.GetLogger().Error().Err(err).Msg("failed to get user by email")
		Error(c, http.StatusBadRequest, "verification failed", "email or password is incorrect")
		return
	}

	// verify password
	match := internal.VerifyPassword(user.Password, request.Password)
	if !match {
		Error(c, http.StatusUnauthorized, "login failed", "email or password is incorrect")
		return
	}

	// Check KYC verification status without blocking login
	sponsored, err := h.kycClient.IsUserVerified(c.Request.Context(), user.AccountAddress)
	if err != nil {
		logger.GetLogger().Error().Err(err).Msg("failed to check KYC verification status")
		InternalServerError(c)
		return
	}
	if user.Sponsored != sponsored {
		user.Sponsored = sponsored
		if err := h.db.UpdateUserByID(&user); err != nil {
			logger.GetLogger().Error().Err(err).Msg("failed to update user sponsorship status")
		}
	}

	// create token pairs
	tokenPair, err := h.tokenManager.CreateTokenPair(user.ID, user.Username, user.Admin)
	if err != nil {
		logger.GetLogger().Error().Err(err).Msg("Failed to generate token pair")
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
		logger.GetLogger().Error().Err(err).Send()
		Error(c, http.StatusBadRequest, "Invalid request format", err.Error())
		return
	}

	if err := internal.ValidateStruct(request); err != nil {
		Error(c, http.StatusBadRequest, "Validation failed", err.Error())
		return
	}

	accessToken, err := h.tokenManager.AccessTokenFromRefresh(request.RefreshToken)
	if err != nil {
		logger.GetLogger().Error().Err(err).Send()
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
		logger.GetLogger().Error().Err(err).Send()
		Error(c, http.StatusBadRequest, "Invalid request format", err.Error())
		return
	}

	if err := internal.ValidateStruct(request); err != nil {
		Error(c, http.StatusBadRequest, "Validation failed", err.Error())
		return
	}

	// get user by email
	user, err := h.db.GetUserByEmail(request.Email)
	if err != nil {
		logger.GetLogger().Error().Err(err).Msg("failed to get user ")
		Error(c, http.StatusNotFound, "user lookup failed", "failed to get user")
		return

	}

	code := internal.GenerateRandomCode()
	subject, body := h.mailService.ResetPasswordMailContent(code, h.config.MailSender.TimeoutMin, user.Username, h.config.Server.Host)
	err = h.mailService.SendMail(h.config.MailSender.Email, request.Email, subject, body)

	if err != nil {
		logger.GetLogger().Error().Err(err).Msg("failed to send verification code")
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
		logger.GetLogger().Error().Err(err).Msg("error updating user data")
		InternalServerError(c)
		return
	}

	Success(c, http.StatusOK, "Verification code sent", RegisterResponse{
		Email:   request.Email,
		Timeout: fmt.Sprintf("%d minutes", h.config.MailSender.TimeoutMin),
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
		logger.GetLogger().Error().Err(err).Send()
		Error(c, http.StatusBadRequest, "Invalid request format", err.Error())
		return
	}

	if err := internal.ValidateStruct(request); err != nil {
		Error(c, http.StatusBadRequest, "Validation failed", err.Error())
		return
	}

	// get user by email
	user, err := h.db.GetUserByEmail(request.Email)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			Error(c, http.StatusBadRequest, "Invalid request format", err.Error())
			return

		}
		logger.GetLogger().Error().Err(err).Msg("failed to get user by email")
		InternalServerError(c)
		return

	}

	if user.Code != request.Code {
		Error(c, http.StatusBadRequest, "Invalid code", "")
		return
	}

	if user.UpdatedAt.Add(time.Duration(h.config.MailSender.TimeoutMin) * time.Minute).Before(time.Now()) {
		Error(c, http.StatusBadRequest, "code expired", "verification code has expired")

		return
	}
	isAdmin := internal.Contains(h.config.Admins, request.Email)

	// create token pairs
	tokenPair, err := h.tokenManager.CreateTokenPair(user.ID, user.Username, isAdmin)
	if err != nil {
		logger.GetLogger().Error().Err(err).Msg("Failed to generate token pair")
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
		logger.GetLogger().Error().Err(err).Send()
		Error(c, http.StatusBadRequest, "Invalid request format", err.Error())

		return
	}

	if err := internal.ValidateStruct(request); err != nil {
		Error(c, http.StatusBadRequest, "Validation failed", err.Error())
		return
	}

	// hash password
	hashedPassword, err := internal.HashAndSaltPassword([]byte(request.Password))
	if err != nil {
		logger.GetLogger().Error().Err(err).Msg("error hashing password")
		InternalServerError(c)
		return
	}

	err = h.db.UpdatePassword(request.Email, hashedPassword)
	if err == gorm.ErrRecordNotFound {
		logger.GetLogger().Error().Err(err).Msg("user is not found")
		Error(c, http.StatusNotFound, "user is not found", err.Error())

		return
	}

	if err != nil {
		logger.GetLogger().Error().Err(err).Send()
		InternalServerError(c)
		return

	}

	payload := notification.CommonPayload{
		Status:  "password_changed",
		Subject: "Your password was changed",
		Message: "Your account password has been successfully updated.",
	}

	notification := models.NewNotification(c.GetInt("user_id"), models.NotificationTypeUser, notification.MergePayload(payload, map[string]string{}))
	err = h.notificationService.Send(c, notification)
	if err != nil {
		logger.GetLogger().Error().Err(err).Msg("failed to send password changed notification")
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
// @Success 202 {object} ChargeBalanceResponse "workflow_id: string, email: string"
// @Failure 400 {object} APIResponse "Invalid request format or amount"
// @Failure 404 {object} APIResponse "User is not found"
// @Failure 500 {object} APIResponse "Internal server error"
// @Router /user/balance/charge [post]
func (h *Handler) ChargeBalance(c *gin.Context) {
	userID := c.GetInt("user_id")

	var request ChargeBalanceInput
	if err := c.ShouldBindJSON(&request); err != nil {
		logger.GetLogger().Error().Err(err).Send()
		Error(c, http.StatusBadRequest, "Invalid request format", err.Error())
		return
	}

	if err := internal.ValidateStruct(request); err != nil {
		Error(c, http.StatusBadRequest, "Validation failed", err.Error())
		return
	}

	user, err := h.db.GetUserByID(userID)
	if err != nil {
		logger.GetLogger().Error().Err(err).Send()
		Error(c, http.StatusNotFound, "User is not found", "")
		return
	}

	paymentMethod, err := internal.CreatePaymentMethod(request.CardType, request.PaymentToken)
	if err != nil {
		logger.GetLogger().Error().Err(err).Msg("error creating payment method")
		InternalServerError(c)
		return
	}

	_, err = paymentmethod.Attach(paymentMethod.ID, &stripe.PaymentMethodAttachParams{
		Customer: stripe.String(user.StripeCustomerID),
	})
	if err != nil {
		logger.GetLogger().Error().Err(err).Msg("error attaching payment method to customer")
		InternalServerError(c)
		return
	}

	wf, err := h.ewfEngine.NewWorkflow(constants.WorkflowChargeBalance)
	if err != nil {
		logger.GetLogger().Error().Err(err).Send()
		InternalServerError(c)
		return
	}

	wf.State = ewf.State{
		"user_id":            userID,
		"stripe_customer_id": user.StripeCustomerID,
		"payment_method_id":  paymentMethod.ID,
		"amount":             internal.FromUSDToUSDMillicent(request.Amount),
		"mnemonic":           user.Mnemonic,
		"username":           user.Username,
		"transfer_mode":      models.ChargeBalanceMode,
	}

	h.ewfEngine.RunAsync(context.Background(), wf)

	Success(c, http.StatusCreated, "Charge in progress. You can check its status using the workflow id.", ChargeBalanceResponse{
		WorkflowID: wf.UUID,
		Email:      user.Email,
	})
}

// @Summary Get user details
// @Description Retrieves all data of the user
// @Tags users
// @ID get-user
// @Produce json
// @Success 200 {object} GetUserResponse "User is retrieved successfully"
// @Failure 404 {object} APIResponse "User is not found"
// @Failure 500 {object} APIResponse
// @Router /user [get]
// GetUserHandler retrieves all data of the user
func (h *Handler) GetUserHandler(c *gin.Context) {
	userID := c.GetInt("user_id")

	user, err := h.db.GetUserByID(userID)
	if err != nil {
		logger.GetLogger().Error().Err(err).Send()
		Error(c, http.StatusNotFound, "User is not found", "")
		return
	}

	pendingRecords, err := h.db.ListUserPendingRecords(userID)
	if err != nil {
		logger.GetLogger().Error().Err(err).Msg("failed to list pending records")
		InternalServerError(c)
		return
	}

	var tftPendingAmount uint64
	for _, record := range pendingRecords {
		tftPendingAmount += record.TFTAmount - record.TransferredTFTAmount
	}

	usdMillicentPendingAmount, err := internal.FromTFTtoUSDMillicent(h.substrateClient, tftPendingAmount)
	if err != nil {
		logger.GetLogger().Error().Err(err).Msg("failed to convert tft to usd millicent")
		InternalServerError(c)
		return
	}

	userResponse := GetUserResponse{
		User:              user,
		PendingBalanceUSD: internal.FromUSDMilliCentToUSD(usdMillicentPendingAmount),
	}

	Success(c, http.StatusOK, "User is retrieved successfully", gin.H{
		"user": userResponse,
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
		logger.GetLogger().Error().Err(err).Send()
		Error(c, http.StatusNotFound, "User is not found", "")
		return
	}

	usdMillicentBalance, err := internal.GetUserBalanceUSDMillicent(h.substrateClient, user.Mnemonic)
	if err != nil {
		logger.GetLogger().Error().Err(err).Send()
		InternalServerError(c)
		return
	}

	pendingRecords, err := h.db.ListUserPendingRecords(userID)
	if err != nil {
		logger.GetLogger().Error().Err(err).Msg("failed to list pending records")
		InternalServerError(c)
		return
	}

	var tftPendingAmount uint64
	for _, record := range pendingRecords {
		tftPendingAmount += record.TFTAmount - record.TransferredTFTAmount
	}

	usdPendingAmount, err := internal.FromTFTtoUSDMillicent(h.substrateClient, tftPendingAmount)
	if err != nil {
		logger.GetLogger().Error().Err(err).Msg("failed to convert tft to usd millicent")
		InternalServerError(c)
		return
	}

	Success(c, http.StatusOK, "Balance is fetched", UserBalanceResponse{
		BalanceUSD:        internal.FromUSDMilliCentToUSD(usdMillicentBalance),
		DebtUSD:           internal.FromUSDMilliCentToUSD(user.Debt),
		PendingBalanceUSD: internal.FromUSDMilliCentToUSD(usdPendingAmount),
	})
}

// @Summary Redeem voucher
// @Description Redeems a voucher for the user
// @Tags users
// @ID redeem-voucher
// @Param voucher_code path string true "Voucher Code"
// @Produce json
// @Success 202 {object} RedeemVoucherResponse "workflow_id: string, voucher_code: string, amount: float64, email: string"
// @Failure 400 {object} APIResponse "Invalid voucher code, already redeemed, or expired"
// @Failure 404 {object} APIResponse "User or voucher are not found"
// @Failure 500 {object} APIResponse "Internal server error"
// @Router /user/redeem/{voucher_code} [put]
func (h *Handler) RedeemVoucherHandler(c *gin.Context) {
	voucherCodeParam := c.Param("voucher_code")
	if voucherCodeParam == "" {
		Error(c, http.StatusBadRequest, "Voucher Code is required", "")
		return
	}
	userID := c.GetInt("user_id")

	user, err := h.db.GetUserByID(userID)
	if err != nil {
		logger.GetLogger().Error().Err(err).Send()
		Error(c, http.StatusNotFound, "User is not found", "")
		return
	}

	// check voucher exists
	voucher, err := h.db.GetVoucherByCode(voucherCodeParam)
	if err != nil {
		logger.GetLogger().Error().Err(err).Send()
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
		logger.GetLogger().Error().Err(err).Send()
		InternalServerError(c)
		return
	}

	wf, err := h.ewfEngine.NewWorkflow(constants.WorkflowRedeemVoucher)
	if err != nil {
		logger.GetLogger().Error().Err(err).Send()
		InternalServerError(c)
		return
	}
	wf.State = map[string]interface{}{
		"user_id":       user.ID,
		"amount":        internal.FromUSDToUSDMillicent(voucher.Value),
		"mnemonic":      user.Mnemonic,
		"username":      user.Username,
		"transfer_mode": models.RedeemVoucherMode,
	}
	h.ewfEngine.RunAsync(context.Background(), wf)

	Success(c, http.StatusOK, "Voucher is redeemed successfully. Money transfer in progress.", RedeemVoucherResponse{
		WorkflowID:  wf.UUID,
		VoucherCode: voucher.Code,
		Amount:      voucher.Value,
		Email:       user.Email,
	})
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
		logger.GetLogger().Error().Err(err).Msg("failed to list SSH keys")
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
		logger.GetLogger().Error().Err(err).Send()
		Error(c, http.StatusBadRequest, "Invalid request format", err.Error())
		return
	}

	if err := internal.ValidateStruct(request); err != nil {
		Error(c, http.StatusBadRequest, "Validation failed", err.Error())
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
		var sqliteErr sqlite3.Error
		if errors.As(err, &sqliteErr) && sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
			Error(c, http.StatusBadRequest, "Duplicate SSH key", "SSH key name or public key already exists for this user.")
			return
		}
		logger.GetLogger().Error().Err(err).Msg("failed to create SSH key")
		InternalServerError(c)
		return
	}

	payload := notification.CommonPayload{
		Status:  "ssh_key_added",
		Subject: "New SSH key added",
		Message: fmt.Sprintf("SSH key '%s' was added to your account.", sshKey.Name),
	}
	notification := models.NewNotification(userID, models.NotificationTypeUser, notification.MergePayload(payload, map[string]string{}))
	err := h.notificationService.Send(c, notification)
	if err != nil {
		logger.GetLogger().Error().Err(err).Msg("failed to send ssh key added notification")
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
		logger.GetLogger().Error().Err(err).Msg("failed to delete SSH key")
		InternalServerError(c)
		return
	}

	Success(c, http.StatusOK, "SSH key deleted successfully", nil)
}

// @Summary Get workflow status
// @Description Returns the status of a workflow by its ID.
// @Tags workflow
// @ID get-workflow-status
// @Accept json
// @Produce json
// @Param workflow_id path string true "Workflow ID"
// @Success 200 {object} string "Workflow status returned successfully"
// @Failure 400 {object} APIResponse "Invalid request or missing workflow ID"
// @Failure 404 {object} APIResponse "Workflow not found"
// @Failure 500 {object} APIResponse "Internal server error"
// @Router /workflow/{workflow_id} [get]
func (h *Handler) GetWorkflowStatus(c *gin.Context) {

	workflowID := c.Param("workflow_id")
	if workflowID == "" {
		Error(c, http.StatusBadRequest, "Invalid request", "Workflow ID is required")
		return
	}

	workflow, err := h.ewfEngine.Store().LoadWorkflowByUUID(c, workflowID)
	if err != nil {
		InternalServerError(c)
		return
	}
	Success(c, http.StatusOK, "Status returned successfully", workflow.Status)
}

// @Summary List user pending records
// @Description Returns user pending records in the system
// @Tags users
// @ID list-user-pending-records
// @Accept json
// @Produce json
// @Success 200 {array} PendingRecordsResponse
// @Failure 500 {object} APIResponse
// @Security BearerAuth
// @Router /user/pending-records [get]
// ListUserPendingRecordsHandler returns user pending records in the system
func (h *Handler) ListUserPendingRecordsHandler(c *gin.Context) {
	userID := c.GetInt("user_id")

	pendingRecords, err := h.db.ListUserPendingRecords(userID)
	if err != nil {
		logger.GetLogger().Error().Err(err).Msg("failed to list pending records")
		InternalServerError(c)
		return
	}

	var pendingRecordsResponse []PendingRecordsResponse
	for _, record := range pendingRecords {
		usdMillicentAmount, err := internal.FromTFTtoUSDMillicent(h.substrateClient, record.TFTAmount)
		if err != nil {
			logger.GetLogger().Error().Err(err).Msg("failed to convert tft to usd amount")
			InternalServerError(c)
			return
		}

		usdMillicentTransferredAmount, err := internal.FromTFTtoUSDMillicent(h.substrateClient, record.TransferredTFTAmount)
		if err != nil {
			logger.GetLogger().Error().Err(err).Msg("failed to convert tft to usd transferred amount")
			InternalServerError(c)
			return
		}

		pendingRecordsResponse = append(pendingRecordsResponse, PendingRecordsResponse{
			PendingRecord:        record,
			USDAmount:            internal.FromUSDMilliCentToUSD(usdMillicentAmount),
			TransferredUSDAmount: internal.FromUSDMilliCentToUSD(usdMillicentTransferredAmount),
		})
	}

	Success(c, http.StatusOK, "Pending records are retrieved successfully", map[string]any{
		"pending_records": pendingRecordsResponse,
	})
}

func isUserRegistered(user models.User) bool {
	return user.Sponsored &&
		user.Verified &&
		len(strings.TrimSpace(user.AccountAddress)) > 0 &&
		len(strings.TrimSpace(user.StripeCustomerID)) > 0 &&
		len(strings.TrimSpace(user.Mnemonic)) > 0
}
