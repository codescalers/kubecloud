package deployment

import (
	"kubecloud/internal/constants"
	"kubecloud/internal/notification"

	"github.com/xmonader/ewf"
)

func extendBaseTemplate(
	ns *notification.NotificationService,
	steps []ewf.Step,
	afterWorkflow []ewf.AfterWorkflowHook,
	beforeWorkflow []ewf.BeforeWorkflowHook,
	afterStep []ewf.AfterStepHook,
	beforeStep []ewf.BeforeStepHook,
) ewf.WorkflowTemplate {
	// passed steps and hooks should be prioritized
	return ewf.WorkflowTemplate{
		Steps: steps,
		BeforeWorkflowHooks: append(beforeWorkflow,
			notifyWorkflowStartedHook(ns),
			logWorkflowStartedHook(),
		),
		AfterWorkflowHooks: append(afterWorkflow,
			notifyWorkflowProgressHook(ns),
			closeClientHook(),
			logWorkflowDoneHook(),
		),
		BeforeStepHooks: append(beforeStep,
			logStepStartedHook(),
		),
		AfterStepHooks: append(afterStep,
			notifyStepProgressHook(ns),
			logStepDoneHook(),
		),
	}
}

func buildAddNodeTemplate(ns *notification.NotificationService) ewf.WorkflowTemplate {
	steps := []ewf.Step{
		{Name: constants.StepUpdateNetwork, RetryPolicy: criticalRetryPolicy},
		{Name: constants.StepAddNode, RetryPolicy: standardRetryPolicy},
		{Name: constants.StepFetchKubeconfig, RetryPolicy: criticalRetryPolicy},
		{Name: constants.StepVerifyNewNodes, RetryPolicy: longExponentialRetryPolicy},
		{Name: constants.StepStoreDeployment, RetryPolicy: standardRetryPolicy},
	}
	return extendBaseTemplate(ns, steps,
		[]ewf.AfterWorkflowHook{addNodeFailureHook()},
		nil,
		nil,
		nil,
	)
}

func buildRemoveNodeTemplate(ns *notification.NotificationService) ewf.WorkflowTemplate {
	steps := []ewf.Step{
		{Name: constants.StepRemoveNode, RetryPolicy: standardRetryPolicy},
		{Name: constants.StepStoreDeployment, RetryPolicy: standardRetryPolicy},
	}
	return extendBaseTemplate(ns, steps, nil, nil, nil, nil)
}

func buildDeleteClusterTemplate(ns *notification.NotificationService) ewf.WorkflowTemplate {
	steps := []ewf.Step{
		{Name: constants.StepRemoveCluster, RetryPolicy: standardRetryPolicy},
		{Name: constants.StepRemoveClusterFromDB, RetryPolicy: standardRetryPolicy},
	}
	return extendBaseTemplate(ns, steps, nil, nil, nil, nil)
}

func buildDeleteAllClustersTemplate(ns *notification.NotificationService) ewf.WorkflowTemplate {
	steps := []ewf.Step{
		{Name: constants.StepGatherAllContractIDs, RetryPolicy: standardRetryPolicy},
		{Name: constants.StepBatchCancelContracts, RetryPolicy: standardRetryPolicy},
		{Name: constants.StepDeleteAllUserClusters, RetryPolicy: standardRetryPolicy},
	}
	return extendBaseTemplate(ns, steps, nil, nil, nil, nil)
}

func buildRollbackFailedDeploymentTemplate(ns *notification.NotificationService) ewf.WorkflowTemplate {
	steps := []ewf.Step{
		{Name: constants.StepRemoveCluster, RetryPolicy: standardRetryPolicy},
	}
	return extendBaseTemplate(ns, steps, nil, nil, nil, nil)
}

func buildRollbackAddNodeTemplate(ns *notification.NotificationService) ewf.WorkflowTemplate {
	steps := []ewf.Step{
		{Name: constants.StepRemoveNode, RetryPolicy: standardRetryPolicy},
	}
	return extendBaseTemplate(ns, steps, nil, nil, nil, nil)
}
