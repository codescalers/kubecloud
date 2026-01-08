package services

import (
	"context"
	"encoding/json"
	"fmt"
	"kubecloud/internal/core/generators"
	"kubecloud/internal/core/models"
	"kubecloud/internal/core/persistence"
	"kubecloud/internal/core/workflows"
	"kubecloud/internal/infrastructure/gridclient"
	"kubecloud/internal/infrastructure/kyc"
	"kubecloud/internal/infrastructure/metrics"
	"slices"
	"time"

	"github.com/xmonader/ewf"
)

type UserService struct {
	userRepo    models.UserRepository
	voucherRepo models.VoucherRepository

	appCtx     context.Context
	gridClient gridclient.GridClient
	ewfEngine  *ewf.Engine
	kycClient  *kyc.KYCClient
	metrics    *metrics.Metrics

	// configs
	codeTimeoutMin int
	systemAdmins   []string
}

func NewUserService(appCtx context.Context,
	userRepo models.UserRepository,
	voucherRepo models.VoucherRepository,
	gridClient gridclient.GridClient,
	ewfEngine *ewf.Engine,
	kycClient *kyc.KYCClient,
	metrics *metrics.Metrics,
	codeTimeoutMin int,
	systemAdmins []string,
) UserService {
	return UserService{
		userRepo:    userRepo,
		voucherRepo: voucherRepo,

		appCtx:     appCtx,
		gridClient: gridClient,
		ewfEngine:  ewfEngine,
		kycClient:  kycClient,
		metrics:    metrics,

		codeTimeoutMin: codeTimeoutMin,
		systemAdmins:   systemAdmins,
	}
}

type UserWithBalancesInUSD struct {
	models.User
	CreditCardBalanceInUSD float64 `json:"credit_card_balance_in_usd"`
	CreditedBalanceInUSD   float64 `json:"credited_balance_in_usd"`
	DebtInUSD              float64 `json:"debt_in_usd"`
	BalanceInTFT           float64 `json:"balance_in_tft,omitempty"`
}

func (svc *UserService) GetUserByEmail(email string) (models.User, error) {
	return svc.userRepo.GetUserByEmail(email)
}

func (svc *UserService) GetUserByID(userID int) (models.User, error) {
	return svc.userRepo.GetUserByID(userID)
}

func (svc *UserService) GetUserWithBalancesInUSD(userID int) (UserWithBalancesInUSD, error) {
	user, err := svc.GetUserByID(userID)
	if err != nil {
		return UserWithBalancesInUSD{}, err
	}

	return UserWithBalancesInUSD{
		User:                   user,
		CreditedBalanceInUSD:   gridclient.FromUSDMilliCentToUSD(user.CreditedBalance),
		CreditCardBalanceInUSD: gridclient.FromUSDMilliCentToUSD(user.CreditCardBalance),
		DebtInUSD:              gridclient.FromUSDMilliCentToUSD(user.Debt),
	}, nil
}

func (svc *UserService) ListRemainingWorkflowsByUserID(userID int) ([]*ewf.Workflow, error) {
	records, err := svc.userRepo.ListRemainingWorkflowsByUserID(userID)
	if err != nil {
		return nil, err
	}

	workflows := make([]*ewf.Workflow, 0, len(records))

	for _, rec := range records {
		var wf ewf.Workflow
		if err := json.Unmarshal(rec.Data, &wf); err != nil {
			return nil, fmt.Errorf("failed to unmarshal workflow %s: %w", rec.UUID, err)
		}
		workflows = append(workflows, &wf)
	}
	return workflows, nil
}

func (svc *UserService) GetUserBalanceInUSDMillicent(userMnemonic string) (uint64, error) {
	return svc.gridClient.GetUserBalanceUSDMillicent(userMnemonic)
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
	return generators.GenerateVerificationCode(4) // Default to 4-digit verification codes
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
	wf, err := svc.ewfEngine.NewWorkflow(workflows.WorkflowUserRegistration, ewf.WithDisplayName("Register user"))
	if err != nil {
		return "", err
	}

	wf.State = ewf.State{
		"name":     name,
		"email":    email,
		"password": password,
	}

	err = svc.ewfEngine.Run(svc.appCtx, wf, ewf.WithAsync())
	return wf.UUID, err
}

func (svc *UserService) AsyncVerifyUserRegistration(requestCtx context.Context, userID int, userEmail, username string) (string, error) {
	wf, err := svc.ewfEngine.NewWorkflow(workflows.WorkflowUserVerification, ewf.WithDisplayName("Verify user registration"))
	if err != nil {
		return "", err
	}

	if err = svc.ewfEngine.Store().SaveWorkflow(requestCtx, wf); err != nil {
		return "", err
	}

	// Start the user-verification workflow
	wf.State = ewf.State{
		"email": userEmail,
		"name":  username,
		"config": map[string]interface{}{
			"user_id": userID,
		},
	}

	err = svc.ewfEngine.Run(svc.appCtx, wf, ewf.WithAsync())
	return wf.UUID, err
}

func (svc *UserService) AsyncStripeChargeBalance(userID int, userStripeCustomerID, paymentMethodID, userMnemonic, username string, requestAmount float64) (string, error) {
	wf, err := svc.ewfEngine.NewWorkflow(workflows.WorkflowChargeBalance, ewf.WithDisplayName("Charge balance"))
	if err != nil {
		return "", err
	}

	wf.State = ewf.State{
		"stripe_customer_id": userStripeCustomerID,
		"payment_method_id":  paymentMethodID,
		"amount":             gridclient.FromUSDToUSDMillicent(requestAmount),
		"username":           username,
		"config": map[string]interface{}{
			"user_id":  userID,
			"mnemonic": userMnemonic,
		},
	}

	if err = persistence.SetStateUserID(&wf, userID); err != nil {
		return "", err
	}

	err = svc.ewfEngine.Run(svc.appCtx, wf, ewf.WithAsync())
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
