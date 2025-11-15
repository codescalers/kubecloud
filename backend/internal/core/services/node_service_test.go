package services

import (
	"fmt"
	"testing"

	"kubecloud/internal/core/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test 1: NodeService - GetUserNodeByNodeID SUCCESS
func TestNodeService_GetUserNodeByNodeID_Success(t *testing.T) {
	mockNodesRepo := new(mockUserNodesRepo)
	mockUserRepo := new(mockUserRepo)

	node := models.UserNodes{
		ID:         1,
		UserID:     1,
		NodeID:     100,
		ContractID: 1,
	}

	mockNodesRepo.On("GetUserNodeByNodeID", uint64(100)).Return(node, nil)

	service := NodeService{
		nodesRepo: mockNodesRepo,
		userRepo:  mockUserRepo,
	}

	result, err := service.GetUserNodeByNodeID(100)

	require.NoError(t, err)
	assert.Equal(t, uint32(100), result.NodeID)
}

// Test 2: NodeService - GetUserNodeByNodeID NOT FOUND
func TestNodeService_GetUserNodeByNodeID_NotFound(t *testing.T) {
	mockNodesRepo := new(mockUserNodesRepo)
	mockUserRepo := new(mockUserRepo)

	mockNodesRepo.On("GetUserNodeByNodeID", uint64(999)).Return(models.UserNodes{}, fmt.Errorf("node not found"))

	service := NodeService{
		nodesRepo: mockNodesRepo,
		userRepo:  mockUserRepo,
	}

	_, err := service.GetUserNodeByNodeID(999)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "node not found")
}

// Test 3: NodeService - GetUserNodeByContractID SUCCESS
func TestNodeService_GetUserNodeByContractID_Success(t *testing.T) {
	mockNodesRepo := new(mockUserNodesRepo)
	mockUserRepo := new(mockUserRepo)

	node := models.UserNodes{
		ID:         1,
		UserID:     1,
		NodeID:     100,
		ContractID: 123,
	}

	mockNodesRepo.On("GetUserNodeByContractID", uint64(123)).Return(node, nil)

	service := NodeService{
		nodesRepo: mockNodesRepo,
		userRepo:  mockUserRepo,
	}

	result, err := service.GetUserNodeByContractID(123)

	require.NoError(t, err)
	assert.Equal(t, uint64(123), result.ContractID)
}

// Test 4: NodeService - GetUserNodeByContractID ERROR
func TestNodeService_GetUserNodeByContractID_Error(t *testing.T) {
	mockNodesRepo := new(mockUserNodesRepo)
	mockUserRepo := new(mockUserRepo)

	mockNodesRepo.On("GetUserNodeByContractID", uint64(999)).Return(models.UserNodes{}, fmt.Errorf("contract error"))

	service := NodeService{
		nodesRepo: mockNodesRepo,
		userRepo:  mockUserRepo,
	}

	_, err := service.GetUserNodeByContractID(999)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "contract error")
}

// Test 5: NodeService - GetUserByID SUCCESS
func TestNodeService_GetUserByID_Success(t *testing.T) {
	mockNodesRepo := new(mockUserNodesRepo)
	mockUserRepo := new(mockUserRepo)

	user := models.User{
		ID:       1,
		Email:    "user@example.com",
		Username: "testuser",
	}

	mockUserRepo.On("GetUserByID", 1).Return(user, nil)

	service := NodeService{
		nodesRepo: mockNodesRepo,
		userRepo:  mockUserRepo,
	}

	result, err := service.GetUserByID(1)

	require.NoError(t, err)
	assert.Equal(t, "user@example.com", result.Email)
}

// Test 6: NodeService - GetUserByID NOT FOUND
func TestNodeService_GetUserByID_NotFound(t *testing.T) {
	mockNodesRepo := new(mockUserNodesRepo)
	mockUserRepo := new(mockUserRepo)

	mockUserRepo.On("GetUserByID", 999).Return(models.User{}, fmt.Errorf("user not found"))

	service := NodeService{
		nodesRepo: mockNodesRepo,
		userRepo:  mockUserRepo,
	}

	_, err := service.GetUserByID(999)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "user not found")
}
