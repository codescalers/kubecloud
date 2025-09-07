package statemanager

import (
	"fmt"

	"kubecloud/kubedeployer"

	"github.com/xmonader/ewf"
	"kubecloud/internal/logger"
)

// ClientConfig represents the configuration needed to create a kubeclient
type ClientConfig struct {
	SSHPublicKey string `json:"ssh_public_key"`
	Mnemonic     string `json:"mnemonic"`
	UserID       string `json:"user_id"`
	Network      string `json:"network"`
	Debug        bool   `json:"debug"`
}

// ValidateConfig validates the client configuration
func ValidateConfig(config ClientConfig) error {
	if config.Mnemonic == "" {
		return fmt.Errorf("missing mnemonic in config")
	}
	if config.Network == "" {
		return fmt.Errorf("missing network in config")
	}
	if config.SSHPublicKey == "" {
		return fmt.Errorf("missing SSH public key in config")
	}
	if config.UserID == "" {
		return fmt.Errorf("missing user ID in config")
	}

	return nil
}

// GetKubeClient retrieves or creates a kubeclient with proper state management
func GetKubeClient(state ewf.State, config ClientConfig) (*kubedeployer.Client, error) {
	// Try to get existing kubeclient from state
	if value, ok := state["kubeclient"]; ok {
		if client, ok := value.(*kubedeployer.Client); ok && client != nil {
			logger.GetLogger().Debug().Msg("Reusing existing kubeclient from state")
			return client, nil
		}
		// If we found an invalid client, remove it from state
		delete(state, "kubeclient")
		logger.GetLogger().Warn().Msg("Removed invalid kubeclient from state")
	}

	// Validate configuration before creating client
	if err := ValidateConfig(config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	// Create new client
	kubeClient, err := kubedeployer.NewClient(config.Mnemonic, config.Network, config.Debug)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubeclient: %w", err)
	}

	// Restore GridClient state if it exists
	if err := RestoreGridClientState(state, kubeClient); err != nil {
		logger.GetLogger().Warn().Err(err).Msg("Failed to restore GridClient state, continuing with fresh state")
	}

	// Store the new client in state for reuse
	state["kubeclient"] = kubeClient
	SaveGridClientState(state, kubeClient)

	logger.GetLogger().Debug().Msg("Created and stored fresh kubeclient")
	return kubeClient, nil
}

// GetExistingKubeClient returns an existing kubeclient from state without creating a new one
func GetExistingKubeClient(state ewf.State) (*kubedeployer.Client, error) {
	value, ok := state["kubeclient"]
	if !ok {
		return nil, fmt.Errorf("no kubeclient found in state")
	}

	client, ok := value.(*kubedeployer.Client)
	if !ok || client == nil {
		return nil, fmt.Errorf("invalid kubeclient in state")
	}

	return client, nil
}

// CloseClient properly closes a kubeclient and cleans up state
func CloseClient(state ewf.State, kubeClient *kubedeployer.Client) error {
	if kubeClient == nil {
		logger.GetLogger().Debug().Msg("No kubeclient to close")
		return nil
	}

	SaveGridClientState(state, kubeClient)
	kubeClient.Close()
	delete(state, "kubeclient")

	logger.GetLogger().Debug().Msg("Closed kubeclient and cleaned up state")
	return nil
}
