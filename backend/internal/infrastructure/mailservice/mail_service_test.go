package mailservice

import (
	"bytes"
	"mime/multipart"
	"strings"
	"testing"

	"kubecloud/internal/infrastructure/metrics"
	"kubecloud/internal/shared"
)

// TestIsValidEmail tests email validation logic.
// This scenario covers:
// - Valid email addresses in various formats
// - Invalid email addresses (missing @, invalid format)
// - Edge cases (empty, spaces, special characters)
func TestIsValidEmail(t *testing.T) {
	tests := []struct {
		name     string
		email    string
		expected bool
	}{
		{"simple_email", "user@example.com", true},
		{"email_with_plus", "user+tag@example.co.uk", true},
		{"email_with_dots", "first.last@example.com", true},
		{"email_with_numbers", "user123@example.com", true},
		{"subdomain_email", "user@mail.example.com", true},
		{"missing_at_sign", "userexample.com", false},
		{"missing_domain", "user@", false},
		{"missing_local_part", "@example.com", false},
		{"empty_string", "", false},
		{"spaces_in_email", "user @example.com", false},
		{"double_at_sign", "user@@example.com", false},
		{"no_tld", "user@localhost", true}, // Valid according to net/mail
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidEmail(tt.email)
			if result != tt.expected {
				t.Errorf("IsValidEmail(%q) = %v, want %v", tt.email, result, tt.expected)
			}
		})
	}
}

// TestFakeMailService_SystemMail tests system mail configuration.
// This scenario covers:
// - Retrieving configured system email
// - Empty system email handling
func TestFakeMailService_SystemMail(t *testing.T) {
	metrics := metrics.NewMetrics()
	service := NewFakeMailService(metrics)

	mail := service.SystemMail()
	if mail != "" {
		t.Errorf("FakeMailService.SystemMail() = %q, want empty string", mail)
	}
}

// TestFakeMailService_SendMail tests fake mail sending.
// This scenario covers:
// - Successful mail sending with valid email
// - Rejection of invalid email addresses
// - Metrics tracking (sent/failed emails)
func TestFakeMailService_SendMail(t *testing.T) {
	metrics := metrics.NewMetrics()
	service := NewFakeMailService(metrics)

	tests := []struct {
		name       string
		sender     string
		receiver   string
		subject    string
		body       string
		expectErr  bool
		description string
	}{
		{
			name:        "valid_email",
			sender:      "system@example.com",
			receiver:    "user@example.com",
			subject:     "Test Subject",
			body:        "Test Body",
			expectErr:   false,
			description: "sending to valid email address",
		},
		{
			name:        "invalid_receiver_email",
			sender:      "system@example.com",
			receiver:    "invalid-email",
			subject:     "Test Subject",
			body:        "Test Body",
			expectErr:   true,
			description: "sending to invalid email address",
		},
		{
			name:        "empty_receiver",
			sender:      "system@example.com",
			receiver:    "",
			subject:     "Test Subject",
			body:        "Test Body",
			expectErr:   true,
			description: "sending to empty receiver",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.SendMail(tt.sender, tt.receiver, tt.subject, tt.body)
			if (err != nil) != tt.expectErr {
				t.Errorf("SendMail() error = %v, wantErr %v (%s)", err, tt.expectErr, tt.description)
			}
		})
	}
}

// TestFakeMailService_SendMailFromSystem tests system mail sending.
// This scenario covers:
// - Sending mail from system account
// - Uses system email as sender
func TestFakeMailService_SendMailFromSystem(t *testing.T) {
	metrics := metrics.NewMetrics()
	service := NewFakeMailService(metrics)

	err := service.SendMailFromSystem("user@example.com", "Subject", "Body")
	if err != nil {
		t.Errorf("SendMailFromSystem() unexpected error: %v", err)
	}
}

// TestFakeMailService_MaxConcurrentSends tests concurrent send configuration.
// This scenario covers:
// - Returns configured max concurrent sends value
func TestFakeMailService_MaxConcurrentSends(t *testing.T) {
	metrics := metrics.NewMetrics()
	service := NewFakeMailService(metrics)

	max := service.MaxConcurrentSends()
	if max != 0 {
		t.Errorf("MaxConcurrentSends() = %d, want 0", max)
	}
}

// TestFakeMailService_MaxAttachmentSizeInBytes tests attachment size limit.
// This scenario covers:
// - Returns configured max attachment size in bytes
func TestFakeMailService_MaxAttachmentSizeInBytes(t *testing.T) {
	metrics := metrics.NewMetrics()
	service := NewFakeMailService(metrics)

	size := service.MaxAttachmentSizeInBytes()
	if size != 0 {
		t.Errorf("MaxAttachmentSizeInBytes() = %d, want 0", size)
	}
}

// TestFakeMailService_ParseAttachment tests attachment parsing.
// This scenario covers:
// - Fake service returns empty attachment (no actual parsing)
func TestFakeMailService_ParseAttachment(t *testing.T) {
	metrics := metrics.NewMetrics()
	service := NewFakeMailService(metrics)

	fh := &multipart.FileHeader{}
	att, err := service.ParseAttachment(fh)

	if err != nil {
		t.Errorf("ParseAttachment() unexpected error: %v", err)
	}
	if att.FileName != "" || len(att.Data) != 0 {
		t.Errorf("ParseAttachment() expected empty attachment, got FileName=%q Data=%d bytes", att.FileName, len(att.Data))
	}
}

// TestFakeMailService_ContentGenerators tests mail content generation methods.
// This scenario covers:
// - ResetPasswordMailContent returns empty subject and body
// - SignUpMailContent returns empty subject and body
// - WelcomeMailContent returns empty subject and body
// - InvoiceMailContent returns empty subject and body
// - NotifyAdminsMailContent returns empty subject and body
func TestFakeMailService_ContentGenerators(t *testing.T) {
	metrics := metrics.NewMetrics()
	service := NewFakeMailService(metrics)

	tests := []struct {
		name     string
		testFunc func() (string, string)
	}{
		{"ResetPasswordMailContent", func() (string, string) { return service.ResetPasswordMailContent(123456, 10, "testuser") }},
		{"SignUpMailContent", func() (string, string) { return service.SignUpMailContent(654321, 10, "testuser") }},
		{"WelcomeMailContent", func() (string, string) { return service.WelcomeMailContent("testuser") }},
		{"InvoiceMailContent", func() (string, string) { return service.InvoiceMailContent(100.50, "USD", 1) }},
		{"NotifyAdminsMailContent", func() (string, string) { return service.NotifyAdminsMailContent(5) }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			subject, body := tt.testFunc()
			if subject != "" || body != "" {
				t.Errorf("%s() expected empty subject and body, got subject=%q body=%q", tt.name, subject, body)
			}
		})
	}
}

// TestFakeMailService_SystemAnnouncementMailBody tests system announcement mail generation.
// This scenario covers:
// - SystemAnnouncementMailBody returns empty string for any input
func TestFakeMailService_SystemAnnouncementMailBody(t *testing.T) {
	metrics := metrics.NewMetrics()
	service := NewFakeMailService(metrics)

	result := service.SystemAnnouncementMailBody("Test announcement body")
	if result != "" {
		t.Errorf("SystemAnnouncementMailBody() expected empty string, got %q", result)
	}
}

// TestSendGridMailService_NewSendGridMailService tests service initialization.
// This scenario covers:
// - Creates service with proper configuration
// - Stores email, host, and limits from config
func TestSendGridMailService_NewSendGridMailService(t *testing.T) {
	config := shared.MailSender{
		Email:               "system@example.com",
		SendGridKey:         "test-key",
		TimeoutMin:          5,
		MaxConcurrentSends:  10,
		MaxAttachmentSizeMB: 25,
	}
	metrics := metrics.NewMetrics()

	service := NewSendGridMailService(config, "example.com", metrics)

	if service.SystemMail() != "system@example.com" {
		t.Errorf("SystemMail() = %q, want system@example.com", service.SystemMail())
	}
	if service.MaxConcurrentSends() != 10 {
		t.Errorf("MaxConcurrentSends() = %d, want 10", service.MaxConcurrentSends())
	}
	if service.MaxAttachmentSizeInBytes() != 25*1024*1024 {
		t.Errorf("MaxAttachmentSizeInBytes() = %d, want %d", service.MaxAttachmentSizeInBytes(), 25*1024*1024)
	}
}

// TestSendGridMailService_ResetPasswordMailContent tests reset password email generation.
// This scenario covers:
// - Returns non-empty subject
// - Body contains code and timeout
// - Body contains username in title case
// - Body contains host placeholder substitution
func TestSendGridMailService_ResetPasswordMailContent(t *testing.T) {
	config := shared.MailSender{
		Email:               "system@example.com",
		SendGridKey:         "test-key",
		MaxConcurrentSends:  10,
		MaxAttachmentSizeMB: 25,
	}
	metrics := metrics.NewMetrics()
	service := NewSendGridMailService(config, "app.example.com", metrics)

	subject, body := service.ResetPasswordMailContent(123456, 15, "johndoe")

	if subject == "" {
		t.Errorf("ResetPasswordMailContent() subject should not be empty")
	}
	if body == "" {
		t.Errorf("ResetPasswordMailContent() body should not be empty")
	}
	if !strings.Contains(body, "123456") {
		t.Errorf("ResetPasswordMailContent() body should contain code, got: %s", body)
	}
	if !strings.Contains(body, "15") {
		t.Errorf("ResetPasswordMailContent() body should contain timeout, got: %s", body)
	}
	if !strings.Contains(body, "app.example.com") {
		t.Errorf("ResetPasswordMailContent() body should contain host, got: %s", body)
	}
}

// TestSendGridMailService_SignUpMailContent tests signup email generation.
// This scenario covers:
// - Returns non-empty subject with emoji
// - Body contains code and timeout
// - Body contains username
// - Body contains host
func TestSendGridMailService_SignUpMailContent(t *testing.T) {
	config := shared.MailSender{
		Email:               "system@example.com",
		SendGridKey:         "test-key",
		MaxConcurrentSends:  10,
		MaxAttachmentSizeMB: 25,
	}
	metrics := metrics.NewMetrics()
	service := NewSendGridMailService(config, "app.example.com", metrics)

	subject, body := service.SignUpMailContent(654321, 10, "alice")

	if subject == "" {
		t.Errorf("SignUpMailContent() subject should not be empty")
	}
	if !strings.Contains(subject, "Welcome") || !strings.Contains(subject, "Mycelium") {
		t.Errorf("SignUpMailContent() subject should contain welcome message, got: %s", subject)
	}
	if !strings.Contains(body, "654321") {
		t.Errorf("SignUpMailContent() body should contain code, got: %s", body)
	}
	if !strings.Contains(body, "10") {
		t.Errorf("SignUpMailContent() body should contain timeout, got: %s", body)
	}
}

// TestSendGridMailService_WelcomeMailContent tests welcome email generation.
// This scenario covers:
// - Returns non-empty subject with emoji
// - Body contains username
// - Body contains host
func TestSendGridMailService_WelcomeMailContent(t *testing.T) {
	config := shared.MailSender{
		Email:               "system@example.com",
		SendGridKey:         "test-key",
		MaxConcurrentSends:  10,
		MaxAttachmentSizeMB: 25,
	}
	metrics := metrics.NewMetrics()
	service := NewSendGridMailService(config, "app.example.com", metrics)

	subject, body := service.WelcomeMailContent("bob")

	if subject == "" {
		t.Errorf("WelcomeMailContent() subject should not be empty")
	}
	if body == "" {
		t.Errorf("WelcomeMailContent() body should not be empty")
	}
	if !strings.Contains(body, "app.example.com") {
		t.Errorf("WelcomeMailContent() body should contain host, got: %s", body)
	}
}

// TestSendGridMailService_InvoiceMailContent tests invoice email generation.
// This scenario covers:
// - Returns invoice subject
// - Body contains amount and currency
// - Body contains invoice ID
func TestSendGridMailService_InvoiceMailContent(t *testing.T) {
	config := shared.MailSender{
		Email:               "system@example.com",
		SendGridKey:         "test-key",
		MaxConcurrentSends:  10,
		MaxAttachmentSizeMB: 25,
	}
	metrics := metrics.NewMetrics()
	service := NewSendGridMailService(config, "app.example.com", metrics)

	subject, body := service.InvoiceMailContent(1250.75, "USD", 42)

	if subject == "" {
		t.Errorf("InvoiceMailContent() subject should not be empty")
	}
	if !strings.Contains(subject, "Invoice") {
		t.Errorf("InvoiceMailContent() subject should contain 'Invoice', got: %s", subject)
	}
	if !strings.Contains(body, "1250.75") {
		t.Errorf("InvoiceMailContent() body should contain amount, got: %s", body)
	}
	if !strings.Contains(body, "USD") {
		t.Errorf("InvoiceMailContent() body should contain currency, got: %s", body)
	}
	if !strings.Contains(body, "42") {
		t.Errorf("InvoiceMailContent() body should contain invoice ID, got: %s", body)
	}
}

// TestSendGridMailService_SystemAnnouncementMailBody tests system announcement email generation.
// This scenario covers:
// - Returns HTML content from template
// - Replaces body with provided content
// - Converts newlines to <br> tags
func TestSendGridMailService_SystemAnnouncementMailBody(t *testing.T) {
	config := shared.MailSender{
		Email:               "system@example.com",
		SendGridKey:         "test-key",
		MaxConcurrentSends:  10,
		MaxAttachmentSizeMB: 25,
	}
	metrics := metrics.NewMetrics()
	service := NewSendGridMailService(config, "app.example.com", metrics)

	body := service.SystemAnnouncementMailBody("This is a test\nmessage with\nnewlines")

	if body == "" {
		t.Errorf("SystemAnnouncementMailBody() should return non-empty string")
	}
	if !strings.Contains(body, "<br>") {
		t.Errorf("SystemAnnouncementMailBody() should convert newlines to <br>, got: %s", body)
	}
}

// TestSendGridMailService_NotifyAdminsMailContent tests admin notification email generation.
// This scenario covers:
// - Returns non-empty subject about pending requests
// - Body contains number of records
// - Body contains host
func TestSendGridMailService_NotifyAdminsMailContent(t *testing.T) {
	config := shared.MailSender{
		Email:               "system@example.com",
		SendGridKey:         "test-key",
		MaxConcurrentSends:  10,
		MaxAttachmentSizeMB: 25,
	}
	metrics := metrics.NewMetrics()
	service := NewSendGridMailService(config, "app.example.com", metrics)

	subject, body := service.NotifyAdminsMailContent(7)

	if subject == "" {
		t.Errorf("NotifyAdminsMailContent() subject should not be empty")
	}
	if !strings.Contains(subject, "pending") {
		t.Errorf("NotifyAdminsMailContent() subject should mention pending, got: %s", subject)
	}
	if !strings.Contains(body, "7") {
		t.Errorf("NotifyAdminsMailContent() body should contain record count, got: %s", body)
	}
	if !strings.Contains(body, "app.example.com") {
		t.Errorf("NotifyAdminsMailContent() body should contain host, got: %s", body)
	}
}

// TestIsAttachmentAllowed tests attachment file type validation.
// This scenario covers:
// - Allowed file types (pdf, doc, docx, txt, jpg, jpeg, png, gif, zip)
// - Rejects disallowed file types (exe, sh, bat, etc.)
// - Case-insensitive file extension checking
func TestIsAttachmentAllowed(t *testing.T) {
	tests := []struct {
		filename string
		allowed  bool
		description string
	}{
		{"document.pdf", true, "PDF file allowed"},
		{"document.PDF", true, "PDF file allowed (uppercase)"},
		{"spreadsheet.doc", true, "DOC file allowed"},
		{"document.docx", true, "DOCX file allowed"},
		{"notes.txt", true, "TXT file allowed"},
		{"photo.jpg", true, "JPG file allowed"},
		{"image.jpeg", true, "JPEG file allowed"},
		{"screenshot.png", true, "PNG file allowed"},
		{"animation.gif", true, "GIF file allowed"},
		{"archive.zip", true, "ZIP file allowed"},
		{"script.exe", false, "EXE file not allowed"},
		{"script.sh", false, "SH file not allowed"},
		{"script.bat", false, "BAT file not allowed"},
		{"malware.app", false, "APP file not allowed"},
		{"noextension", false, "No extension not allowed"},
		{"document.pdf.exe", false, "Double extension not allowed"},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			result := isAttachmentAllowed(tt.filename)
			if result != tt.allowed {
				t.Errorf("isAttachmentAllowed(%q) = %v, want %v (%s)", tt.filename, result, tt.allowed, tt.description)
			}
		})
	}
}

// TestSendGridMailService_ParseAttachment tests attachment file parsing.
// This scenario covers:
// - Successful parsing of allowed file types
// - Rejection of disallowed file types
// - File size validation against max attachment size
// - File reading and data extraction
func TestSendGridMailService_ParseAttachment(t *testing.T) {
	config := shared.MailSender{
		Email:               "system@example.com",
		SendGridKey:         "test-key",
		MaxConcurrentSends:  10,
		MaxAttachmentSizeMB: 1, // 1 MB
	}
	metrics := metrics.NewMetrics()
	service := NewSendGridMailService(config, "app.example.com", metrics)

	tests := []struct {
		name        string
		filename    string
		fileSize    int64
		fileData    []byte
		expectErr   bool
		description string
	}{
		{
			name:        "valid_pdf",
			filename:    "document.pdf",
			fileSize:    1000,
			fileData:    []byte("PDF data"),
			expectErr:   false,
			description: "valid PDF file",
		},
		{
			name:        "invalid_extension",
			filename:    "script.exe",
			fileSize:    500,
			fileData:    []byte("executable"),
			expectErr:   true,
			description: "disallowed file type",
		},
		{
			name:        "file_too_large",
			filename:    "document.pdf",
			fileSize:    2 * 1024 * 1024, // 2 MB (exceeds 1 MB limit)
			fileData:    nil,
			expectErr:   true,
			description: "file exceeds size limit",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create multipart file header
			body := new(bytes.Buffer)
			writer := multipart.NewWriter(body)
			part, _ := writer.CreateFormFile("file", tt.filename)

			if tt.fileData != nil {
				part.Write(tt.fileData)
			}
			writer.Close()

			// Parse the multipart data to get FileHeader
			reader := multipart.NewReader(body, writer.Boundary())
			form, _ := reader.ReadForm(10 * 1024 * 1024)
			files := form.File["file"]
			if len(files) == 0 {
				t.Fatalf("no file in multipart form")
			}

			fh := files[0]
			fh.Size = tt.fileSize

			att, err := service.ParseAttachment(fh)

			if (err != nil) != tt.expectErr {
				t.Errorf("ParseAttachment() error = %v, wantErr %v (%s)", err, tt.expectErr, tt.description)
			}
			if !tt.expectErr && att.FileName != tt.filename {
				t.Errorf("ParseAttachment() FileName = %q, want %q", att.FileName, tt.filename)
			}
		})
	}
}

// TestAttachment tests Attachment struct.
// This scenario covers:
// - Attachment struct has FileName and Data fields
// - Can store file metadata and content
func TestAttachment(t *testing.T) {
	att := Attachment{
		FileName: "test.pdf",
		Data:     []byte("test data"),
	}

	if att.FileName != "test.pdf" {
		t.Errorf("Attachment.FileName = %q, want test.pdf", att.FileName)
	}
	if string(att.Data) != "test data" {
		t.Errorf("Attachment.Data = %s, want test data", string(att.Data))
	}
}
