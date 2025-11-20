package app

import (
	"context"
	"fmt"
	"kubecloud/internal/api/handlers"
	"kubecloud/internal/auth"
	"kubecloud/internal/billing"
	cfg "kubecloud/internal/config"
	"kubecloud/internal/core/models"
	corepersistence "kubecloud/internal/core/persistence"
	"kubecloud/internal/core/services"
	"kubecloud/internal/core/workers"
	grid "kubecloud/internal/infrastructure/grid"
	"kubecloud/internal/infrastructure/kyc"
	"kubecloud/internal/infrastructure/logger"
	mailservice "kubecloud/internal/infrastructure/mailservice"
	"kubecloud/internal/infrastructure/metrics"
	"kubecloud/internal/infrastructure/notification"
	"kubecloud/internal/infrastructure/persistence"
	"kubecloud/internal/infrastructure/realtime"
	"kubecloud/internal/infrastructure/substrate"

	"net/url"
	"os"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/deployer"
	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/graphql"
	"github.com/vedhavyas/go-subkey"
	"github.com/xmonader/ewf"
)

// Handlers contains application Handlers
type appHandlers struct {
	userHandler         handlers.UserHandler
	statsHandler        handlers.StatsHandler
	notificationHandler handlers.NotificationHandler
	nodeHandler         handlers.NodeHandler
	invoiceHandler      handlers.InvoiceHandler
	deploymentHandler   handlers.DeploymentHandler
	adminHandler        handlers.AdminHandler
	healthHandler       handlers.HealthHandler
	settingsHandler     handlers.SettingsHandler
}

type appDependencies struct {
	config        cfg.Configuration
	core          appCore
	security      appSecurity
	communication appCommunication
	infra         appInfrastructure
}

// appCore contains essential application services
type appCore struct {
	appCtx    context.Context
	db        models.DB
	metrics   *metrics.Metrics
	ewfEngine *ewf.Engine
}

// appSecurity contains authentication and security related services
type appSecurity struct {
	tokenManager   auth.TokenManager
	sponsorKeyPair subkey.KeyPair
	sponsorAddress string
	sshPublicKey   string
	kycClient      *kyc.KYCClient
}

// appCommunication contains notification and communication related services
type appCommunication struct {
	mailService            mailservice.MailService
	sseManager             *realtime.SSEManager
	notificationDispatcher *notification.NotificationDispatcher
}

// appInfrastructure contains grid and blockchain related services
type appInfrastructure struct {
	gridClient      deployer.TFPluginClient
	firesquidClient graphql.GraphQl
	graphql         graphql.GraphQl
	substrateClient substrate.Substrate
}

func createAppDependencies(ctx context.Context, config cfg.Configuration) (appDependencies, error) {
	appCore, err := createAppCore(ctx, config)
	if err != nil {
		return appDependencies{}, err
	}

	appInfrastructure, err := createAppInfrastructure(config)
	if err != nil {
		return appDependencies{}, err
	}

	appSecurity, err := createAppSecurity(ctx, config)
	if err != nil {
		return appDependencies{}, err
	}

	appCommunication, err := createAppCommunication(ctx, config, appCore.db, appCore.ewfEngine, appCore.metrics)
	if err != nil {
		return appDependencies{}, err
	}

	return appDependencies{
		config:        config,
		core:          appCore,
		security:      appSecurity,
		communication: appCommunication,
		infra:         appInfrastructure,
	}, nil
}

func createAppCore(ctx context.Context, config cfg.Configuration) (appCore, error) {
	dbPoolConfig := models.DBPoolConfig{
		MaxOpenConns:           config.Database.MaxOpenConns,
		MaxIdleConns:           config.Database.MaxIdleConns,
		ConnMaxLifetimeMinutes: config.Database.ConnMaxLifetimeMinutes,
		ConnMaxIdleTimeMinutes: config.Database.ConnMaxIdleTimeMinutes,
	}

	db, err := persistence.NewGormDB(config.Database.DSN, dbPoolConfig)
	if err != nil {
		return appCore{}, fmt.Errorf("failed to create user storage: %w", err)
	}

	// create storage for workflows
	ewfStore := corepersistence.NewGormEWFRepository(db)

	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", config.Redis.Hostname, config.Redis.Port),
		Password: config.Redis.Password,
		DB:       config.Redis.DB,
	})

	if err := client.Ping(ctx).Err(); err != nil {
		return appCore{}, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	qEngine := ewf.NewRedisQueueEngine(client)

	// initialize workflow ewfEngine
	ewfEngine, err := ewf.NewEngine(ewf.WithQueueEngine(qEngine), ewf.WithStore(ewfStore))
	if err != nil {
		logger.GetLogger().Error().Err(err).Msg("failed to init EWF engine")
		return appCore{}, fmt.Errorf("failed to init workflow engine: %w", err)
	}

	return appCore{
		appCtx:    ctx,
		db:        db,
		metrics:   metrics.NewMetrics(),
		ewfEngine: ewfEngine,
	}, nil
}

func createAppSecurity(ctx context.Context, config cfg.Configuration) (appSecurity, error) {
	tokenManager := auth.NewTokenHandler(
		config.JwtToken.Secret,
		time.Duration(config.JwtToken.AccessExpiryMinutes)*time.Minute,
		time.Duration(config.JwtToken.RefreshExpiryHours)*time.Hour,
	)

	// Derive sponsor (system) account SS58 address once
	sponsorKeyPair, err := auth.KeyPairFromMnemonic(config.SystemAccount.Mnemonic)
	if err != nil {
		return appSecurity{}, fmt.Errorf("failed to create sponsor keypair from system account: %w", err)
	}
	sponsorAddress, err := auth.AccountAddressFromKeypair(sponsorKeyPair)
	if err != nil {
		return appSecurity{}, fmt.Errorf("failed to create sponsor address from keypair: %w", err)
	}

	// Initialize KYC client
	kycVerifierAPIURL := grid.KYCURLs[config.SystemAccount.Network]
	parsedUrl, err := url.Parse(kycVerifierAPIURL)
	if err != nil {
		return appSecurity{}, fmt.Errorf("failed to parse KYC verifier API URL: %w", err)
	}
	kycChallengeDomain := parsedUrl.Hostname()

	kycClient := kyc.NewKYCClient(
		kycVerifierAPIURL,
		kycChallengeDomain,
		nil, // Use default http.Client
	)

	if valid, err := kycClient.IsValidSponsor(ctx, sponsorAddress); err != nil || !valid {
		if err != nil {
			return appSecurity{}, fmt.Errorf("failed to validate sponsor address, %w", err)
		}
		return appSecurity{}, fmt.Errorf("the provided sponsor address can't be used as a sponsor")
	}

	sshPublicKeyBytes, err := os.ReadFile(config.SSH.PublicKeyPath)
	if err != nil {
		return appSecurity{}, fmt.Errorf("failed to read SSH public key from %s: %w", config.SSH.PublicKeyPath, err)
	}

	return appSecurity{
		tokenManager:   tokenManager,
		sponsorKeyPair: sponsorKeyPair,
		sponsorAddress: sponsorAddress,
		sshPublicKey:   strings.TrimSpace(string(sshPublicKeyBytes)),
		kycClient:      kycClient,
	}, nil
}

func createAppCommunication(ctx context.Context, config cfg.Configuration, db models.DB, ewfEngine *ewf.Engine, metrics *metrics.Metrics) (appCommunication, error) {
	var mailService mailservice.MailService

	if config.DevMode {
		logger.GetLogger().Info().Msg("Dev mode enabled: using FakeMailService for OTP logging")
		mailService = mailservice.NewFakeMailService(metrics)
	} else {
		mailService = mailservice.NewSendGridMailService(config.MailSender, config.Server.Host, metrics)
	}

	// mailService := shared.NewMailService(config.MailSender, config.Server.Host, metrics)
	sseManager := realtime.NewSSEManager()

	notificationRepo := corepersistence.NewGormNotificationRepository(db)
	userRepo := corepersistence.NewGormUserRepository(db)
	notificationDispatcher, err := notification.NewNotificationDispatcher(notificationRepo, userRepo, ewfEngine)
	if err != nil {
		return appCommunication{}, fmt.Errorf("failed to create notification dispatcher: %w", err)
	}

	sseNotifier := notification.NewSSENotifier(sseManager)
	emailNotifier := notification.NewEmailNotifier(mailService, userRepo)
	err = emailNotifier.ParseTemplates()
	if err != nil {
		return appCommunication{}, fmt.Errorf("failed to init notification templates: %w", err)
	}

	notificationDispatcher.RegisterNotifier(sseNotifier)
	notificationDispatcher.RegisterNotifier(emailNotifier)
	if err := notificationDispatcher.ValidateConfigsChannelsAgainstRegistered(); err != nil {
		return appCommunication{}, fmt.Errorf("failed to validate notification configs channels against registered notifiers: %w", err)
	}

	return appCommunication{
		mailService:            mailService,
		sseManager:             sseManager,
		notificationDispatcher: notificationDispatcher,
	}, nil
}

func createAppInfrastructure(config cfg.Configuration) (appInfrastructure, error) {
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
		return appInfrastructure{}, fmt.Errorf("failed to create TF grid client: %w", err)
	}

	fireSquidClient, err := graphql.NewGraphQl(grid.FireSquidURLs[config.SystemAccount.Network]...)
	if err != nil {
		logger.GetLogger().Error().Err(err).Msg("failed to connect to firesquid client")
		return appInfrastructure{}, fmt.Errorf("failed to connect to firesquid client: %w", err)
	}

	graphQl, err := graphql.NewGraphQl(deployer.GraphQlURLs[config.SystemAccount.Network]...)
	if err != nil {
		return appInfrastructure{}, fmt.Errorf("failed to connect to TF graphql: %w", err)
	}

	tfChainClient, err := substrate.NewTFChainClient(
		config.SystemAccount.Network, config.SystemAccount.Mnemonic,
		config.TermsANDConditions.DocumentLink, config.TermsANDConditions.DocumentHash,
	)
	if err != nil {
		return appInfrastructure{}, fmt.Errorf("failed to create tf chain client: %w", err)
	}

	return appInfrastructure{
		gridClient:      gridClient,
		graphql:         graphQl,
		firesquidClient: fireSquidClient,
		substrateClient: tfChainClient,
	}, nil
}

func (app *App) createHandlers() appHandlers {
	// Repositories
	userRepo := corepersistence.NewGormUserRepository(app.core.db)
	voucherRepo := corepersistence.NewGormVoucherRepository(app.core.db)
	transferRecordsRepo := corepersistence.NewGormTransferRecordRepository(app.core.db)
	notificationRepo := corepersistence.NewGormNotificationRepository(app.core.db)
	clusterRepo := corepersistence.NewGormClusterRepository(app.core.db)
	invoiceRepo := corepersistence.NewGormInvoiceRepository(app.core.db)
	contractsRepo := corepersistence.NewGormUserContractDataRepository(app.core.db)
	transactionRepo := corepersistence.NewGormTransactionRepository(app.core.db)
	settingsRepo := corepersistence.NewGormSettingsRepository(app.core.db)

	// Services
	billingService := services.NewBillingService(
		userRepo, contractsRepo, transferRecordsRepo, clusterRepo,
		app.infra.substrateClient, app.infra.graphql, app.infra.gridClient,
		uint64(app.config.MinimumTFTAmountInWallet), services.Discount(app.config.AppliedDiscount),
	)
	userService := services.NewUserService(
		app.core.appCtx, userRepo, voucherRepo,
		app.infra.substrateClient, app.core.ewfEngine,
		app.security.kycClient, app.core.metrics, app.config.MailSender.TimeoutMin,
		app.config.Admins,
	)

	statsService := services.NewStatsService(
		userRepo, clusterRepo, app.infra.gridClient.GridProxyClient,
		app.infra.substrateClient, app.config.SystemAccount.Mnemonic,
	)

	notificationAPIService := services.NewNotificationService(notificationRepo)

	nodeService := services.NewNodeService(
		contractsRepo, userRepo, app.core.appCtx, app.core.ewfEngine,
		app.infra.gridClient, app.infra.substrateClient,
	)

	invoiceService := services.NewInvoiceService(
		invoiceRepo, userRepo,
		app.infra.firesquidClient, app.infra.graphql, app.infra.substrateClient,
		app.config.Invoice,
	)

	deploymentService := services.NewDeploymentService(
		app.core.appCtx, clusterRepo, userRepo, contractsRepo, app.core.ewfEngine,
		app.config.Debug, app.security.sshPublicKey, app.config.SSH.PrivateKeyPath, app.config.SystemAccount.Network,
	)

	adminService := services.NewAdminService(
		app.core.appCtx, userRepo, contractsRepo, transferRecordsRepo, voucherRepo,
		transactionRepo, app.infra.substrateClient, app.core.ewfEngine,
	)

	settingsService := services.NewSettingsService(settingsRepo)

	// Handlers
	stripeClient := &billing.DefaultStripeClient{}
	userHandler := handlers.NewUserHandler(
		userService, billingService, app.communication.notificationDispatcher,
		app.communication.mailService, app.security.tokenManager, stripeClient,
	)
	statsHandler := handlers.NewStatsHandler(statsService)
	notificationHandler := handlers.NewNotificationHandler(notificationAPIService)
	nodeHandler := handlers.NewNodeHandler(nodeService, billingService)
	deploymentHandler := handlers.NewDeploymentHandler(deploymentService, billingService)
	invoiceHandler := handlers.NewInvoiceHandler(invoiceService, app.communication.mailService)
	adminHandler := handlers.NewAdminHandler(adminService, billingService, app.communication.notificationDispatcher, app.communication.mailService)
	healthHandler := handlers.NewHealthHandler(app.config.SystemAccount.Network, app.infra.firesquidClient, app.infra.graphql, app.core.db)
	settingsHandler := handlers.NewSettingsHandler(settingsService)

	return appHandlers{
		userHandler:         userHandler,
		statsHandler:        statsHandler,
		notificationHandler: notificationHandler,
		nodeHandler:         nodeHandler,
		invoiceHandler:      invoiceHandler,
		deploymentHandler:   deploymentHandler,
		adminHandler:        adminHandler,
		healthHandler:       healthHandler,
		settingsHandler:     settingsHandler,
	}
}

func (app *App) createWorkers() workers.Workers {
	userRepo := corepersistence.NewGormUserRepository(app.core.db)
	transferRecordsRepo := corepersistence.NewGormTransferRecordRepository(app.core.db)
	clusterRepo := corepersistence.NewGormClusterRepository(app.core.db)
	invoiceRepo := corepersistence.NewGormInvoiceRepository(app.core.db)
	contractsRepo := corepersistence.NewGormUserContractDataRepository(app.core.db)

	workersService := services.NewWorkersService(
		app.core.appCtx, userRepo, contractsRepo, invoiceRepo, clusterRepo, transferRecordsRepo,
		app.communication.mailService, app.infra.gridClient, app.core.ewfEngine,
		app.communication.notificationDispatcher, app.infra.graphql, app.infra.firesquidClient,
		app.infra.substrateClient, app.config.Invoice, app.config.SystemAccount.Mnemonic,
		app.config.Currency, app.config.ClusterHealthCheckIntervalInHours,
		app.config.NodeHealthCheck.ReservedNodeHealthCheckIntervalInHours,
		app.config.NodeHealthCheck.ReservedNodeHealthCheckTimeoutInMinutes,
		app.config.NodeHealthCheck.ReservedNodeHealthCheckWorkersNum,
		app.config.SettleTransferRecordsIntervalInMinutes,
		app.config.NotifyAdminsForPendingRecordsInHours,
		app.config.MinimumTFTAmountInWallet, services.Discount(app.config.AppliedDiscount),
	)

	billingService := services.NewBillingService(
		userRepo, contractsRepo, transferRecordsRepo, clusterRepo,
		app.infra.substrateClient, app.infra.graphql, app.infra.gridClient,
		uint64(app.config.MinimumTFTAmountInWallet), services.Discount(app.config.AppliedDiscount),
	)

	return workers.NewWorkers(app.core.appCtx, workersService, billingService)
}
