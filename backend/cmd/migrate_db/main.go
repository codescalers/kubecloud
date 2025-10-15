package main

import (
	"context"
	"flag"

	"kubecloud/models"

	"github.com/rs/zerolog/log"
)

func main() {
	var sourceDSN string
	var destinationDSN string
	flag.StringVar(&sourceDSN, "source-db", "", "Source database DSN (e.g., postgres://... or sqlite:///path.db)")
	flag.StringVar(&destinationDSN, "destination-db", "", "Destination database DSN (e.g., postgres://... or sqlite:///path.db)")
	flag.Parse()

	if sourceDSN == "" || destinationDSN == "" {
		log.Error().Msg("Both --source-db and --destination-db DSNs are required")
		return
	}

	srcDB, err := models.NewDB(sourceDSN, models.DBPoolConfig{})
	if err != nil {
		log.Error().Err(err).Msg("failed to open source db")
		return
	}
	defer srcDB.Close()

	dstDB, err := models.NewDB(destinationDSN, models.DBPoolConfig{})
	if err != nil {
		log.Error().Err(err).Msg("failed to open destination db")
		return
	}
	defer dstDB.Close()

	log.Info().Msg("migrating database")
	if err := models.MigrateAll(context.Background(), srcDB, dstDB); err != nil {
		log.Error().Err(err).Msg("migration failed")
		return
	}
}
