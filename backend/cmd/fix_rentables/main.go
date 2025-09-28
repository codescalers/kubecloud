package main

import (
	"flag"
	"fmt"
	"kubecloud/internal"
	"kubecloud/models"
	"os"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	substrate "github.com/threefoldtech/tfchain/clients/tfchain-client-go"
)

var config internal.Configuration

func loadConfig(configPath string) error {
	viper.SetConfigFile(configPath)
	if err := viper.ReadInConfig(); err != nil {
		log.Error().Err(err).Msg("Failed to read config file")
		return err
	}

	config = internal.Configuration{
		Database: internal.DB{
			File: viper.GetString("database.file"),
		},
		TFChainURL: viper.GetString("tfchain_url"),
	}
	return nil
}

func main() {
	configPath := flag.String("config", "../../config.json", "Path to config file")
	flag.Parse()
	err := loadConfig(*configPath)
	if err != nil {
		log.Error().Err(err).Msg("Failed to load config")
		return
	}

	_, err = os.Stat(config.Database.File)
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

	// Get all user_nodes records
	var allRecords []models.UserNodes
	if err := db.GetDB().Order("created_at DESC, id DESC").Find(&allRecords).Error; err != nil {
		log.Error().Err(err).Msg("Failed to get user_nodes records")
		return
	}

	if len(allRecords) == 0 {
		fmt.Println("No user_nodes records found. Nothing to fix.")
		return
	}

	fmt.Printf("Found %d total user_nodes records to check.\n", len(allRecords))

	validContracts := 0
	invalidContracts := 0
	removedTotal := 0

	for _, record := range allRecords {
		contract, err := substrateClient.GetContract(record.ContractID)
		if err != nil && !strings.Contains(err.Error(), "not found") {
			log.Error().Err(err).Uint64("contract_id", record.ContractID).Msg("Failed to get contract")
			continue
		}
		shouldDelete := err != nil || contract.State.IsDeleted

		if shouldDelete {
			// Contract not found or deleted - delete the record
			invalidContracts++
			reason := "not found"
			if err == nil {
				reason = "deleted"
			}
			fmt.Printf("Node %d: Contract %d %s, deleting record\n", record.NodeID, record.ContractID, reason)

			if err := db.GetDB().Delete(&record).Error; err != nil {
				log.Error().Err(err).Uint64("contract_id", record.ContractID).Msg("Failed to delete record")
				continue
			}
			removedTotal++
		} else {
			// Contract exists and is active - keep the record
			validContracts++
			fmt.Printf("Node %d: Contract %d is valid, keeping record\n", record.NodeID, record.ContractID)
		}
	}

	fmt.Printf("Cleanup completed. Valid contracts: %d, Invalid contracts: %d, Removed: %d duplicate rows.\n", validContracts, invalidContracts, removedTotal)
}
