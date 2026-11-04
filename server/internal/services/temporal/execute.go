package temporal

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/datazip-inc/olake-ui/server/internal/appconfig"
	"github.com/datazip-inc/olake-ui/server/internal/constants"
	"github.com/datazip-inc/olake-ui/server/internal/models"
	"github.com/datazip-inc/olake-ui/server/internal/storage"
	"github.com/datazip-inc/olake-ui/server/internal/types"
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
// StartDiscoverStreams starts discover with the job's stored catalog as input and returns the
// operation ID; with an empty catalog, discover runs fresh. The result is read with
// ReadDiscoverOutput once the operation has succeeded.
func (t *Temporal) StartDiscoverStreams(ctx context.Context, projectID, sourceType, version, config string, stored types.StreamsCatalog, jobName string, maxDiscoverThreads *int, targetQueryEngines []string) (string, error) {
	workflowID, err := NewOperationID(OperationDiscoverCatalog, projectID)
	if err != nil {
		return "", err
	}

	configs := []JobConfig{
		{Name: "config.json", Data: config},
		{Name: "user_id.txt", Data: telemetry.GetTelemetryUserID(ctx)},
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
		// The result is read in a later request that does not know the job, so record here that
		// the job is legacy: it keeps streams_config even when the driver also writes the split files.
		configs = append(configs,
			JobConfig{Name: constants.StreamsFile, Data: stored.Streams},
			JobConfig{Name: constants.LegacyJobMarkerFile, Data: "true"},
		)
		cmdArgs = append(cmdArgs, "--catalog", "/mnt/config/"+constants.StreamsFile)
	}

	if err := SetupConfigFiles(ctx, Discover, workflowID, configs); err != nil {
		return "", fmt.Errorf("failed to setup config files: %s", err)
	}

	if encryptionKey := appconfig.Load().EncryptionKey; encryptionKey != "" {
		cmdArgs = append(cmdArgs, "--encryption-key", encryptionKey)
	}

	// every driver writes streams.json, so it is the file the workflow reports
	if _, err := t.startDiscoverWorkflow(ctx, sourceType, version, workflowID, cmdArgs, constants.StreamsFile); err != nil {
		return "", err
	}
	return workflowID, nil
}

// ReadDiscoverOutput reads the catalog a succeeded discover operation wrote, and whether the
// operation was started for a legacy job.
func ReadDiscoverOutput(ctx context.Context, operationID string) (types.StreamsCatalog, bool, error) {
	catalog, err := readDiscoveredCatalog(ctx, operationID)
	if err != nil {
		return types.StreamsCatalog{}, false, err
	}
	// every driver writes streams.json, so a discover without it did not finish its output
	if catalog.Streams == "" {
		return types.StreamsCatalog{}, false, fmt.Errorf("discover wrote no %s", constants.StreamsFile)
	}
	_, err = storage.ReadFile(ctx, path.Join(operationID, constants.LegacyJobMarkerFile))
	switch {
	case err == nil:
		return catalog, true, nil
	case errors.Is(err, fs.ErrNotExist):
		return catalog, false, nil
	default:
		return types.StreamsCatalog{}, false, fmt.Errorf("failed to read %s: %s", constants.LegacyJobMarkerFile, err)
	}
}

// ConvertStreams converts a legacy streams.json into the split format with the CLI's offline
// conversion (discover --convert-streams), which reads the catalog the way every command reads a
// --streams input and never connects to the source. It returns the split catalog it wrote.
// It runs synchronously and its workflow ID is not an operation ID, so it cannot be polled
// through the operations API.
func (t *Temporal) ConvertStreams(ctx context.Context, jobID int, sourceType, version, streamsConfig string) (types.StreamsCatalog, error) {
	workflowID := fmt.Sprintf("convert-streams-%s-%d-%d", sourceType, jobID, time.Now().UnixNano())

	configs := []JobConfig{{Name: constants.StreamsFile, Data: streamsConfig}}
	if err := SetupConfigFiles(ctx, Discover, workflowID, configs); err != nil {
		return types.StreamsCatalog{}, fmt.Errorf("failed to setup config files: %s", err)
	}

	args := []string{"discover", "--streams", "/mnt/config/" + constants.StreamsFile, "--convert-streams"}
	run, err := t.startDiscoverWorkflow(ctx, sourceType, version, workflowID, args, constants.SelectedStreamsFile)
	if err != nil {
		return types.StreamsCatalog{}, err
	}
	if _, err := ExtractWorkflowResponse(ctx, run); err != nil {
		return types.StreamsCatalog{}, fmt.Errorf("failed to extract workflow response: %v", err)
	}
	catalog, err := readDiscoveredCatalog(ctx, workflowID)
	if err != nil {
		return types.StreamsCatalog{}, err
	}
	if !catalog.IsSplit() {
		return types.StreamsCatalog{}, fmt.Errorf("conversion wrote no %s and %s", constants.AvailableStreamsFile, constants.SelectedStreamsFile)
	}
	return catalog, nil
}

// startDiscoverWorkflow starts a discover workflow that writes outputFile, without waiting for it.
func (t *Temporal) startDiscoverWorkflow(ctx context.Context, sourceType, version, workflowID string, args []string, outputFile string) (client.WorkflowRun, error) {
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
		return nil, fmt.Errorf("failed to execute discover workflow: %s", err)
	}
	return run, nil
}

// readDiscoveredCatalog reads the catalog files that a discover or convert workflow wrote.
// Drivers below MinStreamsV2Version write only streams.json, so a missing file returns "".
func readDiscoveredCatalog(ctx context.Context, workflowID string) (types.StreamsCatalog, error) {
	read := func(name string) (string, error) {
		data, err := storage.ReadFile(ctx, path.Join(workflowID, name))
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

	var catalog types.StreamsCatalog
	var err error
	if catalog.Streams, err = read(constants.StreamsFile); err != nil {
		return types.StreamsCatalog{}, err
	}
	if catalog.Available, err = read(constants.AvailableStreamsFile); err != nil {
		return types.StreamsCatalog{}, err
	}
	if catalog.Selected, err = read(constants.SelectedStreamsFile); err != nil {
		return types.StreamsCatalog{}, err
	}
	if (catalog.Available == "") != (catalog.Selected == "") {
		return types.StreamsCatalog{}, fmt.Errorf("%s and %s must be written together", constants.AvailableStreamsFile, constants.SelectedStreamsFile)
	}
	return catalog, nil
}

// StartDriverSpecs returns an empty operation ID when no workflow was needed, in which
// case there is nothing to poll and the caller answers immediately.
func (t *Temporal) StartDriverSpecs(ctx context.Context, projectID, destinationType, sourceType, version string, availableQueryEngines bool) (string, error) {
	// An older image rejects the flag and fails the workflow, so report the feature as absent instead.
	if availableQueryEngines && !supportsQueryEngines(version) {
		return "", nil
	}

	workflowID, err := NewOperationID(OperationSpec, projectID)
	if err != nil {
		return "", err
	}

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

	if _, err := t.Client.ExecuteWorkflow(ctx, workflowOptions, ExecuteWorkflow, req); err != nil {
		return "", fmt.Errorf("failed to execute fetch spec workflow: %s", err)
	}

	return workflowID, nil
}

// supportsQueryEngines reports whether the version takes the engine flags; custom builds are assumed current.
func supportsQueryEngines(version string) bool {
	return utils.GetCustomDriverVersion() != "" || semver.Compare(version, constants.DefaultQueryEnginesVersion) >= 0
}

func (t *Temporal) StartVerifyDriverCredentials(ctx context.Context, projectID, flag, sourceType, version, config string) (string, error) {
	workflowID, err := NewOperationID(OperationTestConnection, projectID)
	if err != nil {
		return "", err
	}

	configs := []JobConfig{
		{Name: "config.json", Data: config},
	}

	if err := SetupConfigFiles(ctx, Check, workflowID, configs); err != nil {
		return "", fmt.Errorf("failed to setup config files: %s", err)
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

	if _, err := t.Client.ExecuteWorkflow(ctx, workflowOptions, ExecuteWorkflow, req); err != nil {
		return "", fmt.Errorf("failed to execute test connection workflow: %s", err)
	}

	return workflowID, nil
}

// DecodeConnectionStatus reshapes the check workflow's raw output into {message, status}.
func DecodeConnectionStatus(result map[string]interface{}) (map[string]interface{}, error) {
	connectionStatus, ok := result["connectionStatus"].(map[string]interface{})
	if !ok || connectionStatus == nil {
		return nil, fmt.Errorf("connection status not found")
	}

	status, statusOk := connectionStatus["status"].(string)
	message, _ := connectionStatus["message"].(string) // message is optional
	if !statusOk {
		return nil, fmt.Errorf("connection status not found")
	}

	response := map[string]interface{}{
		"message": message,
		"status":  status,
	}
	// prerequisites is optional: only source drivers with CDC checks report it
	if prerequisites, ok := connectionStatus["prerequisites"].([]interface{}); ok {
		response["prerequisites"] = prerequisites
	}
	return response, nil
}

func (t *Temporal) ClearDestination(ctx context.Context, job *models.Job, diff types.StreamsCatalog) error {
	workflowID, scheduleID := t.WorkflowAndScheduleID(job.ProjectID, job.ID)

	// update the sync schedule to use clear-destination request
	handle := t.Client.ScheduleClient().GetHandle(ctx, scheduleID)
	if _, err := handle.Describe(ctx); err != nil {
		return fmt.Errorf("schedule does not exist: %s", err)
	}

	// update schedule to use clear-destination request
	clearReq, err := buildExecutionReqForClearDestination(ctx, job, workflowID, diff)
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

// StartStreamDifference starts a comparison of the job's stored catalog with an edited one and
// returns the operation ID. Each catalog is passed in its own format. For a split new catalog OLake
// writes difference_available_streams.json and difference_selected_streams.json; the operation
// reports only the selected one.
func (t *Temporal) StartStreamDifference(ctx context.Context, job *models.Job, oldCatalog, newCatalog types.StreamsCatalog) (string, error) {
	workflowID, err := NewOperationID(OperationStreamDifference, job.ProjectID)
	if err != nil {
		return "", err
	}

	// A split new catalog always passes its available_streams: a newly discovered stream is selectable
	// only if it is in the new streams[], and default changes are read from there.
	configs, cmdArgs := stageCatalog(oldCatalog, "old_", "--streams", "--available-streams", "--selected-streams")
	newConfigs, newArgs := stageCatalog(newCatalog, "new_", "--difference", "--difference-available-streams", "--difference-selected-streams")
	configs = append(configs, newConfigs...)
	cmdArgs = append(append([]string{"discover"}, cmdArgs...), newArgs...)

	if job.Destination != nil {
		destinationConfig, err := utils.Encrypt(job.Destination.Config)
		if err != nil {
			return "", fmt.Errorf("failed to encrypt destination config: %s", err)
		}
		configs = append(configs, JobConfig{Name: "destination.json", Data: destinationConfig})
		cmdArgs = append(cmdArgs, "--destination", "/mnt/config/destination.json")
	}

	outputFile := "difference_streams.json"
	if newCatalog.IsSplit() {
		outputFile = "difference_selected_streams.json"
	}

	if err := SetupConfigFiles(ctx, Discover, workflowID, configs); err != nil {
		return "", fmt.Errorf("failed to setup config files: %s", err)
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
		OutputFile:    outputFile,
	}

	workflowOptions := client.StartWorkflowOptions{
		ID:        workflowID,
		TaskQueue: t.taskQueue,
	}

	if _, err := t.Client.ExecuteWorkflow(ctx, workflowOptions, ExecuteWorkflow, req); err != nil {
		return "", fmt.Errorf("failed to execute stream difference workflow: %s", err)
	}

	return workflowID, nil
}

// stageCatalog returns the config files of a catalog, named with prefix, and the discover flags that
// point at them: streamsFlag for a legacy catalog, availableFlag and selectedFlag for a split one.
func stageCatalog(catalog types.StreamsCatalog, prefix, streamsFlag, availableFlag, selectedFlag string) ([]JobConfig, []string) {
	if catalog.IsSplit() {
		return []JobConfig{
				{Name: prefix + constants.AvailableStreamsFile, Data: catalog.Available},
				{Name: prefix + constants.SelectedStreamsFile, Data: catalog.Selected},
			}, []string{
				availableFlag, "/mnt/config/" + prefix + constants.AvailableStreamsFile,
				selectedFlag, "/mnt/config/" + prefix + constants.SelectedStreamsFile,
			}
	}
	return []JobConfig{{Name: prefix + constants.StreamsFile, Data: catalog.Streams}},
		[]string{streamsFlag, "/mnt/config/" + prefix + constants.StreamsFile}
}
