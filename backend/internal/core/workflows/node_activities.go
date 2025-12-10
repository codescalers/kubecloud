package workflows

import (
	"context"
	"fmt"
	"kubecloud/internal/core/models"
	"kubecloud/internal/infrastructure/gridclient"
	"time"

	"github.com/xmonader/ewf"
)

// Node status constants
const (
	NodeRentable           = "rentable"
	NodeRented             = "rented"
	NodeHasActiveContracts = "NodeHasActiveContracts"
)

func ReserveNodeStep(userNodesRepo models.UserNodesRepository, gridClient gridclient.GridClient) ewf.StepFn {
	return func(ctx context.Context, state ewf.State) error {
		config, err := getConfig(state)
		if err != nil {
			return fmt.Errorf("failed to get config from state: %w", err)
		}

		nodeIDVal, ok := state["node_id"]
		if !ok {
			return fmt.Errorf("missing 'node_id' in state")
		}

		nodeID, err := toUint32(nodeIDVal)
		if err != nil {
			return fmt.Errorf("invalid 'node_id' in state: %w", err)
		}

		contractID, err := gridClient.CreateRentContract(config.Mnemonic, nodeID)
		if err != nil {
			return fmt.Errorf("failed to create rent contract for node_id=%d (user_id=%d): %w", nodeID, config.UserID, err)
		}

		err = userNodesRepo.CreateUserNode(&models.UserNodes{
			UserID:     config.UserID,
			ContractID: contractID,
			NodeID:     nodeID,
			CreatedAt:  time.Now(),
		})
		if err != nil {
			return fmt.Errorf("failed to create user node record for node_id=%d (user_id=%d): %w", nodeID, config.UserID, err)
		}

		state["contract_id"] = contractID
		return nil
	}
}

func UnreserveNodeStep(userNodesRepo models.UserNodesRepository, gridClient gridclient.GridClient) ewf.StepFn {
	return func(ctx context.Context, state ewf.State) error {
		contractIDVal, ok := state["contract_id"]
		if !ok {
			return fmt.Errorf("missing 'contract_id' in state")
		}

		contractID, err := toUint64(contractIDVal)
		if err != nil {
			return fmt.Errorf("invalid 'contract_id' in state: %w", err)
		}

		config, err := getConfig(state)
		if err != nil {
			return fmt.Errorf("failed to get config from state: %w", err)
		}

		err = gridClient.CancelContract(config.Mnemonic, contractID)
		if err != nil {
			return fmt.Errorf("failed to cancel contract: %w", err)
		}

		err = userNodesRepo.DeleteUserNode(contractID)
		if err != nil {
			return fmt.Errorf("failed to delete user node: %w", err)
		}

		return nil
	}
}

// VerifyNodeStateStep checks if node has reached the desired state
func VerifyNodeStateStep(gridClient gridclient.GridClient) ewf.StepFn {
	return func(ctx context.Context, state ewf.State) error {
		targetStatus, ok := state["target_status"].(string)
		if !ok {
			return fmt.Errorf("missing 'target_status' in state")
		}

		nodeIDVal, exists := state["node_id"]
		if !exists {
			return fmt.Errorf("missing 'node_id' in state")
		}

		nodeID, err := toUint32(nodeIDVal)
		if err != nil {
			return fmt.Errorf("invalid 'node_id' in state: %w", err)
		}

		node, err := gridClient.Node(ctx, nodeID)
		if err != nil {
			return fmt.Errorf("failed to get node: %w", err)
		}

		reached := targetStatus == NodeRentable && node.Rentable || targetStatus == NodeRented && !node.Rentable

		if !reached {
			return fmt.Errorf("node %d has not reached target status '%s' (current: rentable=%v)", nodeID, targetStatus, node.Rentable)
		}

		return nil
	}
}
