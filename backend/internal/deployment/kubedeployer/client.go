package kubedeployer

import (
	"fmt"

	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/deployer"
	"go.opentelemetry.io/otel/trace"
)

type Client struct {
	GridClient    deployer.TFPluginClient
	mnemonic      string
	zlogOutputURL string
}

func NewClient(mnemonic, gridNet string, debug bool, tp trace.TracerProvider, zlogOutputURL string) (*Client, error) {
	pluginOpts := []deployer.PluginOpt{
		deployer.WithNetwork(gridNet),
		deployer.WithDisableSentry(),
	}
	if debug {
		pluginOpts = append(pluginOpts, deployer.WithLogs())
	}

	if tp != nil {
		pluginOpts = append(pluginOpts, deployer.WithTraceProvider(tp))
	}

	tfplugin, err := deployer.NewTFPluginClient(
		mnemonic,
		pluginOpts...,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create TFPluginClient: %v", err)
	}

	return &Client{
		GridClient:    tfplugin,
		mnemonic:      mnemonic,
		zlogOutputURL: zlogOutputURL,
	}, nil
}

func (c *Client) Close() {
	c.GridClient.Close()
}
