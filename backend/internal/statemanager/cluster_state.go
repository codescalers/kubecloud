package statemanager

import (
	"encoding/json"
	"fmt"

	"kubecloud/kubedeployer"

	"github.com/xmonader/ewf"
	"kubecloud/internal/logger"
)

// GetCluster retrieves a cluster from workflow state with robust deserialization
func GetCluster(state ewf.State) (kubedeployer.Cluster, error) {
	value, ok := state["cluster"]
	if !ok {
		return kubedeployer.Cluster{}, fmt.Errorf("missing 'cluster' in state")
	}

	// Try direct type assertion first (for newly created clusters)
	if cluster, ok := value.(kubedeployer.Cluster); ok {
		return cluster, nil
	}

	// Handle JSON string format
	if clusterStr, ok := value.(string); ok {
		var cluster kubedeployer.Cluster
		if err := json.Unmarshal([]byte(clusterStr), &cluster); err != nil {
			return kubedeployer.Cluster{}, fmt.Errorf("failed to unmarshal cluster: %w", err)
		}
		return cluster, nil
	}

	// Fallback: handle as map/interface{} and convert to JSON
	clusterBytes, err := json.Marshal(value)
	if err != nil {
		return kubedeployer.Cluster{}, fmt.Errorf("failed to marshal cluster value: %w", err)
	}

	var cluster kubedeployer.Cluster
	if err := json.Unmarshal(clusterBytes, &cluster); err != nil {
		return kubedeployer.Cluster{}, fmt.Errorf("failed to unmarshal cluster: %w", err)
	}

	return cluster, nil
}

// StoreCluster safely stores the cluster in state using JSON marshaling
func StoreCluster(state ewf.State, cluster kubedeployer.Cluster) {
	// Use the cluster's custom marshaling and store as JSON string
	if jsonData, err := json.Marshal(cluster); err == nil {
		state["cluster"] = string(jsonData)
	} else {
		// Fallback to direct storage if marshaling fails
		logger.GetLogger().Warn().Err(err).Msg("Failed to marshal cluster, falling back to direct storage")
		state["cluster"] = cluster
	}
}
