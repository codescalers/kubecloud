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
		mnemonic, err := getFromState[string](state, "mnemonic")
		if err != nil {
			return err
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
		userID, err := getFromState[int](state, "user_id")
		if err != nil {
			return err
		}

		nodeID, err := getFromState[uint32](state, "node_id")
		if err != nil {
			return err
		}

		identity, err := getFromState[substrate.Identity](state, "identity")
		if err != nil {
			return err
		}

		// Reserve the node
		contractID, err := substrateClient.CreateRentContract(identity, nodeID, nil)
		if err != nil {
			return fmt.Errorf("failed to create rent contract for node_id=%d (user_id=%d): %w", nodeID, userID, err)
		}

		err = db.CreateUserNode(&models.UserNodes{
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

func UnreserveNodeStep(db models.DB, substrateClient *substrate.Substrate) ewf.StepFn {
	return func(ctx context.Context, state ewf.State) error {
		contractID, err := getFromState[uint64](state, "contract_id")
		if err != nil {
			return err
		}

		mnemonic, err := getFromState[string](state, "mnemonic")
		if err != nil {
			return err
		}

		identity, err := substrate.NewIdentityFromSr25519Phrase(mnemonic)
		if err != nil {
			return fmt.Errorf("failed to create identity: %w", err)
		}

		err = substrateClient.CancelContract(identity, contractID)
		if err != nil {
			return fmt.Errorf("failed to cancel contract: %w", err)
		}

		err = db.DeleteUserNode(contractID)
		if err != nil {
			return fmt.Errorf("failed to delete user node: %w", err)
		}

		return nil
	}
}

// VerifyNodeStateStep checks if node has reached the desired state
func VerifyNodeStateStep(proxyClient proxy.Client) ewf.StepFn {
	return func(ctx context.Context, state ewf.State) error {
		targetStatus, err := getFromState[string](state, "target_status")
		if err != nil {
			return err
		}

		nodeID, err := getFromState[uint32](state, "node_id")
		if err != nil {
			return err
		}

		node, err := proxyClient.Node(ctx, nodeID)
		if err != nil {
			return fmt.Errorf("failed to get node: %w", err)
		}

		reached := targetStatus == constants.NodeRentable && node.Rentable || targetStatus == constants.NodeRented && !node.Rentable

		if !reached {
			return fmt.Errorf("node %d has not reached target status '%s' (current: rentable=%v)", nodeID, targetStatus, node.Rentable)
		}

		return nil
	}
}
