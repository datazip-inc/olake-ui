package temporal

import (
	"context"
	"fmt"
	"time"

	"github.com/beego/beego/v2/server/web"
	"github.com/datazip-inc/olake-ui/server/internal/constants"
	"github.com/datazip-inc/olake-ui/server/internal/models"
	"github.com/datazip-inc/olake-ui/server/internal/models/dto"
	"github.com/datazip-inc/olake-ui/server/utils/logger"
	"go.temporal.io/sdk/client"
	"golang.org/x/mod/semver"
)

type ExecutionRequest struct {
	Type          string        `json:"type"`
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

// DiscoverStreams runs a workflow to discover catalog data
func (t *Temporal) DiscoverStreams(ctx context.Context, sourceType, version, config, streamsConfig, jobName string) (map[string]interface{}, error) {
	workflowID := fmt.Sprintf("discover-catalog-%s-%d", sourceType, time.Now().Unix())

	configs := []JobConfig{
		{Name: "config.json", Data: config},
		{Name: "streams.json", Data: streamsConfig},
	}

	cmdArgs := []string{
		"discover",
		"--config",
		"/mnt/config/config.json",
	}

	if jobName != "" && semver.Compare(version, "v0.2.0") >= 0 {
		cmdArgs = append(cmdArgs, "--destination-database-prefix", jobName)
	}

	if streamsConfig != "" {
		cmdArgs = append(cmdArgs, "--catalog", "/mnt/config/streams.json")
	}

	if encryptionKey, _ := web.AppConfig.String(constants.ConfEncryptionKey); encryptionKey != "" {
		cmdArgs = append(cmdArgs, "--encryption-key", encryptionKey)
	}

	req := &ExecutionRequest{
		Type:          "docker",
		Command:       Discover,
		ConnectorType: sourceType,
		Version:       version,
		Args:          cmdArgs,
		Configs:       configs,
		WorkflowID:    workflowID,
		JobID:         0,
		Timeout:       GetWorkflowTimeout(Discover),
		OutputFile:    "streams.json",
	}

	workflowOptions := client.StartWorkflowOptions{
		ID:        workflowID,
		TaskQueue: t.taskQueue,
	}

	run, err := t.Client.ExecuteWorkflow(ctx, workflowOptions, ExecuteWorkflow, req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute discover workflow: %s", err)
	}

	result, err := ExtractWorkflowResponse(ctx, run)
	if err != nil {
		return nil, fmt.Errorf("failed to extract workflow response: %v", err)
	}

	return result, nil
}

// FetchSpec runs a workflow to fetch driver specifications
func (t *Temporal) GetDriverSpecs(ctx context.Context, destinationType, sourceType, version string) (dto.SpecOutput, error) {
	workflowID := fmt.Sprintf("fetch-spec-%s-%d", sourceType, time.Now().Unix())

	// spec version >= DefaultSpecVersion is required
	if semver.Compare(version, constants.DefaultSpecVersion) < 0 {
		version = constants.DefaultSpecVersion
	}

	cmdArgs := []string{
		"spec",
	}
	if destinationType != "" {
		cmdArgs = append(cmdArgs, "--destination-type", destinationType)
	}

	req := &ExecutionRequest{
		Type:          "docker",
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

// TestConnection runs a workflow to test connection
func (t *Temporal) VerifyDriverCredentials(ctx context.Context, workflowID, flag, sourceType, version, config string) (map[string]interface{}, error) {
	configs := []JobConfig{
		{Name: "config.json", Data: config},
	}

	cmdArgs := []string{
		"check",
		fmt.Sprintf("--%s", flag),
		"/mnt/config/config.json",
	}
	if encryptionKey, _ := web.AppConfig.String(constants.ConfEncryptionKey); encryptionKey != "" {
		cmdArgs = append(cmdArgs, "--encryption-key", encryptionKey)
	}

	req := &ExecutionRequest{
		Type:          "docker",
		Command:       Check,
		ConnectorType: sourceType,
		Version:       version,
		Args:          cmdArgs,
		Configs:       configs,
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

	handle := t.Client.ScheduleClient().GetHandle(ctx, scheduleID)
	if _, err := handle.Describe(ctx); err != nil {
		return fmt.Errorf("schedule does not exist: %s", err)
	}

	if err := t.PauseSchedule(ctx, job.ProjectID, job.ID); err != nil {
		return fmt.Errorf("failed to pause sync schedule: %s", err)
	}

	// update schedule to use clear-destination request
	clearReq := buildExecutionReqForClearDestination(job, workflowID, streamsConfig)
	err := t.UpdateScheduleAction(ctx, job.ProjectID, job.ID, clearReq)
	if err != nil {
		_ = t.UnpauseSchedule(ctx, job.ProjectID, job.ID)
		return fmt.Errorf("failed to update schedule for clear-destination: %s", err)
	}

	// the next schedule runs with the sync request
	defer func() {
		restoreCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 30*time.Second)
		defer cancel()

		syncReq := buildExecutionReqForSync(job, workflowID)
		if err := t.UpdateScheduleAction(restoreCtx, job.ProjectID, job.ID, syncReq); err != nil {
			logger.Errorf("failed to restore schedule to sync mode for job %d: %s", job.ID, err)
		}

		if err := t.UnpauseSchedule(restoreCtx, job.ProjectID, job.ID); err != nil {
			logger.Errorf("failed to unpause schedule for job %d: %s", job.ID, err)
		}
	}()

	return t.TriggerSchedule(ctx, job.ProjectID, job.ID)
}

// GetDifferenceStreams compares old and new stream configs and returns the difference
func (t *Temporal) GetDifferenceStreams(ctx context.Context, job *models.Job, oldConfig, newConfig string) (map[string]interface{}, error) {
	workflowID := fmt.Sprintf("difference-%s-%d-%d", job.ProjectID, job.ID, time.Now().Unix())

	configs := []JobConfig{
		{Name: "old_streams.json", Data: oldConfig},
		{Name: "new_streams.json", Data: newConfig},
	}

	cmdArgs := []string{
		"discover",
		"--streams", "/mnt/config/old_streams.json",
		"--difference", "/mnt/config/new_streams.json",
	}
	if encryptionKey, _ := web.AppConfig.String(constants.ConfEncryptionKey); encryptionKey != "" {
		cmdArgs = append(cmdArgs, "--encryption-key", encryptionKey)
	}

	req := &ExecutionRequest{
		Type:          "docker",
		Command:       Discover,
		ConnectorType: job.SourceID.Type,
		Version:       job.SourceID.Version,
		Args:          cmdArgs,
		Configs:       configs,
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
