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
	Count       int     `json:"count" binding:"required,gt=0" validate:"required,gt=0"`
	Value       float64 `json:"value" binding:"required,gt=0" validate:"required,gt=0"`
	ExpireAfter int     `json:"expire_after_days" binding:"required,gt=0"`
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
		Error(c, http.StatusInternalServerError, "internal server error", "")
		return
	}

	Success(c, http.StatusOK, "Users retrieved successfully", map[string]interface{}{
		"users": users,
	})
}

// DeleteUsersHandler deletes user from system
func (h *Handler) DeleteUsersHandler(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		Error(c, http.StatusBadRequest, "User ID is required", "")
		return
	}

	authUserID := c.GetString("user_id")
	if userID == authUserID {
		Error(c, http.StatusForbidden, "Admins cannot delete their own account", "")
		return
	}

	ID, err := strconv.Atoi(userID)
	if err != nil {
		log.Error().Err(err).Send()
		Error(c, http.StatusBadRequest, "Invalid user ID", err.Error())
		return
	}

	err = h.db.DeleteUserByID(ID)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID).Msg("Failed to delete user")
		Error(c, http.StatusInternalServerError, "internal server error", "")
		return
	}

	Success(c, http.StatusOK, "User is deleted successfully", nil)

}

// GenerateVouchersHandler generates bulk of vouchers
func (h *Handler) GenerateVouchersHandler(c *gin.Context) {
	var request GenerateVouchersInput

	// check on request format
	if err := c.ShouldBindJSON(&request); err != nil {
		log.Error().Err(err).Send()
		Error(c, http.StatusBadRequest, "invalid request format", err.Error())
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
			ExpiresAt: time.Now().Add(time.Duration(request.ExpireAfter) * 24 * time.Hour),
		}

		if err := h.db.CreateVoucher(&voucher); err != nil {
			log.Error().Err(err).Msg("failed to create voucher")
			Error(c, http.StatusInternalServerError, "internal server error", "")
			return
		}

		vouchers = append(vouchers, voucher)
	}

	Success(c, http.StatusCreated, "Vouchers are generated successfully", map[string]interface{}{
		"vouchers": vouchers,
	})
}

// ListVouchersHandler returns all vouchers in system
func (h *Handler) ListVouchersHandler(c *gin.Context) {

	vouchers, err := h.db.ListAllVouchers()
	if err != nil {
		log.Error().Err(err).Msg("failed to list all vouchers")
		Error(c, http.StatusInternalServerError, "internal server error", "")
		return
	}
	Success(c, http.StatusOK, "Vouchers Retrieved successfully", map[string]interface{}{
		"vouchers": vouchers,
	})
}

// CreditUserHandler lets admin credit specific user with money
func (h *Handler) CreditUserHandler(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		Error(c, http.StatusBadRequest, "User ID is required", "")
		return
	}

	var request CreditRequestInput
	// check on request format
	if err := c.ShouldBindJSON(&request); err != nil {
		Error(c, http.StatusBadRequest, "invalid request format", err.Error())
		return
	}

	ID, err := strconv.Atoi(userID)
	if err != nil {
		log.Error().Msg("invalid user ID or format")
		Error(c, http.StatusBadRequest, "Invalid user ID format", "")
		return
	}

	user, err := h.db.GetUserByID(ID)
	if err != nil {
		log.Error().Err(err).Send()
		Error(c, http.StatusInternalServerError, "internal server error", "")
		return
	}

	// get admin ID from middleware context
	adminID, exists := c.Get("user_id")
	if !exists {
		Error(c, http.StatusUnauthorized, "Admin ID not found in context", "")
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
		Error(c, http.StatusInternalServerError, "internal server error", "")
		return
	}

	if err := h.db.CreditUserBalance(user.ID, request.Amount); err != nil {
		log.Error().Err(err).Msg("Failed to credit user")
		Error(c, http.StatusInternalServerError, "internal server error", "")
		return
	}

	Success(c, http.StatusOK, "User credited successfully", map[string]interface{}{
		"user":   user.Email,
		"amount": request.Amount,
		"memo":   request.Memo,
	})

}
