package mailservice

import (
	cfg "kubecloud/internal/config"
	mailcontentformatter "kubecloud/internal/infrastructure/mailservice/mail_content_formatter"
	mailsender "kubecloud/internal/infrastructure/mailservice/mail_sender"
)

type MailService struct {
	mailSender           mailsender.MailSender
	mailContentFormatter mailcontentformatter.MailContentFormatter
	config               cfg.Configuration
}

func NewMailService(mailSender mailsender.MailSender, mailContentFormatter mailcontentformatter.MailContentFormatter, config cfg.Configuration) MailService {
	return MailService{mailSender: mailSender, mailContentFormatter: mailContentFormatter, config: config}
}

func (m MailService) SendWelcomeEmail(to string, name string) error {
	subject, body := m.mailContentFormatter.FormatWelcomeMailContent(name, m.config.Server.Host)
	return m.mailSender.Send(mailsender.MailRequest{
		From:    m.config.MailSender.Email,
		To:      to,
		Subject: subject,
		Body:    body,
	})
}

func (m MailService) SendResetPasswordMail(to string, code int, name string) error {
	subject, body := m.mailContentFormatter.FormatResetPasswordMailContent(code, m.config.MailSender.TimeoutMin, name, m.config.Server.Host)
	return m.mailSender.Send(mailsender.MailRequest{
		From:    m.config.MailSender.Email,
		To:      to,
		Subject: subject,
		Body:    body,
	})
}

func (m MailService) SendSignUpMail(to string, code int, name string) error {
	subject, body := m.mailContentFormatter.FormatSignUpMailContent(code, m.config.MailSender.TimeoutMin, name, m.config.Server.Host)
	return m.mailSender.Send(mailsender.MailRequest{
		From:    m.config.MailSender.Email,
		To:      to,
		Subject: subject,
		Body:    body,
	})
}

func (m MailService) SendInvoiceMail(to string, invoiceTotal float64, currency string, invoiceID int, attachments []mailsender.Attachment) error {
	subject, body := m.mailContentFormatter.FormatInvoiceMailContent(invoiceTotal, currency, invoiceID)
	return m.mailSender.Send(mailsender.MailRequest{
		From:        m.config.MailSender.Email,
		To:          to,
		Subject:     subject,
		Body:        body,
		Attachments: attachments,
	})
}

func (m MailService) SendSystemAnnouncementMail(to string, body string) error {
	mailBody := m.mailContentFormatter.FormatSystemAnnouncementMailBody(body)
	return m.mailSender.Send(mailsender.MailRequest{
		From:    m.config.MailSender.Email,
		To:      to,
		Subject: "System Announcement",
		Body:    mailBody,
	})
}

func (m MailService) SendNotifyAdminsEmail(to string, recordsNumber int) error {
	subject, body := m.mailContentFormatter.FormatNotifyAdminsMailContent(recordsNumber, m.config.Server.Host)
	return m.mailSender.Send(mailsender.MailRequest{
		From:    m.config.MailSender.Email,
		To:      to,
		Subject: subject,
		Body:    body,
	})
}

func (m MailService) SendEmailNotification(to string, subject string, body string) error {
	return m.mailSender.Send(mailsender.MailRequest{
		From:    m.config.MailSender.Email,
		To:      to,
		Subject: subject,
		Body:    body,
	})
}

func (m MailService) GetMailConfig() cfg.MailSender {
	return m.config.MailSender
}
