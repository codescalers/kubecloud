package deployment

import (
	"kubecloud/internal/constants"

	"github.com/xmonader/ewf"
)

func buildAddNodeTemplate() ewf.WorkflowTemplate {
	return ewf.WorkflowTemplate{
		Steps: []ewf.Step{
			{Name: constants.StepUpdateNetwork, RetryPolicy: criticalRetryPolicy},
			{Name: constants.StepAddNode, RetryPolicy: standardRetryPolicy},
			{Name: constants.StepFetchKubeconfig, RetryPolicy: criticalRetryPolicy},
			{Name: constants.StepVerifyNewNodes, RetryPolicy: longExponentialRetryPolicy},
			{Name: constants.StepStoreDeployment, RetryPolicy: standardRetryPolicy},
		},
		BeforeWorkflowHooks: []ewf.BeforeWorkflowHook{},
		AfterWorkflowHooks:  []ewf.AfterWorkflowHook{},
		BeforeStepHooks:     []ewf.BeforeStepHook{},
		AfterStepHooks:      []ewf.AfterStepHook{},
	}
}

func buildRemoveNodeTemplate() ewf.WorkflowTemplate {
	return ewf.WorkflowTemplate{
		Steps: []ewf.Step{
			{Name: constants.StepRemoveNode, RetryPolicy: standardRetryPolicy},
			{Name: constants.StepStoreDeployment, RetryPolicy: standardRetryPolicy},
		},
		BeforeWorkflowHooks: []ewf.BeforeWorkflowHook{},
		AfterWorkflowHooks:  []ewf.AfterWorkflowHook{},
		BeforeStepHooks:     []ewf.BeforeStepHook{},
		AfterStepHooks:      []ewf.AfterStepHook{},
	}
}

func buildDeleteClusterTemplate() ewf.WorkflowTemplate {
	return ewf.WorkflowTemplate{
		Steps: []ewf.Step{
			{Name: constants.StepRemoveCluster, RetryPolicy: standardRetryPolicy},
			{Name: constants.StepRemoveClusterFromDB, RetryPolicy: standardRetryPolicy},
		},
		BeforeWorkflowHooks: []ewf.BeforeWorkflowHook{},
		AfterWorkflowHooks:  []ewf.AfterWorkflowHook{},
		BeforeStepHooks:     []ewf.BeforeStepHook{},
		AfterStepHooks:      []ewf.AfterStepHook{},
	}
}

func buildDeleteAllClustersTemplate() ewf.WorkflowTemplate {
	return ewf.WorkflowTemplate{
		Steps: []ewf.Step{
			{Name: constants.StepGatherAllContractIDs, RetryPolicy: standardRetryPolicy},
			{Name: constants.StepBatchCancelContracts, RetryPolicy: standardRetryPolicy},
			{Name: constants.StepDeleteAllUserClusters, RetryPolicy: standardRetryPolicy},
		},
		BeforeWorkflowHooks: []ewf.BeforeWorkflowHook{},
		AfterWorkflowHooks:  []ewf.AfterWorkflowHook{},
		BeforeStepHooks:     []ewf.BeforeStepHook{},
		AfterStepHooks:      []ewf.AfterStepHook{},
	}
}

func buildRollbackFailedDeploymentTemplate() ewf.WorkflowTemplate {
	return ewf.WorkflowTemplate{
		Steps: []ewf.Step{
			{Name: constants.StepRemoveCluster, RetryPolicy: standardRetryPolicy},
		},
		BeforeWorkflowHooks: []ewf.BeforeWorkflowHook{},
		AfterWorkflowHooks:  []ewf.AfterWorkflowHook{},
		BeforeStepHooks:     []ewf.BeforeStepHook{},
		AfterStepHooks:      []ewf.AfterStepHook{},
	}
}

func buildRollbackAddNodeTemplate() ewf.WorkflowTemplate {
	return ewf.WorkflowTemplate{
		Steps: []ewf.Step{
			{Name: constants.StepRemoveNode, RetryPolicy: standardRetryPolicy},
		},
		BeforeWorkflowHooks: []ewf.BeforeWorkflowHook{},
		AfterWorkflowHooks:  []ewf.AfterWorkflowHook{},
		BeforeStepHooks:     []ewf.BeforeStepHook{},
		AfterStepHooks:      []ewf.AfterStepHook{},
	}
}
