package logger

import (
	"go.temporal.io/sdk/log"
)

// temporalLogger wraps our existing logger functions to implement Temporal's interface
type temporalLogger struct{}

func (l temporalLogger) Debug(msg string, keyvals ...interface{}) {
	if len(keyvals) > 0 {
		logger.Debug().Fields(keyvals).Msg(msg)
	} else {
		logger.Debug().Msg(msg)
	}
}

func (l temporalLogger) Info(msg string, keyvals ...interface{}) {
	if len(keyvals) > 0 {
		logger.Info().Fields(keyvals).Msg(msg)
	} else {
		logger.Info().Msg(msg)
	}
}

func (l temporalLogger) Warn(msg string, keyvals ...interface{}) {
	if len(keyvals) > 0 {
		logger.Warn().Fields(keyvals).Msg(msg)
	} else {
		logger.Warn().Msg(msg)
	}
}

func (l temporalLogger) Error(msg string, keyvals ...interface{}) {
	if len(keyvals) > 0 {
		logger.Error().Fields(keyvals).Msg(msg)
	} else {
		logger.Error().Msg(msg)
	}
}

// NewTemporalLogger creates a Temporal-compatible logger using our existing zerolog logger
func NewTemporalLogger() log.Logger {
	return temporalLogger{}
}
