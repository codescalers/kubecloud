package activities

import (
	"kubecloud/internal"
	mailservice "kubecloud/internal/mailservice"
	"kubecloud/internal/metrics"
	"kubecloud/internal/notification"
	"kubecloud/internal/substrate"
	"kubecloud/internal/models"
	"time"

	proxy "github.com/threefoldtech/tfgrid-sdk-go/grid-proxy/pkg/client"
	"github.com/vedhavyas/go-subkey"
	"github.com/xmonader/ewf"
)

var workflowsDescriptions = map[string]string{
	internal.WorkflowAddNode:                  "Adding Node",
	internal.WorkflowRemoveNode:               "Removing Node",
	internal.WorkflowDeleteCluster:            "Deleting Cluster",
	internal.WorkflowDeleteAllClusters:        "Deleting All Clusters",
	internal.WorkflowRollbackFailedDeployment: "Rollback",
	internal.WorkflowUserRegistration:         "User Registration",
	internal.WorkflowUserVerification:         "User Verification",
	internal.WorkflowChargeBalance:            "Charge Balance",
	internal.WorkflowAdminCreditBalance:       "Admin Credit Balance",
	internal.WorkflowRedeemVoucher:            "Redeem Voucher",
	internal.WorkflowReserveNode:              "Reserve Node",
	internal.WorkflowUnreserveNode:            "Unreserve Node",
	internal.WorkflowTrackClusterHealth:       "Cluster Health Check",
}

func RegisterEWFWorkflows(
	engine *ewf.Engine,
	config internal.Configuration,
	db models.DB,
	mail mailservice.MailService,
	substrate substrate.Substrate,
	kycClient *internal.KYCClient,
	sponsorAddress string,
	sponsorKeyPair subkey.KeyPair,
	metrics *metrics.Metrics,
	notificationSender notification.NotificationSender,
	proxyClient proxy.Client,
	randomizer internal.Randomizer,
) {
	userRepo := models.NewGormUserRepository(db)
	clusterRepo := models.NewGormClusterRepository(db)
	userNodesRepo := models.NewGormUserNodesRepository(db)
	pendingRecordRepo := models.NewGormPendingRecordRepository(db)

	engine.Register(internal.StepSendVerificationEmail, SendVerificationEmailStep(mail, randomizer, config))
	engine.Register(internal.StepCreateUser, CreateUserStep(config, userRepo))
	engine.Register(internal.StepUpdateCode, UpdateCodeStep(userRepo))
	engine.Register(internal.StepSetupTFChain, SetupTFChainStep(substrate, userRepo))
	engine.Register(internal.StepCreateStripeCustomer, CreateStripeCustomerStep(userRepo))
	engine.Register(internal.StepCreateKYCSponsorship, CreateKYCSponsorship(kycClient, sponsorAddress, sponsorKeyPair, userRepo))
	engine.Register(internal.StepSendWelcomeEmail, SendWelcomeEmailStep(mail, config, metrics))
	engine.Register(internal.StepCreatePaymentIntent, CreatePaymentIntentStep(config.Currency, metrics))
	engine.Register(internal.StepCreatePendingRecord, CreatePendingRecord(substrate, userRepo, pendingRecordRepo, config.SystemAccount.Mnemonic))
	engine.Register(internal.StepUpdateCreditCardBalance, UpdateCreditCardBalanceStep(userRepo))
	engine.Register(internal.StepReserveNode, ReserveNodeStep(userNodesRepo, substrate))
	engine.Register(internal.StepUnreserveNode, UnreserveNodeStep(userNodesRepo, substrate))
	engine.Register(internal.StepUpdateCreditedBalance, UpdateCreditedBalanceStep(userRepo))
	engine.Register(internal.StepSendEmailNotification, SendNotification(userRepo, notificationSender.GetNotifiers()[notification.ChannelEmail]))
	engine.Register(internal.StepSendUINotification, SendNotification(userRepo, notificationSender.GetNotifiers()[notification.ChannelUI]))
	engine.Register(internal.StepVerifyNodeState, VerifyNodeStateStep(proxyClient))
	engine.Register(internal.StepVerifyClusterInDB, VerifyClusterInDBStep(clusterRepo))

	registerWorkflowTemplate := newKubecloudWorkflowTemplate(notificationSender)
	registerWorkflowTemplate.BeforeWorkflowHooks = []ewf.BeforeWorkflowHook{
		// hookNotificationWorkflowStarted,
	}
	registerWorkflowTemplate.Steps = []ewf.Step{
		{Name: internal.StepCreateUser, RetryPolicy: &ewf.RetryPolicy{
			MaxAttempts: 2,
			BackOff:     ewf.ConstantBackoff(2 * time.Second),
		}},
		{Name: internal.StepSendVerificationEmail, RetryPolicy: &ewf.RetryPolicy{
			MaxAttempts: 3,
			BackOff:     ewf.ConstantBackoff(2 * time.Second),
		}},
		{Name: internal.StepUpdateCode, RetryPolicy: &ewf.RetryPolicy{
			MaxAttempts: 2,
			BackOff:     ewf.ConstantBackoff(2 * time.Second),
		}},
	}
	engine.RegisterTemplate(internal.WorkflowUserRegistration, &registerWorkflowTemplate)

	userVerificationTemplate := newKubecloudWorkflowTemplate(notificationSender)
	userVerificationTemplate.Steps = []ewf.Step{
		{Name: internal.StepSetupTFChain, RetryPolicy: &ewf.RetryPolicy{
			MaxAttempts: 5,
			BackOff:     ewf.ConstantBackoff(2 * time.Second),
		}},
		{Name: internal.StepCreateStripeCustomer, RetryPolicy: &ewf.RetryPolicy{
			MaxAttempts: 3,
			BackOff:     ewf.ConstantBackoff(2 * time.Second),
		}},
		{Name: internal.StepCreateKYCSponsorship, RetryPolicy: &ewf.RetryPolicy{
			MaxAttempts: 3,
			BackOff:     ewf.ConstantBackoff(2 * time.Second),
		}},
		{Name: internal.StepSendWelcomeEmail, RetryPolicy: &ewf.RetryPolicy{
			MaxAttempts: 3,
			BackOff:     ewf.ConstantBackoff(2 * time.Second),
		}},
	}

	engine.RegisterTemplate(internal.WorkflowUserVerification, &userVerificationTemplate)

	chargeBalanceTemplate := newKubecloudWorkflowTemplate(notificationSender)
	chargeBalanceTemplate.Steps = []ewf.Step{
		{Name: internal.StepCreatePaymentIntent, RetryPolicy: &ewf.RetryPolicy{MaxAttempts: 2, BackOff: ewf.ConstantBackoff(2 * time.Second)}},
		{Name: internal.StepUpdateCreditCardBalance, RetryPolicy: &ewf.RetryPolicy{MaxAttempts: 2, BackOff: ewf.ConstantBackoff(2 * time.Second)}},
		{Name: internal.StepCreatePendingRecord, RetryPolicy: &ewf.RetryPolicy{MaxAttempts: 2, BackOff: ewf.ConstantBackoff(2 * time.Second)}},
	}
	engine.RegisterTemplate(internal.WorkflowChargeBalance, &chargeBalanceTemplate)

	adminCreditBalanceTemplate := newKubecloudWorkflowTemplate(notificationSender)
	adminCreditBalanceTemplate.Steps = []ewf.Step{
		{Name: internal.StepUpdateCreditedBalance, RetryPolicy: &ewf.RetryPolicy{MaxAttempts: 2, BackOff: ewf.ConstantBackoff(2 * time.Second)}},
		{Name: internal.StepCreatePendingRecord, RetryPolicy: &ewf.RetryPolicy{MaxAttempts: 2, BackOff: ewf.ConstantBackoff(2 * time.Second)}},
	}
	engine.RegisterTemplate(internal.WorkflowAdminCreditBalance, &adminCreditBalanceTemplate)

	redeemVoucherTemplate := newKubecloudWorkflowTemplate(notificationSender)
	redeemVoucherTemplate.Steps = []ewf.Step{
		{Name: internal.StepUpdateCreditedBalance, RetryPolicy: &ewf.RetryPolicy{MaxAttempts: 2, BackOff: ewf.ConstantBackoff(2 * time.Second)}},
		{Name: internal.StepCreatePendingRecord, RetryPolicy: &ewf.RetryPolicy{MaxAttempts: 2, BackOff: ewf.ConstantBackoff(2 * time.Second)}},
	}
	engine.RegisterTemplate(internal.WorkflowRedeemVoucher, &redeemVoucherTemplate)

	reserveNodeTemplate := newKubecloudWorkflowTemplate(notificationSender)
	reserveNodeTemplate.Steps = []ewf.Step{
		{Name: internal.StepReserveNode, RetryPolicy: &ewf.RetryPolicy{MaxAttempts: 2, BackOff: ewf.ConstantBackoff(2 * time.Second)}},
		{Name: internal.StepVerifyNodeState, RetryPolicy: &ewf.RetryPolicy{MaxAttempts: 5, BackOff: ewf.ExponentialBackoff(10*time.Second, 2*time.Minute, 2.0)}},
	}
	engine.RegisterTemplate(internal.WorkflowReserveNode, &reserveNodeTemplate)

	unreserveNodeTemplate := newKubecloudWorkflowTemplate(notificationSender)
	unreserveNodeTemplate.Steps = []ewf.Step{
		{Name: internal.StepUnreserveNode, RetryPolicy: &ewf.RetryPolicy{MaxAttempts: 2, BackOff: ewf.ConstantBackoff(2 * time.Second)}},
		{Name: internal.StepVerifyNodeState, RetryPolicy: &ewf.RetryPolicy{MaxAttempts: 5, BackOff: ewf.ExponentialBackoff(10*time.Second, 2*time.Minute, 2.0)}},
	}
	engine.RegisterTemplate(internal.WorkflowUnreserveNode, &unreserveNodeTemplate)

	trackClusterHealthWFTemplate := newKubecloudWorkflowTemplate(notificationSender)
	trackClusterHealthWFTemplate.Steps = []ewf.Step{
		{Name: internal.StepVerifyClusterInDB, RetryPolicy: standardRetryPolicy},
		{Name: internal.StepFetchKubeconfig, RetryPolicy: standardRetryPolicy},
		{Name: internal.StepVerifyClusterReady, RetryPolicy: standardRetryPolicy},
	}
	trackClusterHealthWFTemplate.AfterWorkflowHooks = []ewf.AfterWorkflowHook{hookClusterHealthCheck(notificationSender)}
	// trackClusterHealthWFTemplate.BeforeWorkflowHooks = []ewf.BeforeWorkflowHook{hookNotificationWorkflowStarted}
	engine.RegisterTemplate(internal.WorkflowTrackClusterHealth, &trackClusterHealthWFTemplate)

	registerDeploymentActivities(engine, metrics, clusterRepo, notificationSender, config)

	notificationTemplate := ewf.WorkflowTemplate{
		Steps: []ewf.Step{
			{Name: internal.StepSendUINotification, RetryPolicy: &ewf.RetryPolicy{MaxAttempts: 2, BackOff: ewf.ConstantBackoff(2 * time.Second)}},
			{Name: internal.StepSendEmailNotification, RetryPolicy: &ewf.RetryPolicy{MaxAttempts: 2, BackOff: ewf.ConstantBackoff(2 * time.Second)}},
		},
	}
	engine.RegisterTemplate(internal.WorkflowSendNotification, &notificationTemplate)
}
