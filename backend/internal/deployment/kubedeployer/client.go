package kubedeployer

import (
	"kubecloud/internal/infrastructure/gridclient"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

type Client struct {
	GridClient gridclient.GridClient
}

// NewClient creates a new Client instance
func NewClient(mnemonic, gridNet string, debug bool, tp *sdktrace.TracerProvider) (*Client, error) {
	var opts []gridclient.ClientOpts
	if gridNet != "" {
		opts = append(opts, gridclient.WithNetwork(gridNet))
	}
	if debug {
		opts = append(opts, gridclient.WithDebug())
	}
	if tp != nil {
		opts = append(opts, gridclient.WithTracerProvider(tp))
	}
	opts = append(opts, gridclient.WithDisableSentry())

	gridCl, err := gridclient.NewGridClient(mnemonic, opts...)
	if err != nil {
		return nil, err
	}

	return &Client{
		GridClient: gridCl,
	}, nil
}

// Close closes the underlying GridClient
func (c *Client) Close() {
	c.GridClient.Close()
}
