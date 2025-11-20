package mailservice

import (
	_ "embed"
	"fmt"
	"html/template"
)

//go:embed templates/notifications/billing.html
var billingTemplate string

//go:embed templates/notifications/deployment.html
var deploymentTemplate string

//go:embed templates/notifications/node.html
var nodeTemplate string

//go:embed templates/notifications/user.html
var userTemplate string

var emailTemplates *template.Template

func init() {
	var err error
	emailTemplates, err = template.New("billing").Parse(billingTemplate)
	if err != nil {
		panic(fmt.Sprintf("failed to parse billing template: %v", err))
	}

	if _, err := emailTemplates.New("deployment").Parse(deploymentTemplate); err != nil {
		panic(fmt.Sprintf("failed to parse deployment template: %v", err))
	}

	if _, err := emailTemplates.New("node").Parse(nodeTemplate); err != nil {
		panic(fmt.Sprintf("failed to parse node template: %v", err))
	}

	if _, err := emailTemplates.New("user").Parse(userTemplate); err != nil {
		panic(fmt.Sprintf("failed to parse user template: %v", err))
	}
}

// GetEmailTemplates returns the embedded email templates
func GetEmailTemplates() *template.Template {
	return emailTemplates
}
