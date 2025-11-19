package workers

import (
	"time"

	"kubecloud/internal/infrastructure/logger"
)

func (w Workers) TrackUserDebt() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-w.ctx.Done():
			return
		case <-ticker.C:
			if err := w.svc.UpdateUserDebt(); err != nil {
				logger.ForOperation("debt_tracker", "update_user_debt").Error().Err(err).Msg("Failed to update user debt")
				logWorkerAudit(
					logger.AuditActionWorkerDebtUpdate,
					logger.AuditSeverityError,
					map[string]any{
						"reason": err.Error(),
					},
				)
				continue
			}

			logWorkerAudit(
				logger.AuditActionWorkerDebtUpdate,
				logger.AuditSeverityInfo,
				map[string]any{
					"result": "user_debt_updated",
				},
			)
		}
	}
}
