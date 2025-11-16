package kubedeployer

import (
	"context"
	"fmt"
	"net"

	"github.com/pkg/errors"
	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/deployer"
	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/workloads"
	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/zos"
)

func getIpForVm(ctx context.Context, tfPluginClient deployer.TFPluginClient, networkName string, nodeID uint32) (string, error) {
	network := tfPluginClient.State.Networks.GetNetwork(networkName)
	ipRange := network.GetNodeSubnet(nodeID)

	ip, ipRangeCIDR, err := net.ParseCIDR(ipRange)
	if err != nil {
		return "", errors.Wrapf(err, "invalid IP range %s for node %d", ipRange, nodeID)
	}

	usedHostIDs, err := getUsedHostIDsFromGrid(ctx, tfPluginClient, nodeID, networkName, ipRangeCIDR)
	if err != nil {
		return "", errors.Wrapf(err, "failed to get used IPs for node %d", nodeID)
	}

	// skip 0, 1, and 255 as they are reserved
	for hostID := byte(2); hostID < 255; hostID++ {
		used := false
		for _, usedID := range usedHostIDs {
			if usedID == hostID {
				used = true
				break
			}
		}
		if !used {
			vmIP := make(net.IP, len(ip.To4()))
			copy(vmIP, ip.To4())
			vmIP[3] = hostID

			return vmIP.String(), nil
		}
	}

	return "", fmt.Errorf("all IPs are exhausted for network %s on node %d", networkName, nodeID)
}

func getUsedHostIDsFromGrid(ctx context.Context, tfPluginClient deployer.TFPluginClient, nodeID uint32, networkName string, ipRangeCIDR *net.IPNet) ([]byte, error) {
	nodeClient, err := tfPluginClient.NcPool.GetNodeClient(tfPluginClient.SubstrateConn, nodeID)
	if err != nil {
		return nil, errors.Wrapf(err, "could not get node client for node %d", nodeID)
	}

	privateIPs, err := nodeClient.NetworkListPrivateIPs(ctx, networkName)
	if err != nil {
		return nil, errors.Wrapf(err, "could not list private IPs from node %d", nodeID)
	}

	var usedHostIDs []byte
	for _, privateIP := range privateIPs {
		parsedIP := net.ParseIP(privateIP).To4()
		if parsedIP != nil && ipRangeCIDR.Contains(parsedIP) {
			usedHostIDs = append(usedHostIDs, parsedIP[3])
		}
	}

	return usedHostIDs, nil
}

func createNetworkWorkload(networkName, projectName string, nodes []uint32) (workloads.ZNet, error) {
	keys := make(map[uint32][]byte)
	for _, node := range nodes {
		key, err := workloads.RandomMyceliumKey()
		if err != nil {
			return workloads.ZNet{}, err
		}
		keys[node] = key
	}

	return workloads.ZNet{
		Name:  networkName,
		Nodes: nodes,
		IPRange: zos.IPNet{IPNet: net.IPNet{
			IP:   net.IPv4(10, 20, 0, 0),
			Mask: net.CIDRMask(16, 32),
		}},
		MyceliumKeys: keys,
		SolutionType: projectName,
	}, nil
}
