package mailcontentformatter

import (
	"fmt"
	"kubecloud/internal/core/models"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type MailTextFormatter struct{}

func NewMailTextFormatter() MailTextFormatter {
	return MailTextFormatter{}
}

func (f MailTextFormatter) FormatResetPasswordMailContent(code int, timeout int, username string, systemHost string) (string, string) {
	subject := "Reset password"
	user := cases.Title(language.Und).String(username)
	body := fmt.Sprintf(
		"Hello %s,\n\nUse the code %d to reset your password. The code expires in %d minutes.\n\nIf you didn't request this, please ignore this email.\n%s",
		user, code, timeout, systemHost,
	)

	return subject, body
}

func (f MailTextFormatter) FormatSignUpMailContent(code int, timeout int, username string, systemHost string) (string, string) {
	subject := "Welcome to Mycelium Cloud"
	user := cases.Title(language.Und).String(username)
	body := fmt.Sprintf(
		"Welcome %s!\n\nYour verification code is %d. It expires in %d minutes.\n\nVisit %s to continue your signup.",
		user, code, timeout, systemHost,
	)

	return subject, body
}

func (f MailTextFormatter) FormatWelcomeMailContent(username string, systemHost string) (string, string) {
	subject := "Welcome to Mycelium Cloud"
	user := cases.Title(language.Und).String(username)
	body := fmt.Sprintf(
		"Hello %s,\n\nYour Mycelium Cloud account is ready. Sign in at %s to get started.",
		user, systemHost,
	)

	return subject, body
}

func (f MailTextFormatter) FormatInvoiceMailContent(invoiceTotal float64, currency string, invoiceID int) (string, string) {
	subject := "Invoice Notification"
	body := fmt.Sprintf(
		"We hope you're well.\n\nInvoice %d is outstanding for %.2f %s. If you already paid or need help, contact us.\n\nThank you for being a valued customer.",
		invoiceID, invoiceTotal, currency,
	)

	return subject, body
}

func (f MailTextFormatter) FormatSystemAnnouncementMailBody(body string) string {
	lines := strings.TrimSpace(body)
	if lines == "" {
		return ""
	}
	return fmt.Sprintf("System announcement from Mycelium Cloud:\n\n%s", lines)
}

func (f MailTextFormatter) FormatNotifyAdminsMailContent(recordsNumber int, systemHost string) (string, string) {
	subject := "Pending payment requests"
	body := fmt.Sprintf(
		"There are %d payment records pending settlement.\nPlease review them at %s.",
		recordsNumber, systemHost,
	)

	return subject, body
}

func (f MailTextFormatter) FormatNotificationMailContent(notification models.Notification) (string, string, error) {
	subject := notification.Payload["subject"]
	if subject == "" {
		subject = fmt.Sprintf("%s notification", cases.Title(language.Und).String(string(notification.Type)))
	}

	return subject, notification.String(), nil
}
