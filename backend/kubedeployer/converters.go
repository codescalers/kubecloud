package kubedeployer

import (
	"fmt"
	"math/rand"

	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/workloads"
	zosTypes "github.com/threefoldtech/tfgrid-sdk-go/grid-client/zos"
)

const (
	K3S_FLIST      = "https://hub.threefold.me/hanafy.3bot/ahmedhanafy725-k3s-pure.flist"
	K3S_ENTRYPOINT = "/sbin/zinit init"
	K3S_DATA_DIR   = "/mnt/data"
	K3S_IFACE      = "flannel-br"
)

func generateRandomString(length int) string {
	chars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := 0; i < length; i++ {
		result[i] = chars[rand.Intn(len(chars))]
	}
	return string(result)
}

func deploymentFromNode(
	node Node,
	projectName string,
	networkName string,
	leaderIP string,
	token string,
	masterSSH string,
) (workloads.Deployment, error) {
	ipSeed, err := workloads.RandomMyceliumIPSeed()
	if err != nil {
		return workloads.Deployment{}, err
	}

	disk := workloads.Disk{
		Name:   fmt.Sprintf("%s_data", node.Name),
		SizeGB: node.DiskSize / 1024,
	}

	var gpus []zosTypes.GPU
	for _, gpuID := range node.GPUIDs {
		gpus = append(gpus, zosTypes.GPU(gpuID))
	}

	vm := workloads.VM{
		Name:           node.Name,
		NodeID:         node.NodeID,
		CPU:            node.CPU,
		MemoryMB:       node.Memory,
		RootfsSizeMB:   node.RootSize,
		EnvVars:        node.EnvVars,
		Flist:          node.Flist,
		Entrypoint:     node.Entrypoint,
		NetworkName:    networkName,
		IP:             node.IP,
		MyceliumIPSeed: ipSeed,
		Mounts: []workloads.Mount{
			{
				Name:       disk.Name,
				MountPoint: K3S_DATA_DIR,
			},
		},
		GPUs: gpus,
	}

	vm.EnvVars["K3S_NODE_NAME"] = node.Name
	vm.EnvVars["DUAL_STACK"] = "true"
	vm.EnvVars["MASTER"] = "false"
	vm.EnvVars["HA"] = "false"
	vm.EnvVars["K3S_URL"] = ""
	vm.EnvVars["K3S_TOKEN"] = token

	if node.Type == NodeTypeMaster || node.Type == NodeTypeLeader {
		vm.EnvVars["MASTER"] = "true"
		vm.EnvVars["HA"] = "true"
	}
	if node.Type != NodeTypeLeader {
		vm.EnvVars["K3S_URL"] = fmt.Sprintf("https://%s:6443", leaderIP)
	}
	if vm.EnvVars["K3S_FLANNEL_IFACE"] == "" {
		vm.EnvVars["K3S_FLANNEL_IFACE"] = K3S_IFACE
	}
	if vm.EnvVars["K3S_DATA_DIR"] == "" {
		vm.EnvVars["K3S_DATA_DIR"] = K3S_DATA_DIR
	}
	if vm.Flist == "" {
		vm.Flist = K3S_FLIST
	}
	if vm.Entrypoint == "" {
		vm.Entrypoint = K3S_ENTRYPOINT
	}

	vm.EnvVars["SSH_KEY"] = node.EnvVars["SSH_KEY"] + "\n" + masterSSH

	depl := workloads.NewDeployment(
		node.Name,
		node.NodeID,
		projectName, nil,
		networkName,
		[]workloads.Disk{disk}, nil,
		[]workloads.VM{vm}, nil, nil, nil,
	)

	return depl, nil
}

func nodeFromDeployment(
	depl workloads.Deployment,
) (Node, error) {
	vm := depl.Vms[0]
	var node Node

	diskSizeMb := uint64(0)
	if len(depl.Disks) > 0 {
		diskSizeMb = depl.Disks[0].SizeGB * 1024
	}

	node.Name = vm.Name
	node.NodeID = vm.NodeID
	node.CPU = vm.CPU
	node.Memory = vm.MemoryMB
	node.RootSize = vm.RootfsSizeMB
	node.DiskSize = diskSizeMb
	node.EnvVars = vm.EnvVars
	node.Flist = vm.Flist
	node.Entrypoint = vm.Entrypoint
	node.DiskSize = depl.Disks[0].SizeGB * 1024
	node.GPUIDs = make([]string, len(vm.GPUs))

	for i, gpu := range vm.GPUs {
		node.GPUIDs[i] = string(gpu)
	}

	// computed fields
	node.IP = vm.IP
	node.MyceliumIP = vm.MyceliumIP
	node.PlanetaryIP = vm.PlanetaryIP
	node.ContractID = depl.ContractID

	return node, nil
}

func GetProjectName(userID, clusterName string) string {
	return "kc" + userID + clusterName
}

func GetNodeName(userID, clusterName, nodeName string) string {
	return GetProjectName(userID, clusterName) + nodeName
}

func (c *Cluster) PrepareCluster(userID string) error {
	projectName := GetProjectName(userID, c.Name)
	networkName := projectName + "net"

	c.ProjectName = projectName
	c.Network.Name = networkName

	hasLeader := false
	for idx, node := range c.Nodes {
		c.Nodes[idx].OriginalName = node.Name // Safe, cause it checks if projectName is not empty
		c.Nodes[idx].Name = projectName + node.Name
		if node.Type == NodeTypeLeader {
			hasLeader = true
		}
	}

	if !hasLeader {
		for i, node := range c.Nodes {
			if node.Type == NodeTypeMaster {
				c.Nodes[i].Type = NodeTypeLeader
				break
			}
		}
	}

	return nil
}
