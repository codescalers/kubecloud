package app

import (
	"context"
	"fmt"
	"kubecloud/internal"
	"kubecloud/internal/activities"
	"kubecloud/internal/metrics"
	"kubecloud/internal/notification"
	"kubecloud/middlewares"
	"kubecloud/models"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	substrate "github.com/threefoldtech/tfchain/clients/tfchain-client-go"
	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/deployer"
	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/graphql"
	proxy "github.com/threefoldtech/tfgrid-sdk-go/grid-proxy/pkg/client"
	"github.com/xmonader/ewf"

	"kubecloud/internal/logger"

	"github.com/gin-gonic/gin"
	"github.com/stripe/stripe-go/v82"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	// Import the generated docs package
	_ "kubecloud/docs"
)

// App holds all configurations for the app
type App struct {
	router              *gin.Engine
	httpServer          *http.Server
	config              internal.Configuration
	handlers            Handler
	db                  models.DB
	redis               *internal.RedisClient
	sseManager          *internal.SSEManager
	notificationService *notification.NotificationService
	gridClient          deployer.TFPluginClient
	appCtx              context.Context
	appCancel           context.CancelFunc
	metrics             *metrics.Metrics
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

	// Add our custom logging middleware
	router.Use(middlewares.GinLoggerMiddleware())

	stripe.Key = config.StripeSecret

	tokenHandler := internal.NewTokenHandler(
		config.JwtToken.Secret,
		time.Duration(config.JwtToken.AccessExpiryMinutes)*time.Minute,
		time.Duration(config.JwtToken.RefreshExpiryHours)*time.Hour,
	)

	db, err := models.NewSqliteDB(config.Database.File)
	if err != nil {
		logger.GetLogger().Error().Err(err).Msg("Failed to create user storage")
		return nil, fmt.Errorf("failed to create user storage: %w", err)
	}

	gridProxy := proxy.NewRetryingClient(proxy.NewClient(config.GridProxyURL))

	manager := substrate.NewManager(config.TFChainURL)
	substrateClient, err := manager.Substrate()

	if err != nil {
		logger.GetLogger().Error().Err(err).Msg("failed to connect to substrate client")
		return nil, fmt.Errorf("failed to connect to substrate client: %w", err)
	}

	graphqlURL := []string{config.GraphqlURL}
	graphqlClient, err := graphql.NewGraphQl(graphqlURL...)
	if err != nil {
		logger.GetLogger().Error().Err(err).Msg("failed to connect to graphql client")
		return nil, fmt.Errorf("failed to connect to graphql client: %w", err)
	}

	firesquidURL := []string{config.FiresquidURL}
	firesquidClient, err := graphql.NewGraphQl(firesquidURL...)
	if err != nil {
		logger.GetLogger().Error().Err(err).Msg("failed to connect to firesquid client")
		return nil, fmt.Errorf("failed to connect to firesquid client: %w", err)
	}

	redisClient, err := internal.NewRedisClient(config.Redis)
	if err != nil {
		logger.GetLogger().Error().Err(err).Msg("Failed to create Redis client")
		return nil, fmt.Errorf("failed to create Redis client: %w", err)
	}

	sseManager := internal.NewSSEManager()
	pluginOpts := []deployer.PluginOpt{
		deployer.WithNetwork(config.SystemAccount.Network),
		deployer.WithDisableSentry(),
	}
	if config.Debug {
		pluginOpts = append(pluginOpts, deployer.WithLogs())
	}

	gridClient, err := deployer.NewTFPluginClient(
		config.SystemAccount.Mnemonic,
		pluginOpts...,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create TF grid client: %w", err)
	}

	// create storage for workflows
	ewfStore := models.NewGormStore(db.GetDB())

	// initialize workflow ewfEngine
	ewfEngine, err := ewf.NewEngine(ewfStore)
	if err != nil {
		logger.GetLogger().Error().Err(err).Msg("failed to init EWF engine")
		return nil, fmt.Errorf("failed to init workflow engine: %w", err)
	}

	metrics := metrics.NewMetrics()
	notificationConfig := config.Notification
	mailService := internal.NewMailService(config.MailSender.SendGridKey, metrics)

	sseNotifier := notification.NewSSENotifier(sseManager)
	emailNotifier := notification.NewEmailNotifier(mailService, config.MailSender.Email, notificationConfig.EmailTemplatesDirPath)
	err = emailNotifier.ParseTemplates()
	if err != nil {
		return nil, fmt.Errorf("failed to init notification templates: %w", err)
	}
	notificationService, err := notification.NewNotificationService(db, ewfEngine, notificationConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create notification service: %w", err)
	}
	notificationService.RegisterNotifier(sseNotifier)
	notificationService.RegisterNotifier(emailNotifier)
	if err := notificationService.ValidateConfigsChannelsAgainstRegistered(); err != nil {
		return nil, fmt.Errorf("failed to validate notification configs channels against registered notifiers: %w", err)
	}

	// Create an app-level context for coordinating shutdown
	systemIdentity, err := substrate.NewIdentityFromSr25519Phrase(config.SystemAccount.Mnemonic)
	if err != nil {
		return nil, fmt.Errorf("failed to create system identity: %w", err)
	}

	sshPublicKeyBytes, err := os.ReadFile(config.SSH.PublicKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read SSH public key from %s: %w", config.SSH.PublicKeyPath, err)
	}
	sshPublicKey := strings.TrimSpace(string(sshPublicKeyBytes))

	appCtx, appCancel := context.WithCancel(ctx)

	// Derive sponsor (system) account SS58 address once
	sponsorKeyPair, err := internal.KeyPairFromMnemonic(config.SystemAccount.Mnemonic)
	if err != nil {
		appCancel()
		return nil, fmt.Errorf("failed to create sponsor keypair from system account: %w", err)
	}
	sponsorAddress, err := internal.AccountAddressFromKeypair(sponsorKeyPair)
	if err != nil {
		appCancel()
		return nil, fmt.Errorf("failed to create sponsor address from keypair: %w", err)
	}

	// Validate KYC configuration
	if strings.TrimSpace(config.KYCVerifierAPIURL) == "" {
		appCancel()
		return nil, fmt.Errorf("KYC verifier API URL is required")
	}
	if strings.TrimSpace(config.KYCChallengeDomain) == "" {
		appCancel()
		return nil, fmt.Errorf("KYC challenge domain is required")
	}

	// Initialize KYC client
	kycClient := internal.NewKYCClient(
		config.KYCVerifierAPIURL,
		config.KYCChallengeDomain,
		nil, // Use default http.Client
	)
	if valid, err := kycClient.IsValidSponsor(appCtx, sponsorAddress); err != nil || !valid {
		appCancel()
		if err != nil {
			return nil, fmt.Errorf("failed to validate sponsor address, %w", err)
		}
		return nil, fmt.Errorf("the provided sponsor address can't be used as a sponsor")
	}

	handler := NewHandler(tokenHandler, db, config, mailService, gridProxy,
		substrateClient, graphqlClient, firesquidClient, redisClient,
		sseManager, ewfEngine, config.SystemAccount.Network, sshPublicKey,
		systemIdentity, kycClient, sponsorKeyPair, sponsorAddress, metrics, notificationService)

	app := &App{
		router:              router,
		config:              config,
		handlers:            *handler,
		redis:               redisClient,
		db:                  db,
		sseManager:          sseManager,
		notificationService: notificationService,
		appCtx:              appCtx,
		appCancel:           appCancel,
		gridClient:          gridClient,
		metrics:             metrics,
	}

	activities.RegisterEWFWorkflows(
		ewfEngine,
		app.config,
		app.db,
		app.handlers.mailService,
		app.handlers.substrateClient,
		app.handlers.kycClient,
		sponsorAddress,
		sponsorKeyPair,
		app.metrics,
		app.notificationService,
		gridProxy,
	)

	app.registerHandlers()

	return app, nil
}

// registerHandlers registers all routes
func (app *App) registerHandlers() {
	app.metrics.RegisterMetricsEndpoint(app.router)

	app.router.Use(middlewares.CorsMiddleware())
	app.router.Use(app.metrics.Middleware())

	app.metrics.StartGORMMetricsCollector(app.db.GetDB(), metrics.MetricsCollectorInterval)
	app.metrics.StartGoRuntimeMetricsCollector(metrics.MetricsCollectorInterval)

	v1 := app.router.Group("/api/v1")
	{
		v1.GET("/health", app.handlers.HealthHandler)
		v1.GET("/workflow/:workflow_id", app.handlers.GetWorkflowStatus)
		v1.GET("/twins/:twin_id/account", app.handlers.GetAccountIDHandler)
		v1.GET("/system/maintenance/status", app.handlers.GetMaintenanceModeHandler)
		v1.GET("/stats", app.handlers.GetStatsHandler)
		v1.GET("/nodes", app.handlers.ListAllGridNodesHandler)

		adminGroup := v1.Group("")
		adminGroup.Use(middlewares.AdminMiddleware(app.handlers.tokenManager))
		{
			usersGroup := adminGroup.Group("/users")
			{
				usersGroup.GET("", app.handlers.ListUsersHandler)
				usersGroup.DELETE("/:user_id", app.handlers.DeleteUsersHandler)
				usersGroup.POST("/:user_id/credit", app.handlers.CreditUserHandler)
			}
			usersGroup.POST("/mail", app.handlers.SendMailToAllUsersHandler)

			adminGroup.GET("/invoices", app.handlers.ListAllInvoicesHandler)
			adminGroup.GET("/pending-records", app.handlers.ListPendingRecordsHandler)

			vouchersGroup := adminGroup.Group("/vouchers")
			{
				vouchersGroup.POST("/generate", app.handlers.GenerateVouchersHandler)
				vouchersGroup.GET("", app.handlers.ListVouchersHandler)

			}

		}

		systemGroup := adminGroup.Group("/system")
		{
			systemGroup.PUT("/maintenance/status", app.handlers.SetMaintenanceModeHandler)
		}

		userGroup := v1.Group("/user")
		{
			userGroup.POST("/register", app.handlers.RegisterHandler)
			userGroup.POST("/register/verify", app.handlers.VerifyRegisterCode)
			userGroup.POST("/login", app.handlers.LoginUserHandler)
			userGroup.POST("/refresh", app.handlers.RefreshTokenHandler)
			userGroup.POST("/forgot_password", app.handlers.ForgotPasswordHandler)
			userGroup.POST("/forgot_password/verify", app.handlers.VerifyForgetPasswordCodeHandler)

			authGroup := userGroup.Group("")
			authGroup.Use(middlewares.UserMiddleware(app.handlers.tokenManager))
			{
				authGroup.GET("/", app.handlers.GetUserHandler)
				authGroup.PUT("/change_password", app.handlers.ChangePasswordHandler)
				authGroup.GET("/nodes", app.handlers.ListNodesHandler)
				authGroup.GET("/nodes/rentable", app.handlers.ListRentableNodesHandler)
				authGroup.GET("/nodes/rented", app.handlers.ListRentedNodesHandler)
				authGroup.POST("/nodes/:node_id", app.handlers.ReserveNodeHandler)
				authGroup.DELETE("/nodes/unreserve/:contract_id", app.handlers.UnreserveNodeHandler)
				authGroup.POST("/balance/charge", app.handlers.ChargeBalance)
				authGroup.GET("/balance", app.handlers.GetUserBalance)
				authGroup.PUT("/redeem/:voucher_code", app.handlers.RedeemVoucherHandler)
				authGroup.GET("/invoice/:invoice_id", app.handlers.DownloadInvoiceHandler)
				authGroup.GET("/invoice", app.handlers.ListUserInvoicesHandler)
				authGroup.GET("/pending-records", app.handlers.ListUserPendingRecordsHandler)
				// SSH Key management
				authGroup.GET("/ssh-keys", app.handlers.ListSSHKeysHandler)
				authGroup.POST("/ssh-keys", app.handlers.AddSSHKeyHandler)
				authGroup.DELETE("/ssh-keys/:ssh_key_id", app.handlers.DeleteSSHKeyHandler)
			}
		}

		deployerGroup := v1.Group("")
		deployerGroup.Use(middlewares.UserMiddleware(app.handlers.tokenManager))
		{
			deployerGroup.GET("/events", app.sseManager.HandleSSE)

			deploymentGroup := deployerGroup.Group("/deployments")
			{
				deploymentGroup.POST("", app.handlers.HandleDeployCluster)
				deploymentGroup.GET("", app.handlers.HandleListDeployments)
				deploymentGroup.DELETE("", app.handlers.HandleDeleteAllDeployments)
				deploymentGroup.GET("/:name", app.handlers.HandleGetDeployment)
				deploymentGroup.GET("/:name/kubeconfig", app.handlers.HandleGetKubeconfig)
				deploymentGroup.DELETE("/:name", app.handlers.HandleDeleteCluster)
				deploymentGroup.POST("/:name/nodes", app.handlers.HandleAddNode)
				deploymentGroup.DELETE("/:name/nodes/:node_name", app.handlers.HandleRemoveNode)
			}

			notificationGroup := deployerGroup.Group("/notifications")
			{
				notificationGroup.GET("", app.handlers.GetAllNotificationsHandler)
				notificationGroup.GET("/unread", app.handlers.GetUnreadNotificationsHandler)
				notificationGroup.PATCH("/read-all", app.handlers.MarkAllNotificationsReadHandler)
				notificationGroup.DELETE("", app.handlers.DeleteAllNotificationsHandler)
				notificationGroup.PATCH("/:notification_id/read", app.handlers.MarkNotificationReadHandler)
				notificationGroup.PATCH("/:notification_id/unread", app.handlers.MarkNotificationUnreadHandler)
				notificationGroup.DELETE("/:notification_id", app.handlers.DeleteNotificationHandler)
			}
		}
	}
	app.router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}

func (app *App) StartBackgroundWorkers() {
	go app.handlers.MonthlyInvoicesHandler()
	go app.handlers.TrackUserDebt(app.gridClient)
	go app.handlers.MonitorSystemBalanceAndHandleSettlement()
	go app.handlers.TrackClusterHealth()
	go app.handlers.TrackReservedNodeHealth(app.notificationService, app.handlers.proxyClient)
}

// Run starts the server
func (app *App) Run() error {
	app.StartBackgroundWorkers()

	// Start command socket
	go app.startCommandSocket()

	app.handlers.ewfEngine.ResumeRunningWorkflows()
	app.httpServer = &http.Server{
		Addr:    fmt.Sprintf(":%s", app.config.Server.Port),
		Handler: app.router,
	}

	logger.GetLogger().Info().Msgf("Starting server at %s:%s", app.config.Server.Host, app.config.Server.Port)

	if err := app.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.GetLogger().Error().Err(err).Msg("Failed to start server")
		return err
	}

	return nil
}

// Shutdown gracefully shuts down the server and worker manager
func (app *App) Shutdown(ctx context.Context) error {
	// First, cancel the app context to signal all components to stop
	if app.appCancel != nil {
		app.appCancel()
	}

	if app.httpServer != nil {
		if err := app.httpServer.Shutdown(ctx); err != nil {
			logger.GetLogger().Error().Err(err).Msg("Failed to shutdown HTTP server")
		}
	}

	if app.sseManager != nil {
		app.sseManager.Stop()
	}

	if app.redis != nil {
		if err := app.redis.Close(); err != nil {
			logger.GetLogger().Error().Err(err).Msg("Failed to close Redis connection")
		}
	}

	if app.db != nil {
		if err := app.db.Close(); err != nil {
			logger.GetLogger().Error().Err(err).Msg("Failed to close database connection")
		}
	}

	if app.handlers.substrateClient != nil {
		app.handlers.substrateClient.Close()
	}

	app.gridClient.Close()

	logger.CloseLogger()

	return nil
}

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
		case <-app.appCtx.Done():
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

		if app.appCtx.Err() != nil {
			return
		}

		logger.GetLogger().Error().Err(err).Msg("socket accept error")
		continue
	}
}

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

func (app *App) reloadNotificationConfig() error {
	cfg, err := internal.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if err = app.notificationService.ReloadNotificationConfig(cfg.Notification); err != nil {
		return fmt.Errorf("failed to reload notification config: %w", err)
	}

	return nil
}
