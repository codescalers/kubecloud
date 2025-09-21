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

type zNetJSON struct {
	Name             string            `json:"name"`
	Description      string            `json:"description"`
	Nodes            []uint32          `json:"nodes"`
	IPRange          string            `json:"ip_range"`
	AddWGAccess      bool              `json:"add_wg_access"`
	MyceliumKeys     map[string]string `json:"mycelium_keys,omitempty"`
	SolutionType     string            `json:"solution_type"`
	AccessWGConfig   string            `json:"access_wg_config"`
	ExternalIP       *string           `json:"external_ip,omitempty"`
	ExternalSK       string            `json:"external_sk,omitempty"`
	PublicNodeID     uint32            `json:"public_node_id"`
	NodesIPRange     map[string]string `json:"nodes_ip_range,omitempty"`
	NodeDeploymentID map[string]uint64 `json:"node_deployment_id,omitempty"`
	WGPort           map[string]int    `json:"wg_port,omitempty"`
	Keys             map[string]string `json:"keys,omitempty"`
}

type Cluster struct {
	Name  string `json:"name" binding:"required,min=3,max=20,alphanum"`
	Token string `json:"token"`
	Nodes []Node `json:"nodes" binding:"required,min=1,dive"`

	// Computed
	Network     workloads.ZNet `json:"network,omitempty"`
	ProjectName string         `json:"project_name,omitempty"`
}

type Node struct {
	Name   string   `json:"name" binding:"required,min=3,max=20,alphanum"`
	Type   NodeType `json:"type" binding:"required,oneof=worker master leader"`
	NodeID uint32   `json:"node_id" binding:"required"`

	CPU      uint8             `json:"cpu" binding:"required,min=1"`
	Memory   uint64            `json:"memory" binding:"required,min=2048"`     // Memory in MB
	RootSize uint64            `json:"root_size" binding:"required,min=5120"`  // Storage in MB
	DiskSize uint64            `json:"disk_size" binding:"required,min=10240"` // Storage in MB
	GPUIDs   []string          `json:"gpu_ids,omitempty"`                      // List of GPU IDs
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

// marshalZNet converts workloads.ZNet → serializable form
func marshalZNet(z workloads.ZNet) zNetJSON {
	out := zNetJSON{
		Name:           z.Name,
		Description:    z.Description,
		Nodes:          z.Nodes,
		AddWGAccess:    z.AddWGAccess,
		SolutionType:   z.SolutionType,
		AccessWGConfig: z.AccessWGConfig,
		PublicNodeID:   z.PublicNodeID,
	}

	if !z.IPRange.Nil() {
		out.IPRange = z.IPRange.String()
	}
	if z.ExternalIP != nil {
		ipStr := z.ExternalIP.String()
		out.ExternalIP = &ipStr
	}
	if z.ExternalSK != (wgtypes.Key{}) {
		out.ExternalSK = base64.StdEncoding.EncodeToString(z.ExternalSK[:])
	}

	if len(z.MyceliumKeys) > 0 {
		out.MyceliumKeys = make(map[string]string, len(z.MyceliumKeys))
		for nodeID, key := range z.MyceliumKeys {
			out.MyceliumKeys[strconv.FormatUint(uint64(nodeID), 10)] = base64.StdEncoding.EncodeToString(key)
		}
	}
	if len(z.NodesIPRange) > 0 {
		out.NodesIPRange = make(map[string]string, len(z.NodesIPRange))
		for nodeID, ip := range z.NodesIPRange {
			if !ip.Nil() {
				out.NodesIPRange[strconv.FormatUint(uint64(nodeID), 10)] = ip.String()
			}
		}
	}
	if len(z.NodeDeploymentID) > 0 {
		out.NodeDeploymentID = make(map[string]uint64, len(z.NodeDeploymentID))
		for nodeID, id := range z.NodeDeploymentID {
			out.NodeDeploymentID[strconv.FormatUint(uint64(nodeID), 10)] = id
		}
	}
	if len(z.WGPort) > 0 {
		out.WGPort = make(map[string]int, len(z.WGPort))
		for nodeID, port := range z.WGPort {
			out.WGPort[strconv.FormatUint(uint64(nodeID), 10)] = port
		}
	}
	if len(z.Keys) > 0 {
		out.Keys = make(map[string]string, len(z.Keys))
		for nodeID, key := range z.Keys {
			out.Keys[strconv.FormatUint(uint64(nodeID), 10)] = base64.StdEncoding.EncodeToString(key[:])
		}
	}

	return out
}

// unmarshalZNet converts zNetJSON → workloads.ZNet
func unmarshalZNet(in zNetJSON) (workloads.ZNet, error) {
	z := workloads.ZNet{
		Name:             in.Name,
		Description:      in.Description,
		Nodes:            in.Nodes,
		AddWGAccess:      in.AddWGAccess,
		SolutionType:     in.SolutionType,
		AccessWGConfig:   in.AccessWGConfig,
		PublicNodeID:     in.PublicNodeID,
		MyceliumKeys:     make(map[uint32][]byte),
		NodesIPRange:     make(map[uint32]zos.IPNet),
		NodeDeploymentID: make(map[uint32]uint64),
		WGPort:           make(map[uint32]int),
		Keys:             make(map[uint32]wgtypes.Key),
	}

	if in.IPRange != "" {
		ipNet, err := zos.ParseIPNet(in.IPRange)
		if err != nil {
			return z, fmt.Errorf("invalid IP range: %w", err)
		}
		z.IPRange = ipNet
	}
	if in.ExternalIP != nil {
		ipNet, err := zos.ParseIPNet(*in.ExternalIP)
		if err != nil {
			return z, fmt.Errorf("invalid external IP: %w", err)
		}
		z.ExternalIP = &ipNet
	}
	if in.ExternalSK != "" {
		decoded, err := base64.StdEncoding.DecodeString(in.ExternalSK)
		if err != nil {
			return z, fmt.Errorf("invalid external SK: %w", err)
		}
		if len(decoded) != 32 {
			return z, fmt.Errorf("external SK wrong length: %d", len(decoded))
		}
		var key [32]byte
		copy(key[:], decoded)
		z.ExternalSK = wgtypes.Key(key)
	}

	parseNode := func(s string) (uint32, error) {
		v, err := strconv.ParseUint(s, 10, 32)
		return uint32(v), err
	}

	for idStr, val := range in.MyceliumKeys {
		id, err := parseNode(idStr)
		if err != nil {
			return z, err
		}
		decoded, err := base64.StdEncoding.DecodeString(val)
		if err != nil {
			return z, err
		}
		z.MyceliumKeys[id] = decoded
	}
	for idStr, val := range in.NodesIPRange {
		id, err := parseNode(idStr)
		if err != nil {
			return z, err
		}
		ipNet, err := zos.ParseIPNet(val)
		if err != nil {
			return z, err
		}
		z.NodesIPRange[id] = ipNet
	}
	for idStr, val := range in.NodeDeploymentID {
		id, err := parseNode(idStr)
		if err != nil {
			return z, err
		}
		z.NodeDeploymentID[id] = val
	}
	for idStr, val := range in.WGPort {
		id, err := parseNode(idStr)
		if err != nil {
			return z, err
		}
		z.WGPort[id] = val
	}
	for idStr, val := range in.Keys {
		id, err := parseNode(idStr)
		if err != nil {
			return z, err
		}
		decoded, err := base64.StdEncoding.DecodeString(val)
		if err != nil {
			return z, err
		}
		if len(decoded) != 32 {
			return z, fmt.Errorf("invalid key length for %d", id)
		}
		var key [32]byte
		copy(key[:], decoded)
		z.Keys[id] = wgtypes.Key(key)
	}

	return z, nil
}

// MarshalJSON implements custom JSON marshaling for Cluster
func (c Cluster) MarshalJSON() ([]byte, error) {
	// Create a serializable version of the cluster
	serializable := struct {
		Name        string   `json:"name"`
		Token       string   `json:"token"`
		Nodes       []Node   `json:"nodes"`
		ProjectName string   `json:"project_name,omitempty"`
		Network     zNetJSON `json:"network,omitempty"`
	}{
		Name:        c.Name,
		Token:       c.Token,
		Nodes:       c.Nodes,
		ProjectName: c.ProjectName,
		Network:     marshalZNet(c.Network),
	}

	return json.Marshal(serializable)
}

// UnmarshalJSON implements custom JSON unmarshaling for Cluster
func (c *Cluster) UnmarshalJSON(data []byte) error {
	// First unmarshal into a temporary structure
	var temp struct {
		Name        string   `json:"name"`
		Token       string   `json:"token"`
		Nodes       []Node   `json:"nodes"`
		ProjectName string   `json:"project_name,omitempty"`
		Network     zNetJSON `json:"network"`
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

	if err := json.Unmarshal(data, &temp); err != nil {
		return fmt.Errorf("failed to unmarshal cluster: %w", err)
	}

	c.Name = temp.Name
	c.Token = temp.Token
	c.Nodes = temp.Nodes
	c.ProjectName = temp.ProjectName

	var err error
	c.Network, err = unmarshalZNet(temp.Network)
	if err != nil {
		return fmt.Errorf("failed to unmarshal network: %w", err)
	}
	return nil
}

type VM struct {
	// Node Node `json:"node" validate:"required"`
	Name   string `json:"name" validate:"required,min=3,max=20,alphanum"`
	NodeID uint32 `json:"node_id" validate:"required"`

	CPU      uint8             `json:"cpu" validate:"required,min=1"`
	Memory   uint64            `json:"memory" validate:"required,min=2048"`     // Memory in MB
	RootSize uint64            `json:"root_size" validate:"required,min=5120"`  // Storage in MB
	DiskSize uint64            `json:"disk_size" validate:"required,min=10240"` // Storage in MB
	GPUIDs   []string          `json:"gpu_ids,omitempty"`                       // List of GPU IDs
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
	// Computed
	Network     workloads.ZNet `json:"network,omitempty"`
	ProjectName string         `json:"project_name,omitempty"`
}

// MarshalJSON implements custom JSON marshaling for VM
func (vm VM) MarshalJSON() ([]byte, error) {
	// Create a serializable version of the VM
	serializable := struct {
		Name         string            `json:"name"`
		NodeID       uint32            `json:"node_id"`
		CPU          uint8             `json:"cpu"`
		Memory       uint64            `json:"memory"`
		RootSize     uint64            `json:"root_size"`
		DiskSize     uint64            `json:"disk_size"`
		GPUIDs       []string          `json:"gpu_ids,omitempty"`
		EnvVars      map[string]string `json:"env_vars"`
		Flist        string            `json:"flist,omitempty"`
		Entrypoint   string            `json:"entrypoint,omitempty"`
		IP           string            `json:"ip,omitempty"`
		MyceliumIP   string            `json:"mycelium_ip,omitempty"`
		PlanetaryIP  string            `json:"planetary_ip,omitempty"`
		ContractID   uint64            `json:"contract_id,omitempty"`
		OriginalName string            `json:"original_name,omitempty"`
		ProjectName  string            `json:"project_name,omitempty"`
		Network      zNetJSON          `json:"network,omitempty"`
	}{
		ProjectName:  vm.ProjectName,
		Name:         vm.Name,
		NodeID:       vm.NodeID,
		CPU:          vm.CPU,
		Memory:       vm.Memory,
		RootSize:     vm.RootSize,
		DiskSize:     vm.DiskSize,
		GPUIDs:       vm.GPUIDs,
		EnvVars:      vm.EnvVars,
		Flist:        vm.Flist,
		Entrypoint:   vm.Entrypoint,
		IP:           vm.IP,
		MyceliumIP:   vm.MyceliumIP,
		PlanetaryIP:  vm.PlanetaryIP,
		ContractID:   vm.ContractID,
		OriginalName: vm.OriginalName,
		Network:      marshalZNet(vm.Network),
	}

	return json.Marshal(serializable)
}

// UnmarshalJSON implements custom JSON unmarshaling for VM
func (vm *VM) UnmarshalJSON(data []byte) error {
	// First unmarshal into a temporary structure
	var temp struct {
		ProjectName  string            `json:"project_name,omitempty"`
		Name         string            `json:"name"`
		NodeID       uint32            `json:"node_id"`
		CPU          uint8             `json:"cpu"`
		Memory       uint64            `json:"memory"`
		RootSize     uint64            `json:"root_size"`
		DiskSize     uint64            `json:"disk_size"`
		GPUIDs       []string          `json:"gpu_ids,omitempty"`
		EnvVars      map[string]string `json:"env_vars"`
		Flist        string            `json:"flist,omitempty"`
		Entrypoint   string            `json:"entrypoint,omitempty"`
		IP           string            `json:"ip,omitempty"`
		MyceliumIP   string            `json:"mycelium_ip,omitempty"`
		PlanetaryIP  string            `json:"planetary_ip,omitempty"`
		ContractID   uint64            `json:"contract_id,omitempty"`
		OriginalName string            `json:"original_name,omitempty"`
		Network      zNetJSON          `json:"network"`
	}
	if err := json.Unmarshal(data, &temp); err != nil {
		return fmt.Errorf("failed to unmarshal VM: %w", err)
	}

	vm.ProjectName = temp.ProjectName
	vm.Name = temp.Name
	vm.NodeID = temp.NodeID
	vm.CPU = temp.CPU
	vm.Memory = temp.Memory
	vm.RootSize = temp.RootSize
	vm.DiskSize = temp.DiskSize
	vm.GPUIDs = temp.GPUIDs
	vm.EnvVars = temp.EnvVars
	vm.Flist = temp.Flist
	vm.Entrypoint = temp.Entrypoint
	vm.IP = temp.IP
	vm.MyceliumIP = temp.MyceliumIP
	vm.PlanetaryIP = temp.PlanetaryIP
	vm.ContractID = temp.ContractID
	vm.OriginalName = temp.OriginalName

	var err error
	vm.Network, err = unmarshalZNet(temp.Network)
	if err != nil {
		return fmt.Errorf("failed to unmarshal network: %w", err)
	}
	return nil
}

// PrepareVM prepares the VM for deployment by setting project name and network name
func (v *VM) PrepareVM() error {
	if v.ProjectName == "" {
		return fmt.Errorf("VM project name is not set")
	}

	if v.Network.Name == "" {
		v.Network.Name = v.ProjectName + "net"
	}

	if v.OriginalName == "" {
		v.OriginalName = v.Name
		v.Name = v.ProjectName + v.OriginalName
	}

	return nil
}

func (v *VM) LoadFromDeployment(deployment workloads.Deployment) error {
	if len(deployment.Vms) == 0 {
		return fmt.Errorf("deployment has no VMs")
	}

	vm := deployment.Vms[0]
	v.IP = vm.IP
	v.MyceliumIP = vm.MyceliumIP
	v.PlanetaryIP = vm.PlanetaryIP
	v.ContractID = deployment.ContractID

	return nil
}

func (c *Cluster) Validate() error {
	nodeNames := make(map[string]struct{})
	for _, node := range c.Nodes {
		if _, exists := nodeNames[node.Name]; exists {
			return fmt.Errorf("duplicate node name found: %s", node.Name)
		}
		nodeNames[node.Name] = struct{}{}
	}

	return nil
}

func (c *Cluster) getAllClusterContracts() ([]uint64, error) {
	var contracts []uint64
	for _, contractId := range c.Network.NodeDeploymentID {
		contracts = append(contracts, contractId)
	}

	for _, contractId := range c.Nodes {
		if contractId.ContractID != 0 {
			contracts = append(contracts, contractId.ContractID)
		}
	}

	return contracts, nil
}
