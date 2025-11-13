package activities

import (
	"kubecloud/internal"
	"kubecloud/internal/constants"
	mailservice "kubecloud/internal/mailservice"
	"kubecloud/internal/metrics"
	"kubecloud/internal/notification"
	"kubecloud/internal/substrate"
	"kubecloud/models"
	"time"

	proxy "github.com/threefoldtech/tfgrid-sdk-go/grid-proxy/pkg/client"
	"github.com/vedhavyas/go-subkey"
	"github.com/xmonader/ewf"
)

var workflowsDescriptions = map[string]string{
	constants.WorkflowAddNode:                  "Adding Node",
	constants.WorkflowRemoveNode:               "Removing Node",
	constants.WorkflowDeleteCluster:            "Deleting Cluster",
	constants.WorkflowDeleteAllClusters:        "Deleting All Clusters",
	constants.WorkflowRollbackFailedDeployment: "Rollback",
	constants.WorkflowUserRegistration:         "User Registration",
	constants.WorkflowUserVerification:         "User Verification",
	constants.WorkflowChargeBalance:            "Charge Balance",
	constants.WorkflowAdminCreditBalance:       "Admin Credit Balance",
	constants.WorkflowRedeemVoucher:            "Redeem Voucher",
	constants.WorkflowReserveNode:              "Reserve Node",
	constants.WorkflowUnreserveNode:            "Unreserve Node",
	constants.WorkflowTrackClusterHealth:       "Cluster Health Check",
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

	engine.Register(constants.StepSendVerificationEmail, SendVerificationEmailStep(mail, randomizer, config))
	engine.Register(constants.StepCreateUser, CreateUserStep(config, userRepo))
	engine.Register(constants.StepUpdateCode, UpdateCodeStep(userRepo))
	engine.Register(constants.StepSetupTFChain, SetupTFChainStep(substrate, userRepo))
	engine.Register(constants.StepCreateStripeCustomer, CreateStripeCustomerStep(userRepo))
	engine.Register(constants.StepCreateKYCSponsorship, CreateKYCSponsorship(kycClient, sponsorAddress, sponsorKeyPair, userRepo))
	engine.Register(constants.StepSendWelcomeEmail, SendWelcomeEmailStep(mail, config, metrics))
	engine.Register(constants.StepCreatePaymentIntent, CreatePaymentIntentStep(config.Currency, metrics))
	engine.Register(constants.StepCreatePendingRecord, CreatePendingRecord(substrate, userRepo, pendingRecordRepo, config.SystemAccount.Mnemonic))
	engine.Register(constants.StepUpdateCreditCardBalance, UpdateCreditCardBalanceStep(userRepo))
	engine.Register(constants.StepReserveNode, ReserveNodeStep(userNodesRepo, substrate))
	engine.Register(constants.StepUnreserveNode, UnreserveNodeStep(userNodesRepo, substrate))
	engine.Register(constants.StepUpdateCreditedBalance, UpdateCreditedBalanceStep(userRepo))
	engine.Register(constants.StepSendEmailNotification, SendNotification(userRepo, notificationSender.GetNotifiers()[notification.ChannelEmail]))
	engine.Register(constants.StepSendUINotification, SendNotification(userRepo, notificationSender.GetNotifiers()[notification.ChannelUI]))
	engine.Register(constants.StepVerifyNodeState, VerifyNodeStateStep(proxyClient))
	engine.Register(constants.StepVerifyClusterInDB, VerifyClusterInDBStep(clusterRepo))

	registerWorkflowTemplate := newKubecloudWorkflowTemplate(notificationSender)
	registerWorkflowTemplate.BeforeWorkflowHooks = []ewf.BeforeWorkflowHook{
		// hookNotificationWorkflowStarted,
	}
	registerWorkflowTemplate.Steps = []ewf.Step{
		{Name: constants.StepCreateUser, RetryPolicy: &ewf.RetryPolicy{
			MaxAttempts: 2,
			BackOff:     ewf.ConstantBackoff(2 * time.Second),
		}},
		{Name: constants.StepSendVerificationEmail, RetryPolicy: &ewf.RetryPolicy{
			MaxAttempts: 3,
			BackOff:     ewf.ConstantBackoff(2 * time.Second),
		}},
		{Name: constants.StepUpdateCode, RetryPolicy: &ewf.RetryPolicy{
			MaxAttempts: 2,
			BackOff:     ewf.ConstantBackoff(2 * time.Second),
		}},
	}
	engine.RegisterTemplate(constants.WorkflowUserRegistration, &registerWorkflowTemplate)

	userVerificationTemplate := newKubecloudWorkflowTemplate(notificationSender)
	userVerificationTemplate.Steps = []ewf.Step{
		{Name: constants.StepSetupTFChain, RetryPolicy: &ewf.RetryPolicy{
			MaxAttempts: 5,
			BackOff:     ewf.ConstantBackoff(2 * time.Second),
		}},
		{Name: constants.StepCreateStripeCustomer, RetryPolicy: &ewf.RetryPolicy{
			MaxAttempts: 3,
			BackOff:     ewf.ConstantBackoff(2 * time.Second),
		}},
		{Name: constants.StepCreateKYCSponsorship, RetryPolicy: &ewf.RetryPolicy{
			MaxAttempts: 3,
			BackOff:     ewf.ConstantBackoff(2 * time.Second),
		}},
		{Name: constants.StepSendWelcomeEmail, RetryPolicy: &ewf.RetryPolicy{
			MaxAttempts: 3,
			BackOff:     ewf.ConstantBackoff(2 * time.Second),
		}},
	}

	engine.RegisterTemplate(constants.WorkflowUserVerification, &userVerificationTemplate)

	chargeBalanceTemplate := newKubecloudWorkflowTemplate(notificationSender)
	chargeBalanceTemplate.Steps = []ewf.Step{
		{Name: constants.StepCreatePaymentIntent, RetryPolicy: &ewf.RetryPolicy{MaxAttempts: 2, BackOff: ewf.ConstantBackoff(2 * time.Second)}},
		{Name: constants.StepUpdateCreditCardBalance, RetryPolicy: &ewf.RetryPolicy{MaxAttempts: 2, BackOff: ewf.ConstantBackoff(2 * time.Second)}},
		{Name: constants.StepCreatePendingRecord, RetryPolicy: &ewf.RetryPolicy{MaxAttempts: 2, BackOff: ewf.ConstantBackoff(2 * time.Second)}},
	}
	engine.RegisterTemplate(constants.WorkflowChargeBalance, &chargeBalanceTemplate)

	adminCreditBalanceTemplate := newKubecloudWorkflowTemplate(notificationSender)
	adminCreditBalanceTemplate.Steps = []ewf.Step{
		{Name: constants.StepUpdateCreditedBalance, RetryPolicy: &ewf.RetryPolicy{MaxAttempts: 2, BackOff: ewf.ConstantBackoff(2 * time.Second)}},
		{Name: constants.StepCreatePendingRecord, RetryPolicy: &ewf.RetryPolicy{MaxAttempts: 2, BackOff: ewf.ConstantBackoff(2 * time.Second)}},
	}
	engine.RegisterTemplate(constants.WorkflowAdminCreditBalance, &adminCreditBalanceTemplate)

	redeemVoucherTemplate := newKubecloudWorkflowTemplate(notificationSender)
	redeemVoucherTemplate.Steps = []ewf.Step{
		{Name: constants.StepUpdateCreditedBalance, RetryPolicy: &ewf.RetryPolicy{MaxAttempts: 2, BackOff: ewf.ConstantBackoff(2 * time.Second)}},
		{Name: constants.StepCreatePendingRecord, RetryPolicy: &ewf.RetryPolicy{MaxAttempts: 2, BackOff: ewf.ConstantBackoff(2 * time.Second)}},
	}
	engine.RegisterTemplate(constants.WorkflowRedeemVoucher, &redeemVoucherTemplate)

	reserveNodeTemplate := newKubecloudWorkflowTemplate(notificationSender)
	reserveNodeTemplate.Steps = []ewf.Step{
		{Name: constants.StepReserveNode, RetryPolicy: &ewf.RetryPolicy{MaxAttempts: 2, BackOff: ewf.ConstantBackoff(2 * time.Second)}},
		{Name: constants.StepVerifyNodeState, RetryPolicy: &ewf.RetryPolicy{MaxAttempts: 5, BackOff: ewf.ExponentialBackoff(10*time.Second, 2*time.Minute, 2.0)}},
	}
	engine.RegisterTemplate(constants.WorkflowReserveNode, &reserveNodeTemplate)

	unreserveNodeTemplate := newKubecloudWorkflowTemplate(notificationSender)
	unreserveNodeTemplate.Steps = []ewf.Step{
		{Name: constants.StepUnreserveNode, RetryPolicy: &ewf.RetryPolicy{MaxAttempts: 2, BackOff: ewf.ConstantBackoff(2 * time.Second)}},
		{Name: constants.StepVerifyNodeState, RetryPolicy: &ewf.RetryPolicy{MaxAttempts: 5, BackOff: ewf.ExponentialBackoff(10*time.Second, 2*time.Minute, 2.0)}},
	}
	engine.RegisterTemplate(constants.WorkflowUnreserveNode, &unreserveNodeTemplate)

	trackClusterHealthWFTemplate := newKubecloudWorkflowTemplate(notificationSender)
	trackClusterHealthWFTemplate.Steps = []ewf.Step{
		{Name: constants.StepVerifyClusterInDB, RetryPolicy: standardRetryPolicy},
		{Name: constants.StepFetchKubeconfig, RetryPolicy: standardRetryPolicy},
		{Name: constants.StepVerifyClusterReady, RetryPolicy: standardRetryPolicy},
	}
	trackClusterHealthWFTemplate.AfterWorkflowHooks = []ewf.AfterWorkflowHook{hookClusterHealthCheck(notificationSender)}
	// trackClusterHealthWFTemplate.BeforeWorkflowHooks = []ewf.BeforeWorkflowHook{hookNotificationWorkflowStarted}
	engine.RegisterTemplate(constants.WorkflowTrackClusterHealth, &trackClusterHealthWFTemplate)

	registerDeploymentActivities(engine, metrics, clusterRepo, notificationSender, config)

	notificationTemplate := ewf.WorkflowTemplate{
		Steps: []ewf.Step{
			{Name: constants.StepSendUINotification, RetryPolicy: &ewf.RetryPolicy{MaxAttempts: 2, BackOff: ewf.ConstantBackoff(2 * time.Second)}},
			{Name: constants.StepSendEmailNotification, RetryPolicy: &ewf.RetryPolicy{MaxAttempts: 2, BackOff: ewf.ConstantBackoff(2 * time.Second)}},
		},
	}
	engine.RegisterTemplate(constants.WorkflowSendNotification, &notificationTemplate)
}
