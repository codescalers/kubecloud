package mailservice

import (
	"mime/multipart"
	"net/mail"
)

// MailService defines the contract for mail services
type MailService interface {
	SystemMail() string
	MaxConcurrentSends() int
	MaxAttachmentSizeInBytes() int64
	ParseAttachment(fh *multipart.FileHeader) (Attachment, error)

	SendMailFromSystem(receiver, subject, body string, attachments ...Attachment) error
	SendMail(sender, receiver, subject, body string, attachments ...Attachment) error

	ResetPasswordMailContent(code int, timeout int, username string) (string, string)
	SignUpMailContent(code int, timeout int, username string) (string, string)
	WelcomeMailContent(username string) (string, string)
	InvoiceMailContent(invoiceTotal float64, currency string, invoiceID int) (string, string)
	SystemAnnouncementMailBody(body string) string
	NotifyAdminsMailContent(recordsNumber int) (string, string)
	SendBulkSystemMails(receivers []string, body string, subject string, attachments ...Attachment) int
}

type Attachment struct {
	FileName string
	Data     []byte
}

// IsValidEmail checks if the provided email is valid
func IsValidEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}
