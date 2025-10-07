package kubedeployer

import (
	"fmt"

	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/deployer"
)

type Client struct {
	GridClient deployer.TFPluginClient
	mnemonic   string
}

func NewClient(mnemonic, gridNet string, debug bool) (*Client, error) {
	pluginOpts := []deployer.PluginOpt{
		deployer.WithNetwork(gridNet),
		deployer.WithDisableSentry(),
	}
	if debug {
		pluginOpts = append(pluginOpts, deployer.WithLogs())
	}

	tfplugin, err := deployer.NewTFPluginClient(
		mnemonic,
		pluginOpts...,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create TFPluginClient: %v", err)
	}

	return &Client{
		GridClient: tfplugin,
		mnemonic:   mnemonic,
	}, nil
}

func (c *Client) Close() {
	c.GridClient.Close()
}
