package moneycollector

import (
	"kubecloud/internal"
	"kubecloud/internal/substrate"
	"kubecloud/internal/models"
	"sync"

	"github.com/rs/zerolog/log"
)

type MoneyCollector struct {
	userRepo        models.UserRepository
	config          internal.Configuration
	substrateClient substrate.Substrate
}

const (
	MinBalanceThreshold = 1e5
)

func NewMoneyCollector(userRepo models.UserRepository, config internal.Configuration, substrateClient substrate.Substrate) *MoneyCollector {
	return &MoneyCollector{
		userRepo:        userRepo,
		config:          config,
		substrateClient: substrateClient,
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

			balance, err := m.substrateClient.GetUserTFTBalance(user.Mnemonic)
			if err != nil {
				log.Error().Err(err).Int("user_id", user.ID).Msg("MoneyCollector: failed to get user balance")
				return
			}
			if balance > MinBalanceThreshold {
				log.Debug().Int("user_id", user.ID).Uint64("balance", balance).Msg("MoneyCollector: transferring balance to system account")
				if err := m.substrateClient.TransferTFTsToSystem(balance-MinBalanceThreshold, user.Mnemonic); err != nil {
					log.Error().Err(err).Int("user_id", user.ID).Msg("MoneyCollector: failed to transfer balance")
				}
				return
			}
		}(user)

	}
	wg.Wait()
	log.Info().Msg("MoneyCollector: finished")
}
