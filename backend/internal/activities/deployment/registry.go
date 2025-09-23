package deployment

import (
	"kubecloud/internal"
	"kubecloud/internal/metrics"
	"kubecloud/internal/notification"
	"kubecloud/internal/workflow"
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
	engine          *ewf.Engine
	metrics         *metrics.Metrics
	db              models.DB
	ns              *notification.NotificationService
	config          internal.Configuration
	templateBuilder workflow.TemplateBuilderFunc
}

func NewTemplateRegistry(
	engine *ewf.Engine,
	metrics *metrics.Metrics,
	db models.DB,
	ns *notification.NotificationService,
	config internal.Configuration,
	templateBuilder workflow.TemplateBuilderFunc,
) *TemplateRegistry {
	return &TemplateRegistry{
		engine:          engine,
		metrics:         metrics,
		db:              db,
		ns:              ns,
		config:          config,
		templateBuilder: templateBuilder,
	}
}

// RegisterAll registers all deployment steps and workflow templates
func (tr *TemplateRegistry) RegisterAll() {
	steps := map[string]ewf.StepFn{
		// Network steps
		workflow.StepDeployNetwork: DeployNetworkStep(tr.metrics),
		workflow.StepUpdateNetwork: UpdateNetworkStep(tr.metrics),

		// Node steps
		workflow.StepDeployNode: DeployNodeStep(tr.metrics),
		workflow.StepAddNode:    AddNodeStep(tr.metrics),
		workflow.StepRemoveNode: RemoveNodeStep(tr.metrics),

		// Cluster management steps
		workflow.StepRemoveCluster:      CancelDeploymentStep(tr.db, tr.metrics),
		workflow.StepFetchKubeconfig:    FetchKubeconfigStep(tr.db, tr.config.SSH.PrivateKeyPath),
		workflow.StepVerifyClusterReady: VerifyClusterReadyStep(),
		workflow.StepVerifyNewNodes:     VerifyAddedNodeStep(tr.db, tr.config.SSH.PrivateKeyPath),

		// Database steps
		workflow.StepStoreDeployment:     StoreDeploymentStep(tr.db, tr.metrics),
		workflow.StepRemoveClusterFromDB: RemoveClusterFromDBStep(tr.db),

		// Bulk operations steps
		workflow.StepGatherAllContractIDs:  GatherAllContractIDsStep(tr.db),
		workflow.StepBatchCancelContracts:  BatchCancelContractsStep(),
		workflow.StepDeleteAllUserClusters: DeleteAllUserClustersStep(tr.db),
	}

	for stepName, stepFn := range steps {
		tr.engine.Register(stepName, stepFn)
	}

	templates := map[string]ewf.WorkflowTemplate{
		workflow.WorkflowAddNode:                  buildAddNodeTemplate(tr.templateBuilder),
		workflow.WorkflowRemoveNode:               buildRemoveNodeTemplate(tr.templateBuilder),
		workflow.WorkflowDeleteCluster:            buildDeleteClusterTemplate(tr.templateBuilder),
		workflow.WorkflowDeleteAllClusters:        buildDeleteAllClustersTemplate(tr.templateBuilder),
		workflow.WorkflowRollbackFailedDeployment: buildRollbackFailedDeploymentTemplate(tr.templateBuilder),
		workflow.WorkflowRollbackAddNode:          buildRollbackAddNodeTemplate(tr.templateBuilder),
	}

	for templateName, template := range templates {
		tr.engine.RegisterTemplate(templateName, &template)
	}
}

// RegisterDynamicDeployWorkflow creates and registers a dynamic deployment workflow with the specified number of nodes
func (tr *TemplateRegistry) RegisterDynamicDeployWorkflow(wfName string, nodesNum int) {
	steps := []ewf.Step{
		{Name: workflow.StepDeployNetwork, RetryPolicy: criticalRetryPolicy},
	}

	for i := 0; i < nodesNum; i++ {
		stepName := getDeployNodeStepName(i + 1)
		tr.engine.Register(stepName, DeployNodeStep(tr.metrics))
		steps = append(steps, ewf.Step{Name: stepName, RetryPolicy: criticalRetryPolicy})
	}

	steps = append(steps,
		ewf.Step{Name: workflow.StepFetchKubeconfig, RetryPolicy: criticalRetryPolicy},
		// TODO: review the verify logic. if it fails?
		ewf.Step{Name: workflow.StepVerifyClusterReady, RetryPolicy: longExponentialRetryPolicy},
		ewf.Step{Name: workflow.StepStoreDeployment, RetryPolicy: standardRetryPolicy},
	)

	if tr.templateBuilder != nil {
		template := tr.templateBuilder(
			steps,
			[]ewf.AfterWorkflowHook{deployFailureHook(*tr.engine, tr.metrics), closeClientHook()},
			nil,
			nil,
			nil,
		)

		tr.engine.RegisterTemplate(wfName, &template)
	}
}
