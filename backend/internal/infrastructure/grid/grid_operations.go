package grid

import (
	"math"

	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/deployer"
)

// FromTFTtoUSDMillicent converts TFT amount to USD Millicent (1/1000 of a dollar)
func FromTFTtoUSDMillicent(gridClient deployer.TFPluginClient, amount uint64) (uint64, error) {
	price, err := gridClient.SubstrateConn.GetTFTPrice()
	if err != nil {
		return 0, err
	}

	usdMillicentBalance := uint64(math.Round((float64(amount) / 1e7) * float64(price)))
	return usdMillicentBalance, nil
}

// FromUSDMillicentToTFT converts USD Millicent to TFT amount
// This avoids floating point precision issues by accepting an integer value
func FromUSDMillicentToTFT(gridClient deployer.TFPluginClient, amountMillicent uint64) (uint64, error) {
	price, err := gridClient.SubstrateConn.GetTFTPrice()
	if err != nil {
		return 0, err
	}

	// Convert Millicent to dollars for the calculation
	amountUSD := FromUSDMilliCentToUSD(amountMillicent)
	tft := (amountUSD * 1e7) / (float64(price) / 1000)
	return uint64(tft), nil
}

// GetUserBalanceUSDMillicent gets balance of user in USD Millicent
// This avoids floating point precision issues by returning an integer value
func GetUserBalanceUSDMillicent(gridClient deployer.TFPluginClient, userMnemonic string) (uint64, error) {
	// Create identity of user from mnemonic
	userIdentity, err := gridClient.SubstrateConn.NewIdentityFromSr25519Phrase(userMnemonic)
	if err != nil {
		return 0, err
	}

	tftBalance, err := gridClient.SubstrateConn.GetBalance(userIdentity)
	if err != nil {
		return 0, err
	}

	return FromTFTtoUSDMillicent(gridClient, tftBalance.Free.Uint64())
}

// GetUserBalanceUSD gets balance of user in USD
func GetUserBalanceUSD(gridClient deployer.TFPluginClient, userMnemonic string) (float64, error) {
	usdMillicentBalance, err := GetUserBalanceUSDMillicent(gridClient, userMnemonic)
	if err != nil {
		return 0, err
	}

	return FromUSDMilliCentToUSD(usdMillicentBalance), nil
}

// TransferTFTsFromSystem transfer balance to users' account
func TransferTFTsFromSystem(gridClient deployer.TFPluginClient, tftBalance uint64, userMnemonic string, systemMnemonic string) error {
	// Create identity of user from mnemonic
	userIdentity, err := gridClient.SubstrateConn.NewIdentityFromSr25519Phrase(userMnemonic)
	if err != nil {
		return err
	}

	// Create identity of system from mnemonic
	systemIdentity, err := gridClient.SubstrateConn.NewIdentityFromSr25519Phrase(systemMnemonic)
	if err != nil {
		return err
	}

	return gridClient.SubstrateConn.Transfer(tftBalance, systemIdentity, userIdentity.PublicKey())
}

// TransferTFTsToSystem transfer balance to system account
func TransferTFTsToSystem(gridClient deployer.TFPluginClient, tftBalance uint64, userMnemonic string, systemMnemonic string) error {
	// Create identity of user from mnemonic
	userIdentity, err := gridClient.SubstrateConn.NewIdentityFromSr25519Phrase(userMnemonic)
	if err != nil {
		return err
	}

	// Create identity of system from mnemonic
	systemIdentity, err := gridClient.SubstrateConn.NewIdentityFromSr25519Phrase(systemMnemonic)
	if err != nil {
		return err
	}

	return gridClient.SubstrateConn.Transfer(tftBalance, userIdentity, systemIdentity.PublicKey())
}

func FromUSDMilliCentToUSD(amountMillicent uint64) float64 {
	return float64(amountMillicent) / 1000
}

func FromUSDToUSDMillicent(amountUSD float64) uint64 {
	return uint64(amountUSD * 1000)
}
