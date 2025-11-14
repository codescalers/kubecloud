package workflows

import (
	"kubecloud/internal/shared"
	"context"
	"fmt"
	"kubecloud/internal/infrastructure/notification"
	"kubecloud/internal/deployment/statemanager"
	"kubecloud/internal/infrastructure/substrate"
	"kubecloud/internal/deployment/kubedeployer"
	"kubecloud/internal/core/models"
	"slices"

	"kubecloud/internal/infrastructure/logger"

	"github.com/xmonader/ewf"
)

func notifyWorkflowProgress(notificationSender notification.NotificationSender) ewf.AfterWorkflowHook {
	return func(ctx context.Context, wf *ewf.Workflow, err error) {
		log := logger.ForOperation("workflow", "notify_workflow_progress").With().Str("workflow_name", wf.Name).Logger()

		notificationType := workflowToNotificationType(wf.Name)
		switch notificationType {
		case models.NotificationTypeDeployment:
			if err := sendDeploymentWorkflowNotification(notificationSender, wf, err); err != nil {
				log.Error().Err(err).Msg("Failed to send deployment workflow notification")
			}
		case models.NotificationTypeBilling:
			if err := sendBillingWorkflowNotifications(notificationSender, wf, err); err != nil {
				log.Error().Err(err).Msg("Failed to send billing workflow notifications")
			}
		case models.NotificationTypeNode:
			if err := sendNodeWorkflowNotification(notificationSender, wf, err); err != nil {
				log.Error().Err(err).Msg("Failed to send node workflow notification")
			}
		case models.NotificationTypeUser:
			if err := sendUserWorkflowNotification(notificationSender, wf, err); err != nil {
				log.Error().Err(err).Msg("Failed to send user workflow notification")
			}
		}
	}
}

func sendDeploymentWorkflowNotification(notificationSender notification.NotificationSender, wf *ewf.Workflow, err error) error {
	log := logger.ForOperation("workflow", "create_deployment_notification").With().Str("workflow_name", wf.Name).Logger()
	config, confErr := getConfig(wf.State)
	if confErr != nil {
		log.Error().Msg("Missing or invalid 'config' in workflow state")
		return confErr
	}

	workflowDesc := getWorkflowDescription(wf.Name)

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

		return notificationSender.SendClusterFailureNotification(config.UserID, clusterName, nodeInfo, workflowDesc, message, nodeID, err)
	}

	cluster, clusterErr := statemanager.GetCluster(wf.State)
	return notificationSender.SendClusterSuccessNotification(config.UserID, len(cluster.Nodes), cluster.Name, workflowDesc, clusterErr)
}

// notifyStepProgress sends step progress notifications
func notifyStepProgress(notificationSender notification.NotificationSender, state ewf.State, workflowName, stepName string, status string, err error, retryCount, maxRetries int) {
	log := logger.ForOperation("workflow", "notify_step_progress").With().
		Str("workflow_name", workflowName).
		Str("step_name", stepName).Logger()

	if stepName != shared.StepDeployNetwork && !isDeployStep(stepName) {
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

	if err := notificationSender.SendClusterStepNotification(
		config.UserID, retryCount, maxRetries, err, status, stepName, workflowName, clusterName); err != nil {
		log.Error().Err(err).Msg("Failed to send step notification")
	}
}

func notifyStepHook(notificationSender notification.NotificationSender) ewf.AfterStepHook {
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
				notifyStepProgress(notificationSender, wf.State, wf.Name, step.Name, "retrying", err, attempts, maxAttempts)
				wf.State[attemptKey] = attempts
				return
			}
			notifyStepProgress(notificationSender, wf.State, wf.Name, step.Name, "failed", err, 0, 0)
		} else {
			notifyStepProgress(notificationSender, wf.State, wf.Name, step.Name, "completed", nil, 0, 0)
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
	return name == shared.WorkflowDeployCluster
}

func isDeployStep(stepName string) bool {
	return stepName == shared.StepDeployLeaderNode ||
		stepName == shared.StepBatchDeployAllNodes
}

func sendBillingWorkflowNotifications(notificationSender notification.NotificationSender, wf *ewf.Workflow, err error) error {
	log := logger.ForOperation("workflow", "create_billing_notification").With().Str("workflow_name", wf.Name).Logger()
	userID, ok := wf.State["user_id"].(int)
	if !ok {
		log.Error().Msg("Missing or invalid 'user_id' in workflow state")
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

	if wf.Name == shared.WorkflowAdminCreditBalance {
		adminID, ok := wf.State["admin_id"].(int)
		if !ok {
			log.Error().Msg("Missing or invalid 'admin_id' in workflow state")
			return nil
		}
		username, ok := wf.State["username"].(string)
		if !ok {
			log.Warn().Msg("Missing or invalid 'username' in workflow state")
		}

		wfDesc := getWorkflowDescription(wf.Name)
		return notificationSender.SendCreditedUserNotification(userID, adminID, wfDesc, username, amountUSD, err)
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

	return notificationSender.SendRedeemedVoucherNotification(userID, wf.Name, getWorkflowDescription(wf.Name), amountUSD, newBalanceUSD, err)
}

func sendNodeWorkflowNotification(notificationSender notification.NotificationSender, wf *ewf.Workflow, err error) error {
	log := logger.ForOperation("workflow", "create_node_notification").With().Str("workflow_name", wf.Name).Logger()
	userID, ok := wf.State["user_id"].(int)
	if !ok {
		log.Error().Msg("Missing or invalid 'user_id' in workflow state")
		return nil
	}

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

	return notificationSender.SendNodeReservationsNotification(userID, wf.Name, getWorkflowDescription(wf.Name), contractID, nodeID, err)
}

func sendUserWorkflowNotification(notificationSender notification.NotificationSender, wf *ewf.Workflow, err error) error {
	log := logger.ForOperation("workflow", "create_user_notification").With().Str("workflow_name", wf.Name).Logger()
	userID, ok := wf.State["user_id"].(int)
	if !ok {
		log.Error().Msg("Missing or invalid 'user_id' in workflow state")
		return nil
	}

	return notificationSender.SendUserRegistrationAndVerificationNotification(userID, wf.Name, getWorkflowDescription(wf.Name), err)
}

func workflowToNotificationType(workflowName string) models.NotificationType {
	billingWf := []string{shared.WorkflowChargeBalance, shared.WorkflowAdminCreditBalance, shared.WorkflowRedeemVoucher}
	deployWf := []string{shared.WorkflowDeleteAllClusters, shared.WorkflowDeleteCluster, shared.WorkflowRemoveNode, shared.WorkflowAddNode, shared.WorkflowRollbackFailedDeployment}
	nodesWf := []string{shared.WorkflowReserveNode, shared.WorkflowUnreserveNode}
	userWf := []string{shared.WorkflowUserVerification, shared.WorkflowUserRegistration}

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
