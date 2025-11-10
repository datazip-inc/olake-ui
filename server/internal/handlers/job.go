package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/datazip-inc/olake-ui/server/internal/models/dto"
	"github.com/datazip-inc/olake-ui/server/utils"
	"github.com/datazip-inc/olake-ui/server/utils/logger"
)

// @router /project/:projectid/jobs [get]
func (h *Handler) ListJobs() {
	projectID, err := GetProjectIDFromPath(&h.Controller)
	if err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusBadRequest, fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}

	logger.Debugf("Get all jobs initiated project_id[%s]", projectID)

	jobs, err := h.etl.ListJobs(h.Ctx.Request.Context(), projectID)
	if err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusInternalServerError, fmt.Sprintf("failed to retrieve jobs by project ID: %s", err), err)
		return
	}
	utils.SuccessResponse(&h.Controller, "jobs listed successfully", jobs)
}

// @router /project/:projectid/jobs [post]
func (h *Handler) CreateJob() {
	userID := GetUserIDFromSession(&h.Controller)
	if userID == nil {
		utils.ErrorResponse(&h.Controller, http.StatusUnauthorized, "Not authenticated", fmt.Errorf("not authenticated"))
		return
	}

	projectID, err := GetProjectIDFromPath(&h.Controller)
	if err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusBadRequest, fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}

	var req dto.CreateJobRequest
	if err := UnmarshalAndValidate(h.Ctx.Input.RequestBody, &req); err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusBadRequest, fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}

	// Conditional validation
	if req.Source.ID == nil {
		if err := dto.ValidateSourceType(req.Source.Type); err != nil {
			utils.ErrorResponse(&h.Controller, http.StatusBadRequest, fmt.Sprintf("failed to validate request: %s", err), err)
			return
		}
		if req.Source.Name == "" || req.Source.Version == "" || req.Source.Config == "" {
			utils.ErrorResponse(&h.Controller, http.StatusBadRequest, "source name, version, and config are required when source id is not provided", err)
			return
		}
	}

	if req.Destination.ID == nil {
		if err := dto.ValidateDestinationType(req.Destination.Type); err != nil {
			utils.ErrorResponse(&h.Controller, http.StatusBadRequest, fmt.Sprintf("failed to validate request: %s", err), err)
			return
		}
		if req.Destination.Name == "" || req.Destination.Version == "" || req.Destination.Config == "" {
			utils.ErrorResponse(&h.Controller, http.StatusBadRequest, "destination name, version, and config are required when destination id is not provided", err)
			return
		}
	}

	logger.Debugf("Create job initiated project_id[%s] job_name[%s] user_id[%v]", projectID, req.Name, userID)

	if err := h.etl.CreateJob(h.Ctx.Request.Context(), &req, projectID, userID); err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusInternalServerError, fmt.Sprintf("failed to create job: %s", err), err)
		return
	}
	utils.SuccessResponse(&h.Controller, fmt.Sprintf("job '%s' created successfully", req.Name), nil)
}

// @router /project/:projectid/jobs/:id [put]
func (h *Handler) UpdateJob() {
	userID := GetUserIDFromSession(&h.Controller)
	if userID == nil {
		utils.ErrorResponse(&h.Controller, http.StatusUnauthorized, "Not authenticated", fmt.Errorf("not authenticated"))
		return
	}

	projectID, err := GetProjectIDFromPath(&h.Controller)
	if err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusBadRequest, fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}

	jobID, err := GetIDFromPath(&h.Controller)
	if err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusBadRequest, fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}

	var req dto.UpdateJobRequest
	if err := UnmarshalAndValidate(h.Ctx.Input.RequestBody, &req); err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusBadRequest, fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}

	if req.Source.ID == nil {
		if err := dto.ValidateSourceType(req.Source.Type); err != nil {
			utils.ErrorResponse(&h.Controller, http.StatusBadRequest, fmt.Sprintf("failed to validate request: %s", err), err)
			return
		}
		if req.Source.Name == "" || req.Source.Version == "" || req.Source.Config == "" {
			utils.ErrorResponse(&h.Controller, http.StatusBadRequest, "source name, version, and config are required when source id is not provided", err)
			return
		}
	}
	if req.Destination.ID == nil {
		if err := dto.ValidateDestinationType(req.Destination.Type); err != nil {
			utils.ErrorResponse(&h.Controller, http.StatusBadRequest, fmt.Sprintf("failed to validate request: %s", err), err)
			return
		}
		if req.Destination.Name == "" || req.Destination.Version == "" || req.Destination.Config == "" {
			utils.ErrorResponse(&h.Controller, http.StatusBadRequest, "destination name, version, and config are required when destination id is not provided", err)
			return
		}
	}

	logger.Debugf("Update job initiated project_id[%s] job_id[%d] job_name[%s] user_id[%v]", projectID, jobID, req.Name, userID)

	if err := h.etl.UpdateJob(h.Ctx.Request.Context(), &req, projectID, jobID, userID); err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusInternalServerError, fmt.Sprintf("failed to update job: %s", err), err)
		return
	}
	utils.SuccessResponse(&h.Controller, fmt.Sprintf("job '%s' updated successfully", req.Name), nil)
}

// @router /project/:projectid/jobs/:id [delete]
func (h *Handler) DeleteJob() {
	id, err := GetIDFromPath(&h.Controller)
	if err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusBadRequest, fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}

	logger.Infof("Delete job initiated job_id[%d]", id)

	jobName, err := h.etl.DeleteJob(h.Ctx.Request.Context(), id)
	if err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusInternalServerError, fmt.Sprintf("failed to delete job: %s", err), err)
		return
	}
	utils.SuccessResponse(&h.Controller, fmt.Sprintf("job '%s' deleted successfully", jobName), nil)
}

// @router /project/:projectid/jobs/check-unique [post]
func (h *Handler) CheckUniqueJobName() {
	projectID, err := GetProjectIDFromPath(&h.Controller)
	if err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusBadRequest, fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}

	var req dto.CheckUniqueJobNameRequest
	if err := UnmarshalAndValidate(h.Ctx.Input.RequestBody, &req); err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusBadRequest, fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}

	logger.Infof("Check unique job name initiated project_id[%s] job_name[%s]", projectID, req.JobName)

	unique, err := h.etl.CheckUniqueJobName(h.Ctx.Request.Context(), projectID, req)
	if err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusInternalServerError, fmt.Sprintf("failed to check job name uniqueness: %s", err), err)
		return
	}
	utils.SuccessResponse(&h.Controller, fmt.Sprintf("job name '%s' uniqueness checked successfully", req.JobName), dto.CheckUniqueJobNameResponse{Unique: unique})
}

// @router /project/:projectid/jobs/:id/sync [post]
func (h *Handler) SyncJob() {
	projectID, err := GetProjectIDFromPath(&h.Controller)
	if err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusBadRequest, fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}

	id, err := GetIDFromPath(&h.Controller)
	if err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusBadRequest, fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}

	logger.Infof("Sync trigger initiated for project_id[%s] job_id[%d]", projectID, id)

	result, err := h.etl.SyncJob(h.Ctx.Request.Context(), projectID, id)
	if err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusInternalServerError, fmt.Sprintf("failed to trigger sync: %s", err), err)
		return
	}
	utils.SuccessResponse(&h.Controller, fmt.Sprintf("sync triggered successfully for job_id[%d]", id), result)
}

// @router /project/:projectid/jobs/:id/activate [put]
func (h *Handler) ActivateJob() {
	userID := GetUserIDFromSession(&h.Controller)
	if userID == nil {
		utils.ErrorResponse(&h.Controller, http.StatusUnauthorized, "Not authenticated", fmt.Errorf("not authenticated"))
		return
	}

	id, err := GetIDFromPath(&h.Controller)
	if err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusBadRequest, fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}

	var req dto.JobStatusRequest
	if err := UnmarshalAndValidate(h.Ctx.Input.RequestBody, &req); err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusBadRequest, fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}

	logger.Debugf("Activate job initiated job_id[%d] user_id[%v]", id, userID)

	if err := h.etl.ActivateJob(h.Ctx.Request.Context(), id, req, userID); err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusInternalServerError, fmt.Sprintf("failed to activate job: %s", err), err)
		return
	}
	utils.SuccessResponse(&h.Controller, fmt.Sprintf("job %d %s successfully", id, utils.Ternary(req.Activate, "resumed", "paused")), nil)
}

// @router /project/:projectid/jobs/:id/cancel [post]
func (h *Handler) CancelJobRun() {
	projectID, err := GetProjectIDFromPath(&h.Controller)
	if err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusBadRequest, fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}

	id, err := GetIDFromPath(&h.Controller)
	if err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusBadRequest, fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}

	logger.Infof("Cancel job run initiated project_id[%s] job_id[%d]", projectID, id)

	if err := h.etl.CancelJobRun(h.Ctx.Request.Context(), projectID, id); err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusInternalServerError, fmt.Sprintf("failed to cancel job run: %s", err), err)
		return
	}
	utils.SuccessResponse(&h.Controller, fmt.Sprintf("job workflow cancel requested successfully for job_id[%d]", id), nil)
}

// @router /project/:projectid/jobs/:id/clear-destination [post]
func (h *Handler) ClearDestination() {
	projectID, err := GetProjectIDFromPath(&h.Controller)
	if err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusBadRequest, fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}

	id, err := GetIDFromPath(&h.Controller)
	if err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusBadRequest, fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}
	if err := h.etl.ClearDestination(h.Ctx.Request.Context(), projectID, id, "", 0); err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusInternalServerError, fmt.Sprintf("failed to trigger clear destination: %s", err), err)
		return
	}
	utils.SuccessResponse(&h.Controller, fmt.Sprintf("clear destination triggered successfully for job_id[%d]", id), nil)
}

// @router /project/:projectid/jobs/:id/stream-difference [post]
func (h *Handler) GetStreamDifference() {
	projectID, err := GetProjectIDFromPath(&h.Controller)
	if err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusBadRequest, fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}

	id, err := GetIDFromPath(&h.Controller)
	if err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusBadRequest, fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}

	var req dto.StreamDifferenceRequest
	if err := UnmarshalAndValidate(h.Ctx.Input.RequestBody, &req); err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusBadRequest, fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}

	logger.Debugf("Get stream difference initiated project_id[%s] job_id[%d]", projectID, id)

	diffStreams, err := h.etl.GetStreamDifference(h.Ctx.Request.Context(), projectID, id, req)
	if err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusInternalServerError, "Failed to get stream difference", err)
		return
	}
	utils.SuccessResponse(&h.Controller, fmt.Sprintf("stream difference retrieved successfully for job_id[%d]", id), dto.StreamDifferenceResponse{
		DifferenceStreams: diffStreams,
	})
}

// @router /project/:projectid/jobs/:id/clear-destination [get]
func (h *Handler) GetClearDestinationStatus() {
	projectID, err := GetProjectIDFromPath(&h.Controller)
	if err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusBadRequest, fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}

	jobID, err := GetIDFromPath(&h.Controller)
	if err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusBadRequest, fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}

	status, err := h.etl.GetClearDestinationStatus(h.Ctx.Request.Context(), projectID, jobID)
	if err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusInternalServerError, fmt.Sprintf("failed to get clear destination status: %s", err), err)
		return
	}
	utils.SuccessResponse(&h.Controller, fmt.Sprintf("clear destination status retrieved successfully for job_id[%d]", jobID), dto.ClearDestinationStatusResponse{
		Running: status,
	})
}

// @router /project/:projectid/jobs/:id/tasks [get]
func (h *Handler) GetJobTasks() {
	projectID, err := GetProjectIDFromPath(&h.Controller)
	if err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusBadRequest, fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}

	id, err := GetIDFromPath(&h.Controller)
	if err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusBadRequest, fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}

	logger.Debugf("Get job tasks initiated project_id[%s] job_id[%d]", projectID, id)

	tasks, err := h.etl.GetJobTasks(h.Ctx.Request.Context(), projectID, id)
	if err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusInternalServerError, fmt.Sprintf("failed to get job tasks: %s", err), err)
		return
	}
	utils.SuccessResponse(&h.Controller, fmt.Sprintf("job tasks listed successfully for job_id[%d]", id), tasks)
}

// @router /project/:projectid/jobs/:id/logs [get]
func (h *Handler) GetTaskLogs() {
	id, err := GetIDFromPath(&h.Controller)
	if err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusBadRequest, fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}

	var req dto.JobTaskRequest
	if err := UnmarshalAndValidate(h.Ctx.Input.RequestBody, &req); err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusBadRequest, fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}

	logger.Debugf("Get task logs initiated job_id[%d] file_path[%s]", id, req.FilePath)

	logs, err := h.etl.GetTaskLogs(h.Ctx.Request.Context(), id, req.FilePath)
	if err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusInternalServerError, fmt.Sprintf("failed to get task logs: %s", err), err)
		return
	}
	utils.SuccessResponse(&h.Controller, fmt.Sprintf("task logs retrieved successfully for job_id[%d]", id), logs)
}

// @router /internal/worker/callback/sync-telemetry [post]
func (h *Handler) UpdateSyncTelemetry() {
	var req struct {
		JobID      int    `json:"job_id"`
		WorkflowID string `json:"workflow_id"`
		Event      string `json:"event"`
	}

	if err := json.Unmarshal(h.Ctx.Input.RequestBody, &req); err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusBadRequest, "Invalid request format", err)
		return
	}

	if req.JobID == 0 || req.WorkflowID == "" {
		utils.ErrorResponse(&h.Controller, http.StatusBadRequest, "job_id and workflow_id are required", nil)
		return
	}

	if err := h.etl.UpdateSyncTelemetry(h.Ctx.Request.Context(), req.JobID, req.WorkflowID, req.Event); err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusInternalServerError, "Failed to update sync telemetry", err)
		return
	}

	utils.SuccessResponse(&h.Controller, fmt.Sprintf("sync telemetry updated successfully for job_id[%d] workflow_id[%s] event[%s]", req.JobID, req.WorkflowID, req.Event), nil)
}

// RecoverClearDestination handles recovery from stuck clear-destination workflows (internal use only)
// @router /projects/:project_id/jobs/:job_id/recover-clear-destination [post]
func (h *Handler) RecoverClearDestination() {
	projectID, err := GetProjectIDFromPath(&h.Controller)
	if err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusBadRequest, fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}

	jobID, err := GetIDFromPath(&h.Controller)
	if err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusBadRequest, fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}

	if err := h.etl.RecoverFromClearDestination(h.Ctx.Request.Context(), projectID, jobID); err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusInternalServerError, fmt.Sprintf("failed to recover from clear-destination: %s", err), err)
		return
	}

	utils.SuccessResponse(&h.Controller, fmt.Sprintf("successfully recovered from clear-destination and restored sync schedule for job_id[%d]", jobID), nil)
}
