package mailsender

import (
	"context"
	"fmt"
	"strings"

	"kubecloud/internal/infrastructure/logger"
)

// FakeMailSender overrides MailSender methods for development purposes
type FakeMailSender struct {
}

// NewFakeMailSender creates a new fake mail sender for development
func NewFakeMailSender() FakeMailSender {
	return FakeMailSender{}
}

// Send overrides to track metrics without actually sending
func (s FakeMailSender) Send(ctx context.Context, req MailRequest) error {
	if !isValidEmail(req.To) {
		return fmt.Errorf("email %v is not valid", req.To)
	}

	files := make([]string, len(req.Attachments))
	for i, att := range req.Attachments {
		files[i] = att.FileName
	}

	logger.GetLogger().Info().Msg(formatMailSummaryBox(req, files))

	return nil
}

func formatMailSummaryBox(req MailRequest, files []string) string {
	attachmentList := "None"
	if len(files) > 0 {
		attachmentList = strings.Join(files, ", ")
	}

	return fmt.Sprintf("\n"+
		"╔══════════════════════════════════════════════════════════╗\n"+
		"║                   FAKE MAIL DISPATCH                     ║\n"+
		"╠══════════════════════════════════════════════════════════╣\n"+
		"║  From: %-47s║\n"+
		"║  To:   %-47s║\n"+
		"║  Subj: %-47s║\n"+
		"║  Body: %-47s║\n"+
		"║  Files:%-48s║\n"+
		"╚══════════════════════════════════════════════════════════╝",
		req.From, req.To, req.Subject, req.Body, attachmentList)
}
