package temporal

import (
	"fmt"
	"path"
	"strings"

	"github.com/datazip-inc/olake-ui/server/internal/appconfig"
	"github.com/datazip-inc/olake-ui/server/internal/constants"
)

// workerConfigPath returns the config path passed to the worker command.
//
// NFS: /mnt/config/{file}.json — worker mounts the workflow subdirectory as a
// volume subPath at /mnt/config, so the hash is not in the CLI arg.
//
// S3: s3://{bucket}/{prefix}/{workflow-dir}/{file}.json
func workerConfigPath(cmd Command, workflowID, filename string) string {
	cfg := appconfig.Load()

	switch strings.ToLower(strings.TrimSpace(cfg.OlakeStorageMode)) {
	case constants.StorageModeNFS:
		return fmt.Sprintf("/mnt/config/%s", filename)
	case constants.StorageModeS3:
		bucket := strings.TrimSpace(cfg.OlakeS3Bucket)
		jobDir := GetWorkflowDirectory(cmd, workflowID)
		prefix := strings.Trim(strings.TrimSpace(cfg.OlakeS3Prefix), "/")
		key := path.Join(jobDir, filename)
		if prefix != "" {
			key = path.Join(prefix, key)
		}
		return fmt.Sprintf("s3://%s/%s", bucket, key)
	}

	return fmt.Sprintf("/mnt/config/%s", filename)
}
