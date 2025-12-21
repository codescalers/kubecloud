package services

import (
	"context"
	"fmt"
	"kubecloud/internal/core/generators"
	"kubecloud/internal/core/models"
	"kubecloud/internal/core/persistence"
	"kubecloud/internal/core/workflows"
	"kubecloud/internal/infrastructure/logger"
	"kubecloud/internal/infrastructure/mailservice"
	mailsender "kubecloud/internal/infrastructure/mailservice/mail_sender"
	"kubecloud/internal/infrastructure/notification"
	"kubecloud/internal/infrastructure/substrate"
	"time"

	"sync"

	"github.com/hashicorp/go-multierror"
	"github.com/xmonader/ewf"
)

type AdminService struct {
	userRepo            models.UserRepository
	contractsRepo       models.ContractDataRepository
	transferRecordsRepo models.TransferRecordRepository
	voucherRepo         models.VoucherRepository
	transRepo           models.TransactionRepository
	ewfRepo             *persistence.GormEWFRepository

	appCtx                 context.Context
	substrateClient        substrate.Substrate
	ewfEngine              *ewf.Engine
	mailService            mailservice.MailService
	notificationDispatcher *notification.NotificationDispatcher
}

func NewAdminService(appCtx context.Context,
	userRepo models.UserRepository,
	contractsRepo models.ContractDataRepository,
	transferRecordsRepo models.TransferRecordRepository,
	voucherRepo models.VoucherRepository,
	transactionRepo models.TransactionRepository,
	substrateClient substrate.Substrate,
	ewfEngine *ewf.Engine,
	mailService mailservice.MailService,
	notificationDispatcher *notification.NotificationDispatcher,
	ewfRepo *persistence.GormEWFRepository,
) AdminService {
	return AdminService{
		userRepo:            userRepo,
		contractsRepo:       contractsRepo,
		transferRecordsRepo: transferRecordsRepo,
		voucherRepo:         voucherRepo,
		transRepo:           transactionRepo,
		ewfRepo:             ewfRepo,

		appCtx:                 appCtx,
		substrateClient:        substrateClient,
		ewfEngine:              ewfEngine,
		mailService:            mailService,
		notificationDispatcher: notificationDispatcher,
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
func (svc *AdminService) AsyncDrainUserUSD(userID, adminID int) error {
	wf, err := svc.ewfEngine.NewWorkflow(workflows.WorkflowDrainUser, ewf.WithDisplayName("Drain user balance"))
	if err != nil {
		return err
	}

	wf.State = map[string]interface{}{
		"target_user_id": userID,
		"config": map[string]interface{}{
			"user_id": adminID,
		},
	}

	return svc.ewfEngine.Run(svc.appCtx, wf, ewf.WithAsync())
}

// AsyncDrainAllUsersUSD drains all users' balances to the system account
func (svc *AdminService) AsyncDrainAllUsersUSD(adminID int) error {
	wf, err := svc.ewfEngine.NewWorkflow(workflows.WorkflowDrainAllUsers, ewf.WithDisplayName("Drain all users balance"))
	if err != nil {
		return err
	}

	wf.State = map[string]interface{}{
		"config": map[string]interface{}{
			"user_id": adminID,
		},
	}

	return svc.ewfEngine.Run(svc.appCtx, wf, ewf.WithAsync())
}

// AdminWorkflow represents a workflow for admin display purposes
type AdminWorkflow struct {
	UUID        string            `json:"uuid"`
	Name        string            `json:"name"`
	DisplayName string            `json:"display_name"`
	Status      string            `json:"status"`
	CurrentStep int               `json:"current_step"`
	TotalSteps  int               `json:"total_steps"`
	StepName    string            `json:"step_name"`
	State       map[string]any    `json:"state"`
	UserID      int               `json:"user_id"`
	CreatedAt   time.Time         `json:"created_at"`
	QueueName   string            `json:"queue_name"`
	Metadata    map[string]string `json:"metadata"`
	Error       string            `json:"error"`
}

// ListAllWorkflows returns all workflows with optional filtering by status
func (svc *AdminService) ListAllWorkflows(status string) ([]AdminWorkflow, error) {
	workflows, err := svc.ewfRepo.ListAllWorkflows(svc.appCtx, status)
	if err != nil {
		return nil, err
	}

	return svc.convertToAdminWorkflows(workflows), nil
}

// ListAllWorkflowsPaginated returns paginated workflows with optional filtering by status
func (svc *AdminService) ListAllWorkflowsPaginated(status string, page, limit int) (int, []AdminWorkflow, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	workflows, err := svc.ewfRepo.ListAllWorkflows(svc.appCtx, status)
	if err != nil {
		return 0, nil, err
	}

	total := len(workflows)
	start := (page - 1) * limit
	end := start + limit

	if start >= total {
		return total, []AdminWorkflow{}, nil
	}

	if end > total {
		end = total
	}

	paginatedWorkflows := workflows[start:end]
	adminWorkflows := svc.convertToAdminWorkflows(paginatedWorkflows)

	return total, adminWorkflows, nil
}

// convertToAdminWorkflows converts ewf workflows to AdminWorkflow format
func (svc *AdminService) convertToAdminWorkflows(workflows []*ewf.Workflow) []AdminWorkflow {
	var adminWorkflows []AdminWorkflow
	for _, wf := range workflows {
		displayName := wf.DisplayName
		if displayName == "" {
			displayName = wf.Name
		}

		stepName := ""
		currentStepIndex := wf.CurrentStep
		// For completed workflows, CurrentStep might be equal to or greater than the number of steps
		// In that case, show the last step name
		if len(wf.Steps) > 0 {
			if currentStepIndex >= len(wf.Steps) {
				currentStepIndex = len(wf.Steps) - 1
			}
			if currentStepIndex >= 0 && currentStepIndex < len(wf.Steps) {
				stepName = wf.Steps[currentStepIndex].Name
			}
		}
		// If step name is still empty, don't include it in the display
		if stepName == "" {
			stepName = "-"
		}

		// Extract user_id from state
		userID := 0
		if uid, ok := wf.State["user_id"]; ok {
			switch v := uid.(type) {
			case int:
				userID = v
			case float64:
				userID = int(v)
			}
		}
		// Also check for gorm_user_id which is set by the persistence layer
		if uid, ok := wf.State["gorm_user_id"]; ok {
			switch v := uid.(type) {
			case int:
				userID = v
			case float64:
				userID = int(v)
			}
		}

		adminWorkflows = append(adminWorkflows, AdminWorkflow{
			UUID:        wf.UUID,
			Name:        wf.Name,
			DisplayName: displayName,
			Status:      string(wf.Status),
			CurrentStep: wf.CurrentStep,
			TotalSteps:  len(wf.Steps),
			StepName:    stepName,
			State:       wf.State,
			UserID:      userID,
			CreatedAt:   wf.CreatedAt,
			QueueName:   wf.QueueName,
			Metadata:    wf.Metadata,
			Error:       wf.Error,
		})
	}

	return adminWorkflows
}

func (svc *AdminService) SendMailToAllUsers(body, subject string, users []models.User, adminID int, attachments ...mailsender.Attachment) {

	failedEmails := svc.sendBulkSystemMails(users, body, subject, attachments...)

	totalUsers := len(users)
	successfulEmails := len(users) - failedEmails

	// after sending, send an SSE notification on the progress
	notif := notification.NewNotification(adminID, models.NotificationTypeAdmin).
		Info(fmt.Sprintf(
			"Mail sent to %d/%d users successfully",
			successfulEmails,
			totalUsers,
		)).
		WithSubject("Mail Sending Progress").
		WithChannels(notification.ChannelUI).
		NoPersist().
		Build()

	if err := svc.notificationDispatcher.Send(svc.appCtx, notif); err != nil {
		logger.GetLogger().Error().Err(err).Msg("failed to send mail progress notification")
	}

}

// SendBulkSystemMails send system mails to all passed emails
func (svc *AdminService) sendBulkSystemMails(users []models.User, body string, subject string, attachments ...mailsender.Attachment) int {
	emailConcurrencyLimiter := make(chan struct{}, svc.mailService.GetMailConfig().MaxConcurrentSends)

	var (
		wg           sync.WaitGroup
		mu           sync.Mutex
		failedEmails []string
	)

	for _, user := range users {
		wg.Add(1)
		emailConcurrencyLimiter <- struct{}{}
		go func(user models.User) {
			defer wg.Done()
			defer func() { <-emailConcurrencyLimiter }()
			err := svc.mailService.SendSystemAnnouncementMail(user.Email, body, subject, attachments...)
			if err != nil {
				logger.GetLogger().Error().Err(err).Str("user_email", user.Email).Msg("failed to send mail to user")
				mu.Lock()
				failedEmails = append(failedEmails, user.Email)
				mu.Unlock()
			}
		}(user)
	}

	wg.Wait()

	return len(failedEmails)
}
