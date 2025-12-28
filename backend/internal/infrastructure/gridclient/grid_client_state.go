package gridclient

import (
	"encoding/json"
	"fmt"

	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/state"
)

// GridClientState represents the critical state that needs to be preserved
type gridClientState struct {
	CurrentNodeDeployments map[uint32][]uint64          `json:"current_node_deployments"`
	NetworkSubnets         map[string]map[uint32]string `json:"network_subnets"`
}

// GetState returns the current GridClient state as JSON bytes
func (s *gridClient) GetState() ([]byte, error) {
	if s.gridClient.State == nil {
		return nil, fmt.Errorf("gridclient state is nil")
	}

	gridState := gridClientState{
		CurrentNodeDeployments: make(map[uint32][]uint64),
		NetworkSubnets:         make(map[string]map[uint32]string),
	}

	// Save CurrentNodeDeployments
	for nodeID, contractIDs := range s.gridClient.State.CurrentNodeDeployments {
		gridState.CurrentNodeDeployments[nodeID] = []uint64(contractIDs)
	}

	// Save network subnet information
	for networkName, network := range s.gridClient.State.Networks.State {
		gridState.NetworkSubnets[networkName] = network.Subnets
	}

	return json.Marshal(gridState)
}

// RestoreState restores the GridClient state from JSON bytes
func (s *gridClient) RestoreState(stateData []byte) error {
	if s.gridClient.State == nil {
		return fmt.Errorf("gridclient state is nil")
	}

	if len(stateData) == 0 {
		return nil // Nothing to restore
	}

	var savedState gridClientState
	if err := json.Unmarshal(stateData, &savedState); err != nil {
		return fmt.Errorf("failed to unmarshal state: %w", err)
	}

	// Restore CurrentNodeDeployments
	s.gridClient.State.CurrentNodeDeployments = make(map[uint32]state.ContractIDs)
	for nodeID, contractIDs := range savedState.CurrentNodeDeployments {
		s.gridClient.State.CurrentNodeDeployments[nodeID] = state.ContractIDs(contractIDs)
	}

	// Restore network subnet information
	s.gridClient.State.Networks.State = make(map[string]state.Network)
	for networkName, subnets := range savedState.NetworkSubnets {
		s.gridClient.State.Networks.State[networkName] = state.Network{
			Subnets: subnets,
		}
	}

	return nil
}
