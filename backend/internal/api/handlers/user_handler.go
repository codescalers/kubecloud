package handlers

import (
	"errors"
	"fmt"
	"kubecloud/internal/auth"
	"kubecloud/internal/billing"
	"kubecloud/internal/core/models"
	"kubecloud/internal/core/services"
	"kubecloud/internal/infrastructure/mailservice"
	"kubecloud/internal/infrastructure/notification"
	"kubecloud/internal/infrastructure/substrate"
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
		auditLogFromContext(
			c,
			logger.AuditActionUserRegister,
			logger.AuditSeverityWarning,
			map[string]any{
				"reason": "invalid_request_format",
			},
		)
		BadRequest(c, "Invalid request format")
		return
	}

	// check if user previously exists
	existingUser, getErr := h.svc.GetUserByEmail(request.Email)
	if getErr != nil && getErr != models.ErrUserNotFound {
		reqLog.Error().Err(getErr).Msg("failed to get user by email")
		auditLogFromContext(
			c,
			logger.AuditActionUserRegister,
			logger.AuditSeverityError,
			map[string]any{
				"reason": getErr.Error(),
			},
		)
		InternalServerError(c)
		return
	}

	if getErr != models.ErrUserNotFound {
		if isUserRegistered(existingUser) {
			auditLogFromContext(
				c,
				logger.AuditActionUserRegister,
				logger.AuditSeverityWarning,
				map[string]any{
					"reason": "user_already_registered",
				},
			)
			Conflict(c, "User is already registered")
			return
		}
	}

	wfUUID, err := h.svc.AsyncRegisterUser(request.Name, request.Email, request.Password)
	if err != nil {
		reqLog.Error().Err(err).Msg("failed to register user")
		auditLogFromContext(
			c,
			logger.AuditActionUserRegister,
			logger.AuditSeverityError,
			map[string]any{
				"reason": err.Error(),
			},
		)
		InternalServerError(c)
		return
	}
	response := RegisterUserResponse{
		WorkflowID: wfUUID,
		Email:      request.Email,
	}
	auditLogFromContext(
		c,
		logger.AuditActionUserRegister,
		logger.AuditSeverityInfo,
		map[string]any{
			"workflow_id": response.WorkflowID,
			"email":       response.Email,
		},
	)

	Accepted(c, "Registration in progress. You can check its status using the workflow id.", response)
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
		auditLogFromContext(
			c,
			logger.AuditActionUserVerify,
			logger.AuditSeverityWarning,
			map[string]any{
				"reason": "invalid_request_format",
			},
		)
		BadRequest(c, "Invalid request format")
		return
	}

	// get user by email
	user, err := h.svc.GetUserByEmail(request.Email)
	if err != nil {
		if errors.Is(err, models.ErrUserNotFound) {
			auditLogFromContext(
				c,
				logger.AuditActionUserVerify,
				logger.AuditSeverityWarning,
				map[string]any{
					"reason": "user_not_found",
				},
			)
			NotFound(c, "User not found")
			return
		}
		reqLog.Error().Err(err).Msg("failed to get user by email")
		auditLogFromContext(
			c,
			logger.AuditActionUserVerify,
			logger.AuditSeverityError,
			map[string]any{
				"reason": err.Error(),
			},
		)
		InternalServerError(c)
		return
	}
	logWithUser := requestLogger(c, "VerifyRegisterCode").With().Int("user_id", user.ID).Logger()
	reqLog = &logWithUser

	// check if user is already registered (all required fields are set)
	if isUserRegistered(user) {
		auditLogFromContext(
			c,
			logger.AuditActionUserVerify,
			logger.AuditSeverityWarning,
			map[string]any{
				"reason": "user_already_registered",
			},
		)
		Conflict(c, "User is already registered")
		return
	}

	// check verification if user is not verified
	if !user.Verified {
		if user.Code != request.Code {
			auditLogFromContext(
				c,
				logger.AuditActionUserVerify,
				logger.AuditSeverityWarning,
				map[string]any{
					"reason": "invalid_code",
				},
			)
			BadRequest(c, "Invalid verification code")
			return
		}

		if h.svc.IsVerificationCodeExpired(user.UpdatedAt) {
			auditLogFromContext(
				c,
				logger.AuditActionUserVerify,
				logger.AuditSeverityWarning,
				map[string]any{
					"reason": "code_expired",
				},
			)
			BadRequest(c, "Code has expired")
			return
		}

		if err := h.svc.UpdateUserByID(&models.User{ID: user.ID, Verified: true}); err != nil {
			if errors.Is(err, models.ErrUserNotFound) {
				auditLogFromContext(
					c,
					logger.AuditActionUserVerify,
					logger.AuditSeverityWarning,
					map[string]any{
						"reason": "user_not_found_on_update",
					},
				)
				NotFound(c, "User not found")
				return
			}
			reqLog.Error().Err(err).Msg("failed to update user data")
			auditLogFromContext(
				c,
				logger.AuditActionUserVerify,
				logger.AuditSeverityError,
				map[string]any{
					"reason": err.Error(),
				},
			)
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
			auditLogFromContext(
				c,
				logger.AuditActionUserVerify,
				logger.AuditSeverityError,
				map[string]any{
					"reason": err.Error(),
				},
			)
			InternalServerError(c)
			return
		}
	}

	wfUUID, err := h.svc.AsyncVerifyUserRegistration(c.Request.Context(), user.ID, user.Email, user.Username)
	if err != nil {
		reqLog.Error().Err(err).Msg("failed to start user verification workflow")
		auditLogFromContext(
			c,
			logger.AuditActionUserVerify,
			logger.AuditSeverityError,
			map[string]any{
				"reason": err.Error(),
			},
		)
		InternalServerError(c)
		return
	}

	tokenPair, err := h.tokenManager.CreateTokenPair(user.ID, user.Username, user.Admin)
	if err != nil {
		reqLog.Error().Err(err).Msg("Failed to generate token pair")
		auditLogFromContext(
			c,
			logger.AuditActionUserVerify,
			logger.AuditSeverityError,
			map[string]any{
				"reason": err.Error(),
			},
		)
		InternalServerError(c)
		return
	}

	response := VerifyRegisterUserResponse{
		WorkflowID: wfUUID,
		Email:      user.Email,
		TokenPair:  tokenPair,
	}
	auditLogWithActor(
		c,
		logger.AuditActionUserVerify,
		logger.AuditSeverityInfo,
		map[string]any{
			"workflow_id": response.WorkflowID,
			"email":       response.Email,
		},
		logger.AuditActorUser,
		map[string]any{
			"user_id": user.ID,
			"email":   user.Email,
		},
	)
	Accepted(c, "Verification is in progress", response)

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
		auditLogFromContext(
			c,
			logger.AuditActionUserLogin,
			logger.AuditSeverityWarning,
			map[string]any{
				"reason": "invalid_request_format",
			},
		)
		BadRequest(c, "Invalid request format")
		return
	}

	// get user by email
	user, err := h.svc.GetUserByEmail(request.Email)
	if err != nil {
		reqLog.Error().Err(err).Msg("failed to get user by email")
		auditLogFromContext(
			c,
			logger.AuditActionUserLogin,
			logger.AuditSeverityWarning,
			map[string]any{
				"reason": "user_not_found",
			},
		)
		BadRequest(c, "email or password is incorrect")
		return
	}

	logWithUser := requestLogger(c, "LoginUserHandler").With().Int("user_id", user.ID).Logger()
	reqLog = &logWithUser

	// verify password
	match := auth.VerifyPassword(user.Password, request.Password)
	if !match {
		auditLogFromContext(
			c,
			logger.AuditActionUserLogin,
			logger.AuditSeverityWarning,
			map[string]any{
				"reason": "invalid_credentials",
			},
		)
		Unauthorized(c, "email or password is incorrect")
		return
	}

	if err := h.svc.CheckKYCVerification(c.Request.Context(), user.ID, user.Sponsored, user.AccountAddress); err != nil {
		reqLog.Error().Err(err).Msg("failed to check KYC verification status")
		auditLogFromContext(
			c,
			logger.AuditActionUserLogin,
			logger.AuditSeverityError,
			map[string]any{
				"reason": err.Error(),
			},
		)
		InternalServerError(c)
		return
	}

	// create token pairs
	tokenPair, err := h.tokenManager.CreateTokenPair(user.ID, user.Username, user.Admin)
	if err != nil {
		reqLog.Error().Err(err).Msg("Failed to generate token pair")
		auditLogFromContext(
			c,
			logger.AuditActionUserLogin,
			logger.AuditSeverityError,
			map[string]any{
				"reason": err.Error(),
			},
		)
		InternalServerError(c)
		return
	}

	auditLogWithActor(
		c,
		logger.AuditActionUserLogin,
		logger.AuditSeverityInfo,
		map[string]any{
			"result": "success",
		},
		logger.AuditActorUser,
		map[string]any{
			"user_id": user.ID,
			"email":   user.Email,
		},
	)
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
		auditLogFromContext(
			c,
			logger.AuditActionUserTokenRefresh,
			logger.AuditSeverityWarning,
			map[string]any{
				"reason": "invalid_request_format",
			},
		)
		BadRequest(c, "Invalid request format")
		return
	}

	accessToken, err := h.tokenManager.AccessTokenFromRefresh(request.RefreshToken)
	if err != nil {
		reqLog.Error().Err(err).Msg("refresh token failed")
		auditLogFromContext(
			c,
			logger.AuditActionUserTokenRefresh,
			logger.AuditSeverityWarning,
			map[string]any{
				"reason": "invalid_or_expired_refresh_token",
			},
		)
		Unauthorized(c, "Invalid or expired refresh token")
		return
	}

	auditLogFromContext(
		c,
		logger.AuditActionUserTokenRefresh,
		logger.AuditSeverityInfo,
		map[string]any{
			"result": "success",
		},
	)

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
		auditLogFromContext(
			c,
			logger.AuditActionUserPasswordResetRequest,
			logger.AuditSeverityWarning,
			map[string]any{
				"reason": "invalid_request_format",
			},
		)
		BadRequest(c, "Invalid request format")
		return
	}

	// get user by email
	user, err := h.svc.GetUserByEmail(request.Email)
	if err != nil {
		if err == models.ErrUserNotFound {
			auditLogFromContext(
				c,
				logger.AuditActionUserPasswordResetRequest,
				logger.AuditSeverityWarning,
				map[string]any{
					"reason": "user_not_found",
				},
			)
			NotFound(c, "user lookup failed")
			return
		}
		reqLog.Error().Err(err).Msg("failed to get user ")

		auditLogFromContext(
			c,
			logger.AuditActionUserPasswordResetRequest,
			logger.AuditSeverityError,
			map[string]any{
				"reason": err.Error(),
			},
		)
		InternalServerError(c)
		return
	}

	logWithUser := requestLogger(c, "ForgotPasswordHandler").With().Int("user_id", user.ID).Logger()
	reqLog = &logWithUser

	code := h.svc.GenerateRandomCode()

	subject, body := h.mailService.ResetPasswordMailContent(code, h.svc.CodeTimeoutInMinutes(), user.Username)
	err = h.mailService.SendMailFromSystem(request.Email, subject, body)
	if err != nil {
		reqLog.Error().Err(err).Msg("failed to send verification code")
		auditLogFromContext(
			c,
			logger.AuditActionUserPasswordResetRequest,
			logger.AuditSeverityError,
			map[string]any{
				"reason": err.Error(),
			},
		)
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
			auditLogFromContext(
				c,
				logger.AuditActionUserPasswordResetRequest,
				logger.AuditSeverityWarning,
				map[string]any{
					"reason": "user_not_found_on_update",
				},
			)
			NotFound(c, "User not found")
			return
		}
		reqLog.Error().Err(err).Msg("error updating user data")
		auditLogFromContext(
			c,
			logger.AuditActionUserPasswordResetRequest,
			logger.AuditSeverityError,
			map[string]any{
				"reason": err.Error(),
			},
		)
		InternalServerError(c)
		return
	}

	response := RegisterResponse{
		Email:   request.Email,
		Timeout: fmt.Sprintf("%d minutes", h.svc.CodeTimeoutInMinutes()),
	}

	auditLogFromContext(
		c,
		logger.AuditActionUserPasswordResetRequest,
		logger.AuditSeverityInfo,
		map[string]any{
			"email": response.Email,
		},
	)
	OK(c, "Verification code sent", response)

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
		auditLogFromContext(
			c,
			logger.AuditActionUserPasswordResetVerify,
			logger.AuditSeverityWarning,
			map[string]any{
				"reason": "invalid_request_format",
			},
		)
		BadRequest(c, "Invalid request format")
		return
	}

	// get user by email
	user, err := h.svc.GetUserByEmail(request.Email)
	if err != nil {
		if err == models.ErrUserNotFound {
			auditLogFromContext(
				c,
				logger.AuditActionUserPasswordResetVerify,
				logger.AuditSeverityWarning,
				map[string]any{
					"reason": "user_not_found",
				},
			)
			NotFound(c, "User not found")
			return
		}

		reqLog.Error().Err(err).Msg("failed to get user by email")
		auditLogFromContext(
			c,
			logger.AuditActionUserPasswordResetVerify,
			logger.AuditSeverityError,
			map[string]any{
				"reason": err.Error(),
			},
		)
		InternalServerError(c)
		return
	}

	logWithUser := requestLogger(c, "VerifyForgetPasswordCodeHandler").With().Int("user_id", user.ID).Logger()
	reqLog = &logWithUser

	if user.Code != request.Code {
		auditLogFromContext(
			c,
			logger.AuditActionUserPasswordResetVerify,
			logger.AuditSeverityWarning,
			map[string]any{
				"reason": "invalid_code",
			},
		)
		BadRequest(c, "Invalid code")
		return
	}

	if h.svc.IsVerificationCodeExpired(user.UpdatedAt) {
		auditLogFromContext(
			c,
			logger.AuditActionUserPasswordResetVerify,
			logger.AuditSeverityWarning,
			map[string]any{
				"reason": "code_expired",
			},
		)
		BadRequest(c, "verification code has expired")
		return
	}

	isAdmin := h.svc.IsSystemAdmin(user.Email)

	// create token pairs
	tokenPair, err := h.tokenManager.CreateTokenPair(user.ID, user.Username, isAdmin)
	if err != nil {
		reqLog.Error().Err(err).Msg("Failed to generate token pair")
		auditLogFromContext(
			c,
			logger.AuditActionUserPasswordResetVerify,
			logger.AuditSeverityError,
			map[string]any{
				"reason": err.Error(),
			},
		)
		InternalServerError(c)
		return
	}
	auditLogFromContext(
		c,
		logger.AuditActionUserPasswordResetVerify,
		logger.AuditSeverityInfo,
		map[string]any{
			"email": user.Email,
		},
	)
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
		auditLogFromContext(
			c,
			logger.AuditActionUserPasswordChange,
			logger.AuditSeverityWarning,
			map[string]any{
				"reason": "invalid_request_format",
			},
		)
		BadRequest(c, "Invalid request format")
		return
	}

	// hash password
	hashedPassword, err := auth.HashAndSaltPassword([]byte(request.Password))
	if err != nil {
		reqLog.Error().Err(err).Msg("error hashing password")
		auditLogFromContext(
			c,
			logger.AuditActionUserPasswordChange,
			logger.AuditSeverityError,
			map[string]any{
				"reason": err.Error(),
			},
		)
		InternalServerError(c)
		return
	}

	if err = h.svc.UpdateUserByID(&models.User{ID: userID, Password: hashedPassword}); err != nil {
		if errors.Is(err, models.ErrUserNotFound) {
			auditLogFromContext(
				c,
				logger.AuditActionUserPasswordChange,
				logger.AuditSeverityWarning,
				map[string]any{
					"reason": "user_not_found",
				},
			)
			NotFound(c, "User is not found")
			return
		}

		reqLog.Error().Err(err).Msg("failed to update password")
		auditLogFromContext(
			c,
			logger.AuditActionUserPasswordChange,
			logger.AuditSeverityError,
			map[string]any{
				"reason": err.Error(),
			},
		)
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
		auditLogFromContext(
			c,
			logger.AuditActionUserPasswordChange,
			logger.AuditSeverityError,
			map[string]any{
				"reason": err.Error(),
			},
		)
		InternalServerError(c)
		return
	}
	auditLogFromContext(
		c,
		logger.AuditActionUserPasswordChange,
		logger.AuditSeverityInfo,
		map[string]any{
			"status": "success",
		},
	)

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
		auditLogFromContext(
			c,
			logger.AuditActionBalanceCharge,
			logger.AuditSeverityWarning,
			map[string]any{
				"reason": "invalid_request_format",
			},
		)
		BadRequest(c, "Invalid request format")
		return
	}

	user, err := h.svc.GetUserByID(userID)
	if err != nil {
		if errors.Is(err, models.ErrUserNotFound) {
			auditLogFromContext(
				c,
				logger.AuditActionBalanceCharge,
				logger.AuditSeverityWarning,
				map[string]any{
					"reason": "user_not_found",
				},
			)
			NotFound(c, "User is not found")
			return
		}
		reqLog.Error().Err(err).Msg("failed to get user by id")
		auditLogFromContext(
			c,
			logger.AuditActionBalanceCharge,
			logger.AuditSeverityError,
			map[string]any{
				"reason": err.Error(),
			},
		)
		InternalServerError(c)
		return
	}

	paymentMethod, err := h.stripeClient.CreatePaymentMethod(request.CardType, request.PaymentToken)
	if err != nil {
		reqLog.Error().Err(err).Msg("error creating payment method")
		h.svc.IncrementStripePaymentFailure()

		if stripeErr, ok := err.(*stripe.Error); ok {
			auditLogFromContext(
				c,
				logger.AuditActionBalanceCharge,
				logger.AuditSeverityWarning,
				map[string]any{
					"reason": stripeErr.Code,
				},
			)
			Error(c, stripeErr.HTTPStatusCode, string(stripeErr.Code))
			return
		}

		auditLogFromContext(
			c,
			logger.AuditActionBalanceCharge,
			logger.AuditSeverityError,
			map[string]any{
				"reason": err.Error(),
			},
		)
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
			auditLogFromContext(
				c,
				logger.AuditActionBalanceCharge,
				logger.AuditSeverityWarning,
				map[string]any{
					"reason": stripeErr.Code,
				},
			)
			Error(c, stripeErr.HTTPStatusCode, string(stripeErr.Code))
			return
		}

		auditLogFromContext(
			c,
			logger.AuditActionBalanceCharge,
			logger.AuditSeverityError,
			map[string]any{
				"reason": err.Error(),
			},
		)
		InternalServerError(c)
		return
	}

	wfUUID, err := h.svc.AsyncStripeChargeBalance(userID, user.StripeCustomerID, paymentMethod.ID, user.Mnemonic, user.Username, request.Amount)
	if err != nil {
		reqLog.Error().Err(err).Msg("failed to create async stripe charge balance workflow")
		auditLogFromContext(
			c,
			logger.AuditActionBalanceCharge,
			logger.AuditSeverityError,
			map[string]any{
				"reason": err.Error(),
			},
		)
		InternalServerError(c)
		return
	}

	response := ChargeBalanceResponse{
		WorkflowID: wfUUID,
		Email:      user.Email,
	}
	auditLogFromContext(
		c,
		logger.AuditActionBalanceCharge,
		logger.AuditSeverityInfo,
		map[string]any{
			"workflow_id": response.WorkflowID,
			"amount":      request.Amount,
			"email":       response.Email,
		},
	)
	Accepted(c, "Charge in progress. You can check its status using the workflow id.", response)

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
			auditLogFromContext(
				c,
				logger.AuditActionBalanceGet,
				logger.AuditSeverityWarning,
				map[string]any{
					"reason": "user_not_found",
				},
			)
			NotFound(c, "User is not found")
			return
		}
		reqLog.Error().Err(err).Msg("User is not found")
		auditLogFromContext(
			c,
			logger.AuditActionBalanceGet,
			logger.AuditSeverityError,
			map[string]any{
				"reason": err.Error(),
			},
		)
		InternalServerError(c)
		return
	}

	usdMillicentBalance, err := h.svc.GetUserBalanceInUSDMillicent(user.Mnemonic)
	if err != nil {
		reqLog.Error().Err(err).Msg("failed to get user balance in usd millicent")
		auditLogFromContext(
			c,
			logger.AuditActionBalanceGet,
			logger.AuditSeverityError,
			map[string]any{
				"reason": err.Error(),
			},
		)
		InternalServerError(c)
		return
	}

	pendingAmountInUSDMillicent, err := h.svc.GetUserPendingBalanceInUSDMillicent(userID)
	if err != nil {
		reqLog.Error().Err(err).Msg("failed to list pending records")
		auditLogFromContext(
			c,
			logger.AuditActionBalanceGet,
			logger.AuditSeverityError,
			map[string]any{
				"reason": err.Error(),
			},
		)
		InternalServerError(c)
		return
	}

	auditLogFromContext(
		c,
		logger.AuditActionBalanceGet,
		logger.AuditSeverityInfo,
		map[string]any{
			"user_id": user.ID,
			"email":   user.Email,
		},
	)
	OK(c, "Balance is fetched", UserBalanceResponse{
		BalanceUSD:        substrate.FromUSDMilliCentToUSD(usdMillicentBalance),
		DebtUSD:           substrate.FromUSDMilliCentToUSD(user.Debt),
		PendingBalanceUSD: substrate.FromUSDMilliCentToUSD(pendingAmountInUSDMillicent),
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
		auditLogFromContext(
			c,
			logger.AuditActionVoucherRedeem,
			logger.AuditSeverityWarning,
			map[string]any{
				"reason": "missing_voucher_code",
			},
		)
		BadRequest(c, "Voucher Code is required")
		return
	}
	userID := c.GetInt("user_id")
	reqLog := requestLogger(c, "RedeemVoucherHandler")

	user, err := h.svc.GetUserByID(userID)
	if err != nil {
		if errors.Is(err, models.ErrUserNotFound) {
			auditLogFromContext(
				c,
				logger.AuditActionVoucherRedeem,
				logger.AuditSeverityWarning,
				map[string]any{
					"reason": "user_not_found",
				},
			)
			NotFound(c, "User is not found")
			return
		}
		reqLog.Error().Err(err).Msg("User is not found")
		auditLogFromContext(
			c,
			logger.AuditActionVoucherRedeem,
			logger.AuditSeverityError,
			map[string]any{
				"reason": err.Error(),
			},
		)
		InternalServerError(c)
		return
	}

	// check voucher exists
	voucher, err := h.svc.GetVoucherByCode(voucherCodeParam)
	if err != nil {
		if errors.Is(err, models.ErrVoucherNotFound) {
			auditLogFromContext(
				c,
				logger.AuditActionVoucherRedeem,
				logger.AuditSeverityWarning,
				map[string]any{
					"reason": "voucher_not_found",
				},
			)
			NotFound(c, "Voucher is not found")
			return
		}
		reqLog.Error().Err(err).Msg("Voucher is not found")
		auditLogFromContext(
			c,
			logger.AuditActionVoucherRedeem,
			logger.AuditSeverityError,
			map[string]any{
				"reason": err.Error(),
			},
		)
		InternalServerError(c)
		return
	}

	// check voucher not redeemed
	if voucher.Redeemed {
		auditLogFromContext(
			c,
			logger.AuditActionVoucherRedeem,
			logger.AuditSeverityWarning,
			map[string]any{
				"reason": "voucher_already_redeemed",
			},
		)
		BadRequest(c, "Voucher is already redeemed")
		return
	}

	// check on expiration time of voucher
	if voucher.ExpiresAt.Before(time.Now()) {
		auditLogFromContext(
			c,
			logger.AuditActionVoucherRedeem,
			logger.AuditSeverityWarning,
			map[string]any{
				"reason": "voucher_expired",
			},
		)
		BadRequest(c, "Voucher is already expired")
		return
	}

	wfUUID, err := h.svc.AsyncRedeemVoucher(user.ID, voucher.Value, user.Mnemonic, user.Username, voucher.Code)
	if err != nil {
		reqLog.Error().Err(err).Msg("failed to redeem voucher")
		auditLogFromContext(
			c,
			logger.AuditActionVoucherRedeem,
			logger.AuditSeverityError,
			map[string]any{
				"reason": err.Error(),
			},
		)
		InternalServerError(c)
		return
	}

	response := RedeemVoucherResponse{
		WorkflowID:  wfUUID,
		VoucherCode: voucher.Code,
		Amount:      voucher.Value,
		Email:       user.Email,
	}
	auditLogFromContext(
		c,
		logger.AuditActionVoucherRedeem,
		logger.AuditSeverityInfo,
		map[string]any{
			"workflow_id": response.WorkflowID,
			"voucher":     response.VoucherCode,
		},
	)
	Accepted(c, "Voucher is redeemed successfully. Money transfer in progress.", response)

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
		auditLogFromContext(
			c,
			logger.AuditActionSSHKeyList,
			logger.AuditSeverityWarning,
			map[string]any{
				"reason": "user_not_authenticated",
			},
		)
		Unauthorized(c, "user not authenticated")
		return
	}

	sshKeys, err := h.svc.ListUserSSHKeys(userID)
	if err != nil {
		reqLog.Error().Err(err).Msg("failed to list SSH keys")
		auditLogFromContext(
			c,
			logger.AuditActionSSHKeyList,
			logger.AuditSeverityError,
			map[string]any{
				"reason": err.Error(),
			},
		)
		InternalServerError(c)
		return
	}
	auditLogFromContext(
		c,
		logger.AuditActionSSHKeyList,
		logger.AuditSeverityInfo,
		nil,
	)

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
		auditLogFromContext(
			c,
			logger.AuditActionSSHKeyAdd,
			logger.AuditSeverityWarning,
			map[string]any{
				"reason": "invalid_request_format",
			},
		)
		BadRequest(c, "Invalid request format")
		return
	}

	// Validate SSH key format
	if _, _, _, _, err := ssh.ParseAuthorizedKey([]byte(request.PublicKey)); err != nil {
		reqLog.Error().Err(err).Msg("Invalid SSH key format")
		auditLogFromContext(
			c,
			logger.AuditActionSSHKeyAdd,
			logger.AuditSeverityWarning,
			map[string]any{
				"reason": "invalid_ssh_key_format",
			},
		)
		BadRequest(c, "Invalid SSH key format")
		return
	}

	sshKey, err := h.svc.CreateSSHKey(userID, request.Name, request.PublicKey)
	if err != nil {
		if errors.Is(err, models.ErrSSHKeyAlreadyExists) {
			auditLogFromContext(
				c,
				logger.AuditActionSSHKeyAdd,
				logger.AuditSeverityWarning,
				map[string]any{
					"reason": "ssh_key_already_exists",
				},
			)
			BadRequest(c, "SSH key name or public key already exists for this user.")
			return
		}

		reqLog.Error().Err(err).Msg("failed to create SSH key")
		auditLogFromContext(
			c,
			logger.AuditActionSSHKeyAdd,
			logger.AuditSeverityError,
			map[string]any{
				"reason": err.Error(),
			},
		)
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
	}

	auditLogFromContext(
		c,
		logger.AuditActionSSHKeyAdd,
		logger.AuditSeverityInfo,
		map[string]any{
			"user_id":      userID,
			"ssh_key_name": sshKey.Name,
		},
	)
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
		auditLogFromContext(
			c,
			logger.AuditActionSSHKeyDelete,
			logger.AuditSeverityWarning,
			map[string]any{
				"reason": "user_not_authenticated",
			},
		)
		Unauthorized(c, "user not authenticated")
		return
	}

	sshKeyID := c.Param("ssh_key_id")
	if sshKeyID == "" {
		auditLogFromContext(
			c,
			logger.AuditActionSSHKeyDelete,
			logger.AuditSeverityWarning,
			map[string]any{
				"reason": "ssh_key_id_required",
			},
		)
		BadRequest(c, "SSH key ID is required")
		return
	}

	// Convert sshKeyID to int
	var keyID int
	keyID, err := strconv.Atoi(sshKeyID)
	if err != nil {
		auditLogFromContext(
			c,
			logger.AuditActionSSHKeyDelete,
			logger.AuditSeverityWarning,
			map[string]any{
				"reason": "invalid_ssh_key_id_format",
			},
		)
		BadRequest(c, "invalid SSH key ID format")
		return
	}

	sshKeyName, err := h.svc.DeleteSSHKey(userID, keyID)
	if err != nil {
		if err.Error() == fmt.Sprintf("no SSH key found with ID %d for user %d", keyID, userID) {
			auditLogFromContext(
				c,
				logger.AuditActionSSHKeyDelete,
				logger.AuditSeverityWarning,
				map[string]any{
					"reason": "ssh_key_not_found",
				},
			)
			NotFound(c, "SSH key not found")
			return
		}
		reqLog.Error().Err(err).Msg("failed to delete SSH key")
		auditLogFromContext(
			c,
			logger.AuditActionSSHKeyDelete,
			logger.AuditSeverityError,
			map[string]any{
				"reason": err.Error(),
			},
		)
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
	auditLogFromContext(
		c,
		logger.AuditActionSSHKeyDelete,
		logger.AuditSeverityInfo,
		map[string]any{
			"user_id":      userID,
			"ssh_key_name": sshKeyName,
		},
	)

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
		auditLogFromContext(
			c,
			logger.AuditActionWorkflowStatusGet,
			logger.AuditSeverityWarning,
			map[string]any{
				"reason": "workflow_id_required",
			},
		)
		BadRequest(c, "Workflow ID is required")
		return
	}

	workflowStatus, err := h.svc.GetWorkflowStatus(c.Request.Context(), workflowID)
	if err != nil {
		reqLog.Error().Err(err).Msg("failed to load workflow by UUID")
		auditLogFromContext(
			c,
			logger.AuditActionWorkflowStatusGet,
			logger.AuditSeverityError,
			map[string]any{
				"reason": err.Error(),
			},
		)
		InternalServerError(c)
		return
	}

	auditLogFromContext(
		c,
		logger.AuditActionWorkflowStatusGet,
		logger.AuditSeverityInfo,
		map[string]any{
			"workflow_id": workflowID,
		},
	)
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
		auditLogFromContext(
			c,
			logger.AuditActionPendingRecordsList,
			logger.AuditSeverityError,
			map[string]any{
				"reason": err.Error(),
			},
		)
		InternalServerError(c)
		return
	}
	auditLogFromContext(
		c,
		logger.AuditActionPendingRecordsList,
		logger.AuditSeverityInfo,
		map[string]any{
			"user_id": userID,
		},
	)

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
		auditLogFromContext(
			c,
			logger.AuditActionWorkflowList,
			logger.AuditSeverityError,
			map[string]any{
				"reason": err.Error(),
			},
		)
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
	auditLogFromContext(
		c,
		logger.AuditActionWorkflowList,
		logger.AuditSeverityInfo,
		map[string]any{
			"user_id":          userID,
			"workflows_length": len(workflows),
		},
	)

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

func auditLogFromContext(c *gin.Context, action logger.AuditActionType, severity logger.AuditSeverity, metadata map[string]any) {
	actorType, actorMetadata := resolveActorFromContext(c)
	auditLogWithActor(c, action, severity, metadata, actorType, actorMetadata)
}

func auditLogWithActor(
	c *gin.Context,
	action logger.AuditActionType,
	severity logger.AuditSeverity,
	metadata map[string]any,
	actorType logger.AuditActorType,
	actorMetadata map[string]any,
) {
	if metadata == nil {
		metadata = map[string]any{}
	}

	opts := []logger.AuditEntryOption{
		logger.WithAuditActionMetadata(metadata),
		logger.WithAuditSeverity(severity),
	}

	if len(actorMetadata) > 0 {
		opts = append(opts, logger.WithAuditActorMetadata(actorMetadata))
	}

	userAgent := ""
	if c.Request != nil {
		userAgent = c.Request.UserAgent()
	}

	logger.LogAudit(actorType, action, c.ClientIP(), userAgent, opts...)
}

func resolveActorFromContext(c *gin.Context) (logger.AuditActorType, map[string]any) {
	actorType := logger.AuditActorSystem
	actorMetadata := map[string]any{}

	if raw, exists := c.Get("is_admin"); exists {
		if isAdmin, ok := raw.(bool); ok && isAdmin {
			actorType = logger.AuditActorAdmin
			actorMetadata["is_admin"] = true
		}
	}

	if raw, exists := c.Get("user_id"); exists {
		switch v := raw.(type) {
		case int:
			if v != 0 {
				actorMetadata["user_id"] = v
				if actorType != logger.AuditActorAdmin {
					actorType = logger.AuditActorUser
				}
			}
		case int64:
			if v != 0 {
				actorMetadata["user_id"] = v
				if actorType != logger.AuditActorAdmin {
					actorType = logger.AuditActorUser
				}
			}
		}
	}

	if raw, exists := c.Get("user_email"); exists {
		if email, ok := raw.(string); ok && strings.TrimSpace(email) != "" {
			actorMetadata["email"] = email
		}
	}

	return actorType, actorMetadata
}
