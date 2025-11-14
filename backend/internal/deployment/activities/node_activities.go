package activities

import (
	"kubecloud/internal/shared"
	"context"
	"fmt"
	"kubecloud/internal/infrastructure/substrate"
	"kubecloud/internal/core/models"
	"time"

	proxy "github.com/threefoldtech/tfgrid-sdk-go/grid-proxy/pkg/client"
	"github.com/xmonader/ewf"
)

func ReserveNodeStep(userNodesRepo models.UserNodesRepository, substrateClient substrate.Substrate) ewf.StepFn {
	return func(ctx context.Context, state ewf.State) error {
		userID, ok := state["user_id"].(int)
		if !ok {
			return fmt.Errorf("missing or invalid 'user_id' in state")
		}
		nodeID, ok := state["node_id"].(uint32)
		if !ok {
			return fmt.Errorf("missing or invalid 'node_id' in state")
		}

		mnemonicVal, ok := state["mnemonic"]
		if !ok {
			return fmt.Errorf("missing 'mnemonic' in state")
		}
		mnemonic, ok := mnemonicVal.(string)
		if !ok {
			return fmt.Errorf("'mnemonic' in state is not a string")
		}

		// Reserve the node
		contractID, err := substrateClient.CreateRentContract(mnemonic, nodeID, nil)
		if err != nil {
			return fmt.Errorf("failed to create rent contract for node_id=%d (user_id=%d): %w", nodeID, userID, err)
		}

		err = userNodesRepo.CreateUserNode(&models.UserNodes{
			UserID:     userID,
			ContractID: contractID,
			NodeID:     nodeID,
			CreatedAt:  time.Now(),
		})
		if err != nil {
			return fmt.Errorf("failed to create user node record for node_id=%d (user_id=%d): %w", nodeID, userID, err)
		}

		state["contract_id"] = contractID
		return nil
	}
}

func UnreserveNodeStep(userNodesRepo models.UserNodesRepository, substrateClient substrate.Substrate) ewf.StepFn {
	return func(ctx context.Context, state ewf.State) error {
		contractID, ok := state["contract_id"].(uint64)
		if !ok {
			return fmt.Errorf("missing or invalid 'contract_id' in state")
		}
		mnemonic, ok := state["mnemonic"].(string)
		if !ok {
			return fmt.Errorf("missing or invalid 'mnemonic' in state")
		}

		err := substrateClient.CancelContract(mnemonic, contractID)
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
func VerifyNodeStateStep(proxyClient proxy.Client) ewf.StepFn {
	return func(ctx context.Context, state ewf.State) error {
		targetStatus, ok := state["target_status"].(string)
		if !ok {
			return fmt.Errorf("missing or invalid 'target_status' in state")
		}

		nodeID, exists := state["node_id"]
		if !exists {
			return fmt.Errorf("missing or invalid 'node_id' in state")
		}

		nodeIDUint32, ok := nodeID.(uint32)
		if !ok {
			return fmt.Errorf("node_id in state is not a uint32")
		}

		node, err := proxyClient.Node(ctx, nodeIDUint32)
		if err != nil {
			return fmt.Errorf("failed to get node: %w", err)
		}

		reached := targetStatus == shared.NodeRentable && node.Rentable || targetStatus == shared.NodeRented && !node.Rentable

		if !reached {
			return fmt.Errorf("node %d has not reached target status '%s' (current: rentable=%v)", nodeIDUint32, targetStatus, node.Rentable)
		}

		return nil
	}
}
