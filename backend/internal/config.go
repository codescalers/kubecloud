package internal

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/go-playground/validator"
)

// Configuration struct holds all configs for the app
type Configuration struct {
	Server   Server   `json:"server" validate:"required,dive"`
	Database DB       `json:"database" validate:"required"`
	JWT      JwtToken `json:"token" validate:"required"`
	Admins   []string `json:"admins"`
}

type Server struct {
	Host string `json:"host" validate:"required,hostname|ip"`
	Port string `json:"port" validate:"required,numeric"`
}

type DB struct {
	File string `json:"file" validate:"required"`
}

type JwtToken struct {
	Secret             string `json:"secret" validate:"required"`
	AccessTokenExpiryMinutes  int    `json:"access_token_expiry_minutes" validate:"required,gt=0"`  // in minutes
	RefreshTokenExpiryHours int    `json:"refresh_token_expiry_hours" validate:"required,gt=0"` // in hours
}

// ReadConfFile read configurations of json file
func ReadConfFile(path string) (Configuration, error) {
	config := Configuration{}
	file, err := os.Open(path)
	if err != nil {
		return Configuration{}, fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	dec := json.NewDecoder(file)
	if err := dec.Decode(&config); err != nil {
		return Configuration{}, fmt.Errorf("failed to load config: %w", err)
	}

	validate := validator.New()
	if err := validate.Struct(config); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			for _, ve := range validationErrors {
				fmt.Printf("Validation error on field '%s': %s\n", ve.Namespace(), ve.Tag())
			}
		}
		return Configuration{}, fmt.Errorf("invalid configuration: %w", err)
	}

	return config, nil
}
