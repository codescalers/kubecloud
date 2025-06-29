package controller

import (
	"context"
	"fmt"

	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/deployer"
	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/workloads"
	"github.com/threefoldtech/tfgrid-sdk-go/grid-proxy/pkg/types"
	"github.com/threefoldtech/zosbase/pkg/gridtypes/zos"
	"k8s.io/klog"
)

type GWRequest struct {
	Hostname string   `json:"hostname"`
	Backends []string `json:"backends"`
}

func deployGateway(pluginClient deployer.TFPluginClient, gw GWRequest) (workloads.GatewayNameProxy, error) {
	var zosBackends []zos.Backend
	for _, backend := range gw.Backends {
		zosBackends = append(zosBackends, zos.Backend(backend))
	}

	gateway := workloads.GatewayNameProxy{
		Name:         gw.Hostname,
		Backends:     zosBackends,
		SolutionType: gw.Hostname,
	}

	node, err := selectNode(pluginClient)
	if err != nil {
		return workloads.GatewayNameProxy{}, fmt.Errorf("failed to select node: %w", err)
	}

	// gateway.NodeID = node
	gateway.NodeID = 14 // TODO: remove this after testing
	if err := pluginClient.GatewayNameDeployer.Deploy(context.TODO(), &gateway); err != nil {
		return workloads.GatewayNameProxy{}, fmt.Errorf("failed to deploy gateway on node %d: %w", node, err)
	}

	res, err := pluginClient.State.LoadGatewayNameFromGrid(context.TODO(), gateway.NodeID, gateway.Name, gateway.Name)
	if err != nil {
		return workloads.GatewayNameProxy{}, fmt.Errorf("failed to load gateway for name %s: %w", gateway.Name, err)
	}

	klog.Infof("Gateway deployed successfully on node %d", node)

	return res, nil
}

func selectNode(pluginClient deployer.TFPluginClient) (uint32, error) {
	trueVal := true
	nodes, err := deployer.FilterNodes(
		context.Background(),
		pluginClient,
		types.NodeFilter{Domain: &trueVal, Status: []string{"up"}},
		nil, nil, nil,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to filter nodes: %w", err)
	}

	if len(nodes) == 0 {
		return 0, fmt.Errorf("no available nodes found")
	}

	return uint32(nodes[0].NodeID), nil
}
