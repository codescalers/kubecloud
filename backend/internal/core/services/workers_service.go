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

	"github.com/cenkalti/backoff"
	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/calculator"
	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/deployer"
	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/graphql"
	"github.com/xmonader/ewf"
)

const (
	// UnitFactor represents the smallest unit conversion factor for both USD and TFT
	TFTUnitFactor = 1e7
	transferFees  = 0.01 * TFTUnitFactor // 0.01 TFT
	nodeCertified = "Certified"

	zeroTFTBalanceValue    = 0.05 * TFTUnitFactor // 0.05 TFT
	defaultPricingPolicyID = uint32(1)

	TrackingDebtPeriod = time.Hour
	reties             = 3
)

type WorkerService struct {
	ctx context.Context

	userRepo            models.UserRepository
	contractsRepo       models.ContractDataRepository
	invoicesRepo        models.InvoiceRepository
	clusterRepo         models.ClusterRepository
	transferRecordsRepo models.TransferRecordRepository

	mailService            mailservice.MailService
	graphql                graphql.GraphQl
	firesquidClient        graphql.GraphQl
	substrateClient        substrate.Substrate
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
	settleTransferRecordsIntervalInMinutes  int
	notifyAdminsForPendingRecordsInHours    int
	minimumTFTAmountInWallet                int
	appliedDiscount                         Discount
}

func NewWorkersService(
	ctx context.Context, userRepo models.UserRepository, contractsRepo models.ContractDataRepository,
	invoicesRepo models.InvoiceRepository, clusterRepo models.ClusterRepository, transferRecordsRepo models.TransferRecordRepository,
	mailService mailservice.MailService,
	gridClient deployer.TFPluginClient, ewfEngine *ewf.Engine, notificationDispatcher *notification.NotificationDispatcher,
	graphql graphql.GraphQl, firesquidClient graphql.GraphQl, substrateClient substrate.Substrate,
	invoiceCompanyData cfg.InvoiceCompanyData, systemMnemonic, currency string,
	clusterHealthCheckIntervalInHours, reservedNodeHealthCheckIntervalInHours,
	reservedNodeHealthCheckTimeoutInMinutes, reservedNodeHealthCheckWorkersNum,
	settleTransferRecordsIntervalInMinutes, notifyAdminsForPendingRecordsInHours int,
	minimumTFTAmountInWallet int, appliedDiscount Discount,
) WorkerService {
	return WorkerService{
		ctx:                 ctx,
		userRepo:            userRepo,
		contractsRepo:       contractsRepo,
		invoicesRepo:        invoicesRepo,
		clusterRepo:         clusterRepo,
		transferRecordsRepo: transferRecordsRepo,

		mailService:            mailService,
		notificationDispatcher: notificationDispatcher,
		ewfEngine:              ewfEngine,
		graphql:                graphql,
		firesquidClient:        firesquidClient,
		substrateClient:        substrateClient,
		gridClient:             gridClient,

		systemMnemonic:     systemMnemonic,
		invoiceCompanyData: invoiceCompanyData,
		currency:           currency,

		clusterHealthCheckIntervalInHours:       clusterHealthCheckIntervalInHours,
		reservedNodeHealthCheckIntervalInHours:  reservedNodeHealthCheckIntervalInHours,
		reservedNodeHealthCheckTimeoutInMinutes: reservedNodeHealthCheckTimeoutInMinutes,
		reservedNodeHealthCheckWorkersNum:       reservedNodeHealthCheckWorkersNum,
		settleTransferRecordsIntervalInMinutes:  settleTransferRecordsIntervalInMinutes,
		notifyAdminsForPendingRecordsInHours:    notifyAdminsForPendingRecordsInHours,

		minimumTFTAmountInWallet: minimumTFTAmountInWallet,
		appliedDiscount:          appliedDiscount,
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

func (svc WorkerService) ListUserClusters(userID int) ([]models.Cluster, error) {
	return svc.clusterRepo.ListUserClusters(userID)
}

func (svc WorkerService) ListAllReservedNodes() ([]models.UserContractData, error) {
	return svc.contractsRepo.ListAllReservedNodes()
}

func (svc WorkerService) ListFailedTransferRecords() ([]models.TransferRecord, error) {
	return svc.transferRecordsRepo.ListFailedTransferRecords()
}

func (svc WorkerService) ListPendingTransferRecords() ([]models.TransferRecord, error) {
	return svc.transferRecordsRepo.ListPendingTransferRecords()
}

func (svc WorkerService) GetClusterHealthCheckInterval() time.Duration {
	return time.Duration(svc.clusterHealthCheckIntervalInHours) * time.Hour
}

func (svc WorkerService) GetReservedNodeHealthCheckInterval() time.Duration {
	return time.Duration(svc.reservedNodeHealthCheckIntervalInHours) * time.Hour
}

func (svc WorkerService) GetSettleTransferRecordsInterval() time.Duration {
	return time.Duration(svc.settleTransferRecordsIntervalInMinutes) * time.Minute
}

func (svc WorkerService) GetNotifyAdminsForPendingRecordsInterval() time.Duration {
	return time.Duration(svc.notifyAdminsForPendingRecordsInHours) * time.Hour
}

func (svc WorkerService) CreateUserInvoice(BillingService BillingService, user models.User) error {
	now := time.Now()
	timeMonthAgo := now.AddDate(0, -1, 0)

	contracts, err := svc.contractsRepo.ListAllContractsInPeriod(user.ID, timeMonthAgo, now)
	if err != nil {
		return err
	}

	if len(contracts) == 0 {
		return nil
	}

	var invoiceItems []models.NodeItem
	var totalInvoiceCostUSD float64

	for _, contract := range contracts {
		billReports, err := billing.ListContractBillReports(svc.graphql, contract.ContractID, timeMonthAgo, now)
		if err != nil {
			return err
		}

		totalAmountBilledInUSDMillicent, err := BillingService.calculateTotalUsageOfReportsInUSDMillicent(billReports.Reports)
		if err != nil {
			return err
		}

		rentRecordStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		if contract.CreatedAt.After(rentRecordStart) {
			rentRecordStart = contract.CreatedAt
		}

		var totalHours int
		cancellationDate, err := billing.GetRentContractCancellationDate(svc.firesquidClient, contract.ContractID)
		if errors.Is(err, billing.ErrorEventsNotFound) {
			totalHours = getHoursOfGivenPeriod(rentRecordStart, time.Now())
		} else if err != nil {
			return err
		} else {
			totalHours = getHoursOfGivenPeriod(rentRecordStart, cancellationDate)
		}

		totalAmountUSD := substrate.FromUSDMilliCentToUSD(totalAmountBilledInUSDMillicent)

		invoiceItems = append(invoiceItems, models.NodeItem{
			NodeID:        contract.NodeID,
			ContractID:    contract.ContractID,
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
		userContracts, err := svc.contractsRepo.ListAllContractsInPeriod(user.ID, time.Now().Add(-TrackingDebtPeriod), time.Now())
		if err != nil {
			logger.ForOperation("debt_tracker", "list_user_contracts").Error().Err(err).Msg("Failed to list user contracts")
			continue
		}

		userDebt, err := svc.calculateDebt(user.Mnemonic, userContracts)
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

func (svc WorkerService) calculateDebt(userMnemonic string, userContracts []models.UserContractData) (uint64, error) {
	if len(userContracts) == 0 {
		return 0, nil
	}

	identity, err := svc.substrateClient.NewIdentityFromSr25519Phrase(userMnemonic)
	if err != nil {
		return 0, fmt.Errorf("failed to create identity: %w", err)
	}

	calculatorClient := calculator.NewCalculator(svc.gridClient.SubstrateConn, identity)

	var totalDebt int64
	for _, userContract := range userContracts {
		var debt int64
		if err = backoff.Retry(func() error {
			debt, err = calculatorClient.CalculateContractOverdue(userContract.ContractID, time.Hour)
			return err
		}, backoff.WithMaxRetries(
			backoff.NewExponentialBackOff(),
			reties,
		)); err != nil {
			logger.ForOperation("debt_tracker", "calc_overdue").Error().Err(err).Msg("Failed to calculate contract overdue")
			continue
		}

		totalDebt += debt
	}

	totalDebtUSDMillicent, err := svc.substrateClient.FromTFTtoUSDMillicent(uint64(totalDebt))
	if err != nil {
		return 0, fmt.Errorf("failed to convert debt to USD millicent: %w", err)
	}

	return totalDebtUSDMillicent, nil
}

// checkNodesWithWorkerPool uses a worker pool to check node health concurrently
func (svc WorkerService) CheckNodesWithWorkerPool(reservedNodes []models.UserContractData) {
	timeout := time.Duration(svc.reservedNodeHealthCheckTimeoutInMinutes) * time.Minute

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	workerCount := svc.reservedNodeHealthCheckWorkersNum
	if workerCount > len(reservedNodes) {
		workerCount = len(reservedNodes)
	}

	jobs := make(chan models.UserContractData, len(reservedNodes))
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

func (svc WorkerService) healthCheckWorker(ctx context.Context, wg *sync.WaitGroup, jobs <-chan models.UserContractData, results chan<- NodeHealthResult) {
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

func (svc WorkerService) SettlePendingPayments(records []models.TransferRecord) {
	log := logger.ForOperation("balance_monitor", "settle_pending_payments")

	for _, record := range records {
		if record.Operation == models.WithdrawOperation {
			continue
		}

		// Already settled
		if record.State == models.SuccessState {
			continue
		}

		transferState := models.SuccessState
		var transferFailure string

		// getting balance every time to ensure we have the latest balance
		systemTFTBalance, err := svc.substrateClient.GetUserTFTBalance(svc.systemMnemonic)
		if err != nil {
			log.Error().Err(err).Int("record_id", record.ID).Msg("Failed to get system TFT balance for pending record")
			continue
		}

		if systemTFTBalance < record.TFTAmount {
			logger.GetLogger().Warn().Msgf("Insufficient system balance to settle pending record ID %d", record.ID)
			continue
		}

		if err = svc.transferTFTsToUser(record.UserID, record.TFTAmount); err != nil {
			logger.GetLogger().Error().Err(err).Msgf("Failed to settle pending record ID %d", record.ID)

			transferState = models.FailedState
			transferFailure = err.Error()
		}

		if err := svc.transferRecordsRepo.UpdateTransferRecordState(record.ID, transferState, transferFailure); err != nil {
			logger.GetLogger().Error().Err(err).Msgf("Failed to update pending record ID %d state", record.ID)
		}
	}
}

func (svc WorkerService) transferTFTsToUser(userID int, amountToTransfer uint64) error {
	user, err := svc.userRepo.GetUserByID(userID)
	if err != nil {
		return fmt.Errorf("failed to get user %d: %w", userID, err)
	}

	err = svc.substrateClient.TransferTFTsFromSystem(amountToTransfer, user.Mnemonic)
	if err != nil {
		return fmt.Errorf("failed to transfer TFTs for user %d: %w", userID, err)
	}

	return nil
}

func (svc *WorkerService) ResetUsersTFTsWithNoUSDBalance(users []models.User) error {
	for _, user := range users {
		if user.CreditedBalance+user.CreditCardBalance-user.Debt <= 0 {
			logger.GetLogger().Info().Msgf("User %d has no USD balance, withdrawing all TFTs except for %d", user.ID, svc.minimumTFTAmountInWallet)

			userTFTBalance, err := svc.substrateClient.GetUserTFTBalance(user.Mnemonic)
			if err != nil {
				logger.GetLogger().Error().Err(err).Msgf("Failed to get user TFT balance for user %d", user.ID)
				continue
			}

			if userTFTBalance <= uint64(svc.minimumTFTAmountInWallet)*TFTUnitFactor {
				continue
			}

			if userTFTBalance <= uint64(svc.minimumTFTAmountInWallet)*TFTUnitFactor+transferFees {
				continue
			}

			transferRecord := models.TransferRecord{
				UserID:    user.ID,
				Username:  user.Username,
				TFTAmount: userTFTBalance - transferFees - uint64(svc.minimumTFTAmountInWallet)*TFTUnitFactor,
				Operation: models.WithdrawOperation,
				State:     models.SuccessState,
			}

			if err = svc.substrateClient.TransferTFTsToSystem(userTFTBalance, user.Mnemonic); err != nil {
				logger.GetLogger().Error().Err(err).Msgf("Failed to withdraw all TFTs for user %d", user.ID)

				transferRecord.State = models.FailedState
				transferRecord.Failure = err.Error()
			}

			if err := svc.transferRecordsRepo.CreateTransferRecord(&transferRecord); err != nil {
				logger.GetLogger().Error().Err(err).Msgf("Failed to create transfer record for user %d", user.ID)
			}
		}
	}

	return nil
}

func (svc WorkerService) NotifyAdminWithPendingRecords(records []models.TransferRecord) error {
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
