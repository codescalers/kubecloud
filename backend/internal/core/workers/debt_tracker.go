package workers

import (
	"time"

	"kubecloud/internal/core/services"
	"kubecloud/internal/infrastructure/logger"
)

func (w Workers) TrackUserDebt() {
	ticker := time.NewTicker(services.TrackingDebtPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-w.ctx.Done():
			return
		case <-ticker.C:
			if err := w.svc.UpdateUserDebt(); err != nil {
				logger.ForOperation("debt_tracker", "update_user_debt").Error().Err(err).Msg("Failed to update user debt")
			}
		}
	}
}
