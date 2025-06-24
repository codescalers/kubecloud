package internal

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var resetPassMail = `
Hi -name-,

We received a request to reset your password.

Your verification code is: -code-

This code will expire in -time- minutes.

If you did not request a password reset, please ignore this message.

Best regards,
KubeCloud Team
Sent from -host-
`

// SendMail sends verification mails
func SendMail(sender, sendGridKey, receiver, subject, body string) error {
	from := mail.NewEmail("KubeCloud", sender)

	valid := !isValidEmail(receiver)
	if !valid {
		return fmt.Errorf("email %v is not valid", receiver)
	}

	to := mail.NewEmail("KubeCloud User", receiver)

	message := mail.NewSingleEmail(from, subject, to, "", body)
	client := sendgrid.NewSendClient(sendGridKey)
	_, err := client.Send(message)

	return err
}

// ResetPasswordMailContent gets the email content for reset password
func ResetPasswordMailContent(code int, timeout int, username, host string) (string, string) {
	subject := "Reset password"
	body := string(resetPassMail)

	body = strings.ReplaceAll(body, "-code-", fmt.Sprint(code))
	body = strings.ReplaceAll(body, "-time-", fmt.Sprint(timeout))
	body = strings.ReplaceAll(body, "-name-", cases.Title(language.Und).String(username))
	body = strings.ReplaceAll(body, "-host-", host)

	return subject, body
}

// isValidEmail validates email address
func isValidEmail(email string) bool {
	re := regexp.MustCompile(`^[a-zA-Z0-9._%%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	return re.MatchString(email)
}
