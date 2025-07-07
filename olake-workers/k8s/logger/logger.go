package logger

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

var logger zerolog.Logger

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

func Init() {
	zerolog.TimestampFunc = func() time.Time { return time.Now().UTC() }

	// Log level colors
	logColors := map[string]string{
		"debug": "\033[36m", // Cyan
		"info":  "\033[32m", // Green
		"warn":  "\033[33m", // Yellow
		"error": "\033[31m", // Red
		"fatal": "\033[31m", // Red
	}

	consoleWriter := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: "2006-01-02 15:04:05",
		FormatLevel: func(i interface{}) string {
			level := strings.ToLower(fmt.Sprintf("%s", i))
			if color, exists := logColors[level]; exists {
				return fmt.Sprintf("%s%s\033[0m", color, strings.ToUpper(level))
			}
			return strings.ToUpper(level)
		},
		FormatMessage: func(i interface{}) string {
			switch v := i.(type) {
			case string:
				return v
			default:
				jsonMsg, _ := json.Marshal(v)
				return string(jsonMsg)
			}
		},
		FormatTimestamp: func(i interface{}) string {
			return fmt.Sprintf("\033[90m%s\033[0m", i)
		},
	}

	logger = zerolog.New(consoleWriter).With().Timestamp().Logger()
}
