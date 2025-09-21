package deployment

import (
	"kubecloud/internal"
	"kubecloud/internal/constants"
	"kubecloud/internal/metrics"
	"kubecloud/internal/notification"
	"kubecloud/models"
	"time"

	"github.com/xmonader/ewf"
)

var (
	criticalRetryPolicy        = &ewf.RetryPolicy{MaxAttempts: 5, BackOff: ewf.ConstantBackoff(5 * time.Second)}
	standardRetryPolicy        = &ewf.RetryPolicy{MaxAttempts: 2, BackOff: ewf.ConstantBackoff(2 * time.Second)}
	longExponentialRetryPolicy = &ewf.RetryPolicy{MaxAttempts: 5, BackOff: ewf.ExponentialBackoff(30*time.Second, 5*time.Minute, 2.0)}
)

// TemplateRegistry manages deployment workflow and step registration
type TemplateRegistry struct {
	engine  *ewf.Engine
	metrics *metrics.Metrics
	db      models.DB
	ns      *notification.NotificationService
	config  internal.Configuration
}

func NewTemplateRegistry(
	engine *ewf.Engine,
	metrics *metrics.Metrics,
	db models.DB,
	ns *notification.NotificationService,
	config internal.Configuration,
) *TemplateRegistry {
	return &TemplateRegistry{
		engine:  engine,
		metrics: metrics,
		db:      db,
		ns:      ns,
		config:  config,
	}
}

// RegisterAll registers all deployment steps and workflow templates
func (tr *TemplateRegistry) RegisterAll() {
	steps := map[string]ewf.StepFn{
		// Network steps
		constants.StepDeployNetwork: DeployNetworkStep(tr.metrics),
		constants.StepUpdateNetwork: UpdateNetworkStep(tr.metrics),

		// Node steps
		constants.StepDeployNode: DeployNodeStep(tr.metrics),
		constants.StepAddNode:    AddNodeStep(tr.metrics),
		constants.StepRemoveNode: RemoveNodeStep(tr.metrics),

		// Cluster management steps
		constants.StepRemoveCluster:      CancelDeploymentStep(tr.db, tr.metrics),
		constants.StepFetchKubeconfig:    FetchKubeconfigStep(tr.db, tr.config.SSH.PrivateKeyPath),
		constants.StepVerifyClusterReady: VerifyClusterReadyStep(),
		constants.StepVerifyNewNodes:     VerifyAddedNodeStep(tr.db, tr.config.SSH.PrivateKeyPath),

		// Database steps
		constants.StepStoreDeployment:     StoreDeploymentStep(tr.db, tr.metrics),
		constants.StepRemoveClusterFromDB: RemoveClusterFromDBStep(tr.db),

		// Bulk operations steps
		constants.StepGatherAllContractIDs:  GatherAllContractIDsStep(tr.db),
		constants.StepBatchCancelContracts:  BatchCancelContractsStep(),
		constants.StepDeleteAllUserClusters: DeleteAllUserClustersStep(tr.db),
	}

	for stepName, stepFn := range steps {
		tr.engine.Register(stepName, stepFn)
	}

	templates := map[string]ewf.WorkflowTemplate{
		constants.WorkflowAddNode:                  buildAddNodeTemplate(tr.ns),
		constants.WorkflowDeleteCluster:            buildDeleteClusterTemplate(tr.ns),
		constants.WorkflowDeleteAllClusters:        buildDeleteAllClustersTemplate(tr.ns),
		constants.WorkflowRemoveNode:               buildRemoveNodeTemplate(tr.ns),
		constants.WorkflowRollbackFailedDeployment: buildRollbackFailedDeploymentTemplate(tr.ns),
		constants.WorkflowRollbackAddNode:          buildRollbackAddNodeTemplate(tr.ns),
	}

	for templateName, template := range templates {
		tr.engine.RegisterTemplate(templateName, &template)
	}
}

// RegisterDynamicDeployWorkflow creates and registers a dynamic deployment workflow with the specified number of nodes
func (tr *TemplateRegistry) RegisterDynamicDeployWorkflow(wfName string, nodesNum int) {
	steps := []ewf.Step{
		{Name: constants.StepDeployNetwork, RetryPolicy: criticalRetryPolicy},
	}

	for i := 0; i < nodesNum; i++ {
		stepName := getDeployNodeStepName(i + 1)
		tr.engine.Register(stepName, DeployNodeStep(tr.metrics))
		steps = append(steps, ewf.Step{Name: stepName, RetryPolicy: criticalRetryPolicy})
	}

	steps = append(steps,
		ewf.Step{Name: constants.StepFetchKubeconfig, RetryPolicy: criticalRetryPolicy},
		// TODO: review the verify logic. if it fails?
		ewf.Step{Name: constants.StepVerifyClusterReady, RetryPolicy: longExponentialRetryPolicy},
		ewf.Step{Name: constants.StepStoreDeployment, RetryPolicy: standardRetryPolicy},
	)

	template := ewf.WorkflowTemplate{
		Steps: steps,
		BeforeWorkflowHooks: []ewf.BeforeWorkflowHook{
			notifyWorkflowStartedHook(tr.ns),
			logWorkflowStartedHook(),
		},
		AfterWorkflowHooks: []ewf.AfterWorkflowHook{
			notifyWorkflowProgressHook(tr.ns),
			deployFailureHook(*tr.engine, tr.metrics),
			closeClientHook(),
			logWorkflowDoneHook(),
		},
		BeforeStepHooks: []ewf.BeforeStepHook{
			logStepStartedHook(),
		},
		AfterStepHooks: []ewf.AfterStepHook{
			notifyStepProgressHook(tr.ns),
			logStepDoneHook(),
		},
	}

	tr.engine.RegisterTemplate(wfName, &template)
}
