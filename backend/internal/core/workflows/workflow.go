package workflows

import (
	"kubecloud/internal/billing"
	cfg "kubecloud/internal/config"
	"kubecloud/internal/core/models"
	"kubecloud/internal/core/persistence"
	"kubecloud/internal/infrastructure/kyc"
	mailservice "kubecloud/internal/infrastructure/mailservice"
	"kubecloud/internal/infrastructure/metrics"
	"kubecloud/internal/infrastructure/notification"
	"kubecloud/internal/infrastructure/substrate"

	"time"

	proxy "github.com/threefoldtech/tfgrid-sdk-go/grid-proxy/pkg/client"
	"github.com/vedhavyas/go-subkey"
	"github.com/xmonader/ewf"
)

var workflowsDescriptions = map[string]string{
	WorkflowAddNode:                  "Adding Node",
	WorkflowRemoveNode:               "Removing Node",
	WorkflowDeleteCluster:            "Deleting Cluster",
	WorkflowDeleteAllClusters:        "Deleting All Clusters",
	WorkflowRollbackFailedDeployment: "Rollback",
	WorkflowUserRegistration:         "User Registration",
	WorkflowUserVerification:         "User Verification",
	WorkflowChargeBalance:            "Charge Balance",
	WorkflowAdminCreditBalance:       "Admin Credit Balance",
	WorkflowRedeemVoucher:            "Redeem Voucher",
	WorkflowReserveNode:              "Reserve Node",
	WorkflowUnreserveNode:            "Unreserve Node",
	WorkflowTrackClusterHealth:       "Cluster Health Check",
}

func RegisterEWFWorkflows(
	engine *ewf.Engine,
	config cfg.Configuration,
	db models.DB,
	mail mailservice.MailService,
	substrate substrate.Substrate,
	kycClient *kyc.KYCClient,
	sponsorAddress string,
	sponsorKeyPair subkey.KeyPair,
	metrics *metrics.Metrics,
	notificationDispatcher *notification.NotificationDispatcher,
	proxyClient proxy.Client,
	stripeClient billing.StripeClient,
) {
	userRepo := persistence.NewGormUserRepository(db)
	clusterRepo := persistence.NewGormClusterRepository(db)
	userNodesRepo := persistence.NewGormUserNodesRepository(db)
	pendingRecordRepo := persistence.NewGormPendingRecordRepository(db)

	engine.Register(StepSendVerificationEmail, SendVerificationEmailStep(mail, config))
	engine.Register(StepCreateUser, CreateUserStep(config, userRepo))
	engine.Register(StepUpdateCode, UpdateCodeStep(userRepo))
	engine.Register(StepSetupTFChain, SetupTFChainStep(substrate, userRepo))
	engine.Register(StepCreateStripeCustomer, CreateStripeCustomerStep(userRepo, stripeClient))
	engine.Register(StepCreateKYCSponsorship, CreateKYCSponsorship(kycClient, sponsorAddress, sponsorKeyPair, userRepo))
	engine.Register(StepSendWelcomeEmail, SendWelcomeEmailStep(mail, config, metrics))
	engine.Register(StepCreatePaymentIntent, CreatePaymentIntentStep(config.Currency, metrics, stripeClient))
	engine.Register(StepCreatePendingRecord, CreatePendingRecord(substrate, userRepo, pendingRecordRepo, config.SystemAccount.Mnemonic))
	engine.Register(StepUpdateCreditCardBalance, UpdateCreditCardBalanceStep(userRepo))
	engine.Register(StepReserveNode, ReserveNodeStep(userNodesRepo, substrate))
	engine.Register(StepUnreserveNode, UnreserveNodeStep(userNodesRepo, substrate))
	engine.Register(StepUpdateCreditedBalance, UpdateCreditedBalanceStep(userRepo))
	engine.Register(StepSendEmailNotification, SendEmailNotificationStep(userRepo, mail))
	engine.Register(StepVerifyNodeState, VerifyNodeStateStep(proxyClient))
	engine.Register(StepVerifyClusterInDB, VerifyClusterInDBStep(clusterRepo))

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
		{Name: StepCreatePendingRecord, RetryPolicy: &ewf.RetryPolicy{MaxAttempts: 2, BackOff: ewf.ConstantBackoff(2 * time.Second)}},
	}
	engine.RegisterTemplate(WorkflowChargeBalance, &chargeBalanceTemplate)

	adminCreditBalanceTemplate := newKubecloudWorkflowTemplate(notificationDispatcher)
	adminCreditBalanceTemplate.Steps = []ewf.Step{
		{Name: StepUpdateCreditedBalance, RetryPolicy: &ewf.RetryPolicy{MaxAttempts: 2, BackOff: ewf.ConstantBackoff(2 * time.Second)}},
		{Name: StepCreatePendingRecord, RetryPolicy: &ewf.RetryPolicy{MaxAttempts: 2, BackOff: ewf.ConstantBackoff(2 * time.Second)}},
	}
	engine.RegisterTemplate(WorkflowAdminCreditBalance, &adminCreditBalanceTemplate)

	redeemVoucherTemplate := newKubecloudWorkflowTemplate(notificationDispatcher)
	redeemVoucherTemplate.Steps = []ewf.Step{
		{Name: StepUpdateCreditedBalance, RetryPolicy: &ewf.RetryPolicy{MaxAttempts: 2, BackOff: ewf.ConstantBackoff(2 * time.Second)}},
		{Name: StepCreatePendingRecord, RetryPolicy: &ewf.RetryPolicy{MaxAttempts: 2, BackOff: ewf.ConstantBackoff(2 * time.Second)}},
	}
	engine.RegisterTemplate(WorkflowRedeemVoucher, &redeemVoucherTemplate)

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
		{Name: StepFetchKubeconfig, RetryPolicy: standardRetryPolicy},
		{Name: StepVerifyClusterReady, RetryPolicy: standardRetryPolicy},
	}
	trackClusterHealthWFTemplate.AfterWorkflowHooks = []ewf.AfterWorkflowHook{hookClusterHealthCheck(notificationDispatcher)}
	// trackClusterHealthWFTemplate.BeforeWorkflowHooks = []ewf.BeforeWorkflowHook{hookNotificationWorkflowStarted}
	engine.RegisterTemplate(WorkflowTrackClusterHealth, &trackClusterHealthWFTemplate)

	registerDeploymentActivities(engine, metrics, clusterRepo, notificationDispatcher, config)

	// Email-only workflow for guaranteed email delivery with retries
	emailNotificationTemplate := ewf.WorkflowTemplate{
		Steps: []ewf.Step{
			{Name: StepSendEmailNotification, RetryPolicy: &ewf.RetryPolicy{MaxAttempts: 2, BackOff: ewf.ConstantBackoff(2 * time.Second)}},
		},
	}
	engine.RegisterTemplate(WorkflowSendEmailNotification, &emailNotificationTemplate)
}
