package workflows

import (
	"kubecloud/internal/billing"
	"kubecloud/internal/core/models"
	"kubecloud/internal/core/persistence"
	"kubecloud/internal/infrastructure/kyc"
	mailservice "kubecloud/internal/infrastructure/mailservice"
	"kubecloud/internal/infrastructure/metrics"
	"kubecloud/internal/infrastructure/notification"
	"kubecloud/internal/infrastructure/substrate"
	"kubecloud/internal/shared"
	"time"

	proxy "github.com/threefoldtech/tfgrid-sdk-go/grid-proxy/pkg/client"
	"github.com/vedhavyas/go-subkey"
	"github.com/xmonader/ewf"
)

var workflowsDescriptions = map[string]string{
	shared.WorkflowAddNode:                  "Adding Node",
	shared.WorkflowRemoveNode:               "Removing Node",
	shared.WorkflowDeleteCluster:            "Deleting Cluster",
	shared.WorkflowDeleteAllClusters:        "Deleting All Clusters",
	shared.WorkflowRollbackFailedDeployment: "Rollback",
	shared.WorkflowUserRegistration:         "User Registration",
	shared.WorkflowUserVerification:         "User Verification",
	shared.WorkflowChargeBalance:            "Charge Balance",
	shared.WorkflowAdminCreditBalance:       "Admin Credit Balance",
	shared.WorkflowRedeemVoucher:            "Redeem Voucher",
	shared.WorkflowReserveNode:              "Reserve Node",
	shared.WorkflowUnreserveNode:            "Unreserve Node",
	shared.WorkflowTrackClusterHealth:       "Cluster Health Check",
}

func RegisterEWFWorkflows(
	engine *ewf.Engine,
	config shared.Configuration,
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

	engine.Register(shared.StepSendVerificationEmail, SendVerificationEmailStep(mail, config))
	engine.Register(shared.StepCreateUser, CreateUserStep(config, userRepo))
	engine.Register(shared.StepUpdateCode, UpdateCodeStep(userRepo))
	engine.Register(shared.StepSetupTFChain, SetupTFChainStep(substrate, userRepo))
	engine.Register(shared.StepCreateStripeCustomer, CreateStripeCustomerStep(userRepo, stripeClient))
	engine.Register(shared.StepCreateKYCSponsorship, CreateKYCSponsorship(kycClient, sponsorAddress, sponsorKeyPair, userRepo))
	engine.Register(shared.StepSendWelcomeEmail, SendWelcomeEmailStep(mail, config, metrics))
	engine.Register(shared.StepCreatePaymentIntent, CreatePaymentIntentStep(config.Currency, metrics, stripeClient))
	engine.Register(shared.StepCreatePendingRecord, CreatePendingRecord(substrate, userRepo, pendingRecordRepo, config.SystemAccount.Mnemonic))
	engine.Register(shared.StepUpdateCreditCardBalance, UpdateCreditCardBalanceStep(userRepo))
	engine.Register(shared.StepReserveNode, ReserveNodeStep(userNodesRepo, substrate))
	engine.Register(shared.StepUnreserveNode, UnreserveNodeStep(userNodesRepo, substrate))
	engine.Register(shared.StepUpdateCreditedBalance, UpdateCreditedBalanceStep(userRepo))
	engine.Register(shared.StepSendEmailNotification, SendNotification(userRepo, notificationDispatcher.GetNotifiers()[notification.ChannelEmail]))
	engine.Register(shared.StepSendUINotification, SendNotification(userRepo, notificationDispatcher.GetNotifiers()[notification.ChannelUI]))
	engine.Register(shared.StepVerifyNodeState, VerifyNodeStateStep(proxyClient))
	engine.Register(shared.StepVerifyClusterInDB, VerifyClusterInDBStep(clusterRepo))

	registerWorkflowTemplate := newKubecloudWorkflowTemplate(notificationDispatcher)
	registerWorkflowTemplate.BeforeWorkflowHooks = []ewf.BeforeWorkflowHook{
		// hookNotificationWorkflowStarted,
	}
	registerWorkflowTemplate.Steps = []ewf.Step{
		{Name: shared.StepCreateUser, RetryPolicy: &ewf.RetryPolicy{
			MaxAttempts: 2,
			BackOff:     ewf.ConstantBackoff(2 * time.Second),
		}},
		{Name: shared.StepSendVerificationEmail, RetryPolicy: &ewf.RetryPolicy{
			MaxAttempts: 3,
			BackOff:     ewf.ConstantBackoff(2 * time.Second),
		}},
		{Name: shared.StepUpdateCode, RetryPolicy: &ewf.RetryPolicy{
			MaxAttempts: 2,
			BackOff:     ewf.ConstantBackoff(2 * time.Second),
		}},
	}
	engine.RegisterTemplate(shared.WorkflowUserRegistration, &registerWorkflowTemplate)

	userVerificationTemplate := newKubecloudWorkflowTemplate(notificationDispatcher)
	userVerificationTemplate.Steps = []ewf.Step{
		{Name: shared.StepSetupTFChain, RetryPolicy: &ewf.RetryPolicy{
			MaxAttempts: 5,
			BackOff:     ewf.ConstantBackoff(2 * time.Second),
		}},
		{Name: shared.StepCreateStripeCustomer, RetryPolicy: &ewf.RetryPolicy{
			MaxAttempts: 3,
			BackOff:     ewf.ConstantBackoff(2 * time.Second),
		}},
		{Name: shared.StepCreateKYCSponsorship, RetryPolicy: &ewf.RetryPolicy{
			MaxAttempts: 3,
			BackOff:     ewf.ConstantBackoff(2 * time.Second),
		}},
		{Name: shared.StepSendWelcomeEmail, RetryPolicy: &ewf.RetryPolicy{
			MaxAttempts: 3,
			BackOff:     ewf.ConstantBackoff(2 * time.Second),
		}},
	}

	engine.RegisterTemplate(shared.WorkflowUserVerification, &userVerificationTemplate)

	chargeBalanceTemplate := newKubecloudWorkflowTemplate(notificationDispatcher)
	chargeBalanceTemplate.Steps = []ewf.Step{
		{Name: shared.StepCreatePaymentIntent, RetryPolicy: &ewf.RetryPolicy{MaxAttempts: 2, BackOff: ewf.ConstantBackoff(2 * time.Second)}},
		{Name: shared.StepUpdateCreditCardBalance, RetryPolicy: &ewf.RetryPolicy{MaxAttempts: 2, BackOff: ewf.ConstantBackoff(2 * time.Second)}},
		{Name: shared.StepCreatePendingRecord, RetryPolicy: &ewf.RetryPolicy{MaxAttempts: 2, BackOff: ewf.ConstantBackoff(2 * time.Second)}},
	}
	engine.RegisterTemplate(shared.WorkflowChargeBalance, &chargeBalanceTemplate)

	adminCreditBalanceTemplate := newKubecloudWorkflowTemplate(notificationDispatcher)
	adminCreditBalanceTemplate.Steps = []ewf.Step{
		{Name: shared.StepUpdateCreditedBalance, RetryPolicy: &ewf.RetryPolicy{MaxAttempts: 2, BackOff: ewf.ConstantBackoff(2 * time.Second)}},
		{Name: shared.StepCreatePendingRecord, RetryPolicy: &ewf.RetryPolicy{MaxAttempts: 2, BackOff: ewf.ConstantBackoff(2 * time.Second)}},
	}
	engine.RegisterTemplate(shared.WorkflowAdminCreditBalance, &adminCreditBalanceTemplate)

	redeemVoucherTemplate := newKubecloudWorkflowTemplate(notificationDispatcher)
	redeemVoucherTemplate.Steps = []ewf.Step{
		{Name: shared.StepUpdateCreditedBalance, RetryPolicy: &ewf.RetryPolicy{MaxAttempts: 2, BackOff: ewf.ConstantBackoff(2 * time.Second)}},
		{Name: shared.StepCreatePendingRecord, RetryPolicy: &ewf.RetryPolicy{MaxAttempts: 2, BackOff: ewf.ConstantBackoff(2 * time.Second)}},
	}
	engine.RegisterTemplate(shared.WorkflowRedeemVoucher, &redeemVoucherTemplate)

	reserveNodeTemplate := newKubecloudWorkflowTemplate(notificationDispatcher)
	reserveNodeTemplate.Steps = []ewf.Step{
		{Name: shared.StepReserveNode, RetryPolicy: &ewf.RetryPolicy{MaxAttempts: 2, BackOff: ewf.ConstantBackoff(2 * time.Second)}},
		{Name: shared.StepVerifyNodeState, RetryPolicy: &ewf.RetryPolicy{MaxAttempts: 5, BackOff: ewf.ExponentialBackoff(10*time.Second, 2*time.Minute, 2.0)}},
	}
	engine.RegisterTemplate(shared.WorkflowReserveNode, &reserveNodeTemplate)

	unreserveNodeTemplate := newKubecloudWorkflowTemplate(notificationDispatcher)
	unreserveNodeTemplate.Steps = []ewf.Step{
		{Name: shared.StepUnreserveNode, RetryPolicy: &ewf.RetryPolicy{MaxAttempts: 2, BackOff: ewf.ConstantBackoff(2 * time.Second)}},
		{Name: shared.StepVerifyNodeState, RetryPolicy: &ewf.RetryPolicy{MaxAttempts: 5, BackOff: ewf.ExponentialBackoff(10*time.Second, 2*time.Minute, 2.0)}},
	}
	engine.RegisterTemplate(shared.WorkflowUnreserveNode, &unreserveNodeTemplate)

	trackClusterHealthWFTemplate := newKubecloudWorkflowTemplate(notificationDispatcher)
	trackClusterHealthWFTemplate.Steps = []ewf.Step{
		{Name: shared.StepVerifyClusterInDB, RetryPolicy: standardRetryPolicy},
		{Name: shared.StepFetchKubeconfig, RetryPolicy: standardRetryPolicy},
		{Name: shared.StepVerifyClusterReady, RetryPolicy: standardRetryPolicy},
	}
	trackClusterHealthWFTemplate.AfterWorkflowHooks = []ewf.AfterWorkflowHook{hookClusterHealthCheck(notificationDispatcher)}
	// trackClusterHealthWFTemplate.BeforeWorkflowHooks = []ewf.BeforeWorkflowHook{hookNotificationWorkflowStarted}
	engine.RegisterTemplate(shared.WorkflowTrackClusterHealth, &trackClusterHealthWFTemplate)

	registerDeploymentActivities(engine, metrics, clusterRepo, notificationDispatcher, config)

	notificationTemplate := ewf.WorkflowTemplate{
		Steps: []ewf.Step{
			{Name: shared.StepSendUINotification, RetryPolicy: &ewf.RetryPolicy{MaxAttempts: 2, BackOff: ewf.ConstantBackoff(2 * time.Second)}},
			{Name: shared.StepSendEmailNotification, RetryPolicy: &ewf.RetryPolicy{MaxAttempts: 2, BackOff: ewf.ConstantBackoff(2 * time.Second)}},
		},
	}
	engine.RegisterTemplate(shared.WorkflowSendNotification, &notificationTemplate)
}
