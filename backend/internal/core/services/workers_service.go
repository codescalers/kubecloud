package services

import (
	"context"
	"errors"
	"fmt"
	"kubecloud/internal/billing"
	cfg "kubecloud/internal/config"
	"kubecloud/internal/core/models"
	"kubecloud/internal/core/workflows"
	"kubecloud/internal/infrastructure/logger"
	"kubecloud/internal/infrastructure/mailservice"
	"kubecloud/internal/infrastructure/notification"
	"kubecloud/internal/infrastructure/substrate"

	"sync"
	"time"

	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/calculator"
	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/deployer"
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
	gridClient             deployer.TFPluginClient
	ewfEngine              *ewf.Engine
	notificationDispatcher *notification.NotificationDispatcher

	// configs
	systemMnemonic                          string
	invoiceCompanyData                      cfg.InvoiceCompanyData
	currency                                string
	clusterHealthCheckIntervalInHours       int
	reservedNodeHealthCheckIntervalInHours  int
	reservedNodeHealthCheckTimeoutInMinutes int
	reservedNodeHealthCheckWorkersNum       int
	monitorBalanceIntervalInMinutes         int
	notifyAdminsForPendingRecordsInHours    int
}

func NewWorkersService(
	ctx context.Context, userRepo models.UserRepository, nodesRepo models.UserNodesRepository,
	invoicesRepo models.InvoiceRepository, clusterRepo models.ClusterRepository, pendingRecordsRepo models.PendingRecordRepository,
	mailService mailservice.MailService,
	gridClient deployer.TFPluginClient, ewfEngine *ewf.Engine, notificationDispatcher *notification.NotificationDispatcher,
	graphql graphql.GraphQl, firesquidClient graphql.GraphQl,
	invoiceCompanyData cfg.InvoiceCompanyData, systemMnemonic, currency string,
	clusterHealthCheckIntervalInHours, reservedNodeHealthCheckIntervalInHours,
	reservedNodeHealthCheckTimeoutInMinutes, reservedNodeHealthCheckWorkersNum,
	monitorBalanceIntervalInMinutes, notifyAdminsForPendingRecordsInHours int,
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

		systemMnemonic:     systemMnemonic,
		invoiceCompanyData: invoiceCompanyData,
		currency:           currency,

		clusterHealthCheckIntervalInHours:       clusterHealthCheckIntervalInHours,
		reservedNodeHealthCheckIntervalInHours:  reservedNodeHealthCheckIntervalInHours,
		reservedNodeHealthCheckTimeoutInMinutes: reservedNodeHealthCheckTimeoutInMinutes,
		reservedNodeHealthCheckWorkersNum:       reservedNodeHealthCheckWorkersNum,
		monitorBalanceIntervalInMinutes:         monitorBalanceIntervalInMinutes,
		notifyAdminsForPendingRecordsInHours:    notifyAdminsForPendingRecordsInHours,
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

func (svc WorkerService) GetMonitorBalanceInterval() time.Duration {
	return time.Duration(svc.monitorBalanceIntervalInMinutes) * time.Minute
}

func (svc WorkerService) GetNotifyAdminsForPendingRecordsInterval() time.Duration {
	return time.Duration(svc.notifyAdminsForPendingRecordsInHours) * time.Hour
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
		totalAmountUSDMillicent, err := svc.gridClient.SubstrateConn.FromTFTtoUSDMillicent(totalAmountTFT)
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

		totalAmountUSD := substrate.FromUSDMilliCentToUSD(totalAmountUSDMillicent)

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

	subject, body := svc.mailService.InvoiceMailContent(totalInvoiceCostUSD, svc.currency, invoice.ID)
	return svc.mailService.SendMailFromSystem(user.Email, subject, body, mailservice.Attachment{
		FileName: fmt.Sprintf("invoice-%d-%d.pdf", invoice.UserID, invoice.ID),
		Data:     invoice.FileData,
	})
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

		userDebt, err := svc.calculateDebt(user.Mnemonic, userNodes)
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

func (svc WorkerService) calculateDebt(userMnemonic string, userNodes []models.UserNodes) (uint64, error) {
	identity, err := svc.gridClient.SubstrateConn.NewIdentityFromSr25519Phrase(userMnemonic)
	if err != nil {
		return 0, fmt.Errorf("failed to create identity: %w", err)
	}

	var totalDebt int64
	for _, node := range userNodes {
		calculatorClient := calculator.NewCalculator(svc.gridClient.SubstrateConn, identity)
		debt, err := calculatorClient.CalculateContractOverdue(node.ContractID, time.Hour)
		if err != nil {
			logger.ForOperation("debt_tracker", "calc_overdue").Error().Err(err).Msg("Failed to calculate contract overdue")
			continue
		}
		totalDebt += debt
	}

	totalDebtUSDMillicent, err := svc.gridClient.SubstrateConn.FromTFTtoUSDMillicent(uint64(totalDebt))
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

		node, err := svc.gridClient.GridProxyClient.Node(ctx, userNode.NodeID)
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
		systemTFTBalance, err := svc.gridClient.SubstrateConn.GetUserTFTBalance(svc.systemMnemonic)
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

	err = svc.gridClient.SubstrateConn.TransferTFTsFromSystem(amountToTransfer, user.Mnemonic, svc.systemMnemonic)
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
	subject, body := svc.mailService.NotifyAdminsMailContent(len(records))

	admins, err := svc.userRepo.ListAdmins()
	if err != nil {
		return err
	}

	for _, admin := range admins {
		err = svc.mailService.SendMailFromSystem(admin.Email, subject, body)
		if err != nil {
			logger.ForOperation("balance_monitor", "send_admin_mail").Error().Err(err).Msg("Failed to send admin notification email")
			continue
		}
	}

	return nil
}

func (svc WorkerService) AsyncTrackClusterHealth(cluster models.Cluster) error {
	wf, err := svc.ewfEngine.NewWorkflow(workflows.WorkflowTrackClusterHealth)
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

func getHoursOfGivenPeriod(startDate, endDate time.Time) int {
	// Calculate the duration between the first day of the month and the specific date
	duration := endDate.Sub(startDate)
	// Convert the duration to hours
	hours := int(duration.Hours())
	return hours
}
