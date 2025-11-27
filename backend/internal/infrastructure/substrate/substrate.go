package substrate

import (
	"fmt"
	"kubecloud/internal/config"

	"github.com/cosmos/go-bip39"
	substrate "github.com/threefoldtech/tfchain/clients/tfchain-client-go"
	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/calculator"
	client "github.com/threefoldtech/tfgrid-sdk-go/grid-client/node"
	proxy "github.com/threefoldtech/tfgrid-sdk-go/grid-proxy/pkg/client"
)

type Substrate interface {
	GridProxyClient() proxy.Client
	FromTFTtoUSDMillicent(amount uint64) (uint64, error)
	FromUSDMillicentToTFT(amountMillicent uint64) (uint64, error)
	GetUserBalanceUSDMillicent(userMnemonic string) (uint64, error)
	GetUserBalanceUSD(userMnemonic string) (float64, error)
	TransferTFTsFromSystem(tftBalance uint64, userMnemonic string) error
	TransferTFTsToSystem(tftBalance uint64, userMnemonic string) error
	SystemIdentity() (substrate.Identity, error)
	GetTwinIDFromUserMnemonic(mnemonic string) (uint64, error)
	GetNodeClient(nodeID uint32) (*client.NodeClient, error)
	NewCalculator(mnemonic string) (calculator.Calculator, error)
	GetFreeBalanceTFT(mnemonic string) (uint64, error)
	GetUserAddress(mnemonic string) (string, error)
	AcceptTermsAndConditions(mnemonic, docLink, docHash string) error
	CreateTwin(mnemonic string) (uint32, error)
	CreateRentContract(mnemonic string, nodeID uint32) (uint64, error)
	CancelContract(mnemonic string, contractID uint64) error
	SetupUserOnTFChain(termsAndConditions config.TermsANDConditions, network string) (mnemonic string, twinID uint32, err error)
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
