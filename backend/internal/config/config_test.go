package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/go-playground/validator"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig(t *testing.T) {
	// Create a temporary directory for test files
	tempDir := t.TempDir()

	// Create a temporary config file
	configPath := filepath.Join(tempDir, "config.json")
	notificationConfigPath := filepath.Join(tempDir, "notification-config.json")

	// Create SSH key paths
	privateKeyPath := filepath.Join(tempDir, "id_rsa")
	publicKeyPath := filepath.Join(tempDir, "id_rsa.pub")

	// Create a valid config JSON
	configJSON := `{
		"server": {
			"host": "0.0.0.0",
			"port": "8080"
		},
		"redis": {
			"hostname": "localhost",
			"port": 6379,
			"password": "pass",
			"db": 0
		},
		"database": {
			"dsn": "sqlite3:///` + filepath.Join(tempDir, "test.db") + `"
		},
		"jwt_token": {
			"secret": "test_secret",
			"access_expiry_minutes": 60,
			"refresh_expiry_hours": 24
		},
		"admins": ["admin@example.com"],
		"mailSender": {
			"email": "test@example.com",
			"sendgrid_key": "test_key",
			"timeout": 5,
			"max_concurrent_sends": 10,
			"max_attachment_size_mb": 10
		},
		"dev_mode": true,
		"currency": "usd",
		"stripe_secret": "secret",
		"voucher_name_length": 8,
		"terms_and_conditions": {
			"document_link": "https://example.com/terms",
			"document_hash": "sha256:example_hash_here"
		},
		"system_account": {
			"mnemonic": "test mnemonic phrase for system account",
			"network": "test"
		},
		"deployer_workers_num": 3,
		"invoice": {
			"name": "Test Company",
			"address": "123 Test St",
			"governorate": "Test Governorate"
		},
		"ssh": {
			"private_key_path": "` + privateKeyPath + `",
			"public_key_path": "` + publicKeyPath + `"
		},
		"monitor_balance_interval_in_minutes": 5,
		"notify_admins_for_pending_records_in_hours": 2,
		"cluster_health_check_interval_in_hours": 1,
		"node_health_check": {
			"reserved_node_health_check_interval_in_hours": 1,
			"reserved_node_health_check_timeout_in_minutes": 1,
			"reserved_node_health_check_workers_num": 10
		},
		"logger": {
			"log_dir": "` + tempDir + `",
			"max_size": 10,
			"max_backups": 5,
			"max_age_days": 30,
			"compress": true
		},
		"loki": {
			"url": "http://localhost:3100/loki/api/v1/push",
			"flush_interval_second": 5
		},
		"notification_config_path": "` + notificationConfigPath + `"
	}`

	// Write the config files
	err := os.WriteFile(configPath, []byte(configJSON), 0644)
	require.NoError(t, err, "Failed to write test config file")

	// Create empty SSH key files
	err = os.WriteFile(privateKeyPath, []byte("test private key"), 0600)
	require.NoError(t, err, "Failed to write test private key file")

	err = os.WriteFile(publicKeyPath, []byte("test public key"), 0644)
	require.NoError(t, err, "Failed to write test public key file")

	// Set up viper to use our test config
	viper.Reset()
	viper.SetConfigFile(configPath)
	err = viper.ReadInConfig()
	require.NoError(t, err, "Failed to read test config file")

	// Load the configuration
	config, err := LoadConfig()
	require.NoError(t, err, "LoadConfig() failed")

	// Verify the loaded configuration
	assert.Equal(t, "0.0.0.0", config.Server.Host)
	assert.Equal(t, "8080", config.Server.Port)
	assert.Equal(t, "sqlite3:///"+filepath.Join(tempDir, "test.db"), config.Database.DSN)
	assert.Equal(t, "test_secret", config.JwtToken.Secret)
	assert.Equal(t, 60, config.JwtToken.AccessExpiryMinutes)
	assert.Equal(t, 24, config.JwtToken.RefreshExpiryHours)
	assert.Contains(t, config.Admins, "admin@example.com")
	assert.Equal(t, "test@example.com", config.MailSender.Email)
	assert.Equal(t, "test_key", config.MailSender.SendGridKey)
	assert.Equal(t, "usd", config.Currency)
	assert.Equal(t, "secret", config.StripeSecret)
	assert.Equal(t, 8, config.VoucherNameLength)
	assert.Equal(t, "https://example.com/terms", config.TermsANDConditions.DocumentLink)
	assert.Equal(t, "test mnemonic phrase for system account", config.SystemAccount.Mnemonic)
	assert.Equal(t, "test", config.SystemAccount.Network)
	assert.Equal(t, 3, config.DeployerWorkersNum)
	assert.Equal(t, "Test Company", config.Invoice.Name)
	assert.Equal(t, "123 Test St", config.Invoice.Address)
	assert.Equal(t, "Test Governorate", config.Invoice.Governorate)
	assert.Equal(t, privateKeyPath, config.SSH.PrivateKeyPath)
	assert.Equal(t, publicKeyPath, config.SSH.PublicKeyPath)
	assert.Equal(t, 5, config.MonitorBalanceIntervalInMinutes)
	assert.Equal(t, 2, config.NotifyAdminsForPendingRecordsInHours)
	assert.Equal(t, 1, config.ClusterHealthCheckIntervalInHours)
	assert.Equal(t, 1, config.NodeHealthCheck.ReservedNodeHealthCheckIntervalInHours)
	assert.Equal(t, 1, config.NodeHealthCheck.ReservedNodeHealthCheckTimeoutInMinutes)
	assert.Equal(t, 10, config.NodeHealthCheck.ReservedNodeHealthCheckWorkersNum)
	assert.Equal(t, tempDir, config.Logger.LogDir)
	assert.Equal(t, 10, config.Logger.MaxSize)
	assert.Equal(t, 5, config.Logger.MaxBackups)
	assert.Equal(t, 30, config.Logger.MaxAgeDays)
	assert.True(t, config.Logger.Compress)
	assert.Equal(t, "http://localhost:3100/loki/api/v1/push", config.Loki.URL)
	assert.Equal(t, 5, config.Loki.FlushIntervalSecond)
}

func TestDefaultTagsInConfig(t *testing.T) {
	// Create a temporary directory for test files
	tempDir := t.TempDir()

	// Create a temporary config file with minimal settings
	configPath := filepath.Join(tempDir, "minimal_config.json")

	// Create a minimal config JSON with only required fields
	configJSON := `{
		"server": {
    "host": "0.0.0.0",
    "port": "8080"
		},
		"database": {
			"dsn": "postgres://user:pass@localhost:5432/db"
		},
		"redis": {
			"hostname": "localhost",
			"port": 6379,
			"password": "pass",
			"db": 0
		},
		"jwt_token": {
			"secret": "your-jwt-secret-key-here-make-it-long-and-secure"
		},
		"admins": ["admin@admin.com"],
		"mailSender": {
			"email": "noreply@example.com",
			"sendgrid_key": "your-sendgrid-api-key-here"
		},
		"stripe_secret": "stripe_secret_key_here",
		"terms_and_conditions": {
			"document_link": "https://example.com/terms",
			"document_hash": "sha256:example_hash_here"
		},
		"system_account": {
			"mnemonic": "your system account 12-word mnemonic phrase goes here for blockchain operations",
			"network": "main"
		},
		"invoice": {
			"name": "KubeCloud Invoice",
			"address": "123 KubeCloud St, Cloud City, CC 12345",
			"governorate": "Cloud Governorate"
		},
		"ssh": {
			"private_key_path": "/root/.ssh/id_rsa",
			"public_key_path": "/root/.ssh/id_rsa.pub"
		},
		"loki": {
			"url": "http://loki:3100/loki/api/v1/push"
		}
	}`

	// Write the config file
	err := os.WriteFile(configPath, []byte(configJSON), 0644)
	require.NoError(t, err, "Failed to write test config file")

	// Create empty SSH key files
	err = os.WriteFile(filepath.Join(tempDir, "id_rsa"), []byte("test private key"), 0600)
	require.NoError(t, err, "Failed to write test private key file")

	err = os.WriteFile(filepath.Join(tempDir, "id_rsa.pub"), []byte("test public key"), 0644)
	require.NoError(t, err, "Failed to write test public key file")

	// Set up viper to use our test config
	viper.Reset()
	viper.SetConfigFile(configPath)
	err = viper.ReadInConfig()
	require.NoError(t, err, "Failed to read test config file")

	// Load the configuration
	config, err := LoadConfig()
	require.NoError(t, err, "LoadConfig() failed")

	// Verify default values were applied
	assert.Equal(t, "0.0.0.0", config.Server.Host, "Default server host not applied")
	assert.Equal(t, "8080", config.Server.Port, "Default server port not applied")
	assert.Equal(t, 25, config.Database.MaxOpenConns, "Default max idle conns not applied")
	assert.Equal(t, 25, config.Database.MaxIdleConns, "Default max idle conns not applied")
	assert.Equal(t, 60, config.Database.ConnMaxLifetimeMinutes, "Default conn max lifetime not applied")
	assert.Equal(t, 30, config.Database.ConnMaxIdleTimeMinutes, "Default conn max idle time not applied")
	assert.Equal(t, "usd", config.Currency, "Default currency not applied")
	assert.Equal(t, 8, config.VoucherNameLength, "Default voucher name length not applied")
	assert.Equal(t, 4, config.VerificationCodeLength, "Default verification code length not applied")
	assert.Equal(t, 1, config.DeployerWorkersNum, "Default deployer workers num not applied")
	assert.Equal(t, 1, config.ClusterHealthCheckIntervalInHours, "Default cluster health check interval not applied")
	assert.Equal(t, 60, config.JwtToken.AccessExpiryMinutes, "Default JWT access expiry not applied")
	assert.Equal(t, 24, config.JwtToken.RefreshExpiryHours, "Default JWT refresh expiry not applied")
	assert.Equal(t, 1, config.NodeHealthCheck.ReservedNodeHealthCheckIntervalInHours, "Default node health check interval not applied")
	assert.Equal(t, 1, config.NodeHealthCheck.ReservedNodeHealthCheckTimeoutInMinutes, "Default node health check timeout not applied")
	assert.Equal(t, 10, config.NodeHealthCheck.ReservedNodeHealthCheckWorkersNum, "Default node health check workers num not applied")
	assert.Equal(t, 120, config.MonitorBalanceIntervalInMinutes, "Default monitor balance interval not applied")
	assert.Equal(t, 24, config.NotifyAdminsForPendingRecordsInHours, "Default notify admins for pending records interval not applied")
}

func TestLoadConfig_InvalidConfig(t *testing.T) {
	// Create a temporary directory for test files
	tempDir := t.TempDir()

	// Create a temporary config file
	configPath := filepath.Join(tempDir, "invalid_config.json")

	// Test cases for invalid configurations
	testCases := []struct {
		name        string
		configJSON  string
		expectedErr string
	}{
		{
			name: "invalid database DSN",
			configJSON: `{
				"server": {
					"host": "0.0.0.0",
					"port": "8080"
				},
				"database": {
					"dsn": "invalid-dsn"
				},
				"jwt_token": {
					"secret": "test_secret"
				},
				"admins": ["admin@example.com"],
				"stripe_secret": "secret",
				"voucher_name_length": 8,
				"ssh": {
					"private_key_path": "/tmp/id_rsa",
					"public_key_path": "/tmp/id_rsa.pub"
				},
				"monitor_balance_interval_in_minutes": 5,
				"notify_admins_for_pending_records_in_hours": 2,
				"cluster_health_check_interval_in_hours": 1,
				"node_health_check": {
					"reserved_node_health_check_interval_in_hours": 1,
					"reserved_node_health_check_timeout_in_minutes": 1,
					"reserved_node_health_check_workers_num": 10
				}
			}`,
			expectedErr: "validation error on field 'Configuration.Database.DSN': dsn",
		},
		{
			name: "invalid max_idle_conns",
			configJSON: `{
				"server": {
					"host": "0.0.0.0",
					"port": "8080"
				},
				"database": {
					"dsn": "sqlite3:///test.db",
					"max_open_conns": 5,
					"max_idle_conns": 10
				},
				"jwt_token": {
					"secret": "test_secret"
				},
				"admins": ["admin@example.com"],
				"stripe_secret": "secret",
				"voucher_name_length": 8,
				"ssh": {
					"private_key_path": "/tmp/id_rsa",
					"public_key_path": "/tmp/id_rsa.pub"
				},
				"monitor_balance_interval_in_minutes": 5,
				"notify_admins_for_pending_records_in_hours": 2,
				"cluster_health_check_interval_in_hours": 1,
				"node_health_check": {
					"reserved_node_health_check_interval_in_hours": 1,
					"reserved_node_health_check_timeout_in_minutes": 1,
					"reserved_node_health_check_workers_num": 10
				}
			}`,
			expectedErr: "validation error on field 'Configuration.Database.MaxIdleConns': lteMaxOpenConns",
		},
		{
			name: "missing jwt_token",
			configJSON: `{
				"server": {
					"host": "0.0.0.0",
					"port": "8080"
				},
				"database": {
					"dsn": "sqlite3:///test.db"
				},
				"admins": ["admin@example.com"],
				"stripe_secret": "secret",
				"voucher_name_length": 8,
				"ssh": {
					"private_key_path": "/tmp/id_rsa",
					"public_key_path": "/tmp/id_rsa.pub"
				},
				"monitor_balance_interval_in_minutes": 5,
				"notify_admins_for_pending_records_in_hours": 2,
				"cluster_health_check_interval_in_hours": 1,
				"node_health_check": {
					"reserved_node_health_check_interval_in_hours": 1,
					"reserved_node_health_check_timeout_in_minutes": 1,
					"reserved_node_health_check_workers_num": 10
				}
			}`,
			expectedErr: "validation error on field 'Configuration.JwtToken.Secret': required",
		},
		{
			name: "missing admins",
			configJSON: `{
				"server": {
					"host": "0.0.0.0",
					"port": "8080"
				},
				"database": {
					"dsn": "sqlite3:///test.db"
				},
				"jwt_token": {
					"secret": "test_secret"
				},
				"stripe_secret": "secret",
				"voucher_name_length": 8,
				"ssh": {
					"private_key_path": "/tmp/id_rsa",
					"public_key_path": "/tmp/id_rsa.pub"
				},
				"monitor_balance_interval_in_minutes": 5,
				"notify_admins_for_pending_records_in_hours": 2,
				"cluster_health_check_interval_in_hours": 1,
				"node_health_check": {
					"reserved_node_health_check_interval_in_hours": 1,
					"reserved_node_health_check_timeout_in_minutes": 1,
					"reserved_node_health_check_workers_num": 10
				}
			}`,
			expectedErr: "validation error on field 'Configuration.Admins': required",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Write the invalid config file
			err := os.WriteFile(configPath, []byte(tc.configJSON), 0644)
			require.NoError(t, err, "Failed to write test config file")

			// Set up viper to use our test config
			viper.Reset()
			viper.SetConfigFile(configPath)
			err = viper.ReadInConfig()
			require.NoError(t, err, "Failed to read test config file")

			// Load the configuration, expect error
			_, err = LoadConfig()
			require.Error(t, err, "LoadConfig() should fail with invalid config")
			assert.Contains(t, err.Error(), tc.expectedErr)
		})
	}
}

func TestSQLiteProductionCheck(t *testing.T) {
	tempDir := t.TempDir()
	privateKeyPath := filepath.Join(tempDir, "id_rsa")
	publicKeyPath := filepath.Join(tempDir, "id_rsa.pub")
	err := os.WriteFile(privateKeyPath, []byte("test"), 0600)
	require.NoError(t, err, "Failed to write test private key file")
	err = os.WriteFile(publicKeyPath, []byte("test"), 0644)
	require.NoError(t, err, "Failed to write test public key file")

	tests := []struct {
		name      string
		devMode   bool
		dsn       string
		shouldErr bool
	}{
		{"sqlite with dev_mode=true allowed", true, "sqlite3:///test.db", false},
		{"sqlite with dev_mode=false blocked", false, "sqlite3:///test.db", true},
		{"sqlite:// scheme also blocked in production", false, "sqlite:///test.db", true},
		{"postgres allowed in production", false, "postgres://user:pass@localhost:5432/db", false},
		{"postgres allowed in dev", true, "postgres://user:pass@localhost:5432/db", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configPath := filepath.Join(tempDir, "test.json")
			configJSON := `{
				"server": {"host": "localhost", "port": "8080"},
				"database": {"dsn": "` + tt.dsn + `"},
				"redis": {"hostname": "localhost", "port": 6379, "db": 0},
				"jwt_token": {"secret": "secret"},
				"admins": ["admin@test.com"],
				"mailSender": {"email": "test@test.com", "sendgrid_key": "key"},
				"dev_mode": ` + map[bool]string{true: "true", false: "false"}[tt.devMode] + `,
				"stripe_secret": "secret",
				"terms_and_conditions": {"document_link": "https://test.com", "document_hash": "hash"},
				"system_account": {"mnemonic": "test mnemonic", "network": "test"},
				"invoice": {"name": "Test", "address": "Test", "governorate": "Test"},
				"ssh": {"private_key_path": "` + privateKeyPath + `", "public_key_path": "` + publicKeyPath + `"},
				"monitor_balance_interval_in_minutes": 1,
				"notify_admins_for_pending_records_in_hours": 1,
				"node_health_check": {"reserved_node_health_check_interval_in_hours": 1, "reserved_node_health_check_timeout_in_minutes": 1, "reserved_node_health_check_workers_num": 1}
			}`
			err := os.WriteFile(configPath, []byte(configJSON), 0644)
			require.NoError(t, err, "Failed to write test config file")

			viper.Reset()
			viper.SetConfigFile(configPath)
			err = viper.ReadInConfig()
			require.NoError(t, err, "Failed to read test config file")

			_, err = LoadConfig()
			if tt.shouldErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "SQLite not allowed in production")
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestRegisterConfigValidators(t *testing.T) {
	// Test DSN validator
	t.Run("DSN Validator", func(t *testing.T) {
		v := validator.New()
		registerConfigValidators(v)

		// Valid DSNs
		validDSNs := []string{
			"postgres://user:pass@localhost:5432/dbname?sslmode=disable",
			"sqlite:///absolute/path.db",
			"sqlite3:///absolute/path.db",
		}

		for _, dsn := range validDSNs {
			err := v.Var(dsn, "dsn")
			assert.NoError(t, err, "DSN %q should be valid", dsn)
		}

		// Invalid DSNs
		invalidDSNs := []string{
			"",
			"invalid-dsn",
			"mysql://user:pass@localhost:3306/dbname", // Not supported
			"sqlite://", // Empty path
		}

		for _, dsn := range invalidDSNs {
			err := v.Var(dsn, "dsn")
			assert.Error(t, err, "DSN %q should be invalid", dsn)
		}
	})

	// Test DB struct validation
	t.Run("DB Struct Validation", func(t *testing.T) {
		v := validator.New()
		registerConfigValidators(v)

		// Valid DB configurations
		validDBs := []DB{
			{DSN: "postgres://user:pass@localhost:5432/dbname"},
			{DSN: "sqlite:///path.db", MaxOpenConns: 10, MaxIdleConns: 5},
			{DSN: "sqlite:///path.db", MaxOpenConns: 10, MaxIdleConns: 10},
			{DSN: "sqlite:///path.db", MaxOpenConns: 0, MaxIdleConns: 10}, // MaxOpenConns <= 0 means no limit
		}

		for _, db := range validDBs {
			err := v.Struct(db)
			assert.NoError(t, err, "DB %+v should be valid", db)
		}

		// Invalid DB configurations
		invalidDBs := []DB{
			{DSN: ""},            // Empty DSN
			{DSN: "invalid-dsn"}, // Invalid DSN format
			{DSN: "sqlite:///path.db", MaxOpenConns: 5, MaxIdleConns: 10}, // MaxIdleConns > MaxOpenConns
		}

		for _, db := range invalidDBs {
			err := v.Struct(db)
			assert.Error(t, err, "DB %+v should be invalid", db)
		}
	})
}
