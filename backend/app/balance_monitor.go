package app

import (
	"kubecloud/internal"
	"kubecloud/models"
	"time"

	"kubecloud/internal/logger"

	"github.com/pkg/errors"
)

func (h *adminHandler) MonitorSystemBalanceAndHandleSettlement() {
	balanceTicker := time.NewTicker(time.Duration(h.monitorBalanceIntervalInMinutes) * time.Minute)
	adminNotifyTicker := time.NewTicker(time.Duration(h.notifyAdminsForPendingRecordsInHours) * time.Hour)
	defer balanceTicker.Stop()
	defer adminNotifyTicker.Stop()

	for {
		select {
		case <-balanceTicker.C:
			records, err := h.svc.prRepo.ListOnlyPendingRecords()
			if err != nil {
				continue
			}

			if err := h.settlePendingPayments(records); err != nil {
				logger.GetLogger().Error().Err(err).Send()
			}

		case <-adminNotifyTicker.C:
			records, err := h.svc.prRepo.ListOnlyPendingRecords()
			if err != nil {
				continue
			}

			if len(records) > 0 {
				if err := h.notifyAdminWithPendingRecords(records); err != nil {
					logger.GetLogger().Error().Err(err).Send()
				}
			}
		}
	}
}

func (h *adminHandler) settlePendingPayments(records []models.PendingRecord) error {
	for _, record := range records {
		// Already settled
		if record.TransferredTFTAmount >= record.TFTAmount {
			continue
		}

		// getting balance every time to ensure we have the latest balance
		systemTFTBalance, err := internal.GetUserTFTBalance(h.substrateClient, h.systemIdentity)
		if err != nil {
			logger.GetLogger().Error().Err(err).Msgf("Failed to get system TFT balance for pending record ID %d", record.ID)
			continue
		}

		amountToTransfer := record.TFTAmount - record.TransferredTFTAmount
		if systemTFTBalance < amountToTransfer {
			logger.GetLogger().Warn().Msgf("Insufficient system balance to settle pending record ID %d", record.ID)
			continue
		}

		if err = h.transferTFTsToUser(record.UserID, record.ID, amountToTransfer); err != nil {
			logger.GetLogger().Error().Err(err).Send()
			continue
		}
	}

	return nil
}

func (h *adminHandler) transferTFTsToUser(userID, recordID int, amountToTransfer uint64) error {
	user, err := h.svc.userRepo.GetUserByID(userID)
	if err != nil {
		return errors.Wrapf(err, "failed to get user for pending record ID %d", recordID)
	}

	err = internal.TransferTFTs(h.substrateClient, amountToTransfer, user.Mnemonic, h.systemIdentity)
	if err != nil {
		return errors.Wrapf(err, "Failed to transfer TFTs for pending record ID %d", recordID)
	}

	err = h.svc.prRepo.UpdatePendingRecordTransferredAmount(recordID, amountToTransfer)
	if err != nil {
		return errors.Wrapf(err, "Failed to update transferred amount for pending record ID %d", recordID)
	}

	return nil
}

func (h *adminHandler) notifyAdminWithPendingRecords(records []models.PendingRecord) error {
	subject, body := h.mailService.NotifyAdminsMailContent(len(records))

	admins, err := h.svc.userRepo.ListAdmins()
	if err != nil {
		return err
	}

	for _, admin := range admins {
		err = h.mailService.SendMailFromSystem(admin.Email, subject, body)
		if err != nil {
			logger.GetLogger().Error().Err(err).Send()
			continue
		}
	}

	return nil
}
