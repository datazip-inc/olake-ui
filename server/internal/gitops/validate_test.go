package gitops

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseAndValidateJobConfig(t *testing.T) {
	for _, field := range []string{"name", "source", "destination", "frequency"} {
		t.Run("missing "+field, func(t *testing.T) {
			_, err := ParseAndValidateJobConfig(jobConfigJSON(map[string]interface{}{field: nil}))
			assert.ErrorIs(t, err, ErrNonRetryable)
			assert.ErrorContains(t, err, "validation failed")
		})
	}

	t.Run("invalid JSON", func(t *testing.T) {
		_, err := ParseAndValidateJobConfig(`{"name":`)
		assert.ErrorIs(t, err, ErrNonRetryable)
		assert.ErrorContains(t, err, "invalid spec.config JSON")
	})

	t.Run("wrong field type", func(t *testing.T) {
		_, err := ParseAndValidateJobConfig(jobConfigJSON(map[string]interface{}{"activate": "yes"}))
		assert.ErrorIs(t, err, ErrNonRetryable)
	})
}

func TestParseAndValidateSource(t *testing.T) {
	tests := []struct {
		name       string
		config     string
		wantConfig string
		wantErr    string
	}{
		{
			name:       "config as an object is compacted",
			config:     `{"name":"pg","type":"postgres","version":"v1.0.0","config":{ "host": "pg" }}`,
			wantConfig: `{"host":"pg"}`,
		},
		{
			name:       "config as a JSON string",
			config:     `{"name":"pg","type":"postgres","version":"v1.0.0","config":"{\"host\":\"pg\"}"}`,
			wantConfig: `{"host":"pg"}`,
		},
		{name: "unknown type", config: `{"name":"pg","type":"mssql2","version":"v1.0.0","config":{}}`, wantErr: "invalid source type 'mssql2'"},
		{name: "destination type", config: `{"name":"pg","type":"iceberg","version":"v1.0.0","config":{}}`, wantErr: "invalid source type"},
		{name: "missing version", config: `{"name":"pg","type":"postgres","config":{}}`, wantErr: "validation failed"},
		{name: "missing config", config: `{"name":"pg","type":"postgres","version":"v1.0.0"}`, wantErr: "validation failed"},
		{name: "config is an array", config: `{"name":"pg","type":"postgres","version":"v1.0.0","config":[]}`, wantErr: "invalid spec.config JSON"},
		{name: "invalid JSON", config: `not json`, wantErr: "invalid spec.config JSON"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := ParseAndValidateSource(tt.config)
			if tt.wantErr != "" {
				assert.ErrorIs(t, err, ErrNonRetryable)
				assert.ErrorContains(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantConfig, req.Config.String())
		})
	}
}

func TestParseAndValidateDestination(t *testing.T) {
	tests := []struct {
		name    string
		config  string
		wantErr string
	}{
		{name: "source type", config: `{"name":"lake","type":"postgres","version":"v1.0.0","config":{}}`, wantErr: "invalid destination type 'postgres'"},
		{name: "missing name", config: `{"type":"iceberg","version":"v1.0.0","config":{}}`, wantErr: "validation failed"},
		{name: "invalid JSON", config: `{"name":`, wantErr: "invalid spec.config JSON"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseAndValidateDestination(tt.config)
			assert.ErrorIs(t, err, ErrNonRetryable)
			assert.ErrorContains(t, err, tt.wantErr)
		})
	}
}
