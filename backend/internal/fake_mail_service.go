package internal

import (
	"fmt"
	"kubecloud/internal/logger"
	"kubecloud/internal/metrics"
)

// FakeMailService overrides MailService methods for development purposes
type FakeMailService struct {
	metrics *metrics.Metrics
}

// NewFakeMailService creates a new fake mail service for development
func NewFakeMailService(metrics *metrics.Metrics) FakeMailService {
	logger.GetLogger().Info().Msg("Dev Mode: Using FakeMailService (emails will be displayed in console)")
	return FakeMailService{
		metrics: metrics,
	}
}

// SendMail overrides to validate email and track metrics without actually sending
func (service FakeMailService) SendMail(sender, receiver, subject, body string, attachments ...Attachment) error {
	if !IsValidEmail(receiver) {
		service.metrics.IncrementEmailFailed()
		return fmt.Errorf("email %v is not valid", receiver)
	}
	service.metrics.IncrementEmailSent()
	return nil
}

// ResetPasswordMailContent displays OTP in a clean format
func (service FakeMailService) ResetPasswordMailContent(code int, timeout int, username, host string) (string, string) {
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
func (service FakeMailService) SignUpMailContent(code int, timeout int, username, host string) (string, string) {
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

func (service FakeMailService) WelcomeMailContent(username, host string) (string, string) {
	logger.GetLogger().Info().Msgf("\n"+
		"╔══════════════════════════════════════════════════════════╗\n"+
		"║               WELCOME TO KUBECLOUD                       ║\n"+
		"╠══════════════════════════════════════════════════════════╣\n"+
		"║  User:    %-47s║\n"+
		"╚══════════════════════════════════════════════════════════╝",
		username)
	return "", ""
}
func (service FakeMailService) NotifyAdminsMailContent(recordsNumber int, host string) (string, string) {
	logger.GetLogger().Info().Msgf("\n"+
		"╔══════════════════════════════════════════════════════════╗\n"+
		"║          ADMIN NOTIFICATION: PENDING RECORDS             ║\n"+
		"╠══════════════════════════════════════════════════════════╣\n"+
		"║  There're pending payment requests for you to settle     ║\n"+
		"║  Pending Records: %-34d║\n"+
		"║  Host:           %-34s║\n"+
		"╚══════════════════════════════════════════════════════════╝",
		recordsNumber, host)
	return "", ""
}

func (service FakeMailService) InvoiceMailContent(invoiceTotal float64, currency string, invoiceID int) (string, string) {
	logger.GetLogger().Info().Msgf("\n"+
		"╔══════════════════════════════════════════════════════════╗\n"+
		"║               INVOICE GENERATED                          ║\n"+
		"╠══════════════════════════════════════════════════════════╣\n"+
		"║  Invoice ID: %-42d║\n"+
		"║  Total:      %-42.2f %s║\n"+
		"╚══════════════════════════════════════════════════════════╝",
		invoiceID, invoiceTotal, currency)
	return "", ""
}

func (service FakeMailService) SystemAnnouncementMailBody(body string) string {
	logger.GetLogger().Info().Msgf("\n"+
		"╔══════════════════════════════════════════════════════════╗\n"+
		"║               SYSTEM ANNOUNCEMENT                        ║\n"+
		"╠══════════════════════════════════════════════════════════╣\n"+
		"║  %s\n"+
		"╚══════════════════════════════════════════════════════════╝",
		body)
	return ""
}
