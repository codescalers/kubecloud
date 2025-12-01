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
			}
		}
	}
}

func (w Workers) TrackUsersBalance() {
	ticker := time.NewTicker(w.svc.GetUsersBalanceCheckInterval())
	defer ticker.Stop()

	for {
		select {
		case <-w.ctx.Done():
			return
		case <-ticker.C:
			if err := w.svc.CheckUsersBalance(); err != nil {
				logger.ForOperation("debt_tracker", "check_users_balance").Error().Err(err).Msg("Failed to check users balance")
			}
		}
	}
}
