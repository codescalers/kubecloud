package activities

import (
	"context"
	"fmt"

	"kubecloud/internal/logger"
	"kubecloud/internal/statemanager"

	"kubecloud/internal/notification"
	"kubecloud/models"

	"github.com/xmonader/ewf"
)

func hookWorkflowStarted(n *notification.NotificationService) ewf.BeforeWorkflowHook {
	return func(ctx context.Context, w *ewf.Workflow) {
		var userID int
		cfg, err := getConfig(w.State)
		if err == nil {
			userID = cfg.UserID
			logger.GetLogger().Info().Int("user_id", userID).Msg("hookWorkflowStarted")
		}
		if err != nil {
			logger.GetLogger().Error().Err(err).Msg("missing or invalid config in workflow state")
			userIDVal, ok := w.State["user_id"].(int)
			if !ok {
				logger.GetLogger().Error().Msg("missing or invalid 'user_id' in workflow state")
				return
			}
			userID = userIDVal
			logger.GetLogger().Info().Int("user_id", userID).Msg("hookWorkflowStarted")
		}

		workflowDesc := getWorkflowDescription(w.Name)
		subject := fmt.Sprintf("%s Started", workflowDesc)
		message := fmt.Sprintf("%s has been started", workflowDesc)

		payload := notification.MergePayload(notification.CommonPayload{
			Subject: subject,
			Message: message,
			Status:  "started",
		}, map[string]string{
			"workflow_name": w.Name,
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
	logger.GetLogger().Info().Str("workflow_name", w.Name).Str("step_name", step.Name).Msg("Starting step")
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
		logger.GetLogger().Error().Err(err).Str("workflow_name", w.Name).Str("step_name", step.Name).Msg("Step failed")
	} else {
		logger.GetLogger().Info().Str("workflow_name", w.Name).Str("step_name", step.Name).Msg("Step completed successfully")
	}
}
func hookClusterHealthCheck(notificationService *notification.NotificationService) ewf.AfterWorkflowHook {
	return func(ctx context.Context, wf *ewf.Workflow, err error) {
		if err == nil {
			return
		}
		config, err := getConfig(wf.State)
		if err != nil {
			logger.GetLogger().Error().Err(err).Str("workflow_name", wf.Name).Msg("Failed to get config from state")
			return
		}
		severity := models.NotificationSeverityError
		payload := notification.CommonPayload{
			Message: "Cluster health check failed",
			Subject: "Cluster health check failed",
			Status:  "failed",
		}
		cluster, errCluster := statemanager.GetCluster(wf.State)
		if errCluster != nil {
			logger.GetLogger().Error().Err(err).Str("workflow_name", wf.Name).Msg("Failed to get cluster from state")

			notification := models.NewNotification(config.UserID, models.NotificationTypeDeployment, notification.MergePayload(payload, map[string]string{}), models.WithSeverity(severity), models.WithChannels(notification.ChannelEmail))
			if err := notificationService.Send(ctx, notification); err != nil {
				logger.GetLogger().Error().Err(err).Msg("Failed to send cluster health check notification")
			}

			return
		}
		payload.Message = fmt.Sprintf("Cluster health check failed for cluster Name: %s, Number of nodes: %d", cluster.Name, len(cluster.Nodes))
		notificationObj := models.NewNotification(config.UserID, models.NotificationTypeDeployment, notification.MergePayload(payload, map[string]string{}), models.WithSeverity(severity), models.WithChannels(notification.ChannelEmail))
		if err := notificationService.Send(ctx, notificationObj); err != nil {
			logger.GetLogger().Error().Err(err).Msg("Failed to send cluster health check notification")
		}

		logger.GetLogger().Error().Err(err).Str("workflow_name", wf.Name).Msg("Cluster health check failed")
	}
}

func hookNotificationWorkflowStarted(ctx context.Context, w *ewf.Workflow) {
	logger.GetLogger().Info().Str("workflow_name", w.Name).Msg("Starting notification workflow")
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

func getOrdinalSuffix(n int) string {
	switch n % 100 {
	case 11, 12, 13:
		return "th"
	default:
		switch n % 10 {
		case 1:
			return "st"
		case 2:
			return "nd"
		case 3:
			return "rd"
		default:
			return "th"
		}
	}
}

func getDeployNodeStepName(index int) string {
	return fmt.Sprintf("deploy-%d%s-node", index, getOrdinalSuffix(index))
}
