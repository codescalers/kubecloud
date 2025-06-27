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

	mailService := internal.NewMailService(config.MailSender.SendGridKey)

	handler := NewHandler(tokenHandler, db, config, mailService)

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
		usersGroup := v1.Group("/user")
		{
			usersGroup.POST("/register", app.handlers.RegisterHandler)
			usersGroup.POST("/register/verify", app.handlers.VerifyRegisterCode)
			usersGroup.POST("/login", app.handlers.LoginUserHandler)
			usersGroup.POST("/refresh", app.handlers.RefreshTokenHandler)
			usersGroup.POST("/forgot_password", app.handlers.ForgotPasswordHandler)
			usersGroup.POST("/forgot_password/verify", app.handlers.VerifyForgetPasswordCodeHandler)

			authGroup := usersGroup.Group("")
			authGroup.Use(middlewares.UserMiddleware(app.handlers.tokenManager))
			{
				authGroup.POST("/change_password", app.handlers.ChangePasswordHandler)
				authGroup.POST("/charge_balance", app.handlers.ChargeBalance)
			}

			adminGroup := usersGroup.Group("")
			adminGroup.Use(middlewares.AdminMiddleware(app.handlers.tokenManager))
			{

				adminGroup.GET("", app.handlers.ListUsersHandler)
				adminGroup.DELETE("/:user_id", app.handlers.DeleteUsersHandler)
				adminGroup.POST("/:user_id/credit", app.handlers.CreditUserHandler)

				vouchersGroup := adminGroup.Group("/vouchers")
				{
					vouchersGroup.POST("/generate", app.handlers.GenerateVouchersHandler)
					vouchersGroup.GET("", app.handlers.ListVouchersHandler)

				}

			}

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
