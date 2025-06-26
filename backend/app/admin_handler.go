package app

import (
	"fmt"
	"kubecloud/internal"
	"kubecloud/models"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// GenerateVouchersInput holds all data needed when creating vouchers
type GenerateVouchersInput struct {
	Count int     `json:"count" binding:"required,gt=0" validate:"required,gt=0"`
	Value float64 `json:"value" binding:"required,gt=0" validate:"required,gt=0"`
}

// CreditRequestInput represents a request to credit a user's balance
type CreditRequestInput struct {
	Amount float64 `json:"amount" binding:"required,gt=0" validate:"required,gt=0"`
	Memo   string  `json:"memo" binding:"required,min=3,max=255" validate:"required"`
}

// ListUsersHandler lists all users
func (h *Handler) ListUsersHandler(c *gin.Context) {

	users, err := h.db.ListAllUsers()
	if err != nil {
		log.Error().Err(err).Msg("failed to list all users")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, users)
}

// DeleteUsersHandler deletes user from system
func (h *Handler) DeleteUsersHandler(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID is required"})
		return
	}

	authUserID := c.GetString("user_id")
	if userID == authUserID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Admins cannot delete their own account"})
		return
	}

	ID, err := strconv.Atoi(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	err = h.db.DeleteUserByID(ID)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID).Msg("Failed to delete user")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User is deleted successfully"})

}

// GenerateVouchersHandler generates bulk of vouchers
func (h *Handler) GenerateVouchersHandler(c *gin.Context) {
	var request GenerateVouchersInput

	// check on request format
	if err := c.ShouldBindJSON(&request); err != nil {
		log.Error().Err(err).Send()
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	var vouchers []models.Voucher
	for i := 0; i < request.Count; i++ {
		voucherCode := internal.GenerateRandomVoucher(h.config.Voucher.NameLength)
		timestampPart := fmt.Sprintf("%02d%02d", time.Now().Minute(), time.Now().Second())
		fullCode := fmt.Sprintf("%s-%s", voucherCode, timestampPart)

		voucher := models.Voucher{
			Voucher:   fullCode,
			Value:     request.Value,
			CreatedAt: time.Now(),
		}

		if err := h.db.CreateVoucher(&voucher); err != nil {
			log.Error().Err(err).Msg("failed to create voucher")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}

		vouchers = append(vouchers, voucher)
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":  "Vouchers generated successfully",
		"vouchers": vouchers,
	})

}

// ListVouchersHandler returns all vouchers in system
func (h *Handler) ListVouchersHandler(c *gin.Context) {

	vouchers, err := h.db.ListAllVouchers()
	if err != nil {
		log.Error().Err(err).Msg("failed to list all vouchers")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, vouchers)
}

func (h *Handler) CreditUserHandler(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID is required"})
		return
	}

	var request CreditRequestInput
	// check on request format
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	ID, err := strconv.Atoi(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	user, err := h.db.GetUserByID(ID)
	if err != nil {
		log.Error().Err(err).Send()
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	adminID, exists := c.Get("admin_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Admin ID not found in context"})
		return
	}

	transaction := models.Transaction{
		UserID:    user.ID,
		AdminID:   adminID.(int),
		Amount:    request.Amount,
		Memo:      request.Memo,
		CreatedAt: time.Now(),
	}

	if err := h.db.CreateTransaction(&transaction); err != nil {
		log.Error().Err(err).Msg("Failed to create credit transaction")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	if err := h.db.CreditUserBalance(user.ID, request.Amount); err != nil {
		log.Error().Err(err).Msg("Failed to credit user")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User credited successfully",
		"user":    user.Email,
		"amount":  request.Amount,
		"memo":    request.Memo,
	})

}
