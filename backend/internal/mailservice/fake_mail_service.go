package mailservice

import (
	"fmt"
	"kubecloud/internal/logger"
	"kubecloud/internal/metrics"
	"mime/multipart"
)

// FakeMailService overrides MailService methods for development purposes
type FakeMailService struct {
	metrics *metrics.Metrics
}

// NewFakeMailService creates a new fake mail service for development
func NewFakeMailService(metrics *metrics.Metrics) FakeMailService {
	return FakeMailService{
		metrics: metrics,
	}
}

func (service FakeMailService) SystemMail() string {
	return ""
}

func (service FakeMailService) MaxConcurrentSends() int {
	return 0
}

func (service FakeMailService) MaxAttachmentSizeInBytes() int64 {
	return 0
}

func (service FakeMailService) SendMailFromSystem(receiver, subject, body string, attachments ...Attachment) error {
	return service.SendMail(service.SystemMail(), receiver, subject, body, attachments...)
}

// SendMail overrides to track metrics without actually sending
func (service FakeMailService) SendMail(sender, receiver, subject, body string, attachments ...Attachment) error {
	if !IsValidEmail(receiver) {
		service.metrics.IncrementEmailFailed()
		return fmt.Errorf("email %v is not valid", receiver)
	}
	service.metrics.IncrementEmailSent()
	return nil
}

func (service FakeMailService) ParseAttachment(fh *multipart.FileHeader) (Attachment, error) {
	return Attachment{}, nil
}

// ResetPasswordMailContent displays OTP in a clean format
func (service FakeMailService) ResetPasswordMailContent(code int, timeout int, username string) (string, string) {
	logger.GetLogger().Info().Msgf("\n"+
		"╔══════════════════════════════════════════════════════════╗\n"+
		"║            RESET PASSWORD OTP CODE                       ║\n"+
		"╠══════════════════════════════════════════════════════════╣\n"+
		"║  User:    %-47s║\n"+
		"║  Code:    %-47d║\n"+
		"║  Expires: %d minutes                                      ║\n"+
		"╚══════════════════════════════════════════════════════════╝",
		username, code, timeout)
	return "", ""
}

// SignUpMailContent displays OTP in a clean format
func (service FakeMailService) SignUpMailContent(code int, timeout int, username string) (string, string) {
	logger.GetLogger().Info().Msgf("\n"+
		"╔══════════════════════════════════════════════════════════╗\n"+
		"║            SIGN UP VERIFICATION CODE                     ║\n"+
		"╠══════════════════════════════════════════════════════════╣\n"+
		"║  User:    %-47s║\n"+
		"║  Code:    %-47d║\n"+
		"║  Expires: %d minutes                                      ║\n"+
		"╚══════════════════════════════════════════════════════════╝",
		username, code, timeout)
	return "", ""
}

// WelcomeMailContent gets the email content for welcome mail
func (service FakeMailService) WelcomeMailContent(username string) (string, string) {
	logger.GetLogger().Info().Msgf("Welcome mail has been sent to %s", username)
	return "", ""
}

// InvoiceMailContent gets the email content for invoice mail
func (service FakeMailService) InvoiceMailContent(invoiceTotal float64, currency string, invoiceID int) (string, string) {
	logger.GetLogger().Info().Float64("invoiceTotal", invoiceTotal).
		Str("currency", currency).Int("invoiceID", invoiceID).
		Msgf("Invoice mail has been sent for invoice %d", invoiceID)

	return "", ""
}

// SystemAnnouncementMailBody gets the email content for system announcement mail
func (service FakeMailService) SystemAnnouncementMailBody(body string) string {
	logger.GetLogger().Info().Str("body", body).Msg("System announcement mail has been sent")
	return ""
}

// NotifyAdminsMailContent gets the email content for notifying admins about pending payment records
func (service FakeMailService) NotifyAdminsMailContent(recordsNumber int) (string, string) {
	logger.GetLogger().Info().Int("recordsNumber", recordsNumber).
		Msg("Notify admins mail has been sent")
	return "", ""
}
