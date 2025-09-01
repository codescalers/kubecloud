package main

import (
	"flag"
	"kubecloud/internal"
	"kubecloud/models"
	"os"

	moneyCollector "kubecloud/cmd/cleanup/moneyCollector"

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
	db, err := models.NewSqliteDB(config.Database.File)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create database")
		return
	}

	substrateClient, err := substrate.NewManager(config.TFChainURL).Substrate()
	if err != nil {
		log.Error().Err(err).Msg("Failed to create substrate client")
		return
	}

	moneyCollector := moneyCollector.NewMoneyCollector(db, config, substrateClient)
	moneyCollector.CollectMoney()

	os.Exit(0)

}
