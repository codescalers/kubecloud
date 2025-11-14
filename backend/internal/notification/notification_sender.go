package notification

import (
	"context"
	"fmt"
	"kubecloud/internal"
	"kubecloud/internal/models"
	"strconv"
	"strings"
	"time"
)

const TimestampFormat = "Mon, 02 Jan 2006 15:04"

type NotificationSender interface {
	ReloadNotificationConfig(cfg internal.NotificationConfig) error
	GetNotifiers() map[string]Notifier

	SendVoucherGenerationNotification(adminID, vouchersCount int) error
	SendUserVerificationNotification(userID int) error
	SendPasswordChangeNotification(userID int) error
	SendUserRegistrationAndVerificationNotification(userID int, wfName, wfDesc string, err error) error

	SendNewSSHKeyAddedNotification(userID int, sshKeyName string) error
	SendDeletedSSHKeyNotification(userID int, sshKeyName string) error

	SendUnhealthyNodesNotification(userID int, nodeIDs []uint32) error
	SendWorkflowStartedNotification(userID int, workflowDesc, workflowName string, notificationType models.NotificationType) error

	SendClusterFailedHealthCheckNotification(
		userID, clusterNodesLen int, clusterFailure, clusterInStateFailure error, wfDesc, clusterName string,
	) error
	SendClusterStepNotification(
		userID, retryCount, maxRetries int, clusterErr error, status, stepName, workflowName string, clusterName string,
	) error
	SendClusterSuccessNotification(userID, nodeCount int, clusterName, wfDesc string, clusterErr error) error
	SendClusterFailureNotification(userID int, clusterName, nodeInfo, wfDesc, message string, nodeID uint32, clusterError error) error

	SendNodeReservationsNotification(userID int, wfName, wfDesc string, contractID uint64, nodeID uint32, nodeErr error) error

	SendCreditedUserNotification(userID, adminID int, wfDesc, username string, amountUSD float64, err error) error
	SendRedeemedVoucherNotification(userID int, wfName, wfDesc string, amountUSD, newBalanceUSD float64, voucherErr error) error
}

type EmailAndUINotificationSender struct {
	ctx                 context.Context
	notificationService NotificationServiceInterface
}

func NewEmailAndUINotificationSender(ctx context.Context, notificationService NotificationServiceInterface) EmailAndUINotificationSender {
	return EmailAndUINotificationSender{
		ctx:                 ctx,
		notificationService: notificationService,
	}
}

func (s EmailAndUINotificationSender) ReloadNotificationConfig(cfg internal.NotificationConfig) error {
	return s.notificationService.ReloadNotificationConfig(cfg)
}

func (s EmailAndUINotificationSender) GetNotifiers() map[string]Notifier {
	return s.notificationService.GetNotifiers()
}

func (s EmailAndUINotificationSender) SendVoucherGenerationNotification(adminID, vouchersCount int) error {
	payload := MergePayload(CommonPayload{
		Message: fmt.Sprintf("%d vouchers generated successfully.", vouchersCount),
		Subject: "Vouchers Generated",
		Status:  "succeeded",
	}, map[string]string{})

	notification := models.NewNotification(
		adminID,
		models.NotificationTypeBilling,
		payload,
		models.WithChannels(ChannelUI),
		models.WithSeverity(models.NotificationSeveritySuccess),
		models.WithNoPersist(),
	)

	return s.notificationService.Send(s.ctx, notification)
}

func (s EmailAndUINotificationSender) SendUserVerificationNotification(userID int) error {
	payload := CommonPayload{
		Message: "User email is verified successfully",
		Subject: "User email verified",
	}

	notification := models.NewNotification(
		userID,
		"user_registration",
		MergePayload(payload, map[string]string{}),
		models.WithNoPersist(), models.WithChannels(ChannelUI),
		models.WithSeverity(models.NotificationSeveritySuccess),
	)

	return s.notificationService.Send(s.ctx, notification)
}

func (s EmailAndUINotificationSender) SendPasswordChangeNotification(userID int) error {
	payload := CommonPayload{
		Status:  "password_changed",
		Subject: "Your password was changed",
		Message: "Your account password has been successfully updated.",
	}

	notification := models.NewNotification(userID, models.NotificationTypeUser, MergePayload(payload, map[string]string{}))
	return s.notificationService.Send(s.ctx, notification)
}

func (s EmailAndUINotificationSender) SendNewSSHKeyAddedNotification(userID int, sshKeyName string) error {
	payload := CommonPayload{
		Status:  "ssh_key_added",
		Subject: "New SSH key added",
		Message: fmt.Sprintf("SSH key '%s' was added to your account.", sshKeyName),
	}

	notification := models.NewNotification(userID, models.NotificationTypeUser, MergePayload(payload, map[string]string{}))
	return s.notificationService.Send(s.ctx, notification)
}

func (s EmailAndUINotificationSender) SendDeletedSSHKeyNotification(userID int, sshKeyName string) error {
	payload := CommonPayload{
		Status:  "ssh_key_deleted",
		Subject: "SSH key deleted",
		Message: fmt.Sprintf("SSH key '%s' was deleted from your account.", sshKeyName),
	}

	notification := models.NewNotification(userID, models.NotificationTypeUser, MergePayload(payload, map[string]string{}), models.WithSeverity(models.NotificationSeveritySuccess))
	return s.notificationService.Send(s.ctx, notification)
}

func (s EmailAndUINotificationSender) SendUnhealthyNodesNotification(userID int, nodeIDs []uint32) error {
	subject := "Reserved Node Health Check Failed"

	var b strings.Builder
	for i, id := range nodeIDs {
		if i > 0 {
			b.WriteString("\n")
		}
		b.WriteString(fmt.Sprintf("Node ID: %d", id))
	}

	message := fmt.Sprintf(
		"You have %d reserved node(s) that are currently unhealthy",
		len(nodeIDs),
	)

	payloadData := map[string]string{
		"unhealthy_count": fmt.Sprintf("%d", len(nodeIDs)),
		"timestamp":       time.Now().UTC().Format("2006-01-02 15:04:05 UTC"),
	}
	payloadData["nodes_list"] = b.String()

	payload := MergePayload(CommonPayload{
		Subject: subject,
		Message: message,
		Status:  "unhealthy",
	}, payloadData)

	notification := models.NewNotification(
		userID,
		models.NotificationTypeNode,
		payload,
		models.WithSeverity(models.NotificationSeverityError),
		models.WithChannels(ChannelEmail),
	)

	return s.notificationService.Send(s.ctx, notification)
}

func (s EmailAndUINotificationSender) SendWorkflowStartedNotification(userID int, workflowDesc, workflowName string, notificationType models.NotificationType) error {
	subject := fmt.Sprintf("%s Started", workflowDesc)
	message := fmt.Sprintf("%s has been started", workflowDesc)

	payload := MergePayload(CommonPayload{
		Subject: subject,
		Message: message,
		Status:  "started",
	}, map[string]string{
		"workflow_name": workflowDesc,
	})

	notification := models.NewNotification(userID, notificationType, payload, models.WithNoPersist())
	return s.notificationService.Send(s.ctx, notification)
}

func (s EmailAndUINotificationSender) SendClusterFailedHealthCheckNotification(
	userID, clusterNodesLen int, clusterFailure, clusterInStateFailure error,
	wfDesc, clusterName string,
) error {
	severity := models.NotificationSeverityError
	payload := MergePayload(CommonPayload{
		Message: "Cluster health check failed",
		Subject: "Cluster health check failed",
		Status:  "failed",
		Error:   clusterFailure.Error(),
	}, map[string]string{
		"workflow_name": wfDesc,
		"timestamp":     time.Now().UTC().Format(TimestampFormat),
	})

	if clusterInStateFailure != nil {
		notification := models.NewNotification(userID, models.NotificationTypeDeployment, payload, models.WithSeverity(severity), models.WithChannels(ChannelEmail))
		return s.notificationService.Send(s.ctx, notification)
	}

	payload["message"] = fmt.Sprintf("Cluster health check failed for cluster Name: %s, Number of nodes: %d", clusterName, clusterNodesLen)
	payload["cluster_name"] = clusterName
	notificationObj := models.NewNotification(userID, models.NotificationTypeDeployment, payload, models.WithSeverity(severity), models.WithChannels(ChannelEmail))

	return s.notificationService.Send(s.ctx, notificationObj)
}

func (s EmailAndUINotificationSender) SendClusterStepNotification(userID, retryCount, maxRetries int, clusterErr error, status, stepName, workflowName string, clusterName string) error {
	currentStep := calculateCurrentStep(stepName)
	total := 4
	progressStr := fmt.Sprintf(" (%d/%d)", currentStep, total)

	var message string
	var notificationType string
	switch {
	case clusterErr != nil:
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
		return nil
	}

	payload := MergePayload(
		CommonPayload{
			Subject: "Cluster Deployment Progress",
			Status:  status,
			Message: fmt.Sprintf("Deploying cluster %q - %s", clusterName, message),
		},
		map[string]string{
			"workflow_name": workflowName,
			"step_name":     stepName,
			"cluster_name":  clusterName,
			"user_id":       fmt.Sprintf("%d", userID),
			"current_step":  strconv.Itoa(currentStep),
			"total_steps":   strconv.Itoa(total),
		},
	)
	if clusterErr != nil {
		payload["error"] = clusterErr.Error()
	}

	notification := models.NewNotification(
		userID,
		models.NotificationType(notificationType),
		payload,
		models.WithNoPersist(),
		models.WithChannels(ChannelUI),
		models.WithSeverity(models.NotificationSeverityInfo),
	)

	return s.notificationService.Send(s.ctx, notification)
}

func (s EmailAndUINotificationSender) SendClusterSuccessNotification(userID, nodeCount int, clusterName, wfDesc string, clusterErr error) error {
	var notificationPayload map[string]string

	if clusterErr != nil {
		notificationPayload = MergePayload(CommonPayload{
			Message: fmt.Sprintf("%s completed successfully", wfDesc),
			Subject: fmt.Sprintf("%s completed successfully", wfDesc),
			Status:  "succeeded",
		}, map[string]string{
			"workflow_name": wfDesc,
			"cluster_name":  clusterName,
			"user_id":       fmt.Sprintf("%d", userID),
			"timestamp":     time.Now().Local().Format(TimestampFormat),
		})

	} else {
		totalSteps := nodeCount + 2
		message := fmt.Sprintf("%s completed successfully for cluster '%s' with %d nodes",
			wfDesc, clusterName, nodeCount)

		notificationPayload = MergePayload(CommonPayload{
			Message: message,
			Subject: fmt.Sprintf("%s completed successfully ", wfDesc),
			Status:  "succeeded",
		}, map[string]string{
			"workflow_name": wfDesc,
			"cluster_name":  clusterName,
			"node_count":    fmt.Sprintf("%d", nodeCount),
			"total_steps":   fmt.Sprintf("%d", totalSteps),
			"user_id":       fmt.Sprintf("%d", userID),
			"timestamp":     time.Now().Local().Format(TimestampFormat),
		})
	}

	notification := models.NewNotification(userID, models.NotificationTypeDeployment, notificationPayload)
	return s.notificationService.Send(s.ctx, notification)
}
func (s EmailAndUINotificationSender) SendClusterFailureNotification(userID int, clusterName, nodeInfo, wfDesc, message string, nodeID uint32, clusterError error) error {
	payloadData := map[string]string{
		"workflow_name": wfDesc,
		"cluster_name":  clusterName,
		"node_name":     nodeInfo,
		"user_id":       fmt.Sprintf("%d", userID),
		"timestamp":     time.Now().Local().Format(TimestampFormat),
	}
	if nodeID > 0 {
		payloadData["node_id"] = fmt.Sprintf("%d", nodeID)
	}

	notificationPayload := MergePayload(CommonPayload{
		Message: message,
		Error:   clusterError.Error(),
		Subject: fmt.Sprintf("%s failed", wfDesc),
		Status:  "failed",
	}, payloadData)

	notification := models.NewNotification(userID, models.NotificationTypeDeployment, notificationPayload)
	return s.notificationService.Send(s.ctx, notification)
}

func (s EmailAndUINotificationSender) SendCreditedUserNotification(userID, adminID int, wfDesc, username string, amountUSD float64, creditErr error) error {
	payloadMap := map[string]string{
		"workflow_name": wfDesc,
		"user_id":       fmt.Sprintf("%d", userID),
		"admin_id":      fmt.Sprintf("%d", adminID),
		"timestamp":     time.Now().Local().Format(TimestampFormat),
	}
	if username != "" {
		payloadMap["username"] = username
	}

	payloadData := CommonPayload{}
	if username != "" {
		payloadData.Message = fmt.Sprintf("User %s was credited successfully, money transferred successfully to their account (Amount: $%.2f)", username, amountUSD)
	} else {
		payloadData.Message = fmt.Sprintf("User was credited successfully, money transferred successfully to their account (Amount: $%.2f)", amountUSD)
	}

	payloadData.Subject = "Money transfer to user's account succeeded"
	payloadData.Status = "succeeded"

	severity := models.NotificationSeveritySuccess
	if creditErr != nil {
		severity = models.NotificationSeverityError
		if username != "" {
			payloadData.Message = fmt.Sprintf("Money transfer to user %s's account failed", username)
		} else {
			payloadData.Message = "Money transfer to user's account failed"
		}
		payloadData.Subject = "Money transfer to user's account failed"
		payloadData.Error = creditErr.Error()

		adminNotification := models.NewNotification(adminID, models.NotificationTypeBilling, MergePayload(payloadData, payloadMap), models.WithSeverity(severity), models.WithChannels(ChannelUI))
		return s.notificationService.Send(s.ctx, adminNotification)
	}

	payloadMap["amount"] = fmt.Sprintf("%.2f", amountUSD)

	adminNotification := models.NewNotification(adminID, models.NotificationTypeBilling, MergePayload(payloadData, payloadMap), models.WithSeverity(severity), models.WithChannels(ChannelUI))
	if err := s.notificationService.Send(s.ctx, adminNotification); err != nil {
		return s.notificationService.Send(s.ctx, adminNotification)
	}

	// also notify the user about success
	userPayload := MergePayload(CommonPayload{
		Message: fmt.Sprintf("Funds were credited to your account. Amount added: $%.2f.", amountUSD),
		Subject: "Your Account Has Been Credited",
		Status:  "succeeded",
	}, map[string]string{
		"workflow_name": wfDesc,
		"user_id":       fmt.Sprintf("%d", userID),
		"timestamp":     time.Now().Local().Format(TimestampFormat),
		"amount":        fmt.Sprintf("%.2f", amountUSD),
	})

	userNotification := models.NewNotification(userID, models.NotificationTypeBilling, userPayload, models.WithSeverity(severity), models.WithChannels(ChannelEmail))
	return s.notificationService.Send(s.ctx, userNotification)
}

func (s EmailAndUINotificationSender) SendRedeemedVoucherNotification(userID int, wfName, wfDesc string, amountUSD, newBalanceUSD float64, voucherErr error) error {
	var status, subject, message string
	if voucherErr == nil {
		status = "funds_succeeded"
		subject = "Adding Funds Succeeded"
		message = fmt.Sprintf("Funds were added successfully to your account. Amount added: $%.2f. New balance will be: $%.2f.", amountUSD, newBalanceUSD)

		if wfName == internal.WorkflowRedeemVoucher {
			status = "voucher_redeemed"
			subject = "Voucher Redeemed"
			message = fmt.Sprintf("Voucher redeemed successfully. Amount added: $%.2f.", amountUSD)
		}
	} else {
		status = "funds_failed"
		subject = "Adding Funds Failed"
		message = fmt.Sprintf("Failed to add funds to your account: %s", voucherErr.Error())
	}
	payloadData := map[string]string{
		"workflow_name": wfDesc,
		"user_id":       fmt.Sprintf("%d", userID),
		"timestamp":     time.Now().Local().Format(TimestampFormat),
	}
	if amountUSD > 0 {
		payloadData["amount"] = fmt.Sprintf("%.2f", amountUSD)
	}
	if newBalanceUSD > 0 {
		payloadData["balance"] = fmt.Sprintf("%.2f", newBalanceUSD)
	}

	payload := MergePayload(CommonPayload{
		Message: message,
		Subject: subject,
		Status:  status,
	}, payloadData)

	if voucherErr != nil {
		payload["error"] = voucherErr.Error()
	}

	notification := models.NewNotification(userID, models.NotificationTypeBilling, payload)
	return s.notificationService.Send(s.ctx, notification)
}

func (s EmailAndUINotificationSender) SendNodeReservationsNotification(userID int, wfName, wfDesc string, contractID uint64, nodeID uint32, nodeErr error) error {
	var subject, message string
	var payload map[string]string

	//default workflow reserve node
	subject = "Node Reserved Successfully"
	message = fmt.Sprintf("Node %d has been reserved successfully (contract_id=%d)", nodeID, contractID)
	severity := models.NotificationSeveritySuccess
	if nodeErr == nil {
		if wfName == internal.WorkflowUnreserveNode {
			subject = "Node Unreserved Successfully"
			message = fmt.Sprintf("Node %d has been unreserved successfully (contract_id=%d)", nodeID, contractID)
		}
	} else {
		severity = models.NotificationSeverityError
		subject = "Node Reservation Failed"
		message = fmt.Sprintf("Failed to reserve node %d: %s", nodeID, nodeErr.Error())
		if wfName == internal.WorkflowUnreserveNode {
			subject = "Node Unreservation Failed"
			message = fmt.Sprintf("Failed to unreserve node %d: %s", nodeID, nodeErr.Error())

			if strings.Contains(nodeErr.Error(), internal.NodeHasActiveContracts) {
				message = fmt.Sprintf("Failed to unreserve node %d (contract_id=%d). This node has active workloads on it, please remove all deployments from it first", nodeID, contractID)
			}
		}
	}

	// Build payload data
	payloadData := map[string]string{
		"workflow_name": wfDesc,
		"user_id":       fmt.Sprintf("%d", userID),
		"timestamp":     time.Now().Local().Format(TimestampFormat),
	}
	if nodeID > 0 {
		payloadData["node_id"] = fmt.Sprintf("%d", nodeID)
	}
	if contractID > 0 {
		payloadData["contract_id"] = fmt.Sprintf("%d", contractID)
	}

	payload = MergePayload(CommonPayload{
		Message: message,
		Subject: subject,
	}, payloadData)

	if nodeErr != nil {
		payload["error"] = nodeErr.Error()
	}

	notification := models.NewNotification(
		userID,
		models.NotificationTypeNode,
		payload,
		models.WithChannels(ChannelUI),
		models.WithSeverity(severity),
		models.WithNoPersist(),
	)
	return s.notificationService.Send(s.ctx, notification)
}

func (s EmailAndUINotificationSender) SendUserRegistrationAndVerificationNotification(userID int, wfName, wfDesc string, err error) error {
	var payload map[string]string
	var subject, message string

	//default workflow verified
	subject = "Account Verified Successfully"
	message = "Your account has been verified successfully"
	severity := models.NotificationSeveritySuccess

	if err == nil {
		if wfName == internal.WorkflowUserRegistration {
			subject = "Registration Completed"
			message = "Your registration has been completed successfully"
		}
	} else {
		severity = models.NotificationSeverityError
		subject = "Account Verification Failed"
		message = fmt.Sprintf("Account verification process failed: %s", err.Error())
		if wfName == internal.WorkflowUserRegistration {
			subject = "User Registration Failed"
			message = fmt.Sprintf("User registration process failed: %s", err.Error())
		}
	}

	payloadData := map[string]string{
		"workflow_name": wfDesc,
		"user_id":       fmt.Sprintf("%d", userID),
		"timestamp":     time.Now().Local().Format(TimestampFormat),
	}

	payload = MergePayload(CommonPayload{
		Message: message,
		Subject: subject,
	}, payloadData)

	if err != nil {
		payload["error"] = err.Error()
	}

	notification := models.NewNotification(
		userID,
		models.NotificationTypeUser,
		payload,
		models.WithNoPersist(),
		models.WithSeverity(severity),
	)
	return s.notificationService.Send(s.ctx, notification)
}

func calculateCurrentStep(stepName string) int {
	if stepName == internal.StepDeployNetwork {
		return 1
	}

	if stepName == internal.StepDeployLeaderNode {
		return 2
	}

	if stepName == internal.StepBatchDeployAllNodes {
		return 3
	}

	return 0
}
