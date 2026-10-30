package gitops

import (
	"bytes"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
)

// fileKeys are the manifest fields that are not part of the entity's config.
var fileKeys = []string{"kind", DataProjectID, DataUserID, DataJob}

// parseFileManifest parses one JSON manifest (a source, destination, job or streams) into a ResourceData.
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
		delete(data, DataUserID) // streams have no userId
	} else {
		if name, err = stringField(doc, "name"); err != nil {
			return nil, err
		}
		if name == "" {
			return nil, fmt.Errorf("%s needs name", kind)
		}
		delete(data, DataJob)
	}

	if data[DataConfig], err = resolveConfig(kind, doc); err != nil {
		return nil, err
	}
	return &ResourceData{Kind: kind, Name: name, Data: data}, nil
}

// resolveConfig returns the entity's config as compact JSON with sorted keys, so equal content
// gives equal bytes. For streams it is the "streams" object: the CLI's selected_streams.json, pasted
// as is. For the other kinds it is every field besides kind, projectId, userId and job.
func resolveConfig(kind string, doc map[string]any) (string, error) {
	for _, key := range fileKeys {
		delete(doc, key)
	}
	var config any = doc
	if kind == KindStreams {
		streams, ok := doc["streams"].(map[string]any)
		if !ok {
			return "", fmt.Errorf(`streams needs "streams": the content of the selected_streams.json written by discover`)
		}
		config = streams
	}
	b, err := json.Marshal(config)
	if err != nil {
		return "", fmt.Errorf("encode config: %w", err)
	}
	return string(b), nil
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
