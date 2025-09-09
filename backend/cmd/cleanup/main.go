package main

import (
	"flag"
	"kubecloud/internal"
	"kubecloud/models"
	"os"

	moneycollector "kubecloud/cmd/cleanup/moneycollector"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	substrate "github.com/threefoldtech/tfchain/clients/tfchain-client-go"
)

var config internal.Configuration

func loadConfig(configPath string) {
	viper.SetConfigFile(configPath)
	if err := viper.ReadInConfig(); err != nil {
		log.Error().Err(err).Msg("Failed to read config file")
		os.Exit(1)
	}

	config = internal.Configuration{
		Database: internal.DB{
			File: viper.GetString("database.file"),
		},
		TFChainURL: viper.GetString("tfchain_url"),
		SystemAccount: internal.GridAccount{
			Mnemonic: viper.GetString("system_account.mnemonic"),
			Network:  viper.GetString("system_account.network"),
		},
		MailSender: internal.MailSender{
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

	_, err := os.Stat(config.Database.File)
	if os.IsNotExist(err) {
		log.Error().Err(err).Msg("Database file does not exist")
		return
	}
	if err != nil {
		log.Error().Err(err).Msg("Error checking database file")
		return
	}

	db, err := models.NewSqliteDB(config.Database.File)

	if err != nil {
		log.Error().Err(err).Msg("Failed to open database")
		return
	}
	defer db.Close()

	substrateClient, err := substrate.NewManager(config.TFChainURL).Substrate()
	if err != nil {
		log.Error().Err(err).Msg("Failed to create substrate client")
		return
	}
	defer substrateClient.Close()

	moneyCollector := moneycollector.NewMoneyCollector(db, config, substrateClient)
	moneyCollector.CollectMoney()

}
