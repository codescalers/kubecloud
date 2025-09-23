package deployment

import (
	"context"
	"kubecloud/internal/logger"
	"kubecloud/internal/metrics"
	"kubecloud/internal/statemanager"
	"kubecloud/internal/workflow"
	"kubecloud/kubedeployer"
	"time"

	"github.com/xmonader/ewf"
)

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
		if err == nil || wf.Name != workflow.WorkflowAddNode {
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
