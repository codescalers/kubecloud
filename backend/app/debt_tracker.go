package app

import (
	"context"
	"kubecloud/internal"
	"time"

	"kubecloud/internal/logger"

	"github.com/cenkalti/backoff/v4"
	substrate "github.com/threefoldtech/tfchain/clients/tfchain-client-go"
	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/calculator"
	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/deployer"
)

const (
	TrackingDebtPeriod = time.Hour
	reties             = 3
)

func (h *Handler) TrackUserDebt(ctx context.Context, gridClient deployer.TFPluginClient) {
	ticker := time.NewTicker(TrackingDebtPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.GetLogger().Info().Msg("Debt tracker stopping due to context cancellation")
			return
		case <-ticker.C:
			if err := h.updateUserDebt(gridClient); err != nil {
				logger.GetLogger().Error().Err(err).Send()
			}
		}
	}
}

func (h *Handler) updateUserDebt(gridClient deployer.TFPluginClient) error {
	users, err := h.db.ListAllUsers()
	if err != nil {
		return err
	}

	for _, user := range users {
		userContracts, err := h.db.ListAllContractsInPeriod(user.ID, time.Now().Add(-TrackingDebtPeriod), time.Now())
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

			var debt int64
			err = backoff.Retry(func() error {
				debt, err = calculatorClient.CalculateContractOverdue(contract.ContractID, time.Hour)
				return err
			}, backoff.WithMaxRetries(
				backoff.NewExponentialBackOff(),
				reties,
			))
			if err != nil {
				logger.GetLogger().Error().
					Uint64("contract_id", contract.ContractID).
					Int("max_retries", reties).
					Err(err).
					Msg("Failed to calculate contract overdue after maximum retries")
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
