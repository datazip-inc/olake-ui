package temporal

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/datazip/olake-frontend/server/internal/constants"
	"github.com/datazip/olake-frontend/server/internal/database"
	"github.com/datazip/olake-frontend/server/internal/docker"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// Retry policy constants
var (
	// DefaultRetryPolicy is used for standard operations like discovery and testing connections
	DefaultRetryPolicy = &temporal.RetryPolicy{
		InitialInterval:    time.Second * 15,
		BackoffCoefficient: 2.0,
		MaximumInterval:    time.Minute * 10,
		MaximumAttempts:    1,
	}
)

// DiscoverCatalogWorkflow is a workflow for discovering catalogs
func DiscoverCatalogWorkflow(ctx workflow.Context, params *ActivityParams) (map[string]interface{}, error) {
	// Execute the DiscoverCatalogActivity directly
	options := workflow.ActivityOptions{
		StartToCloseTimeout: time.Minute * 10,
		RetryPolicy:         DefaultRetryPolicy,
	}
	ctx = workflow.WithActivityOptions(ctx, options)

	var result map[string]interface{}
	err := workflow.ExecuteActivity(ctx, DiscoverCatalogActivity, params).Get(ctx, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// TestConnectionWorkflow is a workflow for testing connections
func TestConnectionWorkflow(ctx workflow.Context, params *ActivityParams) (map[string]interface{}, error) {
	// Execute the TestConnectionActivity directly
	options := workflow.ActivityOptions{
		StartToCloseTimeout: time.Minute * 10,
		RetryPolicy:         DefaultRetryPolicy,
	}
	ctx = workflow.WithActivityOptions(ctx, options)

	var result map[string]interface{}
	err := workflow.ExecuteActivity(ctx, TestConnectionActivity, params).Get(ctx, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// RunSyncWorkflow is a workflow for running data synchronization
func RunSyncWorkflow(ctx workflow.Context, jobID int) (map[string]interface{}, error) {
	options := workflow.ActivityOptions{
		// Using large duration (e.g., 10 years)
		StartToCloseTimeout: time.Hour * 24 * 30, // 30 days
		RetryPolicy:         DefaultRetryPolicy,
	}
	params := SyncParams{
		JobID:      jobID,
		WorkflowID: workflow.GetInfo(ctx).WorkflowExecution.ID,
	}

	// Get job details from database
	jobORM := database.NewJobORM()
	job, err := jobORM.GetByID(jobID)
	if err != nil {
		workflow.GetLogger(ctx).Error("Failed to get job details", "error", err)
	} else if job != nil {
		params.JobName = job.Name
		params.CreatedAt = job.CreatedAt.Format(time.RFC3339)
		if job.CreatedBy != nil {
			userORM := database.NewUserORM()
			if fullUser, err := userORM.GetByID(job.CreatedBy.ID); err == nil {
				params.CreatedBy = fullUser.Username
			}
		}
		if job.SourceID != nil {
			params.SourceType = job.SourceID.Type
			params.SourceName = job.SourceID.Name
		}
		if job.DestID != nil {
			params.DestinationType = job.DestID.DestType
			params.DestinationName = job.DestID.Name
		}
	}

	ctx = workflow.WithActivityOptions(ctx, options)
	var result map[string]interface{}
	err = workflow.ExecuteActivity(ctx, SyncActivity, params).Get(ctx, &result)
	if err != nil {
		// Track sync failure event
		properties := map[string]interface{}{
			"job_id":           jobID,
			"workflow_id":      params.WorkflowID,
			"ended_at":         time.Now().UTC().Format(time.RFC3339),
			"job_name":         params.JobName,
			"created_by":       params.CreatedBy,
			"created_at":       params.CreatedAt,
			"source_type":      params.SourceType,
			"source_name":      params.SourceName,
			"destination_type": params.DestinationType,
			"destination_name": params.DestinationName,
		}
		if err := workflow.ExecuteActivity(ctx, TrackEventActivity, constants.EventSyncFailed, properties).Get(ctx, nil); err != nil {
			workflow.GetLogger(ctx).Error("Failed to track sync failure event", "error", err)
		}
		return nil, err
	}

	// Track sync completion event
	properties := map[string]interface{}{
		"job_id":           jobID,
		"workflow_id":      params.WorkflowID,
		"ended_at":         time.Now().UTC().Format(time.RFC3339),
		"job_name":         params.JobName,
		"created_by":       params.CreatedBy,
		"created_at":       params.CreatedAt,
		"source_type":      params.SourceType,
		"source_name":      params.SourceName,
		"destination_type": params.DestinationType,
		"destination_name": params.DestinationName,
	}

	// Read stats.json file
	syncFolderName := fmt.Sprintf("%x", sha256.Sum256([]byte(params.WorkflowID)))
	homeDir := docker.GetDefaultConfigDir()
	mainSyncDir := filepath.Join(homeDir, syncFolderName)
	statsPath := filepath.Join(mainSyncDir, "stats.json")

	if statsData, err := os.ReadFile(statsPath); err == nil {
		var stats map[string]interface{}
		if err := json.Unmarshal(statsData, &stats); err == nil {
			// Add stats properties to the event
			if recordsSynced, ok := stats["Synced Records"]; ok {
				properties["records_synced"] = recordsSynced
			}
			if memory, ok := stats["Memory"]; ok {
				properties["memory_used"] = memory
			}
		}
		if err != nil {
			fmt.Println("Error unmarshalling stats.json", err)
		}
	}

	// Read streams.json if exists
	streamsPath := filepath.Join(mainSyncDir, "streams.json")
	if streamsData, err := os.ReadFile(streamsPath); err == nil {
		var streamsConfig struct {
			Streams []struct {
				Stream struct {
					Name               string   `json:"name"`
					Namespace          string   `json:"namespace"`
					SyncMode           string   `json:"sync_mode"`
					SupportedSyncModes []string `json:"supported_sync_modes"`
				} `json:"stream"`
			} `json:"streams"`
			SelectedStreams map[string][]struct {
				StreamName     string `json:"stream_name"`
				Normalization  bool   `json:"normalization"`
				PartitionRegex string `json:"partition_regex"`
			} `json:"selected_streams"`
		}

		if err := json.Unmarshal(streamsData, &streamsConfig); err == nil {
			// Count normalized streams
			normalizedCount := 0
			partitionedCount := 0

			// Count normalized and partitioned streams from selected_streams
			for _, streams := range streamsConfig.SelectedStreams {
				for _, stream := range streams {
					if stream.Normalization {
						normalizedCount++
					}
					if stream.PartitionRegex != "" {
						partitionedCount++
					}
				}
			}

			properties["normalized_streams_count"] = normalizedCount
			properties["partitioned_streams_count"] = partitionedCount
		} else {
			fmt.Println("Error unmarshalling streams.json", err)
		}
	}

	if err := workflow.ExecuteActivity(ctx, TrackEventActivity, constants.EventSyncCompleted, properties).Get(ctx, nil); err != nil {
		workflow.GetLogger(ctx).Error("Failed to track sync completion event", "error", err)
	}
	return result, nil
}
