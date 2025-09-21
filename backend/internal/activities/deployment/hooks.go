package deployment

import (
	"context"
	"kubecloud/internal/constants"
	"kubecloud/internal/logger"
	"kubecloud/internal/metrics"
	"kubecloud/internal/notification"
	"kubecloud/internal/statemanager"
	"kubecloud/kubedeployer"
	"time"

	"github.com/xmonader/ewf"
)

// DEPLOYMENT HOOKS

func closeClientHook() ewf.AfterWorkflowHook {
	return func(ctx context.Context, wf *ewf.Workflow, err error) {
		if kubeClient, ok := wf.State["kubeclient"].(*kubedeployer.Client); ok {
			// Save final GridClient state before closing
			statemanager.SaveGridClientState(wf.State, kubeClient)

			kubeClient.Close()
			delete(wf.State, "kubeclient")
		} else {
			logger.GetLogger().Warn().Msg("No kubeclient found in workflow state to close")
		}
	}

}

func addNodeFailureHook() ewf.AfterWorkflowHook {
	return func(ctx context.Context, wf *ewf.Workflow, err error) {
		if err == nil || wf.Name != constants.WorkflowAddNode {
			return
		}

		node, ok := wf.State["node"].(kubedeployer.Node)
		if !ok {
			logger.GetLogger().Error().
				Str("workflow_name", wf.Name).
				Msg("node not found in state for rollback")
			return
		}

		cluster, clusterErr := statemanager.GetCluster(wf.State)
		if clusterErr != nil || cluster.ProjectName == "" {
			logger.GetLogger().Error().
				Err(clusterErr).
				Str("workflow_name", wf.Name).
				Msg("nothing to rollback")
			return
		}

		kubeClient, ok := wf.State["kubeclient"].(*kubedeployer.Client)
		if !ok {
			logger.GetLogger().Error().
				Str("workflow_name", wf.Name).
				Msg("no kubeclient found for rollback")
			return
		}

		logger.GetLogger().Info().
			Str("project_name", cluster.ProjectName).
			Str("node_name", node.Name).
			Msg("Triggering rollback for newly added node")

		if err := kubeClient.RemoveNode(ctx, &cluster, node.Name); err != nil {
			logger.GetLogger().Error().
				Err(err).
				Str("node_name", node.Name).
				Msg("Failed to rollback node")
			return
		}

		statemanager.StoreCluster(wf.State, cluster)
		logger.GetLogger().Info().
			Str("node_name", node.Name).
			Msg("Rollback of new node completed")
	}
}

func deployFailureHook(engine ewf.Engine, metrics *metrics.Metrics) ewf.AfterWorkflowHook {
	return func(ctx context.Context, wf *ewf.Workflow, err error) {
		if err != nil {
			// && isDeployWorkflow(wf.Name)

			cluster, clusterErr := statemanager.GetCluster(wf.State)
			if clusterErr != nil || cluster.ProjectName == "" {
				return
			}

			rollbackWf, rollbackErr := engine.NewWorkflow("rollback-failed-deployment")
			if rollbackErr != nil {
				return
			}

			rollbackWf.State["config"] = wf.State["config"]
			rollbackWf.State["cluster"] = wf.State["cluster"]
			rollbackWf.State["kubeclient"] = wf.State["kubeclient"]
			rollbackWf.State["project_name"] = cluster.ProjectName

			rollbackCtx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
			defer cancel()

			engine.RunSync(rollbackCtx, rollbackWf)
			metrics.DecActiveClusterCount()
		}
	}
}

// NOTIFICATION HOOKS

func notifyWorkflowStartedHook(ns *notification.NotificationService) ewf.BeforeWorkflowHook {
	return func(ctx context.Context, w *ewf.Workflow) {}
	// return activities.NotifyWorkflowStartedHook(ns)
}

func notifyWorkflowProgressHook(ns *notification.NotificationService) ewf.AfterWorkflowHook {
	return func(ctx context.Context, w *ewf.Workflow, err error) {}
	// return activities.NotifyWorkflowProgress(ns)
}

func notifyStepProgressHook(ns *notification.NotificationService) ewf.AfterStepHook {
	return func(ctx context.Context, w *ewf.Workflow, step *ewf.Step, err error) {}
	// return activities.NotifyStepHook(ns)
}

// LOGGING HOOKS

func logWorkflowStartedHook() ewf.BeforeWorkflowHook {
	return func(ctx context.Context, wf *ewf.Workflow) {
		logger.GetLogger().Info().Str("workflow_name", wf.Name).Msg("Starting workflow")
	}
}

func logWorkflowDoneHook() ewf.AfterWorkflowHook {
	return func(ctx context.Context, wf *ewf.Workflow, err error) {
		if err != nil {
			logger.GetLogger().Error().Err(err).Str("workflow_name", wf.Name).Msg("Workflow failed")
		} else {
			logger.GetLogger().Info().Str("workflow_name", wf.Name).Msg("Workflow completed successfully")
		}
	}
}

func logStepStartedHook() ewf.BeforeStepHook {
	return func(ctx context.Context, w *ewf.Workflow, step *ewf.Step) {
		logger.GetLogger().Info().Str("workflow_name", w.Name).Str("step_name", step.Name).Msg("Starting step")
	}
}

func logStepDoneHook() ewf.AfterStepHook {
	return func(ctx context.Context, w *ewf.Workflow, step *ewf.Step, err error) {
		if err != nil {
			logger.GetLogger().Error().Err(err).Str("workflow_name", w.Name).Str("step_name", step.Name).Msg("Step failed")
		} else {
			logger.GetLogger().Info().Str("workflow_name", w.Name).Str("step_name", step.Name).Msg("Step completed successfully")
		}
	}
}

// TODO: delegate the log/notify hooks to the parent package to avoid cyclic imports
