package temporal

import (
	"context"
	"fmt"
	"time"

	"github.com/datazip/olake-ui/server/internal/constants"
	"github.com/datazip/olake-ui/server/internal/docker"
	"github.com/datazip/olake-ui/server/internal/models/dto"
	"go.temporal.io/sdk/client"
	"golang.org/x/mod/semver"
)

// GetCatalog runs a workflow to discover catalog data
func (t *Temporal) GetCatalog(ctx context.Context, sourceType, version, config, streamsConfig, jobName string) (map[string]interface{}, error) {
	params := &ActivityParams{
		SourceType:    sourceType,
		Version:       version,
		Config:        config,
		WorkflowID:    fmt.Sprintf("discover-catalog-%s-%d", sourceType, time.Now().Unix()),
		Command:       docker.Discover,
		StreamsConfig: streamsConfig,
		JobName:       jobName,
	}

	workflowOptions := client.StartWorkflowOptions{
		ID:        params.WorkflowID,
		TaskQueue: t.taskQueue,
	}

	run, err := t.Client.ExecuteWorkflow(ctx, workflowOptions, DiscoverCatalogWorkflow, params)
	if err != nil {
		return nil, fmt.Errorf("failed to execute discover workflow: %s", err)
	}

	var result map[string]interface{}
	if err := run.Get(ctx, &result); err != nil {
		return nil, fmt.Errorf("workflow execution failed: %s", err)
	}

	return result, nil
}

// FetchSpec runs a workflow to fetch connector specifications
func (t *Temporal) FetchSpec(ctx context.Context, destinationType, sourceType, version string) (dto.SpecOutput, error) {
	// spec version >= DefaultSpecVersion is required
	if semver.Compare(version, constants.DefaultSpecVersion) < 0 {
		version = constants.DefaultSpecVersion
	}

	params := &ActivityParams{
		SourceType:      sourceType,
		Version:         version,
		WorkflowID:      fmt.Sprintf("fetch-spec-%s-%d", sourceType, time.Now().Unix()),
		DestinationType: destinationType,
	}

	workflowOptions := client.StartWorkflowOptions{
		ID:        params.WorkflowID,
		TaskQueue: t.taskQueue,
	}

	run, err := t.Client.ExecuteWorkflow(ctx, workflowOptions, FetchSpecWorkflow, params)
	if err != nil {
		return dto.SpecOutput{}, fmt.Errorf("failed to execute fetch spec workflow: %s", err)
	}

	var result dto.SpecOutput
	if err := run.Get(ctx, &result); err != nil {
		return dto.SpecOutput{}, fmt.Errorf("workflow execution failed: %s", err)
	}

	return result, nil
}

// TestConnection runs a workflow to test connection
func (t *Temporal) TestConnection(ctx context.Context, workflowID, flag, sourceType, version, config string) (map[string]interface{}, error) {
	params := &ActivityParams{
		SourceType: sourceType,
		Version:    version,
		Config:     config,
		WorkflowID: workflowID,
		Command:    docker.Check,
		Flag:       flag,
	}

	workflowOptions := client.StartWorkflowOptions{
		ID:        params.WorkflowID,
		TaskQueue: t.taskQueue,
	}

	run, err := t.Client.ExecuteWorkflow(ctx, workflowOptions, TestConnectionWorkflow, params)
	if err != nil {
		return nil, fmt.Errorf("failed to execute test connection workflow: %s", err)
	}

	var result map[string]interface{}
	if err := run.Get(ctx, &result); err != nil {
		return nil, fmt.Errorf("workflow execution failed: %s", err)
	}

	return result, nil
}
