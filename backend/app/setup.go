package app

import (
	"context"
	"fmt"
	"kubecloud/internal"
	"kubecloud/models"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func SetUp(t testing.TB) (*App, error) {
	gin.SetMode(gin.TestMode)
	dir := t.TempDir()

	configPath := filepath.Join(dir, "config.json")
	dbPath := filepath.Join(dir, "testing.db")
	workflowPath := filepath.Join(dir, "workflow_testing.db")
	notificationConfigPath := filepath.Join(dir, "notification-config.json")

	privateKeyPath := filepath.Join(dir, "test_id_rsa")
	publicKeyPath := privateKeyPath + ".pub"

	redisHost := os.Getenv("REDIS_HOST")
	if redisHost == "" {
		redisHost = "localhost"
	}

	// Generate SSH key pair
	cmd := exec.Command("ssh-keygen", "-t", "ed25519", "-f", privateKeyPath, "-N", "", "-q")
	err := cmd.Run()
	if err != nil {
		return nil, err
	}

	mnemonic := os.Getenv("TEST_MNEMONIC")
	if mnemonic == "" {
		return nil, fmt.Errorf("TEST_MNEMONIC environment variable must be set for tests")
	}

	config := fmt.Sprintf(`
{
  "server": {
    "host": "0.0.0.0",
    "port": "3000"
  },
  "database": {
    "file": "%s"
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
  "graphql_url": "https://graphql.dev.grid.tf/graphql",
  "firesquid_url": "https://firesquid.dev.grid.tf/graphql",
  "redis": {
    "host": "%s",
    "port": 6379,
    "password": "pass",
    "db": 0
  },
  "deployer_workers_num": 3,
  "invoice": {
    "name": "Name",
    "address": "Address",
    "governorate": "Cairo Governorate"
  },
  "workflow_db_file": "%s",
  "ssh": {
    "private_key_path": "%s",
    "public_key_path": "%s"
  },
  "monitor_balance_interval_in_minutes": 2,
	"notify_admins_for_pending_records_in_hours": 1,
  "kyc_verifier_api_url": "https://kyc.dev.grid.tf",
  "kyc_challenge_domain": "kyc.dev.grid.tf",
  "notification_config_path": "%s",
  "cluster_health_check_interval_in_hours": 1,
  "reserved_node_health_check_interval_in_hours": 1,
  "reserved_node_health_check_timeout_in_minutes": 1,
  "reserved_node_health_check_workers_num": 10
}
`, dbPath, mnemonic, redisHost, workflowPath, privateKeyPath, publicKeyPath, notificationConfigPath)

	err = os.WriteFile(configPath, []byte(config), 0644)
	if err != nil {
		return nil, err
	}

	notificationConfigBytes, err := os.ReadFile("../notification-config-example.json")
	if err != nil {
		return nil, fmt.Errorf("failed to read notification-config-example.json: %w", err)
	}

	err = os.WriteFile(notificationConfigPath, notificationConfigBytes, 0644)
	if err != nil {
		return nil, err
	}

	viper.Reset()
	viper.SetConfigFile(configPath)
	err = viper.ReadInConfig()
	if err != nil {
		return nil, err
	}

	configuration, err := internal.LoadConfig()
	if err != nil {
		return nil, err
	}

	app, err := NewApp(context.Background(), configuration)
	if err != nil {
		return nil, err
	}

	app.httpServer = nil

	t.Cleanup(func() {
		// Shutdown the app gracefully to close all connections
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := app.Shutdown(ctx); err != nil {
			t.Logf("Warning: failed to shutdown app cleanly: %v", err)
		}

		// Clean up files
		_ = os.Remove(privateKeyPath)
		_ = os.Remove(publicKeyPath)
		_ = os.Remove(configPath)
		_ = os.Remove(dbPath)
		_ = os.Remove(workflowPath)
		_ = os.Remove(notificationConfigPath)

		// Reset viper to avoid config leakage between tests
		viper.Reset()
	})

	return app, nil
}

func GetAuthToken(t *testing.T, app *App, id int, email, username string, isAdmin bool) string {
	tokenPair, err := app.handlers.tokenManager.CreateTokenPair(id, username, isAdmin)
	assert.NoError(t, err)
	return tokenPair.AccessToken
}

// Helper to create a test user
func CreateTestUser(t *testing.T, app *App, email, username string, hashedPassword []byte, verified, admin bool, mnemonicRequired bool, code int, updatedAt time.Time) *models.User {
	mnemonic := ""
	sponseeAddress := ""
	if !mnemonicRequired {
		mnemonic = ""
	} else {
		mnemonic, _, err := internal.SetupUserOnTFChain(app.handlers.substrateClient, app.config)
		require.NoError(t, err)
		sponseeKeyPair, err := internal.KeyPairFromMnemonic(mnemonic)
		require.NoError(t, err)
		sponseeAddress, err = internal.AccountAddressFromKeypair(sponseeKeyPair)
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
	err := app.handlers.db.RegisterUser(user)
	require.NoError(t, err)
	return user
}
