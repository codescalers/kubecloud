package internal

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

// Configuration struct holds all configs for the app
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

	// KYC Verifier config
	KYCVerifierAPIURL  string `json:"kyc_verifier_api_url" validate:"required,url"`
	KYCChallengeDomain string `json:"kyc_challenge_domain" validate:"required"`
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
	Timeout             int    `json:"timeout" validate:"min=30"`
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
