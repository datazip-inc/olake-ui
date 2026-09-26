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

// usesSplitCatalog reports whether a job runs with the split catalog files
func usesSplitCatalog(job *models.Job) bool {
	return job.AvailableStreamsConfig != nil && job.SelectedStreamsConfig != nil && utils.SupportsSplitStreams(job.Source.Version)
}

// catalogFlags returns the catalog flags for the split or the legacy format
func catalogFlags(split bool) []string {
	if split {
		return []string{
			"--available-streams", "/mnt/config/" + constants.AvailableStreamsFile,
			"--selected-streams", "/mnt/config/" + constants.SelectedStreamsFile,
		}
	}
	return []string{"--streams", "/mnt/config/streams.json"}
}

// buildExecutionReqForSync builds the ExecutionRequest for a sync job
func buildExecutionReqForSync(job *models.Job, workflowID string) *ExecutionRequest {
	args := []string{
		"sync",
		"--config", "/mnt/config/source.json",
		"--destination", "/mnt/config/destination.json",
		"--state", "/mnt/config/state.json",
	}
	args = append(args, catalogFlags(usesSplitCatalog(job))...)

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
func buildExecutionReqForClearDestination(job *models.Job, workflowID, streamsConfig string) (*ExecutionRequest, error) {
	streamsDir := fmt.Sprintf("%s-%d", workflowID, time.Now().Unix())

	var files map[string]string
	var tempFile string
	splitCatalog := usesSplitCatalog(job)
	if splitCatalog {
		available, selected := utils.StringValue(job.AvailableStreamsConfig), utils.StringValue(job.SelectedStreamsConfig)
		if streamsConfig != "" {
			var err error
			if available, selected, err = utils.SplitCatalog(streamsConfig); err != nil {
				return nil, fmt.Errorf("failed to split clear-destination catalog: %s", err)
			}
		}
		// The CLI rejects empty split files, and a difference with no changed streams is empty:
		// stage that one in the legacy format, as legacy jobs do.
		splitCatalog = available != "" && selected != ""
		if splitCatalog {
			files = map[string]string{
				constants.AvailableStreamsFile: available,
				constants.SelectedStreamsFile:  selected,
			}
			tempFile = constants.SelectedStreamsFile
		}
	}
	if !splitCatalog {
		catalog := streamsConfig
		if catalog == "" {
			catalog = job.StreamsConfig
		}
		files = map[string]string{"streams.json": catalog}
		tempFile = "streams.json"
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
	args = append(args, catalogFlags(splitCatalog)...)
	relativePath := filepath.Join(streamsDir, tempFile)

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
