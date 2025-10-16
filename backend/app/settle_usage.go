package app

import (
	"context"
	"kubecloud/internal"
	"kubecloud/internal/logger"
	"kubecloud/models"
	"math"
	"strconv"
	"time"
)

type discount string

type DiscountPackage struct {
	DurationInMonth float64
	Discount        int
}

// DeductUSDBalanceBasedOnUsage deducts the user balance based on the usage
// This function is called every 24 hours
func (h *Handler) DeductUSDBalanceBasedOnUsage(ctx context.Context) {
	usageDeductionTicker := time.NewTicker(24 * time.Hour)
	defer usageDeductionTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-usageDeductionTicker.C:
			users, err := h.db.ListAllUsers()
			if err != nil {
				logger.GetLogger().Error().Err(err).Msg("Failed to list users")
				continue
			}

			for _, user := range users {
				if err := h.settleUserUsage(&user); err != nil {
					logger.GetLogger().Error().Err(err).Msgf("Failed to settle daily usage for user %d", user.ID)
				}
			}
		}
	}
}

func (h *Handler) settleUserUsage(user *models.User) error {
	usageInUSDMillicent, err := h.getUserLatestUsageInUSD(user.ID)
	if err != nil {
		return err
	}

	return h.db.DeductUserBalance(user, usageInUSDMillicent)
}

func (h *Handler) getUserLatestUsageInUSD(userID int) (uint64, error) {
	now := time.Now()
	// Define the end of the day (next day at 00:00)
	endOfDay := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, time.Local)

	// Get the last calculation time for this user from the database, or use a default if not available
	lastCalcTime, err := h.db.GetUserLastCalcTime(userID)
	if err != nil {
		return 0, err
	}

	// If this is the first time or no record exists, use the start of the day as default
	if lastCalcTime.IsZero() {
		lastCalcTime = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
	}

	contracts, err := h.db.ListAllContractsInPeriod(userID, lastCalcTime, endOfDay)
	if err != nil {
		return 0, err
	}

	if len(contracts) == 0 {
		return 0, nil
	}

	var totalDailyUsageInUSDMillicent uint64

	for _, record := range contracts {
		// Get bill reports from the last calculation time to the end of day
		billReports, err := internal.ListContractBillReports(h.graphqlClient, record.ContractID, lastCalcTime, endOfDay)
		if err != nil {
			return 0, err
		}

		totalAmountBilledInUSDMillicent, err := h.calculateTotalUsageOfReportsInUSDMillicent(billReports.Reports)
		if err != nil {
			return 0, err
		}

		totalDailyUsageInUSDMillicent += totalAmountBilledInUSDMillicent
	}

	// Update the last calculation time for this user in the database
	if err := h.db.UpdateUserLastCalcTime(userID, now); err != nil {
		logger.GetLogger().Error().Err(err).Msgf("Failed to update last calculation time for user %d", userID)
	}

	return totalDailyUsageInUSDMillicent, nil
}

func (h *Handler) calculateTotalUsageOfReportsInUSDMillicent(reports []internal.Report) (uint64, error) {
	var totalAmountBilledInUSDMillicent uint64
	for _, report := range reports {
		amountInTFT, err := removeDiscountFromReport(&report)
		if err != nil {
			return 0, err
		}

		amountInUSDMillicent, err := h.fromTFTtoUSDMillicent(amountInTFT, report)
		if err != nil {
			return 0, err
		}

		totalAmountBilledInUSDMillicent += amountInUSDMillicent
	}

	return totalAmountBilledInUSDMillicent, nil
}

func (h *Handler) fromTFTtoUSDMillicent(amount uint64, report internal.Report) (uint64, error) {
	price, err := h.getBillingRateAt(report)
	if err != nil {
		return 0, err
	}

	usdMillicentBalance := uint64(math.Round((float64(amount) / TFTUnitFactor) * float64(price)))
	return usdMillicentBalance, nil
}

func removeDiscountFromReport(report *internal.Report) (uint64, error) {
	discountPackage := getDiscountPackage(discount(report.DiscountReceived))

	amountBilled, err := strconv.ParseInt(report.AmountBilled, 10, 64)
	if err != nil {
		return 0, err
	}

	amountBilledWithNoDiscount := float64(amountBilled) / float64(1-discountPackage.Discount/100)
	return uint64(amountBilledWithNoDiscount), nil
}

func getDiscountPackage(discountInput discount) DiscountPackage {
	oneDayMargin := 1.0 / 30.0

	discountPackages := map[discount]DiscountPackage{
		"none": {
			DurationInMonth: oneDayMargin * 3,
			Discount:        0,
		},
		"default": {
			DurationInMonth: 1.5 + oneDayMargin,
			Discount:        20,
		},
		"bronze": {
			DurationInMonth: 3 + oneDayMargin,
			Discount:        30,
		},
		"silver": {
			DurationInMonth: 6 + oneDayMargin,
			Discount:        40,
		},
		"gold": {
			DurationInMonth: 10 + oneDayMargin,
			Discount:        60,
		},
	}

	return discountPackages[discountInput]
}

func (h *Handler) getBillingRateAt(report internal.Report) (float64, error) {
	block_duration := 6 // in seconds
	now := time.Now().Unix()

	reportTimestamp, err := strconv.ParseInt(report.Timestamp, 10, 64)
	if err != nil {
		return 0, err
	}

	timeBetweenNowAndReport := now - reportTimestamp // seconds

	// Calculate number of blocks since report
	numberOfBlocks := math.Round(float64(timeBetweenNowAndReport) / float64(block_duration))

	nowBlock, err := h.substrateClient.GetCurrentHeight()
	if err != nil {
		return 0, err
	}
	reportBlock := nowBlock - uint32(numberOfBlocks)

	tftPrice, err := h.substrateClient.GetTFTBillingRateAt(uint64(reportBlock))
	if err != nil {
		return 0, err
	}

	return float64(tftPrice), nil
}
