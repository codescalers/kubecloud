package substrate

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"time"

	substrate "github.com/threefoldtech/tfchain/clients/tfchain-client-go"
	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/deployer"

	"kubecloud/internal/infrastructure/logger"
	"kubecloud/internal/shared"
)

type TFChainClient struct {
	subConn                        *substrate.Substrate
	systemIdentity                 substrate.Identity
	network                        string
	termsANDConditionsDocumentLink string
	termsANDConditionsDocumentHash string
}

func NewTFChainClient(network, systemMnemonic, termsANDConditionsDocumentLink, termsANDConditionsDocumentHash string) (*TFChainClient, error) {
	manager := substrate.NewManager(deployer.SubstrateURLs[network]...)
	substrateConn, err := manager.Substrate()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to tf chain: %w", err)
	}

	systemIdentity, err := substrate.NewIdentityFromSr25519Phrase(systemMnemonic)
	if err != nil {
		return nil, fmt.Errorf("failed to create system identity: %w", err)
	}

	return &TFChainClient{
		subConn:        substrateConn,
		systemIdentity: systemIdentity,
		network:        network,

		termsANDConditionsDocumentLink: termsANDConditionsDocumentLink,
		termsANDConditionsDocumentHash: termsANDConditionsDocumentHash,
	}, nil
}

func (c *TFChainClient) NewIdentityFromSr25519Phrase(mnemonic string) (substrate.Identity, error) {
	return substrate.NewIdentityFromSr25519Phrase(mnemonic)
}

func (c *TFChainClient) SystemIdentity() substrate.Identity {
	return c.systemIdentity
}

func (c *TFChainClient) GetTwinByPubKey(pubKey []byte) (uint32, error) {
	return c.subConn.GetTwinByPubKey(pubKey)
}

func (c *TFChainClient) CreateRentContract(mnemonic string, nodeID uint32, solutionProviderID *uint64) (uint64, error) {
	identity, err := c.NewIdentityFromSr25519Phrase(mnemonic)
	if err != nil {
		return 0, err
	}

	return c.subConn.CreateRentContract(identity, nodeID, solutionProviderID)
}

func (c *TFChainClient) CancelContract(mnemonic string, contractID uint64) error {
	identity, err := c.NewIdentityFromSr25519Phrase(mnemonic)
	if err != nil {
		return err
	}
	return c.subConn.CancelContract(identity, contractID)
}

// SetupUserOnTFChain performs all TFChain setup steps and returns mnemonic, identity, twin ID
func (c *TFChainClient) SetupUserOnTFChain() (mnemonic string, twinID uint32, err error) {
	mnemonic, err = GenerateMnemonic()
	if err != nil {
		return "", 0, fmt.Errorf("generate mnemonic failed: %w", err)
	}

	identity, err := substrate.NewIdentityFromSr25519Phrase(mnemonic)
	if err != nil {
		return "", 0, fmt.Errorf("identity creation failed: %w", err)
	}

	// Activate account with activation service
	if err := c.activateAccount(identity.Address()); err != nil {
		return "", 0, fmt.Errorf("activation failed: %w", err)
	}

	// Wait a few seconds for account activation to complete
	time.Sleep(7 * time.Second)

	if err := c.subConn.AcceptTermsAndConditions(identity, c.termsANDConditionsDocumentLink, c.termsANDConditionsDocumentHash); err != nil {
		return "", 0, fmt.Errorf("accept terms failed: %w", err)
	}

	// Create Twin
	twinID, err = c.subConn.CreateTwin(identity, "", []byte{})
	if err != nil {
		return "", 0, fmt.Errorf("create twin failed: %w", err)
	}

	log := logger.ForOperation("chain_account", "create_twin")
	log.Debug().
		Uint32("twin_id", twinID).
		Str("address", identity.Address()).
		Msg("Twin created successfully")
	return mnemonic, twinID, nil
}

// TransferTFTsFromSystem transfer balance to users' account
func (c *TFChainClient) TransferTFTsFromSystem(tftBalance uint64, userMnemonic string) error {
	// Create identity of user from mnemonic
	userIdentity, err := substrate.NewIdentityFromSr25519Phrase(userMnemonic)
	if err != nil {
		return err
	}

	return c.subConn.Transfer(c.systemIdentity, tftBalance, substrate.AccountID(userIdentity.PublicKey()))
}

// TransferTFTsToSystem transfer balance to system account
func (c *TFChainClient) TransferTFTsToSystem(tftBalance uint64, userMnemonic string) error {
	// Create identity of user from mnemonic
	userIdentity, err := substrate.NewIdentityFromSr25519Phrase(userMnemonic)
	if err != nil {
		return err
	}

	return c.subConn.Transfer(userIdentity, tftBalance, substrate.AccountID(c.systemIdentity.PublicKey()))
}

// GetUserBalanceUSDMillicent gets balance of user in USD Millicent
// This avoids floating point precision issues by returning an integer value
func (c *TFChainClient) GetUserBalanceUSDMillicent(userMnemonic string) (uint64, error) {
	tftBalance, err := c.GetUserTFTBalance(userMnemonic)
	if err != nil {
		return 0, err
	}

	return c.FromTFTtoUSDMillicent(tftBalance)
}

// GetUserBalanceUSD gets balance of user in USD
func (c *TFChainClient) GetUserBalanceUSD(userMnemonic string) (float64, error) {
	usdMillicentBalance, err := c.GetUserBalanceUSDMillicent(userMnemonic)
	if err != nil {
		return 0, err
	}

	return FromUSDMilliCentToUSD(usdMillicentBalance), nil
}

// GetUserBalanceUSD gets balance of user in TFT
func (c *TFChainClient) GetUserTFTBalance(userMnemonic string) (uint64, error) {
	// Create identity of user from mnemonic
	userIdentity, err := substrate.NewIdentityFromSr25519Phrase(userMnemonic)
	if err != nil {
		return 0, err
	}

	account, err := substrate.FromAddress(userIdentity.Address())
	if err != nil {
		return 0, err
	}

	// get balance in TFT
	tftBalance, err := c.subConn.GetBalance(account)
	if err != nil {
		return 0, err
	}

	return tftBalance.Free.Uint64(), nil
}

// FromTFTtoUSDMillicent converts TFT amount to USD Millicent (1/1000 of a dollar)
func (c *TFChainClient) FromTFTtoUSDMillicent(amount uint64) (uint64, error) {
	price, err := c.getTFTPrice()
	if err != nil {
		return 0, err
	}

	usdMillicentBalance := uint64(math.Round((float64(amount) / 1e7) * price))
	return usdMillicentBalance, nil
}

// FromUSDMillicentToTFT converts USD Millicent to TFT amount
// This avoids floating point precision issues by accepting an integer value
func (c *TFChainClient) FromUSDMillicentToTFT(amountMillicent uint64) (uint64, error) {
	price, err := c.getTFTPrice()
	if err != nil {
		return 0, err
	}

	// Convert Millicent to dollars for the calculation
	amountUSD := FromUSDMilliCentToUSD(amountMillicent)
	tft := (amountUSD * 1e7) / (price / 1000)
	return uint64(tft), nil
}

func FromUSDMilliCentToUSD(amountMillicent uint64) float64 {
	return float64(amountMillicent) / 1000
}

func FromUSDToUSDMillicent(amountUSD float64) uint64 {
	return uint64(amountUSD * 1000)
}

// Activates user account with activation service
func (c *TFChainClient) activateAccount(substrateAccountID string) error {
	activationServiceURL := shared.ActivationServiceURLs[c.network]

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

func (c *TFChainClient) getTFTPrice() (float64, error) {
	price, err := c.subConn.GetTFTPrice()
	if err != nil {
		return 0, err
	}
	return float64(price), nil
}

func (c *TFChainClient) Close() {
	c.subConn.Close()
}
