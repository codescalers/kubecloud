package kubedeployer

import (
	"context"
	"fmt"
	"sync"

	"kubecloud/internal/infrastructure/logger"
	"kubecloud/internal/infrastructure/telemetry"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func getNodeToRemove(ctx context.Context, cluster *Cluster, nodeName string) (*Node, int, error) {
	_, span := getTracer().Start(ctx, "getNodeToRemove",
		trace.WithAttributes(
			attribute.String("cluster.name", cluster.Name),
			attribute.String("node.name", nodeName),
			attribute.Int("cluster.total_nodes", len(cluster.Nodes)),
		),
	)
	defer span.End()

	for i, node := range cluster.Nodes {
		if node.Name == nodeName {
			if node.Type == NodeTypeLeader {
				err := fmt.Errorf("cannot remove leader nodes")
				telemetry.RecordError(span, err)
				return nil, -1, err
			}
			span.SetAttributes(
				attribute.Int("node.index", i),
				attribute.String("node.type", string(node.Type)),
				attribute.Int64("node.contract_id", int64(node.ContractID)),
			)
			span.AddEvent("Node found for removal")
			return &cluster.Nodes[i], i, nil
		}
	}

	err := fmt.Errorf("node %s not found in cluster", nodeName)
	telemetry.RecordError(span, err)
	return nil, -1, err
}

func verifyNetworkNotUsedByOthers(ctx context.Context, cluster *Cluster, nodeToRemove *Node) bool {
	_, span := getTracer().Start(ctx, "verifyNetworkNotUsedByOthers",
		trace.WithAttributes(
			attribute.String("node.name", nodeToRemove.Name),
			attribute.Int("node.id", int(nodeToRemove.NodeID)),
			attribute.Int("cluster.total_nodes", len(cluster.Nodes)),
		),
	)
	defer span.End()

	for _, otherNode := range cluster.Nodes {
		if otherNode.Name == nodeToRemove.Name { // skip the node being removed
			continue
		}
		if otherNode.NodeID == nodeToRemove.NodeID { // multiple vms on same physical node
			span.SetAttributes(attribute.Bool("network.still_in_use", true))
			span.AddEvent("Network still in use by other nodes",
				trace.WithAttributes(
					attribute.String("other_node.name", otherNode.Name),
				),
			)
			return true
		}
	}
	span.SetAttributes(attribute.Bool("network.still_in_use", false))
	span.AddEvent("Network not used by other nodes")
	return false
}

func gatherContractsToCancel(ctx context.Context, cluster *Cluster, nodeToRemove *Node, networkStillInUse bool) []uint64 {
	_, span := getTracer().Start(ctx, "gatherContractsToCancel",
		trace.WithAttributes(
			attribute.String("node.name", nodeToRemove.Name),
			attribute.Int64("node.contract_id", int64(nodeToRemove.ContractID)),
			attribute.Bool("network.still_in_use", networkStillInUse),
		),
	)
	defer span.End()

	var contractsToCancel []uint64

	if nodeToRemove.ContractID != 0 {
		contractsToCancel = append(contractsToCancel, nodeToRemove.ContractID)
		span.AddEvent("Added node contract to cancel list",
			trace.WithAttributes(
				attribute.Int64("contract_id", int64(nodeToRemove.ContractID)),
			),
		)
	}

	// Add the network contract if it exists and is not still in use
	if networkContractID, exists := cluster.Network.NodeDeploymentID[nodeToRemove.NodeID]; exists && networkContractID != 0 {
		if !networkStillInUse {
			contractsToCancel = append(contractsToCancel, networkContractID)
			span.AddEvent("Added network contract to cancel list",
				trace.WithAttributes(
					attribute.Int64("contract_id", int64(networkContractID)),
				),
			)
		} else {
			span.AddEvent("Network contract still in use, not canceling",
				trace.WithAttributes(
					attribute.Int64("contract_id", int64(networkContractID)),
				),
			)
		}
	}

	span.SetAttributes(attribute.Int("contracts.total_to_cancel", len(contractsToCancel)))
	span.AddEvent("Contracts gathered for cancellation")
	return contractsToCancel
}

func (c *Client) isContractActive(ctx context.Context, contractID uint64) bool {
	_, span := getTracer().Start(ctx, "Client.isContractActive",
		trace.WithAttributes(
			attribute.Int64("contract.id", int64(contractID)),
		),
	)
	defer span.End()

	log := logger.ForOperation("kubedeployer", "check_contract_active")
	log.Debug().Uint64("contract_id", contractID).Msg("Checking if contract is active")
	_, err := c.GridClient.SubstrateConn.GetContract(contractID)
	isActive := err == nil

	span.SetAttributes(attribute.Bool("contract.is_active", isActive))
	if isActive {
		span.AddEvent("Contract is active")
	} else {
		span.AddEvent("Contract is not active or does not exist")
	}

	return isActive
}

func (c *Client) cancelNodeContracts(ctx context.Context, contractsToCancel []uint64, name string) error {
	_, span := getTracer().Start(ctx, "Client.cancelNodeContracts",
		trace.WithAttributes(
			attribute.String("node.name", name),
			attribute.Int("contracts.total", len(contractsToCancel)),
		),
	)
	defer span.End()

	log := logger.ForOperation("kubedeployer", "cancel_node_contracts")

	if len(contractsToCancel) == 0 {
		span.AddEvent("No contracts to cancel")
		log.Debug().Str("node_name", name).Msg("No contracts to cancel")
		return nil
	}

	existingContractsToCancel := make([]uint64, 0, len(contractsToCancel))
	activeContractsChan := make(chan uint64, len(contractsToCancel))
	var wg sync.WaitGroup

	for _, contractID := range contractsToCancel {
		wg.Add(1)
		go func(contractID uint64) {
			defer wg.Done()
			if c.isContractActive(ctx, contractID) {
				activeContractsChan <- contractID
			} else {
				span.AddEvent("Contract does not exist or already canceled", trace.WithAttributes(
					attribute.Int64("contract_id", int64(contractID)),
				))
				log.Warn().
					Uint64("contract_id", contractID).
					Str("node_name", name).
					Msg("Contract does not exist or already canceled, skipping")
			}
		}(contractID)
	}
	wg.Wait()

	close(activeContractsChan)
	for contractID := range activeContractsChan {
		existingContractsToCancel = append(existingContractsToCancel, contractID)
	}

	span.SetAttributes(
		attribute.Int("contracts.active", len(existingContractsToCancel)),
		attribute.Int("contracts.skipped", len(contractsToCancel)-len(existingContractsToCancel)),
	)

	span.AddEvent("Canceling active contracts",
		trace.WithAttributes(
			attribute.Int("contract.count", len(existingContractsToCancel)),
		),
	)

	log.Debug().
		Str("node_name", name).
		Interface("contract_ids", existingContractsToCancel).
		Msg("Canceling contracts")
	if err := c.GridClient.BatchCancelContract(existingContractsToCancel); err != nil {
		telemetry.RecordError(span, err)
		return fmt.Errorf("failed to cancel node and/or network contracts: %v", err)
	}

	span.AddEvent("Contracts canceled successfully")
	return nil
}

func updateNetworkWorkload(ctx context.Context, cluster *Cluster, removedNodeId uint32, networkStillInUse bool) {
	_, span := getTracer().Start(ctx, "updateNetworkWorkload",
		trace.WithAttributes(
			attribute.Int("node.id", int(removedNodeId)),
			attribute.Bool("network.still_in_use", networkStillInUse),
			attribute.String("network.name", cluster.Network.Name),
		),
	)
	defer span.End()

	log := logger.ForOperation("kubedeployer", "update_network_workload")

	network := cluster.Network

	_, exists := network.NodeDeploymentID[removedNodeId]
	if !exists || networkStillInUse {
		span.AddEvent("Network workload still in use, skipping cleanup",
			trace.WithAttributes(
				attribute.Bool("contract_exists", exists),
				attribute.Bool("network_in_use", networkStillInUse),
			),
		)
		log.Debug().
			Uint32("node_id", removedNodeId).
			Bool("network_in_use", networkStillInUse).
			Msg("Network workload still in use, skipping cleanup")
		return
	}

	span.AddEvent("Cleaning up network workload data")

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

	span.SetAttributes(
		attribute.Int("network.remaining_nodes", len(network.Nodes)),
	)
	span.AddEvent("Network workload data cleaned up successfully")
	log.Debug().
		Uint32("node_id", removedNodeId).
		Msg("Cleaned up network workload data for canceled network contract")
}

func removeNodeFromCluster(ctx context.Context, cluster *Cluster, nodeIndex int) {
	_, span := getTracer().Start(ctx, "removeNodeFromCluster",
		trace.WithAttributes(
			attribute.String("cluster.name", cluster.Name),
			attribute.Int("node.index", nodeIndex),
			attribute.Int("cluster.nodes_before", len(cluster.Nodes)),
		),
	)
	defer span.End()

	updatedNodes := make([]Node, 0, len(cluster.Nodes)-1)
	updatedNodes = append(updatedNodes, cluster.Nodes[:nodeIndex]...)
	updatedNodes = append(updatedNodes, cluster.Nodes[nodeIndex+1:]...)
	cluster.Nodes = updatedNodes

	span.SetAttributes(attribute.Int("cluster.nodes_after", len(cluster.Nodes)))
	span.AddEvent("Node removed from cluster")
}

// RemoveNode cancel the node contract on chain and remove it from the cluster in db
// also cancel the network contract and clean up the network workload in db if not used by other nodes
func (c *Client) RemoveNode(ctx context.Context, cluster *Cluster, nodeName string) error {
	_, span := getTracer().Start(ctx, "Client.RemoveNode",
		trace.WithAttributes(
			attribute.String("cluster.name", cluster.Name),
			attribute.String("node.name", nodeName),
		),
	)
	defer span.End()

	log := logger.ForOperation("kubedeployer", "remove_node")

	span.AddEvent("Finding node to remove")
	nodeToRemove, nodeIndex, err := getNodeToRemove(ctx, cluster, nodeName)
	if err != nil {
		telemetry.RecordError(span, err)
		return err
	}

	span.SetAttributes(
		attribute.Int("node.index", nodeIndex),
		attribute.String("node.type", string(nodeToRemove.Type)),
		attribute.Int64("node.contract_id", int64(nodeToRemove.ContractID)),
	)

	span.AddEvent("Verifying network usage")
	networkStillInUse := verifyNetworkNotUsedByOthers(ctx, cluster, nodeToRemove)
	span.SetAttributes(attribute.Bool("network.still_in_use", networkStillInUse))

	span.AddEvent("Gathering contracts to cancel")
	contractsToCancel := gatherContractsToCancel(ctx, cluster, nodeToRemove, networkStillInUse)
	span.SetAttributes(attribute.Int("contracts.to_cancel_count", len(contractsToCancel)))

	span.AddEvent("Canceling node contracts",
		trace.WithAttributes(
			attribute.Int("contract.count", len(contractsToCancel)),
		),
	)
	if err := c.cancelNodeContracts(ctx, contractsToCancel, nodeToRemove.Name); err != nil {
		telemetry.RecordError(span, err)
		return err
	}

	span.AddEvent("Updating network workload")
	updateNetworkWorkload(ctx, cluster, nodeToRemove.NodeID, networkStillInUse)

	span.AddEvent("Removing node from cluster")
	removeNodeFromCluster(ctx, cluster, nodeIndex)

	span.AddEvent("Node removed successfully")
	log.Debug().
		Str("node_name", nodeName).
		Str("cluster", cluster.Name).
		Msg("Successfully removed node from cluster")
	return nil
}
