package handlers

import (
	"fmt"
	"kubecloud/internal/auth"
	cfg "kubecloud/internal/config"
	"kubecloud/internal/core/models"
	corepersistence "kubecloud/internal/core/persistence"
	"kubecloud/internal/core/workflows"
	"kubecloud/internal/infrastructure/grid"
	"kubecloud/internal/infrastructure/persistence"

	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/deployer"
)

type setup struct {
	tokenManager    auth.TokenManager
	substrateClient grid.SubstrateClient
	router          *gin.Engine
	network         string

	userRepo           models.UserRepository
	voucherRepo        models.VoucherRepository
	invoicesRepo       models.InvoiceRepository
	termsAndConditions cfg.TermsANDConditions
}

func SetUp(t testing.TB) (setup, error) {
	gin.SetMode(gin.TestMode)
	dir := t.TempDir()

	configPath := filepath.Join(dir, "config.json")

	dbPath := filepath.Join(dir, "testing.db")
	dsn := "sqlite3://" + dbPath

	privateKeyPath := filepath.Join(dir, "test_id_rsa")
	publicKeyPath := privateKeyPath + ".pub"

	// Generate SSH key pair
	cmd := exec.Command("ssh-keygen", "-t", "ed25519", "-f", privateKeyPath, "-N", "", "-q")
	err := cmd.Run()
	if err != nil {
		return setup{}, err
	}

	mnemonic := os.Getenv("TEST_MNEMONIC")
	if mnemonic == "" {
		return setup{}, fmt.Errorf("TEST_MNEMONIC environment variable is not set")
	}

	config := fmt.Sprintf(`
{
  "server": {
    "host": "0.0.0.0",
    "port": "3000"
  },
  "database": {
    "dsn": "%s"
  },
  "jwt_token": {
    "secret": "secret",
    "access_expiry_minutes": 60,
    "refresh_expiry_hours": 24
  },
  "admins": [],
  "mailSender": {
    "email": "email@domain.com",
    "sendgrid_key": "sendgrid_key",
    "timeout": 5,
    "max_concurrent_sends": 20,
    "max_attachment_size_mb": 10
  },
  "currency": "usd",
  "stripe_secret": "sk_test",
  "tfchain_url": "wss://tfchain.dev.grid.tf/wss",
  "gridproxy_url": "https://gridproxy.dev.grid.tf/",
  "voucher_name_length": 5,
  "terms_and_conditions": {
    "document_link": "https://manual.grid.tf/labs/knowledge_base/terms_conditions_all3",
    "document_hash": "6f2b4109704ba2883d978a7b94e5f295"
  },
  "activation_service_url": "https://activation.dev.grid.tf/activation/activate",
  "system_account": {
    "mnemonic": "%s",
    "network": "dev"
  },
  "redis": {
    "hostname": "localhost",
    "port": 6379,
    "password": "pass",
    "db": 0
  },
  "graphql_url": "https://graphql.dev.grid.tf/graphql",
  "firesquid_url": "https://firesquid.dev.grid.tf/graphql",
  "deployer_workers_num": 3,
  "invoice": {
    "name": "Name",
    "address": "Address",
    "governorate": "Cairo Governorate"
  },
  "ssh": {
    "private_key_path": "%s",
    "public_key_path": "%s"
  },
  "monitor_balance_interval_in_minutes": 2,
	"notify_admins_for_pending_records_in_hours": 1,
  "verification_code_length": 4,
  "kyc_verifier_api_url": "https://kyc.dev.grid.tf",
  "kyc_challenge_domain": "kyc.dev.grid.tf",
  "cluster_health_check_interval_in_hours": 1,
	"node_health_check": {
		"reserved_node_health_check_interval_in_hours": 1,
		"reserved_node_health_check_timeout_in_minutes": 1,
		"reserved_node_health_check_workers_num": 10
	}
}
`, dsn, mnemonic, privateKeyPath, publicKeyPath)

	err = os.WriteFile(configPath, []byte(config), 0644)
	if err != nil {
		return setup{}, err
	}

	viper.Reset()
	viper.SetConfigFile(configPath)
	err = viper.ReadInConfig()
	if err != nil {
		return setup{}, err
	}

	configuration, err := cfg.LoadConfig()
	if err != nil {
		return setup{}, err
	}

	gridClient, err := deployer.NewTFPluginClient(
		configuration.SystemAccount.Mnemonic, deployer.WithNetwork(configuration.SystemAccount.Network),
	)
	if err != nil {
		return setup{}, err
	}

	substrateClient := grid.NewSubstrateClient(configuration.SystemAccount.Mnemonic, gridClient)

	tokenManager := auth.NewTokenHandler(
		configuration.JwtToken.Secret,
		time.Duration(configuration.JwtToken.AccessExpiryMinutes)*time.Minute,
		time.Duration(configuration.JwtToken.RefreshExpiryHours)*time.Hour,
	)

	router := gin.New()

	// Add recovery middleware
	router.Use(gin.Recovery())

	db, err := persistence.NewGormDB(configuration.Database.DSN, models.DBPoolConfig{})
	if err != nil {
		return setup{}, fmt.Errorf("failed to create user storage: %w", err)
	}

	t.Cleanup(func() {
		// Clean up files
		_ = os.Remove(privateKeyPath)
		_ = os.Remove(publicKeyPath)
		_ = os.Remove(configPath)
		_ = os.Remove(dbPath)

		// Reset viper to avoid config leakage between tests
		viper.Reset()
		gridClient.Close()
	})

	return setup{
		tokenManager:    tokenManager,
		substrateClient: substrateClient,
		router:          router,
		network:         gridClient.Network,

		userRepo:           corepersistence.NewGormUserRepository(db),
		voucherRepo:        corepersistence.NewGormVoucherRepository(db),
		invoicesRepo:       corepersistence.NewGormInvoiceRepository(db),
		termsAndConditions: configuration.TermsANDConditions,
	}, nil
}

func (s setup) GetAuthToken(t *testing.T, id int, email, username string, isAdmin bool) string {
	tokenPair, err := s.tokenManager.CreateTokenPair(id, username, isAdmin)
	assert.NoError(t, err)
	return tokenPair.AccessToken
}

// Helper to create a test user
func (s setup) CreateTestUser(t *testing.T, email, username string, hashedPassword []byte, verified, admin bool, mnemonicRequired bool, code int, updatedAt time.Time) *models.User {
	mnemonic := ""
	sponseeAddress := ""
	if !mnemonicRequired {
		mnemonic = ""
	} else {
		mnemonic, _, err := workflows.SetupUserOnTFChain(s.substrateClient, s.termsAndConditions, s.network)
		require.NoError(t, err)
		sponseeKeyPair, err := auth.KeyPairFromMnemonic(mnemonic)
		require.NoError(t, err)
		sponseeAddress, err = auth.AccountAddressFromKeypair(sponseeKeyPair)
		require.NoError(t, err)
	}
	user := &models.User{
		Username:       username,
		Email:          email,
		Password:       hashedPassword,
		Verified:       verified,
		Admin:          admin,
		Code:           code,
		UpdatedAt:      updatedAt,
		Mnemonic:       mnemonic,
		AccountAddress: sponseeAddress,
	}

	err := s.userRepo.RegisterUser(user)
	require.NoError(t, err)
	return user
}
