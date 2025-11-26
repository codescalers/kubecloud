package grid

import (
	"fmt"
	"math"

	substrate "github.com/threefoldtech/tfchain/clients/tfchain-client-go"
	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/calculator"
	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/deployer"
	client "github.com/threefoldtech/tfgrid-sdk-go/grid-client/node"
)

type SubstrateClient struct {
	gridClient     deployer.TFPluginClient
	systemMnemonic string
}

func NewSubstrateClient(systemMnemonic string, gridClient deployer.TFPluginClient) SubstrateClient {
	return SubstrateClient{
		systemMnemonic: systemMnemonic,
		gridClient:     gridClient,
	}
}

// FromTFTtoUSDMillicent converts TFT amount to USD Millicent (1/1000 of a dollar)
func (s *SubstrateClient) FromTFTtoUSDMillicent(amount uint64) (uint64, error) {
	price, err := s.gridClient.SubstrateConn.GetTFTPrice()
	if err != nil {
		return 0, err
	}

	usdMillicentBalance := uint64(math.Round((float64(amount) / 1e7) * float64(price)))
	return usdMillicentBalance, nil
}

// FromUSDMillicentToTFT converts USD Millicent to TFT amount
// This avoids floating point precision issues by accepting an integer value
func (s *SubstrateClient) FromUSDMillicentToTFT(amountMillicent uint64) (uint64, error) {
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
func (s *SubstrateClient) GetUserBalanceUSDMillicent(userMnemonic string) (uint64, error) {
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
func (s *SubstrateClient) GetUserBalanceUSD(userMnemonic string) (float64, error) {
	usdMillicentBalance, err := s.GetUserBalanceUSDMillicent(userMnemonic)
	if err != nil {
		return 0, err
	}

	return FromUSDMilliCentToUSD(usdMillicentBalance), nil
}

// TransferTFTsFromSystem transfer balance to users' account
func (s *SubstrateClient) TransferTFTsFromSystem(tftBalance uint64, userMnemonic string) error {
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
func (s *SubstrateClient) TransferTFTsToSystem(tftBalance uint64, userMnemonic string) error {
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
func (s *SubstrateClient) SystemIdentity() (substrate.Identity, error) {
	return s.gridClient.SubstrateConn.NewIdentityFromSr25519Phrase(s.systemMnemonic)
}

// GetTwinIDFromMnemonic gets twinID from user mnemonic
func (s *SubstrateClient) GetTwinIDFromUserMnemonic(mnemonic string) (uint64, error) {
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
func (s *SubstrateClient) GetNodeClient(nodeID uint32) (*client.NodeClient, error) {
	return s.gridClient.NcPool.GetNodeClient(s.gridClient.SubstrateConn, nodeID)
}

// NewCalculator creates a new Calculator from user mnemonic
func (s *SubstrateClient) NewCalculator(mnemonic string) (calculator.Calculator, error) {
	identity, err := s.getIdentity(mnemonic)
	if err != nil {
		return calculator.Calculator{}, err
	}

	calculatorClient := calculator.NewCalculator(s.gridClient.SubstrateConn, identity)

	return calculatorClient, nil
}

// GetFreeBalance returns free balance from user mnemonic
func (s *SubstrateClient) GetFreeBalance(mnemonic string) (uint64, error) {
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
func (s *SubstrateClient) GetUserAddress(mnemonic string) (string, error) {
	identity, err := s.getIdentity(mnemonic)
	if err != nil {
		return "", err
	}

	return identity.Address(), nil
}

func (s *SubstrateClient) getIdentity(mnemonic string) (substrate.Identity, error) {
	identity, err := s.gridClient.SubstrateConn.NewIdentityFromSr25519Phrase(mnemonic)
	if err != nil {
		return nil, fmt.Errorf("identity creation failed: %w", err)
	}
	return identity, nil
}

// AcceptTermsAndConditions accepts terms and conditions given user mnemonic
func (s *SubstrateClient) AcceptTermsAndConditions(mnemonic, docLink, docHash string) error {
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
func (s *SubstrateClient) CreateTwin(mnemonic string) (uint32, error) {
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
func (s *SubstrateClient) CreateRentContract(mnemonic string, nodeID uint32) (uint64, error) {
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
func (s *SubstrateClient) CancelContract(mnemonic string, contractID uint64) error {
	// Get Identity
	identity, err := s.getIdentity(mnemonic)
	if err != nil {
		return err
	}

	return s.gridClient.SubstrateConn.CancelContract(identity, contractID)
}
