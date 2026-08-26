package temporal

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/datazip-inc/olake-ui/server/internal/appconfig"
	"github.com/datazip-inc/olake-ui/server/internal/constants"
	"github.com/datazip-inc/olake-ui/server/internal/models"
	"github.com/datazip-inc/olake-ui/server/internal/storage"
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

	switch strings.ToLower(strings.TrimSpace(appconfig.Load().OlakeStorageMode)) {
	case constants.StorageModeS3:
		if err := storage.WriteFilesToS3(ctx, constants.DefaultConfigDir, []storage.JobConfig{{Name: relativePath, Data: catalog}}); err != nil {
			return nil, fmt.Errorf("failed to write streams config to s3: %v", err)
		}
	case constants.StorageModeNFS:
		streamsPath := filepath.Join(constants.DefaultConfigDir, relativePath)

		if err := utils.WriteFile(streamsPath, []byte(catalog), 0644); err != nil {
			return nil, fmt.Errorf("failed to write streams config to file: %v", err)
		}
	default:
		return nil, fmt.Errorf("unsupported OLAKE_STORAGE_MODE: %q", appconfig.Load().OlakeStorageMode)
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
	switch strings.ToLower(strings.TrimSpace(appconfig.Load().OlakeStorageMode)) {
	case constants.StorageModeS3:
		return readJSONFileFromS3(ctx, response)
	case constants.StorageModeNFS:
		workflowResponse, err := readJSONFileFromNFS(responsePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read workflow response: %v", err)
		}

		return workflowResponse, nil
	default:
		return nil, fmt.Errorf("unsupported OLAKE_STORAGE_MODE: %q", appconfig.Load().OlakeStorageMode)
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
