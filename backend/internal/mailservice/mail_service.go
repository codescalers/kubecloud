package mailservice

import (
	"net/mail"
)

// MailService defines the contract for mail services
type MailService interface {
	SystemMail() string
	MaxConcurrentSends() int
	MaxAttachmentSizeInBytes() int64

	SendMailFromSystem(receiver, subject, body string, attachments ...Attachment) error
	SendMail(sender, receiver, subject, body string, attachments ...Attachment) error

	ResetPasswordMailContent(code int, timeout int, username string) (string, string)
	SignUpMailContent(code int, timeout int, username string) (string, string)
	WelcomeMailContent(username string) (string, string)
	InvoiceMailContent(invoiceTotal float64, currency string, invoiceID int) (string, string)
	SystemAnnouncementMailBody(body string) string
	NotifyAdminsMailContent(recordsNumber int) (string, string)
}

// IsValidEmail checks if the provided email is valid
func IsValidEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}
