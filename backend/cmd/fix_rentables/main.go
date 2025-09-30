package main

import (
	"flag"
	"fmt"
	"kubecloud/models"
	"os"
	"strings"

	"github.com/rs/zerolog/log"
	substrate "github.com/threefoldtech/tfchain/clients/tfchain-client-go"
)

func main() {
	var dbPath string
	var tfchainURL string
	flag.StringVar(&dbPath, "db", "", "Path to SQLite database file")
	flag.StringVar(&tfchainURL, "tfchain-url", "", "TFChain WebSocket/HTTP URL")
	flag.Parse()

	if strings.TrimSpace(dbPath) == "" || strings.TrimSpace(tfchainURL) == "" {
		log.Error().Msg("Both --db and --tfchain-url flags are required")
		return
	}

	_, err := os.Stat(dbPath)
	if os.IsNotExist(err) {
		log.Error().Err(err).Msg("Database file does not exist")
		return
	}
	if err != nil {
		log.Error().Err(err).Msg("Error checking database file")
		return
	}

	db, err := models.NewSqliteDBNoMigrate(dbPath)
	if err != nil {
		log.Error().Err(err).Msg("Failed to open database")
		return
	}
	defer db.Close()

	if err := db.GetDB().Exec("DROP INDEX IF EXISTS idx_user_node_id").Error; err != nil {
		log.Error().Err(err).Msg("Failed to drop idx_user_node_id index")
		return
	}

	substrateClient, err := substrate.NewManager(tfchainURL).Substrate()
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
