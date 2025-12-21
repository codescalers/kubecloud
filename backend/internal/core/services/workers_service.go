package services

import (
	"context"
	"errors"
	"fmt"
	"kubecloud/internal/billing"
	"kubecloud/internal/config"
	"kubecloud/internal/core/models"
	"kubecloud/internal/core/workflows"
	"kubecloud/internal/infrastructure/gridclient"
	"kubecloud/internal/infrastructure/logger"
	"kubecloud/internal/infrastructure/mailservice"
	mailsender "kubecloud/internal/infrastructure/mailservice/mail_sender"
	"kubecloud/internal/infrastructure/notification"

	"sync"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/hashicorp/go-multierror"
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
	gridClient             gridclient.GridClient
	ewfEngine              *ewf.Engine
	notificationDispatcher *notification.NotificationDispatcher
	billingService         BillingService

	// configs
	invoiceCompanyData                      config.InvoiceCompanyData
	currency                                string
	checkUserDebtIntervalInHours            int
	clusterHealthCheckIntervalInHours       int
	reservedNodeHealthCheckIntervalInHours  int
	reservedNodeHealthCheckTimeoutInMinutes int
	reservedNodeHealthCheckWorkersNum       int
	settleTransferRecordsIntervalInMinutes  int
	notifyAdminsForPendingRecordsInHours    int
	minimumTFTAmountInWallet                int
	usersBalanceCheckIntervalInHours        int
}

func NewWorkersService(
	ctx context.Context, userRepo models.UserRepository, contractsRepo models.ContractDataRepository,
	invoicesRepo models.InvoiceRepository, clusterRepo models.ClusterRepository, transferRecordsRepo models.TransferRecordRepository,
	mailService mailservice.MailService,
	gridClient gridclient.GridClient, ewfEngine *ewf.Engine, notificationDispatcher *notification.NotificationDispatcher,
	graphql graphql.GraphQl, firesquidClient graphql.GraphQl, billingService BillingService,
	invoiceCompanyData config.InvoiceCompanyData, currency string,
	clusterHealthCheckIntervalInHours, reservedNodeHealthCheckIntervalInHours,
	reservedNodeHealthCheckTimeoutInMinutes, reservedNodeHealthCheckWorkersNum,
	settleTransferRecordsIntervalInMinutes, notifyAdminsForPendingRecordsInHours,
	minimumTFTAmountInWallet,
	usersBalanceCheckIntervalInHours,
	checkUserDebtIntervalInHours int,
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
		gridClient:             gridClient,
		billingService:         billingService,

		invoiceCompanyData: invoiceCompanyData,
		currency:           currency,

		checkUserDebtIntervalInHours:            checkUserDebtIntervalInHours,
		clusterHealthCheckIntervalInHours:       clusterHealthCheckIntervalInHours,
		reservedNodeHealthCheckIntervalInHours:  reservedNodeHealthCheckIntervalInHours,
		reservedNodeHealthCheckTimeoutInMinutes: reservedNodeHealthCheckTimeoutInMinutes,
		reservedNodeHealthCheckWorkersNum:       reservedNodeHealthCheckWorkersNum,
		settleTransferRecordsIntervalInMinutes:  settleTransferRecordsIntervalInMinutes,
		notifyAdminsForPendingRecordsInHours:    notifyAdminsForPendingRecordsInHours,

		minimumTFTAmountInWallet:         minimumTFTAmountInWallet,
		usersBalanceCheckIntervalInHours: usersBalanceCheckIntervalInHours,
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

func (svc WorkerService) GetCheckUserDebtInterval() time.Duration {
	return time.Duration(svc.checkUserDebtIntervalInHours) * time.Hour
}

func (svc WorkerService) GetNotifyAdminsForPendingRecordsInterval() time.Duration {
	return time.Duration(svc.notifyAdminsForPendingRecordsInHours) * time.Hour
}

func (svc WorkerService) GetUsersBalanceCheckInterval() time.Duration {
	return time.Duration(svc.usersBalanceCheckIntervalInHours) * time.Hour
}

func (svc WorkerService) CreateUserInvoice(user models.User) error {
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

		totalAmountBilledInUSDMillicent, err := svc.billingService.calculateTotalUsageOfReportsInUSDMillicent(billReports.Reports)
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

		totalAmountUSD := gridclient.FromUSDMilliCentToUSD(totalAmountBilledInUSDMillicent)

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
		userContracts, err := svc.contractsRepo.ListAllContractsInPeriod(user.ID, time.Now().Add(-TrackingDebtPeriod), time.Now())
		if err != nil {
			logger.ForOperation("debt_tracker", "list_user_contracts").Error().Err(err).Msg("Failed to list user contracts")
			continue
		}

		var contractIDs []uint64
		for _, contract := range userContracts {
			contractIDs = append(contractIDs, contract.ContractID)
		}

		userDebt, err := svc.calculateDebt(user.Mnemonic, contractIDs)
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

func (svc WorkerService) calculateDebt(userMnemonic string, contractIDs []uint64) (uint64, error) {
	if len(contractIDs) == 0 {
		return 0, nil
	}

	calculatorClient, err := svc.gridClient.NewCalculator(userMnemonic)
	if err != nil {
		return 0, fmt.Errorf("failed to create calculator: %w", err)
	}

	var totalDebt int64
	for _, contractID := range contractIDs {
		var debt int64
		if err = backoff.Retry(func() error {
			debt, err = calculatorClient.CalculateContractOverdue(contractID, time.Hour)
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

	totalDebtUSDMillicent, err := svc.gridClient.FromTFTtoUSDMillicent(uint64(totalDebt))
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
		systemTFTBalance, err := svc.gridClient.GetSystemTFTBalance()
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

	err = svc.gridClient.TransferTFTsFromSystem(amountToTransfer, user.Mnemonic)
	if err != nil {
		return fmt.Errorf("failed to transfer TFTs for user %d: %w", userID, err)
	}

	return nil
}

func (svc *WorkerService) ResetUsersTFTsWithNoUSDBalance(users []models.User) error {
	for _, user := range users {
		if user.CreditedBalance+user.CreditCardBalance-user.Debt <= 0 {
			logger.GetLogger().Info().Msgf("User %d has no USD balance, withdrawing all TFTs except for %d", user.ID, svc.minimumTFTAmountInWallet)

			userTFTBalance, err := svc.gridClient.GetFreeBalanceTFT(user.Mnemonic)
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

			if err = svc.gridClient.TransferTFTsToSystem(userTFTBalance, user.Mnemonic); err != nil {
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

func (svc WorkerService) NotifyAdminWithInsufficientBalance() error {
	currentBalance, err := svc.gridClient.GetSystemTFTBalance()
	if err != nil {
		return err
	}

	users, err := svc.userRepo.ListAllUsers()
	if err != nil {
		return err
	}

	var totalUserDailyUsage uint64

	for _, user := range users {
		dailyUsageInUSDMillicent, err := svc.billingService.calculateResourcesUsageInUSD(svc.ctx, user.ID, user.Mnemonic, nil, nil)
		if err != nil {
			return err
		}

		dailyUsageInTFT, err := svc.gridClient.FromUSDMillicentToTFT(dailyUsageInUSDMillicent)
		if err != nil {
			return err
		}

		totalUserDailyUsage += dailyUsageInTFT
	}

	requiredBalance, err := svc.billingService.ApplyDiscountOnUsage(totalUserDailyUsage)
	if err != nil {
		return err
	}

	if requiredBalance <= currentBalance {
		return nil
	}

	admins, err := svc.userRepo.ListAdmins()
	if err != nil {
		return err
	}

	for _, admin := range admins {
		err = svc.mailService.SendInsufficientBalanceNotificationEmail(
			admin.Email, float64(currentBalance)/TFTUnitFactor,
			float64(requiredBalance)/TFTUnitFactor, fmt.Sprint(svc.billingService.Discount()),
		)
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

func (svc WorkerService) checkUserDebt(user models.User, contractIDs []uint64) error {
	totalDebt, err := svc.calculateDebt(user.Mnemonic, contractIDs)
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

	reservedNodes, err := svc.contractsRepo.ListAllReservedNodes()
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
