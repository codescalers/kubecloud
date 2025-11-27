package services

import (
	"context"
	"errors"
	"fmt"
	cfg "kubecloud/internal/config"
	distributedlocks "kubecloud/internal/core/distributed_locks"
	"kubecloud/internal/core/models"
	"kubecloud/internal/core/persistence"
	"kubecloud/internal/core/workflows"
	"kubecloud/internal/deployment/kubedeployer"
	"kubecloud/internal/deployment/statemanager"
	"kubecloud/internal/infrastructure/telemetry"
	"os"
	"strconv"
	"time"

	"github.com/xmonader/ewf"
	"go.opentelemetry.io/otel/attribute"
)

type DeploymentService struct {
	clusterRepo models.ClusterRepository
	userRepo    models.UserRepository

	appCtx    context.Context
	ewfEngine *ewf.Engine
	tracer    *telemetry.ServiceTracer

	locker distributedlocks.DistributedLocks

	// configs
	debug             bool
	sshPublicKey      string
	sshPrivateKeyPath string
	systemNetwork     string
}

func NewDeploymentService(appCtx context.Context,
	clusterRepo models.ClusterRepository, userRepo models.UserRepository,
	userNodesRepo models.UserNodesRepository, ewfEngine *ewf.Engine,
	locker distributedlocks.DistributedLocks,
	debug bool, sshPublicKey, sshPrivateKeyPath, systemNetwork string,
) DeploymentService {
	return DeploymentService{
		clusterRepo: clusterRepo,
		userRepo:    userRepo,

		appCtx:    appCtx,
		ewfEngine: ewfEngine,
		tracer:    telemetry.NewServiceTracer("deployment_service"),

		locker: locker,

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
	_, span := svc.tracer.StartSpan(context.Background(), "ListUserClustersData")
	defer span.End()

	span.SetAttributes(attribute.Int("user_id", userID))

	clusters, err := svc.clusterRepo.ListUserClusters(userID)
	if err != nil {
		telemetry.RecordError(span, err)
		return nil, err
	}

	clusterData := make([]ClusterData, 0, len(clusters))
	for _, cluster := range clusters {
		clusterDataItem, err := svc.GetClusterData(cluster)
		if err != nil {
			telemetry.RecordError(span, err)
			return nil, err
		}
		clusterData = append(clusterData, clusterDataItem)
	}

	span.SetAttributes(attribute.Int("cluster_count", len(clusterData)))
	return clusterData, nil
}

func (svc *DeploymentService) GetClusterData(cluster models.Cluster) (ClusterData, error) {
	_, span := svc.tracer.StartSpan(context.Background(), "GetClusterData")
	defer span.End()

	span.SetAttributes(
		attribute.Int("cluster_id", cluster.ID),
		attribute.String("project_name", cluster.ProjectName),
	)

	clusterResult, err := cluster.GetClusterResult()
	if err != nil {
		telemetry.RecordError(span, err)
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
	ctx, span := svc.tracer.StartSpan(ctx, "GetClusterKubeconfig")
	defer span.End()

	span.SetAttributes(
		attribute.Int("cluster_id", cluster.ID),
		attribute.String("project_name", cluster.ProjectName),
	)

	if cluster.Kubeconfig != "" {
		span.SetAttributes(attribute.Bool("cached_kubeconfig", true))
		return cluster.Kubeconfig, nil
	}

	span.SetAttributes(attribute.Bool("cached_kubeconfig", false))

	clusterResult, err := cluster.GetClusterResult()
	if err != nil {
		telemetry.RecordError(span, err)
		return "", err
	}

	privateKeyBytes, err := os.ReadFile(svc.sshPrivateKeyPath)
	if err != nil {
		telemetry.RecordError(span, err)
		return "", err
	}

	kubeconfig, err := clusterResult.GetKubeconfig(ctx, string(privateKeyBytes))
	if err != nil {
		telemetry.RecordError(span, err)
		return "", err
	}

	cluster.Kubeconfig = kubeconfig
	if err = svc.clusterRepo.UpdateCluster(cluster); err != nil {
		telemetry.RecordError(span, err)
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

// runWithQueue ensures the workflow is run within the specified queue
// if a queue with the given name does not exist, it creates one
// if a non queued workflow is passed, it sets its queue name to run in the specified queue
func (svc *DeploymentService) runWithQueue(queueName string, wf *ewf.Workflow) error {

	err := svc.ewfEngine.CreateQueue(svc.appCtx, queueName, cfg.DefaultQueueConfig.WorkersDef, cfg.DefaultQueueConfig.QueueOptions)
	if err != nil && !errors.Is(err, ewf.ErrQueueAlreadyExists) {
		return err
	}

	if wf.QueueName == "" {
		wf.QueueName = queueName
	}

	return svc.ewfEngine.Run(svc.appCtx, *wf)
}

func (svc *DeploymentService) handleDeploymentAction(userID int, workflowName string, state ewf.State, displayName string, metadata map[string]string) (string, ewf.WorkflowStatus, error) {
	_, span := svc.tracer.StartSpan(context.Background(), "handleDeploymentAction")
	defer span.End()

	span.SetAttributes(
		attribute.Int("user_id", userID),
		attribute.String("workflow_name", workflowName),
		attribute.String("display_name", displayName),
	)

	queueName := fmt.Sprintf("%s:user_%d", cfg.DefaultQueueConfig.Name, userID)

	wf, err := svc.ewfEngine.NewWorkflow(workflowName, ewf.WithQueue(queueName), ewf.WithDisplayName(displayName), ewf.WithMetadata(metadata))
	if err != nil {
		telemetry.RecordError(span, err)
		return "", "", err
	}

	wf.State = state

	if err = persistence.SetStateUserID(&wf, userID); err != nil {
		telemetry.RecordError(span, err)
		return "", "", err
	}

	if err = svc.runWithQueue(queueName, &wf); err != nil {
		telemetry.RecordError(span, err)
		return "", "", err
	}

	span.SetAttributes(
		attribute.String("workflow_uuid", wf.UUID),
		attribute.String("workflow_status", string(wf.Status)),
	)
	return wf.UUID, wf.Status, nil
}

func (svc *DeploymentService) AsyncDeployCluster(config statemanager.ClientConfig, cluster kubedeployer.Cluster) (string, ewf.WorkflowStatus, error) {
	nodeIDs := make([]uint32, 0, len(cluster.Nodes))
	for _, node := range cluster.Nodes {
		nodeIDs = append(nodeIDs, node.NodeID)
	}
	if err := svc.locker.AcquireNodesLocks(svc.appCtx, nodeIDs); err != nil {
		return "", "", err
	}

	state := ewf.State{
		"config":  config,
		"cluster": cluster,
	}

	displayName := fmt.Sprintf("Deploying cluster %s", cluster.Name)
	metadata := map[string]string{
		"cluster_name": cluster.Name,
		"node_count":   strconv.Itoa(len(cluster.Nodes)),
	}
	return svc.handleDeploymentAction(config.UserID, workflows.WorkflowDeployCluster, state, displayName, metadata)
}

func (svc *DeploymentService) AsyncDeleteCluster(config statemanager.ClientConfig, projectName string) (string, ewf.WorkflowStatus, error) {

	state := ewf.State{
		"config":       config,
		"project_name": projectName,
	}

	displayName := fmt.Sprintf("Deleting cluster %s", projectName)
	metadata := map[string]string{
		"project_name": projectName,
	}
	return svc.handleDeploymentAction(config.UserID, workflows.WorkflowDeleteCluster, state, displayName, metadata)
}

func (svc *DeploymentService) AsyncDeleteAllClusters(config statemanager.ClientConfig) (string, ewf.WorkflowStatus, error) {

	state := ewf.State{
		"config": config,
	}

	displayName := "Deleting all user clusters"
	return svc.handleDeploymentAction(config.UserID, workflows.WorkflowDeleteAllClusters, state, displayName, nil)
}

func (svc *DeploymentService) AsyncAddNode(config statemanager.ClientConfig, cl kubedeployer.Cluster, node kubedeployer.Node) (string, ewf.WorkflowStatus, error) {

	if err := svc.locker.AcquireNodesLocks(svc.appCtx, []uint32{node.NodeID}); err != nil {
		return "", "", err
	}

	state := ewf.State{
		"config":  config,
		"cluster": cl,
		"node":    node,
	}
	displayName := fmt.Sprintf("Adding node %s to cluster %s", node.Name, cl.Name)
	metadata := map[string]string{
		"cluster_name": cl.Name,
		"node_name":    node.Name,
	}
	return svc.handleDeploymentAction(config.UserID, workflows.WorkflowAddNode, state, displayName, metadata)
}

func (svc *DeploymentService) AsyncRemoveNode(config statemanager.ClientConfig, cl kubedeployer.Cluster, nodeName string) (string, ewf.WorkflowStatus, error) {

	state := ewf.State{
		"config":    config,
		"cluster":   cl,
		"node_name": nodeName,
	}

	displayName := fmt.Sprintf("Removing node %s from cluster %s", nodeName, cl.Name)
	metadata := map[string]string{
		"cluster_name": cl.Name,
		"node_name":    nodeName,
	}
	return svc.handleDeploymentAction(config.UserID, workflows.WorkflowRemoveNode, state, displayName, metadata)
}
