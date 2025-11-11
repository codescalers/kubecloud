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
	log := logger.ForOperation("kubedeployer", "assign_node_ip")
	log.Debug().
		Str("node_name", n.Name).
		Uint32("node_id", n.NodeID).
		Str("network", networkName).
		Msg("Assigning IP for node")

	ip, err := getIpForVm(ctx, gridClient, networkName, n.NodeID)
	if err != nil {
		return fmt.Errorf("failed to get IP for node %s: %v", n.Name, err)
	}

	n.IP = ip
	log.Debug().
		Str("node_name", n.Name).
		Uint32("node_id", n.NodeID).
		Str("ip", ip).
		Str("network", networkName).
		Msg("IP assigned successfully")

	return nil
}

func (c *Client) DeployNode(ctx context.Context, cluster *Cluster, node Node, masterPubKey string) error {
	log := logger.ForOperation("kubedeployer", "deploy_node")
	log.Info().
		Str("node_name", node.Name).
		Uint32("node_id", node.NodeID).
		Str("cluster", cluster.Name).
		Str("node_type", string(node.Type)).
		Msg("Starting node deployment")

	var leaderIP string
	if node.Type == NodeTypeLeader {
		leaderIP = ""
		log.Debug().Str("node_name", node.Name).Msg("Deploying as leader node")
	} else {
		leaderNode, err := cluster.GetLeaderNode()
		if err != nil {
			log.Error().Err(err).Str("cluster", cluster.Name).Msg("Failed to get leader node")
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

	log.Debug().
		Str("node_name", node.Name).
		Uint32("node_id", node.NodeID).
		Str("deployment_name", depl.Name).
		Msg("Deploying to grid")

	if err := c.GridClient.DeploymentDeployer.Deploy(ctx, &depl); err != nil {
		log.Error().
			Err(err).
			Str("node_name", node.Name).
			Uint32("node_id", node.NodeID).
			Msg("Failed to deploy node to grid")
		return fmt.Errorf("failed to deploy node %s: %v", node.Name, err)
	}

	log.Debug().
		Str("node_name", node.Name).
		Uint32("node_id", node.NodeID).
		Msg("Loading deployment result from grid")

	result, err := c.GridClient.State.LoadDeploymentFromGrid(ctx, node.NodeID, node.Name)
	if err != nil {
		return fmt.Errorf("failed to load deployment for node %s: %v", node.Name, err)
	}

	log.Debug().
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
			log.Info().
				Str("node_name", res.Name).
				Uint32("node_id", res.NodeID).
				Uint64("contract_id", res.ContractID).
				Msg("Updated existing node in cluster")
			break
		}
	}

	if !updated {
		cluster.Nodes = append(cluster.Nodes, res)
		log.Debug().Str("node_name", res.Name).Msg("Added new node to cluster")
	}

	return nil
}

func (c *Client) BatchDeployNodes(ctx context.Context, cluster *Cluster, nodes []Node, masterPubKey string) error {
	if len(nodes) == 0 {
		return nil
	}

	log := logger.ForOperation("kubedeployer", "batch_deploy_nodes")
	log.Debug().
		Int("node_count", len(nodes)).
		Str("cluster", cluster.Name).
		Msg("Batch deploying nodes")

	var leaderIP string
	leaderNode, err := cluster.GetLeaderNode()
	if err != nil {
		log.Error().Err(err).Str("cluster", cluster.Name).Msg("Failed to get leader node")
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

	log.Debug().
		Int("deployment_count", len(deployments)).
		Msg("Starting batch deployment to grid")
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
			log.Warn().Err(err).Str("node_name", node.Name).Msg("Failed to load deployment for node")
			failedNodes = append(failedNodes, node.Name)
			continue
		}

		res := nodeFromDeployment(result)

		res.OriginalName = nodes[i].OriginalName
		res.Type = nodes[i].Type

		for j, n := range cluster.Nodes {
			if n.Name == res.Name {
				cluster.Nodes[j] = res
				log.Debug().Str("node_name", res.Name).Uint64("contract_id", res.ContractID).Msg("Updated existing node in cluster")
				break
			}
		}
		successCount++
	}

	log.Debug().Int("successful", successCount).Int("failed", len(failedNodes)).Int("total", len(nodes)).Msg("Batch deployment completed")

	if len(failedNodes) > 0 {
		if batchErr != nil {
			return fmt.Errorf("failed to deploy %d nodes (%v): %v", len(failedNodes), failedNodes, batchErr)
		}
		return fmt.Errorf("failed to deploy %d nodes: %v", len(failedNodes), failedNodes)
	}

	return nil
}

func (c *Client) DeployNetwork(ctx context.Context, cluster *Cluster) error {
	log := logger.ForOperation("kubedeployer", "deploy_network")

	seen := make(map[uint32]bool)
	nodeIDs := make([]uint32, 0, len(cluster.Nodes))
	for _, node := range cluster.Nodes {
		if !seen[node.NodeID] {
			seen[node.NodeID] = true
			nodeIDs = append(nodeIDs, node.NodeID)
		}
	}

	log.Info().
		Str("network", cluster.Network.Name).
		Str("cluster", cluster.Name).
		Interface("node_ids", nodeIDs).
		Int("node_count", len(nodeIDs)).
		Msg("Deploying network")

	var net workloads.ZNet
	var err error

	if len(cluster.Network.NodeDeploymentID) > 0 {
		log.Debug().
			Str("network", cluster.Network.Name).
			Int("existing_nodes", len(cluster.Network.Nodes)).
			Msg("Updating existing network workload")

		net = cluster.Network

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
					return fmt.Errorf("failed to generate mycelium key for node %d: %v", nodeID, err)
				}
				net.MyceliumKeys[nodeID] = key
			}
		}

		log.Debug().
			Str("network", cluster.Network.Name).
			Interface("total_nodes", net.Nodes).
			Msg("Network update prepared")
	} else {
		log.Debug().
			Str("network", cluster.Network.Name).
			Str("project", cluster.ProjectName).
			Interface("node_ids", nodeIDs).
			Msg("Creating new network workload")

		net, err = createNetworkWorkload(cluster.Network.Name, cluster.ProjectName, nodeIDs)
		if err != nil {
			return fmt.Errorf("failed to create network workload: %v", err)
		}
	}

	log.Debug().
		Str("network", net.Name).
		Interface("nodes", net.Nodes).
		Msg("Deploying network")
	if err := c.GridClient.NetworkDeployer.Deploy(ctx, &net); err != nil {
		return fmt.Errorf("failed to deploy network: %v", err)
	}

	cluster.Network = net

	log.Info().
		Str("network", net.Name).
		Int("contract_count", len(net.NodeDeploymentID)).
		Msg("Network deployed successfully")

	return nil
}

func (c *Client) CancelCluster(ctx context.Context, cluster Cluster) error {
	log := logger.ForOperation("kubedeployer", "cancel_cluster")

	clusterContracts, err := cluster.getAllClusterContracts()
	if err != nil {
		return fmt.Errorf("failed to get cluster contract IDs: %v", err)
	}

	log.Debug().
		Str("cluster", cluster.Name).
		Int("contract_count", len(clusterContracts)).
		Interface("contract_ids", clusterContracts).
		Msg("Collected cluster contracts")

	if err := c.cancelNodeContracts(clusterContracts, cluster.Name); err != nil {
		return fmt.Errorf("failed to cancel cluster contracts: %v", err)
	}

	log.Info().
		Str("cluster", cluster.Name).
		Int("contracts_canceled", len(clusterContracts)).
		Msg("Cluster canceled successfully")

	return nil
}

func (c *Client) CancelAllContractsForUser(ctx context.Context, contractIDs []uint64) error {
	log := logger.ForOperation("kubedeployer", "cancel_user_contracts")

	if len(contractIDs) == 0 {
		log.Debug().Msg("No contracts to cancel for user")
		return nil
	}

	if err := c.cancelNodeContracts(contractIDs, "user"); err != nil {
		return fmt.Errorf("failed to cancel user contracts: %v", err)
	}

	log.Info().
		Int("contracts_canceled", len(contractIDs)).
		Msg("User contracts canceled successfully")

	return nil
}
