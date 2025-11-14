package main

import (
	"flag"
	"kubecloud/internal/core/models"
	"kubecloud/internal/infrastructure/substrate"
	"kubecloud/internal/shared"
	"os"

	moneycollector "kubecloud/cmd/cleanup/moneycollector"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	// substrate "github.com/threefoldtech/tfchain/clients/tfchain-client-go"
)

var config shared.Configuration

func loadConfig(configPath string) {
	viper.SetConfigFile(configPath)
	if err := viper.ReadInConfig(); err != nil {
		log.Error().Err(err).Msg("Failed to read config file")
		os.Exit(1)
	}

	config = shared.Configuration{
		Database: shared.DB{
			DSN:                    viper.GetString("database.dsn"),
			MaxOpenConns:           viper.GetInt("database.max_open_conns"),
			MaxIdleConns:           viper.GetInt("database.max_idle_conns"),
			ConnMaxLifetimeMinutes: viper.GetInt("database.conn_max_lifetime_minutes"),
			ConnMaxIdleTimeMinutes: viper.GetInt("database.conn_max_idle_time_minutes"),
		},
		SystemAccount: shared.GridAccount{
			Mnemonic: viper.GetString("system_account.mnemonic"),
			Network:  viper.GetString("system_account.network"),
		},
		MailSender: shared.MailSender{
			MaxConcurrentSends: viper.GetInt("mailSender.max_concurrent_sends"),
		},
	}
	if config.MailSender.MaxConcurrentSends == 0 {
		config.MailSender.MaxConcurrentSends = 10
	}
}

func main() {
	configPath := flag.String("config", "../../config.json", "Path to config file")
	flag.Parse()
	loadConfig(*configPath)
	dbPoolConfig := models.DBPoolConfig{
		MaxOpenConns:           config.Database.MaxOpenConns,
		MaxIdleConns:           config.Database.MaxIdleConns,
		ConnMaxLifetimeMinutes: config.Database.ConnMaxLifetimeMinutes,
		ConnMaxIdleTimeMinutes: config.Database.ConnMaxIdleTimeMinutes,
	}

	db, err := models.NewGormDB(config.Database.DSN, dbPoolConfig)

	if err != nil {
		log.Error().Err(err).Msg("Failed to open database")
		return
	}
	defer db.Close()

	substrateClient, err := substrate.NewTFChainClient(
		config.SystemAccount.Network, config.SystemAccount.Mnemonic, "", "",
	)

	if err != nil {
		log.Error().Err(err).Msg("Failed to create TF chain client")
		return
	}

	defer substrateClient.Close()

	moneyCollector := moneycollector.NewMoneyCollector(models.NewGormUserRepository(db), config, substrateClient)
	moneyCollector.CollectMoney()
}
