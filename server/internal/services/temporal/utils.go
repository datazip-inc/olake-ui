package temporal

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/datazip-inc/olake-ui/server/internal/constants"
	"github.com/datazip-inc/olake-ui/server/internal/models"
	"github.com/datazip-inc/olake-ui/server/utils"
	"go.temporal.io/sdk/client"
)

// buildExecutionReqForSync builds the ExecutionRequest for a sync job
func buildExecutionReqForSync(job *models.Job, workflowID string) *ExecutionRequest {
	args := []string{
		"sync",
		"--config", "/mnt/config/source.json",
		"--destination", "/mnt/config/destination.json",
		"--catalog", "/mnt/config/streams.json",
		"--state", "/mnt/config/state.json",
	}

	return &ExecutionRequest{
		Command:       Sync,
		ConnectorType: job.SourceID.Type,
		Version:       job.SourceID.Version,
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
func buildExecutionReqForClearDestination(job *models.Job, workflowID, streamsConfig string) (*ExecutionRequest, error) {
	catalog := streamsConfig
	if catalog == "" {
		catalog = job.StreamsConfig
	}

	streamsDir := fmt.Sprintf("%s-%d", workflowID, time.Now().Unix())
	relativePath := filepath.Join(streamsDir, "streams.json")
	streamsPath := filepath.Join(constants.DefaultConfigDir, relativePath)

	if err := utils.WriteFile(streamsPath, []byte(catalog), 0644); err != nil {
		return nil, fmt.Errorf("failed to write streams config to file: %v", err)
	}

	args := []string{
		"clear-destination",
		"--streams", "/mnt/config/streams.json",
		"--state", "/mnt/config/state.json",
		"--destination", "/mnt/config/destination.json",
	}

	return &ExecutionRequest{
		Command:       ClearDestination,
		ConnectorType: job.SourceID.Type,
		Version:       job.SourceID.Version,
		Args:          args,
		Configs:       nil,
		WorkflowID:    workflowID,
		ProjectID:     job.ProjectID,
		JobID:         job.ID,
		Timeout:       GetWorkflowTimeout(ClearDestination),
		OutputFile:    "state.json",
		Options: ExecutionOptions{
			TempPath: relativePath,
		},
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
	workflowResponse, err := ReadJSONFile(responsePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read workflow response: %v", err)
	}

	return workflowResponse, nil
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
