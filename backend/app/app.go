package app

import (
	"context"
	"fmt"
	"kubecloud/internal"
	"kubecloud/models/sqlite"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// App holds all configurations for the app
type App struct {
	router     *gin.Engine
	httpServer *http.Server
	config     internal.Configuration
	handlers   Handler
}

// NewApp create new instance of the app with all configs
func NewApp(configFile string) (*App, error) {
	router := gin.Default()

	config, err := internal.ReadConfFile(configFile)
	if err != nil {
		log.Error().Err(err).Msg("Failed to read configurations file")
		return nil, err
	}

	tokenHandler := internal.NewTokenHandler(
		config.JWT.Secret,
		time.Duration(config.JWT.AccessTokenExpiry)*time.Minute,
		time.Duration(config.JWT.RefreshTokenExpiry)*time.Hour,
	)

	db, err := sqlite.NewSqliteStorage(config.Database.File)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create user storage")
		return nil, fmt.Errorf("failed to create user storage: %w", err)
	}

	handler := NewHandler(tokenHandler, db, config)

	app := &App{
		router:   router,
		config:   config,
		handlers: *handler,
	}

	app.registerHandlers()

	return app, nil

}

// registerHandlers registers all routes
func (app *App) registerHandlers() {
	app.router.POST("/api/v1/register", app.handlers.RegisterHandler)
	app.router.POST("/api/v1/login", app.handlers.LoginUserHandler)
}

// Run starts the server
func (app *App) Run() error {
	addr := fmt.Sprintf("%s:%s", app.config.Server.Host, app.config.Server.Port)
	app.httpServer = &http.Server{
		Addr:    addr,
		Handler: app.router,
	}

	log.Info().Msgf("Starting server at http://%s", addr)

	if err := app.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Error().Err(err).Msg("Failed to start server")
	}

	return app.httpServer.ListenAndServe()
}

// Shutdown gracefully shuts down the server
func (app *App) Shutdown(ctx context.Context) error {
	if app.httpServer != nil {
		return app.httpServer.Shutdown(ctx)
	}
	return nil
}
