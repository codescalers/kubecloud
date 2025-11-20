package services

import (
	"fmt"
	"testing"

	"kubecloud/internal/core/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockClusterRepo struct {
	mock.Mock
}

func (m *mockClusterRepo) CreateCluster(userID int, cluster *models.Cluster) error {
	args := m.Called(userID, cluster)
	return args.Error(0)
}

func (m *mockClusterRepo) ListUserClusters(userID int) ([]models.Cluster, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Cluster), args.Error(1)
}

func (m *mockClusterRepo) GetClusterByName(userID int, projectName string) (models.Cluster, error) {
	args := m.Called(userID, projectName)
	if args.Get(0) == nil {
		return models.Cluster{}, args.Error(1)
	}
	return args.Get(0).(models.Cluster), args.Error(1)
}

func (m *mockClusterRepo) UpdateCluster(contractsRepo models.ContractDataRepository, cluster *models.Cluster) error {
	args := m.Called(contractsRepo, cluster)
	return args.Error(0)
}

func (m *mockClusterRepo) DeleteCluster(userID int, projectName string) error {
	args := m.Called(userID, projectName)
	return args.Error(0)
}

func (m *mockClusterRepo) DeleteAllUserClusters(contractsRepo models.ContractDataRepository, userID int) error {
	args := m.Called(contractsRepo, userID)
	return args.Error(0)
}

func (m *mockClusterRepo) CountAllClusters() (int64, error) {
	args := m.Called()
	return args.Get(0).(int64), args.Error(1)
}

func (m *mockClusterRepo) ListAllClusters() ([]models.Cluster, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Cluster), args.Error(1)
}

// Test 1: DeploymentService - GetClusterByName SUCCESS
func TestDeploymentService_GetClusterByName_Success(t *testing.T) {
	mockClusterRepo := new(mockClusterRepo)
	mockUserRepo := new(mockUserRepo)

	cluster := models.Cluster{
		ID:          1,
		UserID:      1,
		ProjectName: "test-project",
	}

	mockClusterRepo.On("GetClusterByName", 1, "test-project").Return(cluster, nil)

	service := DeploymentService{
		clusterRepo:   mockClusterRepo,
		userRepo:      mockUserRepo,
		contractsRepo: new(mockContractDataRepo),
	}

	result, err := service.GetClusterByName(1, "test-project")

	require.NoError(t, err)
	assert.Equal(t, "test-project", result.ProjectName)
	mockClusterRepo.AssertCalled(t, "GetClusterByName", 1, "test-project")
}

// Test 2: DeploymentService - GetClusterByName NOT FOUND
func TestDeploymentService_GetClusterByName_NotFound(t *testing.T) {
	mockClusterRepo := new(mockClusterRepo)
	mockUserRepo := new(mockUserRepo)

	mockClusterRepo.On("GetClusterByName", 1, "nonexistent").Return(models.Cluster{}, fmt.Errorf("cluster not found"))

	service := DeploymentService{
		clusterRepo:   mockClusterRepo,
		userRepo:      mockUserRepo,
		contractsRepo: new(mockContractDataRepo),
	}

	_, err := service.GetClusterByName(1, "nonexistent")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "cluster not found")
}

// Test 3: DeploymentService - ListUserClusters SUCCESS
func TestDeploymentService_ListUserClusters_Success(t *testing.T) {
	mockClusterRepo := new(mockClusterRepo)
	mockUserRepo := new(mockUserRepo)

	clusters := []models.Cluster{
		{
			ID:          1,
			UserID:      1,
			ProjectName: "project1",
		},
		{
			ID:          2,
			UserID:      1,
			ProjectName: "project2",
		},
	}

	mockClusterRepo.On("ListUserClusters", 1).Return(clusters, nil)

	service := DeploymentService{
		clusterRepo:   mockClusterRepo,
		userRepo:      mockUserRepo,
		contractsRepo: new(mockContractDataRepo),
	}

	result, err := service.ListUserClusters(1)

	require.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "project1", result[0].ProjectName)
}

// Test 4: DeploymentService - ListUserClusters EMPTY
func TestDeploymentService_ListUserClusters_Empty(t *testing.T) {
	mockClusterRepo := new(mockClusterRepo)
	mockUserRepo := new(mockUserRepo)

	mockClusterRepo.On("ListUserClusters", 999).Return([]models.Cluster{}, nil)

	service := DeploymentService{
		clusterRepo:   mockClusterRepo,
		userRepo:      mockUserRepo,
		contractsRepo: new(mockContractDataRepo),
	}

	result, err := service.ListUserClusters(999)

	require.NoError(t, err)
	assert.Len(t, result, 0)
}

// Test 5: DeploymentService - ListUserClusters ERROR
func TestDeploymentService_ListUserClusters_Error(t *testing.T) {
	mockClusterRepo := new(mockClusterRepo)
	mockUserRepo := new(mockUserRepo)

	mockClusterRepo.On("ListUserClusters", 1).Return(nil, fmt.Errorf("database error"))

	service := DeploymentService{
		clusterRepo:   mockClusterRepo,
		userRepo:      mockUserRepo,
		contractsRepo: new(mockContractDataRepo),
	}

	_, err := service.ListUserClusters(1)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "database error")
}
