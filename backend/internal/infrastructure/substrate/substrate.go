package substrate

import (
	"fmt"

	"github.com/cosmos/go-bip39"
	substrate "github.com/threefoldtech/tfchain/clients/tfchain-client-go"
)

type Substrate interface {
	SystemIdentity() substrate.Identity
	SetupUserOnTFChain() (mnemonic string, twinID uint32, err error)
	NewIdentityFromSr25519Phrase(mnemonic string) (substrate.Identity, error)
	TransferTFTsFromSystem(tftBalance uint64, userMnemonic string) error
	TransferTFTsToSystem(tftBalance uint64, userMnemonic string) error

	GetUserBalanceUSDMillicent(userMnemonic string) (uint64, error)
	GetUserBalanceUSD(userMnemonic string) (float64, error)
	GetUserTFTBalance(userMnemonic string) (uint64, error)

	FromTFTtoUSDMillicent(amount uint64) (uint64, error)
	FromUSDMillicentToTFT(amountMillicent uint64) (uint64, error)

	GetTwinByPubKey(pubKey []byte) (uint32, error)
	CreateRentContract(mnemonic string, nodeID uint32, solutionProviderID *uint64) (uint64, error)
	CancelContract(mnemonic string, contractID uint64) error
	Close()
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
