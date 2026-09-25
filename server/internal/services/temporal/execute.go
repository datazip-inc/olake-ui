package temporal

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/datazip-inc/olake-ui/server/internal/appconfig"
	"github.com/datazip-inc/olake-ui/server/internal/constants"
	"github.com/datazip-inc/olake-ui/server/internal/models"
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

// StartDiscoverStreams starts a catalog discovery and returns its operation ID without
// waiting for it to finish. Callers poll the operation instead of holding the request
func (t *Temporal) StartDiscoverStreams(ctx context.Context, projectID, sourceType, version, config, streamsConfig, jobName string, maxDiscoverThreads *int) (string, error) {
	workflowID, err := NewOperationID(OperationDiscoverCatalog, projectID)
	if err != nil {
		return "", err
	}

	configs := []JobConfig{
		{Name: "config.json", Data: config},
		{Name: "streams.json", Data: streamsConfig},
		{Name: "user_id.txt", Data: telemetry.GetTelemetryUserID()},
	}

	if err := SetupConfigFiles(Discover, workflowID, configs); err != nil {
		return "", fmt.Errorf("failed to setup config files: %s", err)
	}

	cmdArgs := []string{
		"discover",
		"--config",
		"/mnt/config/config.json",
	}

	if jobName != "" && (utils.GetCustomDriverVersion() != "" || semver.Compare(version, "v0.2.0") >= 0) {
		cmdArgs = append(cmdArgs, "--destination-database-prefix", jobName)
	}

	// Only add max-discover-threads flag for versions >= v0.3.18
	if semver.Compare(version, constants.DefaultMaxDiscoverThreadsVersion) >= 0 {
		threads := constants.DefaultMaxDiscoverThreads
		if maxDiscoverThreads != nil && *maxDiscoverThreads > 0 {
			threads = *maxDiscoverThreads
		}
		cmdArgs = append(cmdArgs, constants.MaxDiscoverThreadsFlag, strconv.Itoa(threads))
	}

	if streamsConfig != "" {
		cmdArgs = append(cmdArgs, "--catalog", "/mnt/config/streams.json")
	}

	if encryptionKey := appconfig.Load().EncryptionKey; encryptionKey != "" {
		cmdArgs = append(cmdArgs, "--encryption-key", encryptionKey)
	}

	req := &ExecutionRequest{
		Command:       Discover,
		ConnectorType: sourceType,
		Version:       version,
		Args:          cmdArgs,
		Configs:       nil,
		WorkflowID:    workflowID,
		JobID:         0,
		Timeout:       GetWorkflowTimeout(Discover),
		OutputFile:    "streams.json",
	}

	workflowOptions := client.StartWorkflowOptions{
		ID:        workflowID,
		TaskQueue: t.taskQueue,
	}

	if _, err := t.Client.ExecuteWorkflow(ctx, workflowOptions, ExecuteWorkflow, req); err != nil {
		return "", fmt.Errorf("failed to execute discover workflow: %s", err)
	}

	return workflowID, nil
}

// StartDriverSpecs starts a spec fetch and returns its operation ID without waiting.
// specType and specVersion are what the caller asked for, which is not always what the
// workflow runs: a destination spec runs inside a source connector image. They are
// recorded on the workflow memo so the result endpoint can echo them back verbatim.
func (t *Temporal) StartDriverSpecs(ctx context.Context, projectID, destinationType, sourceType, version, specType, specVersion string) (string, error) {
	workflowID, err := NewOperationID(OperationSpec, projectID)
	if err != nil {
		return "", err
	}

	// spec version >= DefaultSpecVersion is required
	if semver.Compare(version, constants.DefaultSpecVersion) < 0 && utils.GetCustomDriverVersion() == "" {
		version = constants.DefaultSpecVersion
	}

	cmdArgs := []string{
		"spec",
	}
	if destinationType != "" {
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
		Memo: map[string]interface{}{
			MemoSpecType:    specType,
			MemoSpecVersion: specVersion,
		},
	}

	if _, err := t.Client.ExecuteWorkflow(ctx, workflowOptions, ExecuteWorkflow, req); err != nil {
		return "", fmt.Errorf("failed to execute fetch spec workflow: %s", err)
	}

	return workflowID, nil
}

// StartVerifyDriverCredentials starts a connection check and returns its operation ID
// without waiting. The operation ID doubles as the directory name under the shared
// config volume, which is where the check's logs land.
func (t *Temporal) StartVerifyDriverCredentials(ctx context.Context, projectID, flag, sourceType, version, config string) (string, error) {
	workflowID, err := NewOperationID(OperationTestConnection, projectID)
	if err != nil {
		return "", err
	}

	configs := []JobConfig{
		{Name: "config.json", Data: config},
	}

	if err := SetupConfigFiles(Check, workflowID, configs); err != nil {
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

// DecodeConnectionStatus reshapes a check workflow's raw output into the {message, status}
// pair the API has always returned.the workflow that produces the raw shape,
// and is applied when the result is collected rather than when it is started.
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

// StartStreamDifference starts a comparison of old and new stream configs and returns
// its operation ID without waiting.
func (t *Temporal) StartStreamDifference(ctx context.Context, job *models.Job, oldConfig, newConfig string) (string, error) {
	workflowID, err := NewOperationID(OperationStreamDifference, job.ProjectID)
	if err != nil {
		return "", err
	}

	configs := []JobConfig{
		{Name: "old_streams.json", Data: oldConfig},
		{Name: "new_streams.json", Data: newConfig},
	}

	if err := SetupConfigFiles(Discover, workflowID, configs); err != nil {
		return "", fmt.Errorf("failed to setup config files: %s", err)
	}

	cmdArgs := []string{
		"discover",
		"--streams", "/mnt/config/old_streams.json",
		"--difference", "/mnt/config/new_streams.json",
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

	if _, err := t.Client.ExecuteWorkflow(ctx, workflowOptions, ExecuteWorkflow, req); err != nil {
		return "", fmt.Errorf("failed to execute stream difference workflow: %s", err)
	}

	return workflowID, nil
}
