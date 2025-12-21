package services

import (
	"context"
	"fmt"
	"testing"
	"time"

	"kubecloud/internal/core/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// MOCK IMPLEMENTATIONS
// ============================================================================

type mockUserRepo struct {
	mock.Mock
}

func (m *mockUserRepo) RegisterUser(user *models.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *mockUserRepo) GetUserByID(userID int) (models.User, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return models.User{}, args.Error(1)
	}
	return args.Get(0).(models.User), args.Error(1)
}

func (m *mockUserRepo) GetUserByEmail(email string) (models.User, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return models.User{}, args.Error(1)
	}
	return args.Get(0).(models.User), args.Error(1)
}

func (m *mockUserRepo) UpdateUserByID(user *models.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *mockUserRepo) ListAllUsers() ([]models.User, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.User), args.Error(1)
}

func (m *mockUserRepo) ListRemainingWorkflowsByUserID(userID int) ([]models.GormWorkflowRecord, error) {
	args := m.Called(userID)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).([]models.GormWorkflowRecord), args.Error(1)
}

func (m *mockUserRepo) ListAdmins() ([]models.User, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.User), args.Error(1)
}

func (m *mockUserRepo) DeleteUserByID(userID int) error {
	args := m.Called(userID)
	return args.Error(0)
}

func (m *mockUserRepo) CreditUserBalance(userID int, amount uint64) error {
	args := m.Called(userID, amount)
	return args.Error(0)
}

func (m *mockUserRepo) CountAllUsers() (int64, error) {
	args := m.Called()
	return args.Get(0).(int64), args.Error(1)
}

func (m *mockUserRepo) CreateSSHKey(key *models.SSHKey) error {
	args := m.Called(key)
	return args.Error(0)
}

func (m *mockUserRepo) ListUserSSHKeys(userID int) ([]models.SSHKey, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.SSHKey), args.Error(1)
}

func (m *mockUserRepo) DeleteSSHKey(keyID, userID int) (string, error) {
	args := m.Called(keyID, userID)
	return args.String(0), args.Error(1)
}

func (m *mockUserRepo) GetSSHKeyByID(keyID, userID int) (models.SSHKey, error) {
	args := m.Called(keyID, userID)
	if args.Get(0) == nil {
		return models.SSHKey{}, args.Error(1)
	}
	return args.Get(0).(models.SSHKey), args.Error(1)
}

type mockVoucherRepo struct {
	mock.Mock
}

func (m *mockVoucherRepo) CreateVoucher(voucher *models.Voucher) error {
	args := m.Called(voucher)
	return args.Error(0)
}

func (m *mockVoucherRepo) ListAllVouchers() ([]models.Voucher, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Voucher), args.Error(1)
}

func (m *mockVoucherRepo) GetVoucherByCode(code string) (models.Voucher, error) {
	args := m.Called(code)
	if args.Get(0) == nil {
		return models.Voucher{}, args.Error(1)
	}
	return args.Get(0).(models.Voucher), args.Error(1)
}

func (m *mockVoucherRepo) RedeemVoucher(userID int, username, code string) error {
	args := m.Called(userID, username, code)
	return args.Error(0)
}

// Using mockTransferRecordRepo from admin_service_test.go

// ============================================================================
// TESTS FOR METHODS THAT DON'T REQUIRE EXTERNAL DEPENDENCIES
// ============================================================================

// Test 1: GetUserByEmail - SUCCESS
func TestUserService_GetUserByEmail_Success(t *testing.T) {
	mockUserRepo := new(mockUserRepo)
	mockVoucherRepo := new(mockVoucherRepo)

	expectedUser := models.User{
		ID:       1,
		Email:    "test@example.com",
		Username: "testuser",
	}

	mockUserRepo.On("GetUserByEmail", "test@example.com").Return(expectedUser, nil)

	service := NewUserService(
		context.Background(),
		mockUserRepo, mockVoucherRepo, nil,
		nil, nil, nil, 5, []string{},
	)

	user, err := service.GetUserByEmail("test@example.com")

	require.NoError(t, err)
	assert.Equal(t, expectedUser.Email, user.Email)
	mockUserRepo.AssertCalled(t, "GetUserByEmail", "test@example.com")
}

// Test 2: GetUserByEmail - NOT FOUND
func TestUserService_GetUserByEmail_NotFound(t *testing.T) {
	mockUserRepo := new(mockUserRepo)
	mockVoucherRepo := new(mockVoucherRepo)

	mockUserRepo.On("GetUserByEmail", "invalid@example.com").
		Return(nil, fmt.Errorf("user not found"))

	service := NewUserService(
		context.Background(),
		mockUserRepo, mockVoucherRepo, nil,
		nil, nil, nil, 5, []string{},
	)

	_, err := service.GetUserByEmail("invalid@example.com")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "user not found")
}

// Test 3: CreateSSHKey - SUCCESS
func TestUserService_CreateSSHKey_Success(t *testing.T) {
	mockUserRepo := new(mockUserRepo)
	mockVoucherRepo := new(mockVoucherRepo)

	mockUserRepo.On("CreateSSHKey", mock.AnythingOfType("*models.SSHKey")).Return(nil)

	service := NewUserService(
		context.Background(),
		mockUserRepo, mockVoucherRepo, nil,
		nil, nil, nil, 5, []string{},
	)

	key, err := service.CreateSSHKey(1, "my-key", "ssh-rsa...")

	require.NoError(t, err)
	assert.Equal(t, 1, key.UserID)
	assert.Equal(t, "my-key", key.Name)
	mockUserRepo.AssertCalled(t, "CreateSSHKey", mock.AnythingOfType("*models.SSHKey"))
}

// Test 4: DeleteSSHKey - SUCCESS
func TestUserService_DeleteSSHKey_Success(t *testing.T) {
	mockUserRepo := new(mockUserRepo)
	mockVoucherRepo := new(mockVoucherRepo)

	mockUserRepo.On("DeleteSSHKey", 5, 1).Return("my-key", nil)

	service := NewUserService(
		context.Background(),
		mockUserRepo, mockVoucherRepo, nil,
		nil, nil, nil, 5, []string{},
	)

	keyName, err := service.DeleteSSHKey(1, 5)

	require.NoError(t, err)
	assert.Equal(t, "my-key", keyName)
	mockUserRepo.AssertCalled(t, "DeleteSSHKey", 5, 1)
}

// Test 5: IsVerificationCodeExpired - EXPIRED
func TestUserService_IsVerificationCodeExpired_Expired(t *testing.T) {
	mockUserRepo := new(mockUserRepo)
	mockVoucherRepo := new(mockVoucherRepo)

	// Code from 20 minutes ago, timeout is 5 minutes
	oldTime := time.Now().Add(-20 * time.Minute)

	service := NewUserService(
		context.Background(),
		mockUserRepo, mockVoucherRepo, nil,
		nil, nil, nil, 5, []string{}, // 5 minute timeout
	)

	isExpired := service.IsVerificationCodeExpired(oldTime)

	assert.True(t, isExpired)
}

// Test 6: IsVerificationCodeExpired - NOT EXPIRED
func TestUserService_IsVerificationCodeExpired_NotExpired(t *testing.T) {
	mockUserRepo := new(mockUserRepo)
	mockVoucherRepo := new(mockVoucherRepo)

	// Code from 2 minutes ago, timeout is 5 minutes
	recentTime := time.Now().Add(-2 * time.Minute)

	service := NewUserService(
		context.Background(),
		mockUserRepo, mockVoucherRepo, nil,
		nil, nil, nil, 5, []string{}, // 5 minute timeout
	)

	isExpired := service.IsVerificationCodeExpired(recentTime)

	assert.False(t, isExpired)
}

// Test 7: IsSystemAdmin - TRUE
func TestUserService_IsSystemAdmin_True(t *testing.T) {
	mockUserRepo := new(mockUserRepo)
	mockVoucherRepo := new(mockVoucherRepo)

	service := NewUserService(
		context.Background(),
		mockUserRepo, mockVoucherRepo, nil,
		nil, nil, nil, 5, []string{"admin@example.com", "superuser@example.com"},
	)

	isAdmin := service.IsSystemAdmin("admin@example.com")

	assert.True(t, isAdmin)
}

// Test 8: IsSystemAdmin - FALSE
func TestUserService_IsSystemAdmin_False(t *testing.T) {
	mockUserRepo := new(mockUserRepo)
	mockVoucherRepo := new(mockVoucherRepo)

	service := NewUserService(
		context.Background(),
		mockUserRepo, mockVoucherRepo, nil,
		nil, nil, nil, 5, []string{"admin@example.com"},
	)

	isAdmin := service.IsSystemAdmin("user@example.com")

	assert.False(t, isAdmin)
}

// Test 9: GetVoucherByCode - SUCCESS
func TestUserService_GetVoucherByCode_Success(t *testing.T) {
	mockUserRepo := new(mockUserRepo)
	mockVoucherRepo := new(mockVoucherRepo)

	expectedVoucher := models.Voucher{
		ID:   1,
		Code: "PROMO100",
	}

	mockVoucherRepo.On("GetVoucherByCode", "PROMO100").Return(expectedVoucher, nil)

	service := NewUserService(
		context.Background(),
		mockUserRepo, mockVoucherRepo, nil,
		nil, nil, nil, 5, []string{},
	)

	voucher, err := service.GetVoucherByCode("PROMO100")

	require.NoError(t, err)
	assert.Equal(t, "PROMO100", voucher.Code)
}

// Test 10: GetVoucherByCode - NOT FOUND
func TestUserService_GetVoucherByCode_NotFound(t *testing.T) {
	mockUserRepo := new(mockUserRepo)
	mockVoucherRepo := new(mockVoucherRepo)

	mockVoucherRepo.On("GetVoucherByCode", "INVALID").
		Return(nil, fmt.Errorf("voucher not found"))

	service := NewUserService(
		context.Background(),
		mockUserRepo, mockVoucherRepo, nil,
		nil, nil, nil, 5, []string{},
	)

	_, err := service.GetVoucherByCode("INVALID")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "voucher not found")
}

// Test 11: GenerateRandomCode - VALID RANGE
func TestUserService_GenerateRandomCode_ValidRange(t *testing.T) {
	mockUserRepo := new(mockUserRepo)
	mockVoucherRepo := new(mockVoucherRepo)

	service := NewUserService(
		context.Background(),
		mockUserRepo, mockVoucherRepo, nil,
		nil, nil, nil, 5, []string{},
	)

	code := service.GenerateRandomCode()

	// Should be 4-digit code (1000-9999)
	assert.GreaterOrEqual(t, code, 1000)
	assert.Less(t, code, 10000)
}

// Test 12: CodeTimeoutInMinutes
func TestUserService_CodeTimeoutInMinutes(t *testing.T) {
	mockUserRepo := new(mockUserRepo)
	mockVoucherRepo := new(mockVoucherRepo)

	service := NewUserService(
		context.Background(),
		mockUserRepo, mockVoucherRepo, nil,
		nil, nil, nil, 15, []string{},
	)

	timeout := service.CodeTimeoutInMinutes()

	assert.Equal(t, 15, timeout)
}
