package activities

import (
	"context"
	"fmt"
	"kubecloud/internal"
	"kubecloud/internal/constants"
	"kubecloud/internal/notification"
	"kubecloud/internal/statemanager"
	"kubecloud/models"
	"slices"
	"strconv"
	"strings"

	"kubecloud/internal/logger"

	"github.com/xmonader/ewf"
)

func workflowToNotificationType(workflowName string) models.NotificationType {
	billingWf := []string{constants.WorkflowChargeBalance, constants.WorkflowAdminCreditBalance, constants.WorkflowRedeemVoucher}
	deployWf := []string{constants.WorkflowDeleteAllClusters, constants.WorkflowDeleteCluster, constants.WorkflowRemoveNode, constants.WorkflowAddNode, constants.WorkflowRollbackFailedDeployment}
	nodesWf := []string{constants.WorkflowReserveNode, constants.WorkflowUnreserveNode}
	userWf := []string{constants.WorkflowUserVerification, constants.WorkflowUserRegistration}

	switch {
	case slices.Contains(billingWf, workflowName):
		return models.NotificationTypeBilling
	case slices.Contains(deployWf, workflowName):
		return models.NotificationTypeDeployment
	case slices.Contains(nodesWf, workflowName):
		return models.NotificationTypeNode
	case slices.Contains(userWf, workflowName):
		return models.NotificationTypeUser
	default:
		return models.NotificationTypeDeployment
	}
}

func notifyWorkflowProgress(notificationService *notification.NotificationService) ewf.AfterWorkflowHook {
	return func(ctx context.Context, wf *ewf.Workflow, err error) {
		var notifcation *models.Notification
		notifcationType := workflowToNotificationType(wf.Name)
		if notifcationType == models.NotificationTypeDeployment {
			notifcation = CreateDeploymentWorkflowNotification(ctx, wf, err)
		}
		if notifcationType == models.NotificationTypeBilling {
			notifcation = CreateBillingWorkflowNotification(ctx, wf, err)
		}
		if notifcationType == models.NotificationTypeNode {
			notifcation = CreateNodeWorkflowNotification(ctx, wf, err)
		}
		if notifcationType == models.NotificationTypeUser {
			notifcation = CreateUserWorkflowNotification(ctx, wf, err)
		}
		if notifcation != nil {
			err = notificationService.Send(ctx, notifcation)
			if err != nil {
				logger.GetLogger().Error().Err(err).Msg("Failed to send notification")
			}
		}
	}
}

func CreateDeploymentWorkflowNotification(ctx context.Context, wf *ewf.Workflow, err error) *models.Notification {
	config, confErr := getConfig(wf.State)
	if confErr != nil {
		logger.GetLogger().Error().Msg("Missing or invalid 'config' in workflow state")
		return &models.Notification{}
	}

	workflowDesc := getWorkflowDescription(wf.Name)
	var notificationPayload map[string]string

	if err != nil {
		message := fmt.Sprintf("%s failed", workflowDesc)
		if cluster, clusterErr := statemanager.GetCluster(wf.State); clusterErr == nil {
			message = fmt.Sprintf("%s for cluster '%s' failed", workflowDesc, cluster.Name)
		}

		notificationPayload = notification.MergePayload(notification.CommonPayload{
			Message: message,
			Error:   err.Error(),
			Subject: fmt.Sprintf("%s failed", workflowDesc),
			Status:  "failed",
		}, map[string]string{
			"name": wf.Name,
		})

		notification := models.NewNotification(config.UserID, models.NotificationTypeDeployment, notificationPayload)
		return notification
	}
	cluster, clusterErr := statemanager.GetCluster(wf.State)
	if clusterErr != nil {
		notificationPayload = notification.MergePayload(notification.CommonPayload{
			Message: fmt.Sprintf("%s completed successfully", workflowDesc),
			Subject: fmt.Sprintf("%s completed successfully", workflowDesc),
			Status:  "succeeded",
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
			Subject: fmt.Sprintf("%s completed successfully ",
				workflowDesc),
			Status: "succeeded",
		}, map[string]string{
			"name":         wf.Name,
			"cluster_name": cluster.Name,
			"node_count":   fmt.Sprintf("%d", nodeCount),
			"total_steps":  fmt.Sprintf("%d", totalSteps),
		})
	}

	notification := models.NewNotification(config.UserID, models.NotificationTypeDeployment, notificationPayload)
	return notification
}

// notifyStepProgress sends step progress notifications
func notifyStepProgress(notificationService *notification.NotificationService, state ewf.State, workflowName, stepName string, status string, err error, retryCount, maxRetries int) {
	if stepName != constants.StepDeployNetwork && !isDeployStep(stepName) {
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

	payload := notification.MergePayload(
		notification.CommonPayload{
			Subject: "Cluster Deployment Progress",
			Status:  status,
			Message: fmt.Sprintf("Deploying cluster %q - %s", clusterName, message),
		},
		map[string]string{
			"workflow_name": workflowName,
			"step_name":     stepName,
			"cluster_name":  clusterName,
			"current_step":  strconv.Itoa(current),
			"total_steps":   strconv.Itoa(total),
		},
	)
	if err != nil {
		payload["error"] = err.Error()
	}

	notification := models.NewNotification(
		config.UserID,
		models.NotificationType(notificationType),
		payload,
		models.WithNoPersist(),
		models.WithChannels(notification.ChannelUI),
		models.WithSeverity(severity),
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
	if stepName == constants.StepDeployNetwork {
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

func CreateBillingWorkflowNotification(ctx context.Context, wf *ewf.Workflow, err error) *models.Notification {
	userID, ok := wf.State["user_id"].(int)
	if !ok {
		logger.GetLogger().Error().Msg("Missing or invalid 'user_id' in workflow state")
		return nil
	}

	var payload map[string]string

	// Extract amount and balance from workflow state
	var amountUSD, newBalanceUSD float64
	if amountVal, ok := wf.State["amount"]; ok {
		if amount, okAmount := amountVal.(uint64); okAmount {
			amountUSD = internal.FromUSDMilliCentToUSD(amount)
		}
	}
	if balanceVal, ok := wf.State["new_balance"]; ok {
		if balance, okBalance := balanceVal.(uint64); okBalance {
			newBalanceUSD = internal.FromUSDMilliCentToUSD(balance)
		}
	}

	status := "funds_failed"
	subject := "Adding Funds Failed"
	message := "Failed to add funds to your account"
	if err == nil {
		status = "funds_succeeded"
		subject = "Adding Funds Succeeded"
		message = fmt.Sprintf("Funds were added successfully to your account. Amount added: $%.2f. New balance: $%.2f.", amountUSD, newBalanceUSD)
		if wf.Name == constants.WorkflowRedeemVoucher {
			status = "voucher_redeemed"
			subject = "Voucher Redeemed"
			message = "Voucher redeemed successfully."
		}
	}
	payloadData := map[string]string{
		"name": wf.Name,
	}
	if amountUSD > 0 {
		payloadData["amount"] = fmt.Sprintf("%.2f", amountUSD)
	}
	if newBalanceUSD > 0 {
		payloadData["balance"] = fmt.Sprintf("%.2f", newBalanceUSD)
	}

	payload = notification.MergePayload(notification.CommonPayload{
		Message: message,
		Subject: subject,
		Status:  status,
	}, payloadData)

	if err != nil {
		payload["error"] = err.Error()
	}

	return models.NewNotification(
		userID,
		models.NotificationTypeBilling,
		payload,
	)
}

func CreateNodeWorkflowNotification(ctx context.Context, wf *ewf.Workflow, err error) *models.Notification {
	userID, ok := wf.State["user_id"].(int)
	if !ok {
		logger.GetLogger().Error().Msg("Missing or invalid 'user_id' in workflow state")
		return nil
	}

	var payload map[string]string

	// Extract node information from workflow state
	var nodeID uint32
	var contractID uint32

	if nodeIDVal, ok := wf.State["node_id"]; ok {
		if id, okID := nodeIDVal.(uint32); okID {
			nodeID = id
		}
	}
	if contractIDVal, ok := wf.State["contract_id"]; ok {
		if id, okID := contractIDVal.(uint32); okID {
			contractID = id
		}
	}

	var subject, message string
	//default workflow reserve node
	subject = "Node Reserved Successfully"
	message = fmt.Sprintf("Node %d has been reserved successfully", nodeID)
	severity := models.NotificationSeveritySuccess
	if err == nil {
		if wf.Name == constants.WorkflowUnreserveNode {
			subject = "Node Unreserved Successfully"
			message = fmt.Sprintf("Node %d has been unreserved successfully", nodeID)
		}
	} else {
		severity = models.NotificationSeverityError
		subject = "Node Reservation Failed"
		message = fmt.Sprintf("Failed to reserve node %d", nodeID)
		if wf.Name == constants.WorkflowUnreserveNode {
			subject = "Node Unreservation Failed"
			message = fmt.Sprintf("Failed to unreserve node %d", nodeID)
		}
	}

	// Build payload data
	payloadData := map[string]string{
		"name": wf.Name,
	}
	if nodeID > 0 {
		payloadData["node_id"] = fmt.Sprintf("%d", nodeID)
	}
	if contractID > 0 {
		payloadData["contract_id"] = fmt.Sprintf("%d", contractID)
	}

	payload = notification.MergePayload(notification.CommonPayload{
		Message: message,
		Subject: subject,
	}, payloadData)

	if err != nil {
		payload["error"] = err.Error()
	}

	return models.NewNotification(
		userID,
		models.NotificationTypeNode,
		payload,
		models.WithChannels(notification.ChannelUI),
		models.WithSeverity(severity),
		models.WithNoPersist(),
	)
}

func CreateUserWorkflowNotification(ctx context.Context, wf *ewf.Workflow, err error) *models.Notification {
	userID, ok := wf.State["user_id"].(int)
	if !ok {
		logger.GetLogger().Error().Msg("Missing or invalid 'user_id' in workflow state")
		return nil
	}

	var payload map[string]string

	var subject, message string
	//default workflow verified

	subject = "Account Verified Successfully"
	message = "Your account has been verified successfully"
	severity := models.NotificationSeveritySuccess
	if err == nil {
		if wf.Name == constants.WorkflowUserRegistration {

			subject = "Registration Completed"
			message = "Your registration has been completed successfully"
		}
	} else {
		severity = models.NotificationSeverityError
		subject = "Account Verification Failed"
		message = "Account verification process failed"
		if wf.Name == constants.WorkflowUserRegistration {
			subject = "User Registration Failed"
			message = "User registration process failed"
		}
	}

	payloadData := map[string]string{
		"name": wf.Name,
	}

	payload = notification.MergePayload(notification.CommonPayload{
		Message: message,
		Subject: subject,
	}, payloadData)

	if err != nil {
		payload["error"] = err.Error()
	}

	return models.NewNotification(
		userID,
		models.NotificationTypeUser,
		payload,
		models.WithNoPersist(),
		models.WithSeverity(severity),
	)
}
