package substrate

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"kubecloud/internal/config"
	"kubecloud/internal/infrastructure/grid"
	"kubecloud/internal/infrastructure/logger"
	"math"
	"net/http"
	"time"

	substrate "github.com/threefoldtech/tfchain/clients/tfchain-client-go"
	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/calculator"
	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/deployer"
	client "github.com/threefoldtech/tfgrid-sdk-go/grid-client/node"
	proxy "github.com/threefoldtech/tfgrid-sdk-go/grid-proxy/pkg/client"
)

type substrateClient struct {
	gridClient     *deployer.TFPluginClient
	systemMnemonic string
}

var _ Substrate = (*substrateClient)(nil)

func NewSubstrateClient(systemMnemonic string, network string, debug bool) (Substrate, error) {
	pluginOpts := []deployer.PluginOpt{
		deployer.WithNetwork(network),
		deployer.WithDisableSentry(),
	}
	if debug {
		pluginOpts = append(pluginOpts, deployer.WithLogs())
	}

	gridClient, err := deployer.NewTFPluginClient(
		systemMnemonic,
		pluginOpts...,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create TF grid client: %w", err)
	}

	return &substrateClient{
		systemMnemonic: systemMnemonic,
		gridClient:     &gridClient,
	}, nil
}

// GridProxyClient returns the grid proxy client from the grid client
func (s *substrateClient) GridProxyClient() proxy.Client {
	return s.gridClient.GridProxyClient
}

// FromTFTtoUSDMillicent converts TFT amount to USD Millicent (1/1000 of a dollar)
func (s *substrateClient) FromTFTtoUSDMillicent(amount uint64) (uint64, error) {
	price, err := s.gridClient.SubstrateConn.GetTFTPrice()
	if err != nil {
		return 0, err
	}

	usdMillicentBalance := uint64(math.Round((float64(amount) / 1e7) * float64(price)))
	return usdMillicentBalance, nil
}

// FromUSDMillicentToTFT converts USD Millicent to TFT amount
// This avoids floating point precision issues by accepting an integer value
func (s *substrateClient) FromUSDMillicentToTFT(amountMillicent uint64) (uint64, error) {
	price, err := s.gridClient.SubstrateConn.GetTFTPrice()
	if err != nil {
		return 0, err
	}

	// Convert Millicent to dollars for the calculation
	amountUSD := FromUSDMilliCentToUSD(amountMillicent)
	tft := (amountUSD * 1e7) / (float64(price) / 1000)
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
func (s *substrateClient) GetUserBalanceUSDMillicent(userMnemonic string) (uint64, error) {
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
func (s *substrateClient) GetUserBalanceUSD(userMnemonic string) (float64, error) {
	usdMillicentBalance, err := s.GetUserBalanceUSDMillicent(userMnemonic)
	if err != nil {
		return 0, err
	}

	return FromUSDMilliCentToUSD(usdMillicentBalance), nil
}

// TransferTFTsFromSystem transfer balance to users' account
func (s *substrateClient) TransferTFTsFromSystem(tftBalance uint64, userMnemonic string) error {
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
func (s *substrateClient) TransferTFTsToSystem(tftBalance uint64, userMnemonic string) error {
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
func (s *substrateClient) SystemIdentity() (substrate.Identity, error) {
	return s.gridClient.SubstrateConn.NewIdentityFromSr25519Phrase(s.systemMnemonic)
}

// GetTwinIDFromMnemonic gets twinID from user mnemonic
func (s *substrateClient) GetTwinIDFromUserMnemonic(mnemonic string) (uint64, error) {
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
func (s *substrateClient) GetNodeClient(nodeID uint32) (*client.NodeClient, error) {
	return s.gridClient.NcPool.GetNodeClient(s.gridClient.SubstrateConn, nodeID)
}

// NewCalculator creates a new Calculator from user mnemonic
func (s *substrateClient) NewCalculator(mnemonic string) (calculator.Calculator, error) {
	identity, err := s.getIdentity(mnemonic)
	if err != nil {
		return calculator.Calculator{}, err
	}

	calculatorClient := calculator.NewCalculator(s.gridClient.SubstrateConn, identity)

	return calculatorClient, nil
}

// GetFreeBalance returns free balance from user mnemonic
func (s *substrateClient) GetFreeBalanceTFT(mnemonic string) (uint64, error) {
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
func (s *substrateClient) GetUserAddress(mnemonic string) (string, error) {
	identity, err := s.getIdentity(mnemonic)
	if err != nil {
		return "", err
	}

	return identity.Address(), nil
}

func (s *substrateClient) getIdentity(mnemonic string) (substrate.Identity, error) {
	identity, err := s.gridClient.SubstrateConn.NewIdentityFromSr25519Phrase(mnemonic)
	if err != nil {
		return nil, fmt.Errorf("identity creation failed: %w", err)
	}
	return identity, nil
}

// AcceptTermsAndConditions accepts terms and conditions given user mnemonic
func (s *substrateClient) AcceptTermsAndConditions(mnemonic, docLink, docHash string) error {
	identity, err := s.getIdentity(mnemonic)
	if err != nil {
		return err
	}

	err = s.gridClient.SubstrateConn.AcceptTermsAndConditions(identity, docLink, docHash)
	if err != nil {
		return err
	}

	return nil
}

// CreateTwin creates a twin given user mnemonic
func (s *substrateClient) CreateTwin(mnemonic string) (uint32, error) {
	identity, err := s.getIdentity(mnemonic)
	if err != nil {
		return 0, err
	}

	twinID, err := s.gridClient.SubstrateConn.CreateTwin(identity, "", []byte{})
	if err != nil {
		return 0, err
	}

	return twinID, nil
}

// CreateRentContract creates rent contract on a node given its id and user mnemonic
func (s *substrateClient) CreateRentContract(mnemonic string, nodeID uint32) (uint64, error) {
	// Get Identity
	identity, err := s.getIdentity(mnemonic)
	if err != nil {
		return 0, err
	}

	// Reserve the node
	contractID, err := s.gridClient.SubstrateConn.CreateRentContract(identity, nodeID, nil)
	if err != nil {
		return 0, err
	}
	return contractID, nil
}

// CancelContract cancels a contract given its ID and user mnemonic
func (s *substrateClient) CancelContract(mnemonic string, contractID uint64) error {
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
func (s *substrateClient) SetupUserOnTFChain(termsAndConditions config.TermsANDConditions, network string) (mnemonic string, twinID uint32, err error) {
	mnemonic, err = GenerateMnemonic()
	if err != nil {
		return "", 0, fmt.Errorf("generate mnemonic failed: %w", err)
	}

	address, err := s.GetUserAddress(mnemonic)
	if err != nil {
		return "", 0, fmt.Errorf("failed to get user address: %w", err)
	}

	// Activate account with activation service
	if err := activateAccount(address, network); err != nil {
		return "", 0, fmt.Errorf("activation failed: %w", err)
	}

	// Wait a few seconds for account activation to complete
	time.Sleep(7 * time.Second)

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

func (s *substrateClient) Close() {
	s.gridClient.Close()
}
