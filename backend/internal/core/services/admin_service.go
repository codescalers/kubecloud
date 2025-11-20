package services

import (
	"context"
	"fmt"
	"kubecloud/internal/core/generators"
	"kubecloud/internal/core/models"
	"kubecloud/internal/core/workflows"
	"kubecloud/internal/infrastructure/logger"
	"kubecloud/internal/infrastructure/substrate"

	"sync"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/xmonader/ewf"
)

type AdminService struct {
	userRepo            models.UserRepository
	contractsRepo       models.ContractDataRepository
	transferRecordsRepo models.TransferRecordRepository
	voucherRepo         models.VoucherRepository
	transRepo           models.TransactionRepository

	appCtx          context.Context
	substrateClient substrate.Substrate
	ewfEngine       *ewf.Engine
}

func NewAdminService(appCtx context.Context,
	userRepo models.UserRepository,
	contractsRepo models.ContractDataRepository,
	transferRecordsRepo models.TransferRecordRepository,
	voucherRepo models.VoucherRepository,
	transactionRepo models.TransactionRepository,
	substrateClient substrate.Substrate,
	ewfEngine *ewf.Engine,
) AdminService {
	return AdminService{
		userRepo:            userRepo,
		contractsRepo:       contractsRepo,
		transferRecordsRepo: transferRecordsRepo,
		voucherRepo:         voucherRepo,
		transRepo:           transactionRepo,

		appCtx:          appCtx,
		substrateClient: substrateClient,
		ewfEngine:       ewfEngine,
	}
}

const maxConcurrentBalanceFetches = 20

type UserWithTFTBalance struct {
	models.User
	BalanceInTFT float64 `json:"balance_in_tft"`
}

type TransferRecordsWithTFTAmount struct {
	models.TransferRecord
	TFTAmountInWholeUnit float32 `json:"tft_amount_in_whole_unit"`
}

func (svc *AdminService) ListAllUsers() ([]models.User, error) {
	return svc.userRepo.ListAllUsers()
}

func (svc *AdminService) GetUserByID(id int) (models.User, error) {
	return svc.userRepo.GetUserByID(id)
}

func (svc *AdminService) ListAllUsersIncludingUSDBalance() ([]UserWithTFTBalance, error) {
	users, err := svc.ListAllUsers()
	// Here is the only critical errors, not the balance related ones
	if err != nil {
		return nil, err
	}

	var (
		usersWithBalance []UserWithTFTBalance
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

			balanceInTFTUnit, err := svc.substrateClient.GetUserTFTBalance(user.Mnemonic)
			if err != nil {
				logger.GetLogger().Error().Err(err).Int("user_id", user.ID).Msg("failed to get user balance")
				mu.Lock()
				balanceErrors = multierror.Append(balanceErrors, fmt.Errorf("failed to get balance for user %d: %w", user.ID, err))
				mu.Unlock()
				return
			}

			mu.Lock()
			usersWithBalance = append(usersWithBalance, UserWithTFTBalance{
				User:         user,
				BalanceInTFT: float64(balanceInTFTUnit) / TFTUnitFactor,
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

func (svc *AdminService) ListAllTransferRecordsWithTFTAmount() ([]TransferRecordsWithTFTAmount, error) {
	transferRecords, err := svc.transferRecordsRepo.ListTransferRecords()
	if err != nil {
		return nil, fmt.Errorf("failed to list transfer records: %w", err)
	}

	var transferRecordsResponse []TransferRecordsWithTFTAmount
	for _, transferRecord := range transferRecords {
		transferRecordsResponse = append(transferRecordsResponse, TransferRecordsWithTFTAmount{
			TransferRecord:       transferRecord,
			TFTAmountInWholeUnit: float32(transferRecord.TFTAmount) / TFTUnitFactor,
		})
	}

	return transferRecordsResponse, nil
}

func (svc *AdminService) generateVoucherWithTimestamp() string {
	voucherCode := generators.GenerateVoucherCode(8) // Default to 8-character vouchers
	timestampPart := fmt.Sprintf("%02d%02d", time.Now().Minute(), time.Now().Second())
	return fmt.Sprintf("%s-%s", voucherCode, timestampPart)
}

func (svc *AdminService) CreditUserBalance(ctx context.Context, transaction models.Transaction, user *models.User) error {
	if err := svc.transRepo.CreateTransaction(&transaction); err != nil {
		return fmt.Errorf("failed to create transaction: %w", err)
	}

	millicentAmount := substrate.FromUSDToUSDMillicent(transaction.Amount)
	user.CreditedBalance += millicentAmount
	if err := svc.userRepo.UpdateUserByID(&models.User{ID: transaction.UserID}); err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
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

	return svc.ewfEngine.RunAsync(svc.appCtx, wf)
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

		if err := svc.ewfEngine.RunAsync(svc.appCtx, wf); err != nil {
			multiErr = multierror.Append(multiErr, err)
		}
	}

	return multiErr.ErrorOrNil()
}
