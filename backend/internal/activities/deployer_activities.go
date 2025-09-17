package activities

import (
	"context"
	"encoding/json"
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
	// Get config first
	config, err := getConfig(state)
	if err != nil {
		logger.GetLogger().Error().Err(err).Msg("Failed to get config")
		return
	}

	// Use the statemanager to get or create client
	_, err = statemanager.GetKubeClient(state, config)
	if err != nil {
		logger.GetLogger().Error().Err(err).Msg("Failed to ensure kubeclient")
		return
	}

	logger.GetLogger().Debug().Msg("Kubeclient ensured and ready for use")
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
				metrics.IncrementClusterDeploymentFailure()
				return fmt.Errorf("failed to prepare cluster: %w", err)
			}
		}

		if err := kubeClient.DeployNetwork(ctx, &cluster); err != nil {
			if isWorkloadAlreadyDeployedError(err) {
				metrics.IncrementClusterDeploymentFailure()
				return fmt.Errorf("network already deployed for cluster %s: %w", cluster.Name, ewf.ErrFailWorkflowNow)
			}
			if isWorkloadInvalid(err) {
				metrics.IncrementClusterDeploymentFailure()
				return fmt.Errorf("network invalid for cluster %s: %w", cluster.Name, ewf.ErrFailWorkflowNow)
			}
			metrics.IncrementClusterDeploymentFailure()
			return fmt.Errorf("failed to deploy network: %w", err)
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
			metrics.IncrementClusterDeploymentFailure()
			return fmt.Errorf("failed to update network: %w", err)
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
			metrics.IncrementClusterDeploymentFailure()
			return fmt.Errorf("failed to assign IP for node %s: %w", node.Name, err)
		}

		if err := kubeClient.DeployNode(ctx, &cluster, node, config.SSHPublicKey); err != nil {
			metrics.IncrementClusterDeploymentFailure()
			return fmt.Errorf("failed to deploy node %s to existing cluster: %w", node.Name, err)
		}

		metrics.IncrementClusterDeploymentSuccess()

		// Save GridClient state after node deployment
		statemanager.SaveGridClientState(state, kubeClient)
		statemanager.StoreCluster(state, cluster)
		return nil
	}
}

func DeployNodeStep(metrics *metrics.Metrics) ewf.StepFn {
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

		nodeIdx, ok := state["node_index"].(int)
		if !ok {
			nodeIdx = 0
		}
		node := cluster.Nodes[nodeIdx]

		if err := node.AssignNodeIP(ctx, kubeClient.GridClient, cluster.Network.Name); err != nil {
			metrics.IncrementClusterDeploymentFailure()
			return fmt.Errorf("failed to assign node IPs: %w", err)
		}
		cluster.Nodes[nodeIdx].IP = node.IP

		if err := kubeClient.DeployNode(ctx, &cluster, node, config.SSHPublicKey); err != nil {
			if isWorkloadAlreadyDeployedError(err) {
				metrics.IncrementClusterDeploymentFailure()
				return fmt.Errorf("node already deployed for cluster %s: %w", cluster.Name, ewf.ErrFailWorkflowNow)
			}
			if isWorkloadInvalid(err) {
				metrics.IncrementClusterDeploymentFailure()
				return fmt.Errorf("node invalid for cluster %s: %w", cluster.Name, ewf.ErrFailWorkflowNow)
			}
			metrics.IncrementClusterDeploymentFailure()
			return fmt.Errorf("failed to deploy node %s: %w", node.Name, err)
		}

		metrics.IncrementClusterDeploymentSuccess()

		// Save GridClient state after node deployment
		statemanager.SaveGridClientState(state, kubeClient)
		statemanager.StoreCluster(state, cluster)
		state["node_index"] = nodeIdx + 1
		return nil
	}
}

func StoreDeploymentStep(db models.DB, metrics *metrics.Metrics) ewf.StepFn {
	return func(ctx context.Context, state ewf.State) error {
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
			return fmt.Errorf("failed to set cluster result: %w", err)
		}

		existingCluster, err := db.GetClusterByName(config.UserID, cluster.ProjectName)
		if err != nil { // cluster not found, create a new one
			if err := db.CreateCluster(config.UserID, dbCluster); err != nil {
				return fmt.Errorf("failed to create cluster in database: %w", err)
			}
		} else { // cluster exists, update it
			existingCluster.Result = dbCluster.Result
			if err := db.UpdateCluster(&existingCluster); err != nil {
				return fmt.Errorf("failed to update cluster in database: %w", err)
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
				return fmt.Errorf("failed to get cluster from database: %w", err)
			}

			cluster, err = dbCluster.GetClusterResult()
			if err != nil {
				return fmt.Errorf("failed to get cluster result: %w", err)
			}
		}

		if err := kubeClient.CancelCluster(ctx, cluster); err != nil {
			return fmt.Errorf("failed to cancel deployment: %w", err)
		}

		metrics.DecActiveClusterCount()
		return nil
	}
}

func RemoveClusterFromDBStep(db models.DB) ewf.StepFn {
	return func(ctx context.Context, state ewf.State) error {
		config, err := getConfig(state)
		if err != nil {
			return err
		}

		projectName, ok := state["project_name"].(string)
		if !ok {
			return fmt.Errorf("missing or invalid 'project_name' in state")
		}

		if err := db.DeleteCluster(config.UserID, projectName); err != nil {
			return fmt.Errorf("failed to delete cluster from database: %w", err)
		}

		return nil
	}
}

func GatherAllContractIDsStep(db models.DB) ewf.StepFn {
	return func(ctx context.Context, state ewf.State) error {
		config, err := getConfig(state)
		if err != nil {
			return err
		}

		clusters, err := db.ListUserClusters(config.UserID)
		if err != nil {
			return fmt.Errorf("failed to list user clusters: %w", err)
		}

		var allContractIDs []uint64
		for _, cluster := range clusters {
			clusterResult, err := cluster.GetClusterResult()
			if err != nil {
				logger.GetLogger().Error().Err(err).Int("cluster_id", cluster.ID).Msg("Failed to deserialize cluster result")
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
			logger.GetLogger().Info().Int("user_id", config.UserID).Msg("No contracts to cancel")
			return nil
		}

		if err := kubeClient.CancelAllContractsForUser(ctx, contractIDs); err != nil {
			return fmt.Errorf("failed to cancel contracts: %w", err)
		}

		return nil
	}
}

func DeleteAllUserClustersStep(db models.DB) ewf.StepFn {
	return func(ctx context.Context, state ewf.State) error {
		config, err := getConfig(state)
		if err != nil {
			return err
		}

		if err := db.DeleteAllUserClusters(config.UserID); err != nil {
			return fmt.Errorf("failed to delete all user clusters from database: %w", err)
		}

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
			return fmt.Errorf("failed to remove node %s from existing cluster: %w", nodeName, err)
		}

		// Save GridClient state after node removal
		statemanager.SaveGridClientState(state, kubeClient)
		statemanager.StoreCluster(state, existingCluster)
		return nil
	}
}

func NewDynamicDeployWorkflowTemplate(engine *ewf.Engine, metrics *metrics.Metrics, notificationService *notification.NotificationService, wfName string, nodesNum int) {
	steps := []ewf.Step{
		{Name: constants.StepDeployNetwork, RetryPolicy: criticalRetryPolicy},
	}

	for i := 0; i < nodesNum; i++ {
		stepName := getDeployNodeStepName(i + 1)
		engine.Register(stepName, DeployNodeStep(metrics))

		steps = append(steps, ewf.Step{Name: stepName, RetryPolicy: criticalRetryPolicy})
	}

	steps = append(steps, ewf.Step{Name: constants.StepFetchKubeconfig, RetryPolicy: criticalRetryPolicy})
	steps = append(steps, ewf.Step{Name: constants.StepVerifyClusterReady, RetryPolicy: longExponentialRetryPolicy})
	steps = append(steps, ewf.Step{Name: constants.StepStoreDeployment, RetryPolicy: standardRetryPolicy})

	workflow := createDeployerWorkflowTemplate(notificationService, engine, metrics)
	workflow.Steps = steps
	workflow.AfterStepHooks = []ewf.AfterStepHook{
		notifyStepHook(notificationService),
	}

	engine.RegisterTemplate(wfName, &workflow)
}

func CloseClient(ctx context.Context, wf *ewf.Workflow, err error) {
	if kubeClient, ok := wf.State["kubeclient"].(*kubedeployer.Client); ok {
		// Save final GridClient state before closing
		statemanager.SaveGridClientState(wf.State, kubeClient)

		kubeClient.Close()
		delete(wf.State, "kubeclient")
	} else {
		logger.GetLogger().Warn().Msg("No kubeclient found in workflow state to close")
	}

}

func deploymentFailureHook(engine *ewf.Engine, metrics *metrics.Metrics) ewf.AfterWorkflowHook {
	return func(ctx context.Context, wf *ewf.Workflow, err error) {
		if err != nil && isDeployWorkflow(wf.Name) {
			cluster, clusterErr := statemanager.GetCluster(wf.State)
			if clusterErr != nil || cluster.ProjectName == "" {
				logger.GetLogger().Error().Err(clusterErr).Str("workflow_name", wf.Name).Msg("nothing to rollback")
				return
			}

			logger.GetLogger().Info().Str("project_name", cluster.ProjectName).Str("workflow_name", wf.Name).Msg("Triggering rollback workflow for failed deployment")

			rollbackWf, rollbackErr := engine.NewWorkflow(constants.WorkflowRollbackFailedDeployment)
			if rollbackErr != nil {
				logger.GetLogger().Error().Err(rollbackErr).Str("project_name", cluster.ProjectName).Msg("Failed to create rollback workflow")
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
				logger.GetLogger().Error().Err(err).Str("project_name", cluster.ProjectName).Msg("Failed to run rollback workflow")
				return
			}

			metrics.DecActiveClusterCount()
		}
	}
}

func createDeployerWorkflowTemplate(notificationService *notification.NotificationService, engine *ewf.Engine, metrics *metrics.Metrics) ewf.WorkflowTemplate {
	template := newKubecloudWorkflowTemplate(notificationService)
	template.AfterWorkflowHooks = append(template.AfterWorkflowHooks,
		[]ewf.AfterWorkflowHook{
			deploymentFailureHook(engine, metrics),
			CloseClient,
		}...)

	return template
}

func registerDeploymentActivities(engine *ewf.Engine, metrics *metrics.Metrics, db models.DB, notificationService *notification.NotificationService, config internal.Configuration) {

	engine.Register(constants.StepDeployNetwork, DeployNetworkStep(metrics))
	engine.Register(constants.StepDeployNode, DeployNodeStep(metrics))
	engine.Register(constants.StepRemoveCluster, CancelDeploymentStep(db, metrics))
	engine.Register(constants.StepAddNode, AddNodeStep(metrics))
	engine.Register(constants.StepUpdateNetwork, UpdateNetworkStep(metrics))
	engine.Register(constants.StepRemoveNode, RemoveDeploymentNodeStep())
	engine.Register(constants.StepStoreDeployment, StoreDeploymentStep(db, metrics))
	engine.Register(constants.StepFetchKubeconfig, FetchKubeconfigStep(config.SSH.PrivateKeyPath))
	engine.Register(constants.StepVerifyClusterReady, VerifyClusterReadyStep())
	engine.Register(constants.StepRemoveClusterFromDB, RemoveClusterFromDBStep(db))
	engine.Register(constants.StepGatherAllContractIDs, GatherAllContractIDsStep(db))
	engine.Register(constants.StepBatchCancelContracts, BatchCancelContractsStep())
	engine.Register(constants.StepDeleteAllUserClusters, DeleteAllUserClustersStep(db))

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

	addNodeWFTemplate := createDeployerWorkflowTemplate(notificationService, engine, metrics)
	addNodeWFTemplate.Steps = []ewf.Step{
		{Name: constants.StepUpdateNetwork, RetryPolicy: criticalRetryPolicy},
		{Name: constants.StepAddNode, RetryPolicy: standardRetryPolicy},
		{Name: constants.StepStoreDeployment, RetryPolicy: standardRetryPolicy},
	}
	engine.RegisterTemplate(constants.WorkflowAddNode, &addNodeWFTemplate)

	removeNodeWFTemplate := createDeployerWorkflowTemplate(notificationService, engine, metrics)
	removeNodeWFTemplate.Steps = []ewf.Step{
		{Name: constants.StepRemoveNode, RetryPolicy: standardRetryPolicy},
		{Name: constants.StepStoreDeployment, RetryPolicy: standardRetryPolicy},
	}
	engine.RegisterTemplate(constants.WorkflowRemoveNode, &removeNodeWFTemplate)

	rollbackWFTemplate := createDeployerWorkflowTemplate(notificationService, engine, metrics)
	rollbackWFTemplate.Steps = []ewf.Step{
		{Name: constants.StepRemoveCluster, RetryPolicy: standardRetryPolicy},
	}
	engine.RegisterTemplate(constants.WorkflowRollbackFailedDeployment, &rollbackWFTemplate)
}

func getFromState[T any](state ewf.State, key string) (T, error) {
	value, ok := state[key]
	if !ok {
		var zero T
		return zero, fmt.Errorf("missing '%s' in state", key)
	}

	val, ok := value.(T)
	if !ok {
		var zero T
		logger.GetLogger().Error().Msgf("Expected '%s' to be of %+v, but got %+v", key, zero, value)
		return zero, fmt.Errorf("invalid '%s' in state", key)
	}
	return val, nil
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

func FetchKubeconfigStep(privateKeyPath string) ewf.StepFn {
	return func(ctx context.Context, state ewf.State) error {
		cluster, err := statemanager.GetCluster(state)
		if err != nil {
			return fmt.Errorf("failed to get cluster from state: %w", err)
		}

		master, err := cluster.GetLeaderNode()
		if err != nil {
			return fmt.Errorf("failed to get leader node in cluster")
		}

		privateKeyBytes, err := os.ReadFile(privateKeyPath)
		if err != nil {
			return fmt.Errorf("failed to read SSH private key: %w", err)
		}

		kubeconfig, err := internal.GetKubeconfigViaSSH(string(privateKeyBytes), &master)
		if err != nil {
			return fmt.Errorf("failed to retrieve kubeconfig via SSH: %w", err)
		}

		state["kubeconfig"] = kubeconfig
		return nil
	}
}

func VerifyClusterReadyStep() ewf.StepFn {
	return func(ctx context.Context, state ewf.State) error {

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
			return fmt.Errorf("failed to parse kubeconfig: %w", err)
		}

		clientset, err := kubernetes.NewForConfig(restConfig)
		if err != nil {
			return fmt.Errorf("failed to create kubernetes client: %w", err)
		}

		nodes, err := clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
		if err != nil {
			return fmt.Errorf("failed to list nodes: %w", err)
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
				return fmt.Errorf("node %s is not ready", n.Name)
			}
		}

		logger.GetLogger().Info().
			Str("cluster", cluster.Name).
			Msg("All nodes are Ready")

		return nil
	}
}
