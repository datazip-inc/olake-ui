package temporal

import (
	"context"
	"encoding/json"
	"fmt"
	"path"

	"github.com/datazip-inc/olake-ui/server/internal/storage"
)

func writeConfigFilesToS3(ctx context.Context, cmd Command, workflowID string, configs []JobConfig) error {
	jobDir := GetWorkflowDirectory(cmd, workflowID)
	for _, jobConfig := range configs {
		relPath := path.Join(jobDir, jobConfig.Name)
		if err := storage.WriteFileToS3AtRelPath(ctx, relPath, []byte(jobConfig.Data)); err != nil {
			return fmt.Errorf("failed to upload %s: %s", jobConfig.Name, err)
		}
	}
	return nil
}

func readJSONFileFromS3(ctx context.Context, relPath string) (map[string]interface{}, error) {
	body, err := storage.ReadFileFromS3AtRelPath(ctx, relPath)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse JSON from %s: %s", relPath, err)
	}

	return result, nil
}
