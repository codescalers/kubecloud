package workers

import (
	"kubecloud/internal/infrastructure/logger"
	"time"
)

func (w Workers) MonthlyInvoicesHandler() {
	baseLog := logger.ForOperation("invoices", "monthly_invoices")
	var lastProcessedMonth time.Month
	var lastProcessedYear int

	for {
		now := time.Now()

		// Calculate sleep duration until month-end
		sleepDuration := calculateSleepDurationTillMonthEndForInvoice(now, lastProcessedMonth, lastProcessedYear)

		// Sleep with context awareness (single select)
		if sleepDuration > 0 {
			select {
			case <-w.ctx.Done():
				return
			case <-time.After(sleepDuration):
			}
			continue
		}

		// Process invoices (we're on month-end and haven't processed yet)
		users, err := w.svc.ListAllUsers()
		if err != nil {
			baseLog.Error().Err(err).Msg("failed to retrieve users for invoice creation")
			continue
		}

		for _, user := range users {
			if err = w.svc.CreateUserInvoice(user); err != nil {
				baseLog.Error().Err(err).Int("user_id", user.ID).Msg("failed to create invoice for user")
			}
		}

		//update last processed month and year
		lastProcessedMonth = now.Month()
		lastProcessedYear = now.Year()
	}
}

func calculateSleepDurationTillMonthEndForInvoice(now time.Time, lastProcessedMonth time.Month, lastProcessedYear int) time.Duration {
	monthLastDay := time.Date(now.Year(), now.Month()+1, 1, 0, 0, 0, 0, now.Location()).AddDate(0, 0, -1)

	// Calculate sleep duration until month-end
	var sleepDuration time.Duration
	if now.Day() != monthLastDay.Day() {
		sleepDuration = monthLastDay.Sub(now)
	}

	//check if invoice for this month and year is already processed
	if now.Month() == lastProcessedMonth && now.Year() == lastProcessedYear {
		// Already processed, sleep until the first day of the next month to avoid running multiple times on the last day
		nextMonth := time.Date(now.Year(), now.Month()+1, 1, 0, 5, 0, 0, now.Location())
		sleepDuration = nextMonth.Sub(now)
	}

	return sleepDuration
}
