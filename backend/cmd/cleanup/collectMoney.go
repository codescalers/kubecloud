package cleanup

import (
	"kubecloud/internal"
	"kubecloud/models"

	"github.com/rs/zerolog/log"
	substrate "github.com/threefoldtech/tfchain/clients/tfchain-client-go"
)

type MoneyCollector struct {
	db              models.DB
	config          internal.Configuration
	substrateClient *substrate.Substrate
}

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
		log.Error().Err(err).Msg("MoneyCollector: failed to create system identity")
		return
	}
	users, err := m.db.ListAllUsers()
	if err != nil {
		log.Error().Err(err).Msg("MoneyCollector: failed to list all users")
		return
	}

	for _, user := range users {
		if user.Mnemonic == "" {
			continue
		}
		balance, err := internal.GetUserBalanceUSDMillicent(m.substrateClient, user.Mnemonic)
		if err != nil {
			log.Error().Err(err).Int("user_id", user.ID).Msg("MoneyCollector: failed to get user balance")
			continue
		}
		if balance > 0 {
			m.substrateClient.Transfer(substrate.Identity{mnemonic: user.Mnemonic}, uint64(balance), substrate.AccountID(m.config.SystemAccount.Mnemonic))
		}

	}

}
