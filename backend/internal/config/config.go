package config

import (
	"fmt"
	"kubecloud/internal/core/paths"
	"net/url"
	"strings"
	"time"

	"github.com/go-playground/validator"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
	"github.com/xmonader/ewf"
)

type Configuration struct {
	Server                               Server                        `json:"server" validate:"required,dive"`
	Database                             DB                            `json:"database" validate:"required"`
	JwtToken                             JwtToken                      `json:"jwt_token" validate:"required"`
	Admins                               []string                      `json:"admins" validate:"required"`
	MailSender                           MailSender                    `json:"mailSender"`
	Currency                             string                        `json:"currency" default:"usd"`
	StripeSecret                         string                        `json:"stripe_secret" validate:"required"`
	VoucherNameLength                    int                           `json:"voucher_name_length" validate:"required,gt=0" default:"8"`
	VerificationCodeLength               int                           `json:"verification_code_length" validate:"gt=0" default:"4"`
	TermsANDConditions                   TermsANDConditions            `json:"terms_and_conditions"`
	SystemAccount                        GridAccount                   `json:"system_account"`
	DeployerWorkersNum                   int                           `json:"deployer_workers_num" default:"1"`
	Invoice                              InvoiceCompanyData            `json:"invoice"`
	SSH                                  SSHConfig                     `json:"ssh" validate:"required,dive"`
	Redis                                RedisConfig                   `json:"redis" validate:"dive"`
	Debug                                bool                          `json:"debug"`
	DevMode                              bool                          `json:"dev_mode"` // When true, allows empty SendGridKey and uses FakeMailService
	MonitorBalanceIntervalInMinutes      int                           `json:"monitor_balance_interval_in_minutes" validate:"required,gt=0"`
	NotifyAdminsForPendingRecordsInHours int                           `json:"notify_admins_for_pending_records_in_hours" validate:"required,gt=0"`
	ClusterHealthCheckIntervalInHours    int                           `json:"cluster_health_check_interval_in_hours" validate:"gt=0" default:"1"`
	UsersBalanceCheckIntervalInHours     int                           `json:"users_balance_check_interval_in_hours" validate:"gt=0" default:"6"`
	CheckUserDebtIntervalInHours         int                           `json:"check_user_debt_interval_in_hours" validate:"gt=0" default:"48"`
	NodeHealthCheck                      ReservedNodeHealthCheckConfig `json:"node_health_check" validate:"required,dive"`

	Logger    LoggerConfig    `json:"logger"`
	Loki      LokiConfig      `json:"loki"`
	Telemetry TelemetryConfig `json:"telemetry"`
}

type SSHConfig struct {
	PrivateKeyPath string `json:"private_key_path" validate:"required"`
	PublicKeyPath  string `json:"public_key_path" validate:"required"`
}

type RedisConfig struct {
	Hostname string `json:"hostname" validate:"hostname|ip|url"`
	Port     int    `json:"port" validate:"min=1,max=65535"`
	Password string `json:"password"`
	DB       int    `json:"db" validate:"min=0"`
}

// Server struct holds server's information
type Server struct {
	Host string `json:"host" validate:"required,hostname|ip|url" default:"0.0.0.0"`
	Port string `json:"port" validate:"required,numeric" default:"8080"`
}

// DB struct holds database file
type DB struct {
	DSN string `json:"dsn" validate:"required,dsn"`
	// Optional connection pool settings (Postgres)
	MaxOpenConns           int `json:"max_open_conns" validate:"min=0"`
	MaxIdleConns           int `json:"max_idle_conns" validate:"min=0"`
	ConnMaxLifetimeMinutes int `json:"conn_max_lifetime_minutes" validate:"min=0"`
	ConnMaxIdleTimeMinutes int `json:"conn_max_idle_time_minutes" validate:"min=0"`
}

// JWT Token struct holds info required for JWT Tokens
type JwtToken struct {
	Secret              string `json:"secret" validate:"required"`
	AccessExpiryMinutes int    `json:"access_expiry_minutes" validate:"gt=0" default:"60"` // in minutes
	RefreshExpiryHours  int    `json:"refresh_expiry_hours" validate:"gt=0" default:"24"`  // in hours
}

// MailSender struct to hold sender's email, password
type MailSender struct {
	Email               string `json:"email" validate:"required,email"`
	SendGridKey         string `json:"sendgrid_key"` // Required in production. Can be empty in dev_mode to use FakeMailService
	TimeoutMin          int    `json:"timeout" validate:"min=2" default:"120"`
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
	Network  string `json:"network" validate:"required" default:"main"`
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
	MaxSize    int    `json:"max_size" default:"512"` // in MB
	MaxBackups int    `json:"max_backups" default:"12"`
	MaxAgeDays int    `json:"max_age_days" default:"30"` // in days
	Compress   bool   `json:"compress" default:"true"`
}

type LokiConfig struct {
	URL                 string      `json:"url"`
	FlushIntervalSecond int         `json:"flush_interval_second" default:"5"`
	Labels              *LokiLabels `json:"labels,omitempty"`
}

type LokiLabels struct {
	App  string `json:"app,omitempty" default:"myceliumCloud"`
	Env  string `json:"env,omitempty" default:"main"`
	Host string `json:"host,omitempty"`
}

type TelemetryConfig struct {
	OTLPEndpoint string `json:"otlp_endpoint" default:"jaeger:4317"` // gRPC endpoint for OTLP exporter
}

type ReservedNodeHealthCheckConfig struct {
	ReservedNodeHealthCheckIntervalInHours  int `json:"reserved_node_health_check_interval_in_hours" validate:"required,gt=0" default:"1"`
	ReservedNodeHealthCheckTimeoutInMinutes int `json:"reserved_node_health_check_timeout_in_minutes" validate:"required,gt=0" default:"1"`
	ReservedNodeHealthCheckWorkersNum       int `json:"reserved_node_health_check_workers_num" validate:"required,gt=0" default:"10"`
}

var DefaultQueueConfig = ewf.QueueMetadata{
	Name: "chain_operations_queue",
	WorkersDef: ewf.WorkersDefinition{
		Count:        1,
		PollInterval: 1 * time.Second,
		WorkTimeout:  10 * time.Minute,
	},
	QueueOptions: ewf.QueueOptions{
		AutoDelete:  true,
		DeleteAfter: 10 * time.Minute,
		PopTimeout:  1 * time.Second,
	},
}

// LoadConfig load configurations
func LoadConfig() (Configuration, error) {
	var config Configuration

	// Use mapstructure to ensure JSON tags are respected
	decoderConfig := &mapstructure.DecoderConfig{
		TagName:          "json",
		Result:           &config,
		WeaklyTypedInput: true,
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

	// Apply default values from struct tags
	applyDefaultValues(&config)

	// custom validators
	v := validator.New()
	registerConfigValidators(v)

	config.SSH.PrivateKeyPath, err = paths.ExpandPath(config.SSH.PrivateKeyPath)
	if err != nil {
		return Configuration{}, fmt.Errorf("failed to expand SSH private key path: %w", err)
	}

	config.SSH.PublicKeyPath, err = paths.ExpandPath(config.SSH.PublicKeyPath)
	if err != nil {
		return Configuration{}, fmt.Errorf("failed to expand SSH public key path: %w", err)
	}

	config.Logger.LogDir, err = paths.ExpandPath(config.Logger.LogDir)
	if err != nil {
		return Configuration{}, fmt.Errorf("failed to expand log directory path: %w", err)
	}

	if err := v.Struct(config); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			for _, ve := range validationErrors {
				return Configuration{}, fmt.Errorf("validation error on field '%s': %s", ve.Namespace(), ve.Tag())
			}
		}
		return Configuration{}, fmt.Errorf("invalid configuration: %w", err)
	}

	// custom validation: sendGridKey is required when NOT in dev mode
	config.MailSender.SendGridKey = strings.TrimSpace(config.MailSender.SendGridKey)
	if !config.DevMode && config.MailSender.SendGridKey == "" {
		return Configuration{}, fmt.Errorf("sendgrid_key is required when dev_mode is false. Set dev_mode=true to use FakeMailService for development")
	}

	// custom validation: warn if using SQLite in production mode
	if !config.DevMode {
		dsn := strings.TrimSpace(config.Database.DSN)
		u, err := url.Parse(dsn)
		if err != nil {
			return Configuration{}, fmt.Errorf("failed to parse database DSN: %w", err)
		}
		if u.Scheme == "sqlite" || u.Scheme == "sqlite3" {
			return Configuration{}, fmt.Errorf("SQLite not allowed in production. Use PostgreSQL or set dev_mode=true")
		}
	}

	return config, nil
}

func registerConfigValidators(v *validator.Validate) {
	if v == nil {
		return
	}

	// dsn validator for supported schemes
	_ = v.RegisterValidation("dsn", func(fl validator.FieldLevel) bool {
		dsn := strings.TrimSpace(fl.Field().String())
		if dsn == "" {
			return false
		}
		u, err := url.Parse(dsn)
		if err != nil {
			return false
		}
		switch u.Scheme {
		case "postgres":
			return true
		case "sqlite", "sqlite3":
			return strings.TrimSpace(u.Path) != ""
		default:
			return false
		}
	})

	// if MaxOpenConns>0, MaxIdleConns must be <= MaxOpenConns
	v.RegisterStructValidation(func(sl validator.StructLevel) {
		val, ok := sl.Current().Interface().(DB)
		if !ok {
			return
		}
		if val.MaxOpenConns <= 0 {
			return
		}
		if val.MaxIdleConns > val.MaxOpenConns {
			sl.ReportError(val.MaxIdleConns, "MaxIdleConns", "max_idle_conns", "lteMaxOpenConns", "")
		}
	}, DB{})
}

// applyDefaultValues sets default values from struct tags
func applyDefaultValues(config *Configuration) {
	// Server defaults
	if config.Server.Host == "" {
		config.Server.Host = "0.0.0.0"
	}
	if config.Server.Port == "" {
		config.Server.Port = "8080"
	}

	// Database defaults
	if config.Database.MaxOpenConns == 0 {
		config.Database.MaxOpenConns = 25
	}
	if config.Database.MaxIdleConns == 0 {
		config.Database.MaxIdleConns = 25
	}
	if config.Database.ConnMaxLifetimeMinutes == 0 {
		config.Database.ConnMaxLifetimeMinutes = 60
	}
	if config.Database.ConnMaxIdleTimeMinutes == 0 {
		config.Database.ConnMaxIdleTimeMinutes = 30
	}

	// Currency default
	if config.Currency == "" {
		config.Currency = "usd"
	}

	// VoucherNameLength default
	if config.VoucherNameLength == 0 {
		config.VoucherNameLength = 8
	}

	// VerificationCodeLength default
	if config.VerificationCodeLength == 0 {
		config.VerificationCodeLength = 4
	}

	// DeployerWorkersNum default
	if config.DeployerWorkersNum == 0 {
		config.DeployerWorkersNum = 1
	}

	// ClusterHealthCheckIntervalInHours default
	if config.ClusterHealthCheckIntervalInHours == 0 {
		config.ClusterHealthCheckIntervalInHours = 1
	}

	// JwtToken defaults
	if config.JwtToken.AccessExpiryMinutes == 0 {
		config.JwtToken.AccessExpiryMinutes = 60
	}
	if config.JwtToken.RefreshExpiryHours == 0 {
		config.JwtToken.RefreshExpiryHours = 24
	}

	// MailSender defaults
	if config.MailSender.TimeoutMin == 0 {
		config.MailSender.TimeoutMin = 120
	}

	if config.MailSender.MaxAttachmentSizeMB == 0 {
		config.MailSender.MaxAttachmentSizeMB = 25
	}

	if config.MailSender.MaxConcurrentSends == 0 {
		config.MailSender.MaxConcurrentSends = 10
	}

	// SystemAccount defaults
	if config.SystemAccount.Network == "" {
		config.SystemAccount.Network = "main"
	}

	// Logger defaults
	if config.Logger.MaxSize == 0 {
		config.Logger.MaxSize = 512
	}
	if config.Logger.MaxBackups == 0 {
		config.Logger.MaxBackups = 12
	}
	if config.Logger.MaxAgeDays == 0 {
		config.Logger.MaxAgeDays = 30
	}

	// Loki defaults
	if config.Loki.FlushIntervalSecond == 0 {
		config.Loki.FlushIntervalSecond = 5
	}

	// LokiLabels defaults
	if config.Loki.Labels != nil {
		if config.Loki.Labels.App == "" {
			config.Loki.Labels.App = "myceliumCloud"
		}
		if config.Loki.Labels.Env == "" {
			config.Loki.Labels.Env = "main"
		}
	}

	// NodeHealthCheck defaults
	if config.NodeHealthCheck.ReservedNodeHealthCheckIntervalInHours == 0 {
		config.NodeHealthCheck.ReservedNodeHealthCheckIntervalInHours = 1
	}
	if config.NodeHealthCheck.ReservedNodeHealthCheckTimeoutInMinutes == 0 {
		config.NodeHealthCheck.ReservedNodeHealthCheckTimeoutInMinutes = 1
	}
	if config.NodeHealthCheck.ReservedNodeHealthCheckWorkersNum == 0 {
		config.NodeHealthCheck.ReservedNodeHealthCheckWorkersNum = 10
	}

	if config.MonitorBalanceIntervalInMinutes == 0 {
		config.MonitorBalanceIntervalInMinutes = 120
	}
	if config.NotifyAdminsForPendingRecordsInHours == 0 {
		config.NotifyAdminsForPendingRecordsInHours = 24
	}

	if config.Telemetry.OTLPEndpoint == "" {
		config.Telemetry.OTLPEndpoint = "jaeger:4317"
	}

	if config.UsersBalanceCheckIntervalInHours == 0 {
		config.UsersBalanceCheckIntervalInHours = 6
	}

	if config.CheckUserDebtIntervalInHours == 0 {
		config.CheckUserDebtIntervalInHours = 48
	}
}
