package activities

import (
	"kubecloud/internal"
	"kubecloud/internal/constants"
	"kubecloud/internal/metrics"
	"kubecloud/internal/notification"
	"kubecloud/models"
	"time"

	substrate "github.com/threefoldtech/tfchain/clients/tfchain-client-go"
	"github.com/vedhavyas/go-subkey"
	"github.com/xmonader/ewf"
)

var workflowsDescriptions = map[string]string{
	constants.WorkflowAddNode:                  "Adding Node",
	constants.WorkflowRemoveNode:               "Removing Node",
	constants.WorkflowDeleteCluster:            "Deleting Cluster",
	constants.WorkflowDeleteAllClusters:        "Deleting All Clusters",
	constants.WorkflowRollbackFailedDeployment: "Rollback Deployment",
	constants.WorkflowUserRegistration:         "User Registration",
	constants.WorkflowUserVerification:         "User Verification",
	constants.WorkflowChargeBalance:            "Charge Balance",
	constants.WorkflowAdminCreditBalance:       "Admin Credit Balance",
	constants.WorkflowRedeemVoucher:            "Redeem Voucher",
	constants.WorkflowReserveNode:              "Reserve Node",
	constants.WorkflowUnreserveNode:            "Unreserve Node",
}

func RegisterEWFWorkflows(
	engine *ewf.Engine,
	config internal.Configuration,
	db models.DB,
	mail internal.MailService,
	substrate *substrate.Substrate,
	kycClient *internal.KYCClient,
	sponsorAddress string,
	sponsorKeyPair subkey.KeyPair,
	metrics *metrics.Metrics,
	notificationService *notification.NotificationService,
) {
	engine.Register(constants.StepSendVerificationEmail, SendVerificationEmailStep(mail, config))
	engine.Register(constants.StepCreateUser, CreateUserStep(config, db))
	engine.Register(constants.StepUpdateCode, UpdateCodeStep(db))
	engine.Register(constants.StepSetupTFChain, SetupTFChainStep(substrate, config, notificationService, db))
	engine.Register(constants.StepCreateStripeCustomer, CreateStripeCustomerStep(db))
	engine.Register(constants.StepCreateKYCSponsorship, CreateKYCSponsorship(kycClient, notificationService, sponsorAddress, sponsorKeyPair, db))
	engine.Register(constants.StepSendWelcomeEmail, SendWelcomeEmailStep(mail, config, metrics))
	engine.Register(constants.StepCreatePaymentIntent, CreatePaymentIntentStep(config.Currency, metrics, notificationService))
	engine.Register(constants.StepCreatePendingRecord, CreatePendingRecord(substrate, db, config.SystemAccount.Mnemonic, notificationService))
	engine.Register(constants.StepUpdateCreditCardBalance, UpdateCreditCardBalanceStep(db, notificationService))
	engine.Register(constants.StepCreateIdentity, CreateIdentityStep())
	engine.Register(constants.StepReserveNode, ReserveNodeStep(db, substrate))
	engine.Register(constants.StepUnreserveNode, UnreserveNodeStep(db, substrate))
	engine.Register(constants.StepUpdateCreditedBalance, UpdateCreditedBalanceStep(db, notificationService))
	engine.Register(constants.StepSendEmailNotification, SendNotification(db, notificationService.GetNotifiers()[notification.ChannelEmail]))
	engine.Register(constants.StepSendUINotification, SendNotification(db, notificationService.GetNotifiers()[notification.ChannelUI]))

	registerWorkflowTemplate := newKubecloudWorkflowTemplate(notificationService)
	registerWorkflowTemplate.BeforeWorkflowHooks = []ewf.BeforeWorkflowHook{
		hookNotificationWorkflowStarted,
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

	userVerificationTemplate := newKubecloudWorkflowTemplate(notificationService)
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

	chargeBalanceTemplate := newKubecloudWorkflowTemplate(notificationService)
	chargeBalanceTemplate.Steps = []ewf.Step{
		{Name: constants.StepCreatePaymentIntent, RetryPolicy: &ewf.RetryPolicy{MaxAttempts: 2, BackOff: ewf.ConstantBackoff(2 * time.Second)}},
		{Name: constants.StepUpdateCreditCardBalance, RetryPolicy: &ewf.RetryPolicy{MaxAttempts: 2, BackOff: ewf.ConstantBackoff(2 * time.Second)}},
		{Name: constants.StepCreatePendingRecord, RetryPolicy: &ewf.RetryPolicy{MaxAttempts: 2, BackOff: ewf.ConstantBackoff(2 * time.Second)}},
	}
	engine.RegisterTemplate(constants.WorkflowChargeBalance, &chargeBalanceTemplate)

	adminCreditBalanceTemplate := newKubecloudWorkflowTemplate(notificationService)
	adminCreditBalanceTemplate.Steps = []ewf.Step{
		{Name: constants.StepUpdateCreditedBalance, RetryPolicy: &ewf.RetryPolicy{MaxAttempts: 2, BackOff: ewf.ConstantBackoff(2 * time.Second)}},
		{Name: constants.StepCreatePendingRecord, RetryPolicy: &ewf.RetryPolicy{MaxAttempts: 2, BackOff: ewf.ConstantBackoff(2 * time.Second)}},
	}
	engine.RegisterTemplate(constants.WorkflowAdminCreditBalance, &adminCreditBalanceTemplate)

	redeemVoucherTemplate := newKubecloudWorkflowTemplate(notificationService)
	redeemVoucherTemplate.Steps = []ewf.Step{
		{Name: constants.StepUpdateCreditedBalance, RetryPolicy: &ewf.RetryPolicy{MaxAttempts: 2, BackOff: ewf.ConstantBackoff(2 * time.Second)}},
		{Name: constants.StepCreatePendingRecord, RetryPolicy: &ewf.RetryPolicy{MaxAttempts: 2, BackOff: ewf.ConstantBackoff(2 * time.Second)}},
	}
	engine.RegisterTemplate(constants.WorkflowRedeemVoucher, &redeemVoucherTemplate)

	reserveNodeTemplate := newKubecloudWorkflowTemplate(notificationService)
	reserveNodeTemplate.Steps = []ewf.Step{
		{Name: constants.StepCreateIdentity, RetryPolicy: &ewf.RetryPolicy{MaxAttempts: 2, BackOff: ewf.ConstantBackoff(2 * time.Second)}},
		{Name: constants.StepReserveNode, RetryPolicy: &ewf.RetryPolicy{MaxAttempts: 2, BackOff: ewf.ConstantBackoff(2 * time.Second)}},
	}
	engine.RegisterTemplate(constants.WorkflowReserveNode, &reserveNodeTemplate)

	unreserveNodeTemplate := newKubecloudWorkflowTemplate(notificationService)
	unreserveNodeTemplate.Steps = []ewf.Step{
		{Name: constants.StepUnreserveNode, RetryPolicy: &ewf.RetryPolicy{MaxAttempts: 2, BackOff: ewf.ConstantBackoff(2 * time.Second)}},
	}
	engine.RegisterTemplate(constants.WorkflowUnreserveNode, &unreserveNodeTemplate)

	trackClusterHealthWFTemplate := newKubecloudWorkflowTemplate(notificationService)
	trackClusterHealthWFTemplate.Steps = []ewf.Step{
		{Name: constants.StepFetchKubeconfig, RetryPolicy: standardRetryPolicy},
		{Name: constants.StepVerifyClusterReady, RetryPolicy: standardRetryPolicy},
	}
	trackClusterHealthWFTemplate.AfterWorkflowHooks = append(trackClusterHealthWFTemplate.AfterWorkflowHooks, hookClusterHealthCheck(notificationService))
	engine.RegisterTemplate(constants.WorkflowTrackClusterHealth, &trackClusterHealthWFTemplate)

	registerDeploymentActivities(engine, metrics, db, notificationService, config)

	notificationTemplate := newKubecloudWorkflowTemplate(notificationService)
	notificationTemplate.Steps = []ewf.Step{
		{Name: constants.StepSendUINotification, RetryPolicy: &ewf.RetryPolicy{MaxAttempts: 2, BackOff: ewf.ConstantBackoff(2 * time.Second)}},
		{Name: constants.StepSendEmailNotification, RetryPolicy: &ewf.RetryPolicy{MaxAttempts: 2, BackOff: ewf.ConstantBackoff(2 * time.Second)}},
	}
	notificationTemplate.BeforeWorkflowHooks = []ewf.BeforeWorkflowHook{
		hookNotificationWorkflowStarted,
	}
	notificationTemplate.AfterWorkflowHooks = []ewf.AfterWorkflowHook{
		hookWorkflowDone,
	}
	engine.RegisterTemplate(constants.WorkflowSendNotification, &notificationTemplate)
}
