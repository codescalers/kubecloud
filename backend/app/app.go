package app

import (
	"context"
	"fmt"
	"kubecloud/internal"
	"kubecloud/internal/activities"
	"kubecloud/internal/metrics"
	"kubecloud/middlewares"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"kubecloud/internal/logger"

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
	config     internal.Configuration

	*appDependencies
	*handlers
}

// NewApp create new instance of the app with all configs
func NewApp(ctx context.Context, config internal.Configuration) (*App, error) {
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

	handlers := app.createHandlers()
	app.handlers = &handlers

	app.registerEWFWorkflows()
	app.registerHandlers()

	return app, nil
}

func (app *App) registerEWFWorkflows() {
	activities.RegisterEWFWorkflows(
		app.core.ewfEngine,
		app.config,
		app.core.db,
		app.communication.mailService,
		app.infra.substrateClient,
		app.security.kycClient,
		app.security.sponsorAddress,
		app.security.sponsorKeyPair,
		app.core.metrics,
		app.communication.notificationService,
		app.infra.gridClient.GridProxyClient,
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
		v1.GET("/health", app.healthHandler.HealthHandler)
		v1.GET("/workflow/:workflow_id", app.userHandler.GetWorkflowStatus)
		v1.GET("/twins/:twin_id/account", app.nodeHandler.GetAccountIDHandler)
		v1.GET("/system/maintenance/status", app.adminHandler.GetMaintenanceModeHandler)
		v1.GET("/stats", app.statsHandler.GetStatsHandler)
		v1.GET("/nodes", app.nodeHandler.ListAllGridNodesHandler)
		v1.GET("/nodes/:node_id/storage-pool", app.nodeHandler.GetNodeStoragePoolHandler)

		adminGroup := v1.Group("")
		adminGroup.Use(middlewares.AdminMiddleware(app.security.tokenManager))
		{
			usersGroup := adminGroup.Group("/users")
			{
				usersGroup.GET("", app.adminHandler.ListUsersHandler)
				usersGroup.DELETE("/:user_id", app.adminHandler.DeleteUsersHandler)
				usersGroup.POST("/:user_id/credit", app.adminHandler.CreditUserHandler)
			}
			usersGroup.POST("/mail", app.adminHandler.SendMailToAllUsersHandler)
			adminGroup.GET("/pending-records", app.adminHandler.ListPendingRecordsHandler)
			adminGroup.GET("/invoices", app.invoiceHandler.ListAllInvoicesHandler)

			vouchersGroup := adminGroup.Group("/vouchers")
			{
				vouchersGroup.POST("/generate", app.adminHandler.GenerateVouchersHandler)
				vouchersGroup.GET("", app.adminHandler.ListVouchersHandler)

			}

		}

		systemGroup := adminGroup.Group("/system")
		{
			systemGroup.PUT("/maintenance/status", app.adminHandler.SetMaintenanceModeHandler)
		}

		userGroup := v1.Group("/user")
		{
			userGroup.POST("/register", app.userHandler.RegisterHandler)
			userGroup.POST("/register/verify", app.userHandler.VerifyRegisterCode)
			userGroup.POST("/login", app.userHandler.LoginUserHandler)
			userGroup.POST("/refresh", app.userHandler.RefreshTokenHandler)
			userGroup.POST("/forgot_password", app.userHandler.ForgotPasswordHandler)
			userGroup.POST("/forgot_password/verify", app.userHandler.VerifyForgetPasswordCodeHandler)

			authGroup := userGroup.Group("")
			authGroup.Use(middlewares.UserMiddleware(app.security.tokenManager))
			{
				authGroup.GET("/", app.userHandler.GetUserHandler)
				authGroup.POST("/balance/charge", app.userHandler.ChargeBalance)
				authGroup.PUT("/change_password", app.userHandler.ChangePasswordHandler)
				authGroup.GET("/balance", app.userHandler.GetUserBalance)
				authGroup.PUT("/redeem/:voucher_code", app.userHandler.RedeemVoucherHandler)
				authGroup.GET("/pending-records", app.userHandler.ListUserPendingRecordsHandler)

				authGroup.GET("/nodes", app.nodeHandler.ListNodesHandler)
				authGroup.GET("/nodes/rentable", app.nodeHandler.ListRentableNodesHandler)
				authGroup.GET("/nodes/rented", app.nodeHandler.ListRentedNodesHandler)
				authGroup.POST("/nodes/:node_id", app.nodeHandler.ReserveNodeHandler)
				authGroup.DELETE("/nodes/unreserve/:contract_id", app.nodeHandler.UnreserveNodeHandler)

				authGroup.GET("/invoice/:invoice_id", app.invoiceHandler.DownloadInvoiceHandler)
				authGroup.GET("/invoice", app.invoiceHandler.ListUserInvoicesHandler)
				// SSH Key management
				authGroup.GET("/ssh-keys", app.userHandler.ListSSHKeysHandler)
				authGroup.POST("/ssh-keys", app.userHandler.AddSSHKeyHandler)
				authGroup.DELETE("/ssh-keys/:ssh_key_id", app.userHandler.DeleteSSHKeyHandler)
			}
		}

		deployerGroup := v1.Group("")
		deployerGroup.Use(middlewares.UserMiddleware(app.security.tokenManager))
		{
			deployerGroup.GET("/events", app.communication.sseManager.HandleSSE)

			deploymentGroup := deployerGroup.Group("/deployments")
			{
				deploymentGroup.POST("", app.deploymentHandler.HandleDeployCluster)
				deploymentGroup.GET("", app.deploymentHandler.HandleListDeployments)
				deploymentGroup.DELETE("", app.deploymentHandler.HandleDeleteAllDeployments)
				deploymentGroup.GET("/:name", app.deploymentHandler.HandleGetDeployment)
				deploymentGroup.GET("/:name/kubeconfig", app.deploymentHandler.HandleGetKubeconfig)
				deploymentGroup.DELETE("/:name", app.deploymentHandler.HandleDeleteCluster)
				deploymentGroup.POST("/:name/nodes", app.deploymentHandler.HandleAddNode)
				deploymentGroup.DELETE("/:name/nodes/:node_name", app.deploymentHandler.HandleRemoveNode)
			}

			notificationGroup := deployerGroup.Group("/notifications")
			{
				notificationGroup.GET("", app.notificationHandler.GetAllNotificationsHandler)
				notificationGroup.GET("/unread", app.notificationHandler.GetUnreadNotificationsHandler)
				notificationGroup.PATCH("/read-all", app.notificationHandler.MarkAllNotificationsReadHandler)
				notificationGroup.DELETE("", app.notificationHandler.DeleteAllNotificationsHandler)
				notificationGroup.PATCH("/:notification_id/read", app.notificationHandler.MarkNotificationReadHandler)
				notificationGroup.PATCH("/:notification_id/unread", app.notificationHandler.MarkNotificationUnreadHandler)
				notificationGroup.DELETE("/:notification_id", app.notificationHandler.DeleteNotificationHandler)
			}
		}
	}
	app.router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}

func (app *App) StartBackgroundWorkers() {
	go app.invoiceHandler.MonthlyInvoicesHandler(app.core.appCtx)
	go app.adminHandler.TrackUserDebt(app.core.appCtx, app.infra.gridClient)
	go app.adminHandler.MonitorSystemBalanceAndHandleSettlement(app.core.appCtx)
	go app.deploymentHandler.TrackClusterHealth(app.core.appCtx)
	go app.nodeHandler.TrackReservedNodeHealth(app.core.appCtx, app.communication.notificationService, app.infra.gridClient.GridProxyClient)
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
	cfg, err := internal.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if err = app.communication.notificationService.ReloadNotificationConfig(cfg.Notification); err != nil {
		return fmt.Errorf("failed to reload notification config: %w", err)
	}

	return nil
}
