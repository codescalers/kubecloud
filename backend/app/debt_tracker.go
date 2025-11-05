package app

import (
	"context"
	"kubecloud/internal"
	"time"

	"kubecloud/internal/logger"

	substrate "github.com/threefoldtech/tfchain/clients/tfchain-client-go"
	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/calculator"
	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/deployer"
)

func (h *Handler) TrackUserDebt(ctx context.Context, gridClient deployer.TFPluginClient) {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := h.updateUserDebt(gridClient); err != nil {
				logger.ForOperation("debt_tracker", "update_user_debt").Error().Err(err).Msg("Failed to update user debt")
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
		userNodes, err := h.db.ListUserNodes(user.ID)
		if err != nil {
			logger.ForOperation("debt_tracker", "list_user_nodes").Error().Err(err).Msg("Failed to list user nodes")
			continue
		}
		// Create identity from mnemonic
		identity, err := substrate.NewIdentityFromSr25519Phrase(user.Mnemonic)
		if err != nil {
			logger.ForOperation("debt_tracker", "new_identity").Error().Err(err).Msg("Failed to create identity from mnemonic")
			continue
		}

		var totalDebt int64
		for _, node := range userNodes {
			calculatorClient := calculator.NewCalculator(gridClient.SubstrateConn, identity)
			debt, err := calculatorClient.CalculateContractOverdue(node.ContractID, time.Hour)
			if err != nil {
				logger.ForOperation("debt_tracker", "calc_overdue").Error().Err(err).Msg("Failed to calculate contract overdue")
				continue
			}
			totalDebt += debt

		}

		totalDebtUSD, err := internal.FromTFTtoUSDMillicent(h.substrateClient, uint64(totalDebt))
		if err != nil {
			logger.ForOperation("debt_tracker", "tft_to_usd_millicent").Error().Err(err).Msg("Failed to convert debt to USD millicent")
			continue
		}
		user.Debt = totalDebtUSD
		err = h.db.UpdateUserByID(&user)
		if err != nil {
			logger.ForOperation("debt_tracker", "update_user_debt_db").Error().Err(err).Msg("Failed to update user debt in DB")
		}
	}

	return nil
}
