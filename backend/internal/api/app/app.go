package app

import (
	"context"
	"fmt"
	"kubecloud/internal/api/middlewares"
	"kubecloud/internal/billing"
	"kubecloud/internal/core/workers"
	"kubecloud/internal/core/workflows"
	"kubecloud/internal/infrastructure/metrics"
	"kubecloud/internal/shared"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"kubecloud/internal/infrastructure/logger"

	"github.com/gin-gonic/gin"
	"github.com/stripe/stripe-go/v82"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	// Import the generated docs package
	_ "kubecloud/docs/swagger"
)

// App holds all configurations for the app
type App struct {
	router     *gin.Engine
	httpServer *http.Server
	config     shared.Configuration

	workers  workers.Workers
	handlers appHandlers

	*appDependencies
}

// NewApp create new instance of the app with all configs
func NewApp(ctx context.Context, config shared.Configuration) (*App, error) {
	// Disable gin's default logging since we're using zerolog
	gin.DisableConsoleColor()
	gin.SetMode(gin.ReleaseMode)

	// Create router without default middleware
	router := gin.New()

	// Add recovery middleware
	router.Use(gin.Recovery())

	// Add our custom logging middleware (includes request ID generation)
	router.Use(middlewares.GinLoggerMiddleware())

	stripe.Key = config.StripeSecret

	appDependencies, err := createAppDependencies(ctx, config)
	if err != nil {
		return nil, err
	}

	app := &App{
		router:          router,
		config:          config,
		appDependencies: &appDependencies,
	}

	app.handlers = app.createHandlers()
	app.workers = app.createWorkers()

	app.registerEWFWorkflows()
	app.registerHandlers()

	return app, nil
}

func (app *App) registerEWFWorkflows() {
	stripeClient := &billing.DefaultStripeClient{}
	workflows.RegisterEWFWorkflows(
		app.core.ewfEngine,
		app.config,
		app.core.db,
		app.communication.mailService,
		app.infra.substrateClient,
		app.security.kycClient,
		app.security.sponsorAddress,
		app.security.sponsorKeyPair,
		app.core.metrics,
		app.communication.notificationSender,
		app.infra.gridClient.GridProxyClient,
		stripeClient,
	)
}

// registerHandlers registers all routes
func (app *App) registerHandlers() {
	app.core.metrics.RegisterMetricsEndpoint(app.router)

	app.router.Use(middlewares.CorsMiddleware())
	app.router.Use(app.core.metrics.Middleware())

	app.core.metrics.StartGORMMetricsCollector(app.core.db, metrics.MetricsCollectorInterval)
	app.core.metrics.StartGoRuntimeMetricsCollector(metrics.MetricsCollectorInterval)

	v1 := app.router.Group("/api/v1")
	{
		v1.GET("/health", app.handlers.healthHandler.HealthHandler)
		v1.GET("/workflow/:workflow_id", app.handlers.userHandler.GetWorkflowStatus)
		v1.GET("/system/maintenance/status", app.handlers.settingsHandler.GetMaintenanceModeHandler)
		v1.GET("/stats", app.handlers.statsHandler.GetStatsHandler)
		v1.GET("/twins/:twin_id/account", app.handlers.nodeHandler.GetAccountIDHandler)
		v1.GET("/nodes", app.handlers.nodeHandler.ListAllGridNodesHandler)
		v1.GET("/nodes/:node_id/storage-pool", app.handlers.nodeHandler.GetNodeStoragePoolHandler)

		adminGroup := v1.Group("")
		adminGroup.Use(middlewares.AdminMiddleware(app.security.tokenManager))
		{
			usersGroup := adminGroup.Group("/users")
			{
				usersGroup.GET("", app.handlers.adminHandler.ListUsersHandler)
				usersGroup.DELETE("/:user_id", app.handlers.adminHandler.DeleteUsersHandler)
				usersGroup.POST("/:user_id/credit", app.handlers.adminHandler.CreditUserHandler)
			}
			usersGroup.POST("/mail", app.handlers.adminHandler.SendMailToAllUsersHandler)
			adminGroup.GET("/pending-records", app.handlers.adminHandler.ListPendingRecordsHandler)
			adminGroup.GET("/invoices", app.handlers.invoiceHandler.ListAllInvoicesHandler)

			vouchersGroup := adminGroup.Group("/vouchers")
			{
				vouchersGroup.POST("/generate", app.handlers.adminHandler.GenerateVouchersHandler)
				vouchersGroup.GET("", app.handlers.adminHandler.ListVouchersHandler)

			}

		}

		systemGroup := adminGroup.Group("/system")
		{
			systemGroup.PUT("/maintenance/status", app.handlers.settingsHandler.SetMaintenanceModeHandler)
		}

		userGroup := v1.Group("/user")
		{
			userGroup.POST("/register", app.handlers.userHandler.RegisterHandler)
			userGroup.POST("/register/verify", app.handlers.userHandler.VerifyRegisterCode)
			userGroup.POST("/login", app.handlers.userHandler.LoginUserHandler)
			userGroup.POST("/refresh", app.handlers.userHandler.RefreshTokenHandler)
			userGroup.POST("/forgot_password", app.handlers.userHandler.ForgotPasswordHandler)
			userGroup.POST("/forgot_password/verify", app.handlers.userHandler.VerifyForgetPasswordCodeHandler)

			authGroup := userGroup.Group("")
			authGroup.Use(middlewares.UserMiddleware(app.security.tokenManager))
			{
				authGroup.GET("/", app.handlers.userHandler.GetUserHandler)
				authGroup.POST("/balance/charge", app.handlers.userHandler.ChargeBalance)
				authGroup.PUT("/change_password", app.handlers.userHandler.ChangePasswordHandler)
				authGroup.GET("/balance", app.handlers.userHandler.GetUserBalance)
				authGroup.PUT("/redeem/:voucher_code", app.handlers.userHandler.RedeemVoucherHandler)
				authGroup.GET("/pending-records", app.handlers.userHandler.ListUserPendingRecordsHandler)

				authGroup.GET("/nodes", app.handlers.nodeHandler.ListNodesHandler)
				authGroup.GET("/nodes/rentable", app.handlers.nodeHandler.ListRentableNodesHandler)
				authGroup.GET("/nodes/rented", app.handlers.nodeHandler.ListRentedNodesHandler)
				authGroup.POST("/nodes/:node_id", app.handlers.nodeHandler.ReserveNodeHandler)
				authGroup.DELETE("/nodes/unreserve/:contract_id", app.handlers.nodeHandler.UnreserveNodeHandler)

				authGroup.GET("/invoice/:invoice_id", app.handlers.invoiceHandler.DownloadInvoiceHandler)
				authGroup.GET("/invoice", app.handlers.invoiceHandler.ListUserInvoicesHandler)
				// SSH Key management
				authGroup.GET("/ssh-keys", app.handlers.userHandler.ListSSHKeysHandler)
				authGroup.POST("/ssh-keys", app.handlers.userHandler.AddSSHKeyHandler)
				authGroup.DELETE("/ssh-keys/:ssh_key_id", app.handlers.userHandler.DeleteSSHKeyHandler)
			}
		}

		deployerGroup := v1.Group("")
		deployerGroup.Use(middlewares.UserMiddleware(app.security.tokenManager))
		{
			deployerGroup.GET("/events", app.communication.sseManager.HandleSSE)

			deploymentGroup := deployerGroup.Group("/deployments")
			{
				deploymentGroup.POST("", app.handlers.deploymentHandler.HandleDeployCluster)
				deploymentGroup.GET("", app.handlers.deploymentHandler.HandleListDeployments)
				deploymentGroup.DELETE("", app.handlers.deploymentHandler.HandleDeleteAllDeployments)
				deploymentGroup.GET("/:name", app.handlers.deploymentHandler.HandleGetDeployment)
				deploymentGroup.GET("/:name/kubeconfig", app.handlers.deploymentHandler.HandleGetKubeconfig)
				deploymentGroup.DELETE("/:name", app.handlers.deploymentHandler.HandleDeleteCluster)
				deploymentGroup.POST("/:name/nodes", app.handlers.deploymentHandler.HandleAddNode)
				deploymentGroup.DELETE("/:name/nodes/:node_name", app.handlers.deploymentHandler.HandleRemoveNode)
			}

			notificationGroup := deployerGroup.Group("/notifications")
			{
				notificationGroup.GET("", app.handlers.notificationHandler.GetAllNotificationsHandler)
				notificationGroup.GET("/unread", app.handlers.notificationHandler.GetUnreadNotificationsHandler)
				notificationGroup.PATCH("/read-all", app.handlers.notificationHandler.MarkAllNotificationsReadHandler)
				notificationGroup.DELETE("", app.handlers.notificationHandler.DeleteAllNotificationsHandler)
				notificationGroup.PATCH("/:notification_id/read", app.handlers.notificationHandler.MarkNotificationReadHandler)
				notificationGroup.PATCH("/:notification_id/unread", app.handlers.notificationHandler.MarkNotificationUnreadHandler)
				notificationGroup.DELETE("/:notification_id", app.handlers.notificationHandler.DeleteNotificationHandler)
			}
		}
	}
	app.router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}

func (app *App) StartBackgroundWorkers() {
	go app.workers.MonthlyInvoicesHandler()
	go app.workers.TrackUserDebt()
	go app.workers.MonitorSystemBalanceAndHandleSettlement()
	go app.workers.TrackClusterHealth()
	go app.workers.TrackReservedNodeHealth()
}

// Run starts the server
func (app *App) Run() error {
	app.StartBackgroundWorkers()

	app.core.ewfEngine.ResumeRunningWorkflows()
	app.httpServer = &http.Server{
		Addr:    fmt.Sprintf(":%s", app.config.Server.Port),
		Handler: app.router,
	}

	logger.GetLogger().Info().
		Str("host", app.config.Server.Host).
		Str("port", app.config.Server.Port).
		Msg("Starting server")

	if err := app.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.GetLogger().Error().Err(err).Msg("Failed to start server")
		return err
	}

	return nil
}

// Shutdown gracefully shuts down the server and worker manager
func (app *App) Shutdown() error {
	defer app.core.appCtx.Done()

	if app.httpServer != nil {
		if err := app.httpServer.Shutdown(app.core.appCtx); err != nil {
			logger.GetLogger().Error().Err(err).Msg("Failed to shutdown HTTP server")
		}
	}

	if app.communication.sseManager != nil {
		app.communication.sseManager.Stop()
	}

	if app.core.db != nil {
		if err := app.core.db.Close(); err != nil {
			logger.GetLogger().Error().Err(err).Msg("Failed to close database connection")
		}
	}

	app.infra.gridClient.Close()

	logger.CloseLogger()

	return nil
}

//nolint:unused
func (app *App) startCommandSocket() {
	socketPath := "/tmp/myceliumcloud.sock"

	os.Remove(socketPath)

	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		logger.GetLogger().Error().Err(err).Msg("failed to create command socket")
		return
	}
	defer listener.Close()
	defer os.Remove(socketPath)

	logger.GetLogger().Info().Str("socket", socketPath).Msg("Command socket started")

	for {
		select {
		case <-app.core.appCtx.Done():
			logger.GetLogger().Info().Msg("command socket stopping")
			return
		default:
		}

		if unixListener, ok := listener.(*net.UnixListener); ok {
			if err := unixListener.SetDeadline(time.Now().Add(1 * time.Second)); err != nil {
				logger.GetLogger().Error().Err(err).Msg("failed to set deadline on listener")
			}
		}

		conn, err := listener.Accept()

		if err == nil {
			go app.handleSocketCommand(conn)
			continue
		}

		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			continue
		}

		if app.core.appCtx.Err() != nil {
			return
		}

		logger.GetLogger().Error().Err(err).Msg("socket accept error")
		continue
	}
}

//nolint:unused
func (app *App) handleSocketCommand(conn net.Conn) {
	defer conn.Close()

	if err := conn.SetReadDeadline(time.Now().Add(5 * time.Second)); err != nil {
		logger.GetLogger().Error().Err(err).Msg("failed to set read deadline")
		return
	}

	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		if _, writeErr := conn.Write([]byte("ERROR: Failed to read command\n")); writeErr != nil {
			logger.GetLogger().Error().Err(writeErr).Msg("failed to write error response")
		}
		return
	}

	command := strings.TrimSpace(string(buffer[:n]))
	logger.GetLogger().Debug().Str("command", command).Msg("Received socket command")

	if command == "reload-notifications" {
		app.handleReloadNotifications(conn)
		return
	}

	response := fmt.Sprintf("ERROR: Unknown command '%s'\n", command)
	if _, err := conn.Write([]byte(response)); err != nil {
		logger.GetLogger().Error().Err(err).Msg("failed to write error response")
	}
	logger.GetLogger().Warn().Str("command", command).Msg("Unknown socket command received")
}

//nolint:unused
func (app *App) handleReloadNotifications(conn net.Conn) {
	err := app.reloadNotificationConfig()

	if err != nil {
		response := fmt.Sprintf("ERROR: %v\n", err)
		if _, writeErr := conn.Write([]byte(response)); writeErr != nil {
			logger.GetLogger().Error().Err(writeErr).Msg("failed to write error response")
		}
		logger.GetLogger().Error().Err(err).Msg("Failed to reload notification config via socket")
		return
	}

	if _, err := conn.Write([]byte("OK: Notification config reloaded successfully\n")); err != nil {
		logger.GetLogger().Error().Err(err).Msg("failed to write success response")
		return
	}
	logger.GetLogger().Info().Msg("Notification config reloaded via socket")
}

//nolint:unused
func (app *App) reloadNotificationConfig() error {
	cfg, err := shared.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if err = app.communication.notificationSender.ReloadNotificationConfig(cfg.Notification); err != nil {
		return fmt.Errorf("failed to reload notification config: %w", err)
	}

	return nil
}
