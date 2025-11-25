package mailcontentformatter

type MailContentFormatter interface {
	FormatResetPasswordMailContent(code int, timeout int, username string, systemHost string) (string, string)
	FormatSignUpMailContent(code int, timeout int, username string, systemHost string) (string, string)
	FormatWelcomeMailContent(username string, systemHost string) (string, string)
	FormatInvoiceMailContent(invoiceTotal float64, currency string, invoiceID int) (string, string)
	FormatSystemAnnouncementMailBody(body string) string
	FormatNotifyAdminsMailContent(recordsNumber int, systemHost string) (string, string)
}
