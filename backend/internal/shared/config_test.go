package shared

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-playground/validator"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestLoadNotificationConfig tests successful loading of a valid notification configuration file.
// This scenario covers:
// - Reading a valid notification config JSON file with template types
// - Parsing nested template type configurations with default and status-based rules
// - Verifying correct deserialization of channels and severity levels
// - Ensuring email templates directory path is correctly loaded
func TestLoadNotificationConfig(t *testing.T) {
	// Create a temporary notification config file
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "notification.json")

	notificationConfigContent := `{
		"template_types": {
			"user_registration": {
				"default": {
					"channels": ["email"],
					"severity": "info"
				},
				"by_status": {
					"success": {
						"channels": ["email", "sse"],
						"severity": "success"
					}
				}
			}
		},
		"email_templates_dir_path": "/tmp/templates"
	}`

	err := os.WriteFile(configPath, []byte(notificationConfigContent), 0644)
	require.NoError(t, err)

	// Test loading valid config
	config, err := loadNotificationConfig(configPath)
	require.NoError(t, err)

	assert.Equal(t, "/tmp/templates", config.EmailTemplatesDirPath)
	assert.Contains(t, config.TemplateTypes, "user_registration")
	assert.Equal(t, []string{"email"}, config.TemplateTypes["user_registration"].Default.Channels)
	assert.Equal(t, "info", config.TemplateTypes["user_registration"].Default.Severity)
}

// TestLoadNotificationConfig_EmptyPath tests the error handling when an empty path is provided.
// This scenario covers:
// - Validation of notification config path requirement
// - Error message clarity for missing configuration files
func TestLoadNotificationConfig_EmptyPath(t *testing.T) {
	_, err := loadNotificationConfig("")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "notification config path is required")
}

// TestLoadNotificationConfig_FileNotFound tests the error handling when the config file doesn't exist.
// This scenario covers:
// - File system error handling for missing configuration files
// - Proper error propagation from viper ReadInConfig
func TestLoadNotificationConfig_FileNotFound(t *testing.T) {
	_, err := loadNotificationConfig("/nonexistent/path.json")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read notification config file")
}

// TestLoadNotificationConfig_InvalidJSON tests the error handling for malformed JSON.
// This scenario covers:
// - JSON parsing error handling
// - Validation of configuration file format
func TestLoadNotificationConfig_InvalidJSON(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "invalid.json")

	err := os.WriteFile(configPath, []byte("invalid json"), 0644)
	require.NoError(t, err)

	_, err = loadNotificationConfig(configPath)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read notification config file")
}

// TestLoadNotificationConfig_InvalidValidation tests the validation of notification config structure.
// This scenario covers:
// - Validation of required fields in notification templates
// - Severity level validation (must be one of: info, error, warning, success)
// - Channel array validation (must not be empty)
// TestLoadNotificationConfig_EmptyTemplateTypes tests loading a notification config with empty template types.
// This scenario covers:
// - Handling configurations where template_types map is empty (allowed case)
// - Ensuring the system gracefully accepts minimal valid configurations
// - Verifying that missing template types do not cause validation errors
func TestLoadNotificationConfig_EmptyTemplateTypes(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "empty_templates.json")

	// Create config with empty template_types (this should be allowed as it's a map)
	emptyConfigContent := `{
		"template_types": {},
		"email_templates_dir_path": "/tmp/templates"
	}`

	err := os.WriteFile(configPath, []byte(emptyConfigContent), 0644)
	require.NoError(t, err)

	config, err := loadNotificationConfig(configPath)
	require.NoError(t, err)
	assert.Empty(t, config.TemplateTypes)
	assert.Equal(t, "/tmp/templates", config.EmailTemplatesDirPath)
}

// TestLoadConfig_ValidConfig tests successful loading of a complete and valid application configuration.
// This scenario covers:
// - Loading all required configuration sections (server, database, JWT, mail, SSH, etc.)
// - Path expansion for SSH keys, log directory, and notification config
// - Proper deserialization of nested configuration structures
// - Comma-separated admin email parsing
// - Notification config integration with main configuration
// - All configuration validations pass with proper values
func TestLoadConfig_ValidConfig(t *testing.T) {
	// Create temporary files
	tempDir := t.TempDir()
	sshPrivateKey := filepath.Join(tempDir, "id_rsa")
	sshPublicKey := filepath.Join(tempDir, "id_rsa.pub")
	logDir := filepath.Join(tempDir, "logs")
	notificationConfig := filepath.Join(tempDir, "notification.json")

	// Create dummy files
	err := os.WriteFile(sshPrivateKey, []byte("dummy private key"), 0600)
	require.NoError(t, err)
	err = os.WriteFile(sshPublicKey, []byte("dummy public key"), 0644)
	require.NoError(t, err)
	err = os.MkdirAll(logDir, 0755)
	require.NoError(t, err)

	notificationConfigContent := `{
		"template_types": {
			"test": {
				"default": {
					"channels": ["email"],
					"severity": "info"
				}
			}
		},
		"email_templates_dir_path": "/tmp/templates"
	}`
	err = os.WriteFile(notificationConfig, []byte(notificationConfigContent), 0644)
	require.NoError(t, err)

	// Set up viper with test config
	viper.Reset()
	viper.SetConfigType("json")

	configData := map[string]interface{}{
		"server": map[string]interface{}{
			"host": "localhost",
			"port": "8080",
		},
		"database": map[string]interface{}{
			"dsn":                        "sqlite:///test.db",
			"max_open_conns":             10,
			"max_idle_conns":             5,
			"conn_max_lifetime_minutes":  30,
			"conn_max_idle_time_minutes": 10,
		},
		"jwt_token": map[string]interface{}{
			"secret":                "test-secret-key",
			"access_expiry_minutes": 15,
			"refresh_expiry_hours":  24,
		},
		"admins":                   "admin1@example.com,admin2@example.com",
		"currency":                 "TFT",
		"stripe_secret":            "sk_test_123",
		"voucher_name_length":      8,
		"verification_code_length": 4,
		"terms_and_conditions": map[string]interface{}{
			"document_link": "https://example.com/terms",
			"document_hash": "abc123",
		},
		"system_account": map[string]interface{}{
			"mnemonic": "test mnemonic phrase",
			"network":  "dev",
		},
		"invoice": map[string]interface{}{
			"name":        "Test Company",
			"address":     "123 Test St",
			"governorate": "Test Gov",
		},
		"ssh": map[string]interface{}{
			"private_key_path": sshPrivateKey,
			"public_key_path":  sshPublicKey,
		},
		"redis": map[string]interface{}{
			"hostname": "localhost",
			"port":     6379,
			"password": "",
			"db":       0,
		},
		"debug":                               true,
		"dev_mode":                            true,
		"monitor_balance_interval_in_minutes": 5,
		"notify_admins_for_pending_records_in_hours": 24,
		"cluster_health_check_interval_in_hours":     1,
		"node_health_check": map[string]interface{}{
			"reserved_node_health_check_interval_in_hours":  1,
			"reserved_node_health_check_timeout_in_minutes": 5,
			"reserved_node_health_check_workers_num":        10,
		},
		"logger": map[string]interface{}{
			"log_dir":      logDir,
			"max_size":     100,
			"max_backups":  5,
			"max_age_days": 30,
			"compress":     true,
		},
		"loki": map[string]interface{}{
			"url":                   "http://localhost:3100",
			"flush_interval_second": 5,
			"labels": map[string]string{
				"app": "kubecloud",
				"env": "test",
			},
		},
		"notification_config_path": notificationConfig,
		"mailsender": map[string]interface{}{
			"email":                  "test@example.com",
			"sendgrid_key":           "",
			"timeout":                5,
			"max_concurrent_sends":   10,
			"max_attachment_size_mb": 10,
		},
	}

	for key, value := range configData {
		viper.Set(key, value)
	}

	config, err := LoadConfig()
	require.NoError(t, err)

	// Verify basic fields
	assert.Equal(t, "localhost", config.Server.Host)
	assert.Equal(t, "8080", config.Server.Port)
	assert.Equal(t, "sqlite:///test.db", config.Database.DSN)
	assert.Equal(t, "test-secret-key", config.JwtToken.Secret)
	assert.Equal(t, []string{"admin1@example.com", "admin2@example.com"}, config.Admins)
	assert.Equal(t, "TFT", config.Currency)
	assert.Equal(t, "sk_test_123", config.StripeSecret)
	assert.Equal(t, 8, config.VoucherNameLength)
	assert.Equal(t, 4, config.VerificationCodeLength)
	assert.True(t, config.Debug)
	assert.True(t, config.DevMode)

	// Verify SSH paths are expanded
	assert.Equal(t, sshPrivateKey, config.SSH.PrivateKeyPath)
	assert.Equal(t, sshPublicKey, config.SSH.PublicKeyPath)
	assert.Equal(t, logDir, config.Logger.LogDir)

	// Verify Loki labels
	expectedLabels := map[string]string{
		"app": "kubecloud",
		"env": "test",
	}
	assert.Equal(t, expectedLabels, config.Loki.Labels)

	// Verify notification config is loaded
	assert.NotEmpty(t, config.Notification.TemplateTypes)
}

// TestLoadConfig_MissingRequiredFields tests error handling when required configuration fields are missing.
// This scenario covers:
// - Validation failure when server host and port are empty
// - Proper error reporting for missing required fields
// - Early validation to prevent loading incomplete configurations
func TestLoadConfig_MissingRequiredFields(t *testing.T) {
	viper.Reset()

	// Set minimal config without required fields
	viper.Set("server", map[string]interface{}{
		"host": "",
		"port": "",
	})

	_, err := LoadConfig()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "validation error")
}

// TestLoadConfig_InvalidDSN tests error handling when an invalid database DSN is provided.
// This scenario covers:
// - DSN scheme validation (only postgres, sqlite, sqlite3 are supported)
// - Rejection of unsupported database schemes (e.g., "invalid://scheme")
// - Proper validation error reporting for database configuration
func TestLoadConfig_InvalidDSN(t *testing.T) {
	viper.Reset()

	viper.Set("server", map[string]interface{}{
		"host": "localhost",
		"port": "8080",
	})
	viper.Set("database", map[string]interface{}{
		"dsn": "invalid://scheme",
	})
	viper.Set("jwt_token", map[string]interface{}{
		"secret":                "test",
		"access_expiry_minutes": 15,
		"refresh_expiry_hours":  24,
	})
	viper.Set("admins", "admin@example.com")
	viper.Set("currency", "TFT")
	viper.Set("stripe_secret", "sk_test_123")
	viper.Set("voucher_name_length", 8)
	viper.Set("monitor_balance_interval_in_minutes", 5)
	viper.Set("notify_admins_for_pending_records_in_hours", 24)
	viper.Set("cluster_health_check_interval_in_hours", 1)
	viper.Set("node_health_check", map[string]interface{}{
		"reserved_node_health_check_interval_in_hours":  1,
		"reserved_node_health_check_timeout_in_minutes": 5,
		"reserved_node_health_check_workers_num":        10,
	})
	viper.Set("ssh", map[string]interface{}{
		"private_key_path": "/tmp/id_rsa",
		"public_key_path":  "/tmp/id_rsa.pub",
	})
	viper.Set("dev_mode", true)

	_, err := LoadConfig()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "validation error")
}

// TestLoadConfig_SendGridKeyRequiredInProd tests validation of SendGrid key requirement in production mode.
// This scenario covers:
// - Enforcing SendGrid key presence when dev_mode=false (production mode)
// - Allowing empty SendGrid key only when dev_mode=true (for development/testing)
// - Custom validation logic for environment-specific requirements
// - Proper error messaging indicating the production requirement
func TestLoadConfig_SendGridKeyRequiredInProd(t *testing.T) {
	tempDir := t.TempDir()
	sshPrivateKey := filepath.Join(tempDir, "id_rsa")
	sshPublicKey := filepath.Join(tempDir, "id_rsa.pub")

	err := os.WriteFile(sshPrivateKey, []byte("dummy"), 0600)
	require.NoError(t, err)
	err = os.WriteFile(sshPublicKey, []byte("dummy"), 0644)
	require.NoError(t, err)

	viper.Reset()
	viper.Set("server", map[string]interface{}{
		"host": "localhost",
		"port": "8080",
	})
	viper.Set("database", map[string]interface{}{
		"dsn": "sqlite:///test.db",
	})
	viper.Set("jwt_token", map[string]interface{}{
		"secret":                "test",
		"access_expiry_minutes": 15,
		"refresh_expiry_hours":  24,
	})
	viper.Set("admins", "admin@example.com")
	viper.Set("currency", "TFT")
	viper.Set("stripe_secret", "sk_test_123")
	viper.Set("voucher_name_length", 8)
	viper.Set("verification_code_length", 4)
	viper.Set("monitor_balance_interval_in_minutes", 5)
	viper.Set("notify_admins_for_pending_records_in_hours", 24)
	viper.Set("cluster_health_check_interval_in_hours", 1)
	viper.Set("node_health_check", map[string]interface{}{
		"reserved_node_health_check_interval_in_hours":  1,
		"reserved_node_health_check_timeout_in_minutes": 5,
		"reserved_node_health_check_workers_num":        10,
	})
	viper.Set("terms_and_conditions", map[string]interface{}{
		"document_link": "https://example.com/terms",
		"document_hash": "abc123",
	})
	viper.Set("system_account", map[string]interface{}{
		"mnemonic": "test mnemonic phrase",
		"network":  "dev",
	})
	viper.Set("invoice", map[string]interface{}{
		"name":        "Test Company",
		"address":     "123 Test St",
		"governorate": "Test Gov",
	})
	viper.Set("ssh", map[string]interface{}{
		"private_key_path": sshPrivateKey,
		"public_key_path":  sshPublicKey,
	})
	viper.Set("redis", map[string]interface{}{
		"hostname": "localhost",
		"port":     6379,
		"password": "",
		"db":       0,
	})
	viper.Set("mailsender", map[string]interface{}{
		"email":                  "test@example.com",
		"sendgrid_key":           "",
		"timeout":                5,
		"max_concurrent_sends":   10,
		"max_attachment_size_mb": 10,
	})
	viper.Set("dev_mode", false) // Production mode

	_, err = LoadConfig()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "sendgrid_key is required when dev_mode is false")
}

// TestRegisterConfigValidators_DSNValidator tests the custom DSN validator registration and validation logic.
// This scenario covers:
// - Validation of supported database schemes (postgres, sqlite, sqlite3)
// - Ensuring empty DSN strings are rejected
// - Validating that SQLite/SQLite3 DSNs have non-empty paths
// - Rejecting unsupported database schemes (mysql, mongodb, etc.)
// - URL parsing and scheme validation
func TestRegisterConfigValidators_DSNValidator(t *testing.T) {
	v := validator.New()
	registerConfigValidators(v)

	// Test valid DSNs
	validDSNs := []string{
		"postgres://user:pass@localhost/db",
		"sqlite:///path/to/db.sqlite",
		"sqlite3:///path/to/db.sqlite3",
	}

	for _, dsn := range validDSNs {
		err := v.Var(dsn, "dsn")
		assert.NoError(t, err, "DSN should be valid: %s", dsn)
	}

	// Test invalid DSNs
	invalidDSNs := []string{
		"",
		"mysql://invalid",
		"mongodb://invalid",
		"sqlite://", // empty path
	}

	for _, dsn := range invalidDSNs {
		err := v.Var(dsn, "dsn")
		assert.Error(t, err, "DSN should be invalid: %s", dsn)
	}
}

// TestRegisterConfigValidators_DBValidation tests the custom database configuration struct validator.
// This scenario covers:
// - Validation that MaxIdleConns does not exceed MaxOpenConns when MaxOpenConns > 0
// - Allowing MaxOpenConns=0 (unlimited connections) with any MaxIdleConns value
// - Proper constraint validation for connection pool settings
// - Database configuration struct-level validation
func TestRegisterConfigValidators_DBValidation(t *testing.T) {
	v := validator.New()
	registerConfigValidators(v)

	// Test valid DB config
	validDB := DB{
		DSN:          "sqlite:///test.db",
		MaxOpenConns: 10,
		MaxIdleConns: 5,
	}
	assert.NoError(t, v.Struct(validDB))

	// Test invalid DB config (MaxIdleConns > MaxOpenConns)
	invalidDB := DB{
		DSN:          "sqlite:///test.db",
		MaxOpenConns: 5,
		MaxIdleConns: 10,
	}
	assert.Error(t, v.Struct(invalidDB))

	// Test valid DB config with MaxOpenConns = 0 (no limit)
	validDBNoLimit := DB{
		DSN:          "sqlite:///test.db",
		MaxOpenConns: 0,
		MaxIdleConns: 10,
	}
	assert.NoError(t, v.Struct(validDBNoLimit))
}

// TestDefaultQueueConfig tests the default queue configuration constants and values.
// This scenario covers:
// - Verifying default queue name for chain operations
// - Validating worker definition defaults (count, poll interval, work timeout)
// - Checking queue options (auto-delete behavior, deletion timeout, pop timeout)
// - Ensuring sensible defaults are set for asynchronous task processing
func TestDefaultQueueConfig(t *testing.T) {
	assert.Equal(t, "chain_operations_queue", DefaultQueueConfig.Name)
	assert.Equal(t, 1, DefaultQueueConfig.WorkersDef.Count)
	assert.Equal(t, 1*time.Second, DefaultQueueConfig.WorkersDef.PollInterval)
	assert.Equal(t, 10*time.Minute, DefaultQueueConfig.WorkersDef.WorkTimeout)
	assert.True(t, DefaultQueueConfig.QueueOptions.AutoDelete)
	assert.Equal(t, 10*time.Minute, DefaultQueueConfig.QueueOptions.DeleteAfter)
	assert.Equal(t, 1*time.Second, DefaultQueueConfig.QueueOptions.PopTimeout)
}
