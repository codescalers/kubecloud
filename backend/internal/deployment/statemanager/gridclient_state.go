package statemanager

import (
	"fmt"
	"kubecloud/internal/deployment/kubedeployer"
	"kubecloud/internal/infrastructure/logger"

	"github.com/xmonader/ewf"
)

const gridClientStateKey = "gridclient_state"

// SaveGridClientState saves the GridClient state to the workflow state
func SaveGridClientState(workflowState ewf.State, kubeClient *kubedeployer.Client) error {
	log := logger.ForOperation("statemanager", "save_gridclient_state")

	if kubeClient == nil || kubeClient.GridClient == nil {
		return fmt.Errorf("kubeClient or its GridClient is nil")
	}

	stateData, err := kubeClient.GridClient.GetState()
	if err != nil {
		return fmt.Errorf("failed to get GridClient state: %w", err)
	}

	workflowState[gridClientStateKey] = stateData
	log.Debug().Msg("GridClient state saved successfully")
	return nil
}

// RestoreGridClientState restores the GridClient state from the workflow state
func RestoreGridClientState(workflowState ewf.State, kubeClient *kubedeployer.Client) error {
	if kubeClient == nil || kubeClient.GridClient == nil {
		return nil
	}

	stateData, ok := workflowState[gridClientStateKey]
	if !ok {
		return nil
	}

	stateBytes, ok := stateData.([]byte)
	if !ok {
		return nil
	}

	return kubeClient.GridClient.RestoreState(stateBytes)
}
