package workflows

import (
	"context"
	"errors"
	"fmt"

	"time"

	"kubecloud/internal/deployment/kubedeployer"
	"kubecloud/internal/deployment/statemanager"
	"kubecloud/internal/infrastructure/logger"
	metricsLib "kubecloud/internal/infrastructure/metrics"

	"kubecloud/internal/infrastructure/notification"

	"github.com/xmonader/ewf"
)

const (
	TimestampFormat = "Mon, 02 Jan 2006 15:04"
)

func hookWorkflowStarted(n *notification.NotificationDispatcher) ewf.BeforeWorkflowHook {
	return func(ctx context.Context, w *ewf.Workflow) {
		log := logger.GetLogger().With().Str("workflow_name", w.Name).Logger()
		suppressNotification, _ := getFromState[bool](w.State, "suppress_notification")
		if suppressNotification {
			log.Info().Msg("Suppressing notification for workflow")
			return
		}

		config, err := getConfig(w.State)
		if err != nil {
			log.Error().Err(err).Msg("failed to get config from state")
			return
		}

		notificationType := workflowToNotificationType(w.Name)
		displayName := getWorkflowDisplayName(w)
		notif := notification.NewNotification(config.UserID, notificationType).
			Info(displayName+" has been started").
			WithSubject(displayName+" Started").
			WithStatus("started").
			WithExtra("workflow_name", displayName).
			NoPersist().
			Build()
		if err = n.Send(ctx, notif); err != nil {
			log.Error().Err(err).Msg("failed to send workflow started notification")
		}
		log.Info().Msg("Starting workflow")
	}
}

func hookStepStarted(ctx context.Context, w *ewf.Workflow, step *ewf.Step) {
	log := logger.GetLogger().With().
		Str("workflow_name", w.Name).
		Str("step_name", step.Name).Logger()

	logEvent := log.Info()

	// Add node_id if available
	if node, err := getFromState[kubedeployer.Node](w.State, "node"); err == nil && node.NodeID > 0 {
		logEvent = logEvent.Uint32("node_id", node.NodeID)
	}

	logEvent.Msg("Starting step")
}

func hookWorkflowDone(_ context.Context, wf *ewf.Workflow, err error) {
	log := logger.GetLogger().With().Str("workflow_name", wf.Name).Logger()
	if err != nil {
		log.Error().Err(err).Msg("workflow failed")
	} else {
		log.Info().Msg("workflow completed successfully")
	}
}

func hookStepDone(_ context.Context, w *ewf.Workflow, step *ewf.Step, err error) {
	log := logger.GetLogger().With().
		Str("workflow_name", w.Name).
		Str("step_name", step.Name).Logger()

	if err != nil {
		logEvent := log.Error().Err(err)

		// Add node_id if available
		if node, nodeErr := getFromState[kubedeployer.Node](w.State, "node"); nodeErr == nil && node.NodeID > 0 {
			logEvent = logEvent.Uint32("node_id", node.NodeID)
		}

		logEvent.Msg("step failed")
	} else {
		logEvent := log.Info()

		// Add node_id if available
		if node, nodeErr := getFromState[kubedeployer.Node](w.State, "node"); nodeErr == nil && node.NodeID > 0 {
			logEvent = logEvent.Uint32("node_id", node.NodeID)
		}

		logEvent.Msg("step completed successfully")
	}
}

func hookClusterHealthCheck(notificationService *notification.NotificationDispatcher) ewf.AfterWorkflowHook {
	return func(ctx context.Context, wf *ewf.Workflow, err error) {
		log := logger.GetLogger().With().Str("workflow_name", wf.Name).Logger()
		if err == nil {
			return
		}
		if !errors.Is(err, ErrClusterNotHealthy) {
			log.Warn().Err(err).Msg("could not check cluster health")
			return
		}

		config, cfgErr := getConfig(wf.State)
		if cfgErr != nil {
			log.Error().Err(cfgErr).Msg("failed to get config from state")
			return
		}

		cluster, errCluster := statemanager.GetCluster(wf.State)
		if errCluster != nil {
			log.Error().Err(errCluster).Msg("failed to get cluster from state")
			return
		}

		message := "Cluster health check failed for cluster " + cluster.Name + " with " + fmt.Sprintf("%d", len(cluster.Nodes)) + " nodes"

		notif := notification.ClusterNotification(config.UserID, cluster.Name).
			Failure(message, err).
			WithSubject("Cluster health check failed").
			WithChannels(notification.ChannelEmail).
			WithExtra("workflow_name", getWorkflowDisplayName(wf)).
			Build()

		if err := notificationService.Send(ctx, notif); err != nil {
			log.Error().Err(err).Msg("failed to send cluster health check notification")
		}

		log.Error().Err(err).Msg("cluster health check failed")
	}
}

func newKubecloudWorkflowTemplate(n *notification.NotificationDispatcher) ewf.WorkflowTemplate {
	return ewf.WorkflowTemplate{
		BeforeWorkflowHooks: []ewf.BeforeWorkflowHook{
			hookWorkflowStarted(n),
		},
		AfterWorkflowHooks: []ewf.AfterWorkflowHook{
			hookWorkflowDone,
			notifyWorkflowProgress(n),
		},
		BeforeStepHooks: []ewf.BeforeStepHook{
			hookStepStarted,
		},
		AfterStepHooks: []ewf.AfterStepHook{
			hookStepDone,
		},
	}
}

func addNodeFailureHook(engine *ewf.Engine, metrics *metricsLib.Metrics) ewf.AfterWorkflowHook {
	return func(ctx context.Context, wf *ewf.Workflow, err error) {
		log := logger.ForOperation("workflow", "hook_add_node_failure").With().Str("workflow_name", wf.Name).Logger()
		if err == nil || wf.Name != WorkflowAddNode {
			return
		}
		metrics.IncrementClusterOperationFailure(metricsLib.ClusterOperationAddNode)
		node, err := getFromState[kubedeployer.Node](wf.State, "node")
		if err != nil {
			log.Error().Err(err).Msg("missing or invalid 'node' in workflow state")
			return
		}

		rollbackWf, rollbackErr := engine.NewWorkflow(WorkflowRollbackFailedAddNode, ewf.WithDisplayName(fmt.Sprintf("Rollback failed node %s", node.Name)))
		if rollbackErr != nil {
			log.Error().
				Err(rollbackErr).
				Str("node", node.Name).
				Uint32("node_id", node.NodeID).
				Msg("failed to create rollback workflow")
			return
		}

		rollbackWf.State["config"] = wf.State["config"]
		rollbackWf.State["cluster"] = wf.State["cluster"]
		rollbackWf.State["kubeclient"] = wf.State["kubeclient"]
		rollbackWf.State["node_name"] = node.OriginalName

		rollbackCtx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		defer cancel()

		// wait the rollback workflow to finish before closing the client
		if err := engine.Run(rollbackCtx, rollbackWf); err != nil {
			log.Error().
				Err(err).
				Str("node", node.Name).
				Uint32("node_id", node.NodeID).
				Msg("failed to run rollback workflow")
			return
		}

	}
}

func metricsSuccessHook(metrics *metricsLib.Metrics) ewf.AfterWorkflowHook {
	return func(ctx context.Context, wf *ewf.Workflow, err error) {
		if err != nil {
			return
		}
		if wf.Name == WorkflowAddNode {
			metrics.IncrementClusterOperationSuccess(metricsLib.ClusterOperationAddNode)
		}
		if wf.Name == WorkflowRemoveNode {
			metrics.IncrementClusterOperationSuccess(metricsLib.ClusterOperationRemoveNode)
		}
		if isDeployWorkflow(wf.Name) {
			metrics.IncActiveClusterCount()
			metrics.IncrementClusterDeploymentSuccess()
		}
	}
}

func metricsFailureHook(metrics *metricsLib.Metrics) ewf.AfterWorkflowHook {
	return func(ctx context.Context, wf *ewf.Workflow, err error) {
		if err == nil {
			return
		}
		if isDeployWorkflow(wf.Name) {
			metrics.IncrementClusterDeploymentFailure()
			return
		}
		switch wf.Name {
		case WorkflowDeleteCluster:
			metrics.IncrementClusterOperationFailure(metricsLib.ClusterOperationDeleteCluster)
		case WorkflowRemoveNode:
			metrics.IncrementClusterOperationFailure(metricsLib.ClusterOperationRemoveNode)
		case WorkflowDeleteAllClusters:
			metrics.IncrementClusterOperationFailure(metricsLib.ClusterOperationDeleteAllClusters)
		}
	}
}
