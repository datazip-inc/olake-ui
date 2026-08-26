package temporal

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/datazip-inc/olake-ui/server/internal/constants"
	"github.com/datazip-inc/olake-ui/server/internal/storage"
)

func writeConfigFilesToS3(ctx context.Context, cmd Command, workflowID string, configs []JobConfig) error {
	workDir := filepath.Join(constants.DefaultConfigDir, GetWorkflowDirectory(cmd, workflowID))
	files := make([]storage.JobConfig, 0, len(configs))
	for _, jobConfig := range configs {
		files = append(files, storage.JobConfig{Name: jobConfig.Name, Data: jobConfig.Data})
	}
	return storage.WriteFilesToS3(ctx, workDir, files)
}

func readJSONFileFromS3(ctx context.Context, relPath string) (map[string]interface{}, error) {
	body, err := storage.ReadFileFromS3(ctx, "", relPath, true)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(body), &result); err != nil {
		return nil, fmt.Errorf("failed to parse JSON from %s: %s", relPath, err)
	}

	return result, nil
}
