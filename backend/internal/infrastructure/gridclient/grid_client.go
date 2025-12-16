package gridclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"kubecloud/internal/config"
	"kubecloud/internal/infrastructure/grid"
	"kubecloud/internal/infrastructure/logger"
	"math"
	"net/http"
	"time"

	"github.com/cosmos/go-bip39"
	substrate "github.com/threefoldtech/tfchain/clients/tfchain-client-go"
	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/calculator"
	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/deployer"
	client "github.com/threefoldtech/tfgrid-sdk-go/grid-client/node"
	"github.com/threefoldtech/tfgrid-sdk-go/grid-proxy/pkg/types"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

const (
	tftBaseUnits       = 1e7
	setupSleepDuration = 7 * time.Second
)

type GridClient interface {
	// chain methods
	FromTFTtoUSDMillicent(amount uint64) (uint64, error)
	FromUSDMillicentToTFT(amountMillicent uint64) (uint64, error)
	GetUserBalanceUSDMillicent(userMnemonic string) (uint64, error)
	GetUserBalanceUSD(userMnemonic string) (float64, error)
	TransferTFTsFromSystem(tftBalance uint64, userMnemonic string) error
	TransferTFTsToSystem(tftBalance uint64, userMnemonic string) error
	SystemIdentity() (substrate.Identity, error)
	GetTwinIDFromUserMnemonic(mnemonic string) (uint64, error)
	GetFreeBalanceTFT(mnemonic string) (uint64, error)
	GetUserAddress(mnemonic string) (string, error)
	AcceptTermsAndConditions(mnemonic, docLink, docHash string) error
	CreateTwin(mnemonic string) (uint32, error)
	CreateRentContract(mnemonic string, nodeID uint32) (uint64, error)
	CancelContract(mnemonic string, contractID uint64) error
	SetupUserOnTFChain(termsAndConditions config.TermsANDConditions) (mnemonic string, twinID uint32, err error)

	// node methods
	GetNodeClient(nodeID uint32) (*client.NodeClient, error)

	// calculator methods
	NewCalculator(mnemonic string) (calculator.Calculator, error)

	// grid-proxy client methods
	Node(ctx context.Context, nodeID uint32) (res types.NodeWithNestedCapacity, err error)
	Nodes(ctx context.Context, filter types.NodeFilter, pagination types.Limit) (res []types.Node, totalCount int, err error)
	Twins(ctx context.Context, filter types.TwinFilter, limit types.Limit) (res []types.Twin, totalCount int, err error)
	Stats(ctx context.Context, filter types.StatsFilter) (res types.Stats, err error)

	Close()
}

type gridClient struct {
	gridClient     *deployer.TFPluginClient
	systemMnemonic string
}

var _ GridClient = (*gridClient)(nil)

type ClientOpts func(*clientCfg)
type clientCfg struct {
	network       string
	traceProvider *sdktrace.TracerProvider
}

func WithNetwork(network string) ClientOpts {
	return func(c *clientCfg) {
		c.network = network
	}
}

func WithTracerProvider(tp *sdktrace.TracerProvider) ClientOpts {
	return func(c *clientCfg) {
		c.traceProvider = tp
	}
}

func NewGridClient(systemMnemonic string, debug bool, disableSentry bool, opts ...ClientOpts) (GridClient, error) {
	cfg := &clientCfg{}

	for _, o := range opts {
		o(cfg)
	}

	pluginOpts := []deployer.PluginOpt{
		deployer.WithDisableSentry(),
	}

	if debug {
		pluginOpts = append(pluginOpts, deployer.WithLogs())
	}
	if disableSentry {
		pluginOpts = append(pluginOpts, deployer.WithDisableSentry())
	}
	if cfg.network != "" {
		pluginOpts = append(pluginOpts, deployer.WithNetwork(cfg.network))
	}

	gridCl, err := deployer.NewTFPluginClient(
		systemMnemonic,
		pluginOpts...,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create TF grid client: %w", err)
	}

	return &gridClient{
		systemMnemonic: systemMnemonic,
		gridClient:     &gridCl,
	}, nil
}

// Node returns the node given its ID from grid-proxy
func (s *gridClient) Node(ctx context.Context, nodeID uint32) (res types.NodeWithNestedCapacity, err error) {
	return s.gridClient.GridProxyClient.Node(ctx, nodeID)
}

// Nodes returns the nodes from grid-proxy based on filter and limit
func (s *gridClient) Nodes(ctx context.Context, filter types.NodeFilter, limit types.Limit) (res []types.Node, totalCount int, err error) {
	return s.gridClient.GridProxyClient.Nodes(ctx, filter, limit)
}

// Twins returns the twins from grid-proxy based on filter and limit
func (s *gridClient) Twins(ctx context.Context, filter types.TwinFilter, limit types.Limit) (res []types.Twin, totalCount int, err error) {
	return s.gridClient.GridProxyClient.Twins(ctx, filter, limit)
}

// Stats returns the stats from grid-proxy based on the passed filter
func (s *gridClient) Stats(ctx context.Context, filter types.StatsFilter) (res types.Stats, err error) {
	return s.gridClient.GridProxyClient.Stats(ctx, filter)
}

// FromTFTtoUSDMillicent converts TFT amount to USD Millicent (1/1000 of a dollar)
func (s *gridClient) FromTFTtoUSDMillicent(amount uint64) (uint64, error) {
	price, err := s.gridClient.SubstrateConn.GetTFTPrice()
	if err != nil {
		return 0, err
	}

	usdMillicentBalance := uint64(math.Round((float64(amount) / tftBaseUnits) * float64(price)))
	return usdMillicentBalance, nil
}

// FromUSDMillicentToTFT converts USD Millicent to TFT amount
// This avoids floating point precision issues by accepting an integer value
func (s *gridClient) FromUSDMillicentToTFT(amountMillicent uint64) (uint64, error) {
	price, err := s.gridClient.SubstrateConn.GetTFTPrice()
	if err != nil {
		return 0, err
	}

	// Convert Millicent to dollars for the calculation
	amountUSD := FromUSDMilliCentToUSD(amountMillicent)
	tft := (amountUSD * tftBaseUnits) / (float64(price) / 1000)
	return uint64(tft), nil
}

func FromUSDMilliCentToUSD(amountMillicent uint64) float64 {
	return float64(amountMillicent) / 1000
}

func FromUSDToUSDMillicent(amountUSD float64) uint64 {
	return uint64(amountUSD * 1000)
}

// GetUserBalanceUSDMillicent gets balance of user in USD Millicent
// This avoids floating point precision issues by returning an integer value
func (s *gridClient) GetUserBalanceUSDMillicent(userMnemonic string) (uint64, error) {
	// Create identity of user from mnemonic
	userIdentity, err := s.getIdentity(userMnemonic)
	if err != nil {
		return 0, err
	}

	tftBalance, err := s.gridClient.SubstrateConn.GetBalance(userIdentity)
	if err != nil {
		return 0, err
	}

	return s.FromTFTtoUSDMillicent(tftBalance.Free.Uint64())
}

// GetUserBalanceUSD gets balance of user in USD
func (s *gridClient) GetUserBalanceUSD(userMnemonic string) (float64, error) {
	usdMillicentBalance, err := s.GetUserBalanceUSDMillicent(userMnemonic)
	if err != nil {
		return 0, err
	}

	return FromUSDMilliCentToUSD(usdMillicentBalance), nil
}

// TransferTFTsFromSystem transfer balance to users' account
func (s *gridClient) TransferTFTsFromSystem(tftBalance uint64, userMnemonic string) error {
	// Create identity of user from mnemonic
	userIdentity, err := s.getIdentity(userMnemonic)
	if err != nil {
		return err
	}

	systemIdentity, err := s.SystemIdentity()
	if err != nil {
		return err
	}

	return s.gridClient.SubstrateConn.Transfer(tftBalance, systemIdentity, userIdentity.PublicKey())
}

// TransferTFTsToSystem transfer balance to system account
func (s *gridClient) TransferTFTsToSystem(tftBalance uint64, userMnemonic string) error {
	// Create identity of user from mnemonic
	userIdentity, err := s.getIdentity(userMnemonic)
	if err != nil {
		return err
	}

	systemIdentity, err := s.SystemIdentity()
	if err != nil {
		return err
	}

	return s.gridClient.SubstrateConn.Transfer(tftBalance, userIdentity, systemIdentity.PublicKey())
}

// SystemIdentity creates identity of system from mnemonic
func (s *gridClient) SystemIdentity() (substrate.Identity, error) {
	return s.gridClient.SubstrateConn.NewIdentityFromSr25519Phrase(s.systemMnemonic)
}

// GetTwinIDFromMnemonic gets twinID from user mnemonic
func (s *gridClient) GetTwinIDFromUserMnemonic(mnemonic string) (uint64, error) {
	identity, err := s.getIdentity(mnemonic)
	if err != nil {
		return 0, err
	}

	twinID, err := s.gridClient.SubstrateConn.GetTwinByPubKey(identity.PublicKey())
	if err != nil {
		return 0, err
	}

	return uint64(twinID), nil
}

// GetNodeClient gents the node client given nodeID
func (s *gridClient) GetNodeClient(nodeID uint32) (*client.NodeClient, error) {
	return s.gridClient.NcPool.GetNodeClient(s.gridClient.SubstrateConn, nodeID)
}

// NewCalculator creates a new Calculator from user mnemonic
func (s *gridClient) NewCalculator(mnemonic string) (calculator.Calculator, error) {
	identity, err := s.getIdentity(mnemonic)
	if err != nil {
		return calculator.Calculator{}, err
	}

	return calculator.NewCalculator(s.gridClient.SubstrateConn, identity), nil
}

// GetFreeBalance returns free balance from user mnemonic
func (s *gridClient) GetFreeBalanceTFT(mnemonic string) (uint64, error) {
	identity, err := s.getIdentity(mnemonic)
	if err != nil {
		return 0, err
	}

	balance, err := s.gridClient.SubstrateConn.GetBalance(identity)
	if err != nil {
		return 0, err
	}
	return balance.Free.Uint64(), nil
}

// GetUserAddress gets user address from user mnemonic
func (s *gridClient) GetUserAddress(mnemonic string) (string, error) {
	identity, err := s.getIdentity(mnemonic)
	if err != nil {
		return "", err
	}

	return identity.Address(), nil
}

func (s *gridClient) getIdentity(mnemonic string) (substrate.Identity, error) {
	identity, err := s.gridClient.SubstrateConn.NewIdentityFromSr25519Phrase(mnemonic)
	if err != nil {
		return nil, fmt.Errorf("identity creation failed: %w", err)
	}
	return identity, nil
}

// AcceptTermsAndConditions accepts terms and conditions given user mnemonic
func (s *gridClient) AcceptTermsAndConditions(mnemonic, docLink, docHash string) error {
	identity, err := s.getIdentity(mnemonic)
	if err != nil {
		return err
	}

	return s.gridClient.SubstrateConn.AcceptTermsAndConditions(identity, docLink, docHash)
}

// CreateTwin creates a twin given user mnemonic
func (s *gridClient) CreateTwin(mnemonic string) (uint32, error) {
	identity, err := s.getIdentity(mnemonic)
	if err != nil {
		return 0, err
	}

	return s.gridClient.SubstrateConn.CreateTwin(identity, "", []byte{})
}

// CreateRentContract creates rent contract on a node given its id and user mnemonic
func (s *gridClient) CreateRentContract(mnemonic string, nodeID uint32) (uint64, error) {
	// Get Identity
	identity, err := s.getIdentity(mnemonic)
	if err != nil {
		return 0, err
	}

	// Reserve the node
	return s.gridClient.SubstrateConn.CreateRentContract(identity, nodeID, nil)
}

// CancelContract cancels a contract given its ID and user mnemonic
func (s *gridClient) CancelContract(mnemonic string, contractID uint64) error {
	// Get Identity
	identity, err := s.getIdentity(mnemonic)
	if err != nil {
		return err
	}

	return s.gridClient.SubstrateConn.CancelContract(identity, contractID)
}

// Activates user account with activation service
func activateAccount(substrateAccountID, network string) error {
	activationServiceURL := grid.ActivationServiceURLs[network]

	body := map[string]string{"substrateAccountID": substrateAccountID}
	jsonData, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal activation payload: %w", err)
	}

	resp, err := http.Post(activationServiceURL, "application/json", bytes.NewReader(jsonData))
	if err != nil {
		return fmt.Errorf("activation request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("activation failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}

// SetupUserOnTFChain performs all TFChain setup steps and returns mnemonic, identity, twin ID
func (s *gridClient) SetupUserOnTFChain(termsAndConditions config.TermsANDConditions) (mnemonic string, twinID uint32, err error) {
	mnemonic, err = GenerateMnemonic()
	if err != nil {
		return "", 0, fmt.Errorf("generate mnemonic failed: %w", err)
	}

	address, err := s.GetUserAddress(mnemonic)
	if err != nil {
		return "", 0, fmt.Errorf("failed to get user address: %w", err)
	}

	// Activate account with activation service
	if err := activateAccount(address, s.gridClient.Network); err != nil {
		return "", 0, fmt.Errorf("activation failed: %w", err)
	}

	// Wait a few seconds for account activation to complete
	time.Sleep(setupSleepDuration)

	// Accept terms and conditions
	err = s.AcceptTermsAndConditions(mnemonic, termsAndConditions.DocumentLink, termsAndConditions.DocumentHash)
	if err != nil {
		return "", 0, fmt.Errorf("failed to accept terms and conditions: %w", err)
	}

	// Create Twin
	twinID, err = s.CreateTwin(mnemonic)
	if err != nil {
		return "", 0, fmt.Errorf("failed to create twin: %w", err)
	}

	log := logger.ForOperation("chain_account", "create_twin")
	log.Debug().
		Uint32("twin_id", twinID).
		Str("address", address).
		Msg("Twin created successfully")
	return mnemonic, twinID, nil
}

func (s *gridClient) Close() {
	s.gridClient.Close()
}

// GenerateMnemonic generate mnemonic
func GenerateMnemonic() (string, error) {
	entropy, err := bip39.NewEntropy(128)
	if err != nil {
		return "", fmt.Errorf("failed to generate entropy: %w", err)
	}

	mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil {
		return "", fmt.Errorf("failed to generate mnemonic: %w", err)
	}

	if !bip39.IsMnemonicValid(mnemonic) {
		return "", fmt.Errorf("generated mnemonic is not valid")
	}

	return mnemonic, nil
}
