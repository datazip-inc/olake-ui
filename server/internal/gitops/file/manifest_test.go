package file

import (
	"testing"

	"github.com/datazip-inc/olake-ui/server/internal/gitops/reconciler"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseFileManifest(t *testing.T) {
	t.Run("source", func(t *testing.T) {
		res, err := parseFileManifest([]byte(`{"kind":"source","projectId":"123","userId":1,"name":"pg","type":"postgres","version":"v1.0.0","config":{"port":12345678901234567890}}`))
		require.NoError(t, err)
		assert.Equal(t, reconciler.KindSource, res.Kind)
		assert.Equal(t, "pg", res.Name)
		// a numeric userId reads as the string a ConfigMap holds; numbers in the config are kept as written
		assert.Equal(t, map[string]string{
			reconciler.DataProjectID: "123",
			reconciler.DataUserID:    "1",
			reconciler.DataConfig:    `{"config":{"port":12345678901234567890},"name":"pg","type":"postgres","version":"v1.0.0"}`,
		}, res.Data)
	})

	t.Run("streams", func(t *testing.T) {
		res, err := parseFileManifest([]byte(`{"kind":"streams","projectId":"123","userId":1,"job":"orders","streams":` + testSelection + `}`))
		require.NoError(t, err)
		assert.Equal(t, "orders", res.Name, "streams are identified by their job")
		// the config is the pasted selected_streams.json, as a Streams ConfigMap holds it; userId is dropped
		assert.JSONEq(t, testSelection, res.Config())
		delete(res.Data, reconciler.DataConfig)
		assert.Equal(t, map[string]string{reconciler.DataProjectID: "123", reconciler.DataJob: "orders"}, res.Data)
	})

	t.Run("key order and whitespace do not change the config", func(t *testing.T) {
		a, err := parseFileManifest([]byte(`{"kind":"source","name":"pg","projectId":"123","config":{"host":"pg","port":5432}}`))
		require.NoError(t, err)
		b, err := parseFileManifest([]byte("{\n  \"config\": { \"port\": 5432, \"host\": \"pg\" },\n  \"projectId\": \"123\", \"name\": \"pg\", \"kind\": \"source\"\n}"))
		require.NoError(t, err)
		assert.Equal(t, a.Data, b.Data)
	})

	for _, tt := range []struct {
		name, raw, wantErr string
	}{
		{name: "invalid JSON", raw: `{"kind":`, wantErr: "invalid JSON"},
		{name: "not an object", raw: `null`, wantErr: "expected a JSON object"},
		{name: "missing kind", raw: `{"name":"pg"}`, wantErr: "missing kind"},
		{name: "unknown kind", raw: `{"kind":"pipeline","name":"pg"}`, wantErr: `invalid kind "pipeline"`},
		{name: "source without name", raw: `{"kind":"source","projectId":"123"}`, wantErr: "source needs name"},
		{name: "streams without job", raw: `{"kind":"streams","streams":{}}`, wantErr: "streams needs job"},
		{name: "streams selection at the top level", raw: `{"kind":"streams","job":"orders","selected_streams":{}}`, wantErr: `streams needs "streams"`},
		{name: "projectId of the wrong type", raw: `{"kind":"source","name":"pg","projectId":true}`, wantErr: "projectId must be a string"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseFileManifest([]byte(tt.raw))
			assert.ErrorContains(t, err, tt.wantErr)
		})
	}
}

func TestIsManifest(t *testing.T) {
	for _, name := range []string{"pg.json", "jobs/orders/streams.JSON"} {
		assert.True(t, isManifest(name), name)
	}
	for _, name := range []string{stateFileName, "jobs/.draft.json", "pg.yaml", "pg.json.tmp"} {
		assert.False(t, isManifest(name), name)
	}
}
