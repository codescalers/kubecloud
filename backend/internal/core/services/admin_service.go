package services

import (
	"context"
	"fmt"
	"kubecloud/internal/core/generators"
	"kubecloud/internal/core/models"
	"kubecloud/internal/core/persistence"
	"kubecloud/internal/core/workflows"
	"kubecloud/internal/infrastructure/logger"
	"kubecloud/internal/infrastructure/substrate"

	"sync"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/xmonader/ewf"
)

type AdminService struct {
	userRepo    models.UserRepository
	nodesRepo   models.UserNodesRepository
	prRepo      models.PendingRecordRepository
	voucherRepo models.VoucherRepository
	transRepo   models.TransactionRepository

	appCtx          context.Context
	substrateClient substrate.Substrate
	ewfEngine       *ewf.Engine
}

func NewAdminService(appCtx context.Context,
	userRepo models.UserRepository,
	userNodeRepo models.UserNodesRepository,
	pendingRecordRepo models.PendingRecordRepository,
	voucherRepo models.VoucherRepository,
	transactionRepo models.TransactionRepository,
	substrateClient substrate.Substrate,
	ewfEngine *ewf.Engine,
) AdminService {
	return AdminService{
		userRepo:    userRepo,
		nodesRepo:   userNodeRepo,
		prRepo:      pendingRecordRepo,
		voucherRepo: voucherRepo,
		transRepo:   transactionRepo,

		appCtx:          appCtx,
		substrateClient: substrateClient,
		ewfEngine:       ewfEngine,
	}
}

const maxConcurrentBalanceFetches = 20

type UserWithUSDBalance struct {
	models.User
	Balance float64 `json:"balance"` // USD balance
}

type PendingRecordsWithUSDAmounts struct {
	models.PendingRecord
	USDAmount            float64 `json:"usd_amount"`
	TransferredUSDAmount float64 `json:"transferred_usd_amount"`
}

func (svc *AdminService) ListAllUsers() ([]models.User, error) {
	return svc.userRepo.ListAllUsers()
}

func (svc *AdminService) ListAllUsersIncludingUSDBalance() ([]UserWithUSDBalance, error) {
	users, err := svc.ListAllUsers()
	// Here is the only critical errors, not the balance related ones
	if err != nil {
		return nil, err
	}

	var (
		usersWithBalance []UserWithUSDBalance
		wg               sync.WaitGroup
		mu               sync.Mutex
		balanceErrors    *multierror.Error
	)

	balanceConcurrencyLimiter := make(chan struct{}, maxConcurrentBalanceFetches)

	for _, user := range users {
		wg.Add(1)
		balanceConcurrencyLimiter <- struct{}{}

		go func(user models.User) {
			defer wg.Done()
			defer func() { <-balanceConcurrencyLimiter }()

			balance, err := svc.substrateClient.GetUserBalanceUSD(user.Mnemonic)
			if err != nil {
				mu.Lock()
				balanceErrors = multierror.Append(balanceErrors, fmt.Errorf("failed to get balance for user %d: %w", user.ID, err))
				mu.Unlock()
				balance = 0.0
			}

			mu.Lock()
			usersWithBalance = append(usersWithBalance, UserWithUSDBalance{
				User:    user,
				Balance: balance,
			})
			mu.Unlock()
		}(user)
	}

	wg.Wait()

	// Log balance errors but continue - return users with 0 balance for unregistered or failed users
	if balanceErrors.ErrorOrNil() != nil {
		// Log but don't fail the request - users with missing accounts will have 0 balance
		logger.GetLogger().Warn().Err(balanceErrors.ErrorOrNil()).Msg("some users had balance fetch errors, returning them with 0 balance")
	}

	return usersWithBalance, nil
}

func (svc *AdminService) DeleteUserByID(userID int) error {
	return svc.userRepo.DeleteUserByID(userID)
}

func (svc *AdminService) AsyncCreditUserUSD(transaction *models.Transaction) error {
	if err := svc.transRepo.CreateTransaction(transaction); err != nil {
		return err
	}

	user, err := svc.userRepo.GetUserByID(transaction.UserID)
	if err != nil {
		return err
	}

	wf, err := svc.ewfEngine.NewWorkflow(workflows.WorkflowAdminCreditBalance)
	if err != nil {
		return err
	}

	wf.State = map[string]interface{}{
		"user_id":       transaction.UserID,
		"amount":        substrate.FromUSDToUSDMillicent(transaction.Amount),
		"mnemonic":      user.Mnemonic,
		"username":      user.Username,
		"transfer_mode": models.AdminCreditMode,
		"admin_id":      transaction.AdminID,
	}

	if err = persistence.SetStateUserID(&wf, transaction.AdminID); err != nil {
		return err
	}

	return svc.ewfEngine.Run(svc.appCtx, wf, ewf.WithAsync())
}

func (svc *AdminService) GenerateVouchers(count, expireAfterDays int, voucherValue float64) ([]models.Voucher, error) {
	var vouchers []models.Voucher

	for range count {
		voucher := models.Voucher{
			Code:      svc.generateVoucherWithTimestamp(),
			Value:     voucherValue,
			CreatedAt: time.Now(),
			ExpiresAt: time.Now().Add(time.Duration(expireAfterDays) * 24 * time.Hour),
		}

		if err := svc.voucherRepo.CreateVoucher(&voucher); err != nil {
			return nil, err
		}
		vouchers = append(vouchers, voucher)
	}

	return vouchers, nil
}

func (svc *AdminService) ListAllVouchers() ([]models.Voucher, error) {
	return svc.voucherRepo.ListAllVouchers()
}

func (svc *AdminService) ListAllPendingRecordsWithUSDAmounts() ([]PendingRecordsWithUSDAmounts, error) {
	pendingRecords, err := svc.prRepo.ListAllPendingRecords()
	if err != nil {
		return nil, err
	}

	var pendingRecordsWithUSDAmounts []PendingRecordsWithUSDAmounts
	for _, record := range pendingRecords {
		usdAmount, err := svc.substrateClient.FromTFTtoUSDMillicent(record.TFTAmount)
		if err != nil {
			return nil, err
		}

		usdTransferredAmount, err := svc.substrateClient.FromTFTtoUSDMillicent(record.TransferredTFTAmount)
		if err != nil {
			return nil, err
		}

		pendingRecordsWithUSDAmounts = append(pendingRecordsWithUSDAmounts, PendingRecordsWithUSDAmounts{
			PendingRecord:        record,
			USDAmount:            substrate.FromUSDMilliCentToUSD(usdAmount),
			TransferredUSDAmount: substrate.FromUSDMilliCentToUSD(usdTransferredAmount),
		})
	}

	return pendingRecordsWithUSDAmounts, nil
}

func (svc *AdminService) generateVoucherWithTimestamp() string {
	voucherCode := generators.GenerateVoucherCode(8) // Default to 8-character vouchers
	timestampPart := fmt.Sprintf("%02d%02d", time.Now().Minute(), time.Now().Second())
	return fmt.Sprintf("%s-%s", voucherCode, timestampPart)
}

// AsyncDrainUserUSD drains a specific user's balance to the system account
func (svc *AdminService) AsyncDrainUserUSD(userID int) error {
	wf, err := svc.ewfEngine.NewWorkflow(workflows.WorkflowDrainUser)
	if err != nil {
		return err
	}

	wf.State = map[string]interface{}{
		"user_id": userID,
	}

	return svc.ewfEngine.Run(svc.appCtx, wf, ewf.WithAsync())
}

// AsyncDrainAllUsersUSD drains all users' balances to the system account
func (svc *AdminService) AsyncDrainAllUsersUSD() error {
	users, err := svc.ListAllUsers()
	if err != nil {
		return err
	}

	var multiErr *multierror.Error

	for _, user := range users {
		wf, err := svc.ewfEngine.NewWorkflow(workflows.WorkflowDrainAllUsers)
		if err != nil {
			multiErr = multierror.Append(multiErr, err)
			continue
		}

		wf.State = map[string]interface{}{
			"user_id": user.ID,
		}

		if err := svc.ewfEngine.Run(svc.appCtx, wf, ewf.WithAsync()); err != nil {
			multiErr = multierror.Append(multiErr, err)
		}
	}

	return multiErr.ErrorOrNil()
}
