package mailsender

import (
	"context"
	"net/mail"
)

type MailRequest struct {
	From        string
	To          string
	Subject     string
	Body        string
	Attachments []Attachment
}
type Attachment struct {
	FileName string
	Data     []byte
}

type MailSender interface {
	Send(ctx context.Context, req MailRequest) error
}

func isValidEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}
