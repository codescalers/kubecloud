package activities

import (
	"context"
	"kubecloud/internal"
	"kubecloud/internal/metrics"
	"kubecloud/models"
	"time"

	"github.com/rs/zerolog/log"
	substrate "github.com/threefoldtech/tfchain/clients/tfchain-client-go"
	"github.com/vedhavyas/go-subkey"
	"github.com/xmonader/ewf"
)

var userWorkflowTemplate = ewf.WorkflowTemplate{
	BeforeWorkflowHooks: []ewf.BeforeWorkflowHook{
		func(ctx context.Context, w *ewf.Workflow) {
			log.Info().Str("workflow_name", w.Name).Msg("Starting workflow")
		},
	},
	BeforeStepHooks: []ewf.BeforeStepHook{
		func(ctx context.Context, w *ewf.Workflow, step *ewf.Step) {
			log.Info().Str("workflow_name", w.Name).Str("step_name", step.Name).Msg("Starting step")
		},
	},
	AfterStepHooks: []ewf.AfterStepHook{
		func(ctx context.Context, w *ewf.Workflow, step *ewf.Step, err error) {
			if err != nil {
				log.Error().Err(err).Str("workflow_name", w.Name).Str("step_name", step.Name).Msg("Step failed")
			} else {
				log.Info().Str("workflow_name", w.Name).Str("step_name", step.Name).Msg("Step completed successfully")
			}
		},
	},
	AfterWorkflowHooks: []ewf.AfterWorkflowHook{
		func(ctx context.Context, w *ewf.Workflow, err error) {
			if err != nil {
				log.Error().Err(err).Str("workflow_name", w.Name).Msg("Workflow completed with error")
			} else {
				log.Info().Str("workflow_name", w.Name).Msg("Workflow completed successfully")
			}
		},
	},
}

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
	engine.Register(StepCreatePendingRecord, CreatePendingRecord(substrate, db, config.SystemAccount.Mnemonic, sse))
	engine.Register(StepUpdateCreditCardBalance, UpdateCreditCardBalanceStep(db))
	engine.Register(StepCreateIdentity, CreateIdentityStep())
	engine.Register(StepReserveNode, ReserveNodeStep(db, substrate))
	engine.Register(StepUnreserveNode, UnreserveNodeStep(db, substrate))
	engine.Register(StepUpdateCreditedBalance, UpdateCreditedBalanceStep(db))

	registerWorkflowTemplate := userWorkflowTemplate
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

	userVerificationTemplate := userWorkflowTemplate
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

	chargeBalanceTemplate := userWorkflowTemplate
	chargeBalanceTemplate.Steps = []ewf.Step{
		{Name: StepCreatePaymentIntent, RetryPolicy: &ewf.RetryPolicy{MaxAttempts: 2, BackOff: ewf.ConstantBackoff(2 * time.Second)}},
		{Name: StepUpdateCreditCardBalance, RetryPolicy: &ewf.RetryPolicy{MaxAttempts: 2, BackOff: ewf.ConstantBackoff(2 * time.Second)}},
		{Name: StepCreatePendingRecord, RetryPolicy: &ewf.RetryPolicy{MaxAttempts: 2, BackOff: ewf.ConstantBackoff(2 * time.Second)}},
	}
	engine.RegisterTemplate(WorkflowChargeBalance, &chargeBalanceTemplate)

	adminCreditBalanceTemplate := userWorkflowTemplate
	adminCreditBalanceTemplate.Steps = []ewf.Step{
		{Name: StepUpdateCreditedBalance, RetryPolicy: &ewf.RetryPolicy{MaxAttempts: 2, BackOff: ewf.ConstantBackoff(2 * time.Second)}},
		{Name: StepCreatePendingRecord, RetryPolicy: &ewf.RetryPolicy{MaxAttempts: 2, BackOff: ewf.ConstantBackoff(2 * time.Second)}},
	}
	engine.RegisterTemplate(WorkflowAdminCreditBalance, &adminCreditBalanceTemplate)

	redeemVoucherTemplate := userWorkflowTemplate
	redeemVoucherTemplate.Steps = []ewf.Step{
		{Name: StepUpdateCreditedBalance, RetryPolicy: &ewf.RetryPolicy{MaxAttempts: 2, BackOff: ewf.ConstantBackoff(2 * time.Second)}},
		{Name: StepCreatePendingRecord, RetryPolicy: &ewf.RetryPolicy{MaxAttempts: 2, BackOff: ewf.ConstantBackoff(2 * time.Second)}},
	}
	engine.RegisterTemplate(WorkflowRedeemVoucher, &redeemVoucherTemplate)

	reserveNodeTemplate := userWorkflowTemplate
	reserveNodeTemplate.Steps = []ewf.Step{
		{Name: StepCreateIdentity, RetryPolicy: &ewf.RetryPolicy{MaxAttempts: 2, BackOff: ewf.ConstantBackoff(2 * time.Second)}},
		{Name: StepReserveNode, RetryPolicy: &ewf.RetryPolicy{MaxAttempts: 2, BackOff: ewf.ConstantBackoff(2 * time.Second)}},
	}
	engine.RegisterTemplate(WorkflowReserveNode, &reserveNodeTemplate)

	unreserveNodeTemplate := userWorkflowTemplate
	unreserveNodeTemplate.Steps = []ewf.Step{
		{Name: StepUnreserveNode, RetryPolicy: &ewf.RetryPolicy{MaxAttempts: 2, BackOff: ewf.ConstantBackoff(2 * time.Second)}},
	}
	engine.RegisterTemplate(WorkflowUnreserveNode, &unreserveNodeTemplate)

	registerDeploymentActivities(engine, metrics, db, sse)
}
