package app

import (
	"kubecloud/internal"
	"kubecloud/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// Handler struct holds configs for all handlers
type Handler struct {
	tokenManager internal.TokenManager
	db           models.DB
	config       internal.Configuration
	mailService  internal.MailService
	redis        *internal.RedisClient
	sseManager   *internal.SSEManager
}

// NewHandler create new handler
func NewHandler(tokenManager internal.TokenManager, db models.DB, config internal.Configuration, mailService internal.MailService, redis *internal.RedisClient, sseManager *internal.SSEManager) *Handler {
	return &Handler{
		tokenManager: tokenManager,
		db:           db,
		config:       config,
		mailService:  mailService,
		redis:        redis,
		sseManager:   sseManager,
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// password and confirm password should match
	if request.Password != request.ConfirmPassword {
		c.JSON(http.StatusBadRequest, gin.H{"error": "password and confirm password don't match"})
		return
	}

	// check if user previously exists
	existingUser, getErr := h.db.GetUserByEmail(request.Email)
	if getErr != nil && getErr != gorm.ErrRecordNotFound {
		c.JSON(http.StatusConflict, gin.H{"error": "user already registered"})
		return
	}

	code := internal.GenerateRandomCode()
	subject, body := h.mailService.SignUpMailContent(code, h.config.MailSender.Timeout, request.Name, h.config.Server.Host)

	err := h.mailService.SendMail(h.config.MailSender.Email, request.Email, subject, body)
	if err != nil {
		log.Error().Err(err).Msg("failed to send verification code")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	// hash password
	hashedPassword, err := internal.HashAndSaltPassword([]byte(request.Password))
	if err != nil {
		log.Error().Err(err).Msg("error hashing password")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	isAdmin := internal.Contains(h.config.Admins, request.Email)

	user := models.User{
		Username: request.Name,
		Email:    request.Email,
		Password: hashedPassword,
		Admin:    isAdmin,
		Code:     code,
	}

	// If user exists but not verified
	if getErr != nil {
		if existingUser.Verified {
			user.ID = existingUser.ID
			user.UpdatedAt = time.Now()
			err = h.db.UpdateUserByID(&user)
			if err != nil {
				log.Error().Err(err).Send()
				c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
				return
			}
		}
	}

	// create user model in db
	if getErr != nil {
		err = h.db.RegisterUser(&user)
		if err != nil {
			log.Error().Err(err).Send()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Verification code has been sent to " + request.Email,
		"timeout": h.config.MailSender.Timeout,
	})

}

func (h *Handler) VerifyRegisterCode(c *gin.Context) {
	var request VerifyCodeInput

	if err := c.ShouldBindJSON(&request); err != nil {
		log.Error().Err(err).Send()
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// get user by email
	user, err := h.db.GetUserByEmail(request.Email)
	if err != nil {
		log.Error().Err(err).Msg("failed to get user by email")
		c.JSON(http.StatusBadRequest, gin.H{"error": "email or password is incorrect"})
		return

	}

	if user.Verified {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user already registered"})
		return
	}

	if user.Code != request.Code {
		c.JSON(http.StatusBadRequest, gin.H{"error": "wrong code"})
		return
	}

	if user.UpdatedAt.Add(time.Duration(h.config.MailSender.Timeout) * time.Second).Before(time.Now()) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "code has expired"})
		return
	}

	err = h.db.UpdateUserVerification(user.ID, true)
	if err != nil {
		log.Error().Err(err).Send()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return

	}

	subject, body := h.mailService.WelcomeMailContent(user.Username, h.config.Server.Host)
	err = h.mailService.SendMail(h.config.MailSender.Email, request.Email, subject, body)
	if err != nil {
		log.Error().Err(err).Send()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	// create token pairs
	tokenPair, err := h.tokenManager.CreateTokenPair(user.ID, user.Username, user.Admin)
	if err != nil {
		log.Error().Err(err).Msg("Failed to generate token pair")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusCreated, tokenPair)
}

// LoginUserHandler logs user into the system
func (h *Handler) LoginUserHandler(c *gin.Context) {
	var request LoginInput

	// check on request format
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// get user by email
	user, err := h.db.GetUserByEmail(request.Email)
	if err != nil {
		log.Error().Err(err).Msg("failed to get user by email")
		c.JSON(http.StatusBadRequest, gin.H{"error": "email or password is incorrect"})
		return

	}

	// verify password
	match := internal.VerifyPassword(user.Password, request.Password)
	if !match {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "email or password is incorrect"})
		return
	}

	// create token pairs
	tokenPair, err := h.tokenManager.CreateTokenPair(user.ID, user.Username, user.Admin)
	if err != nil {
		log.Error().Err(err).Msg("Failed to generate token pair")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, tokenPair)

}

// RefreshTokenHandler handles token refresh requests
func (h *Handler) RefreshTokenHandler(c *gin.Context) {
	var request RefreshTokenInput

	if err := c.ShouldBindJSON(&request); err != nil {
		log.Error().Err(err).Send()
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	accessToken, err := h.tokenManager.AccessTokenFromRefresh(request.RefreshToken)
	if err != nil {
		log.Error().Err(err).Send()
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired refresh token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"access_token": accessToken})
}

// ForgotPasswordHandler sends user verification code
func (h *Handler) ForgotPasswordHandler(c *gin.Context) {
	var request EmailInput

	if err := c.ShouldBindJSON(&request); err != nil {
		log.Error().Err(err).Send()
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// get user by email
	user, err := h.db.GetUserByEmail(request.Email)
	if err != nil {
		log.Error().Err(err).Msg("failed to get user ")
		c.JSON(http.StatusNotFound, gin.H{"error": "failed to get user"})
		return

	}

	code := internal.GenerateRandomCode()
	subject, body := h.mailService.ResetPasswordMailContent(code, h.config.MailSender.Timeout, user.Username, h.config.Server.Host)
	err = h.mailService.SendMail(h.config.MailSender.Email, request.Email, subject, body)

	if err != nil {
		log.Error().Err(err).Msg("failed to send verification code")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Verification code has been sent to " + request.Email,
		"timeout": h.config.MailSender.Timeout,
	})

}

// VerifyForgetPasswordCodeHandler verifies code sent to user when forgetting password
func (h *Handler) VerifyForgetPasswordCodeHandler(c *gin.Context) {
	var request VerifyCodeInput

	if err := c.ShouldBindJSON(&request); err != nil {
		log.Error().Err(err).Send()
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// get user by email
	user, err := h.db.GetUserByEmail(request.Email)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return

		}
		log.Error().Err(err).Msg("failed to get user by email")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return

	}

	if user.Code != request.Code {
		c.JSON(http.StatusBadRequest, gin.H{"error": "wrong code"})
		return
	}

	if user.UpdatedAt.Add(time.Duration(h.config.MailSender.Timeout) * time.Second).Before(time.Now()) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "code has expired"})
		return
	}
	isAdmin := internal.Contains(h.config.Admins, request.Email)

	// create token pairs
	tokenPair, err := h.tokenManager.CreateTokenPair(user.ID, user.Username, isAdmin)
	if err != nil {
		log.Error().Err(err).Msg("Failed to generate token pair")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusCreated, tokenPair)

}

// ChangePasswordHandler changes password of user
func (h *Handler) ChangePasswordHandler(c *gin.Context) {
	var request ChangePasswordInput

	if err := c.ShouldBindJSON(&request); err != nil {
		log.Error().Err(err).Send()
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	if request.Password != request.ConfirmPassword {
		c.JSON(http.StatusBadRequest, gin.H{"error": "password and confirm password don't match"})
		return
	}

	// hash password
	hashedPassword, err := internal.HashAndSaltPassword([]byte(request.Password))
	if err != nil {
		log.Error().Err(err).Msg("error hashing password")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	err = h.db.UpdatePassword(request.Email, hashedPassword)
	if err == gorm.ErrRecordNotFound {
		log.Error().Err(err).Msg("user not found")
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	if err != nil {
		log.Error().Err(err).Send()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return

	}

	c.JSON(http.StatusOK, gin.H{"message": "Password is updated successfully"})

}
