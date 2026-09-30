package temporal

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/datazip-inc/olake-ui/server/internal/constants"
	"github.com/datazip-inc/olake-ui/server/internal/models"
	"github.com/datazip-inc/olake-ui/server/internal/utils"
	"go.temporal.io/sdk/client"
)

// buildExecutionReqForSync builds the ExecutionRequest for a sync job.
// Split jobs use --available-streams/--selected-streams; legacy jobs use --streams.
// Configs for sync are written by the worker, so only the flag style is set here.
func buildExecutionReqForSync(job *models.Job, workflowID string) *ExecutionRequest {
	args := []string{
		"sync",
		"--config", "/mnt/config/source.json",
		"--destination", "/mnt/config/destination.json",
		"--state", "/mnt/config/state.json",
	}
	if job.IsStreamsV2() {
		args = append(args,
			"--available-streams", "/mnt/config/"+constants.AvailableStreamsFile,
			"--selected-streams", "/mnt/config/"+constants.SelectedStreamsFile,
		)
	} else {
		args = append(args, "--streams", "/mnt/config/"+constants.StreamsFile)
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

// buildExecutionReqForClearDestination builds the ExecutionRequest for a clear-destination job.
// streamsConfig is the stream difference to clear, in the combined {streams, selected_streams}
// format that every CLI reads with --streams; empty clears the job's whole stored catalog, which
// a v2 job keeps only in the split files.
func buildExecutionReqForClearDestination(job *models.Job, workflowID, streamsConfig string) (*ExecutionRequest, error) {
	streamsDir := fmt.Sprintf("%s-%d", workflowID, time.Now().Unix())

	var catalogArgs []string
	var tempPath string
	var files map[string]string

	if streamsConfig == "" && job.IsStreamsV2() {
		files = map[string]string{
			constants.AvailableStreamsFile: utils.StringValue(job.AvailableStreamsConfig),
			constants.SelectedStreamsFile:  utils.StringValue(job.SelectedStreamsConfig),
		}
		catalogArgs = []string{
			"--available-streams", "/mnt/config/" + constants.AvailableStreamsFile,
			"--selected-streams", "/mnt/config/" + constants.SelectedStreamsFile,
		}
		tempPath = filepath.Join(streamsDir, constants.SelectedStreamsFile)
	} else {
		catalog := streamsConfig
		if catalog == "" {
			catalog = utils.StringValue(job.StreamsConfig)
		}
		files = map[string]string{constants.StreamsFile: catalog}
		catalogArgs = []string{"--streams", "/mnt/config/" + constants.StreamsFile}
		tempPath = filepath.Join(streamsDir, constants.StreamsFile)
	}

	for name, data := range files {
		path := filepath.Join(constants.DefaultConfigDir, streamsDir, name)
		if err := utils.WriteFile(path, []byte(data), constants.DefaultFileMode); err != nil {
			return nil, fmt.Errorf("failed to write %s to file: %v", name, err)
		}
	}

	args := []string{
		"clear-destination",
		"--state", "/mnt/config/state.json",
		"--destination", "/mnt/config/destination.json",
	}
	args = append(args, catalogArgs...)

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
		TempPath:      tempPath,
	}, nil
}

// ExtractWorkflowResponse extracts and parses the JSON response from a workflow execution result
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
	workflowResponse, err := utils.ReadJSONFile(responsePath)
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
