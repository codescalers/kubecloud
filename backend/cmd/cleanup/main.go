package main

import (
	"flag"
	cfg "kubecloud/internal/config"
	"kubecloud/internal/core/models"
	corepersistence "kubecloud/internal/core/persistence"
	"kubecloud/internal/infrastructure/persistence"
	"kubecloud/internal/infrastructure/substrate"

	"os"

	moneycollector "kubecloud/cmd/cleanup/moneycollector"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/deployer"
	// substrate "github.com/threefoldtech/tfchain/clients/tfchain-client-go"
)

var config cfg.Configuration

func loadConfig(configPath string) {
	viper.SetConfigFile(configPath)
	if err := viper.ReadInConfig(); err != nil {
		log.Error().Err(err).Msg("Failed to read config file")
		os.Exit(1)
	}

	config = cfg.Configuration{
		Database: cfg.DB{
			DSN:                    viper.GetString("database.dsn"),
			MaxOpenConns:           viper.GetInt("database.max_open_conns"),
			MaxIdleConns:           viper.GetInt("database.max_idle_conns"),
			ConnMaxLifetimeMinutes: viper.GetInt("database.conn_max_lifetime_minutes"),
			ConnMaxIdleTimeMinutes: viper.GetInt("database.conn_max_idle_time_minutes"),
		},
		SystemAccount: cfg.GridAccount{
			Mnemonic: viper.GetString("system_account.mnemonic"),
			Network:  viper.GetString("system_account.network"),
		},
		MailSender: cfg.MailSender{
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

	db, err := persistence.NewGormDB(config.Database.DSN, dbPoolConfig)

	if err != nil {
		log.Error().Err(err).Msg("Failed to open database")
		return
	}
	defer db.Close()

	gridClient, err := deployer.NewTFPluginClient(
		config.SystemAccount.Mnemonic, deployer.WithNetwork(config.SystemAccount.Network),
	)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create TF plugin client")
		return
	}
	substrateClient := substrate.NewSubstrateClient(config.SystemAccount.Mnemonic, gridClient)

	defer gridClient.Close()

	moneyCollector := moneycollector.NewMoneyCollector(corepersistence.NewGormUserRepository(db), config, substrateClient)
	moneyCollector.CollectMoney()
}
