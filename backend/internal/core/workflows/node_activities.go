package workflows

import (
	"context"
	"fmt"
	"kubecloud/internal/core/models"
	"time"

	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/deployer"
	proxy "github.com/threefoldtech/tfgrid-sdk-go/grid-proxy/pkg/client"
	"github.com/xmonader/ewf"
)

// Node status constants
const (
	NodeRentable           = "rentable"
	NodeRented             = "rented"
	NodeHasActiveContracts = "NodeHasActiveContracts"
)

func ReserveNodeStep(userNodesRepo models.UserNodesRepository, gridClient deployer.TFPluginClient) ewf.StepFn {
	return func(ctx context.Context, state ewf.State) error {
		userIDVal, ok := state["user_id"]
		if !ok {
			return fmt.Errorf("missing 'user_id' in state")
		}

		userID, err := toInt(userIDVal)
		if err != nil {
			return fmt.Errorf("invalid 'user_id' in state: %w", err)
		}

		nodeIDVal, ok := state["node_id"]
		if !ok {
			return fmt.Errorf("missing 'node_id' in state")
		}

		nodeID, err := toUint32(nodeIDVal)
		if err != nil {
			return fmt.Errorf("invalid 'node_id' in state: %w", err)
		}

		mnemonicVal, ok := state["mnemonic"]
		if !ok {
			return fmt.Errorf("missing 'mnemonic' in state")
		}
		mnemonic, ok := mnemonicVal.(string)
		if !ok {
			return fmt.Errorf("'mnemonic' in state is not a string")
		}

		// Get Identity
		identity, err := gridClient.SubstrateConn.NewIdentityFromSr25519Phrase(mnemonic)
		if err != nil {
			return fmt.Errorf("failed to get identity from mnemonic: %w", err)
		}
		// Reserve the node
		contractID, err := gridClient.SubstrateConn.CreateRentContract(identity, nodeID, nil)
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

func UnreserveNodeStep(userNodesRepo models.UserNodesRepository, gridClient deployer.TFPluginClient) ewf.StepFn {
	return func(ctx context.Context, state ewf.State) error {
		contractIDVal, ok := state["contract_id"]
		if !ok {
			return fmt.Errorf("missing 'contract_id' in state")
		}

		contractID, err := toUint64(contractIDVal)
		if err != nil {
			return fmt.Errorf("invalid 'contract_id' in state: %w", err)
		}

		mnemonic, ok := state["mnemonic"].(string)
		if !ok {
			return fmt.Errorf("missing 'mnemonic' in state")
		}

		identity, err := gridClient.SubstrateConn.NewIdentityFromSr25519Phrase(mnemonic)
		if err != nil {
			return fmt.Errorf("failed to get identity from mnemonic %v", identity)
		}

		err = gridClient.SubstrateConn.CancelContract(identity, contractID)
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

		node, err := proxyClient.Node(ctx, nodeID)
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
