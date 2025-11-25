package mailsender

import (
	"context"
	"encoding/base64"
	"fmt"
	"mime"
	"path/filepath"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

type SendGridMailSender struct {
	client  *sendgrid.Client
}

func NewSendGridMailSender(sendGridKey string) SendGridMailSender {
	return SendGridMailSender{
		client:  sendgrid.NewSendClient(sendGridKey),
	}
}

func (s SendGridMailSender) Send(ctx context.Context, req MailRequest) error {
	from := mail.NewEmail("Mycelium Cloud", req.From)

	if !isValidEmail(req.To) {
		return fmt.Errorf("email %v is not valid", req.To)
	}

	to := mail.NewEmail("Mycelium Cloud User", req.To)

	message := mail.NewSingleEmail(from, req.Subject, to, "", req.Body)
	message.Content = []*mail.Content{
		mail.NewContent("text/html", req.Body),
	}

	for _, att := range req.Attachments {
		attachment := mail.NewAttachment()
		attachment = attachment.SetContent(base64.StdEncoding.EncodeToString(att.Data))
		attachment = attachment.SetType(mime.TypeByExtension(filepath.Ext(att.FileName)))
		attachment = attachment.SetFilename(att.FileName)
		attachment = attachment.SetDisposition("attachment")
		message = message.AddAttachment(attachment)
	}

	_, err := s.client.Send(message)

	if err != nil {
		return err
	}
	return nil
}
