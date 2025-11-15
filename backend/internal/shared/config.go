package shared

import (
	"fmt"
	"kubecloud/internal/shared/path"
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
	Currency                             string                        `json:"currency" validate:"required"`
	StripeSecret                         string                        `json:"stripe_secret" validate:"required"`
	VoucherNameLength                    int                           `json:"voucher_name_length"  validate:"required,gt=0"`
	VerificationCodeLength               int                           `json:"verification_code_length"  validate:"gt=0" default:"4"`
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
	ClusterHealthCheckIntervalInHours    int                           `json:"cluster_health_check_interval_in_hours" validate:"required,gt=0" default:"1"`
	NodeHealthCheck                      ReservedNodeHealthCheckConfig `json:"node_health_check" validate:"required,dive"`

	Logger LoggerConfig `json:"logger"`
	Loki   LokiConfig   `json:"loki"`
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
	Host string `json:"host" validate:"required,hostname|ip|url"`
	Port string `json:"port" validate:"required,numeric"`
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
	AccessExpiryMinutes int    `json:"access_expiry_minutes" validate:"required,gt=0"` // in minutes
	RefreshExpiryHours  int    `json:"refresh_expiry_hours" validate:"required,gt=0"`  // in hours
}

// MailSender struct to hold sender's email, password
type MailSender struct {
	Email               string `json:"email" validate:"required,email"`
	SendGridKey         string `json:"sendgrid_key"` // Required in production. Can be empty in dev_mode to use FakeMailService
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

	// custom validators
	v := validator.New()
	registerConfigValidators(v)

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

	config.SSH.PrivateKeyPath, err = path.ExpandPath(config.SSH.PrivateKeyPath)
	if err != nil {
		return Configuration{}, fmt.Errorf("failed to expand SSH private key path: %w", err)
	}

	config.SSH.PublicKeyPath, err = path.ExpandPath(config.SSH.PublicKeyPath)
	if err != nil {
		return Configuration{}, fmt.Errorf("failed to expand SSH public key path: %w", err)
	}

	config.Logger.LogDir, err = path.ExpandPath(config.Logger.LogDir)
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
