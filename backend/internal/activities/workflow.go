package activities

import (
	"kubecloud/internal"
	"kubecloud/internal/activities/deployment"
	"kubecloud/internal/metrics"
	"kubecloud/internal/notification"
	"kubecloud/internal/workflow"
	"kubecloud/models"
	"time"

	substrate "github.com/threefoldtech/tfchain/clients/tfchain-client-go"
	proxy "github.com/threefoldtech/tfgrid-sdk-go/grid-proxy/pkg/client"
	"github.com/vedhavyas/go-subkey"
	"github.com/xmonader/ewf"
)

var workflowsDescriptions = map[string]string{
	workflow.WorkflowAddNode:                  "Adding Node",
	workflow.WorkflowRemoveNode:               "Removing Node",
	workflow.WorkflowDeleteCluster:            "Deleting Cluster",
	workflow.WorkflowDeleteAllClusters:        "Deleting All Clusters",
	workflow.WorkflowRollbackFailedDeployment: "Rollback",
	workflow.WorkflowUserRegistration:         "User Registration",
	workflow.WorkflowUserVerification:         "User Verification",
	workflow.WorkflowChargeBalance:            "Charge Balance",
	workflow.WorkflowAdminCreditBalance:       "Admin Credit Balance",
	workflow.WorkflowRedeemVoucher:            "Redeem Voucher",
	workflow.WorkflowReserveNode:              "Reserve Node",
	workflow.WorkflowUnreserveNode:            "Unreserve Node",
	workflow.WorkflowTrackClusterHealth:       "Cluster Health Check",
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
	proxyClient proxy.Client,
) {
	engine.Register(workflow.StepSendVerificationEmail, SendVerificationEmailStep(mail, config))
	engine.Register(workflow.StepCreateUser, CreateUserStep(config, db))
	engine.Register(workflow.StepUpdateCode, UpdateCodeStep(db))
	engine.Register(workflow.StepSetupTFChain, SetupTFChainStep(substrate, config, notificationService, db))
	engine.Register(workflow.StepCreateStripeCustomer, CreateStripeCustomerStep(db))
	engine.Register(workflow.StepCreateKYCSponsorship, CreateKYCSponsorship(kycClient, notificationService, sponsorAddress, sponsorKeyPair, db))
	engine.Register(workflow.StepSendWelcomeEmail, SendWelcomeEmailStep(mail, config, metrics))
	engine.Register(workflow.StepCreatePaymentIntent, CreatePaymentIntentStep(config.Currency, metrics, notificationService))
	engine.Register(workflow.StepCreatePendingRecord, CreatePendingRecord(substrate, db, config.SystemAccount.Mnemonic))
	engine.Register(workflow.StepUpdateCreditCardBalance, UpdateCreditCardBalanceStep(db))
	engine.Register(workflow.StepCreateIdentity, CreateIdentityStep())
	engine.Register(workflow.StepReserveNode, ReserveNodeStep(db, substrate))
	engine.Register(workflow.StepUnreserveNode, UnreserveNodeStep(db, substrate))
	engine.Register(workflow.StepUpdateCreditedBalance, UpdateCreditedBalanceStep(db))
	engine.Register(workflow.StepSendEmailNotification, SendNotification(db, notificationService.GetNotifiers()[notification.ChannelEmail]))
	engine.Register(workflow.StepSendUINotification, SendNotification(db, notificationService.GetNotifiers()[notification.ChannelUI]))
	engine.Register(workflow.StepVerifyNodeState, VerifyNodeStateStep(proxyClient))

	registerWorkflowTemplate := newKubecloudWorkflowTemplate(notificationService)
	registerWorkflowTemplate.BeforeWorkflowHooks = []ewf.BeforeWorkflowHook{
		hookLogWorkflowStarted,
	}
	registerWorkflowTemplate.Steps = []ewf.Step{
		{Name: workflow.StepCreateUser, RetryPolicy: &ewf.RetryPolicy{
			MaxAttempts: 2,
			BackOff:     ewf.ConstantBackoff(2 * time.Second),
		}},
		{Name: workflow.StepSendVerificationEmail, RetryPolicy: &ewf.RetryPolicy{
			MaxAttempts: 3,
			BackOff:     ewf.ConstantBackoff(2 * time.Second),
		}},
		{Name: workflow.StepUpdateCode, RetryPolicy: &ewf.RetryPolicy{
			MaxAttempts: 2,
			BackOff:     ewf.ConstantBackoff(2 * time.Second),
		}},
	}
	engine.RegisterTemplate(workflow.WorkflowUserRegistration, &registerWorkflowTemplate)

	userVerificationTemplate := newKubecloudWorkflowTemplate(notificationService)
	userVerificationTemplate.Steps = []ewf.Step{
		{Name: workflow.StepSetupTFChain, RetryPolicy: &ewf.RetryPolicy{
			MaxAttempts: 5,
			BackOff:     ewf.ConstantBackoff(2 * time.Second),
		}},
		{Name: workflow.StepCreateStripeCustomer, RetryPolicy: &ewf.RetryPolicy{
			MaxAttempts: 3,
			BackOff:     ewf.ConstantBackoff(2 * time.Second),
		}},
		{Name: workflow.StepCreateKYCSponsorship, RetryPolicy: &ewf.RetryPolicy{
			MaxAttempts: 3,
			BackOff:     ewf.ConstantBackoff(2 * time.Second),
		}},
		{Name: workflow.StepSendWelcomeEmail, RetryPolicy: &ewf.RetryPolicy{
			MaxAttempts: 3,
			BackOff:     ewf.ConstantBackoff(2 * time.Second),
		}},
	}

	engine.RegisterTemplate(workflow.WorkflowUserVerification, &userVerificationTemplate)

	chargeBalanceTemplate := newKubecloudWorkflowTemplate(notificationService)
	chargeBalanceTemplate.Steps = []ewf.Step{
		{Name: workflow.StepCreatePaymentIntent, RetryPolicy: &ewf.RetryPolicy{MaxAttempts: 2, BackOff: ewf.ConstantBackoff(2 * time.Second)}},
		{Name: workflow.StepUpdateCreditCardBalance, RetryPolicy: &ewf.RetryPolicy{MaxAttempts: 2, BackOff: ewf.ConstantBackoff(2 * time.Second)}},
		{Name: workflow.StepCreatePendingRecord, RetryPolicy: &ewf.RetryPolicy{MaxAttempts: 2, BackOff: ewf.ConstantBackoff(2 * time.Second)}},
	}
	engine.RegisterTemplate(workflow.WorkflowChargeBalance, &chargeBalanceTemplate)

	adminCreditBalanceTemplate := newKubecloudWorkflowTemplate(notificationService)
	adminCreditBalanceTemplate.Steps = []ewf.Step{
		{Name: workflow.StepUpdateCreditedBalance, RetryPolicy: &ewf.RetryPolicy{MaxAttempts: 2, BackOff: ewf.ConstantBackoff(2 * time.Second)}},
		{Name: workflow.StepCreatePendingRecord, RetryPolicy: &ewf.RetryPolicy{MaxAttempts: 2, BackOff: ewf.ConstantBackoff(2 * time.Second)}},
	}
	engine.RegisterTemplate(workflow.WorkflowAdminCreditBalance, &adminCreditBalanceTemplate)

	redeemVoucherTemplate := newKubecloudWorkflowTemplate(notificationService)
	redeemVoucherTemplate.Steps = []ewf.Step{
		{Name: workflow.StepUpdateCreditedBalance, RetryPolicy: &ewf.RetryPolicy{MaxAttempts: 2, BackOff: ewf.ConstantBackoff(2 * time.Second)}},
		{Name: workflow.StepCreatePendingRecord, RetryPolicy: &ewf.RetryPolicy{MaxAttempts: 2, BackOff: ewf.ConstantBackoff(2 * time.Second)}},
	}
	engine.RegisterTemplate(workflow.WorkflowRedeemVoucher, &redeemVoucherTemplate)

	reserveNodeTemplate := newKubecloudWorkflowTemplate(notificationService)
	reserveNodeTemplate.Steps = []ewf.Step{
		{Name: workflow.StepCreateIdentity, RetryPolicy: &ewf.RetryPolicy{MaxAttempts: 2, BackOff: ewf.ConstantBackoff(2 * time.Second)}},
		{Name: workflow.StepReserveNode, RetryPolicy: &ewf.RetryPolicy{MaxAttempts: 2, BackOff: ewf.ConstantBackoff(2 * time.Second)}},
		{Name: workflow.StepVerifyNodeState, RetryPolicy: &ewf.RetryPolicy{MaxAttempts: 5, BackOff: ewf.ExponentialBackoff(10*time.Second, 2*time.Minute, 2.0)}},
	}
	engine.RegisterTemplate(workflow.WorkflowReserveNode, &reserveNodeTemplate)

	unreserveNodeTemplate := newKubecloudWorkflowTemplate(notificationService)
	unreserveNodeTemplate.Steps = []ewf.Step{
		{Name: workflow.StepUnreserveNode, RetryPolicy: &ewf.RetryPolicy{MaxAttempts: 2, BackOff: ewf.ConstantBackoff(2 * time.Second)}},
		{Name: workflow.StepVerifyNodeState, RetryPolicy: &ewf.RetryPolicy{MaxAttempts: 5, BackOff: ewf.ExponentialBackoff(10*time.Second, 2*time.Minute, 2.0)}},
	}
	engine.RegisterTemplate(workflow.WorkflowUnreserveNode, &unreserveNodeTemplate)

	trackClusterHealthWFTemplate := newKubecloudWorkflowTemplate(notificationService)
	trackClusterHealthWFTemplate.Steps = []ewf.Step{
		{Name: workflow.StepFetchKubeconfig, RetryPolicy: standardRetryPolicy},
		{Name: workflow.StepVerifyClusterReady, RetryPolicy: standardRetryPolicy},
	}
	trackClusterHealthWFTemplate.AfterWorkflowHooks = append(trackClusterHealthWFTemplate.AfterWorkflowHooks, hookClusterHealthCheck(notificationService))
	trackClusterHealthWFTemplate.BeforeWorkflowHooks = []ewf.BeforeWorkflowHook{hookNotifyWorkflowStarted(notificationService)}
	engine.RegisterTemplate(workflow.WorkflowTrackClusterHealth, &trackClusterHealthWFTemplate)

	// registerDeploymentActivities(engine, metrics, db, notificationService, config)
	templateBuilder := CreateTemplateBuilder(notificationService)
	registry := deployment.NewTemplateRegistry(engine, metrics, db, notificationService, config, templateBuilder)
	registry.RegisterAll()

	notificationTemplate := newKubecloudWorkflowTemplate(notificationService)
	notificationTemplate.Steps = []ewf.Step{
		{Name: workflow.StepSendUINotification, RetryPolicy: &ewf.RetryPolicy{MaxAttempts: 2, BackOff: ewf.ConstantBackoff(2 * time.Second)}},
		{Name: workflow.StepSendEmailNotification, RetryPolicy: &ewf.RetryPolicy{MaxAttempts: 2, BackOff: ewf.ConstantBackoff(2 * time.Second)}},
	}
	notificationTemplate.BeforeWorkflowHooks = []ewf.BeforeWorkflowHook{
		hookLogWorkflowStarted,
	}
	notificationTemplate.AfterWorkflowHooks = []ewf.AfterWorkflowHook{
		hookLogWorkflowDone,
	}
	engine.RegisterTemplate(workflow.WorkflowSendNotification, &notificationTemplate)
}
