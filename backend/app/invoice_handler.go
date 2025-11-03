package app

import (
	"context"
	"errors"
	"fmt"
	"kubecloud/internal"
	"kubecloud/models"
	"net/http"
	"strconv"

	"time"

	"kubecloud/internal/logger"
	mailservice "kubecloud/internal/mailservice"

	"github.com/gin-gonic/gin"
)

// @Summary Get all invoices
// @Description Returns a list of all invoices
// @Tags admin
// @ID get-all-invoices
// @Accept json
// @Produce json
// @Success 200 {object} APIResponse{data=[]models.Invoice}
// @Failure 500 {object} APIResponse
// @Security AdminMiddleware
// @Router /invoices [get]
// ListAllInvoicesHandler lists all invoices in system
func (h *Handler) ListAllInvoicesHandler(c *gin.Context) {
	reqLog := requestLogger(c, "ListAllInvoicesHandler")
	invoices, err := h.db.ListInvoices()
	if err != nil {
		reqLog.Error().Err(err).Msg("failed to retrieve invoices")
		InternalServerError(c)
		return
	}

	OK(c, "Invoices are retrieved successfully", gin.H{
		"invoices": invoices,
	})

}

// @Summary Get invoices
// @Description Returns a list of invoices for a user
// @Tags invoices
// @ID get-invoices
// @Accept json
// @Produce json
// @Success 200 {object} APIResponse{data=[]models.Invoice}
// @Failure 500 {object} APIResponse
// @Security UserMiddleware
// @Router /user/invoice [get]
// ListUserInvoicesHandler lists user invoices by its ID
func (h *Handler) ListUserInvoicesHandler(c *gin.Context) {
	userID := c.GetInt("user_id")
	reqLog := requestLogger(c, "ListUserInvoicesHandler")

	invoices, err := h.db.ListUserInvoices(userID)
	if err != nil {
		reqLog.Error().Err(err).Msg("failed to retrieve user invoices")
		InternalServerError(c)
		return
	}

	OK(c, "Invoices are retrieved successfully", gin.H{
		"invoices": invoices,
	})
}

func (h *Handler) MonthlyInvoicesHandler(ctx context.Context) {
	baseLog := logger.ForOperation("invoices", "monthly_invoices")
	var lastProcessedMonth time.Month
	var lastProcessedYear int

	for {
		now := time.Now()
		monthLastDay := time.Date(now.Year(), now.Month()+1, 1, 0, 0, 0, 0, now.Location()).AddDate(0, 0, -1)

		// Calculate sleep duration until month-end
		var sleepDuration time.Duration
		if now.Day() != monthLastDay.Day() {
			sleepDuration = monthLastDay.Sub(now)
		}

		//check if invoice for this month and year is already processed
		if now.Month() == lastProcessedMonth && now.Year() == lastProcessedYear {
			// Already processed, sleep until the first day of the next month to avoid running multiple times on the last day
			nextMonth := time.Date(now.Year(), now.Month()+1, 1, 0, 5, 0, 0, now.Location())
			sleepDuration = nextMonth.Sub(now)
		}

		// Sleep with context awareness (single select)
		if sleepDuration > 0 {
			select {
			case <-ctx.Done():
				return
			case <-time.After(sleepDuration):
			}
			continue
		}
		// Process invoices (we're on month-end and haven't processed yet)
		users, err := h.db.ListAllUsers()
		if err != nil {
			baseLog.Error().Err(err).Msg("failed to retrieve users for invoice creation")
			continue
		}

		for _, user := range users {
			if err = h.createUserInvoice(user); err != nil {
				baseLog.Error().Err(err).Int("user_id", user.ID).Msg("failed to create invoice for user")
			}
		}
		//update last processed month and year
		lastProcessedMonth = now.Month()
		lastProcessedYear = now.Year()
	}
}

// @Summary Download invoice
// @Description Downloads an invoice by ID
// @Tags invoices
// @ID download-invoice
// @Accept json
// @Produce octet-stream
// @Param invoice_id path string true "Invoice ID"
// @Success 200 {file} Invoice
// @Failure 404 {object} APIResponse "Invoice is not found"
// @Failure 500 {object} APIResponse
// @Security UserMiddleware
// @Router /user/invoice/{invoice_id} [get]
func (h *Handler) DownloadInvoiceHandler(c *gin.Context) {
	userID := c.GetInt("user_id")
	reqLog := requestLogger(c, "DownloadInvoiceHandler")

	invoiceID := c.Param("invoice_id")
	if invoiceID == "" {
		BadRequest(c, "Invoice ID is required")
		return
	}

	logWithInvoice := reqLog.With().Str("invoice_id", invoiceID).Logger()
	reqLog = &logWithInvoice

	id, err := strconv.Atoi(invoiceID)
	if err != nil {
		reqLog.Error().Err(err).Msg("failed to parse invoice ID")
		BadRequest(c, "Invalid invoice ID")
		return
	}

	invoice, err := h.db.GetInvoice(id)
	if err != nil {
		reqLog.Error().Err(err).Msg("failed to retrieve invoice")
		NotFound(c, "Invoice is not found")
		return
	}

	// Creating pdf for invoice if it doesn't have it
	if len(invoice.FileData) == 0 {
		user, err := h.db.GetUserByID(userID)
		if err != nil {
			reqLog.Error().Err(err).Msg("failed to retrieve user")
			InternalServerError(c)
			return
		}

		pdfContent, err := internal.CreateInvoicePDF(invoice, user, h.config.Invoice)
		if err != nil {
			reqLog.Error().Err(err).Msg("failed to create invoice PDF")
			InternalServerError(c)
			return
		}

		invoice.FileData = pdfContent
		if err := h.db.UpdateInvoicePDF(id, invoice.FileData); err != nil {
			reqLog.Error().Err(err).Msg("failed to update invoice PDF")
			InternalServerError(c)
			return
		}
	}

	if userID != invoice.UserID {
		Forbidden(c, "User is not authorized to download this invoice")
		return
	}

	c.Writer.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fmt.Sprintf("invoice-%d-%d.pdf", invoice.UserID, invoice.ID)))
	c.Writer.Header().Set("Content-Type", "application/pdf")
	c.Data(http.StatusOK, "application/pdf", invoice.FileData)

}

func (h *Handler) createUserInvoice(user models.User) error {
	records, err := h.db.ListUserNodes(user.ID)
	if err != nil {
		return err
	}

	if len(records) == 0 {
		return nil
	}

	now := time.Now()

	var nodeItems []models.NodeItem
	var totalInvoiceCostUSD float64

	for _, record := range records {
		billReports, err := internal.ListContractBillReportsPerMonth(h.graphqlClient, record.ContractID, now)
		if err != nil {
			return err
		}

		totalAmountTFT, err := internal.AmountBilledPerMonth(billReports)
		if err != nil {
			return err
		}
		totalAmountUSDMillicent, err := internal.FromTFTtoUSDMillicent(h.substrateClient, totalAmountTFT)
		if err != nil {
			return err
		}
		rentRecordStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		if record.CreatedAt.After(rentRecordStart) {
			rentRecordStart = record.CreatedAt
		}

		var totalHours int
		cancellationDate, err := internal.GetRentContractCancellationDate(h.firesquidClient, record.ContractID)

		if errors.Is(err, internal.ErrorEventsNotFound) {
			totalHours = GetHoursOfGivenPeriod(rentRecordStart, time.Now())
		} else if err != nil {
			return err
		} else {
			totalHours = GetHoursOfGivenPeriod(rentRecordStart, cancellationDate)
		}

		totalAmountUSD := internal.FromUSDMilliCentToUSD(totalAmountUSDMillicent)

		nodeItems = append(nodeItems, models.NodeItem{
			NodeID:        record.NodeID,
			ContractID:    record.ContractID,
			RentCreatedAt: rentRecordStart,
			PeriodInHours: float64(totalHours),
			Cost:          totalAmountUSD,
		})
		totalInvoiceCostUSD += totalAmountUSD

	}

	invoice := models.Invoice{
		UserID:    user.ID,
		Total:     totalInvoiceCostUSD,
		Nodes:     nodeItems,
		Tax:       0, //TODO:
		CreatedAt: time.Now(),
	}

	file, err := internal.CreateInvoicePDF(invoice, user, h.config.Invoice)
	if err != nil {
		return err
	}

	invoice.FileData = file
	if err = h.db.CreateInvoice(&invoice); err != nil {
		return err
	}

	subject, body := mailservice.InvoiceMailContent(totalInvoiceCostUSD, h.config.Currency, invoice.ID)
	err = h.mailService.SendMail(h.config.MailSender.Email, user.Email, subject, body, mailservice.Attachment{
		FileName: fmt.Sprintf("invoice-%d-%d.pdf", invoice.UserID, invoice.ID),
		Data:     invoice.FileData,
	})
	if err != nil {
		return err
	}
	return nil
}

func GetHoursOfGivenPeriod(startDate, endDate time.Time) int {
	// Calculate the duration between the first day of the month and the specific date
	duration := endDate.Sub(startDate)
	// Convert the duration to hours
	hours := int(duration.Hours())
	return hours
}
