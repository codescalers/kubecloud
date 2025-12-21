package workflows

import (
	"kubecloud/internal/billing"
	cfg "kubecloud/internal/config"
	"kubecloud/internal/core/models"
	"kubecloud/internal/core/persistence"
	"kubecloud/internal/infrastructure/gridclient"
	"kubecloud/internal/infrastructure/kyc"
	"kubecloud/internal/infrastructure/mailservice"
	"kubecloud/internal/infrastructure/metrics"
	"kubecloud/internal/infrastructure/notification"

	"time"

	"github.com/vedhavyas/go-subkey"
	"github.com/xmonader/ewf"
)

func RegisterEWFWorkflows(
	engine *ewf.Engine,
	config cfg.Configuration,
	db models.DB,
	mailService mailservice.MailService,
	gridClient gridclient.GridClient,
	kycClient *kyc.KYCClient,
	sponsorAddress string,
	sponsorKeyPair subkey.KeyPair,
	metrics *metrics.Metrics,
	notificationDispatcher *notification.NotificationDispatcher,
	stripeClient billing.StripeClient,
) {
	userRepo := persistence.NewGormUserRepository(db)
	clusterRepo := persistence.NewGormClusterRepository(db)
	contractsRepo := persistence.NewGormUserContractDataRepository(db)

	engine.Register(StepSendVerificationEmail, SendVerificationEmailStep(mailService, config))
	engine.Register(StepCreateUser, CreateUserStep(config, userRepo))
	engine.Register(StepUpdateCode, UpdateCodeStep(userRepo))
	engine.Register(StepSetupTFChain, SetupTFChainStep(gridClient, userRepo, config.TermsANDConditions))
	engine.Register(StepCreateStripeCustomer, CreateStripeCustomerStep(userRepo, stripeClient))
	engine.Register(StepCreateKYCSponsorship, CreateKYCSponsorship(kycClient, sponsorAddress, sponsorKeyPair, userRepo))
	engine.Register(StepSendWelcomeEmail, SendWelcomeEmailStep(mailService, metrics))
	engine.Register(StepCreatePaymentIntent, CreatePaymentIntentStep(config.Currency, metrics, stripeClient))
	engine.Register(StepUpdateCreditCardBalance, UpdateCreditCardBalanceStep(userRepo))
	engine.Register(StepReserveNode, ReserveNodeStep(contractsRepo, gridClient))
	engine.Register(StepUnreserveNode, UnreserveNodeStep(contractsRepo, gridClient))
	engine.Register(StepSendEmailNotification, SendEmailNotificationStep(userRepo, mailService))
	engine.Register(StepVerifyNodeState, VerifyNodeStateStep(gridClient))
	engine.Register(StepVerifyClusterInDB, VerifyClusterInDBStep(clusterRepo))
	engine.Register(StepDrainUserBalance, DrainUserBalanceStep(userRepo, gridClient))
	engine.Register(StepDrainAllUsersBalance, DrainAllUsersBalanceStep(userRepo, engine, config.MailSender.MaxConcurrentSends))
	engine.Register(StepCheckClusterNodesHealth, CheckClusterNodesHealthStep(clusterRepo, contractsRepo))
	engine.Register(StepCheckClusterHealth, CheckClusterHealthStep(config.SSH.PrivateKeyPath))

	registerWorkflowTemplate := newKubecloudWorkflowTemplate(notificationDispatcher)
	registerWorkflowTemplate.BeforeWorkflowHooks = []ewf.BeforeWorkflowHook{
		// hookNotificationWorkflowStarted,
	}
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

	userVerificationTemplate := newKubecloudWorkflowTemplate(notificationDispatcher)
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

	chargeBalanceTemplate := newKubecloudWorkflowTemplate(notificationDispatcher)
	chargeBalanceTemplate.Steps = []ewf.Step{
		{Name: StepCreatePaymentIntent, RetryPolicy: &ewf.RetryPolicy{MaxAttempts: 2, BackOff: ewf.ConstantBackoff(2 * time.Second)}},
		{Name: StepUpdateCreditCardBalance, RetryPolicy: &ewf.RetryPolicy{MaxAttempts: 2, BackOff: ewf.ConstantBackoff(2 * time.Second)}},
	}
	engine.RegisterTemplate(WorkflowChargeBalance, &chargeBalanceTemplate)

	reserveNodeTemplate := newKubecloudWorkflowTemplate(notificationDispatcher)
	reserveNodeTemplate.Steps = []ewf.Step{
		{Name: StepReserveNode, RetryPolicy: &ewf.RetryPolicy{MaxAttempts: 2, BackOff: ewf.ConstantBackoff(2 * time.Second)}},
		{Name: StepVerifyNodeState, RetryPolicy: &ewf.RetryPolicy{MaxAttempts: 5, BackOff: ewf.ExponentialBackoff(10*time.Second, 2*time.Minute, 2.0)}},
	}
	engine.RegisterTemplate(WorkflowReserveNode, &reserveNodeTemplate)

	unreserveNodeTemplate := newKubecloudWorkflowTemplate(notificationDispatcher)
	unreserveNodeTemplate.Steps = []ewf.Step{
		{Name: StepUnreserveNode, RetryPolicy: &ewf.RetryPolicy{MaxAttempts: 2, BackOff: ewf.ConstantBackoff(2 * time.Second)}},
		{Name: StepVerifyNodeState, RetryPolicy: &ewf.RetryPolicy{MaxAttempts: 5, BackOff: ewf.ExponentialBackoff(10*time.Second, 2*time.Minute, 2.0)}},
	}
	engine.RegisterTemplate(WorkflowUnreserveNode, &unreserveNodeTemplate)

	trackClusterHealthWFTemplate := newKubecloudWorkflowTemplate(notificationDispatcher)
	trackClusterHealthWFTemplate.Steps = []ewf.Step{
		{Name: StepVerifyClusterInDB, RetryPolicy: standardRetryPolicy},
		{Name: StepCheckClusterNodesHealth, RetryPolicy: standardRetryPolicy},
		{Name: StepCheckClusterHealth, RetryPolicy: standardRetryPolicy},
	}
	trackClusterHealthWFTemplate.AfterWorkflowHooks = []ewf.AfterWorkflowHook{hookClusterHealthCheck(notificationDispatcher)}
	// trackClusterHealthWFTemplate.BeforeWorkflowHooks = []ewf.BeforeWorkflowHook{hookNotificationWorkflowStarted}
	engine.RegisterTemplate(WorkflowTrackClusterHealth, &trackClusterHealthWFTemplate)

	registerDeploymentActivities(engine, metrics, clusterRepo, contractsRepo, notificationDispatcher, config)

	// Email-only workflow for guaranteed email delivery with retries
	emailNotificationTemplate := ewf.WorkflowTemplate{
		Steps: []ewf.Step{
			{Name: StepSendEmailNotification, RetryPolicy: &ewf.RetryPolicy{MaxAttempts: 2, BackOff: ewf.ConstantBackoff(2 * time.Second)}},
		},
	}
	engine.RegisterTemplate(WorkflowSendEmailNotification, &emailNotificationTemplate)

	// Drain user balance workflow
	drainUserTemplate := newKubecloudWorkflowTemplate(notificationDispatcher)
	drainUserTemplate.Steps = []ewf.Step{
		{Name: StepDrainUserBalance, RetryPolicy: &ewf.RetryPolicy{MaxAttempts: 2, BackOff: ewf.ConstantBackoff(2 * time.Second)}},
	}
	engine.RegisterTemplate(WorkflowDrainUser, &drainUserTemplate)

	// Drain all users balances workflow
	drainAllUsersTemplate := newKubecloudWorkflowTemplate(notificationDispatcher)
	drainAllUsersTemplate.Steps = []ewf.Step{
		{Name: StepDrainAllUsersBalance, RetryPolicy: &ewf.RetryPolicy{MaxAttempts: 2, BackOff: ewf.ConstantBackoff(2 * time.Second)}},
	}
	engine.RegisterTemplate(WorkflowDrainAllUsers, &drainAllUsersTemplate)
}
