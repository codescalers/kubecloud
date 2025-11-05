package activities

import (
	"context"
	"errors"
	"fmt"
	"time"

	"kubecloud/internal/constants"
	"kubecloud/internal/logger"
	"kubecloud/internal/metrics"
	"kubecloud/internal/statemanager"
	"kubecloud/kubedeployer"

	"kubecloud/internal/notification"
	"kubecloud/models"

	"github.com/xmonader/ewf"
)

const (
	TimestampFormat = "Mon, 02 Jan 2006 15:04"
)

func hookWorkflowStarted(n *notification.NotificationService) ewf.BeforeWorkflowHook {
	return func(ctx context.Context, w *ewf.Workflow) {
		var userID int
		cfg, err := getConfig(w.State)
		if err == nil {
			userID = cfg.UserID
			logger.GetLogger().Debug().Int("user_id", userID).Msg("hookWorkflowStarted")
		}
		if err != nil {
			logger.GetLogger().Error().Err(err).Msg("missing or invalid config in workflow state")
			userIDVal, ok := w.State["user_id"].(int)
			if !ok {
				logger.GetLogger().Error().Msg("missing or invalid 'user_id' in workflow state")
				return
			}
			userID = userIDVal
			logger.GetLogger().Debug().Int("user_id", userID).Msg("hookWorkflowStarted")
		}

		workflowDesc := getWorkflowDescription(w.Name)
		subject := fmt.Sprintf("%s Started", workflowDesc)
		message := fmt.Sprintf("%s has been started", workflowDesc)

		payload := notification.MergePayload(notification.CommonPayload{
			Subject: subject,
			Message: message,
			Status:  "started",
		}, map[string]string{
			"workflow_name": workflowDesc,
		})

		notificationType := workflowToNotificationType(w.Name)
		notification := models.NewNotification(userID, notificationType, payload, models.WithNoPersist())
		err = n.Send(ctx, notification)
		if err != nil {
			logger.GetLogger().Error().Err(err).Msg("Failed to send notification")
		}
		logger.GetLogger().Info().Str("workflow_name", w.Name).Msg("Starting workflow")
	}
}

func hookStepStarted(ctx context.Context, w *ewf.Workflow, step *ewf.Step) {
	logEvent := logger.GetLogger().Info().
		Str("workflow_name", w.Name).
		Str("step_name", step.Name)

	// Add node_id if available
	if node, err := getFromState[kubedeployer.Node](w.State, "node"); err == nil && node.NodeID > 0 {
		logEvent = logEvent.Uint32("node_id", node.NodeID)
	}

	logEvent.Msg("Starting step")
}

func hookWorkflowDone(_ context.Context, wf *ewf.Workflow, err error) {
	if err != nil {
		logger.GetLogger().Error().Err(err).Str("workflow_name", wf.Name).Msg("Workflow failed")
	} else {
		logger.GetLogger().Info().Str("workflow_name", wf.Name).Msg("Workflow completed successfully")
	}
}

func hookStepDone(_ context.Context, w *ewf.Workflow, step *ewf.Step, err error) {
	if err != nil {
		logEvent := logger.GetLogger().Error().
			Err(err).
			Str("workflow_name", w.Name).
			Str("step_name", step.Name)

		// Add node_id if available
		if node, nodeErr := getFromState[kubedeployer.Node](w.State, "node"); nodeErr == nil && node.NodeID > 0 {
			logEvent = logEvent.Uint32("node_id", node.NodeID)
		}

		logEvent.Msg("Step failed")
	} else {
		logEvent := logger.GetLogger().Info().
			Str("workflow_name", w.Name).
			Str("step_name", step.Name)

		// Add node_id if available
		if node, nodeErr := getFromState[kubedeployer.Node](w.State, "node"); nodeErr == nil && node.NodeID > 0 {
			logEvent = logEvent.Uint32("node_id", node.NodeID)
		}

		logEvent.Msg("Step completed successfully")
	}
}
func hookClusterHealthCheck(notificationService *notification.NotificationService) ewf.AfterWorkflowHook {
	return func(ctx context.Context, wf *ewf.Workflow, err error) {
		if err == nil {
			return
		}
		if errors.Is(err, ewf.ErrFailWorkflowNow) {
			logger.GetLogger().Warn().Str("workflow_name", wf.Name).Msg("Cluster not found in database")
			return
		}
		config, cfgErr := getConfig(wf.State)
		if cfgErr != nil {
			logger.GetLogger().Error().Err(cfgErr).Str("workflow_name", wf.Name).Msg("Failed to get config from state")
			return
		}
		severity := models.NotificationSeverityError
		payload := notification.MergePayload(notification.CommonPayload{
			Message: "Cluster health check failed",
			Subject: "Cluster health check failed",
			Status:  "failed",
			Error:   err.Error(),
		}, map[string]string{
			"workflow_name": getWorkflowDescription(wf.Name),
			"timestamp":     time.Now().UTC().Format(TimestampFormat),
		})
		cluster, errCluster := statemanager.GetCluster(wf.State)
		if errCluster != nil {
			logger.GetLogger().Error().Err(errCluster).Str("workflow_name", wf.Name).Msg("Failed to get cluster from state")

			notification := models.NewNotification(config.UserID, models.NotificationTypeDeployment, payload, models.WithSeverity(severity), models.WithChannels(notification.ChannelEmail))
			if err := notificationService.Send(ctx, notification); err != nil {
				logger.GetLogger().Error().Err(err).Msg("Failed to send cluster health check notification")
			}

			return
		}
		payload["message"] = fmt.Sprintf("Cluster health check failed for cluster Name: %s, Number of nodes: %d", cluster.Name, len(cluster.Nodes))
		payload["cluster_name"] = cluster.Name
		notificationObj := models.NewNotification(config.UserID, models.NotificationTypeDeployment, payload, models.WithSeverity(severity), models.WithChannels(notification.ChannelEmail))
		if err := notificationService.Send(ctx, notificationObj); err != nil {
			logger.GetLogger().Error().Err(err).Msg("Failed to send cluster health check notification")
		}

		logger.GetLogger().Error().Err(err).Str("workflow_name", wf.Name).Msg("Cluster health check failed")
	}
}

func newKubecloudWorkflowTemplate(n *notification.NotificationService) ewf.WorkflowTemplate {
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

func addNodeFailureHook(engine *ewf.Engine, metrics *metrics.Metrics) ewf.AfterWorkflowHook {
	return func(ctx context.Context, wf *ewf.Workflow, err error) {
		if err == nil || wf.Name != constants.WorkflowAddNode {
			return
		}

		node, err := getFromState[kubedeployer.Node](wf.State, "node")
		if err != nil {
			logger.GetLogger().Error().Err(err).Str("workflow_name", wf.Name).Msg("missing or invalid 'node' in workflow state")
			return
		}

		rollbackWf, rollbackErr := engine.NewWorkflow(constants.WorkflowRollbackFailedAddNode)
		if rollbackErr != nil {
			logger.GetLogger().Error().
				Err(rollbackErr).
				Str("node", node.Name).
				Uint32("node_id", node.NodeID).
				Str("workflow_name", wf.Name).
				Msg("Failed to create rollback workflow")
			return
		}

		rollbackWf.State["config"] = wf.State["config"]
		rollbackWf.State["cluster"] = wf.State["cluster"]
		rollbackWf.State["kubeclient"] = wf.State["kubeclient"]
		rollbackWf.State["node_name"] = node.OriginalName

		rollbackCtx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		defer cancel()

		// wait the rollback workflow to finish before closing the client
		if err := engine.RunSync(rollbackCtx, rollbackWf); err != nil {
			logger.GetLogger().Error().
				Err(err).
				Str("node", node.Name).
				Uint32("node_id", node.NodeID).
				Str("workflow_name", wf.Name).
				Msg("Failed to run rollback workflow")
			return
		}

		metrics.DecActiveClusterCount()
	}
}
