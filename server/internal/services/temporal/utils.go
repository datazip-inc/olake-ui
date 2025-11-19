package temporal

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/datazip-inc/olake-ui/server/internal/constants"
	"github.com/datazip-inc/olake-ui/server/internal/models"
	"github.com/datazip-inc/olake-ui/server/utils"
	"go.temporal.io/sdk/client"
)

// buildExecutionReqForSync builds the ExecutionRequest for a sync job
func buildExecutionReqForSync(job *models.Job, workflowID string) (*ExecutionRequest, error) {
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
		Timeout:       GetWorkflowTimeout(Sync),
		OutputFile:    "state.json",
	}, nil
}

// buildExecutionReqForClearDestination builds the ExecutionRequest for a clear-destination job
func buildExecutionReqForClearDestination(job *models.Job, workflowID, streamsConfig string) (*ExecutionRequest, error) {
	catalog := streamsConfig
	if catalog == "" {
		catalog = job.StreamsConfig
	}
	configFile := fmt.Sprintf("clear-destination-%s-%d", job.ProjectID, job.ID)
	configPath := filepath.Join(constants.DefaultConfigDir, configFile, "streams.json")

	if err := utils.WriteFile(configPath, []byte(catalog), 0644); err != nil {
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
	}, nil
}

// extractWorkflowResponse extracts and parses the JSON response from a workflow execution result
func ExtractWorkflowResponse(ctx context.Context, run client.WorkflowRun) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	if err := run.Get(ctx, &result); err != nil {
		return nil, fmt.Errorf("workflow execution failed: %v", err)
	}

	response, ok := result["response"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid response format from worker")
	}

	filePath := filepath.Join(constants.DefaultConfigDir, response)
	fileOutput, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %v", err)
	}

	var workflowResponse map[string]interface{}
	if err := json.Unmarshal(fileOutput, &workflowResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal file: %v", err)
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
