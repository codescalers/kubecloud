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
				logWorkerAudit(
					logger.AuditActionWorkerBalanceSettle,
					logger.AuditSeverityError,
					map[string]any{
						"reason": err.Error(),
					},
				)
				continue
			}

			w.svc.SettlePendingPayments(records)
			logWorkerAudit(
				logger.AuditActionWorkerBalanceSettle,
				logger.AuditSeverityInfo,
				map[string]any{
					"records": len(records),
					"result":  "settlement_triggered",
				},
			)

		case <-adminNotifyTicker.C:
			records, err := w.svc.ListOnlyPendingRecords()
			if err != nil {
				logWorkerAudit(
					logger.AuditActionWorkerPendingNotify,
					logger.AuditSeverityError,
					map[string]any{
						"reason": err.Error(),
					},
				)
				continue
			}

			if len(records) == 0 {
				continue
			}

			if err := w.svc.NotifyAdminWithPendingRecords(records); err != nil {
				logger.ForOperation("balance_monitor", "notify_admins_pending_records").Error().Err(err).Msg("Failed to notify admins with pending records")
				logWorkerAudit(
					logger.AuditActionWorkerPendingNotify,
					logger.AuditSeverityError,
					map[string]any{
						"reason":  err.Error(),
						"records": len(records),
					},
				)
				continue
			}

			logWorkerAudit(
				logger.AuditActionWorkerPendingNotify,
				logger.AuditSeverityInfo,
				map[string]any{
					"records": len(records),
					"result":  "admins_notified",
				},
			)
		}
	}
}
