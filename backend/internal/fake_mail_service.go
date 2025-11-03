package internal

import (
	"fmt"
	"kubecloud/internal/logger"
	"kubecloud/internal/metrics"
)

// FakeMailService overrides MailService methods for development purposes
type FakeMailService struct {
	*MailService
}

// NewFakeMailService creates a new fake mail service for development
func NewFakeMailService(metrics *metrics.Metrics) MailServiceInterface {
	logger.GetLogger().Info().Msg("Dev Mode: Using FakeMailService (emails will be displayed in console)")
	return &FakeMailService{
		MailService: &MailService{
			client:  nil,
			metrics: metrics,
		},
	}
}

// SendMail overrides to validate email and track metrics without actually sending
func (service *FakeMailService) SendMail(sender, receiver, subject, body string, attachments ...Attachment) error {
	if !IsValidEmail(receiver) {
		service.metrics.IncrementEmailFailed()
		return fmt.Errorf("email %v is not valid", receiver)
	}
	service.metrics.IncrementEmailSent()
	return nil
}

// ResetPasswordMailContent displays OTP in a clean format
func (service *FakeMailService) ResetPasswordMailContent(code int, timeout int, username, host string) (string, string) {
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
func (service *FakeMailService) SignUpMailContent(code int, timeout int, username, host string) (string, string) {
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
