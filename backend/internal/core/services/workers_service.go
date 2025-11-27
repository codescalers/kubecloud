package services

import (
	"context"
	"errors"
	"fmt"
	"kubecloud/internal/billing"
	"kubecloud/internal/config"
	distributedlocks "kubecloud/internal/core/distributed_locks"
	"kubecloud/internal/core/models"
	"kubecloud/internal/core/workflows"
	"kubecloud/internal/infrastructure/gridclient"
	"kubecloud/internal/infrastructure/logger"
	"kubecloud/internal/infrastructure/mailservice"
	mailsender "kubecloud/internal/infrastructure/mailservice/mail_sender"
	"kubecloud/internal/infrastructure/notification"
	"slices"
	"strconv"
	"strings"

	"sync"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/graphql"
	"github.com/xmonader/ewf"
)

type WorkerService struct {
	ctx context.Context

	userRepo           models.UserRepository
	nodesRepo          models.UserNodesRepository
	invoicesRepo       models.InvoiceRepository
	clusterRepo        models.ClusterRepository
	pendingRecordsRepo models.PendingRecordRepository

	mailService            mailservice.MailService
	graphql                graphql.GraphQl
	firesquidClient        graphql.GraphQl
	gridClient             gridclient.GridClient
	ewfEngine              *ewf.Engine
	notificationDispatcher *notification.NotificationDispatcher

	// configs
	systemMnemonic                          string
	invoiceCompanyData                      config.InvoiceCompanyData
	currency                                string
	checkUserDebtIntervalInHours            int
	clusterHealthCheckIntervalInHours       int
	reservedNodeHealthCheckIntervalInHours  int
	reservedNodeHealthCheckTimeoutInMinutes int
	reservedNodeHealthCheckWorkersNum       int
	monitorBalanceIntervalInMinutes         int
	notifyAdminsForPendingRecordsInHours    int
	locksReleaseIntervalInMinutes           int

	locker distributedlocks.DistributedLocks
	usersBalanceCheckIntervalInHours        int
}

func NewWorkersService(
	ctx context.Context, userRepo models.UserRepository, nodesRepo models.UserNodesRepository,
	invoicesRepo models.InvoiceRepository, clusterRepo models.ClusterRepository, pendingRecordsRepo models.PendingRecordRepository,
	mailService mailservice.MailService,
	gridClient gridclient.GridClient, ewfEngine *ewf.Engine, notificationDispatcher *notification.NotificationDispatcher,
	graphql graphql.GraphQl, firesquidClient graphql.GraphQl,
	invoiceCompanyData config.InvoiceCompanyData, systemMnemonic, currency string,
	clusterHealthCheckIntervalInHours, reservedNodeHealthCheckIntervalInHours,
	reservedNodeHealthCheckTimeoutInMinutes, reservedNodeHealthCheckWorkersNum,
	monitorBalanceIntervalInMinutes, notifyAdminsForPendingRecordsInHours, locksReleaseIntervalInMinutes int,
	locker distributedlocks.DistributedLocks,
	usersBalanceCheckIntervalInHours int,
	checkUserDebtIntervalInHours int,
) WorkerService {
	return WorkerService{
		ctx:                ctx,
		userRepo:           userRepo,
		nodesRepo:          nodesRepo,
		invoicesRepo:       invoicesRepo,
		clusterRepo:        clusterRepo,
		pendingRecordsRepo: pendingRecordsRepo,

		mailService:            mailService,
		notificationDispatcher: notificationDispatcher,
		ewfEngine:              ewfEngine,
		graphql:                graphql,
		firesquidClient:        firesquidClient,
		gridClient:             gridClient,

		locker: locker,

		systemMnemonic:     systemMnemonic,
		invoiceCompanyData: invoiceCompanyData,
		currency:           currency,

		checkUserDebtIntervalInHours:            checkUserDebtIntervalInHours,
		clusterHealthCheckIntervalInHours:       clusterHealthCheckIntervalInHours,
		reservedNodeHealthCheckIntervalInHours:  reservedNodeHealthCheckIntervalInHours,
		reservedNodeHealthCheckTimeoutInMinutes: reservedNodeHealthCheckTimeoutInMinutes,
		reservedNodeHealthCheckWorkersNum:       reservedNodeHealthCheckWorkersNum,
		monitorBalanceIntervalInMinutes:         monitorBalanceIntervalInMinutes,
		notifyAdminsForPendingRecordsInHours:    notifyAdminsForPendingRecordsInHours,
		locksReleaseIntervalInMinutes:           locksReleaseIntervalInMinutes,
		usersBalanceCheckIntervalInHours:        usersBalanceCheckIntervalInHours,
	}
}

type NodeHealthResult struct {
	userID          int
	unhealthyNodeID uint32
}

func (svc WorkerService) ListAllUsers() ([]models.User, error) {
	return svc.userRepo.ListAllUsers()
}

func (svc WorkerService) ListAllClusters() ([]models.Cluster, error) {
	return svc.clusterRepo.ListAllClusters()
}

func (svc WorkerService) ListAllReservedNodes() ([]models.UserNodes, error) {
	return svc.nodesRepo.ListAllReservedNodes()
}

func (svc WorkerService) ListOnlyPendingRecords() ([]models.PendingRecord, error) {
	return svc.pendingRecordsRepo.ListOnlyPendingRecords()
}

func (svc WorkerService) GetClusterHealthCheckInterval() time.Duration {
	return time.Duration(svc.clusterHealthCheckIntervalInHours) * time.Hour
}

func (svc WorkerService) GetReservedNodeHealthCheckInterval() time.Duration {
	return time.Duration(svc.reservedNodeHealthCheckIntervalInHours) * time.Hour
}

func (svc WorkerService) GetCheckUserDebtInterval() time.Duration {
	return time.Duration(svc.checkUserDebtIntervalInHours) * time.Hour
}

func (svc WorkerService) GetMonitorBalanceInterval() time.Duration {
	return time.Duration(svc.monitorBalanceIntervalInMinutes) * time.Minute
}

func (svc WorkerService) GetNotifyAdminsForPendingRecordsInterval() time.Duration {
	return time.Duration(svc.notifyAdminsForPendingRecordsInHours) * time.Hour
}

func (svc WorkerService) GetLocksReleaseInterval() time.Duration {
	return time.Duration(svc.locksReleaseIntervalInMinutes) * time.Minute
}

func (svc WorkerService) GetUsersBalanceCheckInterval() time.Duration {
	return time.Duration(svc.usersBalanceCheckIntervalInHours) * time.Hour
}

func (svc WorkerService) CreateUserInvoice(user models.User) error {
	records, err := svc.nodesRepo.ListUserNodes(user.ID)
	if err != nil {
		return err
	}

	if len(records) == 0 {
		return nil
	}

	now := time.Now()

	var invoiceItems []models.NodeItem
	var totalInvoiceCostUSD float64

	for _, record := range records {
		billReports, err := billing.ListContractBillReportsPerMonth(svc.graphql, record.ContractID, now)
		if err != nil {
			return err
		}

		totalAmountTFT, err := billing.AmountBilledPerMonth(billReports)
		if err != nil {
			return err
		}
		totalAmountUSDMillicent, err := svc.gridClient.FromTFTtoUSDMillicent(totalAmountTFT)
		if err != nil {
			return err
		}
		rentRecordStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		if record.CreatedAt.After(rentRecordStart) {
			rentRecordStart = record.CreatedAt
		}

		var totalHours int
		cancellationDate, err := billing.GetRentContractCancellationDate(svc.firesquidClient, record.ContractID)

		if errors.Is(err, billing.ErrorEventsNotFound) {
			totalHours = getHoursOfGivenPeriod(rentRecordStart, time.Now())
		} else if err != nil {
			return err
		} else {
			totalHours = getHoursOfGivenPeriod(rentRecordStart, cancellationDate)
		}

		totalAmountUSD := gridclient.FromUSDMilliCentToUSD(totalAmountUSDMillicent)

		invoiceItems = append(invoiceItems, models.NodeItem{
			NodeID:        record.NodeID,
			ContractID:    record.ContractID,
			RentCreatedAt: rentRecordStart,
			PeriodInHours: float64(totalHours),
			Cost:          totalAmountUSD,
		})
		totalInvoiceCostUSD += totalAmountUSD

	}

	invoice := models.Invoice{
		UserID:    user.ID,
		Total:     totalInvoiceCostUSD,
		Nodes:     invoiceItems,
		Tax:       0, //TODO:
		CreatedAt: time.Now(),
	}

	file, err := billing.CreateInvoicePDF(invoice, user, svc.invoiceCompanyData)
	if err != nil {
		return err
	}

	invoice.FileData = file
	if err = svc.invoicesRepo.CreateInvoice(&invoice); err != nil {
		return err
	}

	attachments := []mailsender.Attachment{
		{
			FileName: fmt.Sprintf("invoice-%d-%d.pdf", invoice.UserID, invoice.ID),
			Data:     invoice.FileData,
		},
	}

	return svc.mailService.SendInvoiceMail(user.Email, totalInvoiceCostUSD, svc.currency, invoice.ID, attachments)
}

func (svc WorkerService) UpdateUserDebt() error {
	users, err := svc.userRepo.ListAllUsers()
	if err != nil {
		return err
	}

	for _, user := range users {
		userNodes, err := svc.nodesRepo.ListUserNodes(user.ID)
		if err != nil {
			logger.ForOperation("debt_tracker", "list_user_nodes").Error().Err(err).Msg("Failed to list user nodes")
			continue
		}
		contractIDs := make([]uint64, len(userNodes))
		for i, node := range userNodes {
			contractIDs[i] = node.ContractID
		}
		userDebt, err := svc.calculateDebt(user.Mnemonic, contractIDs, time.Hour)
		if err != nil {
			logger.ForOperation("debt_tracker", "calculate_debt").Error().Err(err).Msg("Failed to calculate user debt")
			continue
		}

		user.Debt = userDebt
		err = svc.userRepo.UpdateUserByID(&user)
		if err != nil {
			logger.ForOperation("debt_tracker", "update_user_debt_db").Error().Err(err).Msg("Failed to update user debt in DB")
		}
	}

	return nil
}

func (svc WorkerService) calculateDebt(userMnemonic string, contractIDs []uint64, debtPeriod time.Duration) (uint64, error) {

	calculatorClient, err := svc.gridClient.NewCalculator(userMnemonic)
	if err != nil {
		return 0, fmt.Errorf("failed to create new calculator: %w", err)
	}

	var totalDebt int64
	for _, contractID := range contractIDs {
		debt, err := calculatorClient.CalculateContractOverdue(contractID, debtPeriod)
		if err != nil {
			logger.ForOperation("debt_tracker", "calc_overdue").Error().Err(err).Msg("Failed to calculate contract overdue")
			continue
		}
		totalDebt += debt
	}

	totalDebtUSDMillicent, err := svc.gridClient.FromTFTtoUSDMillicent(uint64(totalDebt))
	if err != nil {
		return 0, fmt.Errorf("failed to convert debt to USD millicent: %w", err)
	}

	return totalDebtUSDMillicent, nil
}

// checkNodesWithWorkerPool uses a worker pool to check node health concurrently
func (svc WorkerService) CheckNodesWithWorkerPool(reservedNodes []models.UserNodes) {
	timeout := time.Duration(svc.reservedNodeHealthCheckTimeoutInMinutes) * time.Minute

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	workerCount := svc.reservedNodeHealthCheckWorkersNum
	if workerCount > len(reservedNodes) {
		workerCount = len(reservedNodes)
	}

	jobs := make(chan models.UserNodes, len(reservedNodes))
	results := make(chan NodeHealthResult, len(reservedNodes))

	var wg sync.WaitGroup
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go svc.healthCheckWorker(ctx, &wg, jobs, results)
	}

	go func() {
		defer close(jobs)
		for _, userNode := range reservedNodes {
			select {
			case <-ctx.Done():
				logger.ForOperation("health_tracker", "health_check_worker").Info().Msg("Context done, stopping health check worker")
				return
			case jobs <- userNode:
			}
		}
	}()

	go func() {
		wg.Wait()
		close(results)
	}()

	userNodes := make(map[int][]uint32)
	for res := range results {
		userNodes[res.userID] = append(userNodes[res.userID], res.unhealthyNodeID)
	}

	for userID, nodeIDs := range userNodes {
		if len(nodeIDs) == 0 {
			continue
		}

		var nodesList string
		for i, id := range nodeIDs {
			if i > 0 {
				nodesList += "\n"
			}
			nodesList += fmt.Sprintf("Node ID: %d", id)
		}

		notif := notification.NewNotification(userID, models.NotificationTypeNode).
			Warning(fmt.Sprintf("You have %d reserved node(s) that are currently unhealthy", len(nodeIDs))).
			WithSubject("Reserved Node Health Check Failed").
			WithStatus("unhealthy").
			WithChannels(notification.ChannelEmail).
			WithExtra("unhealthy_count", fmt.Sprintf("%d", len(nodeIDs))).
			WithExtra("nodes_list", nodesList).
			Build()

		if err := svc.notificationDispatcher.Send(ctx, notif); err != nil {
			logger.ForOperation("health_tracker", "send_unhealthy_nodes_notification").Error().Err(err).Msg("Failed to send unhealthy nodes notification")
		}
	}
}

func (svc WorkerService) healthCheckWorker(ctx context.Context, wg *sync.WaitGroup, jobs <-chan models.UserNodes, results chan<- NodeHealthResult) {
	defer wg.Done()
	log := logger.ForOperation("health_tracker", "health_check_worker")

	for userNode := range jobs {

		node, err := svc.gridClient.Node(ctx, userNode.NodeID)
		if err != nil {
			log.Error().Err(err).Uint32("node_id", userNode.NodeID).Msg("failed to get node for health check")
			continue
		}

		if node.Rentable {
			continue
		}

		if node.Healthy {
			continue
		}

		results <- NodeHealthResult{
			userID:          userNode.UserID,
			unhealthyNodeID: userNode.NodeID,
		}
	}
}

func (svc WorkerService) SettlePendingPayments(records []models.PendingRecord) {
	log := logger.ForOperation("balance_monitor", "settle_pending_payments")

	for _, record := range records {
		// Already settled
		if record.TransferredTFTAmount >= record.TFTAmount {
			continue
		}

		// getting balance every time to ensure we have the latest balance
		systemTFTBalance, err := svc.gridClient.GetFreeBalanceTFT(svc.systemMnemonic)
		if err != nil {
			log.Error().Err(err).Int("record_id", record.ID).Msg("Failed to get system TFT balance for pending record")
			continue
		}

		amountToTransfer := record.TFTAmount - record.TransferredTFTAmount
		if systemTFTBalance < amountToTransfer {
			log.Warn().
				Int("record_id", record.ID).
				Uint64("system_balance", systemTFTBalance).
				Uint64("amount_needed", amountToTransfer).
				Msg("Insufficient system balance to settle pending record")
			continue
		}

		if err = svc.transferTFTsToUser(record.UserID, record.ID, amountToTransfer); err != nil {
			log.Error().Err(err).Int("user_id", record.UserID).Int("record_id", record.ID).Msg("Failed to transfer TFTs to user")
			continue
		}
	}
}

func (svc WorkerService) transferTFTsToUser(userID, recordID int, amountToTransfer uint64) error {
	user, err := svc.userRepo.GetUserByID(userID)
	if err != nil {
		return fmt.Errorf("failed to get user for pending record ID %d: %w", recordID, err)
	}

	err = svc.gridClient.TransferTFTsFromSystem(amountToTransfer, user.Mnemonic)
	if err != nil {
		return fmt.Errorf("failed to transfer TFTs for pending record ID %d: %w", recordID, err)
	}

	err = svc.pendingRecordsRepo.UpdatePendingRecordTransferredAmount(recordID, amountToTransfer)
	if err != nil {
		return fmt.Errorf("failed to update transferred amount for pending record ID %d: %w", recordID, err)
	}

	return nil
}

func (svc WorkerService) NotifyAdminWithPendingRecords(records []models.PendingRecord) error {

	admins, err := svc.userRepo.ListAdmins()
	if err != nil {
		return err
	}

	for _, admin := range admins {
		err = svc.mailService.SendNotifyAdminsEmail(admin.Email, len(records))
		if err != nil {
			logger.ForOperation("balance_monitor", "send_admin_mail").Error().Err(err).Msg("Failed to send admin notification email")
			continue
		}
	}

	return nil
}

func (svc WorkerService) AsyncTrackClusterHealth(cluster models.Cluster) error {
	wf, err := svc.ewfEngine.NewWorkflow(workflows.WorkflowTrackClusterHealth, ewf.WithDisplayName("Cluster Health Check"))
	if err != nil {
		return fmt.Errorf("failed to create health tracking workflow: %w", err)
	}

	cl, err := cluster.GetClusterResult()
	if err != nil {
		return fmt.Errorf("failed to get cluster result during health tracking: %w", err)
	}

	wf.State = ewf.State{
		"cluster": cl,
		"config": map[string]interface{}{
			"user_id": cluster.UserID,
		},
	}

	return svc.ewfEngine.Run(svc.ctx, wf, ewf.WithAsync())
}

func (svc WorkerService) GetAllWorkflowsLocks() ([]string, error) {
	return svc.locker.GetAllWorkflowsLocks(svc.ctx)
}

func (svc WorkerService) ReleaseLocks(keys []string) {
	log := logger.ForOperation("locks_worker", "release_locks")
	workflowsNodes := map[string][]uint32{}
	for _, key := range keys {
		parts := strings.Split(key, ":")
		if len(parts) != 3 {
			log.Error().Str("key", key).Msg("invalid lock key format")
			continue
		}

		workflowID := parts[2]
		nodeID, err := strconv.ParseUint(parts[1], 10, 32)
		if err != nil {
			log.Error().Str("key", key).Msg("invalid node ID")
			continue
		}
		workflowsNodes[workflowID] = append(workflowsNodes[workflowID], uint32(nodeID))
	}

	for workflowID := range workflowsNodes {
		workflow, err := svc.ewfEngine.Store().LoadWorkflowByUUID(svc.ctx, workflowID)
		if err != nil {
			log.Error().Str("workflow_id", workflowID).Msg("failed to load workflow")
			continue
		}
		if !slices.Contains([]ewf.WorkflowStatus{ewf.StatusCompleted, ewf.StatusFailed}, workflow.Status) {
			continue
		}
		nodeIDs := workflowsNodes[workflowID]
		if err := svc.locker.ReleaseLock(svc.ctx, nodeIDs, workflowID); err != nil {
			log.Error().Str("workflow_id", workflow.UUID).Msg("failed to release locks")
			continue
		}
	}

}

func (svc WorkerService) checkUserDebt(user models.User, contractIDs []uint64) error {
	totalDebt, err := svc.calculateDebt(user.Mnemonic, contractIDs, svc.GetCheckUserDebtInterval())
	if err != nil {
		return fmt.Errorf("failed to calculate debt: %w", err)
	}
	userBalance, err := svc.gridClient.GetUserBalanceUSDMillicent(user.Mnemonic)
	if err != nil {
		return fmt.Errorf("failed to get user balance: %w", err)
	}
	if userBalance >= totalDebt {
		return nil
	}
	days := int(svc.GetCheckUserDebtInterval() / (24 * time.Hour))
	if days == 0 {
		days = 1
	}
	totalDebtUSD := gridclient.FromUSDMilliCentToUSD(totalDebt)
	userBalanceUSD := gridclient.FromUSDMilliCentToUSD(userBalance)

	message := fmt.Sprintf(
		"Your balance is not enough to cover the debt for upcoming %d day(s).\nTotal debt: $%.2f\nUser balance: $%.2f",
		days, totalDebtUSD, userBalanceUSD,
	)
	notif := notification.NewNotification(user.ID, models.NotificationTypeBilling).
		Warning(message).
		WithSubject("User Balance Not Enough").
		WithChannels(notification.ChannelEmail, notification.ChannelUI).
		WithExtra("user_id", fmt.Sprintf("%d", user.ID)).
		Build()
	if err := svc.notificationDispatcher.Send(svc.ctx, notif); err != nil {
		return fmt.Errorf("failed to send notification: %w", err)
	}

	return nil
}

func (svc WorkerService) CheckUsersBalance() error {
	userContractIDs, err := svc.getUserContractIDs()
	if err != nil {
		return fmt.Errorf("failed to build contract IDs: %w", err)
	}
	users, err := svc.userRepo.ListAllUsers()
	if err != nil {
		return fmt.Errorf("failed to list users: %w", err)
	}
	multiErr := &multierror.Error{}
	var multiErrMu sync.Mutex
	var wg sync.WaitGroup
	balanceCheckLimiter := make(chan struct{}, svc.mailService.GetMailConfig().MaxConcurrentSends)
	for _, user := range users {
		if _, ok := userContractIDs[user.ID]; !ok {
			continue
		}
		wg.Add(1)
		balanceCheckLimiter <- struct{}{}

		go func(user models.User, contractIDs []uint64) {
			defer wg.Done()
			defer func() { <-balanceCheckLimiter }()
			select {
			case <-svc.ctx.Done():
				return
			default:
			}

			err := svc.checkUserDebt(user, contractIDs)
			if err == nil {
				return
			}
			multiErrMu.Lock()
			multiErr = multierror.Append(multiErr, err)
			multiErrMu.Unlock()
		}(user, userContractIDs[user.ID])
	}
	wg.Wait()
	if err := multiErr.ErrorOrNil(); err != nil {
		return fmt.Errorf("failed to check users balance: %w", err)
	}
	return nil
}

func (svc WorkerService) getUserContractIDs() (map[int][]uint64, error) {
	clusters, err := svc.clusterRepo.ListAllClusters()
	if err != nil {
		return nil, fmt.Errorf("failed to list clusters: %w", err)
	}

	reservedNodes, err := svc.nodesRepo.ListAllReservedNodes()
	if err != nil {
		return nil, fmt.Errorf("failed to list reserved nodes: %w", err)
	}

	userContractIDs := make(map[int][]uint64)
	for _, cluster := range clusters {
		clusterResult, err := cluster.GetClusterResult()
		if err != nil {
			logger.ForOperation("balance_monitor", "build_contract_ids").
				Error().Err(err).Int("user_id", cluster.UserID).Msg("failed to get cluster result, skipping cluster")
			continue
		}
		for _, node := range clusterResult.Nodes {
			if node.ContractID == 0 {
				continue
			}
			userContractIDs[cluster.UserID] = append(userContractIDs[cluster.UserID], node.ContractID)
		}
	}

	for _, reservedNode := range reservedNodes {
		userContractIDs[reservedNode.UserID] = append(userContractIDs[reservedNode.UserID], reservedNode.ContractID)
	}
	return userContractIDs, nil

}

func getHoursOfGivenPeriod(startDate, endDate time.Time) int {
	// Calculate the duration between the first day of the month and the specific date
	duration := endDate.Sub(startDate)
	// Convert the duration to hours
	hours := int(duration.Hours())
	return hours
}
