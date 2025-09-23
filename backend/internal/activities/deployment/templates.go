package deployment

import (
	"kubecloud/internal/workflow"

	"github.com/xmonader/ewf"
)

func buildAddNodeTemplate(templateBuilder workflow.TemplateBuilderFunc) ewf.WorkflowTemplate {
	return templateBuilder(
		[]ewf.Step{
			{Name: workflow.StepUpdateNetwork, RetryPolicy: criticalRetryPolicy},
			{Name: workflow.StepAddNode, RetryPolicy: standardRetryPolicy},
			{Name: workflow.StepFetchKubeconfig, RetryPolicy: criticalRetryPolicy},
			{Name: workflow.StepVerifyNewNodes, RetryPolicy: longExponentialRetryPolicy},
			{Name: workflow.StepStoreDeployment, RetryPolicy: standardRetryPolicy},
		},
		[]ewf.AfterWorkflowHook{addNodeFailureHook(), closeClientHook()},
		nil,
		nil,
		nil,
	)
}

func buildRemoveNodeTemplate(templateBuilder workflow.TemplateBuilderFunc) ewf.WorkflowTemplate {
	return templateBuilder(
		[]ewf.Step{
			{Name: workflow.StepRemoveNode, RetryPolicy: standardRetryPolicy},
			{Name: workflow.StepStoreDeployment, RetryPolicy: standardRetryPolicy},
		},
		[]ewf.AfterWorkflowHook{closeClientHook()},
		nil,
		nil,
		nil,
	)
}

func buildDeleteClusterTemplate(templateBuilder workflow.TemplateBuilderFunc) ewf.WorkflowTemplate {
	return templateBuilder(
		[]ewf.Step{
			{Name: workflow.StepRemoveCluster, RetryPolicy: standardRetryPolicy},
			{Name: workflow.StepRemoveClusterFromDB, RetryPolicy: standardRetryPolicy},
		},
		[]ewf.AfterWorkflowHook{closeClientHook()},
		nil,
		nil,
		nil,
	)
}

func buildDeleteAllClustersTemplate(templateBuilder workflow.TemplateBuilderFunc) ewf.WorkflowTemplate {
	return templateBuilder(
		[]ewf.Step{
			{Name: workflow.StepGatherAllContractIDs, RetryPolicy: standardRetryPolicy},
			{Name: workflow.StepBatchCancelContracts, RetryPolicy: standardRetryPolicy},
			{Name: workflow.StepDeleteAllUserClusters, RetryPolicy: standardRetryPolicy},
		},
		[]ewf.AfterWorkflowHook{closeClientHook()},
		nil,
		nil,
		nil,
	)
}

func buildRollbackFailedDeploymentTemplate(templateBuilder workflow.TemplateBuilderFunc) ewf.WorkflowTemplate {
	return templateBuilder(
		[]ewf.Step{
			{Name: workflow.StepRemoveCluster, RetryPolicy: standardRetryPolicy},
		},
		[]ewf.AfterWorkflowHook{closeClientHook()},
		nil,
		nil,
		nil,
	)
}

func buildRollbackAddNodeTemplate(templateBuilder workflow.TemplateBuilderFunc) ewf.WorkflowTemplate {
	return templateBuilder(
		[]ewf.Step{
			{Name: workflow.StepRemoveNode, RetryPolicy: standardRetryPolicy},
		},
		[]ewf.AfterWorkflowHook{closeClientHook()},
		nil,
		nil,
		nil,
	)
}
