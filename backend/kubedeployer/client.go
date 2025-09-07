package kubedeployer

import (
	"fmt"

	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/deployer"
)

type Client struct {
	GridClient deployer.TFPluginClient
}

func NewClient(mnemonic, gridNet string, debug bool) (*Client, error) {
	plugingOpts := []deployer.PluginOpt{
		deployer.WithNetwork(gridNet),
	}
	if debug {
		plugingOpts = append(plugingOpts, deployer.WithLogs())
	}

	tfplugin, err := deployer.NewTFPluginClient(
		mnemonic,
		plugingOpts...,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create TFPluginClient: %v", err)
	}

	return &Client{
		GridClient: tfplugin,
	}, nil
}

func (c *Client) Close() {
	c.GridClient.Close()
}
