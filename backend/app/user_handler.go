package app

import (
	"kubecloud/internal"
	"kubecloud/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
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

type LoginInput struct {
	Email    string `json:"email" binding:"required" validate:"mail"`
	Password string `json:"password" binding:"required" validate:"password"`
}

// RegisterHandler registers user to the system
func (h *Handler) RegisterHandler(c *gin.Context) {
	var request RegisterInput

	// check on request format
	if err := c.ShouldBindJSON(&request); err != nil {
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

	user := models.User{
		Username: request.Name,
		Email:    request.Email,
		Password: hashed,
		Salt:     salt,
		Admin:    isAdmin,
	}

	// create user model in db
	err = h.db.RegisterUser(&user)
	if err != nil {
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
