package mailservice

import (
	_ "embed"
	"encoding/base64"
	"fmt"
	"io"
	"kubecloud/internal"
	"kubecloud/internal/metrics"
	"mime"
	"mime/multipart"
	"path/filepath"
	"slices"
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

// MailService struct hods all functionalities of mail service
type SendGridMailService struct {
	client              *sendgrid.Client
	metrics             *metrics.Metrics
	systemEmail         string
	systemHost          string
	maxConcurrentSends  int
	maxAttachmentSizeMB int64
}

// NewSendGridMailService creates new instance of sendgrid mail service
func NewSendGridMailService(mailConfigs internal.MailSender, systemHost string, metrics *metrics.Metrics) SendGridMailService {
	return SendGridMailService{
		client:              sendgrid.NewSendClient(mailConfigs.SendGridKey),
		metrics:             metrics,
		systemEmail:         mailConfigs.Email,
		systemHost:          systemHost,
		maxConcurrentSends:  mailConfigs.MaxConcurrentSends,
		maxAttachmentSizeMB: mailConfigs.MaxAttachmentSizeMB,
	}
}

func (service SendGridMailService) SystemMail() string {
	return service.systemEmail
}

func (service SendGridMailService) MaxConcurrentSends() int {
	return service.maxConcurrentSends
}

func (service SendGridMailService) MaxAttachmentSizeInBytes() int64 {
	return service.maxAttachmentSizeMB * 1024 * 1024
}

// SendMail sends verification mails
func (service SendGridMailService) SendMailFromSystem(receiver, subject, body string, attachments ...Attachment) error {
	return service.SendMail(service.systemEmail, receiver, subject, body, attachments...)
}

// SendMail sends verification mails
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

func (service SendGridMailService) ParseAttachment(fh *multipart.FileHeader) (Attachment, error) {
	if !isAttachmentAllowed(fh.Filename) {
		return Attachment{}, fmt.Errorf("file type not allowed for %s", fh.Filename)
	}

	maxFileSizeBytes := service.MaxAttachmentSizeInBytes()

	if fh.Size > maxFileSizeBytes {
		return Attachment{}, fmt.Errorf("file %s is too large: %d bytes (max %d bytes)", fh.Filename, fh.Size, maxFileSizeBytes)
	}

	file, err := fh.Open()
	if err != nil {
		return Attachment{}, fmt.Errorf("failed to open attachment file: %w", err)
	}
	defer file.Close()

	fileData, err := io.ReadAll(file)
	if err != nil {
		return Attachment{}, fmt.Errorf("failed to read attachment file: %w", err)
	}

	return Attachment{
		FileName: fh.Filename,
		Data:     fileData,
	}, nil
}

func isAttachmentAllowed(filename string) bool {
	allowedAttachmentTypes := []string{
		".pdf", ".doc", ".docx", ".txt", ".jpg", ".jpeg", ".png", ".gif", ".zip",
	}

	ext := strings.ToLower(filepath.Ext(filename))
	return slices.Contains(allowedAttachmentTypes, ext)
}

// ResetPasswordMailContent gets the email content for reset password
func (service SendGridMailService) ResetPasswordMailContent(code, timeout int, username string) (string, string) {
	subject := "Reset password"
	body := string(resetPassTemplate)

	body = strings.ReplaceAll(body, "-code-", fmt.Sprint(code))
	body = strings.ReplaceAll(body, "-time-", fmt.Sprint(timeout))
	body = strings.ReplaceAll(body, "-name-", cases.Title(language.Und).String(username))
	body = strings.ReplaceAll(body, "-host-", service.systemHost)

	return subject, body
}

// SignUpMailContent gets the email content for sign up
func (service SendGridMailService) SignUpMailContent(code int, timeout int, username string) (string, string) {
	subject := "Welcome to Mycelium Cloud 🎉"
	body := string(signUpTemplate)

	body = strings.ReplaceAll(body, "-code-", fmt.Sprint(code))
	body = strings.ReplaceAll(body, "-time-", fmt.Sprint(timeout))
	body = strings.ReplaceAll(body, "-name-", cases.Title(language.Und).String(username))
	body = strings.ReplaceAll(body, "-host-", service.systemHost)

	return subject, body
}

// WelcomeMailContent gets the email content for welcome mail
func (service SendGridMailService) WelcomeMailContent(username string) (string, string) {
	subject := "Welcome to Mycelium Cloud 🎉"
	body := string(welcomeMail)

	body = strings.ReplaceAll(body, "-name-", cases.Title(language.Und).String(username))
	body = strings.ReplaceAll(body, "-host-", service.systemHost)

	return subject, body
}

// InvoiceMailContent gets the email content for invoice mail
func (service SendGridMailService) InvoiceMailContent(invoiceTotal float64, currency string, invoiceID int) (string, string) {
	mailBody := "We hope this message finds you well. <br>"
	mailBody += fmt.Sprintf("Our records show that there is an outstanding invoice (%d) for %v %s associated with your account. ", invoiceID, invoiceTotal, currency)

	mailBody += "If you have already made the payment or need any assistance, "
	mailBody += "please don't hesitate to reach out to us. <br><br>"
	mailBody += "We appreciate your prompt attention to this matter and thank you for being a valued customer."

	subject := "Invoice Notification"
	return subject, mailBody

}

// SystemAnnouncementMailBody gets the email content for system announcement mail
func (service SendGridMailService) SystemAnnouncementMailBody(body string) string {
	template := string(systemAnnouncementMail)
	body = strings.ReplaceAll(body, "\n", "<br>")
	template = strings.ReplaceAll(template, "-body-", body)

	return template
}

// NotifyAdminsMailContent gets the email content for notifying admins about pending payment records
func (service SendGridMailService) NotifyAdminsMailContent(recordsNumber int) (string, string) {
	subject := "There're pending payment requests for you to settle"
	body := string(notifyPaymentRecordsMail)

	body = strings.ReplaceAll(body, "-records-", fmt.Sprint(recordsNumber))
	body = strings.ReplaceAll(body, "-host-", service.systemHost)

	return subject, body
}
