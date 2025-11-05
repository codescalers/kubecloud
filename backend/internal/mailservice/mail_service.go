package mailservice

import (
	"fmt"
	"net/mail"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// MailService defines the contract for mail services
type MailService interface {
	SendMail(sender, receiver, subject, body string, attachments ...Attachment) error
	ResetPasswordMailContent(code int, timeout int, username, host string) (string, string)
	SignUpMailContent(code int, timeout int, username, host string) (string, string)
}

// WelcomeMailContent gets the email content for welcome mail
func WelcomeMailContent(username, host string) (string, string) {
	subject := "Welcome to Mycelium Cloud 🎉"
	body := string(welcomeMail)

	body = strings.ReplaceAll(body, "-name-", cases.Title(language.Und).String(username))
	body = strings.ReplaceAll(body, "-host-", host)

	return subject, body
}

// InvoiceMailContent gets the email content for invoice mail
func InvoiceMailContent(invoiceTotal float64, currency string, invoiceID int) (string, string) {
	mailBody := "We hope this message finds you well. <br>"
	mailBody += fmt.Sprintf("Our records show that there is an outstanding invoice (%d) for %v %s associated with your account. ", invoiceID, invoiceTotal, currency)

	mailBody += "If you have already made the payment or need any assistance, "
	mailBody += "please don't hesitate to reach out to us. <br><br>"
	mailBody += "We appreciate your prompt attention to this matter and thank you for being a valued customer."

	subject := "Invoice Notification"
	return subject, mailBody

}

// SystemAnnouncementMailBody gets the email content for system announcement mail
func SystemAnnouncementMailBody(body string) string {
	template := string(systemAnnouncementMail)
	body = strings.ReplaceAll(body, "\n", "<br>")
	template = strings.ReplaceAll(template, "-body-", body)

	return template
}

// NotifyAdminsMailContent gets the email content for notifying admins about pending payment records
func NotifyAdminsMailContent(recordsNumber int, host string) (string, string) {
	subject := "There're pending payment requests for you to settle"
	body := string(notifyPaymentRecordsMail)

	body = strings.ReplaceAll(body, "-records-", fmt.Sprint(recordsNumber))
	body = strings.ReplaceAll(body, "-host-", host)

	return subject, body
}

// IsValidEmail checks if the provided email is valid
func IsValidEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}
