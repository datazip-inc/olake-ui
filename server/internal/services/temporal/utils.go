package temporal

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/datazip-inc/olake-ui/server/internal/models"
	"github.com/datazip-inc/olake-ui/server/utils/telemetry"
	"go.temporal.io/sdk/client"
)

// buildExecutionReqForSync builds the ExecutionRequest for a sync job
func buildExecutionReqForSync(job *models.Job, workflowID string) ExecutionRequest {
	args := []string{
		"sync",
		"--config", "/mnt/config/source.json",
		"--destination", "/mnt/config/destination.json",
		"--catalog", "/mnt/config/streams.json",
		"--state", "/mnt/config/state.json",
	}

	configs := []JobConfig{
		{Name: "source.json", Data: job.SourceID.Config},
		{Name: "destination.json", Data: job.DestID.Config},
		{Name: "streams.json", Data: job.StreamsConfig},
		{Name: "state.json", Data: job.State},
		{Name: "user_id.txt", Data: telemetry.GetTelemetryUserID()},
	}

	return ExecutionRequest{
		Command:       Sync,
		ConnectorType: job.SourceID.Type,
		Version:       job.SourceID.Version,
		Args:          args,
		Configs:       configs,
		WorkflowID:    workflowID,
		JobID:         job.ID,
		JobName:       job.Name,
		ProjectID:     job.ProjectID,
		Timeout:       GetWorkflowTimeout(Sync),
		OutputFile:    "state.json",
	}
}

// buildExecutionReqForClearDestination builds the ExecutionRequest for a clear-destination job
func buildExecutionReqForClearDestination(job *models.Job, workflowID, streamsConfig string) ExecutionRequest {
	catalog := streamsConfig
	if catalog == "" {
		catalog = job.StreamsConfig
	}

	configs := []JobConfig{
		{Name: "streams.json", Data: catalog},
		{Name: "state.json", Data: job.State},
		{Name: "destination.json", Data: job.DestID.Config},
	}

	args := []string{
		"clear-destination",
		"--streams", "/mnt/config/streams.json",
		"--state", "/mnt/config/state.json",
		"--destination", "/mnt/config/destination.json",
	}

	return ExecutionRequest{
		Command:       ClearDestination,
		ConnectorType: job.SourceID.Type,
		Version:       job.SourceID.Version,
		Args:          args,
		Configs:       configs,
		WorkflowID:    workflowID,
		ProjectID:     job.ProjectID,
		JobID:         job.ID,
		JobName:       job.Name,
		Timeout:       GetWorkflowTimeout(ClearDestination),
		OutputFile:    "state.json",
	}
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

	jsonResponse, err := ExtractJSON(response)
	if err != nil {
		return nil, err
	}

	return jsonResponse, nil
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

// ExtractJSON extracts and returns the last valid JSON block from output
func ExtractJSON(output string) (map[string]interface{}, error) {
	outputStr := strings.TrimSpace(output)
	if outputStr == "" {
		return nil, fmt.Errorf("empty output")
	}

	lines := strings.Split(outputStr, "\n")

	// Find the last non-empty line with valid JSON
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}

		start := strings.Index(line, "{")
		end := strings.LastIndex(line, "}")
		if start != -1 && end != -1 && end > start {
			jsonPart := line[start : end+1]
			var result map[string]interface{}
			if err := json.Unmarshal([]byte(jsonPart), &result); err != nil {
				continue // Skip invalid JSON
			}
			return result, nil
		}
	}

	return nil, fmt.Errorf("no valid JSON block found in output")
}
