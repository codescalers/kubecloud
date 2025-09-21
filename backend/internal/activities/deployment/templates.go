package deployment

import (
	"kubecloud/internal/constants"
	"kubecloud/internal/notification"

	"github.com/xmonader/ewf"
)

func buildAddNodeTemplate(ns *notification.NotificationService) ewf.WorkflowTemplate {
	return ewf.WorkflowTemplate{
		Steps: []ewf.Step{
			{Name: constants.StepUpdateNetwork, RetryPolicy: criticalRetryPolicy},
			{Name: constants.StepAddNode, RetryPolicy: standardRetryPolicy},
			{Name: constants.StepFetchKubeconfig, RetryPolicy: criticalRetryPolicy},
			{Name: constants.StepVerifyNewNodes, RetryPolicy: longExponentialRetryPolicy},
			{Name: constants.StepStoreDeployment, RetryPolicy: standardRetryPolicy},
		},
		BeforeWorkflowHooks: []ewf.BeforeWorkflowHook{
			notifyWorkflowStartedHook(ns),
			logWorkflowStartedHook(),
		},
		AfterWorkflowHooks: []ewf.AfterWorkflowHook{
			notifyWorkflowProgressHook(ns),
			addNodeFailureHook(),
			closeClientHook(),
			logWorkflowDoneHook(),
		},
		BeforeStepHooks: []ewf.BeforeStepHook{
			logStepStartedHook(),
		},
		AfterStepHooks: []ewf.AfterStepHook{
			notifyStepProgressHook(ns),
			logStepDoneHook(),
		},
	}
}

func buildRemoveNodeTemplate(ns *notification.NotificationService) ewf.WorkflowTemplate {
	return ewf.WorkflowTemplate{
		Steps: []ewf.Step{
			{Name: constants.StepRemoveNode, RetryPolicy: standardRetryPolicy},
			{Name: constants.StepStoreDeployment, RetryPolicy: standardRetryPolicy},
		},
		BeforeWorkflowHooks: []ewf.BeforeWorkflowHook{
			notifyWorkflowStartedHook(ns),
			logWorkflowStartedHook(),
		},
		AfterWorkflowHooks: []ewf.AfterWorkflowHook{
			notifyWorkflowProgressHook(ns),
			closeClientHook(),
			logWorkflowDoneHook(),
		},
		BeforeStepHooks: []ewf.BeforeStepHook{
			logStepStartedHook(),
		},
		AfterStepHooks: []ewf.AfterStepHook{
			notifyStepProgressHook(ns),
			logStepDoneHook(),
		},
	}
}

func buildDeleteClusterTemplate(ns *notification.NotificationService) ewf.WorkflowTemplate {
	return ewf.WorkflowTemplate{
		Steps: []ewf.Step{
			{Name: constants.StepRemoveCluster, RetryPolicy: standardRetryPolicy},
			{Name: constants.StepRemoveClusterFromDB, RetryPolicy: standardRetryPolicy},
		},
		BeforeWorkflowHooks: []ewf.BeforeWorkflowHook{
			notifyWorkflowStartedHook(ns),
			logWorkflowStartedHook(),
		},
		AfterWorkflowHooks: []ewf.AfterWorkflowHook{
			notifyWorkflowProgressHook(ns),
			closeClientHook(),
			logWorkflowDoneHook(),
		},
		BeforeStepHooks: []ewf.BeforeStepHook{
			logStepStartedHook(),
		},
		AfterStepHooks: []ewf.AfterStepHook{
			notifyStepProgressHook(ns),
			logStepDoneHook(),
		},
	}
}

func buildDeleteAllClustersTemplate(ns *notification.NotificationService) ewf.WorkflowTemplate {
	return ewf.WorkflowTemplate{
		Steps: []ewf.Step{
			{Name: constants.StepGatherAllContractIDs, RetryPolicy: standardRetryPolicy},
			{Name: constants.StepBatchCancelContracts, RetryPolicy: standardRetryPolicy},
			{Name: constants.StepDeleteAllUserClusters, RetryPolicy: standardRetryPolicy},
		},
		BeforeWorkflowHooks: []ewf.BeforeWorkflowHook{
			notifyWorkflowStartedHook(ns),
			logWorkflowStartedHook(),
		},
		AfterWorkflowHooks: []ewf.AfterWorkflowHook{
			notifyWorkflowProgressHook(ns),
			closeClientHook(),
			logWorkflowDoneHook(),
		},
		BeforeStepHooks: []ewf.BeforeStepHook{
			logStepStartedHook(),
		},
		AfterStepHooks: []ewf.AfterStepHook{
			notifyStepProgressHook(ns),
			logStepDoneHook(),
		},
	}
}

func buildRollbackFailedDeploymentTemplate(ns *notification.NotificationService) ewf.WorkflowTemplate {
	return ewf.WorkflowTemplate{
		Steps: []ewf.Step{
			{Name: constants.StepRemoveCluster, RetryPolicy: standardRetryPolicy},
		},
		BeforeWorkflowHooks: []ewf.BeforeWorkflowHook{
			notifyWorkflowStartedHook(ns),
			logWorkflowStartedHook(),
		},
		AfterWorkflowHooks: []ewf.AfterWorkflowHook{
			notifyWorkflowProgressHook(ns),
			closeClientHook(),
			logWorkflowDoneHook(),
		},
		BeforeStepHooks: []ewf.BeforeStepHook{
			logStepStartedHook(),
		},
		AfterStepHooks: []ewf.AfterStepHook{
			notifyStepProgressHook(ns),
			logStepDoneHook(),
		},
	}
}

func buildRollbackAddNodeTemplate(ns *notification.NotificationService) ewf.WorkflowTemplate {
	return ewf.WorkflowTemplate{
		Steps: []ewf.Step{
			{Name: constants.StepRemoveNode, RetryPolicy: standardRetryPolicy},
		},
		BeforeWorkflowHooks: []ewf.BeforeWorkflowHook{
			notifyWorkflowStartedHook(ns),
			logWorkflowStartedHook(),
		},
		AfterWorkflowHooks: []ewf.AfterWorkflowHook{
			notifyWorkflowProgressHook(ns),
			closeClientHook(),
			logWorkflowDoneHook(),
		},
		BeforeStepHooks: []ewf.BeforeStepHook{
			logStepStartedHook(),
		},
		AfterStepHooks: []ewf.AfterStepHook{
			notifyStepProgressHook(ns),
			logStepDoneHook(),
		},
	}
}
