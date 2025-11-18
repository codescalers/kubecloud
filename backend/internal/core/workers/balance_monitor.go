package workers

import (
	"time"

	"kubecloud/internal/infrastructure/logger"
)

func (w Workers) MonitorSystemBalanceAndHandleSettlement() {
	balanceTicker := time.NewTicker(w.svc.GetMonitorBalanceInterval())
	adminNotifyTicker := time.NewTicker(w.svc.GetNotifyAdminsForPendingRecordsInterval())
	defer balanceTicker.Stop()
	defer adminNotifyTicker.Stop()

	for {
		select {
		case <-w.ctx.Done():
			return
		case <-balanceTicker.C:
			records, err := w.svc.ListOnlyPendingRecords()
			if err != nil {
				continue
			}

			w.svc.SettlePendingPayments(records)

		case <-adminNotifyTicker.C:
			records, err := w.svc.ListOnlyPendingRecords()
			if err != nil {
				continue
			}

			if len(records) > 0 {
				if err := w.svc.NotifyAdminWithPendingRecords(records); err != nil {
					logger.ForOperation("balance_monitor", "notify_admins_pending_records").Error().Err(err).Msg("Failed to notify admins with pending records")
				}
			}
		}
	}
}
