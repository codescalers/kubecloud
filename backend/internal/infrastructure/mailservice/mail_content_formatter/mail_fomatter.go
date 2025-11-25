package mailcontentformatter


type MailContentFormatter interface {
	FormatResetPasswordMailContent(code int, timeout int, username string) string
	FormatSignUpMailContent(code int, timeout int, username string) string
	FormatWelcomeMailContent(username string) string
	FormatInvoiceMailContent(invoiceTotal float64, currency string, invoiceID int) string
	FormatSystemAnnouncementMailBody(body string) string
	FormatNotifyAdminsMailContent(recordsNumber int) string
}
