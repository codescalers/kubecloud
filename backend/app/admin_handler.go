package app

import (
	"context"
	"errors"
	"fmt"
	"io"
	"kubecloud/internal"
	"kubecloud/internal/activities"
	"kubecloud/models"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hashicorp/go-multierror"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

type UserResponse struct {
	models.User
	Balance float64 `json:"balance"` // USD balance
}

// GenerateVouchersInput holds all data needed when creating vouchers
type GenerateVouchersInput struct {
	Count       int     `json:"count" validate:"required,gt=0"`
	Value       float64 `json:"value" validate:"required,gt=0"`
	ExpireAfter int     `json:"expire_after_days" validate:"required,gt=0"`
}

// CreditRequestInput represents a request to credit a user's balance
type CreditRequestInput struct {
	AmountUSD float64 `json:"amount" validate:"required,gt=0"`
	Memo      string  `json:"memo" validate:"required,min=3,max=255"`
}

// CreditUserResponse holds the response data after crediting a user
type CreditUserResponse struct {
	User      string  `json:"user"`
	AmountUSD float64 `json:"amount"`
	Memo      string  `json:"memo"`
}

type PendingRecordsResponse struct {
	models.PendingRecord
	USDAmount            float64 `json:"usd_amount"`
	TransferredUSDAmount float64 `json:"transferred_usd_amount"`
}

// AdminMailInput represents the form data for sending emails to all users
type AdminMailInput struct {
	Subject     string                  `form:"subject" binding:"required"`
	Body        string                  `form:"body" binding:"required"`
	Attachments []*multipart.FileHeader `form:"attachments"`
}

type SendMailResponse struct {
	TotalUsers        int      `json:"total_users"`
	SuccessfulEmails  int      `json:"successful_emails"`
	FailedEmailsCount int      `json:"failed_emails_count"`
	FailedEmails      []string `json:"failed_emails,omitempty"`
}

type MaintenanceModeStatus struct {
	Enabled bool `json:"enabled"`
}

// @Summary Get all users
// @Description Returns a list of all users
// @Tags admin
// @ID get-all-users
// @Accept json
// @Produce json
// @Success 200 {array} UserResponse
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
	var usersWithBalance []UserResponse
	const maxConcurrentBalanceFetches = 20
	balanceConcurrencyLimiter := make(chan struct{}, maxConcurrentBalanceFetches)

	var (
		wg       sync.WaitGroup
		mu       sync.Mutex
		multiErr *multierror.Error
	)

	for _, user := range users {
		wg.Add(1)
		balanceConcurrencyLimiter <- struct{}{}

		go func(user models.User) {
			defer wg.Done()
			defer func() { <-balanceConcurrencyLimiter }()

			balance, err := internal.GetUserBalanceUSDMillicent(h.substrateClient, user.Mnemonic)
			if err != nil {
				log.Error().Err(err).Int("user_id", user.ID).Msg("failed to get user balance")
				mu.Lock()
				multiErr = multierror.Append(multiErr, fmt.Errorf("failed to get balance for user %d: %w", user.ID, err))
				mu.Unlock()
				return
			}

			balanceUSD := internal.FromUSDMilliCentToUSD(balance)
			mu.Lock()
			usersWithBalance = append(usersWithBalance, UserResponse{
				User:    user,
				Balance: balanceUSD,
			})
			mu.Unlock()
		}(user)
	}

	wg.Wait()

	// Check if there were any errors during balance fetching
	if multiErr != nil {
		log.Error().Err(multiErr).Msg("errors occurred while fetching user balances")
		InternalServerError(c)
		return
	}

	Success(c, http.StatusOK, "Users are retrieved successfully", map[string]interface{}{
		"users": usersWithBalance,
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
		if errors.Is(err, gorm.ErrRecordNotFound) {
			Error(c, http.StatusNotFound, "User not found", "")
		} else {
			InternalServerError(c)
		}
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

	if err := internal.ValidateStruct(request); err != nil {
		Error(c, http.StatusBadRequest, "Validation failed", err.Error())
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

	if err := internal.ValidateStruct(request); err != nil {
		Error(c, http.StatusBadRequest, "Validation failed", err.Error())
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
		Amount:    request.AmountUSD,
		Memo:      request.Memo,
		CreatedAt: time.Now(),
	}

	wf, err := h.ewfEngine.NewWorkflow(activities.WorkflowAdminCreditBalance)
	if err != nil {
		log.Error().Err(err).Send()
		InternalServerError(c)
		return
	}

	if err := h.db.CreateTransaction(&transaction); err != nil {
		log.Error().Err(err).Msg("Failed to create credit transaction")
		InternalServerError(c)
		return
	}

	wf.State = map[string]interface{}{
		"user_id":       user.ID,
		"amount":        internal.FromUSDToUSDMillicent(request.AmountUSD),
		"mnemonic":      user.Mnemonic,
		"username":      user.Username,
		"transfer_mode": models.AdminCreditMode,
	}
	h.ewfEngine.RunAsync(context.Background(), wf)

	Success(c, http.StatusCreated, "Transaction is created successfully, Money transfer is in progress", CreditUserResponse{
		User:      user.Email,
		AmountUSD: request.AmountUSD,
		Memo:      request.Memo,
	})
}

// @Summary List pending records
// @Description Returns all pending records in the system
// @Tags admin
// @ID list-pending-records
// @Accept json
// @Produce json
// @Success 200 {array} PendingRecordsResponse
// @Failure 500 {object} APIResponse
// @Security AdminMiddleware
// @Router /pending-records [get]
// ListPendingRecordsHandler returns all pending records in the system
func (h *Handler) ListPendingRecordsHandler(c *gin.Context) {
	pendingRecords, err := h.db.ListAllPendingRecords()
	if err != nil {
		log.Error().Err(err).Msg("failed to list all pending records")
		InternalServerError(c)
		return
	}

	var pendingRecordsResponse []PendingRecordsResponse
	for _, record := range pendingRecords {
		usdAmount, err := internal.FromTFTtoUSDMillicent(h.substrateClient, record.TFTAmount)
		if err != nil {
			log.Error().Err(err).Msg("failed to convert tft to usd amount")
			InternalServerError(c)
			return
		}

		usdTransferredAmount, err := internal.FromTFTtoUSDMillicent(h.substrateClient, record.TransferredTFTAmount)
		if err != nil {
			log.Error().Err(err).Msg("failed to convert tft to usd transferred amount")
			InternalServerError(c)
			return
		}

		pendingRecordsResponse = append(pendingRecordsResponse, PendingRecordsResponse{
			PendingRecord:        record,
			USDAmount:            internal.FromUSDMilliCentToUSD(usdAmount),
			TransferredUSDAmount: internal.FromUSDMilliCentToUSD(usdTransferredAmount),
		})
	}

	Success(c, http.StatusOK, "Pending records are retrieved successfully", map[string]any{
		"pending_records": pendingRecordsResponse,
	})
}

// Only accessible by admins
// @Summary Send mail to all users
// @Description Allows admin to send a custom email to all users with optional file attachments. Returns detailed statistics about successful and failed email deliveries.
// @Tags admin
// @ID admin-mail-all-users
// @Accept multipart/form-data
// @Produce json
// @Param subject formData string true "Email subject"
// @Param body formData string true "Email body content"
// @Param attachments formData file false "Email attachments (multiple files allowed)"
// @Success 200 {object} APIResponse{data=SendMailResponse} "Email sending results with delivery statistics"
// @Failure 400 {object} APIResponse "Invalid request format"
// @Failure 500 {object} APIResponse "Internal server error"
// @Security AdminMiddleware
// @Router /users/mail [post]
func (h *Handler) SendMailToAllUsersHandler(c *gin.Context) {
	var input AdminMailInput
	if err := c.ShouldBind(&input); err != nil {
		Error(c, http.StatusBadRequest, "Invalid request format", err.Error())
		return
	}

	var attachments []internal.Attachment
	if form, err := c.MultipartForm(); err == nil {
		if uploaded, ok := form.File["attachments"]; ok {
			log.Info().Int("attachment_count", len(uploaded)).Msg("parsed email attachments")

			attachments, err = h.parseAttachments(uploaded)
			if err != nil {
				log.Error().Err(err).Msg("failed to parse attachments")
				InternalServerError(c)
				return
			}
		}
	}

	users, err := h.db.ListAllUsers()
	if err != nil {
		log.Error().Err(err).Msg("failed to list all users")
		InternalServerError(c)
		return
	}

	body := h.mailService.SystemAnnouncementMailBody(input.Body)

	emailConcurrencyLimiter := make(chan struct{}, h.config.MailSender.MaxConcurrentSends)

	var (
		wg           sync.WaitGroup
		mu           sync.Mutex
		failedEmails []string
	)

	log.Info().Int("attachment_count", len(attachments)).Msg("parsed email attachments")
	for _, user := range users {
		wg.Add(1)
		emailConcurrencyLimiter <- struct{}{}
		go func(user models.User) {
			defer wg.Done()
			defer func() { <-emailConcurrencyLimiter }()
			err := h.mailService.SendMail(h.config.MailSender.Email, user.Email, input.Subject, body, attachments...)
			if err != nil {
				log.Error().Err(err).Str("user_email", user.Email).Msg("failed to send mail to user")
				mu.Lock()
				failedEmails = append(failedEmails, user.Email)
				mu.Unlock()
			}
		}(user)
	}

	wg.Wait()

	totalUsers := len(users)
	responseData := SendMailResponse{
		TotalUsers:        totalUsers,
		SuccessfulEmails:  totalUsers - len(failedEmails),
		FailedEmailsCount: len(failedEmails),
	}

	if responseData.SuccessfulEmails == 0 {
		Error(c, http.StatusInternalServerError, "failed to send mail to all users", "")
		return
	}
	if responseData.FailedEmailsCount > 0 {
		Success(c, http.StatusOK, fmt.Sprintf("Mail sent to %d/%d users successfully", responseData.SuccessfulEmails, responseData.TotalUsers), responseData)
		return
	}
	Success(c, http.StatusOK, "Mail sent successfully to all users", responseData)
}

func (h *Handler) parseAttachments(fileHeaders []*multipart.FileHeader) ([]internal.Attachment, error) {
	if len(fileHeaders) == 0 {
		return nil, nil
	}

	allowedTypes := map[string]bool{
		".pdf": true, ".doc": true, ".docx": true, ".txt": true,
		".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".zip": true,
	}

	var (
		mu       sync.Mutex
		multiErr *multierror.Error
		results  []internal.Attachment
		wg       sync.WaitGroup
	)

	wg.Add(len(fileHeaders))
	for _, fileHeader := range fileHeaders {
		go func(fh *multipart.FileHeader) {
			defer wg.Done()

			ext := strings.ToLower(filepath.Ext(fh.Filename))
			if !allowedTypes[ext] {
				mu.Lock()
				multiErr = multierror.Append(multiErr, fmt.Errorf("file type %s not allowed for %s", ext, fh.Filename))
				mu.Unlock()
				return
			}

			maxFileSizeBytes := h.config.MailSender.MaxAttachmentSizeMB * 1024 * 1024

			if fh.Size > maxFileSizeBytes {
				mu.Lock()
				multiErr = multierror.Append(multiErr, fmt.Errorf("file %s is too large: %d bytes (max %d bytes)", fh.Filename, fh.Size, maxFileSizeBytes))
				mu.Unlock()
				return
			}

			file, err := fh.Open()
			if err != nil {
				log.Error().Err(err).Str("filename", fh.Filename).Msg("failed to open attachment file")
				mu.Lock()
				multiErr = multierror.Append(multiErr, err)
				mu.Unlock()
				return
			}
			defer file.Close()

			fileData, err := io.ReadAll(file)
			if err != nil {
				log.Error().Err(err).Str("filename", fh.Filename).Msg("failed to read attachment file")
				mu.Lock()
				multiErr = multierror.Append(multiErr, err)
				mu.Unlock()
				return
			}

			attachment := internal.Attachment{
				FileName: fh.Filename,
				Data:     fileData,
			}

			mu.Lock()
			results = append(results, attachment)
			mu.Unlock()
		}(fileHeader)
	}

	wg.Wait()
	return results, multiErr.ErrorOrNil()
}

// @Summary Set maintenance mode
// @Description Sets maintenance mode for the system
// @Tags admin
// @ID set-maintenance-mode
// @Accept json
// @Produce json
// @Param body body MaintenanceModeStatus true "Maintenance Mode Status"
// @Success 200 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Security AdminMiddleware
// @Router /system/maintenance/status [put]
// SetMaintenanceModeHandler sets maintenance mode for the system
func (h *Handler) SetMaintenanceModeHandler(c *gin.Context) {
	var request MaintenanceModeStatus

	// check on request format
	if err := c.ShouldBindJSON(&request); err != nil {
		log.Error().Err(err).Send()
		Error(c, http.StatusBadRequest, "Invalid request format", err.Error())
		return
	}

	if err := internal.ValidateStruct(request); err != nil {
		Error(c, http.StatusBadRequest, "Validation failed", err.Error())
		return
	}

	if err := h.redis.SetMaintenanceMode(c.Request.Context(), request.Enabled); err != nil {
		log.Error().Err(err).Send()
		InternalServerError(c)
		return
	}

	Success(c, http.StatusOK, "Maintenance mode is set successfully", nil)
}

// @Summary Get maintenance mode
// @Description Gets maintenance mode for the system
// @Tags admin
// @ID get-maintenance-mode
// @Accept json
// @Produce json
// @Success 200 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Security AdminMiddleware
// @Router /system/maintenance/status [get]
// GetMaintenanceModeHandler gets maintenance mode for the system
func (h *Handler) GetMaintenanceModeHandler(c *gin.Context) {
	enabled, err := h.redis.GetMaintenanceMode(c.Request.Context())
	if err != nil {
		log.Error().Err(err).Send()
		InternalServerError(c)
		return
	}

	Success(c, http.StatusOK, "Maintenance mode is retrieved successfully", MaintenanceModeStatus{
		Enabled: enabled,
	})
}
