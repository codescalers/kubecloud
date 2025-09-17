package app

import (
	"context"
	"kubecloud/internal"
	"kubecloud/models"
	"time"

	"kubecloud/internal/logger"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	substrate "github.com/threefoldtech/tfchain/clients/tfchain-client-go"
	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/calculator"
	"github.com/threefoldtech/tfgrid-sdk-go/grid-proxy/pkg/types"
)

const zeroTFTBalanceValue = 0.05 * 1e7 // 0.05 TFT

func (h *Handler) MonitorSystemBalanceAndHandleSettlement(ctx context.Context) {
	settleTransfersTicker := time.NewTicker(time.Duration(h.config.SettleTransferRecordsIntervalInMinutes) * time.Minute)
	adminNotifyTicker := time.NewTicker(time.Duration(h.config.NotifyAdminsForPendingRecordsInHours) * time.Hour)
	zeroUSDBalanceTicker := time.NewTicker(time.Minute)
	zeroTFTBalanceTicker := time.NewTicker(time.Minute)
	fundUserTFTBalanceTicker := time.NewTicker(24 * time.Hour)
	defer settleTransfersTicker.Stop()
	defer adminNotifyTicker.Stop()
	defer zeroUSDBalanceTicker.Stop()
	defer zeroTFTBalanceTicker.Stop()
	defer fundUserTFTBalanceTicker.Stop()

	for {
		users, err := h.db.ListAllUsers()
		if err != nil {
			continue
		}

		select {
		case <-settleTransfersTicker.C:
			records, err := h.db.ListPendingTransferRecords()
			if err != nil {
				continue
			}

			if err := h.settlePendingPayments(records); err != nil {
				logger.GetLogger().Error().Err(err).Send()
			}

		case <-adminNotifyTicker.C:
			records, err := h.db.ListPendingTransferRecords()
			if err != nil {
				continue
			}

			if len(records) > 0 {
				if err := h.notifyAdminWithPendingRecords(records); err != nil {
					logger.GetLogger().Error().Err(err).Send()
				}
			}

		case <-zeroUSDBalanceTicker.C:
			if err := h.resetUsersTFTsWithNoUSDBalance(users); err != nil {
				log.Error().Err(err).Send()
			}

		case <-zeroTFTBalanceTicker.C:
			for _, user := range users {
				if user.CreditedBalance+user.CreditCardBalance > zeroTFTBalanceValue {
					continue
				}

				if err := h.db.CreateTransferRecord(&models.TransferRecord{
					UserID:    user.ID,
					Username:  user.Username,
					TFTAmount: uint64(h.config.MinimumTFTAmountInWallet) * 1e7,
					Operation: models.DepositOperation,
				}); err != nil {
					log.Error().Err(err).Msgf("Failed to create transfer record for user %d", user.ID)
				}
			}

		case <-fundUserTFTBalanceTicker.C:
			for _, user := range users {
				if err = h.fundUsersToClaimDiscount(ctx, user.ID, user.Username, user.Mnemonic, discount(h.config.AppliedDiscount)); err != nil {
					log.Error().Err(err).Msgf("Failed to fund user %d to claim discount", user.ID)
				}
			}
		}
	}
}

func (h *Handler) resetUsersTFTsWithNoUSDBalance(users []models.User) error {
	for _, user := range users {
		if user.CreditedBalance+user.CreditCardBalance == 0 {
			log.Info().Msgf("User %d has no USD balance, withdrawing all TFTs except for %d", user.ID, h.config.MinimumTFTAmountInWallet)

			userTFTBalance, err := internal.GetUserTFTBalance(h.substrateClient, user.Mnemonic)
			if err != nil {
				log.Error().Err(err).Msgf("Failed to get user TFT balance for user %d", user.ID)
				continue
			}

			if userTFTBalance <= uint64(h.config.MinimumTFTAmountInWallet)*1e7 {
				continue
			}

			transferRecord := models.TransferRecord{
				UserID:    user.ID,
				Username:  user.Username,
				TFTAmount: userTFTBalance,
				Operation: models.WithdrawOperation,
				State:     models.SuccessState,
			}

			if err = h.withdrawTFTsFromUser(user.ID, user.Mnemonic, userTFTBalance); err != nil {
				log.Error().Err(err).Msgf("Failed to withdraw all TFTs for user %d", user.ID)

				// TODO: handle retries
				transferRecord.State = models.FailedState
				transferRecord.Failure = err.Error()
			}

			if err := h.db.CreateTransferRecord(&transferRecord); err != nil {
				log.Error().Err(err).Msgf("Failed to create transfer record for user %d", user.ID)
			}
		}
	}

	return nil
}

func (h *Handler) settlePendingPayments(records []models.TransferRecord) error {
	for _, record := range records {
		if record.Operation == models.WithdrawOperation {
			continue
		}

		// Already settled
		if record.State == models.SuccessState {
			continue
		}

		transferState := models.SuccessState
		var transferFailure string

		// getting balance every time to ensure we have the latest balance
		systemTFTBalance, err := internal.GetUserTFTBalance(h.substrateClient, h.config.SystemAccount.Mnemonic)
		if err != nil {
			logger.GetLogger().Error().Err(err).Msgf("Failed to get system TFT balance for pending record ID %d", record.ID)
			continue
		}

		if systemTFTBalance < record.TFTAmount {
			log.Warn().Msgf("Insufficient system balance to settle pending record ID %d", record.ID)
			continue
		}

		if err = h.transferTFTsToUser(record.UserID, record.TFTAmount); err != nil {
			log.Error().Err(err).Msgf("Failed to settle pending record ID %d", record.ID)

			transferState = models.FailedState
			transferFailure = err.Error()
		}

		if err := h.db.UpdateTransferRecordState(record.ID, transferState, transferFailure); err != nil {
			log.Error().Err(err).Msgf("Failed to update pending record ID %d state", record.ID)
		}
	}

	return nil
}

func (h *Handler) transferTFTsToUser(userID int, amountToTransfer uint64) error {
	user, err := h.db.GetUserByID(userID)
	if err != nil {
		return errors.Wrapf(err, "failed to get user %d", userID)
	}

	err = internal.TransferTFTs(h.substrateClient, amountToTransfer, user.Mnemonic, h.systemIdentity)
	if err != nil {
		return errors.Wrapf(err, "Failed to transfer TFTs to user %d", userID)
	}

	return nil
}

func (h *Handler) withdrawTFTsFromUser(userID int, userMnemonic string, amountToWithdraw uint64) error {
	userIdentity, err := substrate.NewIdentityFromSr25519Phrase(userMnemonic)
	if err != nil {
		return errors.Wrapf(err, "Failed to create identity for user %d", userID)
	}

	err = internal.TransferTFTs(h.substrateClient, amountToWithdraw, h.config.SystemAccount.Mnemonic, userIdentity)
	if err != nil {
		return errors.Wrapf(err, "Failed to transfer TFTs to user %d", userID)
	}

	return nil
}

func (h *Handler) fundUsersToClaimDiscount(ctx context.Context, userID int, Username, userMnemonic string, configuredDiscount discount) error {
	rentedNodes, _, err := h.getRentedNodesForUser(ctx, userID, true)
	if err != nil {
		return err
	}

	dailyUsageInUSDMillicent, err := h.calculateResourcesUsageInUSDApplyingDiscount(userID, userMnemonic, rentedNodes, configuredDiscount)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to calculate resources usage in USD for user %d", userID)
		return err
	}

	if dailyUsageInUSDMillicent > 0 {
		tftAmount, err := internal.FromUSDMillicentToTFT(h.substrateClient, dailyUsageInUSDMillicent)
		if err != nil {
			log.Error().Err(err).Msgf("Failed to convert USD to TFTs for user %d", userID)
			return err
		}

		if err := h.db.CreateTransferRecord(&models.TransferRecord{
			UserID:    userID,
			Username:  Username,
			TFTAmount: tftAmount,
			Operation: models.DepositOperation,
		}); err != nil {
			log.Error().Err(err).Msgf("Failed to create transfer record for user %d", userID)
			return err
		}
	}

	return nil
}

func (h *Handler) calculateResourcesUsageInUSDApplyingDiscount(
	userID int,
	userMnemonic string,
	rentedNodes []types.Node,
	configuredDiscount discount,
) (uint64, error) {
	userIdentity, err := substrate.NewIdentityFromSr25519Phrase(userMnemonic)
	if err != nil {
		return 0, errors.Wrapf(err, "Failed to create identity for user %d", userID)
	}

	calculator := calculator.NewCalculator(h.gridClient.SubstrateConn, userIdentity)

	var totalResourcesCostMillicent uint64
	for _, node := range rentedNodes {
		resourcesCost, err := calculator.CalculateCost(
			node.TotalResources.CRU,
			uint64(node.TotalResources.MRU),
			uint64(node.TotalResources.HRU),
			uint64(node.TotalResources.SRU),
			len(node.PublicConfig.Ipv4) > 0,
			len(node.CertificationType) > 0,
		)
		if err != nil {
			return 0, err
		}

		// resources cost per month
		totalResourcesCostMillicent += internal.FromUSDToUSDMillicent(resourcesCost)
	}

	return uint64(float64(totalResourcesCostMillicent) * getDiscountPackage(configuredDiscount).DurationInMonth), nil
}

func (h *Handler) notifyAdminWithPendingRecords(records []models.TransferRecord) error {
	subject, body := h.mailService.NotifyAdminsMailContent(len(records), h.config.Server.Host)

	admins, err := h.db.ListAdmins()
	if err != nil {
		return err
	}

	for _, admin := range admins {
		err = h.mailService.SendMail(h.config.MailSender.Email, admin.Email, subject, body)
		if err != nil {
			logger.GetLogger().Error().Err(err).Send()
			continue
		}
	}

	return nil
}
