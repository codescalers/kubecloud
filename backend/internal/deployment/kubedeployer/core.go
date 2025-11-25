package kubedeployer

import (
	"context"
	"fmt"
	"slices"

	"kubecloud/internal/infrastructure/logger"

	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/deployer"
	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/workloads"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// getTracer returns a tracer for kubedeployer operations.
func getTracer() trace.Tracer {
	return otel.Tracer("kubedeployer")
}

func (c *Cluster) GetLeaderNode() (Node, error) {
	return c.Nodes[0], nil
}

func (n *Node) AssignNodeIP(ctx context.Context, gridClient deployer.TFPluginClient, networkName string) error {
	ctx, span := getTracer().Start(ctx, "Node.AssignNodeIP",
		trace.WithAttributes(
			attribute.String("node.name", n.Name),
			attribute.Int("node.id", int(n.NodeID)),
			attribute.String("network.name", networkName),
		),
	)
	defer span.End()

	log := logger.ForOperation("kubedeployer", "assign_node_ip")
	log.Debug().
		Str("node_name", n.Name).
		Uint32("node_id", n.NodeID).
		Str("network", networkName).
		Msg("Assigning IP for node")

	ip, err := getIpForVm(ctx, gridClient, networkName, n.NodeID)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to get IP for node %s: %v", n.Name, err)
	}

	n.IP = ip
	span.SetAttributes(attribute.String("node.ip", ip))
	span.AddEvent("IP assigned successfully")
	log.Debug().
		Str("node_name", n.Name).
		Uint32("node_id", n.NodeID).
		Str("ip", ip).
		Str("network", networkName).
		Msg("IP assigned successfully")

	return nil
}

func (c *Client) DeployNode(ctx context.Context, cluster *Cluster, node Node, masterPubKey string) error {
	ctx, span := getTracer().Start(ctx, "Client.DeployNode",
		trace.WithAttributes(
			attribute.String("node.name", node.Name),
			attribute.Int("node.id", int(node.NodeID)),
			attribute.String("cluster.name", cluster.Name),
			attribute.String("node.type", string(node.Type)),
			attribute.Int("node.cpu", int(node.CPU)),
			attribute.Int64("node.memory_mb", int64(node.Memory)),
			attribute.Int64("node.root_size_mb", int64(node.RootSize)),
			attribute.Int("node.data_disks_count", len(node.DataDisks)),
		),
	)
	defer span.End()

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
		span.AddEvent("Generated cluster token")
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
		span.RecordError(err)
		return fmt.Errorf("failed to create VM for node: %v", err)
	}

	span.AddEvent("Deploying to grid",
		trace.WithAttributes(
			attribute.String("deployment.name", depl.Name),
		),
	)
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
		span.RecordError(err)
		return fmt.Errorf("failed to deploy node %s: %v", node.Name, err)
	}

	span.AddEvent("Loading deployment result from grid")
	log.Debug().
		Str("node_name", node.Name).
		Uint32("node_id", node.NodeID).
		Msg("Loading deployment result from grid")

	result, err := c.GridClient.State.LoadDeploymentFromGrid(ctx, node.NodeID, node.Name)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to load deployment for node %s: %v", node.Name, err)
	}

	span.AddEvent("Grid deployment successful")
	log.Debug().
		Str("node_name", node.Name).
		Uint32("node_id", node.NodeID).
		Msg("Grid deployment successful")

	res := nodeFromDeployment(result)
	res.OriginalName = node.OriginalName
	res.Type = node.Type

	span.SetAttributes(
		attribute.String("node.deployed_ip", res.IP),
		attribute.String("node.mycelium_ip", res.MyceliumIP),
		attribute.Int64("node.contract_id", int64(res.ContractID)),
	)

	// used to handle adding new nodes or updating existing ones
	updated := false
	for i, n := range cluster.Nodes {
		if n.Name == res.Name {
			cluster.Nodes[i] = res
			updated = true
			span.AddEvent("Updated existing node in cluster")
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
		span.AddEvent("Added new node to cluster")
		log.Debug().Str("node_name", res.Name).Msg("Added new node to cluster")
	}

	span.AddEvent("Node deployment completed")
	log.Info().
		Str("node_name", res.Name).
		Uint32("node_id", res.NodeID).
		Uint64("contract_id", res.ContractID).
		Msg("Node deployment completed")

	return nil
}

func (c *Client) BatchDeployNodes(ctx context.Context, cluster *Cluster, nodes []Node, masterPubKey string) error {
	if len(nodes) == 0 {
		return nil
	}

	ctx, span := getTracer().Start(ctx, "Client.BatchDeployNodes",
		trace.WithAttributes(
			attribute.Int("node.count", len(nodes)),
			attribute.String("cluster.name", cluster.Name),
		),
	)
	defer span.End()

	log := logger.ForOperation("kubedeployer", "batch_deploy_nodes")
	log.Debug().
		Int("node_count", len(nodes)).
		Str("cluster", cluster.Name).
		Msg("Batch deploying nodes")

	var leaderIP string
	leaderNode, err := cluster.GetLeaderNode()
	if err != nil {
		log.Error().Err(err).Str("cluster", cluster.Name).Msg("Failed to get leader node")
		span.RecordError(err)
		return fmt.Errorf("failed to get leader node IP: %v", err)
	}
	leaderIP = leaderNode.IP
	span.SetAttributes(attribute.String("node.leader_ip", leaderIP))

	if cluster.Token == "" {
		span.AddEvent("Generating cluster token")
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
			span.RecordError(err)
			return fmt.Errorf("failed to create VM for node %s: %v", node.Name, err)
		}
		deployments = append(deployments, &depl)
	}

	span.AddEvent("Starting batch deployment to grid",
		trace.WithAttributes(
			attribute.Int("deployment.count", len(deployments)),
		),
	)
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

		span.AddEvent("Loading deployment result from grid",
			trace.WithAttributes(
				attribute.Int("node.index", i),
				attribute.String("node.name", node.Name),
				attribute.Int64("node.contract_id", int64(deployments[i].ContractID)),
			),
		)

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
				span.AddEvent("Updated existing node in cluster")
				log.Debug().Str("node_name", res.Name).Uint64("contract_id", res.ContractID).Msg("Updated existing node in cluster")
				break
			}
		}
		successCount++
	}

	span.SetAttributes(
		attribute.Int("deployment.successful", successCount),
		attribute.Int("deployment.failed", len(failedNodes)),
	)
	span.AddEvent("Batch deployment completed",
		trace.WithAttributes(
			attribute.Int("successful", successCount),
			attribute.Int("failed", len(failedNodes)),
		),
	)
	log.Debug().Int("successful", successCount).Int("failed", len(failedNodes)).Int("total", len(nodes)).Msg("Batch deployment completed")

	if len(failedNodes) > 0 {
		span.SetAttributes(attribute.StringSlice("deployment.failed_nodes", failedNodes))
		if batchErr != nil {
			span.RecordError(batchErr)
			return fmt.Errorf("failed to deploy %d nodes (%v): %v", len(failedNodes), failedNodes, batchErr)
		}
		err := fmt.Errorf("failed to deploy %d nodes: %v", len(failedNodes), failedNodes)
		span.RecordError(err)
		return err
	}

	return nil
}

func (c *Client) DeployNetwork(ctx context.Context, cluster *Cluster) error {
	ctx, span := getTracer().Start(ctx, "Client.DeployNetwork",
		trace.WithAttributes(
			attribute.String("network.name", cluster.Network.Name),
			attribute.String("cluster.name", cluster.Name),
		),
	)
	defer span.End()

	log := logger.ForOperation("kubedeployer", "deploy_network")

	seen := make(map[uint32]bool)
	nodeIDs := make([]uint32, 0, len(cluster.Nodes))
	for _, node := range cluster.Nodes {
		if !seen[node.NodeID] {
			seen[node.NodeID] = true
			nodeIDs = append(nodeIDs, node.NodeID)
		}
	}

	span.SetAttributes(
		attribute.Int("network.node_count", len(nodeIDs)),
	)
	log.Info().
		Str("network", cluster.Network.Name).
		Str("cluster", cluster.Name).
		Interface("node_ids", nodeIDs).
		Int("node_count", len(nodeIDs)).
		Msg("Deploying network")

	var net workloads.ZNet
	var err error

	if len(cluster.Network.NodeDeploymentID) > 0 {
		span.AddEvent("Updating existing network workload",
			trace.WithAttributes(
				attribute.Int("existing_nodes", len(cluster.Network.Nodes)),
			),
		)
		log.Debug().
			Str("network", cluster.Network.Name).
			Int("existing_nodes", len(cluster.Network.Nodes)).
			Msg("Updating existing network workload")

		net = cluster.Network

		for _, nodeID := range nodeIDs {
			if !slices.Contains(net.Nodes, nodeID) {
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
					span.RecordError(err)
					return fmt.Errorf("failed to generate mycelium key for node %d: %v", nodeID, err)
				}
				net.MyceliumKeys[nodeID] = key
			}
		}

		span.AddEvent("Network update prepared")

		log.Debug().
			Str("network", cluster.Network.Name).
			Interface("total_nodes", net.Nodes).
			Msg("Network update prepared")
	} else {
		span.AddEvent("Creating new network workload")
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

	span.AddEvent("Deploying network")
	err = c.GridClient.NetworkDeployer.Deploy(ctx, &net)
	cluster.Network = net
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to deploy network: %v", err)
	}

	log.Info().
		Str("network", net.Name).
		Int("contract_count", len(net.NodeDeploymentID)).
		Msg("Network deployed successfully")

	span.AddEvent("Network deployed successfully")
	return nil
}

func (c *Client) CancelCluster(ctx context.Context, cluster Cluster) error {
	_, span := getTracer().Start(ctx, "Client.CancelCluster",
		trace.WithAttributes(
			attribute.String("cluster.name", cluster.Name),
		),
	)
	defer span.End()

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

	if err := c.cancelNodeContracts(ctx, clusterContracts, cluster.Name); err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to cancel cluster contracts: %v", err)
	}

	log.Info().
		Str("cluster", cluster.Name).
		Int("contracts_canceled", len(clusterContracts)).
		Msg("Cluster canceled successfully")

	span.AddEvent("Cluster canceled successfully")

	return nil
}

func (c *Client) CancelAllContractsForUser(ctx context.Context, contractIDs []uint64) error {
	log := logger.ForOperation("kubedeployer", "cancel_user_contracts")
	_, span := getTracer().Start(ctx, "Client.CancelAllContractsForUser",
		trace.WithAttributes(
			attribute.Int("contract.count", len(contractIDs)),
		),
	)
	defer span.End()

	if len(contractIDs) == 0 {
		log.Debug().Msg("No contracts to cancel for user")
		span.AddEvent("No contracts to cancel for user")
		return nil
	}

	if err := c.cancelNodeContracts(ctx, contractIDs, "user"); err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to cancel user contracts: %v", err)
	}

	log.Info().
		Int("contracts_canceled", len(contractIDs)).
		Msg("User contracts canceled successfully")

	span.AddEvent("User contracts canceled successfully")

	return nil
}
