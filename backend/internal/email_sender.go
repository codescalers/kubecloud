package internal

import (
	_ "embed"
	"fmt"
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
var signupTemplate []byte

// MailService struct hods all functionalities of mail service
type MailService struct {
	client *sendgrid.Client
}

// NewMailService creates new instance of mail service
func NewMailService(sendGridKey string) MailService {
	return MailService{
		client: sendgrid.NewSendClient(sendGridKey),
	}
}

// SendMail sends verification mails
func (service *MailService) SendMail(sender, receiver, subject, body string) error {
	from := mail.NewEmail("KubeCloud", sender)

	if !isValidEmail(receiver) {
		return fmt.Errorf("email %v is not valid", receiver)
	}

	to := mail.NewEmail("KubeCloud User", receiver)

	message := mail.NewSingleEmail(from, subject, to, "", body)
	message.Content = []*mail.Content{
		mail.NewContent("text/html", body),
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
	body := string(signupTemplate)

	body = strings.ReplaceAll(body, "-code-", fmt.Sprint(code))
	body = strings.ReplaceAll(body, "-time-", fmt.Sprint(timeout))
	body = strings.ReplaceAll(body, "-name-", cases.Title(language.Und).String(username))
	body = strings.ReplaceAll(body, "-host-", host)

	return subject, body
}
