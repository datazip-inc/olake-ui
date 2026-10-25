package gitops

import (
	"bytes"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
)

// File mode manifests are plain JSON, one entity per file, with no Kubernetes fields:
//
//	{"kind": "source", "projectId": "123", "userId": 1, "name": "pg", "type": "postgres", "version": "v1.0.0", "config": {...}}
//	{"kind": "destination", "projectId": "123", "userId": 1, "name": "lake", "type": "iceberg", "version": "v1.0.0", "config": {...}}
//	{"kind": "job", "projectId": "123", "userId": 1, "name": "orders", "source": "pg", "destination": "lake", "frequency": "0 * * * *"}
//	{"kind": "streams", "projectId": "123", "job": "orders", "selected_streams": {...}}
//
// kind, projectId, userId and job map to the ConfigMap's label and data keys; the remaining fields
// are the entity itself, in the shape the API takes, and become data.config. A source, destination
// or job is identified by its OLake name, a streams file by the job it belongs to.

// fileKeys are the manifest fields that are not part of the entity's config.
var fileKeys = []string{"kind", DataProjectID, DataUserID, DataJob}

// parseFileManifest reads one file mode manifest into the ResourceData the reconcilers take.
func parseFileManifest(raw []byte) (*ResourceData, error) {
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber() // keep numbers as written, so the config and its hash do not change on re-read
	var doc map[string]any
	if err := dec.Decode(&doc); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}
	if doc == nil {
		return nil, fmt.Errorf("expected a JSON object")
	}

	kind, err := stringField(doc, "kind")
	if err != nil {
		return nil, err
	}
	if kind == "" {
		return nil, fmt.Errorf("missing kind (source, destination, job, or streams)")
	}
	if !isValidManagedKind(kind) {
		return nil, fmt.Errorf("invalid kind %q (want source, destination, job, or streams)", kind)
	}
	data := map[string]string{}
	for _, key := range []string{DataProjectID, DataUserID, DataJob} {
		if v, err := stringField(doc, key); err != nil {
			return nil, err
		} else if v != "" {
			data[key] = v
		}
	}

	var name string
	if kind == KindStreams {
		if name = data[DataJob]; name == "" {
			return nil, fmt.Errorf("streams needs job (the name or id of the job it belongs to)")
		}
		delete(data, DataUserID) // a Streams ConfigMap has no userId either
	} else {
		if name, err = stringField(doc, "name"); err != nil {
			return nil, err
		}
		if name == "" {
			return nil, fmt.Errorf("%s needs name", kind)
		}
		delete(data, DataJob)
	}

	for _, key := range fileKeys {
		delete(doc, key)
	}
	config, err := json.Marshal(doc) // map keys are sorted, so equal content gives equal bytes
	if err != nil {
		return nil, fmt.Errorf("encode config: %w", err)
	}
	data[DataConfig] = string(config)
	return &ResourceData{Kind: kind, Name: name, Data: data}, nil
}

// stringField returns doc[key] as a string. Numbers are accepted, so "userId": 1 and "userId": "1"
// are the same. A missing key is "".
func stringField(doc map[string]any, key string) (string, error) {
	switch v := doc[key].(type) {
	case nil:
		return "", nil
	case string:
		return strings.TrimSpace(v), nil
	case json.Number:
		return v.String(), nil
	default:
		return "", fmt.Errorf("%s must be a string", key)
	}
}

func isValidManagedKind(kind string) bool {
	switch kind {
	case KindSource, KindDestination, KindJob, KindStreams:
		return true
	default:
		return false
	}
}

// isManifest reports whether a file is read as a manifest: *.json except dotfiles, which leaves out
// the state file (.olake-gitops-state.json).
func isManifest(name string) bool {
	base := filepath.Base(name)
	return !strings.HasPrefix(base, ".") && strings.EqualFold(filepath.Ext(base), ".json")
}
