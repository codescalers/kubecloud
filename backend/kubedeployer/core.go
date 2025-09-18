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
	logger.GetLogger().Debug().Msgf("Assigning IP for node %s in network %s", n.Name, networkName)
	ip, err := getIpForVm(ctx, gridClient, networkName, n.NodeID)
	if err != nil {
		return fmt.Errorf("failed to get IP for node %s: %v", n.Name, err)
	}
	n.IP = ip
	return nil
}

func (c *Client) DeployNode(ctx context.Context, cluster *Cluster, node Node, masterPubKey string) error {
	logger.GetLogger().Debug().Msgf("Deploying node %s in cluster %s", node.Name, cluster.Name)
	var leaderIP string
	if node.Type == NodeTypeLeader {
		leaderIP = ""
	} else {
		leaderNode, err := cluster.GetLeaderNode()
		if err != nil {
			logger.GetLogger().Error().Err(err).Msgf("Failed to get leader node for cluster %s", cluster.Name)
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
	)
	if err != nil {
		return fmt.Errorf("failed to create VM for node: %v", err)
	}

	logger.GetLogger().Debug().Str("node_name", node.Name).Msg("Starting deployment to grid")
	if err := c.GridClient.DeploymentDeployer.Deploy(ctx, &depl); err != nil {
		logger.GetLogger().Error().Err(err).Str("node_name", node.Name).Msg("Failed to deploy node to grid")
		return fmt.Errorf("failed to deploy node %s: %v", node.Name, err)
	}

	result, err := c.GridClient.State.LoadDeploymentFromGrid(ctx, node.NodeID, node.Name)
	if err != nil {
		return fmt.Errorf("failed to load deployment for node %s: %v", node.Name, err)
	}
	logger.GetLogger().Debug().Str("node_name", node.Name).Msg("Grid deployment successful")

	res, err := nodeFromDeployment(result)
	if err != nil {
		logger.GetLogger().Error().Err(err).Str("node_name", node.Name).Msg("Failed to extract node from deployment")
		return fmt.Errorf("failed to get node from deployment: %v", err)
	}
	res.OriginalName = node.OriginalName
	res.Type = node.Type

	// used to handling adding new nodes or updating existing ones
	updated := false
	for i, n := range cluster.Nodes {
		if n.Name == res.Name {
			cluster.Nodes[i] = res
			updated = true
			logger.GetLogger().Debug().Str("node_name", res.Name).Msg("Updated existing node in cluster")
			break
		}
	}

	if !updated {
		cluster.Nodes = append(cluster.Nodes, res)
		logger.GetLogger().Debug().Str("node_name", res.Name).Msg("Added new node to cluster")
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

	var net workloads.ZNet
	var err error

	if len(cluster.Network.NodeDeploymentID) > 0 {
		logger.GetLogger().Debug().Msgf("updating network workload for network: %s", cluster.Network.Name)
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

		logger.GetLogger().Debug().Msgf("Appending nodes %v to existing network %s. Total nodes: %v", nodeIDs, cluster.Network.Name, net.Nodes)
	} else {
		logger.GetLogger().Debug().Msgf("Creating new network workload for network: %s", cluster.Network.Name)
		net, err = createNetworkWorkload(cluster.Network.Name, cluster.ProjectName, nodeIDs)
		if err != nil {
			return fmt.Errorf("failed to create network workload: %v", err)
		}
	}

	logger.GetLogger().Debug().Msgf("Deploying network %s with nodes %v", net.Name, net.Nodes)
	if err := c.GridClient.NetworkDeployer.Deploy(ctx, &net); err != nil {
		return fmt.Errorf("failed to deploy network: %v", err)
	}

	cluster.Network = net

	return nil
}

func (c *Client) CancelCluster(ctx context.Context, cluster Cluster) error {
	clusterContracts, err := cluster.getAllClusterContracts()
	if err != nil {
		return fmt.Errorf("failed to get cluster contract IDs: %v", err)
	}

	if err := c.cancelNodeContracts(clusterContracts, cluster.Name); err != nil {
		return fmt.Errorf("failed to cancel cluster contracts: %v", err)
	}

	return nil
}

func (c *Client) CancelAllContractsForUser(ctx context.Context, contractIDs []uint64) error {
	if len(contractIDs) == 0 {
		return nil
	}

	if err := c.cancelNodeContracts(contractIDs, "user"); err != nil {
		return fmt.Errorf("failed to cancel user contracts: %v", err)
	}

	return nil
}
