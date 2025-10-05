package app

import (
	"kubecloud/internal"
	"time"

	"kubecloud/internal/logger"

	substrate "github.com/threefoldtech/tfchain/clients/tfchain-client-go"
	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/calculator"
	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/deployer"
)

const TrackingDeptPeriod = time.Hour

func (h *Handler) TrackUserDebt(gridClient deployer.TFPluginClient) {
	ticker := time.NewTicker(TrackingDeptPeriod)
	defer ticker.Stop()

	for range ticker.C {
		if err := h.updateUserDebt(gridClient); err != nil {
			logger.GetLogger().Error().Err(err).Send()
		}
	}
}

func (h *Handler) updateUserDebt(gridClient deployer.TFPluginClient) error {
	users, err := h.db.ListAllUsers()
	if err != nil {
		return err
	}

	for _, user := range users {
		userContracts, err := h.db.ListAllContractsInPeriod(user.ID, time.Now().Add(-TrackingDeptPeriod), time.Now())
		if err != nil {
			logger.GetLogger().Error().Err(err).Send()
			continue
		}

		// Create identity from mnemonic
		identity, err := substrate.NewIdentityFromSr25519Phrase(user.Mnemonic)
		if err != nil {
			logger.GetLogger().Error().Err(err).Send()
			continue
		}

		var totalDebt int64
		for _, contract := range userContracts {
			calculatorClient := calculator.NewCalculator(gridClient.SubstrateConn, identity)
			debt, err := calculatorClient.CalculateContractOverdue(contract.ContractID, time.Hour)
			if err != nil {
				logger.GetLogger().Error().Err(err).Send()
				continue
			}
			totalDebt += debt

		}

		totalDebtUSD, err := internal.FromTFTtoUSDMillicent(h.substrateClient, uint64(totalDebt))
		if err != nil {
			logger.GetLogger().Error().Err(err).Send()
			continue
		}
		user.Debt = totalDebtUSD
		err = h.db.UpdateUserByID(&user)
		if err != nil {
			logger.GetLogger().Error().Err(err).Send()
		}
	}

	return nil
}
