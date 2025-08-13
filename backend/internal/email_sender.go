package internal

import (
	_ "embed"
	"encoding/base64"
	"fmt"
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

// MailService struct hods all functionalities of mail service
type MailService struct {
	client *sendgrid.Client
}

type Attachment struct {
	FileName string
	Data     []byte
}

// NewMailService creates new instance of mail service
func NewMailService(sendGridKey string) MailService {
	return MailService{
		client: sendgrid.NewSendClient(sendGridKey),
	}
}

// SendMail sends verification mails
func (service *MailService) SendMail(sender, receiver, subject, body string, attachments ...Attachment) error {
	from := mail.NewEmail("KubeCloud", sender)

	if !isValidEmail(receiver) {
		return fmt.Errorf("email %v is not valid", receiver)
	}

	to := mail.NewEmail("KubeCloud User", receiver)

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

	return err
}

// ResetPasswordMailContent gets the email content for reset password
func (service *MailService) ResetPasswordMailContent(code int, timeout int, username, host string) (string, string) {
	subject := "Reset password"
	body := string(resetPassTemplate)

	body = strings.ReplaceAll(body, "-code-", fmt.Sprint(code))
	body = strings.ReplaceAll(body, "-time-", fmt.Sprint(timeout))
	body = strings.ReplaceAll(body, "-name-", cases.Title(language.Und).String(username))
	body = strings.ReplaceAll(body, "-host-", host)

	return subject, body
}

// WelcomeMailContent gets the email content for welcome messages
func (service *MailService) WelcomeMailContent(username, host string) (string, string) {
	subject := "Welcome to KubeCloud 🎉"
	body := string(welcomeMail)

	body = strings.ReplaceAll(body, "-name-", cases.Title(language.Und).String(username))
	body = strings.ReplaceAll(body, "-host-", host)

	return subject, body
}

// SignUpMailContent gets the email content for sign up
func (service *MailService) SignUpMailContent(code int, timeout int, username, host string) (string, string) {
	subject := "Welcome to KubeCloud 🎉"
	body := string(signUpTemplate)

	body = strings.ReplaceAll(body, "-code-", fmt.Sprint(code))
	body = strings.ReplaceAll(body, "-time-", fmt.Sprint(timeout))
	body = strings.ReplaceAll(body, "-name-", cases.Title(language.Und).String(username))
	body = strings.ReplaceAll(body, "-host-", host)

	return subject, body
}

// NotifyAdminsMailContent gets the content for notifying admins
func (service *MailService) NotifyAdminsMailContent(recordsNumber int, host string) (string, string) {
	subject := "There're pending payment requests for you to settle"
	body := string(notifyPaymentRecordsMail)

	body = strings.ReplaceAll(body, "-records-", fmt.Sprint(recordsNumber))
	body = strings.ReplaceAll(body, "-host-", host)

	return subject, body
}

func (service *MailService) InvoiceMailContent(invoiceTotal float64, currency string, invoiceID int) (string, string) {
	mailBody := "We hope this message finds you well. <br>"
	mailBody += fmt.Sprintf("Our records show that there is an outstanding invoice (%d) for %v %s associated with your account. ", invoiceID, invoiceTotal, currency)

	mailBody += "If you have already made the payment or need any assistance, "
	mailBody += "please don't hesitate to reach out to us. <br><br>"
	mailBody += "We appreciate your prompt attention to this matter and thank you for being a valued customer."

	subject := "Invoice Notification"
	return subject, mailBody

}

func (service *MailService) SystemAnnouncementMailBody(subject, body string) string {
	template := string(systemAnnouncementMail)
	body = strings.ReplaceAll(body, "\n", "<br>")
	template = strings.ReplaceAll(template, "-subject-", subject)
	template = strings.ReplaceAll(template, "-body-", body)

	return template
}
