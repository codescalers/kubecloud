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
	Amount uint64 `json:"amount" binding:"required,gt=0" validate:"required,gt=0"`
	Memo   string `json:"memo" binding:"required,min=3,max=255" validate:"required"`
}

// CreditUserResponse holds the response data after crediting a user
type CreditUserResponse struct {
	User   string `json:"user"`
	Amount uint64 `json:"amount"`
	Memo   string `json:"memo"`
}

// AdminMailInput represents the input for sending a mail to all users
// Only accessible by admins
// @Summary Send mail to all users
// @Description Allows admin to send a custom email to all users
// @Tags admin
// @ID admin-mail-all-users
// @Accept json
// @Produce json
// @Param body body AdminMailInput true "Admin Mail Input"
// @Success 200 {object} APIResponse
// @Failure 400 {object} APIResponse "Invalid request format"
// @Failure 500 {object} APIResponse
// @Security AdminMiddleware
// @Router /users/mail [post]
type AdminMailInput struct {
	Subject string `json:"subject" binding:"required"`
	Body    string `json:"body" binding:"required"`
}

// @Summary Get all users
// @Description Returns a list of all users
// @Tags admin
// @ID get-all-users
// @Accept json
// @Produce json
// @Success 200 {array} models.User
// @Failure 500 {object} APIResponse
// @Security AdminMiddleware
// @Router /users [get]
// ListUsersHandler lists all users
func (h *Handler) ListUsersHandler(c *gin.Context) {
	users, err := h.db.ListAllUsers()
	if err != nil {
		log.Error().Err(err).Msg("failed to list all users")
		InternalServerError(c)
		return
	}

	Success(c, http.StatusOK, "Users are retrieved successfully", map[string]interface{}{
		"users": users,
	})
}

// @Summary Delete a user
// @Description Deletes a user from the system
// @Tags admin
// @ID delete-user
// @Accept json
// @Produce json
// @Param user_id path string true "User ID"
// @Success 200 {object} APIResponse
// @Failure 400 {object} APIResponse "Invalid user ID"
// @Failure 403 {object} APIResponse "Admins cannot delete their own account"
// @Failure 500 {object} APIResponse
// @Security AdminMiddleware
// @Router /users/{user_id} [delete]
// DeleteUsersHandler deletes user from system
func (h *Handler) DeleteUsersHandler(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		Error(c, http.StatusBadRequest, "User ID is required", "")
		return
	}

	id, err := strconv.Atoi(userID)
	if err != nil {
		log.Error().Err(err).Send()
		Error(c, http.StatusBadRequest, "Invalid user ID", err.Error())
		return
	}

	authUserID := c.GetInt("user_id")
	if id == authUserID {
		Error(c, http.StatusForbidden, "Admins cannot delete their own account", "")
		return
	}

	err = h.db.DeleteUserByID(id)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID).Msg("Failed to delete user")
		InternalServerError(c)
		return
	}

	Success(c, http.StatusOK, "User is deleted successfully", nil)

}

// @Summary Generate vouchers
// @Description Generates a bulk of vouchers
// @Tags admin
// @ID generate-vouchers
// @Accept json
// @Produce json
// @Param body body GenerateVouchersInput true "Generate Vouchers Input"
// @Success 201 {array} models.Voucher
// @Failure 400 {object} APIResponse "Invalid request format"
// @Failure 500 {object} APIResponse
// @Security AdminMiddleware
// @Router /vouchers/generate [post]
// GenerateVouchersHandler generates bulk of vouchers
func (h *Handler) GenerateVouchersHandler(c *gin.Context) {
	var request GenerateVouchersInput

	// check on request format
	if err := c.ShouldBindJSON(&request); err != nil {
		log.Error().Err(err).Send()
		Error(c, http.StatusBadRequest, "Invalid request format", err.Error())
		return
	}

	var vouchers []models.Voucher
	for i := 0; i < request.Count; i++ {
		voucherCode := internal.GenerateRandomVoucher(h.config.VoucherNameLength)
		timestampPart := fmt.Sprintf("%02d%02d", time.Now().Minute(), time.Now().Second())
		fullCode := fmt.Sprintf("%s-%s", voucherCode, timestampPart)

		voucher := models.Voucher{
			Code:      fullCode,
			Value:     request.Value,
			CreatedAt: time.Now(),
			ExpiresAt: time.Now().Add(time.Duration(request.ExpireAfter) * 24 * time.Hour),
		}

		if err := h.db.CreateVoucher(&voucher); err != nil {
			log.Error().Err(err).Msg("failed to create voucher")
			InternalServerError(c)
			return
		}

		vouchers = append(vouchers, voucher)
	}

	Success(c, http.StatusCreated, "Vouchers are generated successfully", map[string]interface{}{
		"vouchers": vouchers,
	})
}

// @Summary List vouchers
// @Description Returns all vouchers in the system
// @Tags admin
// @ID list-vouchers
// @Accept json
// @Produce json
// @Success 200 {array} models.Voucher
// @Failure 500 {object} APIResponse
// @Security AdminMiddleware
// @Router /vouchers [get]
// ListVouchersHandler returns all vouchers in system
func (h *Handler) ListVouchersHandler(c *gin.Context) {

	vouchers, err := h.db.ListAllVouchers()
	if err != nil {
		log.Error().Err(err).Msg("failed to list all vouchers")
		InternalServerError(c)
		return
	}
	Success(c, http.StatusOK, "Vouchers are Retrieved successfully", map[string]interface{}{
		"vouchers": vouchers,
	})
}

// @Summary Credit user balance
// @Description Credits a specific user's balance
// @Tags admin
// @ID credit-user
// @Accept json
// @Produce json
// @Param user_id path string true "User ID"
// @Param body body CreditRequestInput true "Credit Request Input"
// @Success 201 {object} CreditUserResponse
// @Failure 400 {object} APIResponse "Invalid request format or user ID"
// @Failure 500 {object} APIResponse
// @Security AdminMiddleware
// @Router /users/{user_id}/credit [post]
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
		Error(c, http.StatusBadRequest, "Invalid request format", err.Error())
		return
	}

	id, err := strconv.Atoi(userID)
	if err != nil {
		log.Error().Err(err).Send()
		Error(c, http.StatusBadRequest, "Invalid user ID format", "")
		return
	}

	user, err := h.db.GetUserByID(id)
	if err != nil {
		log.Error().Err(err).Send()
		InternalServerError(c)
		return
	}

	// get admin ID from middleware context
	adminID := c.GetInt("user_id")

	transaction := models.Transaction{
		UserID:    user.ID,
		AdminID:   adminID,
		Amount:    request.Amount,
		Memo:      request.Memo,
		CreatedAt: time.Now(),
	}

	if err := h.db.CreateTransaction(&transaction); err != nil {
		log.Error().Err(err).Msg("Failed to create credit transaction")
		InternalServerError(c)
		return
	}

	if err := h.db.CreditUserBalance(user.ID, request.Amount); err != nil {
		log.Error().Err(err).Msg("Failed to credit user")
		InternalServerError(c)
		return
	}

	err = internal.TransferTFTs(h.substrateClient, request.Amount, user.Mnemonic, h.config.SystemAccount.Mnemonic)
	if err != nil {
		log.Error().Err(err).Send()
		InternalServerError(c)
		return
	}

	Success(c, http.StatusCreated, "User is credited successfully", CreditUserResponse{
		User:   user.Email,
		Amount: request.Amount,
		Memo:   request.Memo,
	})

}

func (h *Handler) AdminMailAllUsersHandler(c *gin.Context) {
	var input AdminMailInput
	if err := c.ShouldBindJSON(&input); err != nil {
		Error(c, http.StatusBadRequest, "Invalid request format", err.Error())
		return
	}
	users, err := h.db.ListAllUsers()
	if err != nil {
		InternalServerError(c)
		return
	}
	var (
		sent     []string
		failed   []string
		failures = make(map[string]string)
	)
	for _, user := range users {
		subject, body := h.mailService.AdminCustomMailContent(input.Subject, input.Body)
		err := h.mailService.SendMail(h.config.MailSender.Email, user.Email, subject, body)
		if err != nil {
			failed = append(failed, user.Email)
			failures[user.Email] = err.Error()
		} else {
			sent = append(sent, user.Email)
		}
	}
	Success(c, http.StatusOK, "Mail send attempt complete", map[string]interface{}{
		"sent":     sent,
		"failed":   failed,
		"failures": failures,
	})
}
