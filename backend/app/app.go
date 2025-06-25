package app

import (
	"context"
	"fmt"
	"kubecloud/internal"
	"kubecloud/middlewares"
	"kubecloud/models/sqlite"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"github.com/stripe/stripe-go/v82"
)

// App holds all configurations for the app
type App struct {
	router     *gin.Engine
	httpServer *http.Server
	config     internal.Configuration
	handlers   Handler
}

// NewApp create new instance of the app with all configs
func NewApp(config internal.Configuration) (*App, error) {
	router := gin.Default()

	stripe.Key = config.StripeSecret

	tokenHandler := internal.NewTokenHandler(
		config.JWT.Secret,
		time.Duration(config.JWT.AccessTokenExpiryMinutes)*time.Minute,
		time.Duration(config.JWT.RefreshTokenExpiryHours)*time.Hour,
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
	v1 := app.router.Group("/api/v1")
	{
		v1.POST("/register", app.handlers.RegisterHandler)
		v1.POST("/login", app.handlers.LoginUserHandler)
		v1.POST("/refresh", app.handlers.RefreshTokenHandler)
		v1.POST("/forgot_password", app.handlers.ForgotPasswordHandler)
		v1.POST("/forgot_password/verify_email", app.handlers.VerifyForgetPasswordCodeHandler)

		authGroup := v1.Group("")
		authGroup.Use(middlewares.UserMiddleware(app.handlers.tokenManager))
		{
			authGroup.POST("/change_password", app.handlers.ChangePasswordHandler)
			authGroup.POST("/charge_balance", app.handlers.ChargeBalance)
		}

	}

	adminGroup := v1.Group("/admin")
	adminGroup.Use(middlewares.AdminMiddleware(app.handlers.tokenManager))
	{

		usersGroup := adminGroup.Group("/users")
		{
			usersGroup.GET("", app.handlers.ListUsersHandler)
			usersGroup.POST("", app.handlers.RegisterHandler)
			usersGroup.DELETE("/:user_id", app.handlers.DeleteUsersHandler)
			usersGroup.POST("/:user_id/credit", app.handlers.CreditUserHandler)

		}

		vouchersGroup := adminGroup.Group("/vouchers")
		{
			vouchersGroup.POST("/generate", app.handlers.GenerateVouchersHandler)
			vouchersGroup.GET("", app.handlers.ListVouchersHandler)

		}

	}

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
