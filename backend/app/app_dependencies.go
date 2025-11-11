package app

import (
	"context"
	"fmt"
	"kubecloud/internal"
	"kubecloud/internal/constants"
	"kubecloud/internal/logger"
	mailservice "kubecloud/internal/mailservice"
	"kubecloud/internal/metrics"
	"kubecloud/internal/notification"
	"kubecloud/models"
	"net/url"
	"os"
	"strings"
	"time"

	substrate "github.com/threefoldtech/tfchain/clients/tfchain-client-go"
	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/deployer"
	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/graphql"
	"github.com/vedhavyas/go-subkey"
	"github.com/xmonader/ewf"
)

// handlers contains application handlers
type handlers struct {
	userHandler         *userHandler
	statsHandler        *statsHandler
	notificationHandler *notificationHandler
	nodeHandler         *nodeHandler
	invoiceHandler      *invoiceHandler
	deploymentHandler   *deploymentHandler
	adminHandler        *adminHandler
	healthHandler       *healthHandler
}

type appDependencies struct {
	config        internal.Configuration
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
	tokenManager   internal.TokenManager
	systemIdentity substrate.Identity
	sponsorKeyPair subkey.KeyPair
	sponsorAddress string
	sshPublicKey   string
	kycClient      *internal.KYCClient
}

// appCommunication contains notification and communication related services
type appCommunication struct {
	mailService         mailservice.MailService
	sseManager          *internal.SSEManager
	notificationService *notification.NotificationService
}

// appInfrastructure contains grid and blockchain related services
type appInfrastructure struct {
	gridClient      deployer.TFPluginClient
	firesquidClient graphql.GraphQl
	graphql         graphql.GraphQl
	substrateClient *substrate.Substrate
}

func createAppDependencies(ctx context.Context, config internal.Configuration) (appDependencies, error) {
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

	appCommunication, err := createAppCommunication(config, appCore.db, appCore.ewfEngine, appCore.metrics)
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

func createAppCore(ctx context.Context, config internal.Configuration) (appCore, error) {
	dbPoolConfig := models.DBPoolConfig{
		MaxOpenConns:           config.Database.MaxOpenConns,
		MaxIdleConns:           config.Database.MaxIdleConns,
		ConnMaxLifetimeMinutes: config.Database.ConnMaxLifetimeMinutes,
		ConnMaxIdleTimeMinutes: config.Database.ConnMaxIdleTimeMinutes,
	}

	db, err := models.NewGormDB(config.Database.DSN, dbPoolConfig)
	if err != nil {
		return appCore{}, fmt.Errorf("failed to create user storage: %w", err)
	}

	// create storage for workflows
	ewfStore := models.NewGormEWFRepository(db)

	// initialize workflow ewfEngine
	ewfEngine, err := ewf.NewEngine(ewfStore)
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

func createAppSecurity(ctx context.Context, config internal.Configuration) (appSecurity, error) {
	tokenManager := internal.NewTokenHandler(
		config.JwtToken.Secret,
		time.Duration(config.JwtToken.AccessExpiryMinutes)*time.Minute,
		time.Duration(config.JwtToken.RefreshExpiryHours)*time.Hour,
	)

	systemIdentity, err := substrate.NewIdentityFromSr25519Phrase(config.SystemAccount.Mnemonic)
	if err != nil {
		return appSecurity{}, fmt.Errorf("failed to create system identity: %w", err)
	}

	// Derive sponsor (system) account SS58 address once
	sponsorKeyPair, err := internal.KeyPairFromMnemonic(config.SystemAccount.Mnemonic)
	if err != nil {
		return appSecurity{}, fmt.Errorf("failed to create sponsor keypair from system account: %w", err)
	}
	sponsorAddress, err := internal.AccountAddressFromKeypair(sponsorKeyPair)
	if err != nil {
		return appSecurity{}, fmt.Errorf("failed to create sponsor address from keypair: %w", err)
	}

	// Initialize KYC client
	kycVerifierAPIURL := constants.KYCURLs[config.SystemAccount.Network]
	parsedUrl, err := url.Parse(kycVerifierAPIURL)
	if err != nil {
		return appSecurity{}, fmt.Errorf("failed to parse KYC verifier API URL: %w", err)
	}
	kycChallengeDomain := parsedUrl.Hostname()

	kycClient := internal.NewKYCClient(
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
		systemIdentity: systemIdentity,
		sponsorKeyPair: sponsorKeyPair,
		sponsorAddress: sponsorAddress,
		sshPublicKey:   strings.TrimSpace(string(sshPublicKeyBytes)),
		kycClient:      kycClient,
	}, nil
}

func createAppCommunication(config internal.Configuration, db models.DB, ewfEngine *ewf.Engine, metrics *metrics.Metrics) (appCommunication, error) {
	var mailService mailservice.MailService

	if config.DevMode {
		logger.GetLogger().Info().Msg("Dev mode enabled: using FakeMailService for OTP logging")
		mailService = mailservice.NewFakeMailService(metrics)
	} else {
		mailService = mailservice.NewSendGridMailService(config.MailSender, config.Server.Host, metrics)
	}

	// mailService := internal.NewMailService(config.MailSender, config.Server.Host, metrics)
	sseManager := internal.NewSSEManager()

	notificationRepo := models.NewGormNotificationRepository(db)
	notificationService, err := notification.NewNotificationService(notificationRepo, ewfEngine, config.Notification)
	if err != nil {
		return appCommunication{}, fmt.Errorf("failed to create notification service: %w", err)
	}

	sseNotifier := notification.NewSSENotifier(sseManager)
	emailNotifier := notification.NewEmailNotifier(mailService, config.Notification.EmailTemplatesDirPath)
	err = emailNotifier.ParseTemplates()
	if err != nil {
		return appCommunication{}, fmt.Errorf("failed to init notification templates: %w", err)
	}

	notificationService.RegisterNotifier(sseNotifier)
	notificationService.RegisterNotifier(emailNotifier)
	if err := notificationService.ValidateConfigsChannelsAgainstRegistered(); err != nil {
		return appCommunication{}, fmt.Errorf("failed to validate notification configs channels against registered notifiers: %w", err)
	}

	return appCommunication{
		mailService:         mailService,
		sseManager:          sseManager,
		notificationService: notificationService,
	}, nil
}

func createAppInfrastructure(config internal.Configuration) (appInfrastructure, error) {
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

	fireSquidClient, err := graphql.NewGraphQl(constants.FireSquidURLs[config.SystemAccount.Network]...)
	if err != nil {
		logger.GetLogger().Error().Err(err).Msg("failed to connect to firesquid client")
		return appInfrastructure{}, fmt.Errorf("failed to connect to firesquid client: %w", err)
	}

	manager := substrate.NewManager(deployer.SubstrateURLs[config.SystemAccount.Network]...)
	substrateClient, err := manager.Substrate()
	if err != nil {
		return appInfrastructure{}, fmt.Errorf("failed to connect to TF chain: %w", err)
	}

	graphQl, err := graphql.NewGraphQl(deployer.GraphQlURLs[config.SystemAccount.Network]...)
	if err != nil {
		return appInfrastructure{}, fmt.Errorf("failed to connect to TF graphql: %w", err)
	}

	return appInfrastructure{
		gridClient:      gridClient,
		graphql:         graphQl,
		substrateClient: substrateClient,
		firesquidClient: fireSquidClient,
	}, nil
}

func (app *App) createHandlers() handlers {
	// Repositories
	userRepo := models.NewGormUserRepository(app.core.db)
	voucherRepo := models.NewGormVoucherRepository(app.core.db)
	pendingRecordRepo := models.NewGormPendingRecordRepository(app.core.db)
	notificationRepo := models.NewGormNotificationRepository(app.core.db)
	clusterRepo := models.NewGormClusterRepository(app.core.db)
	invoiceRepo := models.NewGormInvoiceRepository(app.core.db)
	userNodesRepo := models.NewGormUserNodesRepository(app.core.db)
	transactionRepo := models.NewGormTransactionRepository(app.core.db)
	settingsRepo := models.NewGormSettingsRepository(app.core.db)

	// Services
	userService := NewUserService(userRepo, voucherRepo, pendingRecordRepo)
	statsService := NewStatsService(userRepo, clusterRepo)
	notificationService := NewNotificationService(notificationRepo)
	nodeService := NewNodeService(userNodesRepo, userRepo)
	invoiceService := NewInvoiceService(invoiceRepo, userRepo, userNodesRepo)
	deploymentService := NewDeploymentService(clusterRepo, userRepo, userNodesRepo)
	adminService := NewAdminService(userRepo, userNodesRepo, pendingRecordRepo, voucherRepo, transactionRepo, settingsRepo)

	// Handlers
	userHandler := newUserHandler(
		app.core.appCtx, userService, app.security, app.core.metrics,
		app.core.ewfEngine, app.infra.substrateClient, app.communication.notificationService, app.communication.mailService,
		app.config.MailSender.TimeoutMin, app.config.Admins,
	)
	statsHandler := newStatsHandler(statsService, app.infra.gridClient.GridProxyClient)
	notificationHandler := newNotificationHandler(notificationService)

	nodeHandler := newNodeHandler(app.core.appCtx, nodeService, app.core.ewfEngine,
		app.infra.gridClient, app.infra.substrateClient, app.config.NodeHealthCheck)

	invoiceHandler := newInvoiceHandler(invoiceService, app.infra.firesquidClient, app.infra.graphql,
		app.infra.substrateClient, app.communication.mailService, app.config.Invoice, app.config.Currency)
	deploymentHandler := newDeploymentHandler(app.core.appCtx, deploymentService,
		app.core.ewfEngine, app.config.SystemAccount.Network, app.config.SSH.PrivateKeyPath,
		app.config.ClusterHealthCheckIntervalInHours, app.security.sshPublicKey, app.config.Debug,
	)
	adminHandler := newAdminHandler(app.core.appCtx, adminService, app.core.ewfEngine,
		app.communication.notificationService, app.communication.mailService,
		app.infra.substrateClient, app.security.systemIdentity,
		app.config.VoucherNameLength, app.config.MonitorBalanceIntervalInMinutes,
		app.config.NotifyAdminsForPendingRecordsInHours,
	)
	healthHandler := newHealthHandler(app.config.SystemAccount.Network, app.infra.firesquidClient, app.infra.graphql, app.core.db)

	return handlers{
		userHandler:         &userHandler,
		statsHandler:        &statsHandler,
		notificationHandler: &notificationHandler,
		nodeHandler:         &nodeHandler,
		invoiceHandler:      &invoiceHandler,
		deploymentHandler:   &deploymentHandler,
		adminHandler:        &adminHandler,
		healthHandler:       &healthHandler,
	}
}
