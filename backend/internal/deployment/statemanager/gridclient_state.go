package statemanager

import (
	"kubecloud/internal/deployment/kubedeployer"

	"github.com/xmonader/ewf"
)

const gridClientStateKey = "gridclient_state"

// SaveGridClientState saves the GridClient state to the workflow state
func SaveGridClientState(workflowState ewf.State, kubeClient *kubedeployer.Client) {
	if kubeClient == nil || kubeClient.GridClient == nil {
		return
	}

	stateData, err := kubeClient.GridClient.GetState()
	if err != nil {
		return
	}

	workflowState[gridClientStateKey] = stateData
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
