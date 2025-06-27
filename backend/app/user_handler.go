package app

import (
	"fmt"
	"kubecloud/internal"
	"kubecloud/models"
	"net/http"
	"time"

	substrate "github.com/threefoldtech/tfchain/clients/tfchain-client-go"
	proxy "github.com/threefoldtech/tfgrid-sdk-go/grid-proxy/pkg/client"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
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
}

// NewHandler create new handler
func NewHandler(tokenManager internal.TokenManager, db models.DB, config internal.Configuration, mailService internal.MailService, gridproxy proxy.Client, substrateClient *substrate.Substrate) *Handler {
	return &Handler{
		tokenManager:    tokenManager,
		db:              db,
		config:          config,
		mailService:     mailService,
		proxyClient:     gridproxy,
		substrateClient: substrateClient,
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

// RegisterHandler registers user to the system
func (h *Handler) RegisterHandler(c *gin.Context) {
	var request RegisterInput

	// check on request format
	if err := c.ShouldBindJSON(&request); err != nil {
		log.Error().Err(err).Send()
		Error(c, http.StatusBadRequest, "invalid request format", err.Error())
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
	log.Debug().Int("generated_code", code).Send()
	subject, body := h.mailService.SignUpMailContent(code, h.config.MailSender.Timeout, request.Name, h.config.Server.Host)

	err := h.mailService.SendMail(h.config.MailSender.Email, request.Email, subject, body)
	if err != nil {
		log.Error().Err(err).Msg("failed to send verification code")
		Error(c, http.StatusInternalServerError, "internal server error", "")
		return
	}

	// hash password
	hashedPassword, err := internal.HashAndSaltPassword([]byte(request.Password))
	if err != nil {
		log.Error().Err(err).Msg("error hashing password")
		Error(c, http.StatusInternalServerError, "internal server error", "")
		return
	}

	isAdmin := internal.Contains(h.config.Admins, request.Email)

	mnemonic, _, err := internal.SetupUserOnTFChain(h.substrateClient, h.config)
	if err != nil {
		log.Error().Err(err).Msg("failed to setup user on TFChain")
		Error(c, http.StatusInternalServerError, "internal server error", "")
		return
	}

	user := models.User{
		Username: request.Name,
		Email:    request.Email,
		Password: hashedPassword,
		Admin:    isAdmin,
		Code:     code,
		Mnemonic: mnemonic,
	}

	// If user exists but not verified
	if getErr != gorm.ErrRecordNotFound {
		if !existingUser.Verified {
			user.ID = existingUser.ID
			user.UpdatedAt = time.Now()
			err = h.db.UpdateUserByID(&user)
			if err != nil {
				log.Error().Err(err).Send()
				Error(c, http.StatusInternalServerError, "internal server error", "")
				return
			}
		}
	}

	// create user model in db
	if getErr != nil {
		err = h.db.RegisterUser(&user)
		if err != nil {
			log.Error().Err(err).Send()
			Error(c, http.StatusInternalServerError, "internal server error", "")
			return
		}
	}

	Success(c, http.StatusOK, "Verification code sent successfully", map[string]interface{}{
		"email":   request.Email,
		"timeout": fmt.Sprintf("%d seconds", h.config.MailSender.Timeout),
	})

}

// VerifyRegisterCode verifies email when signing uo
func (h *Handler) VerifyRegisterCode(c *gin.Context) {
	var request VerifyCodeInput

	if err := c.ShouldBindJSON(&request); err != nil {
		log.Error().Err(err).Send()
		Error(c, http.StatusBadRequest, "invalid request format", err.Error())
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
		Error(c, http.StatusInternalServerError, "internal server error", "")
		return

	}

	subject, body := h.mailService.WelcomeMailContent(user.Username, h.config.Server.Host)
	err = h.mailService.SendMail(h.config.MailSender.Email, request.Email, subject, body)
	if err != nil {
		log.Error().Err(err).Send()
		Error(c, http.StatusInternalServerError, "internal server error", "")
		return
	}

	// create token pairs
	tokenPair, err := h.tokenManager.CreateTokenPair(user.ID, user.Username, user.Admin)
	if err != nil {
		log.Error().Err(err).Msg("Failed to generate token pair")
		Error(c, http.StatusInternalServerError, "internal server error", "")
		return
	}
	Success(c, http.StatusCreated, "token pair generated", tokenPair)

}

// LoginUserHandler logs user into the system
func (h *Handler) LoginUserHandler(c *gin.Context) {
	var request LoginInput

	// check on request format
	if err := c.ShouldBindJSON(&request); err != nil {
		Error(c, http.StatusBadRequest, "invalid request format", err.Error())
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
		Error(c, http.StatusInternalServerError, "internal server error", "")
		return
	}
	Success(c, http.StatusCreated, "token pair generated", tokenPair)

}

// RefreshTokenHandler handles token refresh requests
func (h *Handler) RefreshTokenHandler(c *gin.Context) {
	var request RefreshTokenInput

	if err := c.ShouldBindJSON(&request); err != nil {
		log.Error().Err(err).Send()
		Error(c, http.StatusBadRequest, "invalid request format", err.Error())
		return
	}

	accessToken, err := h.tokenManager.AccessTokenFromRefresh(request.RefreshToken)
	if err != nil {
		log.Error().Err(err).Send()
		Error(c, http.StatusUnauthorized, "refresh token failed", "Invalid or expired refresh token")

		return
	}

	Success(c, http.StatusOK, "access token refreshed successfully", map[string]interface{}{
		"access_token": accessToken,
	})
}

// ForgotPasswordHandler sends user verification code
func (h *Handler) ForgotPasswordHandler(c *gin.Context) {
	var request EmailInput

	if err := c.ShouldBindJSON(&request); err != nil {
		log.Error().Err(err).Send()
		Error(c, http.StatusBadRequest, "invalid request format", err.Error())
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
		Error(c, http.StatusInternalServerError, "internal server error", "")
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
		Error(c, http.StatusInternalServerError, "internal server error", "")
		return
	}

	Success(c, http.StatusOK, "Verification code sent", map[string]interface{}{
		"email":   request.Email,
		"timeout": fmt.Sprintf("%d seconds", h.config.MailSender.Timeout),
	})

}

// VerifyForgetPasswordCodeHandler verifies code sent to user when forgetting password
func (h *Handler) VerifyForgetPasswordCodeHandler(c *gin.Context) {
	var request VerifyCodeInput

	if err := c.ShouldBindJSON(&request); err != nil {
		log.Error().Err(err).Send()
		Error(c, http.StatusBadRequest, "invalid request format", err.Error())
		return
	}

	// get user by email
	user, err := h.db.GetUserByEmail(request.Email)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			Error(c, http.StatusBadRequest, "invalid request format", err.Error())
			return

		}
		log.Error().Err(err).Msg("failed to get user by email")
		Error(c, http.StatusInternalServerError, "internal server error", "")
		return

	}

	if user.Code != request.Code {
		Error(c, http.StatusBadRequest, "invalid code", "")
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
		Error(c, http.StatusInternalServerError, "internal server error", "")
		return
	}
	Success(c, http.StatusCreated, "verification successful", tokenPair)

}

// ChangePasswordHandler changes password of user
func (h *Handler) ChangePasswordHandler(c *gin.Context) {
	var request ChangePasswordInput

	if err := c.ShouldBindJSON(&request); err != nil {
		log.Error().Err(err).Send()
		Error(c, http.StatusBadRequest, "invalid request format", err.Error())

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
		Error(c, http.StatusInternalServerError, "internal server error", "")
		return
	}

	err = h.db.UpdatePassword(request.Email, hashedPassword)
	if err == gorm.ErrRecordNotFound {
		log.Error().Err(err).Msg("user not found")
		Error(c, http.StatusNotFound, "user not found", err.Error())

		return
	}

	if err != nil {
		log.Error().Err(err).Send()
		Error(c, http.StatusInternalServerError, "internal server error", "")
		return

	}

	Success(c, http.StatusOK, "password updated successfully", nil)

}
