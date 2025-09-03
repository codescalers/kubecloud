package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"time"

	substrate "github.com/threefoldtech/tfchain/clients/tfchain-client-go"

	"github.com/tyler-smith/go-bip39"
	"kubecloud/internal/logger"
)

// SetupUserOnTFChain performs all TFChain setup steps and returns mnemonic, identity, twin ID
func SetupUserOnTFChain(client *substrate.Substrate, config Configuration) (mnemonic string, twinID uint32, err error) {
	mnemonic, err = GenerateMnemonic()
	if err != nil {
		return "", 0, fmt.Errorf("generate mnemonic failed: %w", err)
	}

	identity, err := substrate.NewIdentityFromSr25519Phrase(mnemonic)
	if err != nil {
		return "", 0, fmt.Errorf("identity creation failed: %w", err)
	}

	// Activate account with activation service
	if err := ActivateAccount(identity.Address(), config.ActivationServiceURL); err != nil {
		return "", 0, fmt.Errorf("activation failed: %w", err)
	}

	// Wait a few seconds for account activation to complete
	time.Sleep(7 * time.Second)

	if err := client.AcceptTermsAndConditions(identity, config.TermsANDConditions.DocumentLink, config.TermsANDConditions.DocumentHash); err != nil {
		return "", 0, fmt.Errorf("accept terms failed: %w", err)
	}

	// Create Twin
	twinID, err = client.CreateTwin(identity, "", []byte{})
	if err != nil {
		return "", 0, fmt.Errorf("create twin failed: %w", err)
	}

	logger.GetLogger().Debug().Msgf("Twin created with ID %d for %s", twinID, identity.Address())
	return mnemonic, twinID, nil
}

// GenerateMnemonic generate mnemonic for each user
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

// Activates user account with activation service
func ActivateAccount(substrateAccountID string, url string) error {
	body := map[string]string{"substrateAccountID": substrateAccountID}
	jsonData, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal activation payload: %w", err)
	}

	resp, err := http.Post(url, "application/json", bytes.NewReader(jsonData))
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

// TransferTFTs transfer balance to users' account
func TransferTFTs(substrateClient *substrate.Substrate, tftBalance uint64, userMnemonic string, systemIdentity substrate.Identity) error {
	// Create identity of user from mnemonic
	userIdentity, err := substrate.NewIdentityFromSr25519Phrase(userMnemonic)
	if err != nil {
		return err
	}

	return substrateClient.Transfer(systemIdentity, tftBalance, substrate.AccountID(userIdentity.PublicKey()))
}

// GetUserBalanceUSDMillicent gets balance of user in USD Millicent
// This avoids floating point precision issues by returning an integer value
func GetUserBalanceUSDMillicent(substrateClient *substrate.Substrate, userMnemonic string) (uint64, error) {
	tftBalance, err := GetUserTFTBalance(substrateClient, userMnemonic)
	if err != nil {
		return 0, err
	}

	return FromTFTtoUSDMillicent(substrateClient, tftBalance)
}

// GetUserBalanceUSD gets balance of user in TFT
func GetUserTFTBalance(substrateClient *substrate.Substrate, userMnemonic string) (uint64, error) {
	if userMnemonic == "" {
		return 0, nil
	}

	// Create identity from mnemonic
	identity, err := substrate.NewIdentityFromSr25519Phrase(userMnemonic)
	if err != nil {
		return 0, err
	}

	account, err := substrate.FromAddress(identity.Address())
	if err != nil {
		return 0, err
	}

	// get balance in TFT
	tftBalance, err := substrateClient.GetBalance(account)
	if err != nil {
		return 0, err
	}

	return tftBalance.Free.Uint64(), nil
}

// FromTFTtoUSDMillicent converts TFT amount to USD Millicent (1/1000 of a dollar)
func FromTFTtoUSDMillicent(substrateClient *substrate.Substrate, amount uint64) (uint64, error) {
	price, err := substrateClient.GetTFTPrice()
	if err != nil {
		return 0, err
	}

	usdMillicentBalance := uint64(math.Round((float64(amount) / 1e7) * float64(price)))
	return usdMillicentBalance, nil
}

// FromUSDMillicentToTFT converts USD Millicent to TFT amount
// This avoids floating point precision issues by accepting an integer value
func FromUSDMillicentToTFT(substrateClient *substrate.Substrate, amountMillicent uint64) (uint64, error) {
	price, err := substrateClient.GetTFTPrice()
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
