package services

import (
	"context"
	"fmt"
	"kubecloud/internal/constants"
	"kubecloud/internal/statemanager"
	"kubecloud/kubedeployer"
	"kubecloud/models"
	"os"
	"time"

	"github.com/xmonader/ewf"
)

type DeploymentService struct {
	clusterRepo models.ClusterRepository
	userRepo    models.UserRepository

	appCtx    context.Context
	ewfEngine *ewf.Engine

	// configs
	debug             bool
	sshPublicKey      string
	sshPrivateKeyPath string
	systemNetwork     string
}

func NewDeploymentService(appCtx context.Context,
	clusterRepo models.ClusterRepository, userRepo models.UserRepository,
	userNodesRepo models.UserNodesRepository, ewfEngine *ewf.Engine,
	debug bool, sshPublicKey, sshPrivateKeyPath, systemNetwork string,
) DeploymentService {
	return DeploymentService{
		clusterRepo: clusterRepo,
		userRepo:    userRepo,

		appCtx:    appCtx,
		ewfEngine: ewfEngine,

		debug:             debug,
		sshPublicKey:      sshPublicKey,
		sshPrivateKeyPath: sshPrivateKeyPath,
		systemNetwork:     systemNetwork,
	}
}

type ClusterData struct {
	ID          int                  `json:"id"`
	ProjectName string               `json:"project_name"`
	Cluster     kubedeployer.Cluster `json:"cluster"`
	CreatedAt   time.Time            `json:"created_at"`
	UpdatedAt   time.Time            `json:"updated_at"`
}

func (svc *DeploymentService) GetClusterByName(userID int, projectName string) (models.Cluster, error) {
	return svc.clusterRepo.GetClusterByName(userID, projectName)
}

func (svc *DeploymentService) ListUserClusters(userID int) ([]models.Cluster, error) {
	return svc.clusterRepo.ListUserClusters(userID)
}

func (svc *DeploymentService) GetClusterDataByProjectName(userID int, projectName string) (ClusterData, error) {
	cluster, err := svc.clusterRepo.GetClusterByName(userID, projectName)
	if err != nil {
		return ClusterData{}, err
	}

	return svc.GetClusterData(cluster)
}

func (svc *DeploymentService) ListUserClustersData(userID int) ([]ClusterData, error) {
	clusters, err := svc.clusterRepo.ListUserClusters(userID)
	if err != nil {
		return nil, err
	}

	clusterData := make([]ClusterData, 0, len(clusters))
	for _, cluster := range clusters {
		clusterDataItem, err := svc.GetClusterData(cluster)
		if err != nil {
			return nil, err
		}
		clusterData = append(clusterData, clusterDataItem)
	}

	return clusterData, nil
}

func (svc *DeploymentService) GetClusterData(cluster models.Cluster) (ClusterData, error) {
	clusterResult, err := cluster.GetClusterResult()
	if err != nil {
		return ClusterData{}, err
	}

	return ClusterData{
		ID:          cluster.ID,
		ProjectName: cluster.ProjectName,
		Cluster:     clusterResult,
		CreatedAt:   cluster.CreatedAt,
		UpdatedAt:   cluster.UpdatedAt,
	}, nil
}

func (svc *DeploymentService) GetClusterKubeconfig(ctx context.Context, cluster *models.Cluster) (string, error) {
	if cluster.Kubeconfig != "" {
		return cluster.Kubeconfig, nil
	}

	clusterResult, err := cluster.GetClusterResult()
	if err != nil {
		return "", err
	}

	privateKeyBytes, err := os.ReadFile(svc.sshPrivateKeyPath)
	if err != nil {
		return "", err
	}

	kubeconfig, err := clusterResult.GetKubeconfig(ctx, string(privateKeyBytes))
	if err != nil {
		return "", err
	}

	cluster.Kubeconfig = kubeconfig
	if err = svc.clusterRepo.UpdateCluster(cluster); err != nil {
		return "", err
	}

	return kubeconfig, nil
}

func (svc *DeploymentService) GetClientConfig(userID int) (statemanager.ClientConfig, error) {
	user, err := svc.userRepo.GetUserByID(userID)
	if err != nil {
		return statemanager.ClientConfig{}, fmt.Errorf("failed to get user: %v", err)
	}

	return statemanager.ClientConfig{
		SSHPublicKey: svc.sshPublicKey,
		Mnemonic:     user.Mnemonic,
		UserID:       userID,
		Network:      svc.systemNetwork,
		Debug:        svc.debug,
	}, nil
}

func (svc *DeploymentService) AsyncDeployCluster(config statemanager.ClientConfig, cluster kubedeployer.Cluster) (string, ewf.WorkflowStatus, error) {
	wf, err := svc.ewfEngine.NewWorkflow(constants.WorkflowDeployCluster)
	if err != nil {
		return "", "", err
	}

	wf.State = ewf.State{
		"config":  config,
		"cluster": cluster,
	}

	svc.ewfEngine.RunAsync(svc.appCtx, wf)
	return wf.UUID, wf.Status, nil
}

func (svc *DeploymentService) AsyncDeleteCluster(config statemanager.ClientConfig, projectName string) (string, ewf.WorkflowStatus, error) {
	wf, err := svc.ewfEngine.NewWorkflow(constants.WorkflowDeleteCluster)
	if err != nil {
		return "", "", err
	}

	wf.State = ewf.State{
		"config":       config,
		"project_name": projectName,
	}

	svc.ewfEngine.RunAsync(svc.appCtx, wf)
	return wf.UUID, wf.Status, nil
}

func (h *DeploymentService) AsyncDeleteAllClusters(config statemanager.ClientConfig) (string, ewf.WorkflowStatus, error) {
	wf, err := h.ewfEngine.NewWorkflow(constants.WorkflowDeleteAllClusters)
	if err != nil {
		return "", "", err
	}

	wf.State = ewf.State{
		"config": config,
	}

	h.ewfEngine.RunAsync(h.appCtx, wf)
	return wf.UUID, wf.Status, nil
}

func (svc *DeploymentService) AsyncAddNode(config statemanager.ClientConfig, cl kubedeployer.Cluster, node kubedeployer.Node) (string, ewf.WorkflowStatus, error) {
	wf, err := svc.ewfEngine.NewWorkflow(constants.WorkflowAddNode)
	if err != nil {
		return "", "", err
	}

	wf.State = ewf.State{
		"config":  config,
		"cluster": cl,
		"node":    node,
	}

	svc.ewfEngine.RunAsync(svc.appCtx, wf)
	return wf.UUID, wf.Status, nil
}

func (h *DeploymentService) AsyncRemoveNode(config statemanager.ClientConfig, cl kubedeployer.Cluster, nodeName string) (string, ewf.WorkflowStatus, error) {
	wf, err := h.ewfEngine.NewWorkflow(constants.WorkflowRemoveNode)
	if err != nil {
		return "", "", err
	}

	wf.State = ewf.State{
		"config":    config,
		"cluster":   cl,
		"node_name": nodeName,
	}

	h.ewfEngine.RunAsync(h.appCtx, wf)
	return wf.UUID, wf.Status, nil
}
