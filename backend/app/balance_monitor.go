package app

import (
	"context"
	"kubecloud/internal"
	"kubecloud/kubedeployer"
	"kubecloud/models"
	"time"

	"kubecloud/internal/logger"

	"github.com/pkg/errors"
	substrate "github.com/threefoldtech/tfchain/clients/tfchain-client-go"
	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/calculator"
	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/graphql"
	"github.com/threefoldtech/tfgrid-sdk-go/grid-proxy/pkg/types"
)

const (
	// UnitFactor represents the smallest unit conversion factor for both USD and TFT
	TFTUnitFactor = 1e7
	transferFees  = 0.01 * TFTUnitFactor // 0.01 TFT
	nodeCertified = "Certified"

	zeroTFTBalanceValue    = 0.05 * TFTUnitFactor // 0.05 TFT
	defaultPricingPolicyID = uint32(1)
)

func (h *Handler) MonitorSystemBalanceAndHandleSettlement(ctx context.Context) {
	settleTransfersTicker := time.NewTicker(time.Duration(h.config.SettleTransferRecordsIntervalInMinutes) * time.Minute)
	adminNotifyTicker := time.NewTicker(time.Duration(h.config.NotifyAdminsForPendingRecordsInHours) * time.Hour)
	zeroUSDBalanceTicker := time.NewTicker(time.Minute)
	zeroTFTBalanceTicker := time.NewTicker(time.Minute)
	fundUserTFTBalanceTicker := time.NewTicker(24 * time.Hour)
	defer settleTransfersTicker.Stop()
	defer adminNotifyTicker.Stop()
	defer zeroUSDBalanceTicker.Stop()
	defer zeroTFTBalanceTicker.Stop()
	defer fundUserTFTBalanceTicker.Stop()

	for {
		users, err := h.db.ListAllUsers()
		if err != nil {
			continue
		}

		select {
		case <-settleTransfersTicker.C:
			records, err := h.db.ListPendingTransferRecords()
			if err != nil {
				continue
			}

			failedRecords, err := h.db.ListFailedTransferRecords()
			if err != nil {
				continue
			}

			records = append(records, failedRecords...)

			if err := h.settlePendingPayments(records); err != nil {
				logger.GetLogger().Error().Err(err).Send()
			}

		case <-adminNotifyTicker.C:
			records, err := h.db.ListPendingTransferRecords()
			if err != nil {
				continue
			}

			if len(records) > 0 {
				if err := h.notifyAdminWithPendingRecords(records); err != nil {
					logger.GetLogger().Error().Err(err).Send()
				}
			}

		case <-zeroUSDBalanceTicker.C:
			if err := h.resetUsersTFTsWithNoUSDBalance(users); err != nil {
				logger.GetLogger().Error().Err(err).Send()
			}

		case <-zeroTFTBalanceTicker.C:
			for _, user := range users {
				clusters, err := h.db.ListUserClusters(user.ID)
				if err != nil {
					logger.GetLogger().Error().Err(err).Msgf("Failed to list user clusters")
					continue
				}

				if len(clusters) > 0 {
					// user has deployed workloads, skip
					continue
				}

				zeroUSDMillicentBalanceValue, err := internal.FromTFTtoUSDMillicent(h.substrateClient, zeroTFTBalanceValue)
				if err != nil {
					logger.GetLogger().Error().Err(err).Msgf("failed to convert TFT to USD millicent")
					continue
				}

				if user.CreditedBalance+user.CreditCardBalance-user.Debt > zeroUSDMillicentBalanceValue {
					continue
				}

				if err := h.db.CreateTransferRecord(&models.TransferRecord{
					UserID:    user.ID,
					Username:  user.Username,
					TFTAmount: uint64(h.config.MinimumTFTAmountInWallet) * TFTUnitFactor,
					Operation: models.DepositOperation,
				}); err != nil {
					logger.GetLogger().Error().Err(err).Msgf("Failed to create transfer record for user %d", user.ID)
				}
			}

		case <-fundUserTFTBalanceTicker.C:
			for _, user := range users {
				if err := h.fundUserToFulfillDiscount(ctx, user, []types.Node{}, []kubedeployer.Node{}, discount(h.config.AppliedDiscount)); err != nil {
					logger.GetLogger().Error().Err(err).Msgf("Failed to fund user %d to claim discount", user.ID)
				}
			}
		}
	}
}

func (h *Handler) resetUsersTFTsWithNoUSDBalance(users []models.User) error {
	for _, user := range users {
		if user.CreditedBalance+user.CreditCardBalance-user.Debt <= 0 {
			logger.GetLogger().Info().Msgf("User %d has no USD balance, withdrawing all TFTs except for %d", user.ID, h.config.MinimumTFTAmountInWallet)

			userTFTBalance, err := internal.GetUserTFTBalance(h.substrateClient, user.Mnemonic)
			if err != nil {
				logger.GetLogger().Error().Err(err).Msgf("Failed to get user TFT balance for user %d", user.ID)
				continue
			}

			if userTFTBalance <= uint64(h.config.MinimumTFTAmountInWallet)*TFTUnitFactor {
				continue
			}

			if userTFTBalance <= transferFees {
				continue
			}

			transferRecord := models.TransferRecord{
				UserID:    user.ID,
				Username:  user.Username,
				TFTAmount: userTFTBalance,
				Operation: models.WithdrawOperation,
				State:     models.SuccessState,
			}

			if err = h.withdrawTFTsFromUser(user.ID, user.Mnemonic, userTFTBalance); err != nil {
				logger.GetLogger().Error().Err(err).Msgf("Failed to withdraw all TFTs for user %d", user.ID)

				// TODO: handle retries
				transferRecord.State = models.FailedState
				transferRecord.Failure = err.Error()
			}

			if err := h.db.CreateTransferRecord(&transferRecord); err != nil {
				logger.GetLogger().Error().Err(err).Msgf("Failed to create transfer record for user %d", user.ID)
			}
		}
	}

	return nil
}

func (h *Handler) settlePendingPayments(records []models.TransferRecord) error {
	for _, record := range records {
		if record.Operation == models.WithdrawOperation {
			continue
		}

		// Already settled
		if record.State == models.SuccessState {
			continue
		}

		transferState := models.SuccessState
		var transferFailure string

		// getting balance every time to ensure we have the latest balance
		systemTFTBalance, err := internal.GetUserTFTBalance(h.substrateClient, h.config.SystemAccount.Mnemonic)
		if err != nil {
			logger.GetLogger().Error().Err(err).Msgf("Failed to get system TFT balance for pending record ID %d", record.ID)
			continue
		}

		if systemTFTBalance < record.TFTAmount {
			logger.GetLogger().Warn().Msgf("Insufficient system balance to settle pending record ID %d", record.ID)
			continue
		}

		if err = h.transferTFTsToUser(record.UserID, record.TFTAmount); err != nil {
			logger.GetLogger().Error().Err(err).Msgf("Failed to settle pending record ID %d", record.ID)

			transferState = models.FailedState
			transferFailure = err.Error()
		}

		if err := h.db.UpdateTransferRecordState(record.ID, transferState, transferFailure); err != nil {
			logger.GetLogger().Error().Err(err).Msgf("Failed to update pending record ID %d state", record.ID)
		}
	}

	return nil
}

func (h *Handler) transferTFTsToUser(userID int, amountToTransfer uint64) error {
	user, err := h.db.GetUserByID(userID)
	if err != nil {
		return errors.Wrapf(err, "failed to get user %d", userID)
	}

	err = internal.TransferTFTs(h.substrateClient, amountToTransfer, user.Mnemonic, h.systemIdentity)
	if err != nil {
		return errors.Wrapf(err, "Failed to transfer TFTs to user %d", userID)
	}

	return nil
}

func (h *Handler) withdrawTFTsFromUser(userID int, userMnemonic string, amountToWithdraw uint64) error {
	userIdentity, err := substrate.NewIdentityFromSr25519Phrase(userMnemonic)
	if err != nil {
		return errors.Wrapf(err, "Failed to create identity for user %d", userID)
	}

	err = internal.TransferTFTs(h.substrateClient, amountToWithdraw, h.config.SystemAccount.Mnemonic, userIdentity)
	if err != nil {
		return errors.Wrapf(err, "Failed to transfer TFTs to user %d", userID)
	}

	return nil
}

func (h *Handler) fundUserToFulfillDiscount(ctx context.Context, user models.User, addedRentedNodes []types.Node, addedSharedNodes []kubedeployer.Node, discount discount) error {
	// calculate resources usage in USD applying discount
	// I took the cluster nodes since only the new node is in cluster.Nodes
	dailyUsageInUSDMillicent, err := h.calculateResourcesUsageInUSDApplyingDiscount(ctx, user.ID, user.Mnemonic, addedRentedNodes, addedSharedNodes, discount)
	if err != nil {
		return err
	}

	dailyUsageInTFT, err := internal.FromUSDMillicentToTFT(h.substrateClient, dailyUsageInUSDMillicent)
	if err != nil {
		return err
	}

	totalPendingTFTAmount, err := h.db.CalculateTotalPendingTFTAmountPerUser(user.ID)
	if err != nil {
		return err
	}

	userTFTBalance, err := internal.GetUserTFTBalance(h.substrateClient, user.Mnemonic)
	if err != nil {
		return err
	}

	// fund user to fulfill discount
	// make sure no old payments will fund more than needed
	if totalPendingTFTAmount+userTFTBalance < dailyUsageInTFT &&
		dailyUsageInTFT > 0 {
		if err := h.db.CreateTransferRecord(&models.TransferRecord{
			UserID:    user.ID,
			Username:  user.Username,
			TFTAmount: dailyUsageInTFT - userTFTBalance - totalPendingTFTAmount,
			Operation: models.DepositOperation,
		}); err != nil {
			return err
		}
	}

	return nil
}

func (h *Handler) calculateResourcesUsageInUSDApplyingDiscount(
	ctx context.Context,
	userID int,
	userMnemonic string,
	addedRentedNodes []types.Node,
	addedSharedNodes []kubedeployer.Node,
	configuredDiscount discount,
) (uint64, error) {
	userIdentity, err := substrate.NewIdentityFromSr25519Phrase(userMnemonic)
	if err != nil {
		return 0, errors.Wrapf(err, "Failed to create identity for user %d", userID)
	}

	calculator := calculator.NewCalculator(h.gridClient.SubstrateConn, userIdentity)

	var totalResourcesCostMillicent uint64

	rentedNodes, _, err := h.getRentedNodesForUser(ctx, userID, true)
	if err != nil {
		return 0, err
	}
	rentedNodes = append(rentedNodes, addedRentedNodes...)

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
		pricingPolicy, err := h.substrateClient.GetPricingPolicy(defaultPricingPolicyID)
		if err != nil {
			return 0, err
		}
		dedicatedDiscountPercentage := float64(pricingPolicy.DedicatedNodesDiscount / 100)
		totalResourcesCostMillicent += internal.FromUSDToUSDMillicent(resourcesCost * dedicatedDiscountPercentage)
	}

	sharedNodes, err := h.getUserNodes(userID)
	if err != nil {
		return 0, err
	}
	sharedNodes = append(sharedNodes, addedSharedNodes...)

	// Calculate shared nodes
	for _, node := range sharedNodes {
		proxyNode, err := h.proxyClient.Node(ctx, node.NodeID)
		if err != nil {
			return 0, err
		}

		if proxyNode.Rented {
			twinID, err := h.substrateClient.GetTwinByPubKey(userIdentity.PublicKey())
			if err != nil {
				return 0, err
			}

			if proxyNode.RentedByTwinID == uint(twinID) {
				// skip rented nodes as they are already calculated
				continue
			}
		}

		resourcesCost, err := calculator.CalculateCost(
			uint64(node.CPU),
			node.Memory,
			0,
			node.DiskSize+node.RootSize,
			false,
			proxyNode.CertificationType == nodeCertified,
		)
		if err != nil {
			return 0, err
		}

		// resources cost per month
		totalResourcesCostMillicent += internal.FromUSDToUSDMillicent(resourcesCost)
	}

	// Calculate name contracts
	nameContracts, err := h.listNameContractsForUser(userIdentity)
	if err != nil {
		return 0, err
	}

	nameContractMonthlyCostInUSD, err := h.calculateUniqueNameMonthlyCost()
	if err != nil {
		return 0, err
	}

	totalResourcesCostMillicent += internal.FromUSDToUSDMillicent(float64(len(nameContracts)) * nameContractMonthlyCostInUSD)

	discount := getDiscountPackage(configuredDiscount).DurationInMonth
	if discount == 0 {
		return totalResourcesCostMillicent, nil
	}

	return uint64(float64(totalResourcesCostMillicent) * discount), nil
}

func (h *Handler) notifyAdminWithPendingRecords(records []models.TransferRecord) error {
	subject, body := h.mailService.NotifyAdminsMailContent(len(records), h.config.Server.Host)

	admins, err := h.db.ListAdmins()
	if err != nil {
		return err
	}

	for _, admin := range admins {
		err = h.mailService.SendMail(h.config.MailSender.Email, admin.Email, subject, body)
		if err != nil {
			logger.GetLogger().Error().Err(err).Send()
			continue
		}
	}

	return nil
}

func (h *Handler) listNameContractsForUser(userIdentity substrate.Identity) ([]graphql.Contract, error) {
	graphQl, err := graphql.NewGraphQl(h.config.GraphqlURL)
	if err != nil {
		return nil, errors.Wrapf(err, "could not create a new graphql with url: %s", h.config.GraphqlURL)
	}

	twinID, err := h.substrateClient.GetTwinByPubKey(userIdentity.PublicKey())
	if err != nil {
		return nil, err
	}

	contractGetter := graphql.NewContractsGetter(
		twinID,
		graphQl,
		h.gridClient.SubstrateConn,
		h.gridClient.NcPool,
	)

	contractsList, err := contractGetter.ListContractsByTwinID([]string{"Created, GracePeriod"})
	if err != nil {
		return nil, err
	}

	return contractsList.NameContracts, nil
}

func (h *Handler) calculateUniqueNameMonthlyCost() (float64, error) {
	pricingPolicy, err := h.substrateClient.GetPricingPolicy(defaultPricingPolicyID)
	if err != nil {
		return 0, err
	}

	// cost in unit-USD
	monthlyCost := float64(pricingPolicy.UniqueName.Value) * 24 * 30

	costInUSD := monthlyCost / TFTUnitFactor
	return costInUSD, nil
}

func (h *Handler) getUserNodes(userID int) ([]kubedeployer.Node, error) {
	userClusters, err := h.db.ListUserClusters(userID)
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
