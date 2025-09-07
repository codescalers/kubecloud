package moneycollector

import (
	"kubecloud/internal"
	"kubecloud/models"
	"sync"

	"github.com/rs/zerolog/log"
	substrate "github.com/threefoldtech/tfchain/clients/tfchain-client-go"
)

type MoneyCollector struct {
	db              models.DB
	config          internal.Configuration
	substrateClient *substrate.Substrate
}

const (
	MinBalanceThreshold = 1e5
)

func NewMoneyCollector(db models.DB, config internal.Configuration, substrateClient *substrate.Substrate) *MoneyCollector {
	return &MoneyCollector{
		db:              db,
		config:          config,
		substrateClient: substrateClient,
	}
}

func (m *MoneyCollector) CollectMoney() {
	system, err := substrate.NewIdentityFromSr25519Phrase(m.config.SystemAccount.Mnemonic)
	if err != nil {
		log.Error().Err(err).Msg("MoneyCollector: failed to load system identity")
		return
	}
	users, err := m.db.ListAllUsers()
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
			balance, err := internal.GetUserTFTBalance(m.substrateClient, user.Mnemonic)
			if err != nil {
				log.Error().Err(err).Int("user_id", user.ID).Msg("MoneyCollector: failed to get user balance")
				return
			}
			if balance > MinBalanceThreshold {
				userIdentity, err := substrate.NewIdentityFromSr25519Phrase(user.Mnemonic)
				if err != nil {
					log.Error().Err(err).Int("user_id", user.ID).Msg("MoneyCollector: failed to load user identity")
					return
				}
				log.Debug().Int("user_id", user.ID).Uint64("balance", balance).Msg("MoneyCollector: transferring balance to system account")
				if err := m.substrateClient.Transfer(userIdentity, balance-MinBalanceThreshold, substrate.AccountID(system.PublicKey())); err != nil {
					log.Error().Err(err).Int("user_id", user.ID).Msg("MoneyCollector: failed to transfer balance")
				}
				return
			}
		}(user)

	}
	wg.Wait()
	log.Info().Msg("MoneyCollector: finished")
}
