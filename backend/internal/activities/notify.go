package activities

import (
	"context"
	"fmt"
	"kubecloud/internal"
	"kubecloud/internal/statemanager"
	"strconv"
	"strings"

	"github.com/xmonader/ewf"
	"kubecloud/internal/logger"
)

func notifyWorkflowProgress(sse *internal.SSEManager) ewf.AfterWorkflowHook {
	return func(ctx context.Context, wf *ewf.Workflow, err error) {
		config, confErr := getConfig(wf.State)
		if confErr != nil {
			logger.GetLogger().Error().Msg("Missing or invalid 'config' in workflow state")
			return
		}

		workflowDesc := getWorkflowDescription(wf.Name)
		var notificationData map[string]interface{}

		if err != nil {
			message := fmt.Sprintf("%s failed", workflowDesc)
			if cluster, clusterErr := statemanager.GetCluster(wf.State); clusterErr == nil {
				message = fmt.Sprintf("%s for cluster '%s' failed", workflowDesc, cluster.Name)
			}

			notificationData = map[string]interface{}{
				"type":    "workflow_update",
				"message": message,
				"data":    map[string]interface{}{"name": wf.Name, "error": err.Error()},
			}
		} else {
			cluster, clusterErr := statemanager.GetCluster(wf.State)
			if clusterErr != nil {
				notificationData = map[string]interface{}{
					"type":    "workflow_update",
					"message": fmt.Sprintf("%s completed successfully", workflowDesc),
					"data":    map[string]interface{}{"name": wf.Name, "error": false},
				}
			} else {
				nodeCount := len(cluster.Nodes)
				totalSteps := nodeCount + 2
				message := fmt.Sprintf("%s completed successfully for cluster '%s' with %d nodes",
					workflowDesc, cluster.Name, nodeCount)

				notificationData = map[string]interface{}{
					"type":    "workflow_update",
					"message": message,
					"data": map[string]interface{}{
						"name":         wf.Name,
						"cluster_name": cluster.Name,
						"node_count":   nodeCount,
						"total_steps":  totalSteps,
						"cluster":      cluster,
						"error":        false,
					},
				}
			}
		}

		sse.Notify(config.UserID, "workflow_update", notificationData)
	}
}

// notifyStepProgress sends step progress notifications
func notifyStepProgress(sse *internal.SSEManager, state ewf.State, workflowName, stepName string, status string, err error, retryCount, maxRetries int) {
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
	switch {
	case err != nil:
		notificationType = "step_failed"
		message = fmt.Sprintf("Step failed%s", progressStr)
	case status == "completed":
		notificationType = "step_completed"
		message = fmt.Sprintf("Step completed%s", progressStr)
	case status == "retrying":
		notificationType = "step_retrying"
		retryStr := fmt.Sprintf(" - retry %d/%d", retryCount, maxRetries)
		message = fmt.Sprintf("Retrying Step %s%s", progressStr, retryStr)
	default:
		return
	}

	notificationData := map[string]interface{}{
		"type":    notificationType,
		"message": fmt.Sprintf("Deploying cluster %q - %s", clusterName, message),
		"data": map[string]interface{}{
			"workflow_name": workflowName,
			"step_name":     stepName,
			"cluster_name":  clusterName,
			"progress": map[string]interface{}{
				"current": current,
				"total":   total,
			},
		},
	}

	if err != nil {
		notificationData["data"].(map[string]interface{})["error"] = err.Error()
	}

	sse.Notify(config.UserID, notificationType, notificationData)
}

func notifyStepHook(sse *internal.SSEManager) ewf.AfterStepHook {
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
				notifyStepProgress(sse, wf.State, wf.Name, step.Name, "retrying", err, attempts, maxAttempts)
				wf.State[attemptKey] = attempts
				return
			}
			notifyStepProgress(sse, wf.State, wf.Name, step.Name, "failed", err, 0, 0)
		} else {
			notifyStepProgress(sse, wf.State, wf.Name, step.Name, "completed", nil, 0, 0)
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
