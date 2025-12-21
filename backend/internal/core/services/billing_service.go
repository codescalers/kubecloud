package services

import (
	"context"
	"fmt"
	"kubecloud/internal/billing"
	"kubecloud/internal/core/models"
	"kubecloud/internal/deployment/kubedeployer"
	"kubecloud/internal/infrastructure/gridclient"
	"kubecloud/internal/infrastructure/logger"
	"math"
	"strconv"
	"time"

	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/graphql"
	"github.com/threefoldtech/tfgrid-sdk-go/grid-proxy/pkg/types"
)

type Discount string

type DiscountPackage struct {
	DurationInMonth float64
	Discount        int
}

type BillingService struct {
	userRepo            models.UserRepository
	contractsRepo       models.ContractDataRepository
	transferRecordsRepo models.TransferRecordRepository
	clusterRepo         models.ClusterRepository

	gridClient gridclient.GridClient
	graphql    graphql.GraphQl

	minimumTFTAmountInWallet uint64
	appliedDiscount          Discount
}

func NewBillingService(userRepo models.UserRepository, contractsRepo models.ContractDataRepository,
	transferRecordsRepo models.TransferRecordRepository, clusterRepo models.ClusterRepository,
	graphql graphql.GraphQl, gridClient gridclient.GridClient,
	minimumTFTAmountInWallet uint64, appliedDiscount Discount,
) BillingService {
	return BillingService{
		userRepo:            userRepo,
		contractsRepo:       contractsRepo,
		transferRecordsRepo: transferRecordsRepo,
		clusterRepo:         clusterRepo,

		gridClient: gridClient,
		graphql:    graphql,

		appliedDiscount:          appliedDiscount,
		minimumTFTAmountInWallet: minimumTFTAmountInWallet,
	}
}

func (svc *BillingService) Discount() Discount {
	return svc.appliedDiscount
}

func (svc *BillingService) SettleUserUsage(user *models.User) error {
	usageInUSDMillicent, err := svc.getUserLatestUsageInUSD(user.ID)
	if err != nil {
		return err
	}

	return svc.userRepo.DeductUserBalance(user, usageInUSDMillicent)
}

func (svc *BillingService) AfterUserGetCredit(ctx context.Context, user *models.User) error {
	if err := svc.FundUserWithTFTs(ctx, user); err != nil {
		return err
	}

	return svc.SettleUserUsage(user)
}

func (svc *BillingService) FundUserWithTFTs(ctx context.Context, user *models.User) error {
	userGridAccountTFTBalance, err := svc.gridClient.GetFreeBalanceTFT(user.Mnemonic)
	if err != nil {
		return err
	}

	totalPendingTFTAmount, err := svc.transferRecordsRepo.CalculateTotalPendingTFTAmountPerUser(user.ID)
	if err != nil {
		return err
	}

	userBalanceInTFT, err := svc.gridClient.FromUSDMillicentToTFT(user.CreditedBalance + user.CreditCardBalance)
	if err != nil {
		return err
	}

	if userGridAccountTFTBalance+totalPendingTFTAmount >= userBalanceInTFT {
		return nil
	}

	return svc.transferRecordsRepo.CreateTransferRecord(&models.TransferRecord{
		UserID:    user.ID,
		Username:  user.Username,
		TFTAmount: (userBalanceInTFT - (userGridAccountTFTBalance + totalPendingTFTAmount)) * TFTUnitFactor,
		Operation: models.DepositOperation,
	})
}

func (svc *BillingService) CreateTransferRecordToChargeUserWithMinTFTAmount(userID int, username, userMnemonic string) error {
	userTFTBalance, err := svc.gridClient.GetFreeBalanceTFT(userMnemonic)
	if err != nil {
		return err
	}

	totalPendingTFTAmount, err := svc.transferRecordsRepo.CalculateTotalPendingTFTAmountPerUser(userID)
	if err != nil {
		return err
	}

	if userTFTBalance+totalPendingTFTAmount >= zeroTFTBalanceValue {
		return nil
	}

	return svc.transferRecordsRepo.CreateTransferRecord(&models.TransferRecord{
		UserID:    userID,
		Username:  username,
		TFTAmount: svc.minimumTFTAmountInWallet * TFTUnitFactor,
		Operation: models.DepositOperation,
	})
}

func (svc *BillingService) calculateResourcesUsageInUSD(
	ctx context.Context,
	userID int,
	userMnemonic string,
	addedRentedNodes []types.Node,
	addedSharedNodes []kubedeployer.Node,
) (uint64, error) {
	calculator, err := svc.gridClient.NewCalculator(userMnemonic)
	if err != nil {
		return 0, fmt.Errorf("failed to create calculator: %w", err)
	}

	var totalResourcesCostMillicent uint64

	rentedNodes, _, err := svc.getRentedNodesForUser(ctx, userID, true)
	if err != nil {
		return 0, err
	}
	if addedRentedNodes == nil {
		rentedNodes = append(rentedNodes, addedRentedNodes...)
	}

	// Calculate rented nodes
	for _, node := range rentedNodes {
		resourcesCost, err := calculator.CalculateCost(
			node.TotalResources.CRU,
			uint64(node.TotalResources.MRU),
			uint64(node.TotalResources.HRU),
			uint64(node.TotalResources.SRU),
			len(node.PublicConfig.Ipv4) > 0,
			node.CertificationType == nodeCertified,
		)
		if err != nil {
			return 0, err
		}

		// resources cost per month
		pricingPolicy, err := svc.gridClient.GetPricingPolicy(defaultPricingPolicyID)
		if err != nil {
			return 0, err
		}
		dedicatedDiscountPercentage := float64(pricingPolicy.DedicatedNodesDiscount / 100)
		totalResourcesCostMillicent += gridclient.FromUSDToUSDMillicent(resourcesCost * dedicatedDiscountPercentage)
	}

	sharedNodes, err := svc.getUserNodes(userID)
	if err != nil {
		return 0, err
	}
	if addedSharedNodes != nil {
		sharedNodes = append(sharedNodes, addedSharedNodes...)
	}

	// Calculate shared nodes
	for _, node := range sharedNodes {
		proxyNode, err := svc.gridClient.Node(ctx, node.NodeID)
		if err != nil {
			return 0, err
		}

		if proxyNode.Rented {
			twinID, err := svc.gridClient.GetTwin(userMnemonic)
			if err != nil {
				return 0, err
			}

			if proxyNode.RentedByTwinID == uint(twinID) {
				// skip rented nodes as they are already calculated
				continue
			}
		}

		// Calculate total disk size (sum all data disks + root size)
		totalDiskSize := node.RootSize
		for _, diskSize := range node.DataDisks {
			totalDiskSize += diskSize
		}

		resourcesCost, err := calculator.CalculateCost(
			uint64(node.CPU),
			node.Memory,
			0,
			totalDiskSize,
			false,
			proxyNode.CertificationType == nodeCertified,
		)
		if err != nil {
			return 0, err
		}

		// resources cost per month
		totalResourcesCostMillicent += gridclient.FromUSDToUSDMillicent(resourcesCost)
	}

	// Calculate name contracts
	nameContracts, err := svc.listNameContractsForUser(userID)
	if err != nil {
		return 0, err
	}

	nameContractMonthlyCostInUSD, err := svc.calculateUniqueNameMonthlyCost()
	if err != nil {
		return 0, err
	}

	totalResourcesCostMillicent += gridclient.FromUSDToUSDMillicent(float64(len(nameContracts)) * nameContractMonthlyCostInUSD)
	return totalResourcesCostMillicent, nil
}

func (svc *BillingService) ApplyDiscountOnUsage(totalResourcesCostMillicent uint64) (uint64, error) {
	discount := getDiscountPackage(svc.appliedDiscount).DurationInMonth
	if discount == 0 {
		return totalResourcesCostMillicent, nil
	}

	return uint64(float64(totalResourcesCostMillicent) * discount), nil
}

func (svc *BillingService) getUserNodes(userID int) ([]kubedeployer.Node, error) {
	userClusters, err := svc.clusterRepo.ListUserClusters(userID)
	if err != nil {
		return nil, err
	}

	var sharedNodes []kubedeployer.Node
	for _, cluster := range userClusters {
		clusterResult, err := cluster.GetClusterResult()
		if err != nil {
			return nil, err
		}
		sharedNodes = append(sharedNodes, clusterResult.Nodes...)
	}

	return sharedNodes, nil
}

func (svc *BillingService) calculateUniqueNameMonthlyCost() (float64, error) {
	pricingPolicy, err := svc.gridClient.GetPricingPolicy(defaultPricingPolicyID)
	if err != nil {
		return 0, err
	}

	// cost in unit-USD
	monthlyCost := float64(pricingPolicy.UniqueName.Value) * 24 * 30

	costInUSD := monthlyCost / TFTUnitFactor
	return costInUSD, nil
}

func (svc *BillingService) getRentedNodesForUser(ctx context.Context, userID int, healthy bool) ([]types.Node, int, error) {
	twinID, err := svc.getTwinIDFromUserID(userID)
	if err != nil {
		return nil, 0, err
	}

	filter := types.NodeFilter{
		RentedBy: &twinID,
		Features: Zos3NodeFeatures,
	}

	if healthy {
		filter.Healthy = &healthy
	}

	limit := types.DefaultLimit()

	nodes, count, err := svc.gridClient.Nodes(ctx, filter, limit)
	if err != nil {
		return nil, 0, err
	}

	return nodes, count, nil
}

func (svc *BillingService) listNameContractsForUser(userID int) ([]graphql.Contract, error) {
	twinID, err := svc.getTwinIDFromUserID(userID)
	if err != nil {
		return nil, err
	}

	contractGetter := svc.gridClient.NewContractsGetter(
		uint32(twinID),
		svc.graphql,
	)

	contractsList, err := contractGetter.ListContractsByTwinID([]string{"Created, GracePeriod"})
	if err != nil {
		return nil, err
	}

	return contractsList.NameContracts, nil
}

func (svc *BillingService) getTwinIDFromUserID(userID int) (uint64, error) {
	user, err := svc.userRepo.GetUserByID(userID)
	if err != nil {
		return 0, err
	}

	twinID, err := svc.gridClient.GetTwin(user.Mnemonic)
	if err != nil {
		return 0, err
	}

	return uint64(twinID), nil
}

func (svc *BillingService) getUserLatestUsageInUSD(userID int) (uint64, error) {
	now := time.Now()
	// Define the end of the day (next day at 00:00)
	endOfDay := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, time.Local)

	// Get the last calculation time for this user from the database, or use a default if not available
	lastCalcTime, err := svc.userRepo.GetUserLastCalcTime(userID)
	if err != nil {
		return 0, err
	}

	// If this is the first time or no record exists, use the start of the day as default
	if lastCalcTime.IsZero() {
		lastCalcTime = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
	}

	contracts, err := svc.contractsRepo.ListAllContractsInPeriod(userID, lastCalcTime, endOfDay)
	if err != nil {
		return 0, err
	}

	if len(contracts) == 0 {
		return 0, nil
	}

	var totalDailyUsageInUSDMillicent uint64

	for _, record := range contracts {
		// Get bill reports from the last calculation time to the end of day
		billReports, err := billing.ListContractBillReports(svc.graphql, record.ContractID, lastCalcTime, endOfDay)
		if err != nil {
			return 0, err
		}

		totalAmountBilledInUSDMillicent, err := svc.calculateTotalUsageOfReportsInUSDMillicent(billReports.Reports)
		if err != nil {
			return 0, err
		}

		totalDailyUsageInUSDMillicent += totalAmountBilledInUSDMillicent
	}

	// Update the last calculation time for this user in the database
	if err := svc.userRepo.UpdateUserLastCalcTime(userID, now); err != nil {
		logger.GetLogger().Error().Err(err).Msgf("Failed to update last calculation time for user %d", userID)
	}

	return totalDailyUsageInUSDMillicent, nil
}

func (svc *BillingService) calculateTotalUsageOfReportsInUSDMillicent(reports []billing.Report) (uint64, error) {
	var totalAmountBilledInUSDMillicent uint64
	for _, report := range reports {
		amountInTFT, err := removeDiscountFromReport(&report)
		if err != nil {
			return 0, err
		}

		amountInUSDMillicent, err := svc.fromTFTtoUSDMillicent(amountInTFT, report)
		if err != nil {
			return 0, err
		}

		totalAmountBilledInUSDMillicent += amountInUSDMillicent
	}

	return totalAmountBilledInUSDMillicent, nil
}

func (svc *BillingService) fromTFTtoUSDMillicent(amount uint64, report billing.Report) (uint64, error) {
	price, err := svc.getBillingRateAt(report)
	if err != nil {
		return 0, err
	}

	usdMillicentBalance := uint64(math.Round((float64(amount) / TFTUnitFactor) * float64(price)))
	return usdMillicentBalance, nil
}

func (svc *BillingService) getBillingRateAt(report billing.Report) (float64, error) {
	block_duration := 6 // in seconds
	now := time.Now().Unix()

	reportTimestamp, err := strconv.ParseInt(report.Timestamp, 10, 64)
	if err != nil {
		return 0, err
	}

	timeBetweenNowAndReport := now - reportTimestamp // seconds

	// Calculate number of blocks since report
	numberOfBlocks := math.Round(float64(timeBetweenNowAndReport) / float64(block_duration))

	nowBlock, err := svc.gridClient.GetCurrentHeight()
	if err != nil {
		return 0, err
	}
	reportBlock := nowBlock - uint32(numberOfBlocks)

	tftPrice, err := svc.gridClient.GetTFTBillingRateAt(uint64(reportBlock))
	if err != nil {
		return 0, err
	}

	return float64(tftPrice), nil
}

func removeDiscountFromReport(report *billing.Report) (uint64, error) {
	discountPackage := getDiscountPackage(Discount(report.DiscountReceived))

	amountBilled, err := strconv.ParseInt(report.AmountBilled, 10, 64)
	if err != nil {
		return 0, err
	}

	amountBilledWithNoDiscount := float64(amountBilled) / float64(1-discountPackage.Discount/100)
	return uint64(amountBilledWithNoDiscount), nil
}

func getDiscountPackage(discountInput Discount) DiscountPackage {
	oneDayMargin := 1.0 / 30.0

	discountPackages := map[Discount]DiscountPackage{
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
