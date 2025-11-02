package activities

import (
	"context"
	"fmt"
	"kubecloud/internal"
	"kubecloud/internal/constants"
	"kubecloud/internal/notification"
	"kubecloud/internal/statemanager"
	"kubecloud/kubedeployer"
	"kubecloud/models"
	"slices"
	"strconv"
	"strings"
	"time"

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
		var notifications []*models.Notification
		notifcationType := workflowToNotificationType(wf.Name)
		switch notifcationType {
		case models.NotificationTypeDeployment:
			if n := CreateDeploymentWorkflowNotification(ctx, wf, err); n != nil {
				notifications = append(notifications, n)
			}
		case models.NotificationTypeBilling:
			if n := CreateBillingWorkflowNotifications(ctx, wf, err); n != nil {
				notifications = append(notifications, n...)
			}
		case models.NotificationTypeNode:
			if n := CreateNodeWorkflowNotification(ctx, wf, err); n != nil {
				notifications = append(notifications, n)
			}
		case models.NotificationTypeUser:
			if n := CreateUserWorkflowNotification(ctx, wf, err); n != nil {
				notifications = append(notifications, n)
			}
		}

		for _, n := range notifications {

			if err := notificationService.Send(ctx, n); err != nil {
				logger.GetLogger().Error().
					Str("notification_id", n.ID).
					Str("notification_type", string(n.Type)).
					Int("user_id", n.UserID).
					Err(err).
					Msg("Failed to send notification")
			}
		}
	}
}

func CreateDeploymentWorkflowNotification(ctx context.Context, wf *ewf.Workflow, err error) *models.Notification {
	config, confErr := getConfig(wf.State)
	if confErr != nil {
		logger.GetLogger().Error().Str("workflow_name", wf.Name).Msg("Missing or invalid 'config' in workflow state")
		return &models.Notification{}
	}

	workflowDesc := getWorkflowDescription(wf.Name)
	var notificationPayload map[string]string

	if err != nil {
		var clusterName string
		var nodeInfo string
		var nodeID uint32
		message := fmt.Sprintf("%s failed", workflowDesc)

		if cluster, clusterErr := statemanager.GetCluster(wf.State); clusterErr == nil {
			clusterName = cluster.Name
			message = fmt.Sprintf("%s for cluster '%s' failed", workflowDesc, cluster.Name)

			// Add node information if available
			if node, nodeErr := getFromState[kubedeployer.Node](wf.State, "node"); nodeErr == nil {
				nodeInfo = node.Name
				nodeID = node.NodeID
				message = fmt.Sprintf("%s for cluster '%s', node '%s' (node_id=%d) failed", workflowDesc, cluster.Name, node.Name, node.NodeID)
			}
		}

		payloadData := map[string]string{
			"workflow_name": workflowDesc,
			"cluster_name":  clusterName,
			"node_name":     nodeInfo,
			"user_id":       fmt.Sprintf("%d", config.UserID),
			"timestamp":     time.Now().Local().Format(TimestampFormat),
		}
		if nodeID > 0 {
			payloadData["node_id"] = fmt.Sprintf("%d", nodeID)
		}

		notificationPayload = notification.MergePayload(notification.CommonPayload{
			Message: message,
			Error:   err.Error(),
			Subject: fmt.Sprintf("%s failed", workflowDesc),
			Status:  "failed",
		}, payloadData)

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
			"workflow_name": getWorkflowDescription(wf.Name),
			"cluster_name":  cluster.Name,
			"user_id":       fmt.Sprintf("%d", config.UserID),
			"timestamp":     time.Now().Local().Format(TimestampFormat),
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
			"workflow_name": workflowDesc,
			"cluster_name":  cluster.Name,
			"node_count":    fmt.Sprintf("%d", nodeCount),
			"total_steps":   fmt.Sprintf("%d", totalSteps),
			"user_id":       fmt.Sprintf("%d", config.UserID),
			"timestamp":     time.Now().Local().Format(TimestampFormat),
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
		logger.GetLogger().Error().Str("workflow_name", workflowName).Str("step_name", stepName).Msg("Missing or invalid 'config' in workflow state for step notification")
		return
	}

	clusterName := ""
	if cluster, clusterErr := statemanager.GetCluster(state); clusterErr == nil {
		clusterName = cluster.Name
	}

	current := calculateCurrentStep(stepName)
	total := 4
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
			"user_id":       fmt.Sprintf("%d", config.UserID),
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
		models.WithSeverity(models.NotificationSeverityInfo),
	)
	err = notificationService.Send(context.Background(), notification)
	if err != nil {
		logger.GetLogger().Error().Err(err).Str("workflow_name", workflowName).Str("step_name", stepName).Msg("Failed to send notification")
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

	if stepName == constants.StepDeployLeaderNode {
		return 2
	}

	if stepName == constants.StepBatchDeployAllNodes {
		return 3
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
	return name == constants.WorkflowDeployCluster
}

func isDeployStep(stepName string) bool {
	return stepName == constants.StepDeployLeaderNode ||
		stepName == constants.StepBatchDeployAllNodes
}

func CreateBillingWorkflowNotifications(ctx context.Context, wf *ewf.Workflow, err error) []*models.Notification {
	userID, ok := wf.State["user_id"].(int)
	if !ok {
		logger.GetLogger().Error().Str("workflow_name", wf.Name).Msg("Missing or invalid 'user_id' in workflow state")
		return nil
	}

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

	if wf.Name == constants.WorkflowAdminCreditBalance {
		adminID, ok := wf.State["admin_id"].(int)
		if !ok {
			logger.GetLogger().Error().Str("workflow_name", wf.Name).Msg("Missing or invalid 'admin_id' in workflow state")
			return nil
		}
		payloadData := notification.CommonPayload{
			Message: fmt.Sprintf("User %d was credited successfully money transferred successfully to their account (Amount: $%.2f)", userID, amountUSD),
			Subject: "Money transfer to user's account succeeded",
			Status:  "succeeded",
		}
		severity := models.NotificationSeveritySuccess
		if err != nil {
			severity = models.NotificationSeverityError
			payloadData.Message = fmt.Sprintf("Money transfer to user %d's account failed", userID)
			payloadData.Subject = "Money transfer to user's account failed"
			payloadData.Error = err.Error()
			adminNotif := models.NewNotification(adminID, models.NotificationTypeBilling, notification.MergePayload(payloadData, map[string]string{
				"workflow_name": getWorkflowDescription(wf.Name),
				"user_id":       fmt.Sprintf("%d", userID),
				"admin_id":      fmt.Sprintf("%d", adminID),
				"timestamp":     time.Now().Local().Format(TimestampFormat),
			}), models.WithSeverity(severity), models.WithChannels(notification.ChannelUI))
			return []*models.Notification{adminNotif}
		}
		adminNotif := models.NewNotification(adminID, models.NotificationTypeBilling, notification.MergePayload(payloadData, map[string]string{
			"workflow_name": getWorkflowDescription(wf.Name),
			"user_id":       fmt.Sprintf("%d", userID),
			"admin_id":      fmt.Sprintf("%d", adminID),
			"amount":        fmt.Sprintf("%.2f", amountUSD),
			"timestamp":     time.Now().Local().Format(TimestampFormat),
		}), models.WithSeverity(severity), models.WithChannels(notification.ChannelUI))
		// also notify the user about success
		userPayload := notification.MergePayload(notification.CommonPayload{
			Message: fmt.Sprintf("Funds were credited to your account. Amount added: $%.2f.", amountUSD),
			Subject: "Your Account Has Been Credited",
			Status:  "succeeded",
		}, map[string]string{
			"workflow_name": getWorkflowDescription(wf.Name),
			"user_id":       fmt.Sprintf("%d", userID),
			"timestamp":     time.Now().Local().Format(TimestampFormat),
			"amount":        fmt.Sprintf("%.2f", amountUSD),
		})
		userNotif := models.NewNotification(userID, models.NotificationTypeBilling, userPayload, models.WithSeverity(severity), models.WithChannels(notification.ChannelEmail))
		return []*models.Notification{adminNotif, userNotif}
	}

	var payload map[string]string

	// Extract amount and balance from workflow state
	if amountVal, ok := wf.State["amount"]; ok {
		if amount, okAmount := amountVal.(uint64); okAmount {
			amountUSD = internal.FromUSDMilliCentToUSD(amount)
		}
	}

	if balanceVal, exists := wf.State["net_balance"]; exists {
		if balance, okBalance := balanceVal.(uint64); okBalance {
			newBalanceUSD = internal.FromUSDMilliCentToUSD(balance)
		}
	}

	var status, subject, message string

	if err == nil {
		status = "funds_succeeded"
		subject = "Adding Funds Succeeded"
		message = fmt.Sprintf("Funds were added successfully to your account. Amount added: $%.2f. New balance will be: $%.2f.", amountUSD, newBalanceUSD)

		if wf.Name == constants.WorkflowRedeemVoucher {
			status = "voucher_redeemed"
			subject = "Voucher Redeemed"
			message = fmt.Sprintf("Voucher redeemed successfully. Amount added: $%.2f.", amountUSD)
		}
	} else {
		status = "funds_failed"
		subject = "Adding Funds Failed"
		message = fmt.Sprintf("Failed to add funds to your account: %s", err.Error())
	}
	payloadData := map[string]string{
		"workflow_name": getWorkflowDescription(wf.Name),
		"user_id":       fmt.Sprintf("%d", userID),
		"timestamp":     time.Now().Local().Format(TimestampFormat),
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

	return []*models.Notification{models.NewNotification(
		userID,
		models.NotificationTypeBilling,
		payload,
	)}
}

func CreateNodeWorkflowNotification(ctx context.Context, wf *ewf.Workflow, err error) *models.Notification {
	userID, ok := wf.State["user_id"].(int)
	if !ok {
		logger.GetLogger().Error().Str("workflow_name", wf.Name).Msg("Missing or invalid 'user_id' in workflow state")
		return nil
	}

	var payload map[string]string

	// Extract node information from workflow state
	var nodeID uint32
	var contractID uint64

	if nodeIDVal, ok := wf.State["node_id"]; ok {
		if id, okID := nodeIDVal.(uint32); okID {
			nodeID = id
		}
	}
	if contractIDVal, ok := wf.State["contract_id"]; ok {
		if id, okID := contractIDVal.(uint64); okID {
			contractID = id
		}
	}

	var subject, message string
	//default workflow reserve node
	subject = "Node Reserved Successfully"
	message = fmt.Sprintf("Node %d has been reserved successfully (contract_id=%d)", nodeID, contractID)
	severity := models.NotificationSeveritySuccess
	if err == nil {
		if wf.Name == constants.WorkflowUnreserveNode {
			subject = "Node Unreserved Successfully"
			message = fmt.Sprintf("Node %d has been unreserved successfully (contract_id=%d)", nodeID, contractID)
		}
	} else {
		severity = models.NotificationSeverityError
		subject = "Node Reservation Failed"
		message = fmt.Sprintf("Failed to reserve node %d: %s", nodeID, err.Error())
		if wf.Name == constants.WorkflowUnreserveNode {
			subject = "Node Unreservation Failed"
			message = fmt.Sprintf("Failed to unreserve node %d: %s", nodeID, err.Error())

			if strings.Contains(err.Error(), constants.NodeHasActiveContracts) {
				message = fmt.Sprintf("Failed to unreserve node %d (contract_id=%d). This node has active workloads on it, please remove all deployments from it first", nodeID, contractID)
			}
		}
	}

	// Build payload data
	payloadData := map[string]string{
		"workflow_name": getWorkflowDescription(wf.Name),
		"user_id":       fmt.Sprintf("%d", userID),
		"timestamp":     time.Now().Local().Format(TimestampFormat),
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
		logger.GetLogger().Error().Str("workflow_name", wf.Name).Msg("Missing or invalid 'user_id' in workflow state")
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
		message = fmt.Sprintf("Account verification process failed: %s", err.Error())
		if wf.Name == constants.WorkflowUserRegistration {
			subject = "User Registration Failed"
			message = fmt.Sprintf("User registration process failed: %s", err.Error())
		}
	}

	payloadData := map[string]string{
		"workflow_name": getWorkflowDescription(wf.Name),
		"user_id":       fmt.Sprintf("%d", userID),
		"timestamp":     time.Now().Local().Format(TimestampFormat),
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
