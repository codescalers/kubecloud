package logger

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/natefinch/lumberjack"
	"github.com/rs/zerolog"
)

// LoggerInstance holds the singleton logger instance
type LoggerInstance struct {
	logger      zerolog.Logger
	auditLogger zerolog.Logger
	lokiWriter  *LokiWriter
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

type AuditLogConfig struct {
	Enabled bool      `json:"enabled"`
	Sink    io.Writer `json:"sink"`
}

// InitLogger initializes the singleton logger with file and console output
func InitLogger(config LoggerConfig, lokiConfig *LokiConfig, auditLogConfig *AuditLogConfig, debug bool) error {
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

	// Ensure the file is writable by opening it first
	if _, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644); err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	writers := []io.Writer{
		zerolog.ConsoleWriter{Out: os.Stderr},
		rotator,
	}

	var lokiWriter *LokiWriter
	if lokiConfig != nil && lokiConfig.URL != "" {
		lokiWriter = NewLokiWriter(lokiConfig.URL, lokiConfig.Labels, lokiConfig.FlushInterval)

		if err := pingLoki(lokiWriter); err != nil {
			lokiWriter = nil
		} else {
			writers = append(writers, lokiWriter)
		}
	}

	multi := zerolog.MultiLevelWriter(writers...)

	// Initialize the singleton instance
	instance = &LoggerInstance{
		logger:     zerolog.New(multi).With().Timestamp().Logger(),
		lokiWriter: lokiWriter,
	}

	// Initialize dedicated audit logger (discard output when disabled/missing)
	instance.auditLogger = setupAuditLogger(auditLogConfig)

	// Log to stderr that we've set up file logging
	fmt.Fprintf(os.Stderr, "Logger initialized with file output to: %s\n", logFile)

	// Set log level based on debug configuration
	if debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	outputTypes := []string{"console", "file"}
	if lokiWriter != nil {
		outputTypes = append(outputTypes, "Loki")
	}

	instance.logger.Info().
		Bool("level-debug", debug).
		Strs("writers", outputTypes).
		Str("log_file", logFile).
		Msgf("Logger initialized")

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

func pingLoki(lw *LokiWriter) error {
	if lw == nil {
		return fmt.Errorf("loki writer is nil")
	}

	testPayload := `{"streams":[]}`
	resp, err := lw.client.Post(lw.url, "application/json", bytes.NewBufferString(testPayload))
	if err != nil {
		return fmt.Errorf("loki post test failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("loki returned bad status: %s", resp.Status)
	}

	return nil
}

func ForRequest(userID int, requestID, handler string) *zerolog.Logger {
	log := GetLogger().With().
		Str("request_id", requestID).
		Str("handler", handler)
	if userID != 0 {
		log = log.Int("user_id", userID)
	}
	logger := log.Logger()
	return &logger
}

func ForOperation(component, operation string) *zerolog.Logger {
	log := GetLogger().With().
		Str("component", component).
		Str("operation", operation)
	logger := log.Logger()
	return &logger
}

// GetAuditLogger returns the dedicated audit logger instance.
func GetAuditLogger() *zerolog.Logger {
	once.Do(func() {
		if instance == nil {
			// Create a default logger if not initialized
			instance = &LoggerInstance{
				auditLogger: zerolog.New(io.Discard).With().Timestamp().Logger(),
			}
		}
	})
	return &instance.auditLogger
}
