package handlers

import (
	"errors"
	"fmt"
	"kubecloud/internal/core/models"
	"mime/multipart"
	"net/http"
	"strconv"
	"sync"
	"time"

	"kubecloud/internal/core/services"
	mailservice "kubecloud/internal/infrastructure/mailservice"
	"kubecloud/internal/infrastructure/notification"

	"github.com/gin-gonic/gin"
	"github.com/hashicorp/go-multierror"
	"github.com/rs/zerolog"
)

type AdminHandler struct {
	svc                    services.AdminService
	billingService         services.BillingService
	notificationDispatcher *notification.NotificationDispatcher
	mailService            mailservice.MailService
}

func NewAdminHandler(svc services.AdminService, billingService services.BillingService,
	notificationDispatcher *notification.NotificationDispatcher, mailService mailservice.MailService,
) AdminHandler {
	return AdminHandler{
		svc:                    svc,
		billingService:         billingService,
		notificationDispatcher: notificationDispatcher,
		mailService:            mailService,
	}
}

// GenerateVouchersInput holds all data needed when creating vouchers
type GenerateVouchersInput struct {
	Count       int     `json:"count" binding:"required,gt=0"`
	Value       float64 `json:"value" binding:"required,gt=0"`
	ExpireAfter int     `json:"expire_after_days" binding:"required,gt=0"`
}

// CreditRequestInput represents a request to credit a user's balance
type CreditRequestInput struct {
	AmountUSD float64 `json:"amount" binding:"required,gt=0"`
	Memo      string  `json:"memo" binding:"required,min=3,max=255"`
}

// CreditUserResponse holds the response data after crediting a user
type CreditUserResponse struct {
	User      string  `json:"user"`
	AmountUSD float64 `json:"amount"`
	Memo      string  `json:"memo"`
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
// @Success 200 {array} services.UserWithTFTBalance
// @Failure 500 {object} APIResponse
// @Security AdminMiddleware
// @Router /users [get]
// ListUsersHandler lists all users
func (h *AdminHandler) ListUsersHandler(c *gin.Context) {
	reqLog := requestLogger(c, "ListUsersHandler")

	usersWithBalance, err := h.svc.ListAllUsersIncludingUSDBalance()
	if err != nil {
		reqLog.Error().Err(err).Msg("failed to list all users")
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
func (h *AdminHandler) DeleteUsersHandler(c *gin.Context) {
	userID := c.Param("user_id")
	authUserID := c.GetInt("user_id")
	reqLog := requestLogger(c, "DeleteUsersHandler")

	if userID == "" {
		BadRequest(c, "User ID is required")
		return
	}

	id, err := strconv.Atoi(userID)
	if err != nil || id == 0 {
		reqLog.Error().Err(err).Msg("invalid user id to delete")
		BadRequest(c, "Invalid user ID")
		return
	}

	if id == authUserID {
		Forbidden(c, "Admins cannot delete their own account")
		return
	}

	err = h.svc.DeleteUserByID(id)
	if err != nil {
		if errors.Is(err, models.ErrUserNotFound) {
			NotFound(c, "User not found")
			return
		}
		reqLog.Error().Err(err).Msg("failed to delete user by id")
		InternalServerError(c)
		return
	}

	OK(c, "User is deleted successfully", nil)
}

// @Summary Generate vouchers
// @Description Generates a bulk of vouchers
// @Tags admin
// @ID generate-vouchers
// @Accept json
// @Produce json
// @Param body body GenerateVouchersInput true "Generate Vouchers Input"
// @Success 201 {object} APIResponse{data=[]models.Voucher}
// @Failure 400 {object} APIResponse "Invalid request format"
// @Failure 500 {object} APIResponse
// @Security AdminMiddleware
// @Router /vouchers/generate [post]
// GenerateVouchersHandler generates bulk of vouchers
func (h *AdminHandler) GenerateVouchersHandler(c *gin.Context) {
	var request GenerateVouchersInput
	reqLog := requestLogger(c, "GenerateVouchersHandler")
	adminID := c.GetInt("user_id")

	// check on request format
	if err := c.ShouldBindJSON(&request); err != nil {
		reqLog.Error().Err(err).Msg("Invalid request format")
		BadRequest(c, "Invalid request format")
		return
	}

	vouchers, err := h.svc.GenerateVouchers(request.Count, request.ExpireAfter, request.Value)
	if err != nil {
		reqLog.Error().Err(err).Msg("failed to generate vouchers")
		InternalServerError(c)
		return
	}

	notif := notification.BillingNotification(adminID).
		Success(fmt.Sprintf("%d vouchers generated successfully.", request.Count)).
		WithSubject("Vouchers Generated").
		WithStatus("succeeded").
		WithChannels(notification.ChannelUI).
		NoPersist().
		Build()

	if err := h.notificationDispatcher.Send(c.Request.Context(), notif); err != nil {
		reqLog.Error().Err(err).Msg("failed to send UI notification for voucher generation")
	}

	Created(c, "Vouchers are generated successfully", gin.H{
		"vouchers": vouchers,
	})
}

// @Summary List vouchers
// @Description Returns all vouchers in the system
// @Tags admin
// @ID list-vouchers
// @Accept json
// @Produce json
// @Success 200 {object} APIResponse{data=[]models.Voucher}
// @Failure 500 {object} APIResponse
// @Security AdminMiddleware
// @Router /vouchers [get]
// ListVouchersHandler returns all vouchers in system
func (h *AdminHandler) ListVouchersHandler(c *gin.Context) {
	reqLog := requestLogger(c, "ListVouchersHandler")
	vouchers, err := h.svc.ListAllVouchers()
	if err != nil {
		reqLog.Error().Err(err).Msg("failed to list all vouchers")
		InternalServerError(c)
		return
	}

	OK(c, "Vouchers are retrieved successfully", gin.H{
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
// @Success 202 {object} APIResponse{data=CreditUserResponse}
// @Failure 400 {object} APIResponse "Invalid request format or user ID"
// @Failure 500 {object} APIResponse
// @Security AdminMiddleware
// @Router /users/{user_id}/credit [post]
// CreditUserHandler lets admin credit specific user with money
func (h *AdminHandler) CreditUserHandler(c *gin.Context) {
	userID := c.Param("user_id")
	// get admin ID from middleware context
	reqLog := requestLogger(c, "CreditUserHandler")
	adminID := c.GetInt("user_id")

	var request CreditRequestInput
	// check on request format
	if err := c.ShouldBindJSON(&request); err != nil {
		BadRequest(c, "Invalid request format")
		return
	}

	id, err := strconv.Atoi(userID)
	if err != nil || id == 0 {
		reqLog.Error().Err(err).Msg("invalid user ID format")
		BadRequest(c, "Invalid user ID format")
		return
	}

	user, err := h.svc.GetUserByID(id)
	if err != nil {
		if errors.Is(err, models.ErrUserNotFound) {
			NotFound(c, "User is not found")
			return
		}
		reqLog.Error().Err(err).Msg("failed to retrieve user")
		InternalServerError(c)
		return
	}

	transaction := models.Transaction{
		UserID:    id,
		AdminID:   adminID,
		Amount:    request.AmountUSD,
		Memo:      request.Memo,
		CreatedAt: time.Now(),
	}

	if err := h.svc.CreditUserBalance(c.Request.Context(), transaction, &user); err != nil {
		reqLog.Error().Err(err).Msg("Failed to credit user balance")
		InternalServerError(c)
		return
	}

	if err := h.billingService.AfterUserGetCredit(c.Request.Context(), &user); err != nil {
		reqLog.Error().Err(err).Msg("Failed to credit user balance")
		InternalServerError(c)
		return
	}

	notif := notification.BillingNotification(adminID).
		Success(fmt.Sprintf("Admin %s has credited your account with %v$ successfully", user.Username, request.AmountUSD)).
		WithSubject("Admin Credited Your Account").
		WithStatus("succeeded").
		WithChannels(notification.ChannelUI).
		NoPersist().
		Build()

	if err := h.notificationDispatcher.Send(c.Request.Context(), notif); err != nil {
		reqLog.Error().Err(err).Msg("failed to send UI ")
	}

	Success(c, http.StatusCreated, fmt.Sprintf("User is credited with %v$ successfully", request.AmountUSD), CreditUserResponse{
		AmountUSD: request.AmountUSD,
		Memo:      request.Memo,
	})
}

// @Summary List transfer records
// @Description Returns all transfer records in the system
// @Tags admin
// @ID list-transfer-records
// @Accept json
// @Produce json
// @Success 200 {array} []services.TransferRecordsWithTFTAmount
// @Failure 500 {object} APIResponse
// @Security AdminMiddleware
// @Router /transfer-records [get]
// ListTransferRecordsHandler returns all transfer records in the system
func (h *AdminHandler) ListTransferRecordsHandler(c *gin.Context) {
	reqLog := requestLogger(c, "ListPendingRecordsHandler")

	transferRecordsResponse, err := h.svc.ListAllTransferRecordsWithTFTAmount()
	if err != nil {
		reqLog.Error().Err(err).Msg("failed to list all pending records")
		InternalServerError(c)
		return
	}

	OK(c, "Transfer records are retrieved successfully", gin.H{
		"transfer_records": transferRecordsResponse,
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
func (h *AdminHandler) SendMailToAllUsersHandler(c *gin.Context) {
	reqLog := requestLogger(c, "SendMailToAllUsersHandler")
	var input AdminMailInput
	if err := c.ShouldBind(&input); err != nil {
		BadRequest(c, "Invalid request format")
		return
	}

	var attachments []mailservice.Attachment
	if form, err := c.MultipartForm(); err == nil {
		if uploaded, ok := form.File["attachments"]; ok {
			reqLog.Info().Int("attachment_count", len(uploaded)).Msg("parsed email attachments")

			attachments, err = h.parseAttachments(uploaded, reqLog)
			if err != nil {
				reqLog.Error().Err(err).Msg("failed to parse attachments")
				InternalServerError(c)
				return
			}
		}
	}

	users, err := h.svc.ListAllUsers()
	if err != nil {
		reqLog.Error().Err(err).Msg("failed to list all users")
		InternalServerError(c)
		return
	}

	body := h.mailService.SystemAnnouncementMailBody(input.Body)
	emailConcurrencyLimiter := make(chan struct{}, h.mailService.MaxConcurrentSends())

	var (
		wg           sync.WaitGroup
		mu           sync.Mutex
		failedEmails []string
	)

	reqLog.Info().Int("attachment_count", len(attachments)).Msg("parsed email attachments")
	for _, user := range users {
		wg.Add(1)
		emailConcurrencyLimiter <- struct{}{}
		go func(user models.User) {
			defer wg.Done()
			defer func() { <-emailConcurrencyLimiter }()
			err := h.mailService.SendMailFromSystem(user.Email, input.Subject, body, attachments...)
			if err != nil {
				reqLog.Error().Err(err).Str("user_email", user.Email).Msg("failed to send mail to user")
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
		reqLog.Error().Msg("failed to send email to all users")
		InternalServerError(c)
		return
	}
	if responseData.FailedEmailsCount > 0 {
		OK(c, fmt.Sprintf("Mail sent to %d/%d users successfully", responseData.SuccessfulEmails, responseData.TotalUsers), responseData)
		return
	}
	OK(c, "Mail sent successfully to all users", responseData)
}

func (h *AdminHandler) parseAttachments(fileHeaders []*multipart.FileHeader, reqLogger *zerolog.Logger) ([]mailservice.Attachment, error) {
	if len(fileHeaders) == 0 {
		return nil, nil
	}

	var (
		mu       sync.Mutex
		multiErr *multierror.Error
		results  []mailservice.Attachment
		wg       sync.WaitGroup
	)

	wg.Add(len(fileHeaders))
	for _, fileHeader := range fileHeaders {
		go func(fh *multipart.FileHeader) {
			defer wg.Done()

			attachment, err := h.mailService.ParseAttachment(fh)
			if err != nil {
				reqLogger.Error().Err(err).Str("filename", fh.Filename).Msg("failed to parse attachment")
				mu.Lock()
				multiErr = multierror.Append(multiErr, err)
				mu.Unlock()
				return
			}

			mu.Lock()
			results = append(results, attachment)
			mu.Unlock()
		}(fileHeader)
	}

	wg.Wait()
	return results, multiErr.ErrorOrNil()
}

// @Summary Drain user balance
// @Description Drains a specific user's balance to the system account
// @Tags admin
// @ID drain-user
// @Accept json
// @Produce json
// @Param user_id path string true "User ID"
// @Success 202 {object} APIResponse "User balance drain initiated"
// @Failure 400 {object} APIResponse "Invalid user ID"
// @Failure 500 {object} APIResponse
// @Security AdminMiddleware
// @Router /users/{user_id}/drain [post]
// DrainUserHandler drains a specific user's balance to the system account
func (h *AdminHandler) DrainUserHandler(c *gin.Context) {
	userID := c.Param("user_id")
	reqLog := requestLogger(c, "DrainUserHandler")

	id, err := strconv.Atoi(userID)
	if err != nil || id == 0 {
		reqLog.Error().Err(err).Msg("invalid user ID format")
		BadRequest(c, "Invalid user ID format")
		return
	}

	if err := h.svc.AsyncDrainUserUSD(id); err != nil {
		reqLog.Error().Err(err).Msg("failed to drain user balance")
		InternalServerError(c)
		return
	}

	Accepted(c, "User balance drain initiated, transfer in progress", nil)
}

// @Summary Drain all users' balances
// @Description Drains all users' balances to the system account
// @Tags admin
// @ID drain-all-users
// @Accept json
// @Produce json
// @Success 202 {object} APIResponse "All users' balance drain initiated"
// @Failure 500 {object} APIResponse
// @Security AdminMiddleware
// @Router /users/drain-all [post]
// DrainAllUsersHandler drains all users' balances to the system account
func (h *AdminHandler) DrainAllUsersHandler(c *gin.Context) {
	reqLog := requestLogger(c, "DrainAllUsersHandler")

	if err := h.svc.AsyncDrainAllUsersUSD(); err != nil {
		reqLog.Error().Err(err).Msg("failed to drain all users' balances")
		InternalServerError(c)
		return
	}

	Accepted(c, "All users' balance drain initiated, transfers in progress", nil)
}
