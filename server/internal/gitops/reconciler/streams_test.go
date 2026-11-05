package reconciler

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseStreamsCM(t *testing.T) {
	tests := []struct {
		name    string
		config  string
		wantErr string
	}{
		{name: "selection only", config: userSelectedStreams},
		{name: "empty streams array is not legacy", config: `{"streams":[],"selected_streams":{"public":[]}}`},
		{name: "null streams is not legacy", config: `{"streams":null,"selected_streams":{"public":[]}}`},
		{name: "invalid JSON", config: `{"selected_streams":`, wantErr: "invalid streams config"},
		{name: "legacy streams.json", config: `{"streams":[{"stream":{"name":"users"}}],"selected_streams":{"public":[]}}`, wantErr: "legacy streams config (streams[]) is not supported"},
		{name: "legacy streams without selection", config: `{"streams":[{"stream":{"name":"users"}}]}`, wantErr: "legacy streams config"},
		{name: "no selected_streams", config: `{}`, wantErr: "streams config has no selected_streams"},
		{name: "null selected_streams", config: `{"selected_streams":null}`, wantErr: "streams config has no selected_streams"},
		{name: "empty selected_streams", config: `{"selected_streams":{}}`, wantErr: "streams config has no selected_streams"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			catalog, err := parseStreamsCM(tt.config)
			if tt.wantErr != "" {
				assert.ErrorIs(t, err, ErrNonRetryable)
				assert.ErrorContains(t, err, tt.wantErr)
				return
			}
			assert.NoError(t, err)
			// the ConfigMap is stored as written; discover fills the available streams later
			assert.Equal(t, tt.config, catalog.Selected)
			assert.Empty(t, catalog.Available)
			assert.Empty(t, catalog.Streams)
		})
	}
}
