package main

import (
	"flag"
	"fmt"
	"kubecloud/models"
	"strings"

	"github.com/rs/zerolog/log"
	substrate "github.com/threefoldtech/tfchain/clients/tfchain-client-go"
)

func main() {
	var dsn string
	var tfchainURL string
	var maxOpen int
	var maxIdle int
	var maxLife int
	var maxIdleTime int
	flag.StringVar(&dsn, "dsn", "", "Database DSN (postgres://... or sqlite:///path.db)")
	flag.StringVar(&tfchainURL, "tfchain-url", "", "TFChain WebSocket/HTTP URL")
	flag.IntVar(&maxOpen, "db-max-open-conns", 0, "DB max open connections (postgres only)")
	flag.IntVar(&maxIdle, "db-max-idle-conns", 0, "DB max idle connections (postgres only)")
	flag.IntVar(&maxLife, "db-conn-max-lifetime", 0, "DB connection max lifetime (e.g. 30m) (postgres only)")
	flag.IntVar(&maxIdleTime, "db-conn-max-idle-time", 0, "DB connection max idle time (e.g. 5m) (postgres only)")
	flag.Parse()

	if strings.TrimSpace(dsn) == "" || strings.TrimSpace(tfchainURL) == "" {
		log.Error().Msg("Both --dsn and --tfchain-url flags are required")
		return
	}

	pool := models.DBPoolConfig{MaxOpenConns: maxOpen, MaxIdleConns: maxIdle, ConnMaxLifetimeMinutes: maxLife, ConnMaxIdleTimeMinutes: maxIdleTime}
	db, err := models.NewDB(dsn, pool)
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
