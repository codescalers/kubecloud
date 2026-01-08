package services

import (
	"context"
	"fmt"
	"testing"
	"time"

	"kubecloud/internal/core/models"
	"kubecloud/internal/infrastructure/gridclient"
	"kubecloud/internal/infrastructure/mailservice"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// We're using the mock implementations from user_service_test.go
// Adding only the missing methods needed for DeductUserBalance

func (m *mockUserRepo) DeductUserBalance(user *models.User, amount uint64) error {
	args := m.Called(user, amount)
	return args.Error(0)
}

func (m *mockUserRepo) GetUserLastCalcTime(userID int) (time.Time, error) {
	args := m.Called(userID)
	return args.Get(0).(time.Time), args.Error(1)
}

func (m *mockUserRepo) UpdateUserLastCalcTime(userID int, lastCalcTime time.Time) error {
	args := m.Called(userID, lastCalcTime)
	return args.Error(0)
}

// Contract data repository mock
type mockContractDataRepo struct {
	mock.Mock
}

func (m *mockContractDataRepo) CreateUserContractData(contractData *models.UserContractData) error {
	args := m.Called(contractData)
	return args.Error(0)
}

func (m *mockContractDataRepo) DeleteUserContract(contractID uint64) error {
	args := m.Called(contractID)
	return args.Error(0)
}

func (m *mockContractDataRepo) ListUserRentedNodes(userID int) ([]models.UserContractData, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.UserContractData), args.Error(1)
}

func (m *mockContractDataRepo) GetUserNodeByNodeID(nodeID uint64) (models.UserContractData, error) {
	args := m.Called(nodeID)
	if args.Get(0) == nil {
		return models.UserContractData{}, args.Error(1)
	}
	return args.Get(0).(models.UserContractData), args.Error(1)
}

func (m *mockContractDataRepo) GetUserNodeByContractID(contractID uint64) (models.UserContractData, error) {
	args := m.Called(contractID)
	if args.Get(0) == nil {
		return models.UserContractData{}, args.Error(1)
	}
	return args.Get(0).(models.UserContractData), args.Error(1)
}

func (m *mockContractDataRepo) ListAllReservedNodes() ([]models.UserContractData, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.UserContractData), args.Error(1)
}

func (m *mockContractDataRepo) ListAllContractsInPeriod(userID int, start, end time.Time) ([]models.UserContractData, error) {
	args := m.Called(userID, start, end)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.UserContractData), args.Error(1)
}

// Transaction repository mock
type mockTransactionRepo struct {
	mock.Mock
}

func (m *mockTransactionRepo) CreateTransaction(transaction *models.Transaction) error {
	args := m.Called(transaction)
	return args.Error(0)
}

// Transfer record repository mock
type mockTransferRecordRepo struct {
	mock.Mock
}

func (m *mockTransferRecordRepo) CreateTransferRecord(record *models.TransferRecord) error {
	args := m.Called(record)
	return args.Error(0)
}

func (m *mockTransferRecordRepo) ListTransferRecords() ([]models.TransferRecord, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.TransferRecord), args.Error(1)
}

func (m *mockTransferRecordRepo) ListUserTransferRecords(userID int) ([]models.TransferRecord, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.TransferRecord), args.Error(1)
}

func (m *mockTransferRecordRepo) ListPendingTransferRecords() ([]models.TransferRecord, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.TransferRecord), args.Error(1)
}

func (m *mockTransferRecordRepo) ListFailedTransferRecords() ([]models.TransferRecord, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.TransferRecord), args.Error(1)
}

func (m *mockTransferRecordRepo) UpdateTransferRecordState(recordID int, state models.State, failure string) error {
	args := m.Called(recordID, state, failure)
	return args.Error(0)
}

func (m *mockTransferRecordRepo) CalculateTotalPendingTFTAmountPerUser(userID int) (uint64, error) {
	args := m.Called(userID)
	return args.Get(0).(uint64), args.Error(1)
}

var dummyMailService = mailservice.MailService{}

// Test 1: ListAllUsers - SUCCESS
func TestAdminService_ListAllUsers_Success(t *testing.T) {
	mockUserRepo := new(mockUserRepo)
	mockContractsRepo := new(mockContractDataRepo)
	mockTransferRecordsRepo := new(mockTransferRecordRepo)
	mockVoucherRepo := new(mockVoucherRepo)
	mockTransRepo := new(mockTransactionRepo)

	users := []models.User{
		{ID: 1, Email: "user1@example.com", Username: "user1"},
		{ID: 2, Email: "user2@example.com", Username: "user2"},
	}

	mockUserRepo.On("ListAllUsers").Return(users, nil)

	var gridClient gridclient.GridClient

	service := NewAdminService(
		context.Background(),
		mockUserRepo, mockContractsRepo, mockTransferRecordsRepo, mockVoucherRepo, mockTransRepo,
		gridClient, nil, dummyMailService, nil, nil,
	)

	result, err := service.ListAllUsers()

	require.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "user1@example.com", result[0].Email)
	mockUserRepo.AssertCalled(t, "ListAllUsers")
}

// Test 2: ListAllUsers - EMPTY
func TestAdminService_ListAllUsers_Empty(t *testing.T) {
	mockUserRepo := new(mockUserRepo)
	mockContractsRepo := new(mockContractDataRepo)
	mockTransferRecordsRepo := new(mockTransferRecordRepo)
	mockVoucherRepo := new(mockVoucherRepo)
	mockTransRepo := new(mockTransactionRepo)

	mockUserRepo.On("ListAllUsers").Return([]models.User{}, nil)

	var gridClient gridclient.GridClient

	service := NewAdminService(
		context.Background(),
		mockUserRepo, mockContractsRepo, mockTransferRecordsRepo, mockVoucherRepo, mockTransRepo,
		gridClient, nil, dummyMailService, nil, nil,
	)

	result, err := service.ListAllUsers()

	require.NoError(t, err)
	assert.Len(t, result, 0)
}

// Test 3: ListAllUsers - ERROR
func TestAdminService_ListAllUsers_Error(t *testing.T) {
	mockUserRepo := new(mockUserRepo)
	mockContractsRepo := new(mockContractDataRepo)
	mockTransferRecordsRepo := new(mockTransferRecordRepo)
	mockVoucherRepo := new(mockVoucherRepo)
	mockTransRepo := new(mockTransactionRepo)

	mockUserRepo.On("ListAllUsers").Return(nil, fmt.Errorf("database error"))

	var gridClient gridclient.GridClient

	service := NewAdminService(
		context.Background(),
		mockUserRepo, mockContractsRepo, mockTransferRecordsRepo, mockVoucherRepo, mockTransRepo,
		gridClient, nil, dummyMailService, nil, nil,
	)

	_, err := service.ListAllUsers()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "database error")
}

// Test 4: DeleteUserByID - SUCCESS
func TestAdminService_DeleteUserByID_Success(t *testing.T) {
	mockUserRepo := new(mockUserRepo)
	mockContractsRepo := new(mockContractDataRepo)
	mockTransferRecordsRepo := new(mockTransferRecordRepo)
	mockVoucherRepo := new(mockVoucherRepo)
	mockTransRepo := new(mockTransactionRepo)

	mockUserRepo.On("DeleteUserByID", 1).Return(nil)

	var gridClient gridclient.GridClient

	service := NewAdminService(
		context.Background(),
		mockUserRepo, mockContractsRepo, mockTransferRecordsRepo, mockVoucherRepo, mockTransRepo,
		gridClient, nil, dummyMailService, nil, nil,
	)

	err := service.DeleteUserByID(1)

	require.NoError(t, err)
	mockUserRepo.AssertCalled(t, "DeleteUserByID", 1)
}

// Test 5: DeleteUserByID - USER NOT FOUND
func TestAdminService_DeleteUserByID_NotFound(t *testing.T) {
	mockUserRepo := new(mockUserRepo)
	mockContractsRepo := new(mockContractDataRepo)
	mockTransferRecordsRepo := new(mockTransferRecordRepo)
	mockVoucherRepo := new(mockVoucherRepo)
	mockTransRepo := new(mockTransactionRepo)

	mockUserRepo.On("DeleteUserByID", 999).Return(fmt.Errorf("user not found"))

	var gridClient gridclient.GridClient

	service := NewAdminService(
		context.Background(),
		mockUserRepo, mockContractsRepo, mockTransferRecordsRepo, mockVoucherRepo, mockTransRepo,
		gridClient, nil, dummyMailService, nil, nil,
	)

	err := service.DeleteUserByID(999)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "user not found")
}

// Test 6: GenerateVouchers - SUCCESS
func TestAdminService_GenerateVouchers_Success(t *testing.T) {
	mockUserRepo := new(mockUserRepo)
	mockContractsRepo := new(mockContractDataRepo)
	mockTransferRecordsRepo := new(mockTransferRecordRepo)
	mockVoucherRepo := new(mockVoucherRepo)
	mockTransRepo := new(mockTransactionRepo)

	// Setup expectation for CreateVoucher calls
	mockVoucherRepo.On("CreateVoucher", mock.MatchedBy(func(v *models.Voucher) bool {
		return v != nil && v.Value == 100.0
	})).Return(nil)

	var gridClient gridclient.GridClient

	service := NewAdminService(
		context.Background(),
		mockUserRepo, mockContractsRepo, mockTransferRecordsRepo, mockVoucherRepo, mockTransRepo,
		gridClient, nil, dummyMailService, nil, nil,
	)

	vouchers, err := service.GenerateVouchers(5, 30, 100.0)

	require.NoError(t, err)
	assert.Len(t, vouchers, 5)
	assert.Equal(t, 100.0, vouchers[0].Value)
	mockVoucherRepo.AssertNumberOfCalls(t, "CreateVoucher", 5)
}

// Test 7: GenerateVouchers - ZERO COUNT
func TestAdminService_GenerateVouchers_ZeroCount(t *testing.T) {
	mockUserRepo := new(mockUserRepo)
	mockContractsRepo := new(mockContractDataRepo)
	mockTransferRecordsRepo := new(mockTransferRecordRepo)
	mockVoucherRepo := new(mockVoucherRepo)
	mockTransRepo := new(mockTransactionRepo)

	var gridClient gridclient.GridClient

	service := NewAdminService(
		context.Background(),
		mockUserRepo, mockContractsRepo, mockTransferRecordsRepo, mockVoucherRepo, mockTransRepo,
		gridClient, nil, dummyMailService, nil, nil,
	)

	vouchers, err := service.GenerateVouchers(0, 30, 100.0)

	require.NoError(t, err)
	assert.Len(t, vouchers, 0)
}

// Test 8: GenerateVouchers - LARGE COUNT
func TestAdminService_GenerateVouchers_LargeCount(t *testing.T) {
	mockUserRepo := new(mockUserRepo)
	mockContractsRepo := new(mockContractDataRepo)
	mockTransferRecordsRepo := new(mockTransferRecordRepo)
	mockVoucherRepo := new(mockVoucherRepo)
	mockTransRepo := new(mockTransactionRepo)

	// Setup expectation for CreateVoucher calls - 100 times
	mockVoucherRepo.On("CreateVoucher", mock.MatchedBy(func(v *models.Voucher) bool {
		return v != nil && v.Value == 50.0
	})).Return(nil)

	var gridClient gridclient.GridClient

	service := NewAdminService(
		context.Background(),
		mockUserRepo, mockContractsRepo, mockTransferRecordsRepo, mockVoucherRepo, mockTransRepo,
		gridClient, nil, dummyMailService, nil, nil,
	)

	vouchers, err := service.GenerateVouchers(100, 30, 50.0)

	require.NoError(t, err)
	assert.Len(t, vouchers, 100)
	assert.NotEmpty(t, vouchers[0].Code)
	mockVoucherRepo.AssertNumberOfCalls(t, "CreateVoucher", 100)
}
