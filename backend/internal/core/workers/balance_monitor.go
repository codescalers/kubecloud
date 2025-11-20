package workers

import (
	"time"

	"kubecloud/internal/infrastructure/logger"
)

func (w Workers) MonitorSystemBalanceAndHandleSettlement() {
	settleTransfersTicker := time.NewTicker(w.svc.GetSettleTransferRecordsInterval())
	adminNotifyTicker := time.NewTicker(w.svc.GetNotifyAdminsForPendingRecordsInterval())
	zeroUSDBalanceTicker := time.NewTicker(time.Minute)
	zeroTFTBalanceTicker := time.NewTicker(time.Minute)
	fundUserTFTBalanceTicker := time.NewTicker(24 * time.Hour)
	defer settleTransfersTicker.Stop()
	defer adminNotifyTicker.Stop()
	defer zeroUSDBalanceTicker.Stop()
	defer zeroTFTBalanceTicker.Stop()
	defer fundUserTFTBalanceTicker.Stop()

	for {
		select {
		case <-w.ctx.Done():
			return
		case <-settleTransfersTicker.C:
			records, err := w.svc.ListPendingTransferRecords()
			if err != nil {
				continue
			}

			failedRecords, err := w.svc.ListFailedTransferRecords()
			if err != nil {
				continue
			}

			w.svc.SettlePendingPayments(append(records, failedRecords...))

		case <-adminNotifyTicker.C:
			records, err := w.svc.ListPendingTransferRecords()
			if err != nil {
				continue
			}

			if len(records) == 0 {
				continue
			}

			if err := w.svc.NotifyAdminWithPendingRecords(records); err != nil {
				logger.ForOperation("balance_monitor", "notify_admins_pending_records").Error().Err(err).Msg("Failed to notify admins with pending records")
			}

		case <-zeroUSDBalanceTicker.C:
			users, err := w.svc.ListAllUsers()
			if err != nil {
				logger.GetLogger().Error().Err(err).Msg("Failed to list users")
				continue
			}

			if err := w.svc.ResetUsersTFTsWithNoUSDBalance(users); err != nil {
				logger.GetLogger().Error().Err(err).Send()
			}

		case <-zeroTFTBalanceTicker.C:
			users, err := w.svc.ListAllUsers()
			if err != nil {
				logger.GetLogger().Error().Err(err).Msg("Failed to list users")
				continue
			}

			for _, user := range users {
				clusters, err := w.svc.ListUserClusters(user.ID)
				if err != nil {
					logger.GetLogger().Error().Err(err).Msgf("Failed to list user clusters")
					continue
				}

				if len(clusters) > 0 {
					// user has deployed workloads, skip
					continue
				}

				if user.CreditedBalance+user.CreditCardBalance-user.Debt <= 0 {
					continue
				}

				if err := w.billingService.CreateTransferRecordToChargeUserWithMinTFTAmount(user.ID, user.Username, user.Mnemonic); err != nil {
					logger.GetLogger().Error().Err(err).Msgf("Failed to create transfer record for user %d", user.ID)
				}
			}

		case <-fundUserTFTBalanceTicker.C:
			users, err := w.svc.ListAllUsers()
			if err != nil {
				logger.GetLogger().Error().Err(err).Msg("Failed to list users")
				continue
			}

			for _, user := range users {
				if err := w.billingService.FundUserToFulfillDiscount(w.ctx, &user, nil, nil); err != nil {
					logger.GetLogger().Error().Err(err).Msgf("Failed to fund user %d to claim discount", user.ID)
				}
			}
		}
	}
}
