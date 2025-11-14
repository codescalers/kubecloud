package services

import (
	"kubecloud/internal"
	"context"
	"kubecloud/internal/metrics"
	"kubecloud/internal/substrate"
	"kubecloud/internal/models"
	"slices"
	"time"

	"github.com/xmonader/ewf"
)

type UserService struct {
	userRepo    models.UserRepository
	voucherRepo models.VoucherRepository
	prRepo      models.PendingRecordRepository

	appCtx          context.Context
	substrateClient substrate.Substrate
	randomizer      internal.Randomizer
	ewfEngine       *ewf.Engine
	kycClient       *internal.KYCClient
	metrics         *metrics.Metrics

	// configs
	codeTimeoutMin int
	systemAdmins   []string
}

func NewUserService(appCtx context.Context,
	userRepo models.UserRepository,
	voucherRepo models.VoucherRepository,
	pendingRecordRepo models.PendingRecordRepository,
	substrateClient substrate.Substrate,
	randomizer internal.Randomizer,
	ewfEngine *ewf.Engine,
	kycClient *internal.KYCClient,
	metrics *metrics.Metrics,
	codeTimeoutMin int,
	systemAdmins []string,
) UserService {
	return UserService{
		userRepo:    userRepo,
		voucherRepo: voucherRepo,
		prRepo:      pendingRecordRepo,

		appCtx:          appCtx,
		substrateClient: substrateClient,
		randomizer:      randomizer,
		ewfEngine:       ewfEngine,
		kycClient:       kycClient,
		metrics:         metrics,

		codeTimeoutMin: codeTimeoutMin,
		systemAdmins:   systemAdmins,
	}
}

type UserWithPendingBalance struct {
	models.User
	PendingBalanceUSD float64 `json:"pending_balance_usd"`
}

func (svc *UserService) GetUserByEmail(email string) (models.User, error) {
	return svc.userRepo.GetUserByEmail(email)
}

func (svc *UserService) GetUserByID(userID int) (models.User, error) {
	return svc.userRepo.GetUserByID(userID)
}

func (svc *UserService) GetUserWithPendingBalance(userID int) (UserWithPendingBalance, error) {
	user, err := svc.GetUserByID(userID)
	if err != nil {
		return UserWithPendingBalance{}, err
	}

	usdMillicentPendingAmount, err := svc.GetUserPendingBalanceInUSDMillicent(userID)
	if err != nil {
		return UserWithPendingBalance{}, err
	}

	return UserWithPendingBalance{
		User:              user,
		PendingBalanceUSD: substrate.FromUSDMilliCentToUSD(usdMillicentPendingAmount),
	}, nil
}

func (svc *UserService) GetUserPendingBalanceInUSDMillicent(userID int) (uint64, error) {
	pendingRecords, err := svc.prRepo.ListUserPendingRecords(userID)
	if err != nil {
		return 0, err
	}

	var tftPendingAmount uint64
	for _, record := range pendingRecords {
		tftPendingAmount += record.TFTAmount - record.TransferredTFTAmount
	}

	usdMillicentPendingAmount, err := svc.substrateClient.FromTFTtoUSDMillicent(tftPendingAmount)
	if err != nil {
		return 0, err
	}

	return usdMillicentPendingAmount, nil
}

func (svc *UserService) ListUserPendingRecordsWithUSDAmounts(userID int) ([]PendingRecordsWithUSDAmounts, error) {
	pendingRecords, err := svc.prRepo.ListUserPendingRecords(userID)
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

func (svc *UserService) GetUserBalanceInUSDMillicent(userMnemonic string) (uint64, error) {
	return svc.substrateClient.GetUserBalanceUSDMillicent(userMnemonic)
}

func (svc *UserService) UpdateUserByID(user *models.User) error {
	return svc.userRepo.UpdateUserByID(user)
}

func (svc *UserService) ListUserSSHKeys(userID int) ([]models.SSHKey, error) {
	return svc.userRepo.ListUserSSHKeys(userID)
}

func (svc *UserService) CreateSSHKey(userID int, sshKeyName, publicKey string) (models.SSHKey, error) {
	sshKey := models.SSHKey{
		UserID:    userID,
		Name:      sshKeyName,
		PublicKey: publicKey,
	}

	return sshKey, svc.userRepo.CreateSSHKey(&sshKey)
}

func (svc *UserService) DeleteSSHKey(userID, sshKeyID int) (string, error) {
	return svc.userRepo.DeleteSSHKey(sshKeyID, userID)
}

func (svc *UserService) CheckKYCVerification(requestCtx context.Context, userID int, userSponsored bool, userAccountAddress string) error {
	// Check KYC verification status without blocking login
	sponsored, err := svc.kycClient.IsUserVerified(requestCtx, userAccountAddress)
	if err != nil {
		return err
	}

	if userSponsored != sponsored {
		if err := svc.UpdateUserByID(&models.User{ID: userID, Sponsored: sponsored}); err != nil {
			return err
		}
	}

	return nil
}

func (svc *UserService) GetVoucherByCode(voucherCode string) (models.Voucher, error) {
	return svc.voucherRepo.GetVoucherByCode(voucherCode)
}

func (svc *UserService) GenerateRandomCode() int {
	return svc.randomizer.GenerateRandomCode()
}

func (svc *UserService) IsVerificationCodeExpired(userLastUpdatedAt time.Time) bool {
	return userLastUpdatedAt.Add(time.Duration(svc.codeTimeoutMin) * time.Minute).Before(time.Now())
}

func (svc *UserService) CodeTimeoutInMinutes() int {
	return svc.codeTimeoutMin
}

func (svc *UserService) IsSystemAdmin(userEmail string) bool {
	return slices.Contains(svc.systemAdmins, userEmail)
}

func (svc *UserService) AsyncRegisterUser(name, email, password string) (string, error) {
	wf, err := svc.ewfEngine.NewWorkflow(internal.WorkflowUserRegistration)
	if err != nil {
		return "", err
	}

	wf.State = ewf.State{
		"name":     name,
		"email":    email,
		"password": password,
	}

	err = svc.ewfEngine.RunAsync(svc.appCtx, wf)
	return wf.UUID, err
}

func (svc *UserService) AsyncVerifyUserRegistration(requestCtx context.Context, userID int, userEmail, username string) (string, error) {
	wf, err := svc.ewfEngine.NewWorkflow(internal.WorkflowUserVerification)
	if err != nil {
		return "", err
	}

	if err = svc.ewfEngine.Store().SaveWorkflow(requestCtx, wf); err != nil {
		return "", err
	}

	// Start the user-verification workflow
	wf.State = ewf.State{
		"email":   userEmail,
		"name":    username,
		"user_id": userID,
	}

	err = svc.ewfEngine.RunAsync(svc.appCtx, wf)
	return wf.UUID, err
}

func (svc *UserService) AsyncStripeChargeBalance(userID int, userStripeCustomerID, paymentMethodID, userMnemonic, username string, requestAmount float64) (string, error) {
	wf, err := svc.ewfEngine.NewWorkflow(internal.WorkflowChargeBalance)
	if err != nil {
		return "", err
	}

	wf.State = ewf.State{
		"user_id":            userID,
		"stripe_customer_id": userStripeCustomerID,
		"payment_method_id":  paymentMethodID,
		"amount":             substrate.FromUSDToUSDMillicent(requestAmount),
		"mnemonic":           userMnemonic,
		"username":           username,
		"transfer_mode":      models.ChargeBalanceMode,
	}

	err = svc.ewfEngine.RunAsync(svc.appCtx, wf)
	return wf.UUID, err
}

func (svc *UserService) AsyncRedeemVoucher(userID int, voucherValue float64, userMnemonic, userUsername, voucherCode string) (string, error) {
	err := svc.voucherRepo.RedeemVoucher(voucherCode)
	if err != nil {
		return "", err
	}

	wf, err := svc.ewfEngine.NewWorkflow(internal.WorkflowRedeemVoucher)
	if err != nil {
		return "", err
	}

	wf.State = map[string]interface{}{
		"user_id":       userID,
		"amount":        substrate.FromUSDToUSDMillicent(voucherValue),
		"mnemonic":      userMnemonic,
		"username":      userUsername,
		"transfer_mode": models.RedeemVoucherMode,
	}

	err = svc.ewfEngine.RunAsync(svc.appCtx, wf)
	return wf.UUID, err
}

func (svc *UserService) GetWorkflowStatus(ctx context.Context, wfUUID string) (ewf.WorkflowStatus, error) {
	workflow, err := svc.ewfEngine.Store().LoadWorkflowByUUID(ctx, wfUUID)
	if err != nil {
		return "", err
	}

	return workflow.Status, nil
}

func (svc *UserService) IncrementStripePaymentFailure() {
	svc.metrics.IncrementStripePaymentFailure()
}
