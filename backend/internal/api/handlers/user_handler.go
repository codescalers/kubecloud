package handlers

import (
	"errors"
	"fmt"
	"kubecloud/internal/auth"
	"kubecloud/internal/billing"
	"kubecloud/internal/core/models"
	"kubecloud/internal/core/services"
	"kubecloud/internal/core/workflows"
	"kubecloud/internal/infrastructure/mailservice"
	"kubecloud/internal/infrastructure/notification"
	"strconv"
	"strings"
	"time"

	"kubecloud/internal/infrastructure/logger"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/paymentmethod"
	"golang.org/x/crypto/ssh"
)

type UserHandler struct {
	svc                    services.UserService
	notificationDispatcher *notification.NotificationDispatcher
	mailService            mailservice.MailService
	tokenManager           auth.TokenManager
	stripeClient           billing.StripeClient
}

func NewUserHandler(
	svc services.UserService,
	notificationDispatcher *notification.NotificationDispatcher,
	mailService mailservice.MailService,
	tokenManager auth.TokenManager,
	stripeClient billing.StripeClient,
) UserHandler {
	return UserHandler{
		svc:                    svc,
		notificationDispatcher: notificationDispatcher,
		mailService:            mailService,
		tokenManager:           tokenManager,
		stripeClient:           stripeClient,
	}
}

// RegisterInput struct for data needed when user register
type RegisterInput struct {
	Name            string `json:"name" binding:"required,min=3,max=64"`
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
	Code  int    `json:"code" binding:"required"`
}

// ChangePasswordInput struct for user to change password
type ChangePasswordInput struct {
	Email           string `json:"email" binding:"required,email"`
	Password        string `json:"password" binding:"required,min=8,max=64"`
	ConfirmPassword string `json:"confirm_password" binding:"required,eqfield=Password"`
}

// ChargeBalanceInput struct holds required data to charge users' balance
type ChargeBalanceInput struct {
	CardType     string  `json:"card_type" binding:"required"`
	PaymentToken string  `json:"payment_method_id" binding:"required"`
	Amount       float64 `json:"amount" binding:"required,gt=0"`
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
	Name      string `json:"name" binding:"required"`
	PublicKey string `json:"public_key" binding:"required"`
}

// RegisterUserResponse holds the response for user registration
type RegisterUserResponse struct {
	WorkflowID string `json:"workflow_id"`
	Email      string `json:"email"`
}

type VerifyRegisterUserResponse struct {
	WorkflowID string `json:"workflow_id"`
	Email      string `json:"email"`
	*auth.TokenPair
}

// PendingRecordsResponse swagger model
type PendingRecordsResponse struct {
	PendingRecords []services.PendingRecordsWithUSDAmounts `json:"pending_records"`
}

// RedeemVoucherResponse holds the response for redeeming a voucher
type RedeemVoucherResponse struct {
	WorkflowID  string  `json:"workflow_id"`
	VoucherCode string  `json:"voucher_code"`
	Amount      float64 `json:"amount"`
	Email       string  `json:"email"`
}

// UserWorkflow holds the response for listing user workflows
type UserWorkflow struct {
	WorkflowID  string    `json:"workflow_id"`
	Name        string    `json:"name"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	CurrentStep int       `json:"current_step"`
	TotalSteps  int       `json:"total_steps"`
}

// UserWorkflowsResponse swagger model
type UserWorkflowsResponse struct {
	WorkflowResponse []UserWorkflow `json:"workflows"`
}

// RegisterHandler registers user to the system
// @Summary Register user (with KYC sponsorship)
// @Description Registers a new user, sets up blockchain account, and creates KYC sponsorship. Sends verification code to email.
// @Tags users
// @ID register-user
// @Accept json
// @Produce json
// @Param body body RegisterInput true "Register Input"
// @Success 202 {object} APIResponse{data=RegisterUserResponse} "workflow_id: string, email: string"
// @Failure 400 {object} APIResponse "Invalid request format"
// @Failure 409 {object} APIResponse "User is already registered"
// @Failure 500 {object} APIResponse "Internal server error"
// @Router /user/register [post]
func (h *UserHandler) RegisterHandler(c *gin.Context) {
	reqLog := requestLogger(c, "RegisterHandler")
	var request RegisterInput

	// check on request format
	if err := c.ShouldBindJSON(&request); err != nil {
		reqLog.Error().Err(err).Msg("Invalid request format")
		BadRequest(c, "Invalid request format")
		return
	}

	// check if user previously exists
	existingUser, getErr := h.svc.GetUserByEmail(request.Email)
	if getErr != nil && getErr != models.ErrUserNotFound {
		reqLog.Error().Err(getErr).Msg("failed to get user by email")
		InternalServerError(c)
		return
	}

	if getErr != models.ErrUserNotFound {
		if isUserRegistered(existingUser) {
			Conflict(c, "User is already registered")
			return
		}
	}

	wfUUID, err := h.svc.AsyncRegisterUser(request.Name, request.Email, request.Password)
	if err != nil {
		reqLog.Error().Err(err).Msg("failed to register user")
		InternalServerError(c)
		return
	}

	Accepted(c, "Registration in progress. You can check its status using the workflow id.", RegisterUserResponse{
		WorkflowID: wfUUID,
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
// @Success 202 {object} APIResponse{data=VerifyRegisterUserResponse} "workflow_id: string, email: string, token_pair: object"
// @Failure 400 {object} APIResponse "Invalid request"
// @Failure 409 {object} APIResponse "User is already registered"
// @Failure 500 {object} APIResponse "Internal server error"
// @Router /user/register/verify [post]
func (h *UserHandler) VerifyRegisterCode(c *gin.Context) {
	reqLog := requestLogger(c, "VerifyRegisterCode")
	var request VerifyCodeInput

	if err := c.ShouldBindJSON(&request); err != nil {
		reqLog.Error().Err(err).Msg("Invalid request format")
		BadRequest(c, "Invalid request format")
		return
	}

	// get user by email
	user, err := h.svc.GetUserByEmail(request.Email)
	if err != nil {
		if errors.Is(err, models.ErrUserNotFound) {
			NotFound(c, "User not found")
			return
		}
		reqLog.Error().Err(err).Msg("failed to get user by email")
		InternalServerError(c)
		return
	}
	logWithUser := requestLogger(c, "VerifyRegisterCode").With().Int("user_id", user.ID).Logger()
	reqLog = &logWithUser

	// check if user is already registered (all required fields are set)
	if isUserRegistered(user) {
		Conflict(c, "User is already registered")
		return
	}

	// check verification if user is not verified
	if !user.Verified {
		if user.Code != request.Code {
			BadRequest(c, "Invalid verification code")
			return
		}

		if h.svc.IsVerificationCodeExpired(user.UpdatedAt) {
			BadRequest(c, "Code has expired")
			return
		}

		if err := h.svc.UpdateUserByID(&models.User{ID: user.ID, Verified: true}); err != nil {
			if errors.Is(err, models.ErrUserNotFound) {
				NotFound(c, "User not found")
				return
			}
			reqLog.Error().Err(err).Msg("failed to update user data")
			InternalServerError(c)
			return
		}

		notif := notification.UserNotification(user.ID).
			Success("User email is verified successfully").
			WithSubject("User email verified").
			WithChannels(notification.ChannelUI).
			NoPersist().
			Build()

		if err := h.notificationDispatcher.Send(c.Request.Context(), notif); err != nil {
			reqLog.Error().Err(err).Msg("failed to send user verification notification")
			InternalServerError(c)
			return
		}
	}

	wfUUID, err := h.svc.AsyncVerifyUserRegistration(c.Request.Context(), user.ID, user.Email, user.Username)
	if err != nil {
		reqLog.Error().Err(err).Msg("failed to start user verification workflow")
		InternalServerError(c)
		return
	}

	tokenPair, err := h.tokenManager.CreateTokenPair(user.ID, user.Username, user.Admin)
	if err != nil {
		reqLog.Error().Err(err).Msg("Failed to generate token pair")
		InternalServerError(c)
		return
	}

	Accepted(c, "Verification is in progress", VerifyRegisterUserResponse{
		WorkflowID: wfUUID,
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
// @Success 201 {object} APIResponse{data=auth.TokenPair} "token pair generated"
// @Failure 400 {object} APIResponse "Invalid request format"
// @Failure 401 {object} APIResponse "Login failed"
// @Failure 500 {object} APIResponse
// @Router /user/login [post]
// LoginUserHandler logs user into the system
func (h *UserHandler) LoginUserHandler(c *gin.Context) {
	reqLog := requestLogger(c, "LoginUserHandler")
	var request LoginInput

	// check on request format
	if err := c.ShouldBindJSON(&request); err != nil {
		BadRequest(c, "Invalid request format")
		return
	}

	// get user by email
	user, err := h.svc.GetUserByEmail(request.Email)
	if err != nil {
		reqLog.Error().Err(err).Msg("failed to get user by email")
		BadRequest(c, "email or password is incorrect")
		return
	}

	logWithUser := requestLogger(c, "LoginUserHandler").With().Int("user_id", user.ID).Logger()
	reqLog = &logWithUser

	// verify password
	match := auth.VerifyPassword(user.Password, request.Password)
	if !match {
		Unauthorized(c, "email or password is incorrect")
		return
	}

	if err := h.svc.CheckKYCVerification(c.Request.Context(), user.ID, user.Sponsored, user.AccountAddress); err != nil {
		reqLog.Error().Err(err).Msg("failed to check KYC verification status")
		InternalServerError(c)
		return
	}

	// create token pairs
	tokenPair, err := h.tokenManager.CreateTokenPair(user.ID, user.Username, user.Admin)
	if err != nil {
		reqLog.Error().Err(err).Msg("Failed to generate token pair")
		InternalServerError(c)
		return
	}

	Created(c, "token pair generated", tokenPair)
}

// @Summary Refresh access token
// @Description Refreshes the access token using a valid refresh token
// @Tags users
// @ID refresh-token
// @Accept json
// @Produce json
// @Param body body RefreshTokenInput true "Refresh Token Input"
// @Success 201 {object} APIResponse{data=RefreshTokenResponse} "access token refreshed successfully"
// @Failure 400 {object} APIResponse "Invalid request format"
// @Failure 401 {object} APIResponse "Invalid or expired refresh token"
// @Failure 500 {object} APIResponse
// @Router /user/refresh [post]
// RefreshTokenHandler handles token refresh requests
func (h *UserHandler) RefreshTokenHandler(c *gin.Context) {
	reqLog := requestLogger(c, "RefreshTokenHandler")
	var request RefreshTokenInput

	if err := c.ShouldBindJSON(&request); err != nil {
		reqLog.Error().Err(err).Msg("Invalid request format")
		BadRequest(c, "Invalid request format")
		return
	}

	accessToken, err := h.tokenManager.AccessTokenFromRefresh(request.RefreshToken)
	if err != nil {
		reqLog.Error().Err(err).Msg("refresh token failed")
		Unauthorized(c, "Invalid or expired refresh token")
		return
	}

	Created(c, "access token refreshed successfully", RefreshTokenResponse{
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
// @Success 200 {object} APIResponse{data=RegisterResponse} "Verification code sent successfully"
// @Failure 400 {object} APIResponse "Invalid request format"
// @Failure 404 {object} APIResponse "User is not found"
// @Failure 500 {object} APIResponse
// @Router /user/forgot_password [post]
// ForgotPasswordHandler sends user verification code
func (h *UserHandler) ForgotPasswordHandler(c *gin.Context) {
	reqLog := requestLogger(c, "ForgotPasswordHandler")
	var request EmailInput

	if err := c.ShouldBindJSON(&request); err != nil {
		reqLog.Error().Err(err).Msg("Invalid request format")
		BadRequest(c, "Invalid request format")
		return
	}

	// get user by email
	user, err := h.svc.GetUserByEmail(request.Email)
	if err != nil {
		reqLog.Error().Err(err).Msg("failed to get user ")
		NotFound(c, "user lookup failed")
		return
	}

	logWithUser := requestLogger(c, "ForgotPasswordHandler").With().Int("user_id", user.ID).Logger()
	reqLog = &logWithUser

	code := h.svc.GenerateRandomCode()

	subject, body := h.mailService.ResetPasswordMailContent(code, h.svc.CodeTimeoutInMinutes(), user.Username)
	err = h.mailService.SendMailFromSystem(request.Email, subject, body)
	if err != nil {
		reqLog.Error().Err(err).Msg("failed to send verification code")
		InternalServerError(c)
		return
	}

	if err = h.svc.UpdateUserByID(
		&models.User{
			ID:        user.ID,
			UpdatedAt: time.Now(),
			Code:      code,
		},
	); err != nil {
		if errors.Is(err, models.ErrUserNotFound) {
			NotFound(c, "User not found")
			return
		}
		reqLog.Error().Err(err).Msg("error updating user data")
		InternalServerError(c)
		return
	}

	OK(c, "Verification code sent", RegisterResponse{
		Email:   request.Email,
		Timeout: fmt.Sprintf("%d minutes", h.svc.CodeTimeoutInMinutes()),
	})
}

// @Summary Verify forgot password code
// @Description Verifies the code sent to the user's email for password reset
// @Tags users
// @ID verify-forgot-password-code
// @Accept json
// @Produce json
// @Param body body VerifyCodeInput true "Verify Code Input"
// @Success 201 {object} APIResponse{data=auth.TokenPair} "Verification successful"
// @Failure 400 {object} APIResponse "Invalid request format or verification failed"
// @Failure 500 {object} APIResponse
// @Router /user/forgot_password/verify [post]
// VerifyForgetPasswordCodeHandler verifies code sent to user when forgetting password
func (h *UserHandler) VerifyForgetPasswordCodeHandler(c *gin.Context) {
	reqLog := requestLogger(c, "VerifyForgetPasswordCodeHandler")
	var request VerifyCodeInput
	if err := c.ShouldBindJSON(&request); err != nil {
		reqLog.Error().Err(err).Msg("Invalid request format")
		BadRequest(c, "Invalid request format")
		return
	}

	// get user by email
	user, err := h.svc.GetUserByEmail(request.Email)
	if err != nil {
		if err == models.ErrUserNotFound {
			NotFound(c, "User not found")
			return
		}

		reqLog.Error().Err(err).Msg("failed to get user by email")
		InternalServerError(c)
		return
	}

	logWithUser := requestLogger(c, "VerifyForgetPasswordCodeHandler").With().Int("user_id", user.ID).Logger()
	reqLog = &logWithUser

	if user.Code != request.Code {
		BadRequest(c, "Invalid code")
		return
	}

	if h.svc.IsVerificationCodeExpired(user.UpdatedAt) {
		BadRequest(c, "verification code has expired")
		return
	}

	isAdmin := h.svc.IsSystemAdmin(user.Email)

	// create token pairs
	tokenPair, err := h.tokenManager.CreateTokenPair(user.ID, user.Username, isAdmin)
	if err != nil {
		reqLog.Error().Err(err).Msg("Failed to generate token pair")
		InternalServerError(c)
		return
	}

	Created(c, "Verification successful", tokenPair)
}

// @Summary Change password
// @Description Changes the user's password
// @Tags users
// @ID change-password
// @Accept json
// @Produce json
// @Param body body ChangePasswordInput true "Change Password Input"
// @Success 200 {object} APIResponse "Password updated successfully"
// @Failure 400 {object} APIResponse "Invalid request format or password mismatch"
// @Failure 404 {object} APIResponse "User is not found"
// @Failure 500 {object} APIResponse
// @Router /user/change_password [put]
// ChangePasswordHandler changes password of user
func (h *UserHandler) ChangePasswordHandler(c *gin.Context) {
	var request ChangePasswordInput
	userID := c.GetInt("user_id")
	reqLog := requestLogger(c, "ChangePasswordHandler")

	if err := c.ShouldBindJSON(&request); err != nil {
		reqLog.Error().Err(err).Msg("Invalid request format")
		BadRequest(c, "Invalid request format")
		return
	}

	// hash password
	hashedPassword, err := auth.HashAndSaltPassword([]byte(request.Password))
	if err != nil {
		reqLog.Error().Err(err).Msg("error hashing password")
		InternalServerError(c)
		return
	}

	if err = h.svc.UpdateUserByID(&models.User{ID: userID, Password: hashedPassword}); err != nil {
		if errors.Is(err, models.ErrUserNotFound) {
			NotFound(c, "User is not found")
			return
		}

		reqLog.Error().Err(err).Msg("failed to update password")
		InternalServerError(c)
		return
	}

	notif := notification.UserNotification(userID).
		Success("Your account password has been successfully updated.").
		WithSubject("Your password was changed").
		WithStatus("password_changed").
		Build()

	if err := h.notificationDispatcher.Send(c.Request.Context(), notif); err != nil {
		reqLog.Error().Err(err).Msg("failed to send password changed notification")
		InternalServerError(c)
		return
	}

	OK(c, "Password is updated successfully", nil)
}

// @Summary Charge user balance
// @Description Charges the user's balance using a payment method
// @Tags users
// @ID charge-balance
// @Accept json
// @Produce json
// @Param body body ChargeBalanceInput true "Charge Balance Input"
// @Success 202 {object} APIResponse{data=ChargeBalanceResponse} "workflow_id: string, email: string"
// @Failure 400 {object} APIResponse "Invalid request format or amount"
// @Failure 404 {object} APIResponse "User is not found"
// @Failure 500 {object} APIResponse "Internal server error"
// @Router /user/balance/charge [post]
func (h *UserHandler) ChargeBalance(c *gin.Context) {
	userID := c.GetInt("user_id")
	reqLog := requestLogger(c, "ChargeBalance")

	var request ChargeBalanceInput
	if err := c.ShouldBindJSON(&request); err != nil {
		reqLog.Error().Err(err).Msg("Invalid request format")
		BadRequest(c, "Invalid request format")
		return
	}

	user, err := h.svc.GetUserByID(userID)
	if err != nil {
		if errors.Is(err, models.ErrUserNotFound) {
			NotFound(c, "User is not found")
			return
		}
		reqLog.Error().Err(err).Msg("failed to get user by id")
		InternalServerError(c)
		return
	}

	paymentMethod, err := h.stripeClient.CreatePaymentMethod(request.CardType, request.PaymentToken)
	if err != nil {
		reqLog.Error().Err(err).Msg("error creating payment method")
		h.svc.IncrementStripePaymentFailure()

		if stripeErr, ok := err.(*stripe.Error); ok {
			Error(c, stripeErr.HTTPStatusCode, string(stripeErr.Code))
			return
		}

		InternalServerError(c)
		return
	}

	_, err = paymentmethod.Attach(paymentMethod.ID, &stripe.PaymentMethodAttachParams{
		Customer: stripe.String(user.StripeCustomerID),
	})
	if err != nil {
		reqLog.Error().Err(err).Msg("error attaching payment method to customer")
		h.svc.IncrementStripePaymentFailure()

		if stripeErr, ok := err.(*stripe.Error); ok {
			Error(c, stripeErr.HTTPStatusCode, string(stripeErr.Code))
			return
		}

		InternalServerError(c)
		return
	}

	wfUUID, err := h.svc.AsyncStripeChargeBalance(userID, user.StripeCustomerID, paymentMethod.ID, user.Mnemonic, user.Username, request.Amount)
	if err != nil {
		reqLog.Error().Err(err).Msg("failed to create async stripe charge balance workflow")
		InternalServerError(c)
		return
	}

	Accepted(c, "Charge in progress. You can check its status using the workflow id.", ChargeBalanceResponse{
		WorkflowID: wfUUID,
		Email:      user.Email,
	})
}

// @Summary Get user details
// @Description Retrieves all data of the user
// @Tags users
// @ID get-user
// @Produce json
// @Success 200 {object} APIResponse{data=services.UserWithPendingBalance} "User is retrieved successfully"
// @Failure 404 {object} APIResponse "User is not found"
// @Failure 500 {object} APIResponse
// @Router /user [get]
// GetUserHandler retrieves all data of the user
func (h *UserHandler) GetUserHandler(c *gin.Context) {
	userID := c.GetInt("user_id")
	reqLog := requestLogger(c, "GetUserHandler")

	userWithPendingBalance, err := h.svc.GetUserWithPendingBalance(userID)
	if err != nil {
		if errors.Is(err, models.ErrUserNotFound) {
			NotFound(c, "User is not found")
			return
		}

		reqLog.Error().Err(err).Msg("Failed to get user with pending balance")
		InternalServerError(c)
		return
	}

	OK(c, "User is retrieved successfully", gin.H{
		"user": userWithPendingBalance,
	})
}

// @Summary Get user balance
// @Description Retrieves the user's balance in USD
// @Tags users
// @ID get-user-balance
// @Produce json
// @Success 200 {object} APIResponse{data=UserBalanceResponse} "Balance fetched successfully"
// @Failure 404 {object} APIResponse "User is not found"
// @Failure 500 {object} APIResponse
// @Router /user/balance [get]
// GetUserBalance returns user's balance in usd
func (h *UserHandler) GetUserBalance(c *gin.Context) {
	userID := c.GetInt("user_id")
	reqLog := requestLogger(c, "GetUserBalance")

	user, err := h.svc.GetUserByID(userID)
	if err != nil {
		if errors.Is(err, models.ErrUserNotFound) {
			NotFound(c, "User is not found")
			return
		}
		reqLog.Error().Err(err).Msg("User is not found")
		InternalServerError(c)
		return
	}

	usdMillicentBalance, err := h.svc.GetUserBalanceInUSDMillicent(user.Mnemonic)
	if err != nil {
		reqLog.Error().Err(err).Msg("failed to get user balance in usd millicent")
		InternalServerError(c)
		return
	}

	pendingAmountInUSDMillicent, err := h.svc.GetUserPendingBalanceInUSDMillicent(userID)
	if err != nil {
		reqLog.Error().Err(err).Msg("failed to list pending records")
		InternalServerError(c)
		return
	}

	OK(c, "Balance is fetched", UserBalanceResponse{
		BalanceUSD:        workflows.FromUSDMilliCentToUSD(usdMillicentBalance),
		DebtUSD:           workflows.FromUSDMilliCentToUSD(user.Debt),
		PendingBalanceUSD: workflows.FromUSDMilliCentToUSD(pendingAmountInUSDMillicent),
	})
}

// @Summary Redeem voucher
// @Description Redeems a voucher for the user
// @Tags users
// @ID redeem-voucher
// @Param voucher_code path string true "Voucher Code"
// @Produce json
// @Success 202 {object} APIResponse{data=RedeemVoucherResponse} "workflow_id: string, voucher_code: string, amount: float64, email: string"
// @Failure 400 {object} APIResponse "Invalid voucher code, already redeemed, or expired"
// @Failure 404 {object} APIResponse "User or voucher are not found"
// @Failure 500 {object} APIResponse "Internal server error"
// @Router /user/redeem/{voucher_code} [put]
func (h *UserHandler) RedeemVoucherHandler(c *gin.Context) {
	voucherCodeParam := c.Param("voucher_code")
	if voucherCodeParam == "" {
		BadRequest(c, "Voucher Code is required")
		return
	}
	userID := c.GetInt("user_id")
	reqLog := requestLogger(c, "RedeemVoucherHandler")

	user, err := h.svc.GetUserByID(userID)
	if err != nil {
		if errors.Is(err, models.ErrUserNotFound) {
			NotFound(c, "User is not found")
			return
		}
		reqLog.Error().Err(err).Msg("User is not found")
		InternalServerError(c)
		return
	}

	// check voucher exists
	voucher, err := h.svc.GetVoucherByCode(voucherCodeParam)
	if err != nil {
		if errors.Is(err, models.ErrVoucherNotFound) {
			NotFound(c, "Voucher is not found")
			return
		}
		reqLog.Error().Err(err).Msg("Voucher is not found")
		InternalServerError(c)
		return
	}

	// check voucher not redeemed
	if voucher.Redeemed {
		BadRequest(c, "Voucher is already redeemed")
		return
	}

	// check on expiration time of voucher
	if voucher.ExpiresAt.Before(time.Now()) {
		BadRequest(c, "Voucher is already expired")
		return
	}

	wfUUID, err := h.svc.AsyncRedeemVoucher(user.ID, voucher.Value, user.Mnemonic, user.Username, voucher.Code)
	if err != nil {
		reqLog.Error().Err(err).Msg("failed to redeem voucher")
		InternalServerError(c)
		return
	}

	Accepted(c, "Voucher is redeemed successfully. Money transfer in progress.", RedeemVoucherResponse{
		WorkflowID:  wfUUID,
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
// @Success 200 {object} APIResponse{data=[]models.SSHKey}
// @Failure 401 {object} APIResponse "Unauthorized"
// @Failure 500 {object} APIResponse
// @Router /user/ssh-keys [get]
// ListSSHKeysHandler lists all SSH keys for the authenticated user
func (h *UserHandler) ListSSHKeysHandler(c *gin.Context) {
	userID := c.GetInt("user_id")
	reqLog := requestLogger(c, "ListSSHKeysHandler")
	if userID == 0 {
		Unauthorized(c, "user not authenticated")
		return
	}

	sshKeys, err := h.svc.ListUserSSHKeys(userID)
	if err != nil {
		reqLog.Error().Err(err).Msg("failed to list SSH keys")
		InternalServerError(c)
		return
	}

	OK(c, "SSH keys retrieved successfully", sshKeys)
}

// @Summary Add SSH key
// @Description Adds a new SSH key for the authenticated user
// @Tags users
// @ID add-ssh-key
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body SSHKeyInput true "SSH Key Input"
// @Success 201 {object} APIResponse{data=models.SSHKey}
// @Failure 400 {object} APIResponse "Invalid request format"
// @Failure 401 {object} APIResponse "Unauthorized"
// @Failure 500 {object} APIResponse
// @Router /user/ssh-keys [post]
// AddSSHKeyHandler adds a new SSH key for the authenticated user
func (h *UserHandler) AddSSHKeyHandler(c *gin.Context) {
	userID := c.GetInt("user_id")
	reqLog := requestLogger(c, "AddSSHKeyHandler")
	if userID == 0 {
		Unauthorized(c, "user not authenticated")
		return
	}

	var request SSHKeyInput
	if err := c.ShouldBindJSON(&request); err != nil {
		reqLog.Error().Err(err).Msg("Invalid request format")
		BadRequest(c, "Invalid request format")
		return
	}

	// Validate SSH key format
	if _, _, _, _, err := ssh.ParseAuthorizedKey([]byte(request.PublicKey)); err != nil {
		reqLog.Error().Err(err).Msg("Invalid SSH key format")
		BadRequest(c, "Invalid SSH key format")
		return
	}

	sshKey, err := h.svc.CreateSSHKey(userID, request.Name, request.PublicKey)
	if err != nil {
		if errors.Is(err, models.ErrSSHKeyAlreadyExists) {
			BadRequest(c, "SSH key name or public key already exists for this user.")
			return
		}

		reqLog.Error().Err(err).Msg("failed to create SSH key")
		InternalServerError(c)
		return
	}

	notif := notification.UserNotification(userID).
		Success(fmt.Sprintf("SSH key '%s' was added to your account.", sshKey.Name)).
		WithSubject("New SSH key added").
		WithStatus("ssh_key_added").
		Build()

	if err := h.notificationDispatcher.Send(c.Request.Context(), notif); err != nil {
		reqLog.Error().Err(err).Msg("failed to send notification")
		InternalServerError(c)
		return
	}

	Created(c, "SSH key added successfully", sshKey)
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
func (h *UserHandler) DeleteSSHKeyHandler(c *gin.Context) {
	userID := c.GetInt("user_id")
	reqLog := requestLogger(c, "DeleteSSHKeyHandler")
	if userID == 0 {
		Unauthorized(c, "user not authenticated")
		return
	}

	sshKeyID := c.Param("ssh_key_id")
	if sshKeyID == "" {
		BadRequest(c, "SSH key ID is required")
		return
	}

	// Convert sshKeyID to int
	var keyID int
	keyID, err := strconv.Atoi(sshKeyID)
	if err != nil {
		BadRequest(c, "invalid SSH key ID format")
		return
	}

	sshKeyName, err := h.svc.DeleteSSHKey(userID, keyID)
	if err != nil {
		if err.Error() == fmt.Sprintf("no SSH key found with ID %d for user %d", keyID, userID) {
			NotFound(c, "SSH key not found")
			return
		}
		reqLog.Error().Err(err).Msg("failed to delete SSH key")
		InternalServerError(c)
		return
	}

	notif := notification.UserNotification(userID).
		Success(fmt.Sprintf("SSH key '%s' was deleted from your account.", sshKeyName)).
		WithSubject("SSH key deleted").
		WithStatus("ssh_key_deleted").
		Build()

	if err := h.notificationDispatcher.Send(c.Request.Context(), notif); err != nil {
		reqLog.Error().Err(err).Msg("failed to send ssh key deleted notification")
	}

	OK(c, "SSH key deleted successfully", nil)
}

// @Summary Get workflow status
// @Description Returns the status of a workflow by its ID.
// @Tags workflow
// @ID get-workflow-status
// @Accept json
// @Produce json
// @Param workflow_id path string true "Workflow ID"
// @Success 200 {object} APIResponse{data=string} "Workflow status returned successfully"
// @Failure 400 {object} APIResponse "Invalid request or missing workflow ID"
// @Failure 404 {object} APIResponse "Workflow not found"
// @Failure 500 {object} APIResponse "Internal server error"
// @Router /workflow/{workflow_id} [get]
func (h *UserHandler) GetWorkflowStatus(c *gin.Context) {
	reqLog := requestLogger(c, "GetWorkflowStatus")

	workflowID := c.Param("workflow_id")
	if workflowID == "" {
		BadRequest(c, "Workflow ID is required")
		return
	}

	workflowStatus, err := h.svc.GetWorkflowStatus(c.Request.Context(), workflowID)
	if err != nil {
		reqLog.Error().Err(err).Msg("failed to load workflow by UUID")
		InternalServerError(c)
		return
	}

	OK(c, "Status returned successfully", workflowStatus)
}

// @Summary List user pending records
// @Description Returns user pending records in the system
// @Tags users
// @ID list-user-pending-records
// @Accept json
// @Produce json
// @Success 200 {object} APIResponse{data=PendingRecordsResponse} "Pending records returned successfully"
// @Failure 500 {object} APIResponse
// @Security BearerAuth
// @Router /user/pending-records [get]
// ListUserPendingRecordsHandler returns user pending records in the system
func (h *UserHandler) ListUserPendingRecordsHandler(c *gin.Context) {
	userID := c.GetInt("user_id")
	reqLog := requestLogger(c, "ListUserPendingRecordsHandler")

	pendingRecordsWithUSDAmounts, err := h.svc.ListUserPendingRecordsWithUSDAmounts(userID)
	if err != nil {
		reqLog.Error().Err(err).Msg("failed to list pending records with usd amounts")
		InternalServerError(c)
		return
	}

	OK(c, "Pending records are retrieved successfully", gin.H{
		"pending_records": pendingRecordsWithUSDAmounts,
	})
}

// @Summary List remaining user workflows
// @Description Returns all pending/running workflows belonging to the authenticated user.
// @Tags workflow
// @ID list-user-workflows
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} APIResponse{data=UserWorkflowsResponse} "User workflows retrieved successfully"
// @Failure 401 {object} APIResponse "Unauthorized user"
// @Failure 500 {object} APIResponse "Internal server error"
// @Router /user/workflows [get]
func (h *UserHandler) ListUserRemainingWorkflowsHandler(c *gin.Context) {
	userID := c.GetInt("user_id")
	reqLog := requestLogger(c, "ListUserRemainingWorkflowsHandler")

	workflows, err := h.svc.ListRemainingWorkflowsByUserID(userID)
	if err != nil {
		reqLog.Error().Err(err).Msg("failed to list user workflows")
		InternalServerError(c)
		return
	}

	var userWorkflowsResponse []UserWorkflow
	for _, workflow := range workflows {

		userWorkflowsResponse = append(userWorkflowsResponse, UserWorkflow{
			WorkflowID:  workflow.UUID,
			Name:        workflow.Name,
			Status:      string(workflow.Status),
			CreatedAt:   workflow.CreatedAt,
			CurrentStep: workflow.CurrentStep,
			TotalSteps:  len(workflow.Steps),
		})
	}

	OK(c, "User workflows retrieved successfully", gin.H{
		"workflows": userWorkflowsResponse,
	})
}

func isUserRegistered(user models.User) bool {
	return user.Sponsored &&
		user.Verified &&
		len(strings.TrimSpace(user.AccountAddress)) > 0 &&
		len(strings.TrimSpace(user.StripeCustomerID)) > 0 &&
		len(strings.TrimSpace(user.Mnemonic)) > 0
}

// getRequestID retrieves the request ID from the gin context
func getRequestID(c *gin.Context) string {
	if requestID, exists := c.Get("request_id"); exists {
		if id, ok := requestID.(string); ok {
			return id
		}
	}
	return ""
}

// requestLogger creates a contextual logger for a request handler.
func requestLogger(c *gin.Context, handlerName string) *zerolog.Logger {
	requestID := getRequestID(c)
	userID := c.GetInt("user_id")
	return logger.ForRequest(userID, requestID, handlerName)
}
