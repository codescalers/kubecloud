package internal

import (
	"fmt"
	"kubecloud/internal/logger"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/go-playground/validator"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

type Configuration struct {
	Server                               Server             `json:"server" validate:"required,dive"`
	Database                             DB                 `json:"database" validate:"required"`
	JwtToken                             JwtToken           `json:"jwt_token" validate:"required"`
	Admins                               []string           `json:"admins" validate:"required"`
	MailSender                           MailSender         `json:"mailSender"`
	Currency                             string             `json:"currency" validate:"required"`
	StripeSecret                         string             `json:"stripe_secret" validate:"required"`
	VoucherNameLength                    int                `json:"voucher_name_length"  validate:"required,gt=0"`
	GridProxyURL                         string             `json:"gridproxy_url" validate:"required"`
	TFChainURL                           string             `json:"tfchain_url" validate:"required"`
	TermsANDConditions                   TermsANDConditions `json:"terms_and_conditions"`
	ActivationServiceURL                 string             `json:"activation_service_url" validate:"required"`
	GraphqlURL                           string             `json:"graphql_url" validate:"required"`
	FiresquidURL                         string             `json:"firesquid_url" validate:"required"`
	SystemAccount                        GridAccount        `json:"system_account"`
	Redis                                Redis              `json:"redis" validate:"required,dive"`
	DeployerWorkersNum                   int                `json:"deployer_workers_num" default:"1"`
	Invoice                              InvoiceCompanyData `json:"invoice"`
	SSH                                  SSHConfig          `json:"ssh" validate:"required,dive"`
	Debug                                bool               `json:"debug"`
	MonitorBalanceIntervalInMinutes      int                `json:"monitor_balance_interval_in_minutes" validate:"required,gt=0"`
	NotifyAdminsForPendingRecordsInHours int                `json:"notify_admins_for_pending_records_in_hours" validate:"required,gt=0"`
	ClusterHealthCheckIntervalInHours    int                `json:"cluster_health_check_interval_in_hours" validate:"required,gt=0" default:"6"`

	// KYC Verifier config
	KYCVerifierAPIURL  string `json:"kyc_verifier_api_url" validate:"required,url"`
	KYCChallengeDomain string `json:"kyc_challenge_domain" validate:"required"`

	Logger LoggerConfig `json:"logger"`
	Loki   LokiConfig   `json:"loki"`

	// Notification configuration
	NotificationConfigPath string             `json:"notification_config_path"`
	Notification           NotificationConfig `json:"-"`
}

type SSHConfig struct {
	PrivateKeyPath string `json:"private_key_path" validate:"required"`
	PublicKeyPath  string `json:"public_key_path" validate:"required"`
}

// Server struct holds server's information
type Server struct {
	Host string `json:"host" validate:"required,hostname|ip|url"`
	Port string `json:"port" validate:"required,numeric"`
}

// DB struct holds database file
type DB struct {
	File string `json:"file" validate:"required"`
}

// JWT Token struct holds info required for JWT Tokens
type JwtToken struct {
	Secret              string `json:"secret" validate:"required"`
	AccessExpiryMinutes int    `json:"access_expiry_minutes" validate:"required,gt=0"` // in minutes
	RefreshExpiryHours  int    `json:"refresh_expiry_hours" validate:"required,gt=0"`  // in hours
}

// MailSender struct to hold sender's email, password
type MailSender struct {
	Email               string `json:"email" validate:"required,email"`
	SendGridKey         string `json:"sendgrid_key" validate:"required"`
	TimeoutMin          int    `json:"timeout" validate:"min=2"`
	MaxConcurrentSends  int    `json:"max_concurrent_sends" validate:"min=1"`
	MaxAttachmentSizeMB int64  `json:"max_attachment_size_mb" validate:"min=1"`
}

// TermsANDConditions holds required data for accepting terms and conditions
type TermsANDConditions struct {
	DocumentLink string `json:"document_link" validate:"required"`
	DocumentHash string `json:"document_hash" validate:"required"`
}

// GridAccount holds data for system's account
type GridAccount struct {
	Mnemonic string `json:"mnemonic" validate:"required"`
	Network  string `json:"network" validate:"required"`
}

// Redis struct holds Redis connection information
type Redis struct {
	Host     string `json:"host" validate:"required"`
	Port     int    `json:"port" validate:"required,min=1,max=65535"`
	Password string `json:"password"`
	DB       int    `json:"db" validate:"min=0"`
}

// Invoice struct holds needed data for invoice file
type InvoiceCompanyData struct {
	Name        string `json:"name" validate:"required"`
	Address     string `json:"address" validate:"required"`
	Governorate string `json:"governorate" validate:"required"`
}

// Configuration struct holds all configs for the app
type LoggerConfig struct {
	LogDir     string `json:"log_dir"`
	MaxSize    int    `json:"max_size"` // in MB
	MaxBackups int    `json:"max_backups"`
	MaxAgeDays int    `json:"max_age_days"` // in days
	Compress   bool   `json:"compress"`
}

type LokiConfig struct {
	URL                 string            `json:"url"`
	FlushIntervalSecond int               `json:"flush_interval_second"`
	Labels              map[string]string `json:"labels"`
}

// NotificationConfig holds notification template type rules
type NotificationConfig struct {
	TemplateTypes         map[string]NotificationTemplateTypeConfig `json:"template_types"`
	EmailTemplatesDirPath string                                    `json:"email_templates_dir_path"`
}

// ChannelRuleConfig represents channel and severity rules for notification template type
type ChannelRuleConfig struct {
	Channels []string `json:"channels" validate:"required,dive,min=1"`
	Severity string   `json:"severity" validate:"required,oneof=info error warning success"`
}

// NotificationTemplateTypeConfig represents a notification template type configuration
type NotificationTemplateTypeConfig struct {
	Default  ChannelRuleConfig            `json:"default" validate:"required"`
	ByStatus map[string]ChannelRuleConfig `json:"by_status"`
}

// LoadNotificationConfig loads notification configuration from a separate file
func loadNotificationConfig(configPath string) (NotificationConfig, error) {
	var notificationConfig NotificationConfig

	if configPath == "" {
		return NotificationConfig{}, fmt.Errorf("notification config path is required")
	}

	notificationViper := viper.New()
	notificationViper.SetConfigFile(configPath)
	notificationViper.SetConfigType("json")

	if err := notificationViper.ReadInConfig(); err != nil {
		return NotificationConfig{}, fmt.Errorf("failed to read notification config file: %w", err)
	}

	// Use mapstructure to decode the config
	decoderConfig := &mapstructure.DecoderConfig{
		TagName: "json",
		Result:  &notificationConfig,
	}

	decoder, err := mapstructure.NewDecoder(decoderConfig)
	if err != nil {
		return NotificationConfig{}, fmt.Errorf("failed to create decoder: %w", err)
	}

	if err := decoder.Decode(notificationViper.AllSettings()); err != nil {
		return NotificationConfig{}, fmt.Errorf("unable to decode notification config: %w", err)
	}

	v := validator.New()
	if err := v.Struct(notificationConfig); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			ve := validationErrors[0]
			return NotificationConfig{}, fmt.Errorf("notification config validation error on field '%s': %s", ve.Namespace(), ve.Tag())
		}
		return NotificationConfig{}, fmt.Errorf("invalid notification config: %w", err)
	}

	return notificationConfig, nil
}

// LoadConfig load configurations
func LoadConfig() (Configuration, error) {
	var config Configuration

	// Use mapstructure to ensure JSON tags are respected
	decoderConfig := &mapstructure.DecoderConfig{
		TagName: "json",
		Result:  &config,
	}

	// convert comma-separated string of admins into a slice
	adminsRaw := viper.GetString("admins")
	if adminsRaw != "" {
		viper.Set("admins", strings.Split(adminsRaw, ","))
	}

	decoder, err := mapstructure.NewDecoder(decoderConfig)
	if err != nil {
		return Configuration{}, fmt.Errorf("failed to create decoder: %w", err)
	}

	if err := decoder.Decode(viper.AllSettings()); err != nil {
		return Configuration{}, fmt.Errorf("unable to decode into struct, %w", err)
	}

	if labelsRaw := viper.GetString("loki.labels"); labelsRaw != "" {
		parsed := make(map[string]string)
		pairs := strings.Split(labelsRaw, ",")
		for _, p := range pairs {
			if kv := strings.SplitN(p, "=", 2); len(kv) == 2 {
				parsed[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
			}
		}
		config.Loki.Labels = parsed
	}

	config.Database.File, err = expandPath(config.Database.File)
	if err != nil {
		return Configuration{}, fmt.Errorf("failed to expand database file path: %w", err)
	}

	config.SSH.PrivateKeyPath, err = expandPath(config.SSH.PrivateKeyPath)
	if err != nil {
		return Configuration{}, fmt.Errorf("failed to expand SSH private key path: %w", err)
	}

	config.SSH.PublicKeyPath, err = expandPath(config.SSH.PublicKeyPath)
	if err != nil {
		return Configuration{}, fmt.Errorf("failed to expand SSH public key path: %w", err)
	}

	config.Logger.LogDir, err = expandPath(config.Logger.LogDir)
	if err != nil {
		return Configuration{}, fmt.Errorf("failed to expand log directory path: %w", err)
	}

	notificationFilePath, err := expandPath(config.NotificationConfigPath)
	if err != nil {
		return Configuration{}, fmt.Errorf("failed to expand notification config path: %w", err)
	}

	config.Notification, err = loadNotificationConfig(notificationFilePath)
	if err != nil {
		logger.GetLogger().Error().Err(err).Msg("Failed to load notification config")
		config.Notification = NotificationConfig{}
	}

	validate := validator.New()
	if err := validate.Struct(config); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			for _, ve := range validationErrors {
				return Configuration{}, fmt.Errorf("validation error on field '%s': %s", ve.Namespace(), ve.Tag())
			}
		}
		return Configuration{}, fmt.Errorf("invalid configuration: %w", err)
	}

	return config, nil
}

func expandPath(path string) (string, error) {
	if path == "" {
		return os.Getwd()
	}

	path = os.ExpandEnv(path)
	var err error
	if strings.HasPrefix(path, "~") {
		path, err = expandTilde(path)
		if err != nil {
			return "", fmt.Errorf("failed to expand tilde: %w", err)
		}
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path: %w", err)
	}

	return filepath.Clean(absPath), nil
}

func expandTilde(path string) (string, error) {

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	if path == "~" {
		return homeDir, nil
	}
	if strings.HasPrefix(path, "~/") {
		return filepath.Join(homeDir, path[2:]), nil
	}
	parts := strings.SplitN(path, "/", 2)
	username := parts[0][1:]

	user, err := user.Lookup(username)
	if err != nil {
		return "", fmt.Errorf("user %s not found: %w", username, err)
	}

	if len(parts) == 1 {
		return user.HomeDir, nil
	}
	return filepath.Join(user.HomeDir, parts[1]), nil

}
