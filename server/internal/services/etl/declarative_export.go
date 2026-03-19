package services

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/datazip-inc/olake-ui/server/internal/models/dto"
)

func (s *ETLService) ExportCLIBundle(projectID string, jobID int, includeState bool, format string) ([]byte, string, string, error) {
	job, err := s.db.GetJobByID(jobID, true)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to get job: %s", err)
	}
	if job.ProjectID != projectID {
		return nil, "", "", fmt.Errorf("job %d does not belong to project %s", jobID, projectID)
	}
	if job.SourceID == nil || job.DestID == nil {
		return nil, "", "", fmt.Errorf("job %d is missing source or destination", jobID)
	}

	bundleName := job.ApplyID
	if strings.TrimSpace(bundleName) == "" {
		bundleName = sanitizeApplyIdentity(job.Name)
	}

	activate := job.Active
	overlay := dto.ApplyCLIBundleOverlay{
		ApplyIdentity:      bundleName,
		JobName:            job.Name,
		SourceName:         job.SourceID.Name,
		SourceType:         job.SourceID.Type,
		SourceVersion:      job.SourceID.Version,
		DestinationName:    job.DestID.Name,
		DestinationType:    job.DestID.DestType,
		DestinationVersion: job.DestID.Version,
		Frequency:          job.Frequency,
		Activate:           &activate,
	}

	overlayJSON, err := canonicalizeJSONMustMarshal(overlay)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to marshal olake-ui.json: %s", err)
	}

	sourceJSON, err := canonicalizeJSON([]byte(job.SourceID.Config))
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to canonicalize source config: %s", err)
	}

	destJSON, err := canonicalizeJSON([]byte(job.DestID.Config))
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to canonicalize destination config: %s", err)
	}

	streamsJSON, err := canonicalizeJSON([]byte(job.StreamsConfig))
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to canonicalize streams config: %s", err)
	}

	files := []bundleFile{
		{name: filepath.ToSlash(filepath.Join(bundleName, "source.json")), data: []byte(sourceJSON)},
		{name: filepath.ToSlash(filepath.Join(bundleName, "destination.json")), data: []byte(destJSON)},
		{name: filepath.ToSlash(filepath.Join(bundleName, "streams.json")), data: []byte(streamsJSON)},
		{name: filepath.ToSlash(filepath.Join(bundleName, "olake-ui.json")), data: []byte(overlayJSON)},
	}

	if includeState {
		stateJSON, err := canonicalizeJSON([]byte(job.State))
		if err != nil {
			return nil, "", "", fmt.Errorf("failed to canonicalize state file: %s", err)
		}
		files = append(files, bundleFile{name: filepath.ToSlash(filepath.Join(bundleName, "state.json")), data: []byte(stateJSON)})
	}

	switch normalizeExportFormat(format) {
	case "tar.gz":
		data, err := buildTarGzBundle(files)
		return data, fmt.Sprintf("%s.tar.gz", bundleName), "application/gzip", err
	default:
		data, err := buildZipBundle(files)
		return data, fmt.Sprintf("%s.zip", bundleName), "application/zip", err
	}
}

type bundleFile struct {
	name string
	data []byte
}

func buildZipBundle(files []bundleFile) ([]byte, error) {
	buffer := &bytes.Buffer{}
	writer := zip.NewWriter(buffer)

	for _, file := range files {
		entry, err := writer.Create(file.name)
		if err != nil {
			return nil, fmt.Errorf("failed to create zip entry %s: %s", file.name, err)
		}
		if _, err := entry.Write(file.data); err != nil {
			return nil, fmt.Errorf("failed to write zip entry %s: %s", file.name, err)
		}
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to finalize zip bundle: %s", err)
	}
	return buffer.Bytes(), nil
}

func buildTarGzBundle(files []bundleFile) ([]byte, error) {
	buffer := &bytes.Buffer{}
	gzipWriter := gzip.NewWriter(buffer)
	tarWriter := tar.NewWriter(gzipWriter)

	for _, file := range files {
		header := &tar.Header{
			Name: file.name,
			Mode: 0600,
			Size: int64(len(file.data)),
		}
		if err := tarWriter.WriteHeader(header); err != nil {
			return nil, fmt.Errorf("failed to create tar entry %s: %s", file.name, err)
		}
		if _, err := tarWriter.Write(file.data); err != nil {
			return nil, fmt.Errorf("failed to write tar entry %s: %s", file.name, err)
		}
	}

	if err := tarWriter.Close(); err != nil {
		return nil, fmt.Errorf("failed to finalize tar archive: %s", err)
	}
	if err := gzipWriter.Close(); err != nil {
		return nil, fmt.Errorf("failed to finalize gzip stream: %s", err)
	}
	return buffer.Bytes(), nil
}

func canonicalizeJSONMustMarshal(value interface{}) (string, error) {
	raw, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	return canonicalizeJSON(raw)
}

func normalizeExportFormat(format string) string {
	switch strings.ToLower(strings.TrimSpace(format)) {
	case "tar.gz", "tgz":
		return "tar.gz"
	default:
		return "zip"
	}
}
