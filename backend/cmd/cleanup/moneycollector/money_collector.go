package moneycollector

import (
	cfg "kubecloud/internal/config"
	"kubecloud/internal/core/models"
	"sync"

	"github.com/rs/zerolog/log"
	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/deployer"
)

type MoneyCollector struct {
	userRepo   models.UserRepository
	config     cfg.Configuration
	gridClient deployer.TFPluginClient
}

const (
	MinBalanceThreshold = 1e5
)

func NewMoneyCollector(userRepo models.UserRepository, config cfg.Configuration, gridClient deployer.TFPluginClient) *MoneyCollector {
	return &MoneyCollector{
		userRepo:   userRepo,
		config:     config,
		gridClient: gridClient,
	}
}

func (m *MoneyCollector) CollectMoney() {
	users, err := m.userRepo.ListAllUsers()
	if err != nil {
		log.Error().Err(err).Msg("MoneyCollector: failed to list all users")
		return
	}
	log.Debug().Int("total_users", len(users)).Msg("MoneyCollector: total users")
	maxConcurrentBalanceFetches := m.config.MailSender.MaxConcurrentSends

	var wg sync.WaitGroup
	balanceConcurrencyLimiter := make(chan struct{}, maxConcurrentBalanceFetches)
	for _, user := range users {
		wg.Add(1)
		go func(user models.User) {
			balanceConcurrencyLimiter <- struct{}{}
			defer wg.Done()
			defer func() { <-balanceConcurrencyLimiter }()
			if user.Mnemonic == "" {
				return
			}

			userIdentity, err := m.gridClient.SubstrateConn.NewIdentityFromSr25519Phrase(user.Mnemonic)
			if err != nil {
				log.Error().Err(err).Int("user_id", user.ID).Msg("MoneyCollector: failed to get user identity")
				return
			}

			balance, err := m.gridClient.SubstrateConn.GetBalance(userIdentity)
			if err != nil {
				log.Error().Err(err).Int("user_id", user.ID).Msg("MoneyCollector: failed to get user balance")
				return
			}
			freeBalance := balance.Free.Uint64()
			if freeBalance > MinBalanceThreshold {
				log.Debug().Int("user_id", user.ID).Uint64("balance", freeBalance).Msg("MoneyCollector: transferring balance to system account")
				if err := m.gridClient.SubstrateConn.TransferTFTsToSystem(freeBalance-MinBalanceThreshold, user.Mnemonic, m.config.SystemAccount.Mnemonic); err != nil {
					log.Error().Err(err).Int("user_id", user.ID).Msg("MoneyCollector: failed to transfer balance")
				}
				return
			}
		}(user)

	}
	wg.Wait()
	log.Info().Msg("MoneyCollector: finished")
}
