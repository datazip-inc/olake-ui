package handlers

import (
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

	jobs, err := h.etl.GetAllJobs(h.Ctx.Request.Context(), projectID)
	if err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusInternalServerError, fmt.Sprintf("failed to retrieve jobs by project ID: %s", err), err)
		return
	}
	utils.SuccessResponse(&h.Controller, jobs)
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

	// Conditional validation: if id present, we only require id; otherwise, require name/type/version/config.
	if req.Source.ID == nil {
		if err := dto.ValidateSourceType(req.Source.Type); err != nil {
			utils.ErrorResponse(&h.Controller, http.StatusBadRequest, fmt.Sprintf("failed to validate request: %s", err), err)
			return
		}
		if req.Source.Name == "" || req.Source.Version == "" || req.Source.Config == "" {
			utils.ErrorResponse(&h.Controller, http.StatusBadRequest, fmt.Sprintf("source name, version, and config are required when source id is not provided: %s", err), err)
			return
		}
	}

	if req.Destination.ID == nil {
		if err := dto.ValidateDestinationType(req.Destination.Type); err != nil {
			utils.ErrorResponse(&h.Controller, http.StatusBadRequest, fmt.Sprintf("failed to validate request: %s", err), err)
			return
		}
		if req.Destination.Name == "" || req.Destination.Version == "" || req.Destination.Config == "" {
			utils.ErrorResponse(&h.Controller, http.StatusBadRequest, fmt.Sprintf("destination name, version, and config are required when destination id is not provided: %s", err), err)
			return
		}
	}

	logger.Debugf("Create job initiated project_id[%s] job_name[%s] user_id[%v]",
		projectID, req.Name, userID)

	if err := h.etl.CreateJob(h.Ctx.Request.Context(), &req, projectID, userID); err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusInternalServerError, fmt.Sprintf("failed to create job: %s", err), err)
		return
	}
	utils.SuccessResponse(&h.Controller, req.Name)
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

	// Conditional validation: if id present, we only require id; otherwise, require name/type/version/config.
	if req.Source.ID == nil {
		if err := dto.ValidateSourceType(req.Source.Type); err != nil {
			utils.ErrorResponse(&h.Controller, http.StatusBadRequest, fmt.Sprintf("failed to validate request: %s", err), err)
			return
		}
		if req.Source.Name == "" || req.Source.Version == "" || req.Source.Config == "" {
			utils.ErrorResponse(&h.Controller, http.StatusBadRequest, fmt.Sprintf("source name, version, and config are required when source id is not provided: %s", err), err)
			return
		}
	}

	if req.Destination.ID == nil {
		if err := dto.ValidateDestinationType(req.Destination.Type); err != nil {
			utils.ErrorResponse(&h.Controller, http.StatusBadRequest, fmt.Sprintf("failed to validate request: %s", err), err)
			return
		}
		if req.Destination.Name == "" || req.Destination.Version == "" || req.Destination.Config == "" {
			utils.ErrorResponse(&h.Controller, http.StatusBadRequest, fmt.Sprintf("destination name, version, and config are required when destination id is not provided: %s", err), err)
			return
		}
	}

	logger.Debugf("Update job initiated project_id[%s] job_id[%d] job_name[%s] user_id[%v]",
		projectID, jobID, req.Name, userID)

	if err := h.etl.UpdateJob(h.Ctx.Request.Context(), &req, projectID, jobID, userID); err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusInternalServerError, fmt.Sprintf("failed to update job: %s", err), err)
		return
	}
	utils.SuccessResponse(&h.Controller, req.Name)
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
	utils.SuccessResponse(&h.Controller, jobName)
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

	unique, err := h.etl.IsJobNameUnique(h.Ctx.Request.Context(), projectID, req)
	if err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusInternalServerError, fmt.Sprintf("failed to check job name uniqueness: %s", err), err)
		return
	}
	utils.SuccessResponse(&h.Controller, dto.CheckUniqueJobNameResponse{Unique: unique})
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
	utils.SuccessResponse(&h.Controller, result)
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
		statusCode := http.StatusInternalServerError
		if err.Error() == "job not found" {
			statusCode = http.StatusNotFound
		}
		utils.ErrorResponse(&h.Controller, statusCode, fmt.Sprintf("failed to activate job: %s", err), err)
		return
	}
	utils.SuccessResponse(&h.Controller, nil)
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

	resp, err := h.etl.CancelJobRun(h.Ctx.Request.Context(), projectID, id)
	if err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusInternalServerError, fmt.Sprintf("failed to cancel job run: %s", err), err)
		return
	}
	utils.SuccessResponse(&h.Controller, resp)
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
		statusCode := http.StatusInternalServerError
		if err.Error() == "job not found" {
			statusCode = http.StatusNotFound
		}
		utils.ErrorResponse(&h.Controller, statusCode, fmt.Sprintf("failed to get job tasks: %s", err), err)
		return
	}
	utils.SuccessResponse(&h.Controller, tasks)
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
		statusCode := http.StatusInternalServerError
		if err.Error() == "job not found" {
			statusCode = http.StatusNotFound
		}
		utils.ErrorResponse(&h.Controller, statusCode, fmt.Sprintf("failed to get task logs: %s", err), err)
		return
	}
	utils.SuccessResponse(&h.Controller, logs)
}
