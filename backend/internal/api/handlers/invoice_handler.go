package handlers

import (
	"fmt"
	"kubecloud/internal/core/services"
	"net/http"
	"strconv"

	"kubecloud/internal/infrastructure/logger"
	"kubecloud/internal/infrastructure/mailservice"

	"github.com/gin-gonic/gin"
)

type InvoiceHandler struct {
	svc         services.InvoiceService
	mailService mailservice.MailService
}

func NewInvoiceHandler(svc services.InvoiceService, mailService mailservice.MailService) InvoiceHandler {
	return InvoiceHandler{
		svc:         svc,
		mailService: mailService,
	}
}

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
func (h *InvoiceHandler) ListAllInvoicesHandler(c *gin.Context) {
	reqLog := requestLogger(c, "ListAllInvoicesHandler")
	invoices, err := h.svc.ListInvoices()
	if err != nil {
		reqLog.Error().Err(err).Msg("failed to retrieve invoices")
		auditLogFromContext(c, logger.AuditActionInvoiceList, logger.AuditSeverityWarning, map[string]any{
			"reason": err.Error(),
		})
		InternalServerError(c)
		return
	}
	auditLogFromContext(c, logger.AuditActionInvoiceList, logger.AuditSeverityInfo, map[string]any{
		"invoices_length": len(invoices),
	})

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
func (h *InvoiceHandler) ListUserInvoicesHandler(c *gin.Context) {
	userID := c.GetInt("user_id")
	reqLog := requestLogger(c, "ListUserInvoicesHandler")

	invoices, err := h.svc.ListUserInvoices(userID)
	if err != nil {
		reqLog.Error().Err(err).Msg("failed to retrieve user invoices")
		auditLogFromContext(c, logger.AuditActionInvoiceList, logger.AuditSeverityWarning, map[string]any{
			"reason":  err.Error(),
			"user_id": userID,
		})
		InternalServerError(c)
		return
	}

	auditLogFromContext(c, logger.AuditActionInvoiceList, logger.AuditSeverityInfo, map[string]any{
		"invoices_length": len(invoices),
		"user_id":         userID,
	})

	OK(c, "Invoices are retrieved successfully", gin.H{
		"invoices": invoices,
	})
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
func (h *InvoiceHandler) DownloadInvoiceHandler(c *gin.Context) {
	userID := c.GetInt("user_id")
	reqLog := requestLogger(c, "DownloadInvoiceHandler")

	invoiceID := c.Param("invoice_id")
	if invoiceID == "" {
		auditLogFromContext(c, logger.AuditActionInvoiceDownload, logger.AuditSeverityWarning, map[string]any{
			"reason":  "invoice_id_required",
			"user_id": userID,
		})
		BadRequest(c, "Invoice ID is required")
		return
	}

	logWithInvoice := reqLog.With().Str("invoice_id", invoiceID).Logger()
	reqLog = &logWithInvoice

	id, err := strconv.Atoi(invoiceID)
	if err != nil {
		reqLog.Error().Err(err).Msg("failed to parse invoice ID")
		auditLogFromContext(c, logger.AuditActionInvoiceDownload, logger.AuditSeverityWarning, map[string]any{
			"reason":  "failed_to_parse_invoice_id",
			"user_id": userID,
		})
		BadRequest(c, "Invalid invoice ID")
		return
	}

	invoice, err := h.svc.GetInvoiceByID(id, userID)
	if err != nil {
		reqLog.Error().Err(err).Msg("failed to retrieve invoice")
		auditLogFromContext(c, logger.AuditActionInvoiceDownload, logger.AuditSeverityWarning, map[string]any{
			"reason":  "failed_to_retrieve_invoice",
			"user_id": userID,
		})
		NotFound(c, "Invoice is not found")
		return
	}

	if userID != invoice.UserID {
		auditLogFromContext(c, logger.AuditActionInvoiceDownload, logger.AuditSeverityWarning, map[string]any{
			"reason":  "user_not_authorized_to_download_invoice",
			"user_id": userID,
		})
		Forbidden(c, "User is not authorized to download this invoice")
		return
	}
	auditLogFromContext(c, logger.AuditActionInvoiceDownload, logger.AuditSeverityInfo, map[string]any{
		"invoice_id": invoice.ID,
		"user_id":    userID,
	})

	c.Writer.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fmt.Sprintf("invoice-%d-%d.pdf", invoice.UserID, invoice.ID)))
	c.Writer.Header().Set("Content-Type", "application/pdf")
	c.Data(http.StatusOK, "application/pdf", invoice.FileData)
}
