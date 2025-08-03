package activities

import (
	"context"
	"encoding/json"
	"fmt"
	"kubecloud/internal"
	"kubecloud/internal/statemanager"
	"kubecloud/kubedeployer"
	"kubecloud/models"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/xmonader/ewf"
)

var (
	criticalRetryPolicy = &ewf.RetryPolicy{MaxAttempts: 5, BackOff: ewf.ConstantBackoff(5 * time.Second)}
	standardRetryPolicy = &ewf.RetryPolicy{MaxAttempts: 2, BackOff: ewf.ConstantBackoff(2 * time.Second)}
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
		log.Error().Err(err).Msg("Failed to get config")
		return
	}

	// Use the statemanager to get or create client
	_, err = statemanager.GetKubeClient(state, config)
	if err != nil {
		log.Error().Err(err).Msg("Failed to ensure kubeclient")
		return
	}

	log.Debug().Msg("Kubeclient ensured and ready for use")
}

func DeployNetworkStep() ewf.StepFn {
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

		if err := kubeClient.DeployNetwork(ctx, &cluster); err != nil {
			if isWorkloadAlreadyDeployedError(err) {
				return fmt.Errorf("network already deployed for cluster %s: %w", cluster.Name, ewf.ErrFailWorkflowNow)
			}
			if isWorkloadInvalid(err) {
				return fmt.Errorf("network invalid for cluster %s: %w", cluster.Name, ewf.ErrFailWorkflowNow)
			}
			return fmt.Errorf("failed to deploy network: %w", err)
		}

		// Save GridClient state after network deployment
		statemanager.SaveGridClientState(state, kubeClient)
		statemanager.StoreCluster(state, cluster)
		return nil
	}
}

func UpdateNetworkStep() ewf.StepFn {
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
			return fmt.Errorf("failed to update network: %w", err)
		}

		// Save GridClient state after network update
		statemanager.SaveGridClientState(state, kubeClient)
		statemanager.StoreCluster(state, cluster)
		state["node"] = node
		return nil
	}
}

func AddNodeStep() ewf.StepFn {
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
			return fmt.Errorf("failed to assign IP for node %s: %w", node.Name, err)
		}

		if err := kubeClient.DeployNode(ctx, &cluster, node, config.SSHPublicKey); err != nil {
			return fmt.Errorf("failed to deploy node %s to existing cluster: %w", node.Name, err)
		}

		// Save GridClient state after node deployment
		statemanager.SaveGridClientState(state, kubeClient)
		statemanager.StoreCluster(state, cluster)
		return nil
	}
}

func DeployNodeStep() ewf.StepFn {
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
			return fmt.Errorf("failed to assign node IPs: %w", err)
		}
		cluster.Nodes[nodeIdx].IP = node.IP

		if err := kubeClient.DeployNode(ctx, &cluster, node, config.SSHPublicKey); err != nil {
			if isWorkloadAlreadyDeployedError(err) {
				return fmt.Errorf("node already deployed for cluster %s: %w", cluster.Name, ewf.ErrFailWorkflowNow)
			}
			if isWorkloadInvalid(err) {
				return fmt.Errorf("node invalid for cluster %s: %w", cluster.Name, ewf.ErrFailWorkflowNow)
			}
			return fmt.Errorf("failed to deploy node %s: %w", node.Name, err)
		}

		// Save GridClient state after node deployment
		statemanager.SaveGridClientState(state, kubeClient)
		statemanager.StoreCluster(state, cluster)
		state["node_index"] = nodeIdx + 1
		return nil
	}
}

func StoreDeploymentStep(db models.DB) ewf.StepFn {
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

		return nil
	}
}

func CancelDeploymentStep() ewf.StepFn {
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

		projectName, ok := state["project_name"].(string)
		if !ok {
			return fmt.Errorf("missing or invalid 'project_name' in state")
		}

		if err := kubeClient.CancelCluster(ctx, projectName); err != nil {
			return fmt.Errorf("failed to cancel deployment: %w", err)
		}

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

func NewDynamicDeployWorkflowTemplate(engine *ewf.Engine, wfName string, nodesNum int) {
	steps := []ewf.Step{
		{Name: StepDeployNetwork, RetryPolicy: criticalRetryPolicy},
	}

	for i := 0; i < nodesNum; i++ {
		stepName := fmt.Sprintf("deploy_node_%d", i) // TODO: should be cleaned
		engine.Register(stepName, DeployNodeStep())
		steps = append(steps, ewf.Step{Name: stepName, RetryPolicy: criticalRetryPolicy})
	}

	steps = append(steps, ewf.Step{Name: StepStoreDeployment, RetryPolicy: standardRetryPolicy})

	workflow := BaseWFTemplate
	workflow.Steps = steps

	engine.RegisterTemplate(wfName, &workflow)
}

func validateConfig(config statemanager.ClientConfig) error {
	return statemanager.ValidateConfig(config)
}

// DEPRECATED: each setup uses ensureClient now
func SetupClient(ctx context.Context, wf *ewf.Workflow) {
	config, ok := wf.State["config"].(statemanager.ClientConfig)
	if !ok {
		log.Error().Msg("Missing or invalid 'config' in workflow state")
		return
	}

	if err := validateConfig(config); err != nil {
		log.Error().Err(err).Msg("Invalid workflow configuration")
		return
	}

	kubeClient, err := kubedeployer.NewClient(config.Mnemonic, config.Network)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create kubeclient")
		return
	}

	wf.State["kubeclient"] = kubeClient
}

func CloseClient(ctx context.Context, wf *ewf.Workflow, err error) {
	if kubeClient, ok := wf.State["kubeclient"].(*kubedeployer.Client); ok {
		// Save final GridClient state before closing
		statemanager.SaveGridClientState(wf.State, kubeClient)

		kubeClient.Close()
		delete(wf.State, "kubeclient")
	} else {
		log.Warn().Msg("No kubeclient found in workflow state to close")
	}

	if err != nil {
		log.Error().Err(err).Str("workflow_name", wf.Name).Msg("Workflow completed with error")
	} else {
		log.Info().Str("workflow_name", wf.Name).Msg("Workflow completed successfully")
	}
}

func NotifyUser(sse *internal.SSEManager) ewf.AfterWorkflowHook {
	return func(ctx context.Context, wf *ewf.Workflow, err error) {
		config, ok := wf.State["config"].(statemanager.ClientConfig)
		if !ok {
			log.Error().Msg("Missing or invalid 'config' in workflow state")
			return
		}

		notificationData := map[string]interface{}{
			"type":    "workflow_update",
			"message": "Workflow failed",
		}

		if err != nil {
			notificationData["data"] = map[string]interface{}{"name": wf.Name, "error": err.Error()}
		} else {
			cluster, clusterErr := statemanager.GetCluster(wf.State)
			if clusterErr != nil {
				notificationData = map[string]interface{}{
					"type":    "workflow_update",
					"message": "Workflow completed",
					"data":    map[string]interface{}{"name": wf.Name, "error": false},
				}
			} else {
				notificationData = map[string]interface{}{
					"type":    "workflow_update",
					"message": "Workflow completed successfully",
					"data":    map[string]interface{}{"name": wf.Name, "cluster": cluster, "error": false},
				}
			}
		}

		sse.Notify(config.UserID, "workflow_update", notificationData)
	}
}

var BaseWFTemplate = ewf.WorkflowTemplate{
	BeforeWorkflowHooks: []ewf.BeforeWorkflowHook{
		func(ctx context.Context, w *ewf.Workflow) {
			log.Info().Str("workflow_name", w.Name).Msg("Starting workflow")
		},
		// SetupClient,
	},
	AfterWorkflowHooks: []ewf.AfterWorkflowHook{
		// NotifyUser,
		CloseClient,
	},
	BeforeStepHooks: []ewf.BeforeStepHook{
		func(ctx context.Context, w *ewf.Workflow, step *ewf.Step) {
			log.Info().Str("workflow_name", w.Name).Str("step_name", step.Name).Msg("Starting step")
		},
	},
	AfterStepHooks: []ewf.AfterStepHook{
		func(ctx context.Context, w *ewf.Workflow, step *ewf.Step, err error) {
			if err != nil {
				log.Error().Err(err).Str("workflow_name", w.Name).Str("step_name", step.Name).Msg("Step failed")
			} else {
				log.Info().Str("workflow_name", w.Name).Str("step_name", step.Name).Msg("Step completed successfully")
			}
		},
	},
}

func createDeployWorkflowTemplates(engine *ewf.Engine) {
	for i := 1; i <= 10; i++ {
		workflowName := fmt.Sprintf("deploy-%d-nodes", i)

		steps := []ewf.Step{
			{Name: StepDeployNetwork, RetryPolicy: criticalRetryPolicy},
		}

		for j := 0; j < i; j++ {
			stepName := fmt.Sprintf("deploy_node_%d", j)
			engine.Register(stepName, DeployNodeStep())
			steps = append(steps, ewf.Step{Name: stepName, RetryPolicy: criticalRetryPolicy})
		}

		steps = append(steps, ewf.Step{Name: StepStoreDeployment, RetryPolicy: standardRetryPolicy})

		workflowTemplate := BaseWFTemplate
		workflowTemplate.Steps = steps

		engine.RegisterTemplate(workflowName, &workflowTemplate)
	}
}

func registerDeploymentActivities(engine *ewf.Engine, db models.DB, sse *internal.SSEManager) {

	engine.Register(StepDeployNetwork, DeployNetworkStep())
	engine.Register(StepDeployNode, DeployNodeStep())
	engine.Register(StepRemoveCluster, CancelDeploymentStep())
	engine.Register(StepAddNode, AddNodeStep())
	engine.Register(StepUpdateNetwork, UpdateNetworkStep())
	engine.Register(StepRemoveNode, RemoveDeploymentNodeStep())
	engine.Register(StepStoreDeployment, StoreDeploymentStep(db))
	engine.Register(StepRemoveClusterFromDB, RemoveClusterFromDBStep(db))

	createDeployWorkflowTemplates(engine)

	BaseWFTemplate.AfterWorkflowHooks = append(BaseWFTemplate.AfterWorkflowHooks, NotifyUser(sse))

	deleteWFTemplate := BaseWFTemplate
	deleteWFTemplate.Steps = []ewf.Step{
		{Name: StepRemoveCluster, RetryPolicy: standardRetryPolicy},
		{Name: StepRemoveClusterFromDB, RetryPolicy: standardRetryPolicy},
	}
	engine.RegisterTemplate(WorkflowDeleteCluster, &deleteWFTemplate)

	addNodeWFTemplate := BaseWFTemplate
	addNodeWFTemplate.Steps = []ewf.Step{
		{Name: StepUpdateNetwork, RetryPolicy: criticalRetryPolicy},
		{Name: StepAddNode, RetryPolicy: standardRetryPolicy},
		{Name: StepStoreDeployment, RetryPolicy: standardRetryPolicy},
	}
	engine.RegisterTemplate(WorkflowAddNode, &addNodeWFTemplate)

	removeNodeWFTemplate := BaseWFTemplate
	removeNodeWFTemplate.Steps = []ewf.Step{
		{Name: StepRemoveNode, RetryPolicy: standardRetryPolicy},
		{Name: StepStoreDeployment, RetryPolicy: standardRetryPolicy},
	}
	engine.RegisterTemplate(WorkflowRemoveNode, &removeNodeWFTemplate)
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
		log.Error().Msgf("Expected '%s' to be of %+v, but got %+v", key, zero, value)
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
