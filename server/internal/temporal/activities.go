package temporal

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/datazip/olake-server/internal/docker"
	"go.temporal.io/sdk/activity"
)

// ExecuteDockerCommandActivity executes any Docker command using the refactored Docker runner
func ExecuteDockerCommandActivity(ctx context.Context, params ActivityParams) (map[string]interface{}, error) {
	// Get activity logger
	logger := activity.GetLogger(ctx)
	logger.Info("Starting Docker command activity",
		"sourceType", params.SourceType,
		"command", params.Command)

	// Create a Docker runner with the default config directory
	runner := docker.NewRunner(docker.GetDefaultConfigDir())

	// Write config to file
	configPath, err := runner.WriteToFile(params.Config, params.SourceID)
	if err != nil {
		return nil, fmt.Errorf("failed to write config to file: %v", err)
	}

	// Record heartbeat
	activity.RecordHeartbeat(ctx, "Running docker command")

	// Execute Docker command
	_, err = runner.ExecuteDockerCommand(params.Command, params.SourceType, params.Version, configPath)
	if err != nil {
		logger.Error("Docker command failed", "error", err)
		return nil, fmt.Errorf("docker command failed: %v", err)
	}
	if params.Command == docker.Check && err == nil {
		return map[string]interface{}{"status": "success"}, nil
	}

	// For commands that produce catalog.json, parse and return the result
	if configPath == "" {
		return nil, fmt.Errorf("configPath is empty")
	}
	filePath := filepath.Join(filepath.Dir(configPath), string(params.Command)+".json")

	// Record heartbeat
	activity.RecordHeartbeat(ctx, "Parsing results")

	// Parse and return the result
	result, err := runner.ParseJSONFile(filePath)
	if err != nil {
		// Try to find alternative output files for different commands
		outputDir := filepath.Dir(configPath)
		fileData, findErr := runner.FindAlternativeJSONFile(outputDir, filePath, configPath)
		if findErr != nil || fileData == nil {
			return nil, fmt.Errorf("failed to read output file: %v", err)
		}

		// Re-attempt to parse the result
		result, err = runner.ParseJSONFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to parse JSON from file: %v", err)
		}
	}

	return result, nil
}

// DiscoverCatalogActivity runs the discover command to get catalog data
func DiscoverCatalogActivity(ctx context.Context, params ActivityParams) (map[string]interface{}, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting sync activity",
		"sourceType", params.SourceType,
		"workflowID", params.WorkflowID)

	// Create a Docker runner with the default config directory
	runner := docker.NewRunner(docker.GetDefaultConfigDir())

	// Record heartbeat
	activity.RecordHeartbeat(ctx, "Running sync command")

	// Execute the sync operation
	result, err := runner.GetCatalog(
		params.SourceType,
		params.Version,
		params.Config,
		params.WorkflowID,
	)
	if err != nil {
		logger.Error("Sync command failed", "error", err)
		return result, fmt.Errorf("sync command failed: %v", err)
	}

	return result, nil

}

// GetSpecActivity runs the spec command to get connector specifications
func GetSpecActivity(ctx context.Context, params ActivityParams) (map[string]interface{}, error) {
	params.Command = docker.Spec
	return ExecuteDockerCommandActivity(ctx, params)
}

// TestConnectionActivity runs the check command to test connection
func TestConnectionActivity(ctx context.Context, params ActivityParams) (map[string]interface{}, error) {
	params.Command = docker.Check
	return ExecuteDockerCommandActivity(ctx, params)
}

// SyncActivity runs the sync command to transfer data between source and destination
func SyncActivity(ctx context.Context, params SyncParams) (map[string]interface{}, error) {
	// Get activity logger
	logger := activity.GetLogger(ctx)
	logger.Info("Starting sync activity",
		"sourceType", params.SourceType,
		"sourceID", params.SourceID,
		"destID", params.DestID)

	// Create a Docker runner with the default config directory
	runner := docker.NewRunner(docker.GetDefaultConfigDir())

	// Record heartbeat
	activity.RecordHeartbeat(ctx, "Running sync command")

	// Execute the sync operation
	result, err := runner.RunSync(
		params.SourceType,
		params.Version,
		params.SourceConfig,
		params.DestConfig,
		params.StreamsConfig,
		params.JobId,
		params.ProjectID,
		params.SourceID,
		params.DestID,
		params.WorkflowID,
	)
	if err != nil {
		logger.Error("Sync command failed", "error", err)
		return result, fmt.Errorf("sync command failed: %v", err)
	}

	return result, nil
}
