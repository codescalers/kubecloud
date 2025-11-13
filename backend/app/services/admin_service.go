package services

import (
	"context"
	"fmt"
	"kubecloud/internal"
	"kubecloud/internal/constants"
	"kubecloud/internal/substrate"
	"kubecloud/models"
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
	randomizer      internal.Randomizer
	ewfEngine       *ewf.Engine
}

func NewAdminService(appCtx context.Context,
	userRepo models.UserRepository,
	userNodeRepo models.UserNodesRepository,
	pendingRecordRepo models.PendingRecordRepository,
	voucherRepo models.VoucherRepository,
	transactionRepo models.TransactionRepository,
	substrateClient substrate.Substrate,
	randomizer internal.Randomizer,
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
		randomizer:      randomizer,
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

func (svc *AdminService) ListAllUsersIncludingUSDBalance() ([]models.User, error) {
	users, err := svc.ListAllUsers()
	if err != nil {
		return nil, err
	}

	var (
		usersWithBalance []UserWithUSDBalance
		wg               sync.WaitGroup
		mu               sync.Mutex
		multiErr         *multierror.Error
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
				multiErr = multierror.Append(multiErr, fmt.Errorf("failed to get balance for user %d: %w", user.ID, err))
				mu.Unlock()
				return
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

	return nil, multiErr.ErrorOrNil()
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

	wf, err := svc.ewfEngine.NewWorkflow(constants.WorkflowAdminCreditBalance)
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

	svc.ewfEngine.RunAsync(svc.appCtx, wf)
	return nil
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
	voucherCode := svc.randomizer.GenerateRandomVoucher()
	timestampPart := fmt.Sprintf("%02d%02d", time.Now().Minute(), time.Now().Second())
	return fmt.Sprintf("%s-%s", voucherCode, timestampPart)
}
