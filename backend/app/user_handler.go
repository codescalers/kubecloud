package app

import (
	"kubecloud/internal"
	"kubecloud/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/paymentmethod"
	"gorm.io/gorm"
)

// Handler struct holds configs for all handlers
type Handler struct {
	tokenManager internal.TokenManager
	db           models.DB
	config       internal.Configuration
}

// NewHandler create new handler
func NewHandler(tokenManager internal.TokenManager, db models.DB, config internal.Configuration) *Handler {
	return &Handler{
		tokenManager: tokenManager,
		db:           db,
		config:       config,
	}
}

// RegisterInput struct for data needed when user register
type RegisterInput struct {
	Name            string `json:"name" binding:"required" validate:"min=3,max=20"`
	Email           string `json:"email" binding:"required" validate:"mail"`
	Password        string `json:"password" binding:"required" validate:"password"`
	ConfirmPassword string `json:"confirm_password" binding:"required" validate:"password"`
}

// LoginInput struct for login handler
type LoginInput struct {
	Email    string `json:"email" binding:"required" validate:"mail"`
	Password string `json:"password" binding:"required" validate:"password"`
}

// RefreshTokenInput struct when user refresh token
type RefreshTokenInput struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// EmailInput struct for user when forgetting password
type EmailInput struct {
	Email string `json:"email" validate:"required,email"`
}

// VerifyCodeInput struct takes verification code from user
type VerifyCodeInput struct {
	Email string `json:"email" binding:"required"`
	Code  int    `json:"code" binding:"required"`
}

// ChangePasswordInput struct for user to change password
type ChangePasswordInput struct {
	Email           string `json:"email" binding:"required"`
	Password        string `json:"password" binding:"required" validate:"password"`
	ConfirmPassword string `json:"confirm_password" binding:"required" validate:"password"`
}

// ChargeBalanceInput struct holds required data to charge users' balance
type ChargeBalanceInput struct {
	CardType     string  `json:"card_type" binding:"required"`
	PaymentToken string  `json:"payment_method_id" binding:"required"`
	Amount       float64 `json:"amount" binding:"required"`
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
	_, err := h.db.GetUserByEmail(request.Email)
	if err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "user already registered"})
		return
	}

	// hash password
	salt, err := internal.GenerateSalt()
	if err != nil {
		log.Error().Err(err).Msg("Error Generating salt for password")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error hashing password"})
		return
	}
	hashed := internal.HashPassword(request.Password, salt)

	isAdmin := internal.IsAdmin(request.Email, h.config.Admins)

	customer, err := internal.CreateStripeCustomer(request.Name, request.Email)
	if err != nil {
		log.Error().Err(err).Send()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error creating stripe account"})
		return
	}

	user := models.User{
		StripeCustomerID: customer.ID,
		Username:         request.Name,
		Email:            request.Email,
		Password:         hashed,
		Salt:             salt,
		Admin:            isAdmin,
	}

	// create user model in db
	err = h.db.RegisterUser(&user)
	if err != nil {
		log.Error().Err(err).Send()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to register user"})
		return
	}

	saved, _ := h.db.GetUserByEmail(user.Email)

	// create token pairs
	tokenPair, err := h.tokenManager.CreateTokenPair(saved.ID, saved.Username, isAdmin)
	if err != nil {
		log.Error().Err(err).Msg("Failed to generate token pair")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate tokens"})
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
	match := internal.VerifyPassword(user.Password, request.Password, user.Salt)
	if !match {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "email or password is incorrect"})
		return
	}

	// create token pairs
	tokenPair, err := h.tokenManager.CreateTokenPair(user.ID, user.Username, user.Admin)
	if err != nil {
		log.Error().Err(err).Msg("Failed to generate token pair")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate tokens"})
		return
	}
	c.JSON(http.StatusCreated, tokenPair)

}

// RefreshTokenHandler handles token refresh requests
func (h *Handler) RefreshTokenHandler(c *gin.Context) {
	var request RefreshTokenInput

	if err := c.ShouldBindJSON(&request); err != nil {
		log.Error().Err(err).Send()
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	accessToken, err := h.tokenManager.RefreshAccessToken(request.RefreshToken)
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
		log.Error().Err(err).Msg("failed to get user by email")
		c.JSON(http.StatusNotFound, gin.H{"error": "failed to get email"})
		return

	}

	code := internal.GenerateRandomCode()
	subject, body := internal.ResetPasswordMailContent(code, h.config.MailSender.Timeout, user.Username, h.config.Server.Host)
	err = internal.SendMail(h.config.MailSender.Email, h.config.MailSender.SendGridKey, request.Email, subject, body)

	if err != nil {
		log.Error().Err(err).Msg("failed to send verification code")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to send verification code"})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update user data"})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user by email"})
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
	isAdmin := internal.IsAdmin(request.Email, h.config.Admins)

	// create token pairs
	tokenPair, err := h.tokenManager.CreateTokenPair(user.ID, user.Username, isAdmin)
	if err != nil {
		log.Error().Err(err).Msg("Failed to generate token pair")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate tokens"})
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
	salt, err := internal.GenerateSalt()
	if err != nil {
		log.Error().Err(err).Msg("error Generating salt for password")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error hashing password"})
		return
	}
	hashed := internal.HashPassword(request.Password, salt)

	err = h.db.UpdatePassword(request.Email, hashed, salt)
	if err == gorm.ErrRecordNotFound {
		log.Error().Err(err).Msg("user not found")
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	if err != nil {
		log.Error().Err(err).Send()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error updating password"})
		return

	}

	c.JSON(http.StatusOK, gin.H{"message": "Password updated successfully"})

}

func (h *Handler) ChargeBalance(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to extract user ID from context"})
		return
	}

	var request ChargeBalanceInput
	if err := c.ShouldBindJSON(&request); err != nil {
		log.Error().Err(err).Send()
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	if request.Amount <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Amount must be greater than zero"})
		return
	}

	user, err := h.db.GetUserByID(userID)
	if err != nil {
		log.Error().Err(err).Send()
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	paymentMethod, err := internal.CreatePaymentMethod("card", request.PaymentToken)
	if err != nil {
		log.Error().Err(err).Msg("error creating payment method")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create payment method"})
		return
	}

	_, err = paymentmethod.Attach(paymentMethod.ID, &stripe.PaymentMethodAttachParams{
		Customer: stripe.String(user.StripeCustomerID),
	})
	if err != nil {
		log.Error().Err(err).Msg("error attaching payment method to customer")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to attach payment method"})
		return
	}

	intent, err := internal.CreatePaymentIntent(user.StripeCustomerID, paymentMethod.ID, h.config.Currency, request.Amount)
	if err != nil {
		log.Error().Err(err).Msg("error creating payment intent")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create payment intent"})
		return
	}

	user.CreditCardBalance += float64(request.Amount)

	err = h.db.UpdateUserByID(&user)
	if err != nil {
		log.Error().Err(err).Msg("error updating user data")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update user data"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":           "Balance is charged successfully",
		"payment_intent_id": intent.ID,
		"new_balance":       user.CreditCardBalance,
	})

}
