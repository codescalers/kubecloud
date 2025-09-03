package activities

import (
	"context"
	"fmt"
	"kubecloud/internal/notification"
	"kubecloud/internal/statemanager"
	"kubecloud/models"
	"strconv"
	"strings"
	"time"

	"kubecloud/internal/logger"

	"github.com/xmonader/ewf"
)

const (
	timeFormat = "Mon Jan 2, 2006 at 3:04pm (MST)"
)

func notifyWorkflowProgress(notificationService *notification.NotificationService) ewf.AfterWorkflowHook {
	return func(ctx context.Context, wf *ewf.Workflow, err error) {
		config, confErr := getConfig(wf.State)
		if confErr != nil {
			logger.GetLogger().Error().Msg("Missing or invalid 'config' in workflow state")
			return
		}

		workflowDesc := getWorkflowDescription(wf.Name)
		var notificationPayload map[string]string

		if err != nil {
			message := fmt.Sprintf("%s failed", workflowDesc)
			if cluster, clusterErr := statemanager.GetCluster(wf.State); clusterErr == nil {
				message = fmt.Sprintf("%s for cluster '%s' failed", workflowDesc, cluster.Name)
			}

			notificationPayload = map[string]string{
				"message": message,
				"name":    wf.Name,
				"error":   err.Error(),
			}

			notification := models.NewNotification(config.UserID, "workflow_update", notificationPayload, models.WithNoPersist(), models.WithChannels(notification.ChannelUI), models.WithSeverity(models.NotificationSeverityError))
			err = notificationService.Send(ctx, notification)
			if err != nil {
				logger.GetLogger().Error().Err(err).Msg("Failed to send workflow update notification")
			}
			return
		}
		cluster, clusterErr := statemanager.GetCluster(wf.State)
		if clusterErr != nil {
			notificationPayload = map[string]string{
				"message": fmt.Sprintf("%s completed successfully", workflowDesc),
				"name":    wf.Name,
			}
		} else {
			nodeCount := len(cluster.Nodes)
			message := fmt.Sprintf("%s completed successfully for cluster '%s' with %d nodes",
				workflowDesc, cluster.Name, nodeCount)

			notificationPayload = map[string]string{
				"message":      message,
				"name":         wf.Name,
				"cluster_name": cluster.Name,
				"node_count":   fmt.Sprintf("%d", nodeCount),
			}
			return
		}
		cluster, clusterErr := statemanager.GetCluster(wf.State)
		if clusterErr != nil {
			notificationPayload = notification.MergePayload(notification.CommonPayload{
				Message: fmt.Sprintf("%s completed successfully", workflowDesc),
			}, map[string]string{
				"name": wf.Name,
			})

		} else {
			nodeCount := len(cluster.Nodes)
			totalSteps := nodeCount + 2
			message := fmt.Sprintf("%s completed successfully for cluster '%s' with %d nodes",
				workflowDesc, cluster.Name, nodeCount)

			notificationPayload = notification.MergePayload(notification.CommonPayload{
				Message: message,
			}, map[string]string{
				"name":         wf.Name,
				"cluster_name": cluster.Name,
				"node_count":   fmt.Sprintf("%d", nodeCount),
				"total_steps":  fmt.Sprintf("%d", totalSteps),
			})
		}

		notification := models.NewNotification(config.UserID, "workflow_update", notificationPayload, models.WithNoPersist(), models.WithChannels(notification.ChannelUI), models.WithSeverity(models.NotificationSeveritySuccess))
		err = notificationService.Send(ctx, notification)
		if err != nil {
			logger.GetLogger().Error().Err(err).Msg("Failed to send workflow update notification")
		}
	}
}

// notifyStepProgress sends step progress notifications
func notifyStepProgress(notificationService *notification.NotificationService, state ewf.State, workflowName, stepName string, status string, err error, retryCount, maxRetries int) {
	if stepName != StepDeployNetwork && !isDeployStep(stepName) {
		return
	}

	config, confErr := getConfig(state)
	if confErr != nil {
		logger.GetLogger().Error().Msg("Missing or invalid 'config' in workflow state for step notification")
		return
	}

	clusterName := ""
	nodesNum := 0
	if cluster, clusterErr := statemanager.GetCluster(state); clusterErr == nil {
		clusterName = cluster.Name
		nodesNum = len(cluster.Nodes)
	}

	current := calculateCurrentStep(stepName)
	total := nodesNum + 2
	progressStr := fmt.Sprintf(" (%d/%d)", current, total)

	var message string
	var notificationType string
	var severity models.NotificationSeverity
	switch {
	case err != nil:
		notificationType = "step_failed"
		message = fmt.Sprintf("Step failed%s", progressStr)
		severity = models.NotificationSeverityError
	case status == "completed":
		notificationType = "step_completed"
		message = fmt.Sprintf("Step completed%s", progressStr)
		severity = models.NotificationSeveritySuccess
	case status == "retrying":
		notificationType = "step_retrying"
		retryStr := fmt.Sprintf(" - retry %d/%d", retryCount, maxRetries)
		message = fmt.Sprintf("Retrying Step %s%s", progressStr, retryStr)
		severity = models.NotificationSeverityWarning
	default:
		return
	}

	payload := map[string]string{
		"message":       fmt.Sprintf("Deploying cluster %q - %s", clusterName, message),
		"workflow_name": workflowName,
		"step_name":     stepName,
		"cluster_name":  clusterName,
		"current_step":  strconv.Itoa(current),
		"total_steps":   strconv.Itoa(total),
	}
	if err != nil {
		payload["error"] = err.Error()
	}

	notification := models.NewNotification(
		config.UserID,
		models.NotificationType(notificationType),
		payload,
		models.WithNoPersist(),
		models.WithChannels(notification.ChannelUI),
	)
	err = notificationService.Send(context.Background(), notification)
	if err != nil {
		logger.GetLogger().Error().Err(err).Msg("Failed to send notification")
	}
}

func notifyStepHook(notificationService *notification.NotificationService) ewf.AfterStepHook {
	return func(ctx context.Context, wf *ewf.Workflow, step *ewf.Step, err error) {
		attemptKey := fmt.Sprintf("step_%s_attempts", step.Name)
		attempts, _ := wf.State[attemptKey].(int)

		maxAttempts := 1
		if step.RetryPolicy != nil {
			maxAttempts = int(step.RetryPolicy.MaxAttempts)
		}

		if err != nil {
			if attempts < maxAttempts {
				attempts++
				notifyStepProgress(notificationService, wf.State, wf.Name, step.Name, "retrying", err, attempts, maxAttempts)
				wf.State[attemptKey] = attempts
				return
			}
			notifyStepProgress(notificationService, wf.State, wf.Name, step.Name, "failed", err, 0, 0)
		} else {
			notifyStepProgress(notificationService, wf.State, wf.Name, step.Name, "completed", nil, 0, 0)
		}
	}
}

func calculateCurrentStep(stepName string) int {
	if stepName == StepDeployNetwork {
		return 1
	}

	if isDeployStep(stepName) {
		nodeNumStr := strings.TrimPrefix(stepName, "deploy-")
		nodeNumStr = strings.TrimSuffix(nodeNumStr, "-node")

		nodeNumStr = strings.TrimSuffix(nodeNumStr, "st")
		nodeNumStr = strings.TrimSuffix(nodeNumStr, "nd")
		nodeNumStr = strings.TrimSuffix(nodeNumStr, "rd")
		nodeNumStr = strings.TrimSuffix(nodeNumStr, "th")

		if nodeNum, err := strconv.Atoi(nodeNumStr); err == nil {
			return nodeNum + 1 // +1 because network is step 1
		}
	}

	return 0
}

// getWorkflowDescription returns a user-friendly description for the workflow
func getWorkflowDescription(workflowName string) string {
	if desc, exists := workflowsDescriptions[workflowName]; exists {
		return desc
	}

	// Handle deploy-X-nodes workflows
	if isDeployWorkflow(workflowName) {
		return "Deploying Cluster"
	}

	// Fallback to workflow name
	return workflowName
}

func isDeployWorkflow(name string) bool {
	return strings.HasPrefix(name, "deploy-") && strings.HasSuffix(name, "-nodes")
}

func isDeployStep(stepName string) bool {
	return strings.HasPrefix(stepName, "deploy-") && strings.HasSuffix(stepName, "-node")
}

func NotifyCreateDeploymentResult(notificationService *notification.NotificationService) ewf.AfterWorkflowHook {
	return func(ctx context.Context, wf *ewf.Workflow, err error) {
		config, confErr := getConfig(wf.State)
		if confErr != nil {
			log.Error().Msg("Missing or invalid 'config' in workflow state")
			return
		}

		cluster, clusterErr := statemanager.GetCluster(wf.State)
		workflowDesc := getWorkflowDescription(wf.Name)

		var notificationPayload map[string]string
		if err != nil {
			message := fmt.Sprintf("%s failed", workflowDesc)
			if clusterErr == nil {
				message = fmt.Sprintf("%s for cluster '%s' failed", workflowDesc, cluster.Name)
			}
			notificationPayload = map[string]string{
				"subject": fmt.Sprintf("%s failed", workflowDesc),
				"status":  "failed",
				"message": message,
				"error":   err.Error(),
			}
			notification := models.NewNotification(config.UserID, models.NotificationTypeDeployment, notificationPayload, models.WithChannels(notification.ChannelEmail))
			err := notificationService.Send(ctx, notification)
			if err != nil {
				log.Error().Err(err).Msg("Failed to send deployment failure notification")
			}
			return
		}

		message := fmt.Sprintf("%s completed successfully", workflowDesc)
		if clusterErr == nil {
			nodeCount := len(cluster.Nodes)
			message = fmt.Sprintf("%s completed successfully for cluster '%s' with %d nodes",
				workflowDesc, cluster.Name, nodeCount)
		}
		notificationPayload = map[string]string{
			"subject":      fmt.Sprintf("%s completed successfully", workflowDesc),
			"status":       "succeeded",
			"message":      message,
			"timestamp":    time.Now().Format(time.RFC3339),
			"cluster_name": cluster.Name,
		}

		notification := models.NewNotification(config.UserID, models.NotificationTypeDeployment, notificationPayload, models.WithChannels(notification.ChannelEmail))
		err = notificationService.Send(ctx, notification)
		if err != nil {
			log.Error().Err(err).Msg("Failed to send deployment success notification")
		}
	}
}

func NotifyDeploymentDeleted(notificationService *notification.NotificationService) ewf.AfterWorkflowHook {
	return func(ctx context.Context, wf *ewf.Workflow, err error) {
		config, confErr := getConfig(wf.State)
		if confErr != nil {
			log.Error().Msg("Missing or invalid 'config' in workflow state")
			return
		}

		if err != nil {
			return
		}

		// Try to get cluster and more data for richer notification
		cluster, clusterErr := statemanager.GetCluster(wf.State)
		message := "Deployment deleted successfully"
		clusterName := ""
		nodeCount := 0
		if clusterErr == nil {
			clusterName = cluster.Name
			nodeCount = len(cluster.Nodes)
			message = fmt.Sprintf("Deployment for cluster '%s' with %d nodes was deleted", clusterName, nodeCount)
		}
		notificationPayload := map[string]string{
			"subject":      "Deployment deleted",
			"status":       "deleted",
			"message":      message,
			"timestamp":    time.Now().Format(time.RFC3339),
			"cluster_name": clusterName,
		}

		notification := models.NewNotification(config.UserID, models.NotificationTypeDeployment, notificationPayload)
		err = notificationService.Send(ctx, notification)
		if err != nil {
			log.Error().Err(err).Msg("Failed to send deployment deleted notification")
		}
	}
}
