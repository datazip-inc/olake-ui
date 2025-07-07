package logger

import (
	"io"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"

	"olake-ui/olake-workers/k8s/config/types"
)

var logger zerolog.Logger

// Init initializes the global logger based on the provided configuration
func Init(config types.LoggingConfig) {
	zerolog.TimestampFunc = func() time.Time { return time.Now().UTC() }

	var writer io.Writer
	switch strings.ToLower(config.Format) {
	case "console":
		// Use ConsoleWriter with built-in colors and formatting
		writer = zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		}
	case "json":
		writer = os.Stdout
	default:
		// Default to JSON for production safety
		writer = os.Stdout
	}

	logger = zerolog.New(writer).With().Timestamp().Logger()
	zerolog.SetGlobalLevel(parseLogLevel(config.Level))
}

// InitDefault initializes the logger with default settings for backward compatibility
func InitDefault() {
	Init(types.LoggingConfig{
		Level:  "info",
		Format: "console",
	})
}

// parseLogLevel converts a string level to a zerolog.Level
func parseLogLevel(levelStr string) zerolog.Level {
	switch strings.ToLower(levelStr) {
	case "debug":
		return zerolog.DebugLevel
	case "info":
		return zerolog.InfoLevel
	case "warn":
		return zerolog.WarnLevel
	case "error":
		return zerolog.ErrorLevel
	case "fatal":
		return zerolog.FatalLevel
	default:
		return zerolog.InfoLevel // Default to info level
	}
}

// Info writes record with log level INFO
func Info(v ...interface{}) {
	if len(v) == 1 {
		logger.Info().Interface("message", v[0]).Send()
	} else {
		logger.Info().Msgf("%s", v...)
	}
}

func Infof(format string, v ...interface{}) {
	logger.Info().Msgf(format, v...)
}

func Debug(v ...interface{}) {
	logger.Debug().Msgf("%s", v...)
}

func Debugf(format string, v ...interface{}) {
	logger.Debug().Msgf(format, v...)
}

func Error(v ...interface{}) {
	logger.Error().Msgf("%s", v...)
}

func Errorf(format string, v ...interface{}) {
	logger.Error().Msgf(format, v...)
}

func Warn(v ...interface{}) {
	logger.Warn().Msgf("%s", v...)
}

func Warnf(format string, v ...interface{}) {
	logger.Warn().Msgf(format, v...)
}

func Fatal(v ...interface{}) {
	logger.Fatal().Msgf("%s", v...)
	os.Exit(1)
}

func Fatalf(format string, v ...interface{}) {
	logger.Fatal().Msgf(format, v...)
	os.Exit(1)
}