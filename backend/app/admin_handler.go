package app

import (
	"kubecloud/internal"
	"kubecloud/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// GenerateVouchersInput holds all data needed when creating vouchers
type GenerateVouchersInput struct {
	Count int     `json:"count" binding:"required,gt=0"`
	Value float64 `json:"value" binding:"required,gt=0"`
}

type CreditRequestInput struct {
	Amount float64 `json:"amount" binding:"required,gt=0"`
	Memo   string  `json:"memo" binding:"required"`
}

// ListUsersHandler lists all users
func (h *Handler) ListUsersHandler(c *gin.Context) {

	users, err := h.db.ListAllUsers()
	if err != nil {
		log.Error().Err(err).Msg("failed to list all users")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error listing users"})
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

	err := h.db.DeleteUserByID(userID)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID).Msg("Failed to delete user")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})

}

// GenerateVouchersHandler generates bulk of vouchers
func (h *Handler) GenerateVouchersHandler(c *gin.Context) {
	var request GenerateVouchersInput

	// check on request format
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	var vouchers []models.Voucher
	for i := 0; i < request.Count; i++ {
		voucherCode := internal.GenerateRandomVoucher(5)
		voucher := models.Voucher{
			Voucher:   voucherCode,
			Value:     request.Value,
			CreatedAt: time.Now(),
		}

		if err := h.db.CreateVoucher(&voucher); err != nil {
			log.Error().Err(err).Msg("failed to create voucher")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save voucher"})
			return
		}

		vouchers = append(vouchers, voucher)
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":  "Vouchers generated successfully",
		"vouchers": vouchers,
	})

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

	user, err := h.db.GetUserByID(userID)
	if err != nil {
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create credit transaction"})
		return
	}

	if err := h.db.CreditUserBalance(user.ID, request.Amount); err != nil {
		log.Error().Err(err).Msg("Failed to credit user")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to credit user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User credited successfully",
		"user":    user.Email,
		"amount":  request.Amount,
		"memo":    request.Memo,
	})

}
