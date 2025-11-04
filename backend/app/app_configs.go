package app

import (
	"context"
	"fmt"
	"kubecloud/internal"
	"kubecloud/internal/constants"
	"kubecloud/internal/logger"
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

type appConfig struct {
	*coreConfig
	*authConfig
	*communicationConfig
	*gridConfig
}

// coreConfig contains essential application services
type coreConfig struct {
	appCtx    context.Context
	config    internal.Configuration
	db        models.DB
	metrics   *metrics.Metrics
	ewfEngine *ewf.Engine
}

// authConfig contains authentication and security related services
type authConfig struct {
	tokenManager   internal.TokenManager
	systemIdentity substrate.Identity
	sponsorKeyPair subkey.KeyPair
	sponsorAddress string
	sshPublicKey   string
	kycClient      *internal.KYCClient
}

// communicationConfig contains notification and communication related services
type communicationConfig struct {
	mailService         internal.MailService
	sseManager          *internal.SSEManager
	notificationService *notification.NotificationService
}

// gridConfig contains grid and blockchain related services
type gridConfig struct {
	gridClient      deployer.TFPluginClient
	firesquidClient graphql.GraphQl
	graphql         graphql.GraphQl
	substrateClient *substrate.Substrate
}

func newAppConfig(ctx context.Context, config internal.Configuration) (appConfig, error) {
	coreConfig, err := newCoreConfig(ctx, config)
	if err != nil {
		return appConfig{}, err
	}

	gridConfig, err := newGridConfig(ctx, config)
	if err != nil {
		return appConfig{}, err
	}

	authConfig, err := newAuthConfig(ctx, config)
	if err != nil {
		return appConfig{}, err
	}

	communicationConfig, err := newCommunicationConfig(ctx, config, coreConfig.db, coreConfig.ewfEngine, coreConfig.metrics)
	if err != nil {
		return appConfig{}, err
	}

	return appConfig{
		coreConfig:          &coreConfig,
		authConfig:          &authConfig,
		communicationConfig: &communicationConfig,
		gridConfig:          &gridConfig,
	}, nil
}

func newCoreConfig(ctx context.Context, config internal.Configuration) (coreConfig, error) {
	dbPoolConfig := models.DBPoolConfig{
		MaxOpenConns:           config.Database.MaxOpenConns,
		MaxIdleConns:           config.Database.MaxIdleConns,
		ConnMaxLifetimeMinutes: config.Database.ConnMaxLifetimeMinutes,
		ConnMaxIdleTimeMinutes: config.Database.ConnMaxIdleTimeMinutes,
	}

	db, err := models.NewGormDB(config.Database.DSN, dbPoolConfig)
	if err != nil {
		return coreConfig{}, fmt.Errorf("failed to create user storage: %w", err)
	}

	// create storage for workflows
	ewfStore := models.NewGormEWFRepository(db)

	// initialize workflow ewfEngine
	ewfEngine, err := ewf.NewEngine(ewfStore)
	if err != nil {
		logger.GetLogger().Error().Err(err).Msg("failed to init EWF engine")
		return coreConfig{}, fmt.Errorf("failed to init workflow engine: %w", err)
	}

	return coreConfig{
		appCtx:    ctx,
		config:    config,
		db:        db,
		metrics:   metrics.NewMetrics(),
		ewfEngine: ewfEngine,
	}, nil
}

func newAuthConfig(ctx context.Context, config internal.Configuration) (authConfig, error) {
	tokenManager := internal.NewTokenHandler(
		config.JwtToken.Secret,
		time.Duration(config.JwtToken.AccessExpiryMinutes)*time.Minute,
		time.Duration(config.JwtToken.RefreshExpiryHours)*time.Hour,
	)

	systemIdentity, err := substrate.NewIdentityFromSr25519Phrase(config.SystemAccount.Mnemonic)
	if err != nil {
		return authConfig{}, fmt.Errorf("failed to create system identity: %w", err)
	}

	// Derive sponsor (system) account SS58 address once
	sponsorKeyPair, err := internal.KeyPairFromMnemonic(config.SystemAccount.Mnemonic)
	if err != nil {
		return authConfig{}, fmt.Errorf("failed to create sponsor keypair from system account: %w", err)
	}
	sponsorAddress, err := internal.AccountAddressFromKeypair(sponsorKeyPair)
	if err != nil {
		return authConfig{}, fmt.Errorf("failed to create sponsor address from keypair: %w", err)
	}

	// Initialize KYC client
	kycVerifierAPIURL := constants.KYCURLs[config.SystemAccount.Network]
	parsedUrl, err := url.Parse(kycVerifierAPIURL)
	if err != nil {
		return authConfig{}, fmt.Errorf("failed to parse KYC verifier API URL: %w", err)
	}
	kycChallengeDomain := parsedUrl.Hostname()

	kycClient := internal.NewKYCClient(
		kycVerifierAPIURL,
		kycChallengeDomain,
		nil, // Use default http.Client
	)

	if valid, err := kycClient.IsValidSponsor(ctx, sponsorAddress); err != nil || !valid {
		if err != nil {
			return authConfig{}, fmt.Errorf("failed to validate sponsor address, %w", err)
		}
		return authConfig{}, fmt.Errorf("the provided sponsor address can't be used as a sponsor")
	}

	sshPublicKeyBytes, err := os.ReadFile(config.SSH.PublicKeyPath)
	if err != nil {
		return authConfig{}, fmt.Errorf("failed to read SSH public key from %s: %w", config.SSH.PublicKeyPath, err)
	}

	return authConfig{
		tokenManager:   tokenManager,
		systemIdentity: systemIdentity,
		sponsorKeyPair: sponsorKeyPair,
		sponsorAddress: sponsorAddress,
		sshPublicKey:   strings.TrimSpace(string(sshPublicKeyBytes)),
		kycClient:      kycClient,
	}, nil
}

func newCommunicationConfig(ctx context.Context, config internal.Configuration, db models.DB, ewfEngine *ewf.Engine, metrics *metrics.Metrics) (communicationConfig, error) {
	mailService := internal.NewMailService(config.MailSender.SendGridKey, metrics)
	sseManager := internal.NewSSEManager()

	notificationRepo := models.NewGormNotificationRepository(db)
	notificationService, err := notification.NewNotificationService(notificationRepo, ewfEngine, config.Notification)
	if err != nil {
		return communicationConfig{}, fmt.Errorf("failed to create notification service: %w", err)
	}

	sseNotifier := notification.NewSSENotifier(sseManager)
	emailNotifier := notification.NewEmailNotifier(mailService, config.MailSender.Email, config.Notification.EmailTemplatesDirPath)
	err = emailNotifier.ParseTemplates()
	if err != nil {
		return communicationConfig{}, fmt.Errorf("failed to init notification templates: %w", err)
	}

	notificationService.RegisterNotifier(sseNotifier)
	notificationService.RegisterNotifier(emailNotifier)
	if err := notificationService.ValidateConfigsChannelsAgainstRegistered(); err != nil {
		return communicationConfig{}, fmt.Errorf("failed to validate notification configs channels against registered notifiers: %w", err)
	}

	return communicationConfig{
		mailService:         mailService,
		sseManager:          sseManager,
		notificationService: notificationService,
	}, nil
}

func newGridConfig(ctx context.Context, config internal.Configuration) (gridConfig, error) {
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
		return gridConfig{}, fmt.Errorf("failed to create TF grid client: %w", err)
	}

	fireSquidClient, err := graphql.NewGraphQl(constants.FireSquidURLs[config.SystemAccount.Network]...)
	if err != nil {
		logger.GetLogger().Error().Err(err).Msg("failed to connect to firesquid client")
		return gridConfig{}, fmt.Errorf("failed to connect to firesquid client: %w", err)
	}

	manager := substrate.NewManager(deployer.SubstrateURLs[config.SystemAccount.Network]...)
	substrateClient, err := manager.Substrate()
	if err != nil {
		return gridConfig{}, fmt.Errorf("failed to connect to TF chain: %w", err)
	}

	graphQl, err := graphql.NewGraphQl(deployer.GraphQlURLs[config.SystemAccount.Network]...)
	if err != nil {
		return gridConfig{}, fmt.Errorf("failed to connect to TF graphql: %w", err)
	}

	return gridConfig{
		gridClient:      gridClient,
		graphql:         graphQl,
		substrateClient: substrateClient,
		firesquidClient: fireSquidClient,
	}, nil
}

func (app *App) createHandlers() handlers {
	// Repositories
	userRepo := models.NewGormUserRepository(app.db)
	voucherRepo := models.NewGormVoucherRepository(app.db)
	pendingRecordRepo := models.NewGormPendingRecordRepository(app.db)
	notificationRepo := models.NewGormNotificationRepository(app.db)
	clusterRepo := models.NewGormClusterRepository(app.db)
	invoiceRepo := models.NewGormInvoiceRepository(app.db)
	userNodesRepo := models.NewGormUserNodesRepository(app.db)
	transactionRepo := models.NewGormTransactionRepository(app.db)
	settingsRepo := models.NewGormSettingsRepository(app.db)

	// Services
	userService := NewUserService(userRepo, voucherRepo, pendingRecordRepo)
	statsService := NewStatsService(userRepo, clusterRepo)
	notificationService := NewNotificationService(notificationRepo)
	nodeService := NewNodeService(userNodesRepo, userRepo)
	invoiceService := NewInvoiceService(invoiceRepo, userRepo, userNodesRepo)
	deploymentService := NewDeploymentService(clusterRepo, userRepo, userNodesRepo)
	adminService := NewAdminService(userRepo, userNodesRepo, pendingRecordRepo, voucherRepo, transactionRepo, settingsRepo)

	// Handlers
	userHandler := newUserHandler(userService, app.appConfig)
	statsHandler := newStatsHandler(statsService, app.appConfig)
	notificationHandler := newNotificationHandler(notificationService, app.appConfig)
	nodeHandler := newNodeHandler(nodeService, app.appConfig)
	invoiceHandler := newInvoiceHandler(invoiceService, app.appConfig)
	deploymentHandler := newDeploymentHandler(deploymentService, app.appConfig)
	adminHandler := newAdminHandler(adminService, app.appConfig)
	healthHandler := newHealthHandler(app.appConfig)

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
