package activities

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"kubecloud/internal"
	"kubecloud/internal/constants"
	"kubecloud/internal/metrics"
	"kubecloud/internal/notification"
	"kubecloud/internal/statemanager"
	"kubecloud/kubedeployer"
	"kubecloud/models"
	"os"
	"strings"
	"time"

	"kubecloud/internal/logger"

	"github.com/xmonader/ewf"
	"gorm.io/gorm"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	criticalRetryPolicy        = &ewf.RetryPolicy{MaxAttempts: 5, BackOff: ewf.ConstantBackoff(5 * time.Second)}
	standardRetryPolicy        = &ewf.RetryPolicy{MaxAttempts: 2, BackOff: ewf.ConstantBackoff(2 * time.Second)}
	longExponentialRetryPolicy = &ewf.RetryPolicy{MaxAttempts: 5, BackOff: ewf.ExponentialBackoff(30*time.Second, 5*time.Minute, 2.0)}
)

func isWorkloadAlreadyDeployedError(err error) bool {
	errMsg := err.Error()
	return strings.Contains(errMsg, "exists: conflict")
}

func isWorkloadInvalid(err error) bool {
	errMsg := err.Error()
	return strings.Contains(errMsg, "invalid deployment")
}

func ensureClient(state ewf.State) {
	log := logger.ForOperation("deployer_activities", "ensure_client")
	// Get config first
	config, err := getConfig(state)
	if err != nil {
		log.Error().Err(err).Msg("failed to get config")
		return
	}

	// Use the statemanager to get or create client
	_, err = statemanager.GetKubeClient(state, config)
	if err != nil {
		log.Error().Err(err).Int("user_id", config.UserID).Msg("failed to ensure kubeclient")
		return
	}

	log.Debug().Msg("kubeclient ensured and ready for use")
}

func DeployNetworkStep(metrics *metrics.Metrics) ewf.StepFn {
	return func(ctx context.Context, state ewf.State) error {
		ensureClient(state)

		config, err := getConfig(state)
		if err != nil {
			return fmt.Errorf("failed to get config from state: %w", err)
		}

		kubeClient, err := statemanager.GetKubeClient(state, config)
		if err != nil {
			return err
		}

		cluster, err := statemanager.GetCluster(state)
		if err != nil {
			return err
		}

		if cluster.ProjectName == "" {
			// this is a first not a retry
			if err := cluster.PrepareCluster(config.UserID); err != nil {
				return fmt.Errorf("failed to prepare cluster: %w", err)
			}
		}
		statemanager.StoreCluster(state, cluster)

		if err := kubeClient.DeployNetwork(ctx, &cluster); err != nil {
			nodeIDs := make([]uint32, 0, len(cluster.Nodes))
			for _, node := range cluster.Nodes {
				nodeIDs = append(nodeIDs, node.NodeID)
			}

			if isWorkloadAlreadyDeployedError(err) {
				return fmt.Errorf("network already deployed for cluster %s (user_id=%d, node_ids=%v): %w", cluster.Name, config.UserID, nodeIDs, ewf.ErrFailWorkflowNow)
			}
			if isWorkloadInvalid(err) {
				return fmt.Errorf("network invalid for cluster %s (user_id=%d, node_ids=%v): %w", cluster.Name, config.UserID, nodeIDs, ewf.ErrFailWorkflowNow)
			}
			return fmt.Errorf("failed to deploy network for cluster %s (user_id=%d, node_ids=%v): %w", cluster.Name, config.UserID, nodeIDs, err)
		}

		// Save GridClient state after network deployment
		statemanager.SaveGridClientState(state, kubeClient)
		statemanager.StoreCluster(state, cluster)
		return nil
	}
}

func UpdateNetworkStep(metrics *metrics.Metrics) ewf.StepFn {
	return func(ctx context.Context, state ewf.State) error {
		ensureClient(state)

		config, err := getConfig(state)
		if err != nil {
			return fmt.Errorf("failed to get config from state: %w", err)
		}

		kubeClient, err := statemanager.GetKubeClient(state, config)
		if err != nil {
			return err
		}

		cluster, err := statemanager.GetCluster(state)
		if err != nil {
			return fmt.Errorf("failed to get cluster from state while updating network: %w", err)
		}

		node, err := getFromState[kubedeployer.Node](state, "node")
		if err != nil {
			return err
		}

		node.Name = kubedeployer.GetNodeName(config.UserID, cluster.Name, node.OriginalName)
		cluster.Nodes = append(cluster.Nodes, node)

		if err := kubeClient.DeployNetwork(ctx, &cluster); err != nil {
			return fmt.Errorf("failed to update network for cluster %s, node %s (user_id=%d): %w", cluster.Name, node.Name, config.UserID, err)
		}

		// Save GridClient state after network update
		statemanager.SaveGridClientState(state, kubeClient)
		statemanager.StoreCluster(state, cluster)
		state["node"] = node
		return nil
	}
}

func AddNodeStep(metrics *metrics.Metrics) ewf.StepFn {
	return func(ctx context.Context, state ewf.State) error {
		ensureClient(state)

		config, err := getConfig(state)
		if err != nil {
			return fmt.Errorf("failed to get config from state: %w", err)
		}

		kubeClient, err := statemanager.GetKubeClient(state, config)
		if err != nil {
			return err
		}

		cluster, err := statemanager.GetCluster(state)
		if err != nil {
			return err
		}

		node, err := getFromState[kubedeployer.Node](state, "node")
		if err != nil {
			return err
		}

		if err := node.AssignNodeIP(ctx, kubeClient.GridClient, cluster.Network.Name); err != nil {
			return fmt.Errorf("failed to assign IP for node %s, cluster %s (user_id=%d): %w", node.Name, cluster.Name, config.UserID, err)
		}

		if err := kubeClient.DeployNode(ctx, &cluster, node, config.SSHPublicKey); err != nil {
			return fmt.Errorf("failed to deploy node %s to cluster %s (user_id=%d): %w", node.Name, cluster.Name, config.UserID, err)
		}

		// Save GridClient state after node deployment
		statemanager.SaveGridClientState(state, kubeClient)
		statemanager.StoreCluster(state, cluster)

		// Store the deployed node in state for verification step
		state["node"] = node

		return nil
	}
}

func DeployLeaderNodeStep(metrics *metrics.Metrics) ewf.StepFn {
	return func(ctx context.Context, state ewf.State) error {
		log := logger.ForOperation("deployer_activities", "deploy_leader_node")
		ensureClient(state)

		config, err := getConfig(state)
		if err != nil {
			return fmt.Errorf("failed to get config from state: %w", err)
		}

		kubeClient, err := statemanager.GetKubeClient(state, config)
		if err != nil {
			return err
		}

		cluster, err := statemanager.GetCluster(state)
		if err != nil {
			return err
		}

		leaderNode := cluster.Nodes[0]
		if leaderNode.ContractID != 0 {
			log.Debug().Str("node_name", leaderNode.Name).Uint64("contract_id", leaderNode.ContractID).Msg("Leader node already deployed, skipping")
			return nil
		}

		log.Debug().Str("node_name", leaderNode.Name).Msg("Deploying leader node")

		if err := leaderNode.AssignNodeIP(ctx, kubeClient.GridClient, cluster.Network.Name); err != nil {
			return fmt.Errorf("failed to assign IP to leader node: %w", err)
		}
		cluster.Nodes[0].IP = leaderNode.IP

		if err := kubeClient.DeployNode(ctx, &cluster, leaderNode, config.SSHPublicKey); err != nil {
			if isWorkloadAlreadyDeployedError(err) {
				return fmt.Errorf("leader node already deployed for cluster %s: %w", cluster.Name, ewf.ErrFailWorkflowNow)
			}
			if isWorkloadInvalid(err) {
				return fmt.Errorf("leader node invalid for cluster %s: %w", cluster.Name, ewf.ErrFailWorkflowNow)
			}
			return fmt.Errorf("failed to deploy leader node: %w", err)
		}

		log.Debug().Str("node_name", leaderNode.Name).Msg("Leader node deployed successfully")

		statemanager.SaveGridClientState(state, kubeClient)
		statemanager.StoreCluster(state, cluster)
		return nil
	}
}

func BatchDeployAllNodesStep(metrics *metrics.Metrics) ewf.StepFn {
	return func(ctx context.Context, state ewf.State) error {
		log := logger.ForOperation("deployer_activities", "batch_deploy_all_nodes")
		ensureClient(state)

		config, err := getConfig(state)
		if err != nil {
			return fmt.Errorf("failed to get config from state: %w", err)
		}

		kubeClient, err := statemanager.GetKubeClient(state, config)
		if err != nil {
			return err
		}

		cluster, err := statemanager.GetCluster(state)
		if err != nil {
			return err
		}

		// Collect all non-leader nodes that need deployment (ContractID == 0)
		var nodesToDeploy []kubedeployer.Node
		var nodeIndices []int
		for i, node := range cluster.Nodes {
			if node.Type != kubedeployer.NodeTypeLeader && node.ContractID == 0 {
				nodesToDeploy = append(nodesToDeploy, node)
				nodeIndices = append(nodeIndices, i)
			}
		}

		if len(nodesToDeploy) == 0 {
			log.Debug().Msg("No nodes to deploy, all nodes already deployed")
			return nil
		}

		for i, node := range nodesToDeploy {
			if err := node.AssignNodeIP(ctx, kubeClient.GridClient, cluster.Network.Name); err != nil {
				metrics.IncrementClusterDeploymentFailure()
				return fmt.Errorf("failed to assign IP to node %s: %w", node.Name, err)
			}
			nodesToDeploy[i].IP = node.IP
			cluster.Nodes[nodeIndices[i]].IP = node.IP
		}

		batchErr := kubeClient.BatchDeployNodes(ctx, &cluster, nodesToDeploy, config.SSHPublicKey)

		statemanager.SaveGridClientState(state, kubeClient)
		statemanager.StoreCluster(state, cluster)

		if batchErr != nil {
			metrics.IncrementClusterDeploymentFailure()
			return fmt.Errorf("failed to batch deploy nodes: %w", batchErr)
		}
		log.Debug().Int("count", len(nodesToDeploy)).Msg("All nodes deployed successfully")
		metrics.IncrementClusterDeploymentSuccess()
		return nil
	}
}

func StoreDeploymentStep(db models.DB, metrics *metrics.Metrics, fileStorage *internal.FileStorageService, fileStorage *internal.FileStorageService) ewf.StepFn {
	return func(ctx context.Context, state ewf.State) error {
		log := logger.ForOperation("deployer_activities", "store_deployment")
		cluster, err := statemanager.GetCluster(state)
		if err != nil {
			return err
		}

		config, err := getConfig(state)
		if err != nil {
			return err
		}

		dbCluster := &models.Cluster{
			ProjectName: cluster.ProjectName,
		}

		if err := dbCluster.SetClusterResult(cluster); err != nil {
			return fmt.Errorf("failed to set cluster result for %s (user_id=%d): %w", cluster.Name, config.UserID, err)
		}

		existingCluster, err := db.GetClusterByName(config.UserID, cluster.ProjectName)

		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("failed to get cluster by name: %w", err)
		}

		persistedCluster := dbCluster

		if errors.Is(err, gorm.ErrRecordNotFound) {
			if err := db.CreateCluster(config.UserID, dbCluster); err != nil {
				return fmt.Errorf("failed to create cluster in database: %w", err)
			}

		} else {
		persistedCluster := dbCluster
		if err == nil {
			existingCluster.Result = dbCluster.Result
			if err := db.UpdateCluster(&existingCluster); err != nil {
				return fmt.Errorf("failed to update cluster %s in database (user_id=%d): %w", cluster.Name, config.UserID, err)
			}
		}

		metrics.IncActiveClusterCount()

		return nil
	}
}

func CancelDeploymentStep(db models.DB, metrics *metrics.Metrics) ewf.StepFn {
	return func(ctx context.Context, state ewf.State) error {
		ensureClient(state)

		config, err := getConfig(state)
		if err != nil {
			return fmt.Errorf("failed to get config from state: %w", err)
		}

		kubeClient, err := statemanager.GetKubeClient(state, config)
		if err != nil {
			return err
		}

		// in a Rollaback, cluster is in state, in a delete, we need to load from db
		cluster, err := statemanager.GetCluster(state)
		if err != nil {
			projectName, ok := state["project_name"].(string)
			if !ok {
				return fmt.Errorf("missing or invalid 'project_name' in state")
			}

			dbCluster, err := db.GetClusterByName(config.UserID, projectName)
			if err != nil {
				return fmt.Errorf("failed to get cluster %s from database (user_id=%d): %w", projectName, config.UserID, err)
			}

			cluster, err = dbCluster.GetClusterResult()
			if err != nil {
				return fmt.Errorf("failed to get cluster result for %s (user_id=%d): %w", projectName, config.UserID, err)
			}
			state["cluster"] = cluster
		}

		if err := kubeClient.CancelCluster(ctx, cluster); err != nil {
			return fmt.Errorf("failed to cancel deployment for cluster %s (user_id=%d): %w", cluster.Name, config.UserID, err)
		}

		return nil
	}
}

func RemoveClusterFromDBStep(db models.DB, fileStorage *internal.FileStorageService, fileStorage *internal.FileStorageService, metrics *metrics.Metrics) ewf.StepFn {
	return func(ctx context.Context, state ewf.State) error {
		config, err := getConfig(state)
		if err != nil {
			return err
		}

		projectName, ok := state["project_name"].(string)
		if !ok {
			return fmt.Errorf("missing or invalid 'project_name' in state")
		}

		cluster, err := db.GetClusterByName(config.UserID, projectName)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("failed to get cluster by name: %w", err)
		}
		if cluster.ID != 0 {
			if err := fileStorage.DeleteKubeconfigFile(config.UserID, cluster.ID, projectName); err != nil {
				logger.GetLogger().Error().Err(err).Int("user_id", config.UserID).Int("cluster_id", cluster.ID).Str("project_name", projectName).Msg("Failed to delete kubeconfig during cluster removal")
			}
		}

		if err := db.DeleteCluster(config.UserID, projectName); err != nil {
			return fmt.Errorf("failed to delete cluster %s from database (user_id=%d): %w", projectName, config.UserID, err)
		}

		metrics.DecActiveClusterCount()
		metrics.IncrementClusterOperationSuccess(constants.ClusterOperationDeleteCluster)
		return nil
	}
}

func GatherAllContractIDsStep(db models.DB) ewf.StepFn {
	return func(ctx context.Context, state ewf.State) error {
		log := logger.ForOperation("deployer_activities", "gather_all_contract_ids")
		config, err := getConfig(state)
		if err != nil {
			return err
		}

		clusters, err := db.ListUserClusters(config.UserID)
		if err != nil {
			return fmt.Errorf("failed to list user clusters (user_id=%d): %w", config.UserID, err)
		}

		var allContractIDs []uint64
		for _, cluster := range clusters {
			clusterResult, err := cluster.GetClusterResult()
			if err != nil {
				log.Error().Err(err).Int("cluster_id", cluster.ID).Str("project_name", cluster.ProjectName).Int("user_id", config.UserID).Msg("Failed to deserialize cluster result")
				continue
			}

			// Gather contract IDs from all nodes
			for _, node := range clusterResult.Nodes {
				if node.ContractID != 0 {
					allContractIDs = append(allContractIDs, node.ContractID)
				}
			}

			// Gather contract IDs from network deployments
			for _, contractID := range clusterResult.Network.NodeDeploymentID {
				if contractID != 0 {
					allContractIDs = append(allContractIDs, contractID)
				}
			}
		}

		// Remove duplicates
		contractIDSet := make(map[uint64]bool)
		var uniqueContractIDs []uint64
		for _, id := range allContractIDs {
			if !contractIDSet[id] {
				contractIDSet[id] = true
				uniqueContractIDs = append(uniqueContractIDs, id)
			}
		}

		state["contract_ids"] = uniqueContractIDs
		return nil
	}
}

func BatchCancelContractsStep() ewf.StepFn {
	return func(ctx context.Context, state ewf.State) error {
		log := logger.ForOperation("deployer_activities", "batch_cancel_contracts")
		ensureClient(state)

		config, err := getConfig(state)
		if err != nil {
			return fmt.Errorf("failed to get config from state: %w", err)
		}

		kubeClient, err := statemanager.GetKubeClient(state, config)
		if err != nil {
			return err
		}

		contractIDs, ok := state["contract_ids"].([]uint64)
		if !ok {
			return fmt.Errorf("missing or invalid 'contract_ids' in state")
		}

		if len(contractIDs) == 0 {
			log.Info().Int("user_id", config.UserID).Msg("No contracts to cancel")
			return nil
		}

		if err := kubeClient.CancelAllContractsForUser(ctx, contractIDs); err != nil {
			return fmt.Errorf("failed to cancel %d contracts (user_id=%d): %w", len(contractIDs), config.UserID, err)
		}

		return nil
	}
}

func DeleteAllUserClustersStep(db models.DB, fileStorage *internal.FileStorageService, fileStorage *internal.FileStorageService, metrics *metrics.Metrics) ewf.StepFn {
	return func(ctx context.Context, state ewf.State) error {
		config, err := getConfig(state)
		if err != nil {
			return err
		}

		// List clusters to capture IDs and names for kubeconfig cleanup
		clusters, err := db.ListUserClusters(config.UserID)
		if err != nil {
			return fmt.Errorf("failed to list user clusters for deletion: %w", err)
		}
		for _, c := range clusters {
			if err := fileStorage.DeleteKubeconfigFile(config.UserID, c.ID, c.ProjectName); err != nil {
				logger.GetLogger().Error().Err(err).Int("user_id", config.UserID).Int("cluster_id", c.ID).Str("project_name", c.ProjectName).Msg("Failed to delete kubeconfig during bulk cluster removal")
			}
		}

		clusterCount := len(clusters)

		if err := db.DeleteAllUserClusters(config.UserID); err != nil {
			return fmt.Errorf("failed to delete all user clusters from database (user_id=%d): %w", config.UserID, err)
		}

		// Decrement by the actual number of clusters deleted
		metrics.SubActiveClusterCount(clusterCount)
		metrics.IncrementClusterOperationSuccess(constants.ClusterOperationDeleteAllClusters)
		return nil
	}
}

func RemoveDeploymentNodeStep() ewf.StepFn {
	return func(ctx context.Context, state ewf.State) error {
		ensureClient(state)

		config, err := getConfig(state)
		if err != nil {
			return fmt.Errorf("failed to get config from state: %w", err)
		}

		kubeClient, err := statemanager.GetKubeClient(state, config)
		if err != nil {
			return err
		}

		existingCluster, err := statemanager.GetCluster(state)
		if err != nil {
			return err
		}

		nodeName, ok := state["node_name"].(string)
		if !ok {
			return fmt.Errorf("missing or invalid 'node_name' in state")
		}

		nodeName = kubedeployer.GetNodeName(config.UserID, existingCluster.Name, nodeName)

		if err := kubeClient.RemoveNode(ctx, &existingCluster, nodeName); err != nil {
			return fmt.Errorf("failed to remove node %s from cluster %s (user_id=%d): %w", nodeName, existingCluster.Name, config.UserID, err)
		}

		// Save GridClient state after node removal
		statemanager.SaveGridClientState(state, kubeClient)
		statemanager.StoreCluster(state, existingCluster)
		return nil
	}
}

func closeClient(ctx context.Context, wf *ewf.Workflow, err error) {
	log := logger.ForOperation("deployer_activities", "close_client").With().Str("workflow_name", wf.Name).Logger()
	if kubeClient, ok := wf.State["kubeclient"].(*kubedeployer.Client); ok {
		// Save final GridClient state before closing
		statemanager.SaveGridClientState(wf.State, kubeClient)

		kubeClient.Close()
		delete(wf.State, "kubeclient")
	} else {
		log.Warn().Msg("No kubeclient found in workflow state to close")
	}

}

func deploymentFailureHook(engine *ewf.Engine) ewf.AfterWorkflowHook {
	return func(ctx context.Context, wf *ewf.Workflow, err error) {
		log := logger.ForOperation("deployer_activities", "deployment_failure_hook").With().Str("workflow_name", wf.Name).Logger()
		if err != nil && isDeployWorkflow(wf.Name) {
			cluster, clusterErr := statemanager.GetCluster(wf.State)
			if clusterErr != nil || cluster.ProjectName == "" {
				log.Error().Err(clusterErr).Msg("nothing to rollback")
				return
			}

			log.Info().Str("project_name", cluster.ProjectName).Msg("triggering rollback workflow for failed deployment")

			rollbackWf, rollbackErr := engine.NewWorkflow(constants.WorkflowRollbackFailedDeployment)
			if rollbackErr != nil {
				log.Error().Err(rollbackErr).Str("project_name", cluster.ProjectName).Msg("failed to create rollback workflow")
				return
			}

			rollbackWf.State["config"] = wf.State["config"]
			rollbackWf.State["cluster"] = wf.State["cluster"]
			rollbackWf.State["kubeclient"] = wf.State["kubeclient"]
			rollbackWf.State["project_name"] = cluster.ProjectName

			rollbackCtx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
			defer cancel()

			// wait the rollback workflow to finish before closing the client
			if err := engine.RunSync(rollbackCtx, rollbackWf); err != nil {
				log.Error().Err(err).Str("project_name", cluster.ProjectName).Msg("failed to run rollback workflow")
				return
			}
			return
		}

	}
}

func createDeployerWorkflowTemplate(notificationService *notification.NotificationService, engine *ewf.Engine, metrics *metrics.Metrics) ewf.WorkflowTemplate {
	template := newKubecloudWorkflowTemplate(notificationService)
	template.AfterWorkflowHooks = append(template.AfterWorkflowHooks,
		metricsSuccessHook(metrics),
		metricsFailureHook(metrics),
		deploymentFailureHook(engine),
		closeClient,
	)

	return template
}

func createBaseDeployerWorkflowTemplate(notificationService *notification.NotificationService) ewf.WorkflowTemplate {
	template := newKubecloudWorkflowTemplate(notificationService)
	template.AfterWorkflowHooks = append(template.AfterWorkflowHooks,
		closeClient,
	)

	return template
}

func createAddNodeWorkflowTemplate(notificationService *notification.NotificationService, engine *ewf.Engine, metrics *metrics.Metrics) ewf.WorkflowTemplate {
	template := newKubecloudWorkflowTemplate(notificationService)
	template.AfterWorkflowHooks = append(template.AfterWorkflowHooks,
		metricsSuccessHook(metrics),
		addNodeFailureHook(engine, metrics),
		closeClient,
	)
	return template
}

func registerDeploymentActivities(engine *ewf.Engine, metrics *metrics.Metrics, db models.DB, notificationService *notification.NotificationService, config internal.Configuration, fileStorage *internal.FileStorageService) {
	engine.Register(constants.StepDeployNetwork, DeployNetworkStep(metrics))
	engine.Register(constants.StepDeployLeaderNode, DeployLeaderNodeStep(metrics))
	engine.Register(constants.StepBatchDeployAllNodes, BatchDeployAllNodesStep(metrics))
	engine.Register(constants.StepRemoveCluster, CancelDeploymentStep(db, metrics))
	engine.Register(constants.StepAddNode, AddNodeStep(metrics))
	engine.Register(constants.StepUpdateNetwork, UpdateNetworkStep(metrics))
	engine.Register(constants.StepRemoveNode, RemoveDeploymentNodeStep())
	engine.Register(constants.StepStoreDeployment, StoreDeploymentStep(db, metrics, fileStorage, fileStorage))
	engine.Register(constants.StepFetchKubeconfig, FetchKubeconfigStep(db, fileStorage, fileStorage, config.SSH.PrivateKeyPath))
	engine.Register(constants.StepVerifyClusterReady, VerifyClusterReadyStep())
	engine.Register(constants.StepVerifyNewNodes, VerifyAddedNodeStep(db, config.SSH.PrivateKeyPath))
	engine.Register(constants.StepRemoveClusterFromDB, RemoveClusterFromDBStep(db, fileStorage, metrics, fileStorage))
	engine.Register(constants.StepGatherAllContractIDs, GatherAllContractIDsStep(db))
	engine.Register(constants.StepBatchCancelContracts, BatchCancelContractsStep())
	engine.Register(constants.StepDeleteAllUserClusters, DeleteAllUserClustersStep(db, fileStorage, metrics, fileStorage))

	deployWFTemplate := createDeployerWorkflowTemplate(notificationService, engine, metrics)
	deployWFTemplate.Steps = []ewf.Step{
		{Name: constants.StepDeployNetwork, RetryPolicy: criticalRetryPolicy},
		{Name: constants.StepDeployLeaderNode, RetryPolicy: criticalRetryPolicy},
		{Name: constants.StepBatchDeployAllNodes, RetryPolicy: criticalRetryPolicy},
		{Name: constants.StepFetchKubeconfig, RetryPolicy: longExponentialRetryPolicy},
		{Name: constants.StepVerifyClusterReady, RetryPolicy: longExponentialRetryPolicy},
		{Name: constants.StepStoreDeployment, RetryPolicy: standardRetryPolicy},
	}
	deployWFTemplate.AfterStepHooks = []ewf.AfterStepHook{
		notifyStepHook(notificationService),
	}
	engine.RegisterTemplate(constants.WorkflowDeployCluster, &deployWFTemplate)

	deleteWFTemplate := createDeployerWorkflowTemplate(notificationService, engine, metrics)
	deleteWFTemplate.Steps = []ewf.Step{
		{Name: constants.StepRemoveCluster, RetryPolicy: standardRetryPolicy},
		{Name: constants.StepRemoveClusterFromDB, RetryPolicy: standardRetryPolicy},
	}
	engine.RegisterTemplate(constants.WorkflowDeleteCluster, &deleteWFTemplate)

	deleteAllDeploymentsWFTemplate := createDeployerWorkflowTemplate(notificationService, engine, metrics)
	deleteAllDeploymentsWFTemplate.Steps = []ewf.Step{
		{Name: constants.StepGatherAllContractIDs, RetryPolicy: standardRetryPolicy},
		{Name: constants.StepBatchCancelContracts, RetryPolicy: standardRetryPolicy},
		{Name: constants.StepDeleteAllUserClusters, RetryPolicy: standardRetryPolicy},
	}
	engine.RegisterTemplate(constants.WorkflowDeleteAllClusters, &deleteAllDeploymentsWFTemplate)

	addNodeWFTemplate := createAddNodeWorkflowTemplate(notificationService, engine, metrics)
	addNodeWFTemplate.Steps = []ewf.Step{
		{Name: constants.StepUpdateNetwork, RetryPolicy: criticalRetryPolicy},
		{Name: constants.StepAddNode, RetryPolicy: standardRetryPolicy},
		{Name: constants.StepFetchKubeconfig, RetryPolicy: longExponentialRetryPolicy},
		{Name: constants.StepVerifyNewNodes, RetryPolicy: longExponentialRetryPolicy},
		{Name: constants.StepStoreDeployment, RetryPolicy: standardRetryPolicy},
	}
	engine.RegisterTemplate(constants.WorkflowAddNode, &addNodeWFTemplate)

	removeNodeWFTemplate := createDeployerWorkflowTemplate(notificationService, engine, metrics)
	removeNodeWFTemplate.Steps = []ewf.Step{
		{Name: constants.StepRemoveNode, RetryPolicy: standardRetryPolicy},
		{Name: constants.StepFetchKubeconfig, RetryPolicy: longExponentialRetryPolicy},
		{Name: constants.StepStoreDeployment, RetryPolicy: standardRetryPolicy},
	}
	engine.RegisterTemplate(constants.WorkflowRemoveNode, &removeNodeWFTemplate)

	rollbackWFTemplate := createDeployerWorkflowTemplate(notificationService, engine, metrics)
	rollbackWFTemplate.Steps = []ewf.Step{
		{Name: constants.StepRemoveCluster, RetryPolicy: standardRetryPolicy},
	}
	engine.RegisterTemplate(constants.WorkflowRollbackFailedDeployment, &rollbackWFTemplate)

	rollbackAddNodeWFTemplate := createBaseDeployerWorkflowTemplate(notificationService)
	rollbackAddNodeWFTemplate.Steps = []ewf.Step{
		{Name: constants.StepRemoveNode, RetryPolicy: standardRetryPolicy},
		{Name: constants.StepStoreDeployment, RetryPolicy: standardRetryPolicy},
	}
	engine.RegisterTemplate(constants.WorkflowRollbackFailedAddNode, &rollbackAddNodeWFTemplate)
}

func getFromState[T any](state ewf.State, key string) (T, error) {
	value, ok := state[key]
	if !ok {
		var zero T
		return zero, fmt.Errorf("missing '%s' in state", key)
	}

	// Try direct type assertion first (for newly created values)
	if val, ok := value.(T); ok {
		return val, nil
	}

	// Handle the case where value was serialized/deserialized and became a map
	// Use JSON marshaling/unmarshaling to convert map to struct
	valueBytes, err := json.Marshal(value)
	if err != nil {
		var zero T
		return zero, fmt.Errorf("failed to marshal %s value: %w", key, err)
	}

	var result T
	if err := json.Unmarshal(valueBytes, &result); err != nil {
		var zero T
		return zero, fmt.Errorf("failed to unmarshal %s: %w", key, err)
	}

	return result, nil
}

func getConfig(state ewf.State) (statemanager.ClientConfig, error) {
	value, ok := state["config"]
	if !ok {
		return statemanager.ClientConfig{}, fmt.Errorf("missing 'config' in state")
	}

	// Try direct type assertion first (for newly created configs)
	if config, ok := value.(statemanager.ClientConfig); ok {
		return config, nil
	}

	// Handle the case where config was serialized/deserialized and became a map
	// Use JSON marshaling/unmarshaling to convert map to struct
	configBytes, err := json.Marshal(value)
	if err != nil {
		return statemanager.ClientConfig{}, fmt.Errorf("failed to marshal config value: %w", err)
	}

	var config statemanager.ClientConfig
	if err := json.Unmarshal(configBytes, &config); err != nil {
		return statemanager.ClientConfig{}, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return config, nil
}

func retrieveKubeconfig(ctx context.Context, state ewf.State, db models.DB, fileStorage *internal.FileStorageService, fileStorage *internal.FileStorageService, privateKeyPath string) (string, error) {
	log := logger.ForOperation("deployer_activities", "retrieve_kubeconfig")
	// 1. Check if kubeconfig is already in state
	if kc, err := getFromState[string](state, "kubeconfig"); err == nil && kc != "" {
		return kc, nil
	}

	cluster, err := statemanager.GetCluster(state)
	if err != nil {
		return "", fmt.Errorf("failed to get cluster from state: %w", err)
	}

	config, err := getConfig(state)
	if err != nil {
		return "", err
	}

	existingCluster, err := db.GetClusterByName(config.UserID, cluster.ProjectName)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return "", fmt.Errorf("failed to query cluster from database: %w", err)
	}

	if existingCluster.ID != 0 && existingCluster.Kubeconfig != "" {
		logger.GetLogger().Debug().Msgf("Using kubeconfig from DB for cluster %s", existingCluster.ProjectName)
		return existingCluster.Kubeconfig, nil
	}

		if existingCluster.ID != 0 {
			if data, err := fileStorage.ReadKubeconfigFile(config.UserID, existingCluster.ID, existingCluster.ProjectName); err == nil && len(data) > 0 {
				state["kubeconfig"] = string(data)
				return nil
			}
		}

	privateKeyBytes, err := os.ReadFile(privateKeyPath)
	if err != nil {
		return "", fmt.Errorf("failed to read SSH private key (user_id=%d): %w", config.UserID, err)
	}

	// 3. fetch kubeconfig from cluster
	if existingCluster.ID != 0 { // Cluster exists in DB, this is update of existing cluster
		existingClusterResult, err := existingCluster.GetClusterResult()
		if err != nil {
			return "", fmt.Errorf("failed to get cluster result for %s (user_id=%d): %w", cluster.ProjectName, config.UserID, err)
		}
		return existingClusterResult.GetKubeconfig(ctx, string(privateKeyBytes))
	}

	// Brand new cluster
	return cluster.GetKubeconfig(ctx, string(privateKeyBytes))
}

func FetchKubeconfigStep(db models.DB, privateKeyPath string) ewf.StepFn {
	return func(ctx context.Context, state ewf.State) error {
		kubeconfig, err := retrieveKubeconfig(ctx, state, db, privateKeyPath)
		if err != nil {
			return err
		}
		state["kubeconfig"] = kubeconfig
		return nil
	}
}

func VerifyAddedNodeStep(db models.DB, privateKeyPath string) ewf.StepFn {
	return func(ctx context.Context, state ewf.State) error {
		log := logger.ForOperation("deployer_activities", "verify_added_node")
		node, err := getFromState[kubedeployer.Node](state, "node")
		if err != nil {
			return fmt.Errorf("missing or invalid 'node' in state for verification: %w", err)
		}

		kubeconfig, err := retrieveKubeconfig(ctx, state, db, privateKeyPath)
		if err != nil {
			return err
		}
		state["kubeconfig"] = kubeconfig

		restConfig, err := clientcmd.RESTConfigFromKubeConfig([]byte(kubeconfig))
		if err != nil {
			return fmt.Errorf("failed to parse kubeconfig for node %s: %w", node.Name, err)
		}

		clientset, err := kubernetes.NewForConfig(restConfig)
		if err != nil {
			return fmt.Errorf("failed to create kubernetes client for node %s: %w", node.Name, err)
		}

		n, err := clientset.CoreV1().Nodes().Get(ctx, node.Name, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("failed to get node %s from cluster: %w", node.Name, err)
		}

		ready := false
		for _, cond := range n.Status.Conditions {
			if cond.Type == v1.NodeReady && cond.Status == v1.ConditionTrue {
				ready = true
				break
			}
		}

		if !ready {
			return fmt.Errorf("new node %s is not ready", node.Name)
		}

		log.Info().
			Str("node", node.Name).
			Msg("New node is Ready")

		return nil
	}
}

func VerifyClusterReadyStep() ewf.StepFn {
	return func(ctx context.Context, state ewf.State) error {
		log := logger.ForOperation("deployer_activities", "verify_cluster_ready")

		cluster, err := statemanager.GetCluster(state)
		if err != nil {
			return fmt.Errorf("failed to get cluster: %w", err)
		}

		kubeconfig, ok := state["kubeconfig"].(string)
		if !ok || kubeconfig == "" {
			return fmt.Errorf("kubeconfig not found in workflow state")
		}

		restConfig, err := clientcmd.RESTConfigFromKubeConfig([]byte(kubeconfig))
		if err != nil {
			return fmt.Errorf("failed to parse kubeconfig for cluster %s: %w", cluster.Name, err)
		}

		clientset, err := kubernetes.NewForConfig(restConfig)
		if err != nil {
			return fmt.Errorf("failed to create kubernetes client for cluster %s: %w", cluster.Name, err)
		}

		nodes, err := clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
		if err != nil {
			return fmt.Errorf("failed to list nodes for cluster %s: %w", cluster.Name, err)
		}

		for _, n := range nodes.Items {
			ready := false
			for _, cond := range n.Status.Conditions {
				if cond.Type == v1.NodeReady && cond.Status == v1.ConditionTrue {
					ready = true
					break
				}
			}
			if !ready {
				return fmt.Errorf("node %s is not ready in cluster %s", n.Name, cluster.Name)
			}
		}

		log.Info().
			Str("cluster", cluster.Name).
			Msg("All nodes are Ready")

		return nil
	}
}

func VerifyClusterInDBStep(db models.DB) ewf.StepFn {
	return func(ctx context.Context, state ewf.State) error {
		config, err := getConfig(state)
		if err != nil {
			return err
		}

		cluster, err := statemanager.GetCluster(state)
		if err != nil {
			return fmt.Errorf("failed to get cluster from state: %w", err)
		}

		existingCluster, err := db.GetClusterByName(config.UserID, cluster.ProjectName)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("cluster %s not found in database: %w", cluster.ProjectName, ewf.ErrFailWorkflowNow)
			}
			return nil
		}

		if existingCluster.ID == 0 {
			return fmt.Errorf("cluster %s not found in database: %w", cluster.ProjectName, ewf.ErrFailWorkflowNow)
		}
		return nil
	}
}
