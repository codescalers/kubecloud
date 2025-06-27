package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	substrate "github.com/threefoldtech/tfchain/clients/tfchain-client-go"

	"github.com/rs/zerolog/log"
	"github.com/tyler-smith/go-bip39"
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
	if err := ActivateAccount(identity.Address(), config.ActivationService.URL); err != nil {
		return "", 0, fmt.Errorf("activation failed: %w", err)
	}

	// Wait a few seconds for account activation to complete
	time.Sleep(5 * time.Second)

	if err := client.AcceptTermsAndConditions(identity, config.TermsANDConditions.DocumentLink, config.TermsANDConditions.DocumentHash); err != nil {
		return "", 0, fmt.Errorf("accept terms failed: %w", err)
	}

	// Create Twin
	twinID, err = client.CreateTwin(identity, "", []byte{})
	if err != nil {
		return "", 0, fmt.Errorf("create twin failed: %w", err)
	}

	log.Debug().Msgf("Twin created with ID %d for %s", twinID, identity.Address())
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

	fmt.Println(resp.StatusCode)

	return nil
}
