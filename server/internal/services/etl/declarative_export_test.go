package services

import (
	"archive/zip"
	"bytes"
	"io"
	"testing"
)

func TestBuildZipBundleIncludesExpectedFiles(t *testing.T) {
	t.Parallel()

	data, err := buildZipBundle([]bundleFile{
		{name: "bundle/source.json", data: []byte(`{"database":"orders"}`)},
		{name: "bundle/olake-ui.json", data: []byte(`{"job_name":"orders"}`)},
	})
	if err != nil {
		t.Fatalf("buildZipBundle() failed: %v", err)
	}

	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatalf("zip.NewReader() failed: %v", err)
	}

	names := map[string]string{}
	for _, file := range reader.File {
		rc, err := file.Open()
		if err != nil {
			t.Fatalf("Open(%s) failed: %v", file.Name, err)
		}
		content, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			t.Fatalf("ReadAll(%s) failed: %v", file.Name, err)
		}
		names[file.Name] = string(content)
	}

	if names["bundle/source.json"] != `{"database":"orders"}` {
		t.Fatalf("unexpected source.json content: %q", names["bundle/source.json"])
	}
	if names["bundle/olake-ui.json"] != `{"job_name":"orders"}` {
		t.Fatalf("unexpected olake-ui.json content: %q", names["bundle/olake-ui.json"])
	}
}

func TestNormalizeExportFormat(t *testing.T) {
	t.Parallel()

	if normalizeExportFormat("tgz") != "tar.gz" {
		t.Fatalf("expected tgz to normalize to tar.gz")
	}
	if normalizeExportFormat("") != "zip" {
		t.Fatalf("expected empty format to default to zip")
	}
}
