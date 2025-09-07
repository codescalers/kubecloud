package statemanager

import (
	"encoding/json"
	"fmt"

	"kubecloud/kubedeployer"

	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/state"
	"github.com/xmonader/ewf"
	"kubecloud/internal/logger"
)

// GridClientState represents the critical state that needs to be preserved
type GridClientState struct {
	CurrentNodeDeployments map[uint32][]uint64          `json:"current_node_deployments"`
	NetworkSubnets         map[string]map[uint32]string `json:"network_subnets"`
}

// SaveGridClientState saves the critical GridClient state to workflow state
func SaveGridClientState(workflowState ewf.State, kubeClient *kubedeployer.Client) {
	if kubeClient == nil || kubeClient.GridClient.State == nil {
		return
	}

	gridState := GridClientState{
		CurrentNodeDeployments: make(map[uint32][]uint64),
		NetworkSubnets:         make(map[string]map[uint32]string),
	}

	// Save CurrentNodeDeployments
	for nodeID, contractIDs := range kubeClient.GridClient.State.CurrentNodeDeployments {
		gridState.CurrentNodeDeployments[nodeID] = []uint64(contractIDs)
	}

	// Save network subnet information
	for networkName, network := range kubeClient.GridClient.State.Networks.State {
		gridState.NetworkSubnets[networkName] = network.Subnets
	}

	// Store as JSON string in workflow state
	if stateBytes, err := json.Marshal(gridState); err == nil {
		workflowState["gridclient_state"] = string(stateBytes)
		logger.GetLogger().Debug().Msg("Saved GridClient state to workflow state")
	} else {
		logger.GetLogger().Warn().Err(err).Msg("Failed to marshal GridClient state")
	}
}

// RestoreGridClientState restores the critical GridClient state from workflow state
func RestoreGridClientState(workflowState ewf.State, kubeClient *kubedeployer.Client) error {
	if kubeClient == nil || kubeClient.GridClient.State == nil {
		return fmt.Errorf("invalid kubeclient or gridclient state")
	}

	stateStr, ok := workflowState["gridclient_state"].(string)
	if !ok || stateStr == "" {
		logger.GetLogger().Debug().Msg("No GridClient state found in workflow state")
		return nil // Not an error, just no state to restore
	}

	var savedState GridClientState
	if err := json.Unmarshal([]byte(stateStr), &savedState); err != nil {
		return fmt.Errorf("failed to unmarshal GridClient state: %w", err)
	}

	// Restore CurrentNodeDeployments
	kubeClient.GridClient.State.CurrentNodeDeployments = make(map[uint32]state.ContractIDs)
	for nodeID, contractIDs := range savedState.CurrentNodeDeployments {
		kubeClient.GridClient.State.CurrentNodeDeployments[nodeID] = state.ContractIDs(contractIDs)
	}

	// Restore network subnet information
	kubeClient.GridClient.State.Networks.State = make(map[string]state.Network)
	for networkName, subnets := range savedState.NetworkSubnets {
		kubeClient.GridClient.State.Networks.State[networkName] = state.Network{
			Subnets: subnets,
		}
	}

	logger.GetLogger().Debug().Msg("Restored GridClient state from workflow state")
	return nil
}
