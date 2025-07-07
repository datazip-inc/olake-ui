package types

// LoggingConfig contains logging settings
type LoggingConfig struct {
	Level  string `json:"level"`
	Format string `json:"format"`
}
