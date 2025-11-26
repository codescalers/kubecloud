package substrate

import (
	"kubecloud/internal/config"

	substrate "github.com/threefoldtech/tfchain/clients/tfchain-client-go"
	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/calculator"
	client "github.com/threefoldtech/tfgrid-sdk-go/grid-client/node"
)

type Substrate interface {
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
}
