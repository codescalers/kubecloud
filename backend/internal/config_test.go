package internal

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
		"currency": "usd",
		"stripe_secret": "secret",
		"voucher_name_length": 8,
		"terms_and_conditions": {
			"document_link": "https://example.com/terms"
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

	// Create a valid notification config JSON
	notificationConfigJSON := `{
		"template_types": {
			"user_registration": {
				"default": {
					"channels": ["email"],
					"severity": "info"
				}
			},
			"password_reset": {
				"default": {
					"channels": ["email"],
					"severity": "info"
				}
			}
		},
		"email_templates_dir_path": "templates/email"
	}`

	// Write the config files
	err := os.WriteFile(configPath, []byte(configJSON), 0644)
	require.NoError(t, err, "Failed to write test config file")

	err = os.WriteFile(notificationConfigPath, []byte(notificationConfigJSON), 0644)
	require.NoError(t, err, "Failed to write test notification config file")

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

	// Verify notification config
	assert.Len(t, config.Notification.TemplateTypes, 2)
	assert.Contains(t, config.Notification.TemplateTypes, "user_registration")
	assert.Contains(t, config.Notification.TemplateTypes, "password_reset")
	assert.Equal(t, "email", config.Notification.TemplateTypes["user_registration"].Default.Channels[0])
	assert.Equal(t, "info", config.Notification.TemplateTypes["user_registration"].Default.Severity)
	assert.Equal(t, "templates/email", config.Notification.EmailTemplatesDirPath)
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
			name: "missing server",
			configJSON: `{
				"database": {
					"dsn": "sqlite3:///test.db"
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
			expectedErr: "validation error on field 'Configuration.Server.Host': required",
		},
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
			expectedErr: "validation error on field 'Configuration.JwtToken.AccessExpiryMinutes': required",
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

func TestLoadNotificationConfig(t *testing.T) {
	// Create a temporary directory for test files
	tempDir := t.TempDir()

	// Create a temporary notification config file
	configPath := filepath.Join(tempDir, "notification-config.json")

	// Create a valid notification config JSON
	configJSON := `{
		"template_types": {
			"user_registration": {
				"default": {
					"channels": ["email"],
					"severity": "info"
				},
				"by_status": {
					"pending": {
						"channels": ["email", "sms"],
						"severity": "warning"
					},
					"completed": {
						"channels": ["email"],
						"severity": "success"
					}
				}
			},
			"password_reset": {
				"default": {
					"channels": ["email"],
					"severity": "info"
				}
			}
		},
		"email_templates_dir_path": "templates/email"
	}`

	// Write the config file
	err := os.WriteFile(configPath, []byte(configJSON), 0644)
	require.NoError(t, err, "Failed to write test notification config file")

	// Load the notification configuration
	config, err := loadNotificationConfig(configPath)
	require.NoError(t, err, "loadNotificationConfig() failed")

	// Verify the loaded configuration
	assert.Len(t, config.TemplateTypes, 2)
	assert.Contains(t, config.TemplateTypes, "user_registration")
	assert.Contains(t, config.TemplateTypes, "password_reset")

	// Check user_registration template type
	userRegTemplate := config.TemplateTypes["user_registration"]
	assert.Equal(t, "email", userRegTemplate.Default.Channels[0])
	assert.Equal(t, "info", userRegTemplate.Default.Severity)

	// Check by_status configurations
	assert.Len(t, userRegTemplate.ByStatus, 2)
	assert.Contains(t, userRegTemplate.ByStatus, "pending")
	assert.Contains(t, userRegTemplate.ByStatus, "completed")

	// Check pending status
	pendingStatus := userRegTemplate.ByStatus["pending"]
	assert.Len(t, pendingStatus.Channels, 2)
	assert.Contains(t, pendingStatus.Channels, "email")
	assert.Contains(t, pendingStatus.Channels, "sms")
	assert.Equal(t, "warning", pendingStatus.Severity)

	// Check completed status
	completedStatus := userRegTemplate.ByStatus["completed"]
	assert.Len(t, completedStatus.Channels, 1)
	assert.Equal(t, "email", completedStatus.Channels[0])
	assert.Equal(t, "success", completedStatus.Severity)

	// Check password_reset template type
	passwordResetTemplate := config.TemplateTypes["password_reset"]
	assert.Equal(t, "email", passwordResetTemplate.Default.Channels[0])
	assert.Equal(t, "info", passwordResetTemplate.Default.Severity)
	assert.Empty(t, passwordResetTemplate.ByStatus)

	// Check email templates directory path
	assert.Equal(t, "templates/email", config.EmailTemplatesDirPath)
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
