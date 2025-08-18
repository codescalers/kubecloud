package app

import (
	"kubecloud/internal"
	"kubecloud/models"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

func (h *Handler) MonitorSystemBalanceAndHandleSettlement() {
	balanceTicker := time.NewTicker(time.Duration(h.config.MonitorBalanceIntervalInMinutes) * time.Minute)
	adminNotifyTicker := time.NewTicker(time.Duration(h.config.NotifyAdminsForPendingRecordsInHours) * time.Hour)
	defer balanceTicker.Stop()
	defer adminNotifyTicker.Stop()

	for {
		select {
		case <-balanceTicker.C:
			records, err := h.db.ListOnlyPendingRecords()
			if err != nil {
				continue
			}

			if err := h.settlePendingPayments(records); err != nil {
				log.Error().Err(err).Send()
			}

		case <-adminNotifyTicker.C:
			records, err := h.db.ListOnlyPendingRecords()
			if err != nil {
				continue
			}

			if len(records) > 0 {
				if err := h.notifyAdminWithPendingRecords(records); err != nil {
					log.Error().Err(err).Send()
				}
			}
		}
	}
}

func (h *Handler) settlePendingPayments(records []models.PendingRecord) error {
	for _, record := range records {
		// Already settled
		if record.TransferredTFTAmount >= record.TFTAmount {
			continue
		}

		// getting balance every time to ensure we have the latest balance
		systemTFTBalance, err := internal.GetUserTFTBalance(h.substrateClient, h.config.SystemAccount.Mnemonic)
		if err != nil {
			log.Error().Err(err).Msgf("Failed to get system TFT balance for pending record ID %d", record.ID)
			continue
		}

		amountToTransfer := record.TFTAmount - record.TransferredTFTAmount
		if systemTFTBalance < amountToTransfer {
			log.Warn().Msgf("Insufficient system balance to settle pending record ID %d", record.ID)
			continue
		}

		if err = h.transferTFTsToUser(record.UserID, record.ID, amountToTransfer); err != nil {
			log.Error().Err(err).Send()
			continue
		}
	}

	return nil
}

func (h *Handler) transferTFTsToUser(userID, recordID int, amountToTransfer uint64) error {
	user, err := h.db.GetUserByID(userID)
	if err != nil {
		return errors.Wrapf(err, "failed to get user for pending record ID %d", recordID)
	}

	err = internal.TransferTFTs(h.substrateClient, amountToTransfer, user.Mnemonic, h.systemIdentity)
	if err != nil {
		return errors.Wrapf(err, "Failed to transfer TFTs for pending record ID %d", recordID)
	}

	err = h.db.UpdatePendingRecordTransferredAmount(recordID, amountToTransfer)
	if err != nil {
		return errors.Wrapf(err, "Failed to update transferred amount for pending record ID %d", recordID)
	}

	return nil
}

func (h *Handler) notifyAdminWithPendingRecords(records []models.PendingRecord) error {
	subject, body := h.mailService.NotifyAdminsMailContent(len(records), h.config.Server.Host)

	admins, err := h.db.ListAdmins()
	if err != nil {
		return err
	}

	for _, admin := range admins {
		err = h.mailService.SendMail(h.config.MailSender.Email, admin.Email, subject, body)
		if err != nil {
			log.Error().Err(err).Send()
			continue
		}
	}

	return nil
}
