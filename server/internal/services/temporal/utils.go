package temporal

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/datazip-inc/olake-ui/server/internal/models"
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

	return ExecutionRequest{
		Type:          "docker",
		Command:       Sync,
		ConnectorType: job.SourceID.Type,
		Version:       job.SourceID.Version,
		Args:          args,
		WorkflowID:    workflowID,
		JobID:         job.ID,
		Timeout:       GetWorkflowTimeout(Sync),
		OutputFile:    "state.json",
	}
}

// extractWorkflowResponse extracts and parses the JSON response from a workflow execution result
func ExtractWorkflowResponse(ctx context.Context, run client.WorkflowRun) (map[string]interface{}, error) {
	var result map[string]interface{}
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
