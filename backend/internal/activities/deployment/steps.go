package deployment

import (
	"context"
	"errors"
	"fmt"
	"kubecloud/internal/metrics"
	"kubecloud/internal/statemanager"
	"kubecloud/kubedeployer"
	"kubecloud/models"

	"kubecloud/internal/logger"

	"github.com/xmonader/ewf"
)

// =============================================================================
// DEPLOY/UPDATE STEPS
// =============================================================================

func DeployNetworkStep(metrics *metrics.Metrics) ewf.StepFn {
	return func(ctx context.Context, state ewf.State) error {
		stepCtx, err := NewStepContext(state, metrics)
		if err != nil {
			return err
		}
		defer stepCtx.SaveState()

		if stepCtx.Cluster.ProjectName == "" { // first time, not retry
			if err := stepCtx.Cluster.PrepareCluster(stepCtx.Config.UserID); err != nil {
				stepCtx.Metrics.IncrementClusterDeploymentFailure()
				return fmt.Errorf("failed to prepare cluster: %w", err)
			}
		}

		if err := stepCtx.KubeClient.DeployNetwork(ctx, &stepCtx.Cluster); err != nil {
			return handleDeploymentError(err, stepCtx, "network")
		}

		return nil
	}
}

func UpdateNetworkStep(metrics *metrics.Metrics) ewf.StepFn {
	return func(ctx context.Context, state ewf.State) error {
		stepCtx, err := NewStepContext(state, metrics)
		if err != nil {
			return err
		}
		defer stepCtx.SaveState()

		node, err := GetFromState[kubedeployer.Node](stepCtx.State, "node")
		if err != nil {
			return err
		}

		node.Name = kubedeployer.GetNodeName(stepCtx.Config.UserID, stepCtx.Cluster.Name, node.OriginalName)
		stepCtx.Cluster.Nodes = append(stepCtx.Cluster.Nodes, node)

		if err := stepCtx.KubeClient.DeployNetwork(ctx, &stepCtx.Cluster); err != nil {
			return handleDeploymentError(err, stepCtx, "network")
		}

		stepCtx.State["node"] = node
		return nil
	}
}

func DeployNodeStep(metrics *metrics.Metrics) ewf.StepFn {
	return func(ctx context.Context, state ewf.State) error {
		stepCtx, err := NewStepContext(state, metrics)
		if err != nil {
			return err
		}
		defer stepCtx.SaveState()

		nodeIdx, ok := stepCtx.State["node_index"].(int)
		if !ok {
			nodeIdx = 0
		}
		node := stepCtx.Cluster.Nodes[nodeIdx]

		if err := node.AssignNodeIP(ctx, stepCtx.KubeClient.GridClient, stepCtx.Cluster.Network.Name); err != nil {
			stepCtx.Metrics.IncrementClusterDeploymentFailure()
			return fmt.Errorf("failed to assign IP for node %s: %w", node.Name, err)
		}
		stepCtx.Cluster.Nodes[nodeIdx].IP = node.IP

		if err := stepCtx.KubeClient.DeployNode(ctx, &stepCtx.Cluster, node, stepCtx.Config.SSHPublicKey); err != nil {
			return handleDeploymentError(err, stepCtx, "node", node.Name)
		}

		// TODO: this should not be here
		// stepCtx.Metrics.IncrementClusterDeploymentSuccess()
		stepCtx.State["node_index"] = nodeIdx + 1
		return nil
	}
}

func AddNodeStep(metrics *metrics.Metrics) ewf.StepFn {
	return func(ctx context.Context, state ewf.State) error {
		stepCtx, err := NewStepContext(state, metrics)
		if err != nil {
			return err
		}
		defer stepCtx.SaveState()

		node, err := GetFromState[kubedeployer.Node](stepCtx.State, "node")
		if err != nil {
			return err
		}

		if err := node.AssignNodeIP(ctx, stepCtx.KubeClient.GridClient, stepCtx.Cluster.Network.Name); err != nil {
			stepCtx.Metrics.IncrementClusterDeploymentFailure()
			return fmt.Errorf("failed to assign IP for node %s: %w", node.Name, err)
		}

		if err := stepCtx.KubeClient.DeployNode(ctx, &stepCtx.Cluster, node, stepCtx.Config.SSHPublicKey); err != nil {
			return handleDeploymentError(err, stepCtx, "node", node.Name)
		}

		// TODO: this should not be here
		// stepCtx.Metrics.IncrementClusterDeploymentSuccess()
		stepCtx.State["node"] = node
		return nil
	}
}

// =============================================================================
// VERIFICATION STEPS
// =============================================================================

func FetchKubeconfigStep(db models.DB, privateKeyPath string) ewf.StepFn {
	return func(ctx context.Context, state ewf.State) error {
		kubeconfig, err := retrieveKubeconfig(state, db, privateKeyPath)
		if err != nil {
			return err
		}
		state["kubeconfig"] = kubeconfig
		return nil
	}
}

func VerifyClusterReadyStep() ewf.StepFn {
	return func(ctx context.Context, state ewf.State) error {
		kubeconfig, err := GetFromState[string](state, "kubeconfig")
		if err != nil {
			return err
		}

		clientset, err := createKubernetesClient(kubeconfig)
		if err != nil {
			return err
		}

		if err := verifyAllNodesReady(ctx, clientset); err != nil {
			return err
		}

		return nil
	}
}

// TODO: can't it be merged with VerifyClusterReadyStep?
func VerifyAddedNodeStep(db models.DB, privateKeyPath string) ewf.StepFn {
	return func(ctx context.Context, state ewf.State) error {
		node, err := GetFromState[kubedeployer.Node](state, "node")
		if err != nil {
			return fmt.Errorf("missing or invalid 'node' in state for verification: %w", err)
		}

		kubeconfig, err := retrieveKubeconfig(state, db, privateKeyPath)
		if err != nil {
			return err
		}
		state["kubeconfig"] = kubeconfig

		clientset, err := createKubernetesClient(kubeconfig)
		if err != nil {
			return err
		}

		if err := verifySpecificNodeReady(ctx, clientset, node.Name); err != nil {
			return fmt.Errorf("new node verification failed: %w", err)
		}

		return nil
	}
}

// =============================================================================
// STORAGE STEPS
// =============================================================================

func StoreDeploymentStep(db models.DB, metrics *metrics.Metrics) ewf.StepFn {
	return func(ctx context.Context, state ewf.State) error {
		stepCtx, err := NewStepContext(state, metrics)
		if err != nil {
			return err
		}

		kubeconfig, err := GetFromState[string](state, "kubeconfig")
		if err != nil {
			return err
		}

		dbCluster := &models.Cluster{
			ProjectName: stepCtx.Cluster.ProjectName,
			Kubeconfig:  kubeconfig,
		}

		if err := dbCluster.SetClusterResult(stepCtx.Cluster); err != nil {
			return fmt.Errorf("failed to set cluster result: %w", err)
		}

		// TODO: optimize with upsert
		existingCluster, err := db.GetClusterByName(stepCtx.Config.UserID, stepCtx.Cluster.ProjectName)
		if err != nil { // cluster not found, create a new one
			if err := db.CreateCluster(stepCtx.Config.UserID, dbCluster); err != nil {
				return fmt.Errorf("failed to create cluster in database: %w", err)
			}
		} else { // cluster exists, update it
			existingCluster.Result = dbCluster.Result
			existingCluster.Kubeconfig = dbCluster.Kubeconfig
			if err := db.UpdateCluster(&existingCluster); err != nil {
				return fmt.Errorf("failed to update cluster in database: %w", err)
			}
		}

		metrics.IncActiveClusterCount()
		return nil
	}
}

// =============================================================================
// CANCEL/REMOVE STEPS
// =============================================================================

func CancelDeploymentStep(db models.DB, metrics *metrics.Metrics) ewf.StepFn {
	return func(ctx context.Context, state ewf.State) error {
		stepCtx, err := NewStepContext(state, metrics)
		if errors.Is(err, ErrCluster) {
			// in a Rollback, cluster is in state, in a delete, we need to load from db
			projectName, err := GetFromState[string](state, "project_name")
			if err != nil {
				return fmt.Errorf("failed to get project_name from state: %w", err)
			}

			dbCluster, err := db.GetClusterByName(stepCtx.Config.UserID, projectName)
			if err != nil {
				return fmt.Errorf("failed to get cluster from database: %w", err)
			}

			stepCtx.Cluster, err = dbCluster.GetClusterResult()
			if err != nil {
				return fmt.Errorf("failed to get cluster result: %w", err)
			}
		} else if err != nil {
			return err
		}

		defer stepCtx.SaveState()

		if err := stepCtx.KubeClient.CancelCluster(ctx, stepCtx.Cluster); err != nil {
			return fmt.Errorf("failed to cancel deployment: %w", err)
		}

		metrics.DecActiveClusterCount()
		return nil
	}
}

func RemoveNodeStep(metrics *metrics.Metrics) ewf.StepFn {
	return func(ctx context.Context, state ewf.State) error {
		stepCtx, err := NewStepContext(state, metrics)
		if err != nil {
			return err
		}
		defer stepCtx.SaveState()

		nodeName, err := GetFromState[string](state, "node_name")
		if err != nil {
			return fmt.Errorf("failed to get node_name from state: %w", err)
		}
		nodeName = kubedeployer.GetNodeName(stepCtx.Config.UserID, stepCtx.Cluster.Name, nodeName)

		if err := stepCtx.KubeClient.RemoveNode(ctx, &stepCtx.Cluster, nodeName); err != nil {
			return fmt.Errorf("failed to remove node %s from existing cluster: %w", nodeName, err)
		}

		return nil
	}
}

func RemoveClusterFromDBStep(db models.DB) ewf.StepFn {
	return func(ctx context.Context, state ewf.State) error {
		config, err := GetFromState[statemanager.ClientConfig](state, "config")
		if err != nil {
			return fmt.Errorf("failed to get config from state: %w", err)
		}

		projectName, err := GetFromState[string](state, "project_name")
		if err != nil {
			return fmt.Errorf("failed to get project_name from state: %w", err)
		}

		if err := db.DeleteCluster(config.UserID, projectName); err != nil {
			return fmt.Errorf("failed to delete cluster from database: %w", err)
		}

		return nil
	}
}

func GatherAllContractIDsStep(db models.DB) ewf.StepFn {
	return func(ctx context.Context, state ewf.State) error {
		config, err := GetFromState[statemanager.ClientConfig](state, "config")
		if err != nil {
			return fmt.Errorf("failed to get config from state: %w", err)
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
		stepCtx, err := NewStepContext(state, nil)
		if err != nil {
			return err
		}

		contractIDs, err := GetFromState[[]uint64](state, "contract_ids")
		if err != nil {
			return fmt.Errorf("failed to get contract_ids from state: %w", err)
		}

		if len(contractIDs) == 0 {
			logger.GetLogger().Info().Int("user_id", stepCtx.Config.UserID).Msg("No contracts to cancel")
			return nil
		}

		if err := stepCtx.KubeClient.CancelAllContractsForUser(ctx, contractIDs); err != nil {
			return fmt.Errorf("failed to cancel contracts: %w", err)
		}

		return nil
	}
}

func DeleteAllUserClustersStep(db models.DB) ewf.StepFn {
	return func(ctx context.Context, state ewf.State) error {
		config, err := GetFromState[statemanager.ClientConfig](state, "config")
		if err != nil {
			return fmt.Errorf("failed to get config from state: %w", err)
		}

		if err := db.DeleteAllUserClusters(config.UserID); err != nil {
			return fmt.Errorf("failed to delete all user clusters from database: %w", err)
		}

		return nil
	}
}
