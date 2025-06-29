package cmd

import (
	"context"
	"fmt"
	"kubecloud/app"
	"kubecloud/internal"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// TODO: add descriptions
var rootCmd = &cobra.Command{
	Use:   "KubeCloud",
	Short: "This is short description!",
	Long:  "This is long description!",
	RunE: func(cmd *cobra.Command, args []string) error {
		configFile, err := cmd.Flags().GetString("config")
		if err != nil {
			return fmt.Errorf("failed to parse config: %w", err)
		}

		config, err := internal.ReadConfFile(configFile)
		if err != nil {
			log.Error().Err(err).Msg("Failed to read configurations file")
			return fmt.Errorf("failed to read configuration file: %w", err)
		}

		app, err := app.NewApp(config)
		if err != nil {
			return fmt.Errorf("failed to create new app: %w", err)
		}

		return gracefulShutdown(app)
	},
}

func gracefulShutdown(app *app.App) error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Info().Msg("Starting KubeCloud server")

		if err := app.Run(); err != nil && err != http.ErrServerClosed {
			log.Error().Err(err).Msg("Failed to start server")
			stop()
		}
	}()

	<-ctx.Done()
	log.Info().Msg("Shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := app.Shutdown(shutdownCtx); err != nil {
		log.Error().Err(err).Msg("Server shutdown failed")
		return err
	}

	log.Info().Msg("Server gracefully stopped.")
	return nil
}

func Execute() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	err := rootCmd.Execute()
	if err != nil {
		log.Error().Err(err).Msg("Command execution failed")
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().StringP("config", "c", "./config.json", "Path to the configuration file (default: ./config.json)")
}
