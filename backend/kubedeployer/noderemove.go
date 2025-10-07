package kubedeployer

import (
	"context"
	"fmt"
	"sync"

	"kubecloud/internal/logger"
)

func getNodeToRemove(cluster *Cluster, nodeName string) (*Node, int, error) {
	for i, node := range cluster.Nodes {
		if node.Name == nodeName {
			if node.Type == NodeTypeLeader {
				return nil, -1, fmt.Errorf("cannot remove leader nodes")
			}
			return &cluster.Nodes[i], i, nil
		}
	}
	return nil, -1, fmt.Errorf("node %s not found in cluster", nodeName)
}

func verifyNetworkNotUsedByOthers(cluster *Cluster, nodeToRemove *Node) bool {
	for _, otherNode := range cluster.Nodes {
		if otherNode.Name == nodeToRemove.Name { // skip the node being removed
			continue
		}
		if otherNode.NodeID == nodeToRemove.NodeID { // multiple vms on same physical node
			return true
		}
	}
	return false
}

func gatherContractsToCancel(cluster *Cluster, nodeToRemove *Node, networkStillInUse bool) []uint64 {
	var contractsToCancel []uint64

	if nodeToRemove.ContractID != 0 {
		contractsToCancel = append(contractsToCancel, nodeToRemove.ContractID)
	}

	// Add the network contract if it exists and is not still in use
	if networkContractID, exists := cluster.Network.NodeDeploymentID[nodeToRemove.NodeID]; exists && networkContractID != 0 {
		if !networkStillInUse {
			contractsToCancel = append(contractsToCancel, networkContractID)
		}
	}

	return contractsToCancel
}

func (c *Client) isContractActive(contractID uint64) bool {
	logger.GetLogger().Debug().Msgf("Checking if contract %d is active", contractID)
	_, err := c.GridClient.SubstrateConn.GetContract(contractID)
	return err == nil
}

func (c *Client) cancelNodeContracts(contractsToCancel []uint64, name string) error {
	if len(contractsToCancel) == 0 {
		logger.GetLogger().Debug().Msgf("No contracts to cancel for node %q", name)
		return nil
	}

	existingContractsToCancel := make([]uint64, 0, len(contractsToCancel))
	activeContractsChan := make(chan uint64, len(contractsToCancel))
	var wg sync.WaitGroup

	for _, contractID := range contractsToCancel {
		wg.Add(1)
		go func(contractID uint64) {
			defer wg.Done()
			if c.isContractActive(contractID) {
				activeContractsChan <- contractID
			} else {
				logger.GetLogger().Warn().Msgf("Contract %d for node %q does not exist or already canceled, skipping", contractID, name)
			}
		}(contractID)
	}
	wg.Wait()

	close(activeContractsChan)
	for contractID := range activeContractsChan {
		existingContractsToCancel = append(existingContractsToCancel, contractID)
	}

	logger.GetLogger().Debug().Msgf("Canceling contracts for %q: %v", name, existingContractsToCancel)
	if err := c.GridClient.BatchCancelContract(existingContractsToCancel); err != nil {
		return fmt.Errorf("failed to cancel node and/or network contracts: %v", err)
	}

	return nil
}

func updateNetworkWorkload(cluster *Cluster, removedNodeId uint32, networkStillInUse bool) {
	network := cluster.Network

	_, exists := network.NodeDeploymentID[removedNodeId]
	if !exists || networkStillInUse {
		logger.GetLogger().Debug().Msgf("Network workload for node_id %d still in use, skipping cleanup", removedNodeId)
		return
	}

	var updatedNetworkNodes []uint32
	for _, nodeID := range network.Nodes {
		if nodeID != removedNodeId {
			updatedNetworkNodes = append(updatedNetworkNodes, nodeID)
		}
	}
	network.Nodes = updatedNetworkNodes

	delete(network.NodeDeploymentID, removedNodeId)

	if network.NodesIPRange != nil {
		delete(network.NodesIPRange, removedNodeId)
	}
	if network.MyceliumKeys != nil {
		delete(network.MyceliumKeys, removedNodeId)
	}
	if network.Keys != nil {
		delete(network.Keys, removedNodeId)
	}
	if network.WGPort != nil {
		delete(network.WGPort, removedNodeId)
	}

	cluster.Network = network
	logger.GetLogger().Debug().Uint32("node_id", removedNodeId).Msg("Cleaned up network workload data for canceled network contract")
}

func removeNodeFromCluster(cluster *Cluster, nodeIndex int) {
	updatedNodes := make([]Node, 0, len(cluster.Nodes)-1)
	updatedNodes = append(updatedNodes, cluster.Nodes[:nodeIndex]...)
	updatedNodes = append(updatedNodes, cluster.Nodes[nodeIndex+1:]...)
	cluster.Nodes = updatedNodes
}

// RemoveNode cancel the node contract on chain and remove it from the cluster in db
// also cancel the network contract and clean up the network workload in db if not used by other nodes
func (c *Client) RemoveNode(ctx context.Context, cluster *Cluster, nodeName string) error {
	nodeToRemove, nodeIndex, err := getNodeToRemove(cluster, nodeName)
	if err != nil {
		return err
	}

	networkStillInUse := verifyNetworkNotUsedByOthers(cluster, nodeToRemove)

	contractsToCancel := gatherContractsToCancel(cluster, nodeToRemove, networkStillInUse)

	if err := c.cancelNodeContracts(contractsToCancel, nodeToRemove.Name); err != nil {
		return err
	}

	updateNetworkWorkload(cluster, nodeToRemove.NodeID, networkStillInUse)

	removeNodeFromCluster(cluster, nodeIndex)

	logger.GetLogger().Debug().Msgf("Successfully removed node %s from cluster %s", nodeName, cluster.Name)
	return nil
}
