package services

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"testing"
)

func TestCanonicalizeJSONSortsObjectKeys(t *testing.T) {
	t.Parallel()

	left, err := canonicalizeJSON([]byte(`{"b":2,"a":1,"nested":{"d":4,"c":3}}`))
	if err != nil {
		t.Fatalf("canonicalizeJSON(left) failed: %v", err)
	}

	right, err := canonicalizeJSON([]byte(`{"nested":{"c":3,"d":4},"a":1,"b":2}`))
	if err != nil {
		t.Fatalf("canonicalizeJSON(right) failed: %v", err)
	}

	if left != right {
		t.Fatalf("expected canonical JSON to match, left=%s right=%s", left, right)
	}
}

func TestExtractCLIBundleFilesFromZip(t *testing.T) {
	t.Parallel()

	var buffer bytes.Buffer
	zipWriter := zip.NewWriter(&buffer)

	writeZipFile := func(name, content string) {
		t.Helper()
		writer, err := zipWriter.Create(name)
		if err != nil {
			t.Fatalf("Create(%s) failed: %v", name, err)
		}
		if _, err := writer.Write([]byte(content)); err != nil {
			t.Fatalf("Write(%s) failed: %v", name, err)
		}
	}

	writeZipFile("mongo-job/source.json", `{"database":"orders"}`)
	writeZipFile("mongo-job/destination.json", `{"type":"PARQUET"}`)
	writeZipFile("mongo-job/streams.json", `{"streams":[]}`)
	writeZipFile("mongo-job/olake-ui.json", `{"source_type":"mongodb","source_version":"v0.3.0","destination_version":"v0.3.0"}`)

	if err := zipWriter.Close(); err != nil {
		t.Fatalf("Close() failed: %v", err)
	}

	files, err := extractCLIBundleFiles("mongo-job.zip", buffer.Bytes())
	if err != nil {
		t.Fatalf("extractCLIBundleFiles() failed: %v", err)
	}

	if files.BundleName != "mongo-job" {
		t.Fatalf("expected bundle name mongo-job, got %q", files.BundleName)
	}
	if len(files.Source) == 0 || len(files.Dest) == 0 || len(files.Streams) == 0 {
		t.Fatalf("expected required files to be loaded")
	}
	if files.Overlay.SourceType != "mongodb" {
		t.Fatalf("expected overlay source_type mongodb, got %q", files.Overlay.SourceType)
	}
}

func TestExtractCLIBundleFilesFromTarGz(t *testing.T) {
	t.Parallel()

	var buffer bytes.Buffer
	gzipWriter := gzip.NewWriter(&buffer)
	tarWriter := tar.NewWriter(gzipWriter)

	writeTarFile := func(name, content string) {
		t.Helper()
		header := &tar.Header{
			Name: name,
			Mode: 0600,
			Size: int64(len(content)),
		}
		if err := tarWriter.WriteHeader(header); err != nil {
			t.Fatalf("WriteHeader(%s) failed: %v", name, err)
		}
		if _, err := tarWriter.Write([]byte(content)); err != nil {
			t.Fatalf("Write(%s) failed: %v", name, err)
		}
	}

	writeTarFile("bundle/source.json", `{"database":"orders"}`)
	writeTarFile("bundle/destination.json", `{"type":"PARQUET"}`)
	writeTarFile("bundle/streams.json", `{"streams":[]}`)

	if err := tarWriter.Close(); err != nil {
		t.Fatalf("tarWriter.Close() failed: %v", err)
	}
	if err := gzipWriter.Close(); err != nil {
		t.Fatalf("gzipWriter.Close() failed: %v", err)
	}

	files, err := extractCLIBundleFiles("bundle.tar.gz", buffer.Bytes())
	if err != nil {
		t.Fatalf("extractCLIBundleFiles() failed: %v", err)
	}

	if files.BundleName != "bundle" {
		t.Fatalf("expected bundle name bundle, got %q", files.BundleName)
	}
}
