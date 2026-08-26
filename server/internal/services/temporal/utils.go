package temporal

import (
	"context"
	"encoding/json"
	"fmt"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/datazip-inc/olake-ui/server/internal/appconfig"
	"github.com/datazip-inc/olake-ui/server/internal/constants"
	"github.com/datazip-inc/olake-ui/server/internal/models"
	"github.com/datazip-inc/olake-ui/server/internal/storage"
	"github.com/datazip-inc/olake-ui/server/internal/storagemode"
	"github.com/datazip-inc/olake-ui/server/internal/utils"
	"go.temporal.io/sdk/client"
)

// buildExecutionReqForSync builds the ExecutionRequest for a sync job
func buildExecutionReqForSync(job *models.Job, workflowID string) *ExecutionRequest {
	args := []string{
		"sync",
		"--config", workerConfigPath(Sync, workflowID, "source.json"),
		"--destination", workerConfigPath(Sync, workflowID, "destination.json"),
		"--catalog", workerConfigPath(Sync, workflowID, "streams.json"),
		"--state", workerConfigPath(Sync, workflowID, "state.json"),
	}

	return &ExecutionRequest{
		Command:       Sync,
		ConnectorType: job.Source.Type,
		Version:       job.Source.Version,
		Args:          args,
		Configs:       nil,
		WorkflowID:    workflowID,
		JobID:         job.ID,
		ProjectID:     job.ProjectID,
		Timeout:       GetWorkflowTimeout(Sync),
		OutputFile:    "state.json",
	}
}

// buildExecutionReqForClearDestination builds the ExecutionRequest for a clear-destination job
func buildExecutionReqForClearDestination(ctx context.Context, job *models.Job, workflowID, streamsConfig string) (*ExecutionRequest, error) {
	catalog := streamsConfig
	if catalog == "" {
		catalog = job.StreamsConfig
	}

	streamsDir := fmt.Sprintf("%s-%d", workflowID, time.Now().Unix())
	relativePath := filepath.Join(streamsDir, "streams.json")

	switch storagemode.Get() {
	case constants.StorageModeS3:
		if err := storage.WriteFilesToS3(ctx, constants.DefaultConfigDir, []storage.JobConfig{{RelativePath: relativePath, Data: catalog}}); err != nil {
			return nil, fmt.Errorf("failed to write streams config to s3: %v", err)
		}
	default:
		streamsPath := filepath.Join(constants.DefaultConfigDir, relativePath)

		if err := utils.WriteFile(streamsPath, []byte(catalog), 0644); err != nil {
			return nil, fmt.Errorf("failed to write streams config to file: %v", err)
		}
	}

	args := []string{
		"clear-destination",
		"--streams", workerConfigPath(ClearDestination, workflowID, "streams.json"),
		"--state", workerConfigPath(ClearDestination, workflowID, "state.json"),
		"--destination", workerConfigPath(ClearDestination, workflowID, "destination.json"),
	}

	return &ExecutionRequest{
		Command:       ClearDestination,
		ConnectorType: job.Source.Type,
		Version:       job.Source.Version,
		Args:          args,
		Configs:       nil,
		WorkflowID:    workflowID,
		ProjectID:     job.ProjectID,
		JobID:         job.ID,
		Timeout:       GetWorkflowTimeout(ClearDestination),
		OutputFile:    "state.json",
		TempPath:      relativePath,
	}, nil
}

// extractWorkflowResponse extracts and parses the JSON response from a workflow execution result
func ExtractWorkflowResponse(ctx context.Context, run client.WorkflowRun) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	if err := run.Get(ctx, &result); err != nil {
		return nil, fmt.Errorf("workflow execution failed: %v", err)
	}

	// response is the relative path to the file that contains the workflow response
	response, ok := result["response"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid response format from worker")
	}

	responsePath := filepath.Join(constants.DefaultConfigDir, response)
	switch storagemode.Get() {
	case constants.StorageModeS3:
		body, _, err := storage.ReadFileFromS3(ctx, "", response, true)
		if err != nil {
			return nil, err
		}

		var workflowResponse map[string]interface{}
		if err := json.Unmarshal([]byte(body), &workflowResponse); err != nil {
			return nil, fmt.Errorf("failed to parse JSON from %s: %s", response, err)
		}

		return workflowResponse, nil
	default:
		workflowResponse, err := utils.ReadJSONFile(responsePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read workflow response: %v", err)
		}

		return workflowResponse, nil
	}
}

func GetWorkflowTimeout(op Command) time.Duration {
	switch op {
	case Discover:
		return time.Minute * 10
	case Check:
		return time.Minute * 10
	case Spec:
		return time.Minute * 5
	case Sync:
		return time.Hour * 24 * 30
	case ClearDestination:
		return time.Hour * 24 * 30
	// check what can the fallback time be
	default:
		return time.Minute * 5
	}
}

// workerConfigPath returns the config path passed to the worker command.
//
// NFS: /mnt/config/{file}.json — worker mounts the workflow subdirectory as a
// volume subPath at /mnt/config, so the hash is not in the CLI arg.
//
// S3: s3://{bucket}/{prefix}/{workflow-dir}/{file}.json
func workerConfigPath(cmd Command, workflowID, filename string) string {
	switch storagemode.Get() {
	case constants.StorageModeS3:
		cfg := appconfig.Load()
		bucket := strings.TrimSpace(cfg.OlakeS3Bucket)
		jobDir := GetWorkflowDirectory(cmd, workflowID)
		key := path.Join(jobDir, filename)
		if prefix := strings.Trim(cfg.OlakeS3Prefix, "/"); prefix != "" {
			key = path.Join(prefix, key)
		}
		return fmt.Sprintf("s3://%s/%s", bucket, key)
	default:
		return fmt.Sprintf("/mnt/config/%s", filename)
	}
}
