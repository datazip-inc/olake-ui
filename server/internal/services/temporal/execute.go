package temporal

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/datazip-inc/olake-ui/server/internal/appconfig"
	"github.com/datazip-inc/olake-ui/server/internal/constants"
	"github.com/datazip-inc/olake-ui/server/internal/models"
	"github.com/datazip-inc/olake-ui/server/internal/models/dto"
	"github.com/datazip-inc/olake-ui/server/internal/utils"
	"github.com/datazip-inc/olake-ui/server/internal/utils/telemetry"
	"go.temporal.io/sdk/client"
	"golang.org/x/mod/semver"
)

type ExecutionRequest struct {
	Command       Command       `json:"command"`
	ConnectorType string        `json:"connector_type"`
	Version       string        `json:"version"`
	Args          []string      `json:"args"`
	Configs       []JobConfig   `json:"configs"`
	WorkflowID    string        `json:"workflow_id"`
	ProjectID     string        `json:"project_id"`
	JobID         int           `json:"job_id"`
	Timeout       time.Duration `json:"timeout"`
	OutputFile    string        `json:"output_file"` // to get the output file from the workflow

	TempPath string `json:"temp_path"`
}

type JobConfig struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

type Command string

const (
	Discover         Command = "discover"
	Check            Command = "check"
	Sync             Command = "sync"
	Spec             Command = "spec"
	ClearDestination Command = "clear-destination"

	RunSyncWorkflow = "RunSyncWorkflow"
	ExecuteWorkflow = "ExecuteWorkflow"
)

// TODO: check if we can add command args as constants for all the methods

// Scheduled Process (sync, clear-destination):
// For scheduled/triggered workflows, configs are written in the worker-side.
// This is because the BFF doesn't know the actual execution workflowID that
// Temporal will use (the worker gets it from workflow.GetInfo), and the worker
// needs to write configs using the correct execution workflowID for directory
// computation. The worker reads configs from DB and merges with any configs
// provided in the request payload.
//
// Direct Execution (discover, spec, check, stream-difference):
// For direct execution, configs are written in the client-side (BFF) before
// sending to Temporal. The BFF knows the workflowID upfront and can write
// files to the correct directory, avoiding large payloads in Temporal.
//
// ref: https://docs.temporal.io/troubleshooting/blob-size-limit-error
// DiscoverStreams runs discover with the job's stored catalog as input; with an empty catalog,
// discover runs fresh.
func (t *Temporal) DiscoverStreams(ctx context.Context, sourceType, version, config string, stored StreamsCatalog, jobName string, maxDiscoverThreads *int, targetQueryEngines []string) (StreamsCatalog, error) {
	workflowID := fmt.Sprintf("discover-catalog-%s-%d", sourceType, time.Now().UnixNano())

	configs := []JobConfig{
		{Name: "config.json", Data: config},
		{Name: "user_id.txt", Data: telemetry.GetTelemetryUserID()},
	}
	cmdArgs := []string{
		"discover",
		"--config",
		"/mnt/config/config.json",
	}

	if jobName != "" && utils.SupportsDestinationDatabasePrefix(version) {
		cmdArgs = append(cmdArgs, "--destination-database-prefix", jobName)
	}

	if utils.SupportsMaxDiscoverThreads(version) {
		threads := constants.DefaultMaxDiscoverThreads
		if maxDiscoverThreads != nil && *maxDiscoverThreads > 0 {
			threads = *maxDiscoverThreads
		}
		cmdArgs = append(cmdArgs, constants.MaxDiscoverThreadsFlag, strconv.Itoa(threads))
	}

	// OLake stores no engines, so an omitted flag means unconstrained rather than "reuse the last choice".
	if len(targetQueryEngines) > 0 && supportsQueryEngines(version) {
		cmdArgs = append(cmdArgs, constants.TargetQueryEnginesFlag, strings.Join(targetQueryEngines, ","))
	}

	switch {
	case stored.IsSplit():
		configs = append(configs,
			JobConfig{Name: constants.AvailableStreamsFile, Data: stored.Available},
			JobConfig{Name: constants.SelectedStreamsFile, Data: stored.Selected},
		)
		cmdArgs = append(cmdArgs,
			"--available-streams", "/mnt/config/"+constants.AvailableStreamsFile,
			"--selected-streams", "/mnt/config/"+constants.SelectedStreamsFile,
		)
	case stored.Streams != "":
		configs = append(configs, JobConfig{Name: constants.StreamsFile, Data: stored.Streams})
		cmdArgs = append(cmdArgs, "--catalog", "/mnt/config/"+constants.StreamsFile)
	}

	if err := SetupConfigFiles(Discover, workflowID, configs); err != nil {
		return StreamsCatalog{}, fmt.Errorf("failed to setup config files: %s", err)
	}

	if encryptionKey := appconfig.Load().EncryptionKey; encryptionKey != "" {
		cmdArgs = append(cmdArgs, "--encryption-key", encryptionKey)
	}

	// every driver writes streams.json, so it is the file the workflow waits for
	if err := t.discoverWorkflow(ctx, sourceType, version, workflowID, cmdArgs, constants.StreamsFile); err != nil {
		return StreamsCatalog{}, err
	}
	discovered, err := readDiscoveredCatalog(workflowID)
	if err != nil {
		return StreamsCatalog{}, err
	}
	if discovered.Streams == "" {
		return StreamsCatalog{}, fmt.Errorf("discover wrote no %s", constants.StreamsFile)
	}
	return discovered, nil
}

// ConvertStreams converts a legacy streams.json into the split format with the CLI's offline
// conversion (discover --convert-streams), which reads the catalog the way every command reads a
// --streams input and never connects to the source. It returns the split catalog it wrote.
func (t *Temporal) ConvertStreams(ctx context.Context, sourceType, version, streamsConfig string) (StreamsCatalog, error) {
	workflowID := fmt.Sprintf("convert-streams-%s-%d", sourceType, time.Now().UnixNano())

	configs := []JobConfig{{Name: constants.StreamsFile, Data: streamsConfig}}
	if err := SetupConfigFiles(Discover, workflowID, configs); err != nil {
		return StreamsCatalog{}, fmt.Errorf("failed to setup config files: %s", err)
	}

	args := []string{"discover", "--streams", "/mnt/config/" + constants.StreamsFile, "--convert-streams"}
	if err := t.discoverWorkflow(ctx, sourceType, version, workflowID, args, constants.SelectedStreamsFile); err != nil {
		return StreamsCatalog{}, err
	}
	catalog, err := readDiscoveredCatalog(workflowID)
	if err != nil {
		return StreamsCatalog{}, err
	}
	if !catalog.IsSplit() {
		return StreamsCatalog{}, fmt.Errorf("conversion wrote no %s and %s", constants.AvailableStreamsFile, constants.SelectedStreamsFile)
	}
	return catalog, nil
}

// discoverWorkflow runs a discover workflow and waits for it to write outputFile.
func (t *Temporal) discoverWorkflow(ctx context.Context, sourceType, version, workflowID string, args []string, outputFile string) error {
	req := &ExecutionRequest{
		Command:       Discover,
		ConnectorType: sourceType,
		Version:       version,
		Args:          args,
		WorkflowID:    workflowID,
		Timeout:       GetWorkflowTimeout(Discover),
		OutputFile:    outputFile,
	}
	run, err := t.Client.ExecuteWorkflow(ctx, client.StartWorkflowOptions{ID: workflowID, TaskQueue: t.taskQueue}, ExecuteWorkflow, req)
	if err != nil {
		return fmt.Errorf("failed to execute discover workflow: %s", err)
	}
	if _, err := ExtractWorkflowResponse(ctx, run); err != nil {
		return fmt.Errorf("failed to extract workflow response: %v", err)
	}
	return nil
}

// readDiscoveredCatalog reads the catalog files a discover or a conversion wrote in the workflow
// directory, as written. A file the CLI did not write stays empty: a driver below
// MinStreamsV2Version writes only streams.json, a newer one also writes the split files.
func readDiscoveredCatalog(workflowID string) (StreamsCatalog, error) {
	read := func(name string) (string, error) {
		data, err := os.ReadFile(filepath.Join(constants.DefaultConfigDir, workflowID, name))
		if errors.Is(err, fs.ErrNotExist) {
			return "", nil
		}
		if err != nil {
			return "", fmt.Errorf("failed to read %s: %s", name, err)
		}
		if !json.Valid(data) {
			return "", fmt.Errorf("%s is not valid JSON", name)
		}
		return string(data), nil
	}

	var catalog StreamsCatalog
	var err error
	if catalog.Streams, err = read(constants.StreamsFile); err != nil {
		return StreamsCatalog{}, err
	}
	if catalog.Available, err = read(constants.AvailableStreamsFile); err != nil {
		return StreamsCatalog{}, err
	}
	if catalog.Selected, err = read(constants.SelectedStreamsFile); err != nil {
		return StreamsCatalog{}, err
	}
	if (catalog.Available == "") != (catalog.Selected == "") {
		return StreamsCatalog{}, fmt.Errorf("%s and %s must be written together", constants.AvailableStreamsFile, constants.SelectedStreamsFile)
	}
	return catalog, nil
}

// FetchSpec runs a workflow to fetch driver specifications
func (t *Temporal) GetDriverSpecs(ctx context.Context, destinationType, sourceType, version string, availableQueryEngines bool) (dto.SpecOutput, error) {
	// An older image rejects the flag and fails the workflow, so report the feature as absent instead.
	if availableQueryEngines && !supportsQueryEngines(version) {
		return dto.SpecOutput{}, nil
	}

	workflowID := fmt.Sprintf("fetch-spec-%s-%d", sourceType, time.Now().Unix())

	version = utils.ResolveSpecVersion(version)

	cmdArgs := []string{
		"spec",
	}
	// Exclusive: --available-query-engines exits before any spec file resolves, ignoring --destination-type.
	if availableQueryEngines {
		cmdArgs = append(cmdArgs, constants.AvailableQueryEnginesFlag)
	} else if destinationType != "" {
		cmdArgs = append(cmdArgs, "--destination-type", destinationType)
	}

	req := &ExecutionRequest{
		Command:       Spec,
		ConnectorType: sourceType,
		Version:       version,
		Args:          cmdArgs,
		Configs:       nil,
		WorkflowID:    workflowID,
		JobID:         0,
		Timeout:       GetWorkflowTimeout(Spec),
		OutputFile:    "",
	}

	workflowOptions := client.StartWorkflowOptions{
		ID:        workflowID,
		TaskQueue: t.taskQueue,
	}

	run, err := t.Client.ExecuteWorkflow(ctx, workflowOptions, ExecuteWorkflow, req)
	if err != nil {
		return dto.SpecOutput{}, fmt.Errorf("failed to execute fetch spec workflow: %s", err)
	}

	result, err := ExtractWorkflowResponse(ctx, run)
	if err != nil {
		return dto.SpecOutput{}, fmt.Errorf("failed to extract workflow response: %v", err)
	}

	return dto.SpecOutput{
		Spec: result,
	}, nil
}

// supportsQueryEngines reports whether the version takes the engine flags; custom builds are assumed current.
func supportsQueryEngines(version string) bool {
	return utils.GetCustomDriverVersion() != "" || semver.Compare(version, constants.DefaultQueryEnginesVersion) >= 0
}

// TestConnection runs a workflow to test connection
func (t *Temporal) VerifyDriverCredentials(ctx context.Context, workflowID, flag, sourceType, version, config string) (map[string]interface{}, error) {
	configs := []JobConfig{
		{Name: "config.json", Data: config},
	}

	if err := SetupConfigFiles(Check, workflowID, configs); err != nil {
		return nil, fmt.Errorf("failed to setup config files: %s", err)
	}

	cmdArgs := []string{
		"check",
		fmt.Sprintf("--%s", flag),
		"/mnt/config/config.json",
	}
	if encryptionKey := appconfig.Load().EncryptionKey; encryptionKey != "" {
		cmdArgs = append(cmdArgs, "--encryption-key", encryptionKey)
	}

	req := &ExecutionRequest{
		Command:       Check,
		ConnectorType: sourceType,
		Version:       version,
		Args:          cmdArgs,
		Configs:       nil,
		WorkflowID:    workflowID,
		Timeout:       GetWorkflowTimeout(Check),
	}

	workflowOptions := client.StartWorkflowOptions{
		ID:        workflowID,
		TaskQueue: t.taskQueue,
	}

	run, err := t.Client.ExecuteWorkflow(ctx, workflowOptions, ExecuteWorkflow, req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute test connection workflow: %s", err)
	}

	result, err := ExtractWorkflowResponse(ctx, run)
	if err != nil {
		return nil, fmt.Errorf("failed to extract workflow response: %v", err)
	}

	connectionStatus, ok := result["connectionStatus"].(map[string]interface{})
	if !ok || connectionStatus == nil {
		return nil, fmt.Errorf("connection status not found")
	}

	status, statusOk := connectionStatus["status"].(string)
	message, _ := connectionStatus["message"].(string) // message is optional
	if !statusOk {
		return nil, fmt.Errorf("connection status not found")
	}

	return map[string]interface{}{
		"message": message,
		"status":  status,
	}, nil
}

func (t *Temporal) ClearDestination(ctx context.Context, job *models.Job, streamsConfig string) error {
	workflowID, scheduleID := t.WorkflowAndScheduleID(job.ProjectID, job.ID)

	// update the sync schedule to use clear-destination request
	handle := t.Client.ScheduleClient().GetHandle(ctx, scheduleID)
	if _, err := handle.Describe(ctx); err != nil {
		return fmt.Errorf("schedule does not exist: %s", err)
	}

	// update schedule to use clear-destination request
	clearReq, err := buildExecutionReqForClearDestination(job, workflowID, streamsConfig)
	if err != nil {
		return fmt.Errorf("failed to build execution request for clear-destination: %s", err)
	}

	err = t.UpdateSchedule(ctx, job.Frequency, job.ProjectID, job.ID, clearReq)
	if err != nil {
		return fmt.Errorf("failed to update schedule for clear-destination: %s", err)
	}

	if err := t.TriggerSchedule(ctx, job.ProjectID, job.ID); err != nil {
		// revert back to sync
		syncReq := buildExecutionReqForSync(job, workflowID)
		if uerr := t.UpdateSchedule(ctx, job.Frequency, job.ProjectID, job.ID, syncReq); uerr != nil {
			return fmt.Errorf("trigger clear destination workflow failed: %s, revert to sync failed: %s", err, uerr)
		}
		return fmt.Errorf("failed to trigger clear destination workflow: %s", err)
	}
	return nil
}

// GetStreamDifference compares the job's stored catalog with an edited one and returns the
// difference. Each catalog is passed in its own format.
func (t *Temporal) GetStreamDifference(ctx context.Context, job *models.Job, oldCatalog, newCatalog StreamsCatalog) (map[string]interface{}, error) {
	workflowID := fmt.Sprintf("difference-%s-%d-%d", job.ProjectID, job.ID, time.Now().UnixNano())

	var configs []JobConfig
	cmdArgs := []string{"discover"}
	if oldCatalog.IsSplit() {
		configs = append(configs,
			JobConfig{Name: "old_" + constants.AvailableStreamsFile, Data: oldCatalog.Available},
			JobConfig{Name: "old_" + constants.SelectedStreamsFile, Data: oldCatalog.Selected},
		)
		cmdArgs = append(cmdArgs,
			"--available-streams", "/mnt/config/old_"+constants.AvailableStreamsFile,
			"--selected-streams", "/mnt/config/old_"+constants.SelectedStreamsFile,
		)
	} else {
		configs = append(configs, JobConfig{Name: "old_" + constants.StreamsFile, Data: oldCatalog.Streams})
		cmdArgs = append(cmdArgs, "--streams", "/mnt/config/old_"+constants.StreamsFile)
	}
	if newCatalog.IsSplit() {
		configs = append(configs,
			JobConfig{Name: "new_" + constants.AvailableStreamsFile, Data: newCatalog.Available},
			JobConfig{Name: "new_" + constants.SelectedStreamsFile, Data: newCatalog.Selected},
		)
		cmdArgs = append(cmdArgs,
			"--difference-available-streams", "/mnt/config/new_"+constants.AvailableStreamsFile,
			"--difference-selected-streams", "/mnt/config/new_"+constants.SelectedStreamsFile,
		)
	} else {
		configs = append(configs, JobConfig{Name: "new_" + constants.StreamsFile, Data: newCatalog.Streams})
		cmdArgs = append(cmdArgs, "--difference", "/mnt/config/new_"+constants.StreamsFile)
	}
	if err := SetupConfigFiles(Discover, workflowID, configs); err != nil {
		return nil, fmt.Errorf("failed to setup config files: %s", err)
	}

	if encryptionKey := appconfig.Load().EncryptionKey; encryptionKey != "" {
		cmdArgs = append(cmdArgs, "--encryption-key", encryptionKey)
	}

	req := &ExecutionRequest{
		Command:       Discover,
		ConnectorType: job.Source.Type,
		Version:       job.Source.Version,
		Args:          cmdArgs,
		Configs:       nil,
		WorkflowID:    workflowID,
		JobID:         job.ID,
		Timeout:       GetWorkflowTimeout(Discover),
		OutputFile:    "difference_streams.json",
	}

	workflowOptions := client.StartWorkflowOptions{
		ID:        workflowID,
		TaskQueue: t.taskQueue,
	}

	run, err := t.Client.ExecuteWorkflow(ctx, workflowOptions, ExecuteWorkflow, req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute stream difference workflow: %s", err)
	}

	result, err := ExtractWorkflowResponse(ctx, run)
	if err != nil {
		return nil, fmt.Errorf("failed to extract workflow response: %v", err)
	}

	return result, nil
}
