package logger

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/natefinch/lumberjack"
	"github.com/rs/zerolog"
)

// LoggerInstance holds the singleton logger instance
type LoggerInstance struct {
	logger     zerolog.Logger
	lokiWriter *LokiWriter
}

var (
	instance *LoggerInstance
	once     sync.Once
)

// LoggerConfig holds the configuration for the logger
type LoggerConfig struct {
	LogDir     string `json:"log_dir"`
	MaxSize    int    `json:"max_size"`
	MaxBackups int    `json:"max_backups"`
	MaxAge     int    `json:"max_age_days"`
	Compress   bool   `json:"compress"`
}

// LokiConfig holds the configuration for Loki logging
type LokiConfig struct {
	URL           string            `json:"url"`
	FlushInterval time.Duration     `json:"flush_interval"`
	Labels        map[string]string `json:"labels"`
}

// InitLogger initializes the singleton logger with file and console output
func InitLogger(config LoggerConfig, lokiConfig *LokiConfig, debug bool) error {
	logDir := config.LogDir
	if logDir == "" {
		logDir = "logs"
	}
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("failed to create logs directory: %w", err)
	}

	logFile := fmt.Sprintf("%s/app.log", logDir)
	rotator := &lumberjack.Logger{
		Filename:   logFile,
		MaxSize:    config.MaxSize,
		MaxBackups: config.MaxBackups,
		MaxAge:     config.MaxAge,
		Compress:   config.Compress,
	}

	lokiWriter := NewLokiWriter(
		lokiConfig.URL,
		lokiConfig.Labels,
		lokiConfig.FlushInterval,
	)

	multi := zerolog.MultiLevelWriter(
		zerolog.ConsoleWriter{Out: os.Stderr},
		rotator,
		lokiWriter,
	)

	// Initialize the singleton instance
	instance = &LoggerInstance{
		logger:     zerolog.New(multi).With().Timestamp().Logger(),
		lokiWriter: lokiWriter,
	}

	// Set log level based on debug configuration
	if debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		instance.logger.Debug().Msg("debug logging enabled")
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	return nil
}

// GetLogger returns the singleton logger instance
func GetLogger() *zerolog.Logger {
	once.Do(func() {
		if instance == nil {
			// Create a default logger if not initialized
			instance = &LoggerInstance{
				logger: zerolog.New(os.Stderr).With().Timestamp().Logger(),
			}
		}
	})
	return &instance.logger
}

func CloseLogger() {
	if instance != nil && instance.lokiWriter != nil {
		instance.lokiWriter.Close()
	}
}
