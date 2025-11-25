package mailcontentformatter

import (
	"fmt"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	_ "embed"
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

type MailHTMLFormatter struct {
	systemHost string
}

func NewMailHTMLFormatter(systemHost string) MailHTMLFormatter {
	return MailHTMLFormatter{
		systemHost: systemHost,
	}
}

func (f MailHTMLFormatter) FormatResetPasswordMailContent(code int, timeout int, username string) (string, string) {
	subject := "Reset password"
	body := string(resetPassTemplate)

	body = strings.ReplaceAll(body, "-code-", fmt.Sprint(code))
	body = strings.ReplaceAll(body, "-time-", fmt.Sprint(timeout))
	body = strings.ReplaceAll(body, "-name-", cases.Title(language.Und).String(username))
	body = strings.ReplaceAll(body, "-host-", f.systemHost)

	return subject, body
}

func (f MailHTMLFormatter) FormatSignUpMailContent(code int, timeout int, username string) (string, string) {
	subject := "Welcome to Mycelium Cloud 🎉"
	body := string(signUpTemplate)

	body = strings.ReplaceAll(body, "-code-", fmt.Sprint(code))
	body = strings.ReplaceAll(body, "-time-", fmt.Sprint(timeout))
	body = strings.ReplaceAll(body, "-name-", cases.Title(language.Und).String(username))
	body = strings.ReplaceAll(body, "-host-", f.systemHost)

	return subject, body
}

func (f MailHTMLFormatter) FormatWelcomeMailContent(username string) (string, string) {
	subject := "Welcome to Mycelium Cloud 🎉"
	body := string(welcomeMail)

	body = strings.ReplaceAll(body, "-name-", cases.Title(language.Und).String(username))
	body = strings.ReplaceAll(body, "-host-", f.systemHost)

	return subject, body
}

func (f MailHTMLFormatter) FormatInvoiceMailContent(invoiceTotal float64, currency string, invoiceID int) (string, string) {
	mailBody := "We hope this message finds you well. <br>"
	mailBody += fmt.Sprintf("Our records show that there is an outstanding invoice (%d) for %v %s associated with your account. ", invoiceID, invoiceTotal, currency)

	mailBody += "If you have already made the payment or need any assistance, "
	mailBody += "please don't hesitate to reach out to us. <br><br>"
	mailBody += "We appreciate your prompt attention to this matter and thank you for being a valued customer."

	subject := "Invoice Notification"
	return subject, mailBody
}

func (f MailHTMLFormatter) FormatSystemAnnouncementMailBody(body string) string {
	template := string(systemAnnouncementMail)
	body = strings.ReplaceAll(body, "\n", "<br>")
	template = strings.ReplaceAll(template, "-body-", body)

	return template
}

func (f MailHTMLFormatter) FormatNotifyAdminsMailContent(recordsNumber int) (string, string) {
	subject := "There're pending payment requests for you to settle"
	body := string(notifyPaymentRecordsMail)

	body = strings.ReplaceAll(body, "-records-", fmt.Sprint(recordsNumber))
	body = strings.ReplaceAll(body, "-host-", f.systemHost)

	return subject, body
}
