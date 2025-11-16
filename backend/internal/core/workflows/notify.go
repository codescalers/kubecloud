package workflows

import (
	"context"
	"fmt"
	"kubecloud/internal/core/models"
	"kubecloud/internal/deployment/kubedeployer"
	"kubecloud/internal/deployment/statemanager"
	"kubecloud/internal/infrastructure/notification"
	"kubecloud/internal/infrastructure/substrate"

	"slices"

	"kubecloud/internal/infrastructure/logger"

	"github.com/xmonader/ewf"
)

func notifyWorkflowProgress(notificationDispatcher *notification.NotificationDispatcher) ewf.AfterWorkflowHook {
	return func(ctx context.Context, wf *ewf.Workflow, err error) {
		log := logger.ForOperation("workflow", "notify_workflow_progress").With().Str("workflow_name", wf.Name).Logger()

		notificationType := workflowToNotificationType(wf.Name)
		switch notificationType {
		case models.NotificationTypeDeployment:
			if err := sendDeploymentWorkflowNotification(ctx, notificationDispatcher, wf, err); err != nil {
				log.Error().Err(err).Msg("Failed to send deployment workflow notification")
			}
		case models.NotificationTypeBilling:
			if err := sendBillingWorkflowNotifications(ctx, notificationDispatcher, wf, err); err != nil {
				log.Error().Err(err).Msg("Failed to send billing workflow notifications")
			}
		case models.NotificationTypeNode:
			if err := sendNodeWorkflowNotification(ctx, notificationDispatcher, wf, err); err != nil {
				log.Error().Err(err).Msg("Failed to send node workflow notification")
			}
		case models.NotificationTypeUser:
			if err := sendUserWorkflowNotification(ctx, notificationDispatcher, wf, err); err != nil {
				log.Error().Err(err).Msg("Failed to send user workflow notification")
			}
		}
	}
}

func sendDeploymentWorkflowNotification(ctx context.Context, notificationDispatcher *notification.NotificationDispatcher, wf *ewf.Workflow, err error) error {
	log := logger.ForOperation("workflow", "create_deployment_notification").With().Str("workflow_name", wf.Name).Logger()
	config, confErr := getConfig(wf.State)
	if confErr != nil {
		log.Error().Msg("Missing or invalid 'config' in workflow state")
		return confErr
	}

	cluster, clusterErr := statemanager.GetCluster(wf.State)
	if clusterErr != nil {
		log.Error().Err(clusterErr).Msg("failed to get cluster from state")
		return clusterErr
	}

	workflowDesc := getWorkflowDescription(wf.Name)

	if err != nil {
		var nodeInfo string
		var nodeID uint32

		message := fmt.Sprintf("%s for cluster '%s' failed", workflowDesc, cluster.Name)

		// Add node information if available
		if node, nodeErr := getFromState[kubedeployer.Node](wf.State, "node"); nodeErr == nil {
			nodeInfo = node.Name
			nodeID = node.NodeID
			message = fmt.Sprintf("%s for cluster '%s', node '%s' (node_id=%d) failed", workflowDesc, cluster.Name, node.Name, node.NodeID)
		}

		notif := notification.ClusterNotification(config.UserID, cluster.Name).
			Failure(message, err).
			WithSubject(fmt.Sprintf("%s failed", workflowDesc)).
			WithExtra("workflow_name", workflowDesc).
			WithExtra("node_name", nodeInfo).
			WithExtra("node_id", fmt.Sprintf("%d", nodeID)).
			Build()

		return notificationDispatcher.Send(ctx, notif)
	}

	message := fmt.Sprintf("%s completed successfully for cluster '%s' with %d nodes", workflowDesc, cluster.Name, len(cluster.Nodes))
	notif := notification.ClusterNotification(config.UserID, cluster.Name).
		Success(message).
		WithSubject(fmt.Sprintf("%s completed successfully", workflowDesc)).
		WithExtra("workflow_name", workflowDesc).
		WithExtra("node_count", fmt.Sprintf("%d", len(cluster.Nodes))).
		WithExtra("total_steps", fmt.Sprintf("%d", len(cluster.Nodes)+2)).
		Build()

	return notificationDispatcher.Send(ctx, notif)
}

// notifyStepProgress sends step progress notifications
func notifyStepProgress(ctx context.Context, notificationDispatcher *notification.NotificationDispatcher, state ewf.State, workflowName, stepName string, status string, err error, retryCount, maxRetries int) {
	log := logger.ForOperation("workflow", "notify_step_progress").With().
		Str("workflow_name", workflowName).
		Str("step_name", stepName).Logger()

	if stepName != StepDeployNetwork && !isDeployStep(stepName) {
		return
	}

	config, confErr := getConfig(state)
	if confErr != nil {
		log.Error().Msg("Missing or invalid 'config' in workflow state for step notification")
		return
	}

	clusterName := ""
	if cluster, clusterErr := statemanager.GetCluster(state); clusterErr == nil {
		clusterName = cluster.Name
	}

	currentStep := calculateCurrentStep(stepName)
	total := 4
	progressStr := fmt.Sprintf(" (%d/%d)", currentStep, total)

	var builder *notification.NotificationBuilder
	switch {
	case err != nil:
		msg := fmt.Sprintf("Deploying cluster %q - Step failed%s", clusterName, progressStr)
		builder = notification.ClusterNotification(config.UserID, clusterName).Failure(msg, err)
	case status == "completed":
		msg := fmt.Sprintf("Deploying cluster %q - Step completed%s", clusterName, progressStr)
		builder = notification.ClusterNotification(config.UserID, clusterName).Info(msg).WithStatus("completed")
	case status == "retrying":
		retryStr := fmt.Sprintf(" - retry %d/%d", retryCount, maxRetries)
		msg := fmt.Sprintf("Deploying cluster %q - Retrying Step%s%s", clusterName, progressStr, retryStr)
		builder = notification.ClusterNotification(config.UserID, clusterName).Info(msg).WithStatus("retrying")
	default:
		return
	}

	notif := builder.
		WithSubject("Cluster Deployment Progress").
		WithChannels(notification.ChannelUI).
		WithExtra("workflow_name", workflowName).
		WithExtra("step_name", stepName).
		WithExtra("current_step", fmt.Sprintf("%d", currentStep)).
		WithExtra("total_steps", fmt.Sprintf("%d", total)).
		NoPersist().
		Build()

	if err := notificationDispatcher.Send(ctx, notif); err != nil {
		log.Error().Err(err).Msg("Failed to send step notification")
	}
}

func notifyStepHook(notificationDispatcher *notification.NotificationDispatcher) ewf.AfterStepHook {
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
				notifyStepProgress(ctx, notificationDispatcher, wf.State, wf.Name, step.Name, "retrying", err, attempts, maxAttempts)
				wf.State[attemptKey] = attempts
				return
			}
			notifyStepProgress(ctx, notificationDispatcher, wf.State, wf.Name, step.Name, "failed", err, 0, 0)
		} else {
			notifyStepProgress(ctx, notificationDispatcher, wf.State, wf.Name, step.Name, "completed", nil, 0, 0)
		}
	}
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
	return name == WorkflowDeployCluster
}

func isDeployStep(stepName string) bool {
	return stepName == StepDeployLeaderNode ||
		stepName == StepBatchDeployAllNodes
}

func sendBillingWorkflowNotifications(ctx context.Context, notificationDispatcher *notification.NotificationDispatcher, wf *ewf.Workflow, err error) error {
	log := logger.ForOperation("workflow", "create_billing_notification").With().Str("workflow_name", wf.Name).Logger()
	userIDVal, ok := wf.State["user_id"]
	if !ok {
		log.Error().Msg("Missing 'user_id' in workflow state")
		return nil
	}

	var userID int
	switch v := userIDVal.(type) {
	case int:
		userID = v
	case float64:
		userID = int(v)
	case int64:
		userID = int(v)
	default:
		log.Error().Interface("user_id_value", v).Msg("Invalid 'user_id' type in workflow state")
		return nil
	}

	var amountUSD, newBalanceUSD float64
	if amountVal, ok := wf.State["amount"]; ok {
		if amount, okAmount := amountVal.(uint64); okAmount {
			amountUSD = substrate.FromUSDMilliCentToUSD(amount)
		}
	}
	if balanceVal, ok := wf.State["new_balance"]; ok {
		if balance, okBalance := balanceVal.(uint64); okBalance {
			newBalanceUSD = substrate.FromUSDMilliCentToUSD(balance)
		}
	}

	if wf.Name == WorkflowAdminCreditBalance {
		adminIDVal, ok := wf.State["admin_id"]
		if !ok {
			log.Error().Msg("Missing 'admin_id' in workflow state")
			return nil
		}

		var adminID int
		switch v := adminIDVal.(type) {
		case int:
			adminID = v
		case float64:
			adminID = int(v)
		case int64:
			adminID = int(v)
		default:
			log.Error().Interface("admin_id_value", v).Msg("Invalid 'admin_id' type in workflow state")
			return nil
		}
		username, ok := wf.State["username"].(string)
		if !ok {
			log.Warn().Msg("Missing or invalid 'username' in workflow state")
		}

		wfDesc := getWorkflowDescription(wf.Name)

		// Admin notification
		adminNotif := notification.BillingNotification(adminID).
			Success(fmt.Sprintf("User %s was credited successfully, money transferred successfully to their account (Amount: $%.2f)", username, amountUSD)).
			WithSubject("Money transfer to user's account succeeded").
			WithStatus("succeeded").
			WithExtra("amount", fmt.Sprintf("%.2f", amountUSD)).
			WithExtra("workflow_name", wfDesc).
			WithChannels(notification.ChannelUI).
			Build()

		if err != nil {
			adminNotif = notification.BillingNotification(adminID).
				Failure(fmt.Sprintf("Money transfer to user %s's account failed", username), err).
				WithSubject("Money transfer to user's account failed").
				WithChannels(notification.ChannelUI).
				Build()
		}

		if sendErr := notificationDispatcher.Send(ctx, adminNotif); sendErr != nil {
			return sendErr
		}

		// User notification
		userBuilder := notification.BillingNotification(userID)
		if err != nil {
			userBuilder = userBuilder.Failure("Funds transfer to your account failed", err).
				WithSubject("Your Account Credit Failed")
		} else {
			userBuilder = userBuilder.Success(fmt.Sprintf("Funds were credited to your account. Amount added: $%.2f.", amountUSD)).
				WithSubject("Your Account Has Been Credited").
				WithStatus("succeeded").
				WithExtra("amount", fmt.Sprintf("%.2f", amountUSD))
		}

		userNotif := userBuilder.
			WithExtra("workflow_name", wfDesc).
			WithChannels(notification.ChannelEmail).
			Build()

		return notificationDispatcher.Send(ctx, userNotif)
	}

	// Extract amount and balance from workflow state
	if amountVal, ok := wf.State["amount"]; ok {
		if amount, okAmount := amountVal.(uint64); okAmount {
			amountUSD = substrate.FromUSDMilliCentToUSD(amount)
		}
	}

	if balanceVal, exists := wf.State["net_balance"]; exists {
		if balance, okBalance := balanceVal.(uint64); okBalance {
			newBalanceUSD = substrate.FromUSDMilliCentToUSD(balance)
		}
	}

	var status, subject, message string
	if err == nil {
		status = "funds_succeeded"
		subject = "Adding Funds Succeeded"
		message = fmt.Sprintf("Funds were added successfully to your account. Amount added: $%.2f. New balance will be: $%.2f.", amountUSD, newBalanceUSD)

		if wf.Name == WorkflowRedeemVoucher {
			status = "voucher_redeemed"
			subject = "Voucher Redeemed"
			message = fmt.Sprintf("Voucher redeemed successfully. Amount added: $%.2f.", amountUSD)
		}

		notif := notification.BillingNotification(userID).
			Success(message).
			WithSubject(subject).
			WithStatus(status).
			WithExtra("workflow_name", getWorkflowDescription(wf.Name)).
			WithExtra("amount", fmt.Sprintf("%.2f", amountUSD)).
			Build()

		return notificationDispatcher.Send(ctx, notif)
	}

	// Error case
	subject = "Adding Funds Failed"
	message = fmt.Sprintf("Failed to add funds to your account: %s", err.Error())

	notif := notification.BillingNotification(userID).
		Failure(message, err).
		WithSubject(subject).
		WithExtra("workflow_name", getWorkflowDescription(wf.Name)).
		WithExtra("amount", fmt.Sprintf("%.2f", amountUSD)).
		Build()

	return notificationDispatcher.Send(ctx, notif)
}

func sendNodeWorkflowNotification(ctx context.Context, notificationDispatcher *notification.NotificationDispatcher, wf *ewf.Workflow, err error) error {
	log := logger.ForOperation("workflow", "create_node_notification").With().Str("workflow_name", wf.Name).Logger()
	userIDVal, ok := wf.State["user_id"]
	if !ok {
		log.Error().Msg("Missing 'user_id' in workflow state")
		return nil
	}

	var userID int
	switch v := userIDVal.(type) {
	case int:
		userID = v
	case float64:
		userID = int(v)
	case int64:
		userID = int(v)
	default:
		log.Error().Interface("user_id_value", v).Msg("Invalid 'user_id' type in workflow state")
		return nil
	}

	// Extract node information from workflow state
	var nodeID uint32
	var contractID uint64

	if nodeIDVal, ok := wf.State["node_id"]; ok {
		switch v := nodeIDVal.(type) {
		case uint32:
			nodeID = v
		case float64:
			nodeID = uint32(v)
		case int64:
			nodeID = uint32(v)
		case int:
			nodeID = uint32(v)
		}
	}
	if contractIDVal, ok := wf.State["contract_id"]; ok {
		switch v := contractIDVal.(type) {
		case uint64:
			contractID = v
		case float64:
			contractID = uint64(v)
		case int64:
			contractID = uint64(v)
		case int:
			contractID = uint64(v)
		}
	}

	wfDesc := getWorkflowDescription(wf.Name)
	var message, subject string

	// default workflow reserve node
	subject = "Node Reserved Successfully"
	message = fmt.Sprintf("Node %d has been reserved successfully (contract_id=%d)", nodeID, contractID)

	if err == nil {
		if wf.Name == WorkflowUnreserveNode {
			subject = "Node Unreserved Successfully"
			message = fmt.Sprintf("Node %d has been unreserved successfully (contract_id=%d)", nodeID, contractID)
		}
	} else {
		subject = "Node Reservation Failed"
		message = fmt.Sprintf("Failed to reserve node %d: %s", nodeID, err.Error())
		if wf.Name == WorkflowUnreserveNode {
			subject = "Node Unreservation Failed"
			message = fmt.Sprintf("Failed to unreserve node %d: %s", nodeID, err.Error())

			if NodeHasActiveContracts != "" && err.Error() == NodeHasActiveContracts {
				message = fmt.Sprintf("Failed to unreserve node %d (contract_id=%d). This node has active workloads on it, please remove all deployments from it first", nodeID, contractID)
			}
		}
	}

	var builder *notification.NotificationBuilder
	if err != nil {
		builder = notification.NodeNotification(userID, nodeID).
			Failure(message, err).
			WithSubject(subject)
	} else {
		builder = notification.NodeNotification(userID, nodeID).
			Success(message).
			WithSubject(subject)
	}

	notif := builder.
		WithExtra("workflow_name", wfDesc).
		WithExtra("contract_id", fmt.Sprintf("%d", contractID)).
		WithChannels(notification.ChannelUI).
		NoPersist().
		Build()

	return notificationDispatcher.Send(ctx, notif)
}

func sendUserWorkflowNotification(ctx context.Context, notificationDispatcher *notification.NotificationDispatcher, wf *ewf.Workflow, err error) error {
	log := logger.ForOperation("workflow", "create_user_notification").With().Str("workflow_name", wf.Name).Logger()
	userIDVal, ok := wf.State["user_id"]
	if !ok {
		log.Error().Msg("Missing 'user_id' in workflow state")
		return nil
	}

	var userID int
	switch v := userIDVal.(type) {
	case int:
		userID = v
	case float64:
		userID = int(v)
	case int64:
		userID = int(v)
	default:
		log.Error().Interface("user_id_value", v).Msg("Invalid 'user_id' type in workflow state")
		return nil
	}

	wfDesc := getWorkflowDescription(wf.Name)
	var subject, message string

	// default workflow verified
	subject = "Account Verified Successfully"
	message = "Your account has been verified successfully"

	if err == nil {
		if wf.Name == WorkflowUserRegistration {
			subject = "Registration Completed"
			message = "Your registration has been completed successfully"
		}
	} else {
		subject = "Account Verification Failed"
		message = fmt.Sprintf("Account verification process failed: %s", err.Error())
		if wf.Name == WorkflowUserRegistration {
			subject = "User Registration Failed"
			message = fmt.Sprintf("User registration process failed: %s", err.Error())
		}
	}

	var builder *notification.NotificationBuilder
	if err != nil {
		builder = notification.UserNotification(userID).
			Failure(message, err).
			WithSubject(subject)
	} else {
		builder = notification.UserNotification(userID).
			Success(message).
			WithSubject(subject)
	}

	notif := builder.
		WithExtra("workflow_name", wfDesc).
		NoPersist().
		Build()

	return notificationDispatcher.Send(ctx, notif)
}

func workflowToNotificationType(workflowName string) models.NotificationType {
	billingWf := []string{WorkflowChargeBalance, WorkflowAdminCreditBalance, WorkflowRedeemVoucher}
	deployWf := []string{WorkflowDeleteAllClusters, WorkflowDeleteCluster, WorkflowRemoveNode, WorkflowAddNode, WorkflowRollbackFailedDeployment}
	nodesWf := []string{WorkflowReserveNode, WorkflowUnreserveNode}
	userWf := []string{WorkflowUserVerification, WorkflowUserRegistration}

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

func calculateCurrentStep(stepName string) int {
	switch stepName {
	case StepDeployNetwork:
		return 1
	case StepDeployLeaderNode:
		return 2
	case StepBatchDeployAllNodes:
		return 3
	default:
		return 0
	}
}
