package main

import (
	"context"
	"flag"

	"kubecloud/internal/core/models"
	"kubecloud/internal/infrastructure/persistence"

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

	srcDB, err := persistence.NewGormDB(sourceDSN, models.DBPoolConfig{})
	if err != nil {
		log.Error().Err(err).Msg("failed to open source db")
		return
	}
	defer srcDB.Close()

	dstDB, err := persistence.NewGormDB(destinationDSN, models.DBPoolConfig{})
	if err != nil {
		log.Error().Err(err).Msg("failed to open destination db")
		return
	}
	defer dstDB.Close()

	log.Info().Msg("migrating database")
	if err := persistence.MigrateAll(context.Background(), srcDB, dstDB); err != nil {
		log.Error().Err(err).Msg("migration failed")
		return
	}
}
