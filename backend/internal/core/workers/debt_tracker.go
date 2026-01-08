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

func (w Workers) TrackUsersBalance() {
	ticker := time.NewTicker(w.svc.GetUsersBalanceCheckInterval())
	defer ticker.Stop()
	log := logger.ForOperation("debt_tracker", "track_users_balance")
	log.Info().Msg("Starting users balance check")
	// run once on startup
	if err := w.svc.CheckUsersBalance(); err != nil {
		log.Error().Err(err).Msg("Failed to check users balance on startup")
	}

	for {
		select {
		case <-w.ctx.Done():
			return
		case <-ticker.C:
			log.Info().Msg("Checking users balance")
			if err := w.svc.CheckUsersBalance(); err != nil {
				log.Error().Err(err).Msg("Failed to check users balance")
			}
		}
	}
}
