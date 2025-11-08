package kubedeployer

import (
	"context"
	"fmt"

	"kubecloud/internal/logger"

	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/deployer"
	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/workloads"
)

func (c *Cluster) GetLeaderNode() (Node, error) {
	return c.Nodes[0], nil
}

func (n *Node) AssignNodeIP(ctx context.Context, gridClient deployer.TFPluginClient, networkName string) error {
	logger.GetLogger().Debug().
		Str("node_name", n.Name).
		Uint32("node_id", n.NodeID).
		Str("network", networkName).
		Msg("Assigning IP for node")

	ip, err := getIpForVm(ctx, gridClient, networkName, n.NodeID)
	if err != nil {
		return fmt.Errorf("failed to get IP for node %s: %v", n.Name, err)
	}

	n.IP = ip
	logger.GetLogger().Debug().
		Str("node_name", n.Name).
		Uint32("node_id", n.NodeID).
		Str("ip", ip).
		Str("network", networkName).
		Msg("IP assigned successfully")

	return nil
}

func (c *Client) DeployNode(ctx context.Context, cluster *Cluster, node Node, masterPubKey string) error {
	logger.GetLogger().Info().
		Str("node_name", node.Name).
		Uint32("node_id", node.NodeID).
		Str("cluster", cluster.Name).
		Str("node_type", string(node.Type)).
		Msg("Starting node deployment")

	var leaderIP string
	if node.Type == NodeTypeLeader {
		leaderIP = ""
		logger.GetLogger().Debug().Str("node_name", node.Name).Msg("Deploying as leader node")
	} else {
		leaderNode, err := cluster.GetLeaderNode()
		if err != nil {
			logger.GetLogger().Error().Err(err).Str("cluster", cluster.Name).Msg("Failed to get leader node")
			return fmt.Errorf("failed to get leader node IP: %v", err)
		}

		leaderIP = leaderNode.IP
	}

	if cluster.Token == "" {
		cluster.Token = generateRandomString(32)
	}

	depl, err := deploymentFromNode(
		node,
		cluster.ProjectName,
		cluster.Network.Name,
		leaderIP,
		cluster.Token,
		masterPubKey,
		c.mnemonic,
		c.GridClient.Network,
	)
	if err != nil {
		return fmt.Errorf("failed to create VM for node: %v", err)
	}

	logger.GetLogger().Debug().
		Str("node_name", node.Name).
		Uint32("node_id", node.NodeID).
		Str("deployment_name", depl.Name).
		Msg("Deploying to grid")

	if err := c.GridClient.DeploymentDeployer.Deploy(ctx, &depl); err != nil {
		logger.GetLogger().Error().
			Err(err).
			Str("node_name", node.Name).
			Uint32("node_id", node.NodeID).
			Msg("Failed to deploy node to grid")
		return fmt.Errorf("failed to deploy node %s: %v", node.Name, err)
	}

	logger.GetLogger().Debug().
		Str("node_name", node.Name).
		Uint32("node_id", node.NodeID).
		Msg("Loading deployment result from grid")

	result, err := c.GridClient.State.LoadDeploymentFromGrid(ctx, node.NodeID, node.Name)
	if err != nil {
		return fmt.Errorf("failed to load deployment for node %s: %v", node.Name, err)
	}

	logger.GetLogger().Debug().
		Str("node_name", node.Name).
		Uint32("node_id", node.NodeID).
		Msg("Grid deployment successful")

	res := nodeFromDeployment(result)
	res.OriginalName = node.OriginalName
	res.Type = node.Type

	// used to handling adding new nodes or updating existing ones
	updated := false
	for i, n := range cluster.Nodes {
		if n.Name == res.Name {
			cluster.Nodes[i] = res
			updated = true
			logger.GetLogger().Info().
				Str("node_name", res.Name).
				Uint32("node_id", res.NodeID).
				Uint64("contract_id", res.ContractID).
				Msg("Updated existing node in cluster")
			break
		}
	}

	if !updated {
		cluster.Nodes = append(cluster.Nodes, res)
		logger.GetLogger().Debug().Str("node_name", res.Name).Msg("Added new node to cluster")
	}

	return nil
}

func (c *Client) BatchDeployNodes(ctx context.Context, cluster *Cluster, nodes []Node, masterPubKey string) error {
	if len(nodes) == 0 {
		return nil
	}

	logger.GetLogger().Debug().Msgf("Batch deploying %d nodes in cluster %s", len(nodes), cluster.Name)

	var leaderIP string
	leaderNode, err := cluster.GetLeaderNode()
	if err != nil {
		logger.GetLogger().Error().Err(err).Msgf("Failed to get leader node for cluster %s", cluster.Name)
		return fmt.Errorf("failed to get leader node IP: %v", err)
	}
	leaderIP = leaderNode.IP

	if cluster.Token == "" {
		cluster.Token = generateRandomString(32)
	}

	var deployments []*workloads.Deployment
	for _, node := range nodes {
		depl, err := deploymentFromNode(
			node,
			cluster.ProjectName,
			cluster.Network.Name,
			leaderIP,
			cluster.Token,
			masterPubKey,
			c.mnemonic,
			c.GridClient.Network,
		)
		if err != nil {
			return fmt.Errorf("failed to create VM for node %s: %v", node.Name, err)
		}
		deployments = append(deployments, &depl)
	}

	logger.GetLogger().Debug().Msgf("Starting batch deployment of %d nodes to grid", len(deployments))
	batchErr := c.GridClient.DeploymentDeployer.BatchDeploy(ctx, deployments)

	var successCount int
	var failedNodes []string

	for i, node := range nodes {
		if deployments[i].ContractID == 0 {
			failedNodes = append(failedNodes, node.Name)
			continue
		}
		result, err := c.GridClient.State.LoadDeploymentFromGrid(ctx, node.NodeID, node.Name)
		if err != nil {
			logger.GetLogger().Warn().Err(err).Str("node_name", node.Name).Msg("Failed to load deployment for node")
			failedNodes = append(failedNodes, node.Name)
			continue
		}

		res := nodeFromDeployment(result)

		res.OriginalName = nodes[i].OriginalName
		res.Type = nodes[i].Type

		for j, n := range cluster.Nodes {
			if n.Name == res.Name {
				cluster.Nodes[j] = res
				logger.GetLogger().Debug().Str("node_name", res.Name).Uint64("contract_id", res.ContractID).Msg("Updated existing node in cluster")
				break
			}
		}
		successCount++
	}

	logger.GetLogger().Debug().Int("successful", successCount).Int("failed", len(failedNodes)).Int("total", len(nodes)).Msg("Batch deployment completed")

	if len(failedNodes) > 0 {
		if batchErr != nil {
			return fmt.Errorf("failed to deploy %d nodes (%v): %v", len(failedNodes), failedNodes, batchErr)
		}
		return fmt.Errorf("failed to deploy %d nodes: %v", len(failedNodes), failedNodes)
	}

	return nil
}

func (c *Client) DeployNetwork(ctx context.Context, cluster *Cluster) error {
	seen := make(map[uint32]bool)
	nodeIDs := make([]uint32, 0, len(cluster.Nodes))
	for _, node := range cluster.Nodes {
		if !seen[node.NodeID] {
			seen[node.NodeID] = true
			nodeIDs = append(nodeIDs, node.NodeID)
		}
	}

	logger.GetLogger().Info().
		Str("network", cluster.Network.Name).
		Str("cluster", cluster.Name).
		Interface("node_ids", nodeIDs).
		Int("node_count", len(nodeIDs)).
		Msg("Deploying network")

	var net workloads.ZNet
	var err error

	// Check if we have an existing network (either pre-built from retry or already deployed)
	hasExistingNetwork := len(cluster.Network.NodeDeploymentID) > 0
	if hasExistingNetwork {
		logger.GetLogger().Debug().
			Str("network", cluster.Network.Name).
			Msg("Using existing network from cluster state")
		net = cluster.Network

		// Ensure all current nodes are prepared (this shouldn't affect already prepared nodes)
		for _, nodeID := range nodeIDs {
			found := false
			for _, existingNodeID := range net.Nodes {
				if existingNodeID == nodeID {
					found = true
					break
				}
			}
			if !found {
				net.Nodes = append(net.Nodes, nodeID)
			}
		}
		if net.MyceliumKeys == nil {
			net.MyceliumKeys = make(map[uint32][]byte)
		}
		for _, nodeID := range nodeIDs {
			if _, exists := net.MyceliumKeys[nodeID]; !exists {
				key, err := workloads.RandomMyceliumKey()
				if err != nil {
					return fmt.Errorf("failed to generate mycelium key for node %d: %w", nodeID, err)
				}
				net.MyceliumKeys[nodeID] = key
			}
		}

		logger.GetLogger().Debug().
			Str("network", cluster.Network.Name).
			Interface("total_nodes", net.Nodes).
			Msg("Network prepared")
	} else {
		// If the network is not deployed, then it is new network step
		logger.GetLogger().Debug().
			Str("network", cluster.Network.Name).
			Str("project", cluster.ProjectName).
			Interface("node_ids", nodeIDs).
			Msg("Creating new network workload")

		net, err = createNetworkWorkload(cluster.Network.Name, cluster.ProjectName, nodeIDs)
		if err != nil {
			return fmt.Errorf("failed to create network workload: %w", err)
		}
	}

	// Store the built network in cluster before deployment attempt
	// This ensures it's available in state if deployment fails
	cluster.Network = net

	logger.GetLogger().Debug().
		Str("network", net.Name).
		Interface("nodes", net.Nodes).
		Msg("Deploying network")
	if err := c.GridClient.NetworkDeployer.Deploy(ctx, &net); err != nil {
		// Update with what is already deployed
		cluster.Network = net
		return fmt.Errorf("failed to deploy network: %v", err)
	}

	// Update the network in the cluster with deployment results
	// The deployer may have modified the network struct (e.g., NodeDeploymentID)
	cluster.Network = net

	logger.GetLogger().Info().
		Str("network", net.Name).
		Int("contract_count", len(net.NodeDeploymentID)).
		Msg("Network deployed successfully")

	return nil
}

func (c *Client) CancelCluster(ctx context.Context, cluster *Cluster) error {
	clusterContracts, err := cluster.getAllClusterContracts()
	if err != nil {
		return fmt.Errorf("failed to get cluster contract IDs: %w", err)
	}

	if len(clusterContracts) == 0 {
		logger.GetLogger().Debug().
			Str("cluster", cluster.Name).
			Msg("No contracts found to cancel for cluster")
		return nil
	}

	logger.GetLogger().Debug().
		Str("cluster", cluster.Name).
		Int("contract_count", len(clusterContracts)).
		Interface("contract_ids", clusterContracts).
		Msg("Collected cluster contracts")

	if err := c.cancelNodeContracts(clusterContracts, cluster.Name); err != nil {
		return fmt.Errorf("failed to cancel cluster contracts: %w", err)
	}

	logger.GetLogger().Info().
		Str("cluster", cluster.Name).
		Int("contracts_attempted", len(clusterContracts)).
		Msg("Cluster cancellation completed successfully")

	return nil
}

func (c *Client) CancelAllContractsForUser(ctx context.Context, contractIDs []uint64) error {
	if len(contractIDs) == 0 {
		logger.GetLogger().Debug().Msg("No contracts to cancel for user")
		return nil
	}

	if err := c.cancelNodeContracts(contractIDs, "user"); err != nil {
		return fmt.Errorf("failed to cancel user contracts: %v", err)
	}

	logger.GetLogger().Info().
		Int("contracts_canceled", len(contractIDs)).
		Msg("User contracts canceled successfully")

	return nil
}
