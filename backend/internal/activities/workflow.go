package activities

import (
	"kubecloud/internal"
	"kubecloud/internal/metrics"
	"kubecloud/models"
	"time"

	substrate "github.com/threefoldtech/tfchain/clients/tfchain-client-go"
	"github.com/vedhavyas/go-subkey"
	"github.com/xmonader/ewf"
)

func RegisterEWFWorkflows(
	engine *ewf.Engine,
	config internal.Configuration,
	db models.DB,
	mail internal.MailService,
	substrate *substrate.Substrate,
	sse *internal.SSEManager,
	kycClient *internal.KYCClient,
	sponsorAddress string,
	sponsorKeyPair subkey.KeyPair,
	metrics *metrics.Metrics,
) {
	engine.Register(StepSendVerificationEmail, SendVerificationEmailStep(mail, config))
	engine.Register(StepCreateUser, CreateUserStep(config, db))
	engine.Register(StepUpdateCode, UpdateCodeStep(db))
	engine.Register(StepSetupTFChain, SetupTFChainStep(substrate, config, sse, db))
	engine.Register(StepCreateStripeCustomer, CreateStripeCustomerStep(db))
	engine.Register(StepCreateKYCSponsorship, CreateKYCSponsorship(kycClient, sse, sponsorAddress, sponsorKeyPair, db))
	engine.Register(StepSendWelcomeEmail, SendWelcomeEmailStep(mail, config, metrics))
	engine.Register(StepCreatePaymentIntent, CreatePaymentIntentStep(config.Currency, metrics))
	engine.Register(StepUpdateCreditCardBalance, UpdateCreditCardBalanceStep(db))
	engine.Register(StepCreateIdentity, CreateIdentityStep())
	engine.Register(StepReserveNode, ReserveNodeStep(db, substrate))
	engine.Register(StepUnreserveNode, UnreserveNodeStep(db, substrate))

	registerWorkflowTemplate := newKubecloudWorkflowTemplate()
	registerWorkflowTemplate.Steps = []ewf.Step{
		{Name: StepCreateUser, RetryPolicy: &ewf.RetryPolicy{
			MaxAttempts: 2,
			BackOff:     ewf.ConstantBackoff(2 * time.Second),
		}},
		{Name: StepSendVerificationEmail, RetryPolicy: &ewf.RetryPolicy{
			MaxAttempts: 3,
			BackOff:     ewf.ConstantBackoff(2 * time.Second),
		}},
		{Name: StepUpdateCode, RetryPolicy: &ewf.RetryPolicy{
			MaxAttempts: 2,
			BackOff:     ewf.ConstantBackoff(2 * time.Second),
		}},
	}
	engine.RegisterTemplate(WorkflowUserRegistration, &registerWorkflowTemplate)

	userVerificationTemplate := newKubecloudWorkflowTemplate()
	userVerificationTemplate.Steps = []ewf.Step{
		{Name: StepSetupTFChain, RetryPolicy: &ewf.RetryPolicy{
			MaxAttempts: 5,
			BackOff:     ewf.ConstantBackoff(2 * time.Second),
		}},
		{Name: StepCreateStripeCustomer, RetryPolicy: &ewf.RetryPolicy{
			MaxAttempts: 3,
			BackOff:     ewf.ConstantBackoff(2 * time.Second),
		}},
		{Name: StepCreateKYCSponsorship, RetryPolicy: &ewf.RetryPolicy{
			MaxAttempts: 3,
			BackOff:     ewf.ConstantBackoff(2 * time.Second),
		}},
		{Name: StepSendWelcomeEmail, RetryPolicy: &ewf.RetryPolicy{
			MaxAttempts: 3,
			BackOff:     ewf.ConstantBackoff(2 * time.Second),
		}},
	}
	engine.RegisterTemplate(WorkflowUserVerification, &userVerificationTemplate)

	chargeBalanceTemplate := newKubecloudWorkflowTemplate()
	chargeBalanceTemplate.Steps = []ewf.Step{
		{Name: StepCreatePaymentIntent, RetryPolicy: &ewf.RetryPolicy{MaxAttempts: 2, BackOff: ewf.ConstantBackoff(2 * time.Second)}},
		{Name: StepUpdateCreditCardBalance, RetryPolicy: &ewf.RetryPolicy{MaxAttempts: 2, BackOff: ewf.ConstantBackoff(2 * time.Second)}},
	}
	engine.RegisterTemplate(WorkflowChargeBalance, &chargeBalanceTemplate)

	reserveNodeTemplate := newKubecloudWorkflowTemplate()
	reserveNodeTemplate.Steps = []ewf.Step{
		{Name: StepCreateIdentity, RetryPolicy: &ewf.RetryPolicy{MaxAttempts: 2, BackOff: ewf.ConstantBackoff(2 * time.Second)}},
		{Name: StepReserveNode, RetryPolicy: &ewf.RetryPolicy{MaxAttempts: 2, BackOff: ewf.ConstantBackoff(2 * time.Second)}},
	}
	engine.RegisterTemplate(WorkflowReserveNode, &reserveNodeTemplate)

	unreserveNodeTemplate := newKubecloudWorkflowTemplate()
	unreserveNodeTemplate.Steps = []ewf.Step{
		{Name: StepUnreserveNode, RetryPolicy: &ewf.RetryPolicy{MaxAttempts: 2, BackOff: ewf.ConstantBackoff(2 * time.Second)}},
	}
	engine.RegisterTemplate(WorkflowUnreserveNode, &unreserveNodeTemplate)

	registerDeploymentActivities(engine, metrics, db, sse)
}
