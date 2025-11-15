package notification

import (
	"bytes"
	"fmt"
	"kubecloud/internal/core/models"
	mailservice "kubecloud/internal/infrastructure/mailservice"

	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

type EmailNotifier struct {
	mailService mailservice.MailService
}

func NewEmailNotifier(mailService mailservice.MailService, _ string) *EmailNotifier {
	return &EmailNotifier{
		mailService: mailService,
	}
}

func (n *EmailNotifier) GetType() string {
	return ChannelEmail
}

// ParseTemplates is now a no-op since templates are embedded
func (n *EmailNotifier) ParseTemplates() error {
	return nil
}

func (n *EmailNotifier) Notify(notification models.Notification, receiver ...string) error {
	if len(receiver) < 1 {
		return fmt.Errorf("at least one email address is required: receiver")
	}
	if !mailservice.IsValidEmail(receiver[0]) {
		return fmt.Errorf("receiver email address must be valid")
	}

	from := mail.NewEmail("KubeCloud", n.mailService.SystemMail())
	receiverEmail := mail.NewEmail("KubeCloud User", receiver[0])

	tplName := string(notification.Type)

	var buf bytes.Buffer
	emailTpls := mailservice.GetEmailTemplates()
	if err := emailTpls.ExecuteTemplate(&buf, tplName, notification); err != nil {
		return fmt.Errorf("failed to execute notification template '%s': %w", tplName, err)
	}

	subject := notification.Payload["subject"]
	if subject == "" {
		subject = string(notification.Type) + " Notification"
	}

	message := mail.NewSingleEmail(from, subject, receiverEmail, "", buf.String())
	message.Content = []*mail.Content{
		mail.NewContent("text/html", buf.String()),
	}

	err := n.mailService.SendMailFromSystem(receiver[0], subject, buf.String())
	return err
}
