package activities

import (
	"context"
	"fmt"
	"kubecloud/internal/constants"
	"kubecloud/models"
	"time"

	substrate "github.com/threefoldtech/tfchain/clients/tfchain-client-go"
	proxy "github.com/threefoldtech/tfgrid-sdk-go/grid-proxy/pkg/client"
	"github.com/xmonader/ewf"
)

func CreateIdentityStep() ewf.StepFn {
	return func(ctx context.Context, state ewf.State) error {

		mnemonicVal, ok := state["mnemonic"]
		if !ok {
			return fmt.Errorf("missing 'mnemonic' in state")
		}
		mnemonic, ok := mnemonicVal.(string)
		if !ok {
			return fmt.Errorf("'mnemonic' in state is not a string")
		}
		identity, err := substrate.NewIdentityFromSr25519Phrase(mnemonic)
		if err != nil {
			return fmt.Errorf("failed to create identity: %w", err)
		}
		state["identity"] = identity
		return nil
	}
}

func ReserveNodeStep(db models.DB, substrateClient *substrate.Substrate) ewf.StepFn {
	return func(ctx context.Context, state ewf.State) error {
		userID, ok := state["user_id"].(int)
		if !ok {
			return fmt.Errorf("missing or invalid 'user_id' in state")
		}
		nodeID, ok := state["node_id"].(uint32)
		if !ok {
			return fmt.Errorf("missing or invalid 'node_id' in state")
		}
		identity, ok := state["identity"].(substrate.Identity)
		if !ok {
			return fmt.Errorf("missing or invalid 'identity' in state")
		}

		// Reserve the node
		contractID, err := substrateClient.CreateRentContract(identity, nodeID, nil)
		if err != nil {
			return fmt.Errorf("failed to create rent contract: %w", err)
		}

		err = db.CreateUserNode(&models.UserNodes{
			UserID:     userID,
			ContractID: contractID,
			NodeID:     nodeID,
			CreatedAt:  time.Now(),
		})
		if err != nil {
			return fmt.Errorf("failed to create user node: %w", err)
		}

		state["contract_id"] = contractID
		return nil
	}
}

func UnreserveNodeStep(db models.DB, substrateClient *substrate.Substrate) ewf.StepFn {
	return func(ctx context.Context, state ewf.State) error {
		contractID, ok := state["contract_id"].(uint32)
		if !ok {
			return fmt.Errorf("missing or invalid 'contract_id' in state")
		}
		mnemonic, ok := state["mnemonic"].(string)
		if !ok {
			return fmt.Errorf("missing or invalid 'mnemonic' in state")
		}

		identity, err := substrate.NewIdentityFromSr25519Phrase(mnemonic)
		if err != nil {
			return fmt.Errorf("failed to create identity: %w", err)
		}

		err = substrateClient.CancelContract(identity, uint64(contractID))
		if err != nil {
			return fmt.Errorf("failed to cancel contract: %w", err)
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

		reached := targetStatus == constants.NodeRentable && node.Rentable || targetStatus == constants.NodeRented && !node.Rentable

		if !reached {
			return fmt.Errorf("node %d has not reached target status '%s' (current: rentable=%v)", nodeIDUint32, targetStatus, node.Rentable)
		}

		return nil
	}
}
