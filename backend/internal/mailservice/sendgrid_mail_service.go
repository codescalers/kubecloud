package mailservice

import (
	_ "embed"
	"encoding/base64"
	"fmt"
	"kubecloud/internal/logger"
	"kubecloud/internal/metrics"
	"mime"
	"path/filepath"
	"strings"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

//go:embed templates/reset_password.html
var resetPassTemplate []byte

//go:embed templates/welcome.html
var welcomeMail []byte

//go:embed templates/signup.html
var signUpTemplate []byte

//go:embed templates/pending_record_notification.html
var notifyPaymentRecordsMail []byte

//go:embed templates/system_announcement.html
var systemAnnouncementMail []byte

type Attachment struct {
	FileName string
	Data     []byte
}

// SendGridMailService provides functionalities for sending emails using SendGrid.
type SendGridMailService struct {
	client  *sendgrid.Client
	metrics *metrics.Metrics
}

// NewSendGridMailService creates new instance of SendGridMailService
func NewSendGridMailService(sendGridKey string, metrics *metrics.Metrics) SendGridMailService {
	logger.GetLogger().Info().Msg("Using SendGrid mail service")
	return SendGridMailService{
		client:  sendgrid.NewSendClient(sendGridKey),
		metrics: metrics,
	}
}

// SendMail sends mails
func (service SendGridMailService) SendMail(sender, receiver, subject, body string, attachments ...Attachment) error {
	from := mail.NewEmail("Mycelium Cloud", sender)

	if !IsValidEmail(receiver) {
		return fmt.Errorf("email %v is not valid", receiver)
	}

	to := mail.NewEmail("Mycelium Cloud User", receiver)

	message := mail.NewSingleEmail(from, subject, to, "", body)
	message.Content = []*mail.Content{
		mail.NewContent("text/html", body),
	}

	for _, att := range attachments {
		attachment := mail.NewAttachment()
		attachment = attachment.SetContent(base64.StdEncoding.EncodeToString(att.Data))
		attachment = attachment.SetType(mime.TypeByExtension(filepath.Ext(att.FileName)))
		attachment = attachment.SetFilename(att.FileName)
		attachment = attachment.SetDisposition("attachment")
		message = message.AddAttachment(attachment)
	}

	_, err := service.client.Send(message)

	if err != nil {
		service.metrics.IncrementEmailFailed()
		return err
	}
	service.metrics.IncrementEmailSent()
	return nil
}

// ResetPasswordMailContent gets the email content for reset password
func (service SendGridMailService) ResetPasswordMailContent(code int, timeout int, username, host string) (string, string) {
	subject := "Reset password"
	body := string(resetPassTemplate)

	body = strings.ReplaceAll(body, "-code-", fmt.Sprint(code))
	body = strings.ReplaceAll(body, "-time-", fmt.Sprint(timeout))
	body = strings.ReplaceAll(body, "-name-", cases.Title(language.Und).String(username))
	body = strings.ReplaceAll(body, "-host-", host)

	return subject, body
}

// SignUpMailContent gets the email content for sign up
func (service SendGridMailService) SignUpMailContent(code int, timeout int, username, host string) (string, string) {
	subject := "Welcome to Mycelium Cloud 🎉"
	body := string(signUpTemplate)

	body = strings.ReplaceAll(body, "-code-", fmt.Sprint(code))
	body = strings.ReplaceAll(body, "-time-", fmt.Sprint(timeout))
	body = strings.ReplaceAll(body, "-name-", cases.Title(language.Und).String(username))
	body = strings.ReplaceAll(body, "-host-", host)

	return subject, body
}
