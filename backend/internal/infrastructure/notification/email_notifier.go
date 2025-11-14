package notification

import (
	"bytes"
	"fmt"
	"html/template"
	mailservice "kubecloud/internal/infrastructure/mailservice"
	"kubecloud/internal/core/models"
	"os"
	"path/filepath"

	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

var emailTpls *template.Template

type EmailNotifier struct {
	mailService  mailservice.MailService
	templatesDir string
}

func NewEmailNotifier(mailService mailservice.MailService, templatesDir string) *EmailNotifier {
	return &EmailNotifier{
		mailService:  mailService,
		templatesDir: templatesDir,
	}
}

func (n *EmailNotifier) GetType() string {
	return ChannelEmail
}

func (n *EmailNotifier) GetStepName() string {
	return "send-email-notification"
}

func (n *EmailNotifier) ParseTemplates() error {
	if n.templatesDir == "" {

		n.templatesDir = os.Getenv("TEMPLATE_DIR")
		if n.templatesDir == "" {
			n.templatesDir = "./internal/mailservice/templates/notifications"
		}
	}

	tpl, err := template.ParseGlob(filepath.Join(n.templatesDir, "*.html"))
	if err != nil {
		return fmt.Errorf("failed to parse notification templates from directory %s: %w", n.templatesDir, err)
	}
	emailTpls = tpl
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
