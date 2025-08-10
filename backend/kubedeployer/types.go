package kubedeployer

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/workloads"
	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/zos"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

type NodeType string

const (
	NodeTypeWorker NodeType = "worker"
	NodeTypeMaster NodeType = "master"
	NodeTypeLeader NodeType = "leader"
)

type Cluster struct {
	Name  string `json:"name" validate:"required,min=3,alphanum"`
	Token string `json:"token" validate:"required"`
	Nodes []Node `json:"nodes" validate:"required,min=1,dive"`

	// Computed
	Network     workloads.ZNet `json:"network,omitempty"`
	ProjectName string         `json:"project_name,omitempty"`
}

type Node struct {
	Name   string   `json:"name" validate:"required,min=3,alphanum"`
	Type   NodeType `json:"type" validate:"required,oneof=worker master leader"`
	NodeID uint32   `json:"node_id" validate:"required"`

	CPU      uint8             `json:"cpu" validate:"required,gt=0"`
	Memory   uint64            `json:"memory" validate:"required,gt=0"`    // Memory in MB
	RootSize uint64            `json:"root_size" validate:"required,gt=0"` // Storage in MB
	DiskSize uint64            `json:"disk_size" validate:"required,gt=0"` // Storage in MB
	EnvVars  map[string]string `json:"env_vars"`

	// Optional fields
	Flist      string `json:"flist,omitempty"`
	Entrypoint string `json:"entrypoint,omitempty"`

	// Computed
	IP           string `json:"ip,omitempty"`
	MyceliumIP   string `json:"mycelium_ip,omitempty"`
	PlanetaryIP  string `json:"planetary_ip,omitempty"`
	ContractID   uint64 `json:"contract_id,omitempty"`
	OriginalName string `json:"original_name,omitempty"`
}

// MarshalJSON implements custom JSON marshaling for Cluster
func (c Cluster) MarshalJSON() ([]byte, error) {
	// Create a serializable version of the cluster
	serializable := struct {
		Name        string `json:"name"`
		Token       string `json:"token"`
		Nodes       []Node `json:"nodes"`
		ProjectName string `json:"project_name,omitempty"`
		// TODO: add new network object (serialized, minimal, mapped to workloads.ZNet)
		Network struct {
			Name             string            `json:"name"`
			Description      string            `json:"description"`
			Nodes            []uint32          `json:"nodes"`
			IPRange          string            `json:"ip_range"`
			AddWGAccess      bool              `json:"add_wg_access"`
			MyceliumKeys     map[string]string `json:"mycelium_keys,omitempty"` // base64 encoded
			SolutionType     string            `json:"solution_type"`
			AccessWGConfig   string            `json:"access_wg_config"`
			ExternalIP       *string           `json:"external_ip,omitempty"`
			ExternalSK       string            `json:"external_sk,omitempty"` // base64 encoded
			PublicNodeID     uint32            `json:"public_node_id"`
			NodesIPRange     map[string]string `json:"nodes_ip_range,omitempty"`
			NodeDeploymentID map[string]uint64 `json:"node_deployment_id,omitempty"`
			WGPort           map[string]int    `json:"wg_port,omitempty"`
			Keys             map[string]string `json:"keys,omitempty"` // base64 encoded
		} `json:"network,omitempty"`
	}{
		Name:        c.Name,
		Token:       c.Token,
		Nodes:       c.Nodes,
		ProjectName: c.ProjectName,
	}

	// Handle network serialization
	serializable.Network.Name = c.Network.Name
	serializable.Network.Description = c.Network.Description
	serializable.Network.Nodes = c.Network.Nodes
	// Convert IPRange - only if not zero value
	if !c.Network.IPRange.Nil() {
		serializable.Network.IPRange = c.Network.IPRange.String()
	}
	serializable.Network.AddWGAccess = c.Network.AddWGAccess
	serializable.Network.SolutionType = c.Network.SolutionType
	serializable.Network.AccessWGConfig = c.Network.AccessWGConfig
	serializable.Network.PublicNodeID = c.Network.PublicNodeID

	// Convert ExternalIP
	if c.Network.ExternalIP != nil {
		extIPStr := c.Network.ExternalIP.String()
		serializable.Network.ExternalIP = &extIPStr
	}

	// Convert ExternalSK - only if not zero
	if c.Network.ExternalSK != (wgtypes.Key{}) {
		serializable.Network.ExternalSK = base64.StdEncoding.EncodeToString(c.Network.ExternalSK[:])
	}

	// Convert maps with uint32 keys to string keys (only if not empty)
	if len(c.Network.MyceliumKeys) > 0 {
		serializable.Network.MyceliumKeys = make(map[string]string, len(c.Network.MyceliumKeys))
		for nodeID, myceliumKey := range c.Network.MyceliumKeys {
			serializable.Network.MyceliumKeys[fmt.Sprintf("%d", nodeID)] = base64.StdEncoding.EncodeToString(myceliumKey)
		}
	}

	if len(c.Network.NodesIPRange) > 0 {
		serializable.Network.NodesIPRange = make(map[string]string, len(c.Network.NodesIPRange))
		for nodeID, ipRange := range c.Network.NodesIPRange {
			if !ipRange.Nil() {
				serializable.Network.NodesIPRange[fmt.Sprintf("%d", nodeID)] = ipRange.String()
			}
		}
	}

	if len(c.Network.NodeDeploymentID) > 0 {
		serializable.Network.NodeDeploymentID = make(map[string]uint64, len(c.Network.NodeDeploymentID))
		for nodeID, deploymentID := range c.Network.NodeDeploymentID {
			serializable.Network.NodeDeploymentID[fmt.Sprintf("%d", nodeID)] = deploymentID
		}
	}

	if len(c.Network.WGPort) > 0 {
		serializable.Network.WGPort = make(map[string]int, len(c.Network.WGPort))
		for nodeID, port := range c.Network.WGPort {
			serializable.Network.WGPort[fmt.Sprintf("%d", nodeID)] = port
		}
	}

	if len(c.Network.Keys) > 0 {
		serializable.Network.Keys = make(map[string]string, len(c.Network.Keys))
		for nodeID, key := range c.Network.Keys {
			serializable.Network.Keys[fmt.Sprintf("%d", nodeID)] = base64.StdEncoding.EncodeToString(key[:])
		}
	}

	return json.Marshal(serializable)
}

// UnmarshalJSON implements custom JSON unmarshaling for Cluster
func (c *Cluster) UnmarshalJSON(data []byte) error {
	// First unmarshal into a temporary structure
	var temp struct {
		Name        string `json:"name"`
		Token       string `json:"token"`
		Nodes       []Node `json:"nodes"`
		ProjectName string `json:"project_name,omitempty"`
		Network     struct {
			Name             string            `json:"name"`
			Description      string            `json:"description"`
			Nodes            []uint32          `json:"nodes"`
			IPRange          string            `json:"ip_range"`
			AddWGAccess      bool              `json:"add_wg_access"`
			MyceliumKeys     map[string]string `json:"mycelium_keys"`
			SolutionType     string            `json:"solution_type"`
			AccessWGConfig   string            `json:"access_wg_config"`
			ExternalIP       *string           `json:"external_ip"`
			ExternalSK       string            `json:"external_sk"`
			PublicNodeID     uint32            `json:"public_node_id"`
			NodesIPRange     map[string]string `json:"nodes_ip_range"`
			NodeDeploymentID map[string]uint64 `json:"node_deployment_id"`
			WGPort           map[string]int    `json:"wg_port"`
			Keys             map[string]string `json:"keys"`
		} `json:"network"`
	}

	if err := json.Unmarshal(data, &temp); err != nil {
		return fmt.Errorf("failed to unmarshal cluster: %w", err)
	}

	// Set basic cluster fields
	c.Name = temp.Name
	c.Token = temp.Token
	c.Nodes = temp.Nodes
	c.ProjectName = temp.ProjectName

	// Initialize network with basic fields
	c.Network = workloads.ZNet{
		Name:             temp.Network.Name,
		Description:      temp.Network.Description,
		Nodes:            temp.Network.Nodes,
		AddWGAccess:      temp.Network.AddWGAccess,
		SolutionType:     temp.Network.SolutionType,
		AccessWGConfig:   temp.Network.AccessWGConfig,
		PublicNodeID:     temp.Network.PublicNodeID,
		MyceliumKeys:     make(map[uint32][]byte),
		NodesIPRange:     make(map[uint32]zos.IPNet),
		NodeDeploymentID: make(map[uint32]uint64),
		WGPort:           make(map[uint32]int),
		Keys:             make(map[uint32]wgtypes.Key),
	}

	// Parse IPRange
	if temp.Network.IPRange != "" {
		if ipNet, err := zos.ParseIPNet(temp.Network.IPRange); err != nil {
			return fmt.Errorf("failed to parse IP range '%s': %w", temp.Network.IPRange, err)
		} else {
			c.Network.IPRange = ipNet
		}
	}

	// Parse ExternalIP
	if temp.Network.ExternalIP != nil {
		if ipNet, err := zos.ParseIPNet(*temp.Network.ExternalIP); err != nil {
			return fmt.Errorf("failed to parse external IP '%s': %w", *temp.Network.ExternalIP, err)
		} else {
			c.Network.ExternalIP = &ipNet
		}
	}

	// Parse ExternalSK
	if temp.Network.ExternalSK != "" {
		if decoded, err := base64.StdEncoding.DecodeString(temp.Network.ExternalSK); err != nil {
			return fmt.Errorf("failed to decode external SK: %w", err)
		} else if len(decoded) != 32 {
			return fmt.Errorf("invalid external SK length: expected 32 bytes, got %d", len(decoded))
		} else {
			var key [32]byte
			copy(key[:], decoded)
			c.Network.ExternalSK = wgtypes.Key(key)
		}
	}

	// Helper function to convert string node ID to uint32
	parseNodeID := func(nodeIDStr string) (uint32, error) {
		nodeID, err := strconv.ParseUint(nodeIDStr, 10, 32)
		if err != nil {
			return 0, fmt.Errorf("invalid node ID '%s': %w", nodeIDStr, err)
		}
		return uint32(nodeID), nil
	}

	// Convert MyceliumKeys
	for nodeIDStr, myceliumKeyStr := range temp.Network.MyceliumKeys {
		nodeID, err := parseNodeID(nodeIDStr)
		if err != nil {
			return fmt.Errorf("failed to parse node ID for mycelium key: %w", err)
		}

		if decoded, err := base64.StdEncoding.DecodeString(myceliumKeyStr); err != nil {
			return fmt.Errorf("failed to decode mycelium key for node %d: %w", nodeID, err)
		} else {
			c.Network.MyceliumKeys[nodeID] = decoded
		}
	}

	// Convert NodesIPRange
	for nodeIDStr, ipRangeStr := range temp.Network.NodesIPRange {
		nodeID, err := parseNodeID(nodeIDStr)
		if err != nil {
			return fmt.Errorf("failed to parse node ID for IP range: %w", err)
		}

		if ipNet, err := zos.ParseIPNet(ipRangeStr); err != nil {
			return fmt.Errorf("failed to parse IP range '%s' for node %d: %w", ipRangeStr, nodeID, err)
		} else {
			c.Network.NodesIPRange[nodeID] = ipNet
		}
	}

	// Convert NodeDeploymentID
	for nodeIDStr, deploymentID := range temp.Network.NodeDeploymentID {
		nodeID, err := parseNodeID(nodeIDStr)
		if err != nil {
			return fmt.Errorf("failed to parse node ID for deployment ID: %w", err)
		}
		c.Network.NodeDeploymentID[nodeID] = deploymentID
	}

	// Convert WGPort
	for nodeIDStr, port := range temp.Network.WGPort {
		nodeID, err := parseNodeID(nodeIDStr)
		if err != nil {
			return fmt.Errorf("failed to parse node ID for WG port: %w", err)
		}
		c.Network.WGPort[nodeID] = port
	}

	// Convert Keys
	for nodeIDStr, keyStr := range temp.Network.Keys {
		nodeID, err := parseNodeID(nodeIDStr)
		if err != nil {
			return fmt.Errorf("failed to parse node ID for key: %w", err)
		}

		if decoded, err := base64.StdEncoding.DecodeString(keyStr); err != nil {
			return fmt.Errorf("failed to decode key for node %d: %w", nodeID, err)
		} else if len(decoded) != 32 {
			return fmt.Errorf("invalid key length for node %d: expected 32 bytes, got %d", nodeID, len(decoded))
		} else {
			var key [32]byte
			copy(key[:], decoded)
			c.Network.Keys[nodeID] = wgtypes.Key(key)
		}
	}

	return nil
}
