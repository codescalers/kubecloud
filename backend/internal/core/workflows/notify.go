package workflows

import (
	"context"
	"fmt"
	"kubecloud/internal/core/models"
	"kubecloud/internal/deployment/kubedeployer"
	"kubecloud/internal/deployment/statemanager"
	"kubecloud/internal/infrastructure/gridclient"
	"kubecloud/internal/infrastructure/notification"

	"slices"

	"kubecloud/internal/infrastructure/logger"

	"github.com/xmonader/ewf"
)

func notifyWorkflowProgress(notificationDispatcher *notification.NotificationDispatcher) ewf.AfterWorkflowHook {
	return func(ctx context.Context, wf *ewf.Workflow, err error) {
		log := logger.ForOperation("workflow", "notify_workflow_progress").With().Str("workflow_name", wf.Name).Logger()
		suppressNotification, _ := getFromState[bool](wf.State, "suppress_notification")
		if suppressNotification {
			log.Info().Msg("Suppressing notification for workflow")
			return
		}
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
		notif := buildGenericWorkflowNotification(wf, config.UserID, err)
		return notificationDispatcher.Send(ctx, notif)
	}

	displayName := getWorkflowDisplayName(wf)

	if err != nil {
		var nodeInfo string
		var nodeID uint32

		message := fmt.Sprintf("%s for cluster '%s' failed", displayName, cluster.Name)

		// Add node information if available
		if node, nodeErr := getFromState[kubedeployer.Node](wf.State, "node"); nodeErr == nil {
			nodeInfo = node.Name
			nodeID = node.NodeID
			message = fmt.Sprintf("%s for cluster '%s', node '%s' (node_id=%d) failed", displayName, cluster.Name, node.Name, node.NodeID)
		}

		notif := notification.ClusterNotification(config.UserID, cluster.Name).
			Failure(message, err).
			WithSubject(fmt.Sprintf("%s failed", displayName)).
			WithExtra("workflow_name", displayName).
			WithExtra("node_name", nodeInfo).
			WithExtra("node_id", fmt.Sprintf("%d", nodeID)).
			Build()

		return notificationDispatcher.Send(ctx, notif)
	}

	message := fmt.Sprintf("%s completed successfully for cluster '%s' with %d nodes", displayName, cluster.Name, len(cluster.Nodes))
	notif := notification.ClusterNotification(config.UserID, cluster.Name).
		Success(message).
		WithSubject(fmt.Sprintf("%s completed successfully", displayName)).
		WithExtra("workflow_name", displayName).
		WithExtra("node_count", fmt.Sprintf("%d", len(cluster.Nodes))).
		Build()

	return notificationDispatcher.Send(ctx, notif)
}

// notifyStepProgress sends step progress notifications
func notifyStepProgress(ctx context.Context, notificationDispatcher *notification.NotificationDispatcher, state ewf.State, workflowName, stepName string, status string) {
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

	notif := notification.ClusterNotification(config.UserID, clusterName).
		Info(fmt.Sprintf("Deploying cluster %q is in progress - %s: %s", clusterName, stepName, status)).
		WithSubject("Cluster Deployment").
		WithChannels(notification.ChannelUI).
		WithExtra("workflow_name", workflowName).
		WithExtra("step_name", stepName).
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

		displayName := getWorkflowDisplayName(wf)
		if err != nil {
			if attempts < maxAttempts {
				attempts++
				notifyStepProgress(ctx, notificationDispatcher, wf.State, displayName, step.Name, "retrying")
				wf.State[attemptKey] = attempts
				return
			}
			notifyStepProgress(ctx, notificationDispatcher, wf.State, displayName, step.Name, "failed")
		} else {
			notifyStepProgress(ctx, notificationDispatcher, wf.State, displayName, step.Name, "completed")
		}
	}
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

	if isDrainWorkflow(wf.Name) {
		return sendDrainWorkflowNotification(ctx, notificationDispatcher, wf, err)
	}

	config, confErr := getConfig(wf.State)
	if confErr != nil {
		log.Error().Msg("Missing or invalid 'config' in workflow state")
		return confErr
	}

	displayName := getWorkflowDisplayName(wf)

	amount, amountErr := getUint64FromState(wf.State, "amount")
	if amountErr != nil {
		log.Error().Err(amountErr).Msg("failed to get amount from state")
		notif := buildGenericWorkflowNotification(wf, config.UserID, err)
		return notificationDispatcher.Send(ctx, notif)
	}

	newBalance, newBalanceUSDErr := getUint64FromState(wf.State, "new_balance")
	if newBalanceUSDErr != nil {
		log.Error().Err(newBalanceUSDErr).Msg("failed to get new balance from state")
		notif := buildGenericWorkflowNotification(wf, config.UserID, err)
		return notificationDispatcher.Send(ctx, notif)
	}

	amountUSD := gridclient.FromUSDMilliCentToUSD(amount)
	newBalanceUSD := gridclient.FromUSDMilliCentToUSD(newBalance)
	var status, subject, message string
	if err == nil {
		status = "funds_succeeded"
		subject = "Adding Funds Succeeded"
		message = fmt.Sprintf("Funds were added successfully to your account. Amount added: $%.2f. New balance will be: $%.2f.", amountUSD, newBalanceUSD)

		notif := notification.BillingNotification(config.UserID).
			Success(message).
			WithSubject(subject).
			WithStatus(status).
			WithExtra("workflow_name", displayName).
			WithExtra("amount", fmt.Sprintf("%.2f", amountUSD)).
			Build()

		return notificationDispatcher.Send(ctx, notif)
	}

	// Error case
	subject = "Adding Funds Failed"
	message = fmt.Sprintf("Failed to add funds to your account: %s", err.Error())

	notif := notification.BillingNotification(config.UserID).
		Failure(message, err).
		WithSubject(subject).
		WithExtra("workflow_name", displayName).
		WithExtra("amount", fmt.Sprintf("%.2f", amountUSD)).
		Build()

	return notificationDispatcher.Send(ctx, notif)
}

func sendNodeWorkflowNotification(ctx context.Context, notificationDispatcher *notification.NotificationDispatcher, wf *ewf.Workflow, err error) error {
	log := logger.ForOperation("workflow", "create_node_notification").With().Str("workflow_name", wf.Name).Logger()
	config, confErr := getConfig(wf.State)
	if confErr != nil {
		log.Error().Msg("Missing or invalid 'config' in workflow state")
		return confErr
	}

	nodeID, nodeIDErr := getUint32FromState(wf.State, "node_id")
	if nodeIDErr != nil {
		log.Warn().Err(nodeIDErr).Msg("failed to get node ID from state")
		notif := buildGenericWorkflowNotification(wf, config.UserID, err)
		return notificationDispatcher.Send(ctx, notif)
	}
	contractID, contractIDErr := getUint64FromState(wf.State, "contract_id")
	if contractIDErr != nil {
		log.Warn().Err(contractIDErr).Msg("failed to get contract ID from state")
		notif := buildGenericWorkflowNotification(wf, config.UserID, err)
		return notificationDispatcher.Send(ctx, notif)
	}

	displayName := getWorkflowDisplayName(wf)
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
		builder = notification.NodeNotification(config.UserID, nodeID).
			Failure(message, err).
			WithSubject(subject)
	} else {
		builder = notification.NodeNotification(config.UserID, nodeID).
			Success(message).
			WithSubject(subject)
	}

	notif := builder.
		WithExtra("workflow_name", displayName).
		WithExtra("contract_id", fmt.Sprintf("%d", contractID)).
		WithChannels(notification.ChannelUI).
		NoPersist().
		Build()

	return notificationDispatcher.Send(ctx, notif)
}

func sendUserWorkflowNotification(ctx context.Context, notificationDispatcher *notification.NotificationDispatcher, wf *ewf.Workflow, err error) error {
	log := logger.ForOperation("workflow", "create_user_notification").With().Str("workflow_name", wf.Name).Logger()
	config, confErr := getConfig(wf.State)
	if confErr != nil {
		log.Error().Msg("Missing or invalid 'config' in workflow state")
		return confErr
	}

	var subject, message string
	displayName := getWorkflowDisplayName(wf)

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
		builder = notification.UserNotification(config.UserID).
			Failure(message, err).
			WithSubject(subject)
	} else {
		builder = notification.UserNotification(config.UserID).
			Success(message).
			WithSubject(subject)
	}

	notif := builder.
		WithExtra("workflow_name", displayName).
		NoPersist().
		Build()

	return notificationDispatcher.Send(ctx, notif)
}

func workflowToNotificationType(workflowName string) models.NotificationType {
	billingWf := []string{WorkflowChargeBalance, WorkflowDrainUser, WorkflowDrainAllUsers}
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

func getWorkflowDisplayName(workflow *ewf.Workflow) string {
	if workflow.DisplayName != "" {
		return workflow.DisplayName
	}
	return workflow.Name
}

func isDrainWorkflow(name string) bool {
	return name == WorkflowDrainUser || name == WorkflowDrainAllUsers
}

func sendDrainWorkflowNotification(ctx context.Context, notificationDispatcher *notification.NotificationDispatcher, wf *ewf.Workflow, err error) error {
	log := logger.ForOperation("workflow", "create_drain_notification").With().Str("workflow_name", wf.Name).Logger()
	config, confErr := getConfig(wf.State)
	if confErr != nil {
		log.Error().Msg("Missing or invalid 'config' in workflow state")
		return confErr
	}

	if wf.Name == WorkflowDrainAllUsers {
		builder := notification.BillingNotification(config.UserID).
			WithSubject(getWorkflowDisplayName(wf)).
			WithChannels(notification.ChannelUI).
			NoPersist().
			WithExtra("workflow_name", getWorkflowDisplayName(wf))

		if err != nil {
			message := "Draining all users balance failed"
			notif := builder.Failure(message, err).Build()
			return notificationDispatcher.Send(ctx, notif)
		}

		message := "Drained balance for all users successfully"
		notif := builder.Success(message).Build()
		return notificationDispatcher.Send(ctx, notif)
	}

	targetUsername, errTargetUsername := getFromState[string](wf.State, "target_username")
	if errTargetUsername != nil {
		log.Error().Err(errTargetUsername).Msg("failed to get target username from state")
		notif := buildGenericWorkflowNotification(wf, config.UserID, err)
		return notificationDispatcher.Send(ctx, notif)
	}

	targetUserID, errTargetUserID := getFromState[int](wf.State, "target_user_id")
	if errTargetUserID != nil {
		log.Error().Err(errTargetUserID).Msg("failed to get target user ID from state")
		notif := buildGenericWorkflowNotification(wf, config.UserID, err)
		return notificationDispatcher.Send(ctx, notif)
	}

	builder := notification.BillingNotification(config.UserID).
		WithSubject(getWorkflowDisplayName(wf)).
		WithChannels(notification.ChannelUI).
		NoPersist().
		WithExtra("workflow_name", getWorkflowDisplayName(wf)).
		WithExtra("target_user_id", fmt.Sprintf("%d", targetUserID)).
		WithExtra("target_username", targetUsername)

	if err != nil {
		message := fmt.Sprintf("Draining balance for %s failed", targetUsername)
		notif := builder.Failure(message, err).Build()
		return notificationDispatcher.Send(ctx, notif)
	}

	message := fmt.Sprintf("Drained balance for %s successfully", targetUsername)
	notif := builder.Success(message).Build()
	return notificationDispatcher.Send(ctx, notif)
}

func buildGenericWorkflowNotification(wf *ewf.Workflow, userID int, err error) *models.Notification {

	displayName := getWorkflowDisplayName(wf)
	notificationType := workflowToNotificationType(wf.Name)

	var message, subject string
	if err != nil {
		subject = fmt.Sprintf("%s failed", displayName)
		message = fmt.Sprintf("%s failed: %s", displayName, err.Error())
		return notification.NewNotification(userID, notificationType).
			Failure(message, err).
			WithSubject(subject).WithExtra("workflow_name", displayName).Build()
	}
	subject = fmt.Sprintf("%s completed", displayName)
	message = fmt.Sprintf("%s completed successfully", displayName)
	return notification.NewNotification(userID, notificationType).
		Success(message).
		WithSubject(subject).WithExtra("workflow_name", displayName).Build()

}
