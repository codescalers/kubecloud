package workers

import (
	"kubecloud/internal/infrastructure/logger"
	"time"
)

// DeductUSDBalanceBasedOnUsage deducts the user balance based on the usage
// This function is called every 24 hours
func (w Workers) DeductUSDBalanceBasedOnUsage() {
	usageDeductionTicker := time.NewTicker(24 * time.Hour)
	defer usageDeductionTicker.Stop()

	for {
		select {
		case <-w.ctx.Done():
			return
		case <-usageDeductionTicker.C:
			users, err := w.svc.ListAllUsers()
			if err != nil {
				logger.GetLogger().Error().Err(err).Msg("Failed to list users")
				continue
			}

			for _, user := range users {
				if err := w.billingService.SettleUserUsage(&user); err != nil {
					logger.GetLogger().Error().Err(err).Msgf("Failed to settle daily usage for user %d", user.ID)
				}
			}
		}
	}
}
