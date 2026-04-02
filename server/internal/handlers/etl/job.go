package etl

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/datazip-inc/olake-ui/server/internal/constants"
	"github.com/datazip-inc/olake-ui/server/internal/models/dto"
	"github.com/datazip-inc/olake-ui/server/internal/utils"
	"github.com/datazip-inc/olake-ui/server/internal/utils/logger"
)

// @Summary List all jobs
// @Tags Jobs
// @Description Retrieve a list of all jobs associated with a specific project.
// @Param   projectid     path    string  true    "project id (default is 123)"
// @Success 200 {object} dto.JSONResponse{data=[]dto.JobResponse}
// @Failure 400 {object} dto.Error400Response "failed to validate request"
// @Failure 401 {object} dto.Error401Response "unauthorized"
// @Failure 500 {object} dto.Error500Response "failed to retrieve jobs"
// @Router /api/v1/project/{projectid}/jobs [get]
func (h *Handler) ListJobs(c *gin.Context) {
	projectID, err := utils.GetProjectID(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}
	logger.Debugf("List jobs initiated project_id[%s]", projectID)
	jobs, err := h.etl.ListJobs(c.Request.Context(), projectID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, fmt.Sprintf("failed to retrieve jobs by project ID: %s", err), err)
		return
	}
	utils.SuccessResponse(c, "jobs listed successfully", jobs)
}

// @Summary Get job details
// @Tags Jobs
// @Description Retrieve details of a specific job identified by its unique ID.
// @Param   projectid     path    string  true    "project id (default is 123)"
// @Param   id            path    int     true    "job id"
// @Success 200 {object} dto.JSONResponse{data=dto.JobResponse}
// @Failure 400 {object} dto.Error400Response "failed to validate request"
// @Failure 401 {object} dto.Error401Response "unauthorized"
// @Failure 404 {object} dto.Error404Response "job not found"
// @Failure 500 {object} dto.Error500Response "failed to get job"
// @Router /api/v1/project/{projectid}/jobs/{id} [get]
func (h *Handler) GetJob(c *gin.Context) {
	projectID, err := utils.GetProjectID(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}
	jobID, err := utils.GetIDParam(c, "id")
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}
	logger.Debugf("Get job initiated project_id[%s] job_id[%d]", projectID, jobID)
	job, err := h.etl.GetJob(c.Request.Context(), projectID, jobID)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, constants.ErrJobNotFound) {
			status = http.StatusNotFound
		}
		utils.ErrorResponse(c, status, fmt.Sprintf("failed to get job: %s", err), err)
		return
	}
	utils.SuccessResponse(c, fmt.Sprintf("job '%d' retrieved successfully", jobID), job)
}

// @Summary Create a new job
// @Tags Jobs
// @Description Create a new job within a specific project.
// @Param   projectid     path    string  true    "project id (default is 123)"
// @Param   body          body    dto.CreateJobRequest true "job data"
// @Success 200 {object} dto.JSONResponse "job created successfully"
// @Failure 400 {object} dto.Error400Response "failed to validate request"
// @Failure 401 {object} dto.Error401Response "unauthorized"
// @Failure 413 {object} dto.Error413Response "payload too large"
// @Failure 500 {object} dto.Error500Response "failed to create job"
// @Router /api/v1/project/{projectid}/jobs [post]
func (h *Handler) CreateJob(c *gin.Context) {
	userID := utils.GetCurrentUserID(c)
	if userID == nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Not authenticated", fmt.Errorf("not authenticated"))
		return
	}
	projectID, err := utils.GetProjectID(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}
	var req dto.CreateJobRequest
	if err := utils.BindAndValidate(c, &req); err != nil {
		utils.ErrorResponse(c, utils.StatusFromBindError(err), fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}
	if err := validateJobDriverConfig(req.Source, req.Destination); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}
	logger.Debugf("Create job initiated project_id[%s] user_id[%v] job_name[%s]", projectID, userID, req.Name)
	if err := h.etl.CreateJob(c.Request.Context(), &req, projectID, userID); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, fmt.Sprintf("failed to create job: %s", err), err)
		return
	}
	utils.SuccessResponse(c, fmt.Sprintf("job '%s' created successfully", req.Name), nil)
}

// @Summary Update a job
// @Tags Jobs
// @Description Update the configuration details of an existing job.
// @Param   projectid     path    string  true    "project id (default is 123)"
// @Param   id            path    int     true    "job id"
// @Param   body          body    dto.UpdateJobRequest true "job data"
// @Success 200 {object} dto.JSONResponse "job updated successfully"
// @Failure 400 {object} dto.Error400Response "failed to validate request"
// @Failure 401 {object} dto.Error401Response "unauthorized"
// @Failure 404 {object} dto.Error404Response "job not found"
// @Failure 413 {object} dto.Error413Response "payload too large"
// @Failure 500 {object} dto.Error500Response "failed to update job"
// @Router /api/v1/project/{projectid}/jobs/{id} [put]
func (h *Handler) UpdateJob(c *gin.Context) {
	userID := utils.GetCurrentUserID(c)
	if userID == nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Not authenticated", fmt.Errorf("not authenticated"))
		return
	}
	projectID, err := utils.GetProjectID(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}
	jobID, err := utils.GetIDParam(c, "id")
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}
	var req dto.UpdateJobRequest
	if err := utils.BindAndValidate(c, &req); err != nil {
		utils.ErrorResponse(c, utils.StatusFromBindError(err), fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}
	if err := validateJobDriverConfig(req.Source, req.Destination); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}

	logger.Debugf("Update job initiated project_id[%s] job_id[%d] user_id[%v]", projectID, jobID, userID)
	if err := h.etl.UpdateJob(c.Request.Context(), &req, projectID, jobID, userID); err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, constants.ErrJobNotFound) {
			status = http.StatusNotFound
		}
		utils.ErrorResponse(c, status, fmt.Sprintf("failed to update job: %s", err), err)
		return
	}
	utils.SuccessResponse(c, fmt.Sprintf("job '%s' updated successfully", req.Name), nil)
}

// @Summary Delete a job
// @Tags Jobs
// @Description Permanently delete a specified job.
// @Param   projectid     path    string  true    "project id (default is 123)"
// @Param   id            path    int     true    "job id"
// @Success 200 {object} dto.JSONResponse "job deleted successfully"
// @Failure 400 {object} dto.Error400Response "failed to validate request"
// @Failure 401 {object} dto.Error401Response "unauthorized"
// @Failure 404 {object} dto.Error404Response "job not found"
// @Failure 500 {object} dto.Error500Response "failed to delete job"
// @Router /api/v1/project/{projectid}/jobs/{id} [delete]
func (h *Handler) DeleteJob(c *gin.Context) {
	id, err := utils.GetIDParam(c, "id")
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}
	logger.Debugf("Delete job initiated job_id[%d]", id)
	jobName, err := h.etl.DeleteJob(c.Request.Context(), id)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, constants.ErrJobNotFound) {
			status = http.StatusNotFound
		}
		utils.ErrorResponse(c, status, fmt.Sprintf("failed to delete job: %s", err), err)
		return
	}
	utils.SuccessResponse(c, fmt.Sprintf("job '%s' deleted successfully", jobName), nil)
}

// @Summary Check name uniqueness
// @Tags Jobs
// @Description Verify if a given name is unique within the project for a specific entity type.
// @Param   projectid     path    string  true    "project id (default is 123)"
// @Param   body          body    dto.CheckUniqueNameRequest true "unique check data"
// @Success 200 {object} dto.JSONResponse{data=dto.CheckUniqueJobNameResponse}
// @Failure 400 {object} dto.Error400Response "failed to validate request"
// @Failure 401 {object} dto.Error401Response "unauthorized"
// @Failure 409 {object} dto.Error409Response "name is not unique"
// @Failure 413 {object} dto.Error413Response "payload too large"
// @Failure 500 {object} dto.Error500Response "failed to check uniqueness"
// @Router /api/v1/project/{projectid}/check-unique [post]
func (h *Handler) CheckUniqueName(c *gin.Context) {
	projectID, err := utils.GetProjectID(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}
	var req dto.CheckUniqueNameRequest
	if err := utils.BindAndValidate(c, &req); err != nil {
		utils.ErrorResponse(c, utils.StatusFromBindError(err), fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}
	logger.Debugf("Check unique name initiated project_id[%s] entity_type[%s] name[%s]", projectID, req.EntityType, req.Name)
	unique, err := h.etl.CheckUniqueName(c.Request.Context(), projectID, req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, fmt.Sprintf("failed to check name uniqueness: %s", err), err)
		return
	}
	if !unique {
		utils.ErrorResponse(c, http.StatusConflict, fmt.Sprintf("%s name '%s' is not unique", req.EntityType, req.Name), nil)
		return
	}
	utils.SuccessResponse(c, fmt.Sprintf("%s name '%s' uniqueness checked successfully", req.EntityType, req.Name), dto.CheckUniqueJobNameResponse{Unique: unique})
}

// @Summary Trigger job sync
// @Tags Jobs
// @Description Trigger a manual sync for a job
// @Param   projectid     path    string  true    "project id (default is 123)"
// @Param   id            path    int     true    "job id"
// @Success 200 {object} dto.JSONResponse "sync triggered successfully"
// @Failure 400 {object} dto.Error400Response "failed to validate request"
// @Failure 401 {object} dto.Error401Response "unauthorized"
// @Failure 404 {object} dto.Error404Response "job not found"
// @Failure 500 {object} dto.Error500Response "failed to trigger sync"
// @Router /api/v1/project/{projectid}/jobs/{id}/sync [post]
func (h *Handler) SyncJob(c *gin.Context) {
	projectID, err := utils.GetProjectID(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}
	id, err := utils.GetIDParam(c, "id")
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}
	logger.Debugf("Sync job initiated project_id[%s] job_id[%d]", projectID, id)
	result, err := h.etl.SyncJob(c.Request.Context(), projectID, id)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, constants.ErrJobNotFound) {
			status = http.StatusNotFound
		}
		utils.ErrorResponse(c, status, fmt.Sprintf("failed to trigger sync: %s", err), err)
		return
	}
	utils.SuccessResponse(c, fmt.Sprintf("sync triggered successfully for job_id[%d]", id), result)
}

// @Summary Pause or resume job
// @Tags Jobs
// @Description Pause or resume a job.
// @Param   projectid     path    string  true    "project id (default is 123)"
// @Param   id            path    int     true    "job id"
// @Param   body          body    dto.JobStatusRequest true "activation data"
// @Success 200 {object} dto.JSONResponse "job activated/deactivated successfully"
// @Failure 400 {object} dto.Error400Response "failed to validate request"
// @Failure 401 {object} dto.Error401Response "unauthorized"
// @Failure 404 {object} dto.Error404Response "job not found"
// @Failure 413 {object} dto.Error413Response "payload too large"
// @Failure 500 {object} dto.Error500Response "failed to activate job"
// @Router /api/v1/project/{projectid}/jobs/{id}/activate [post]
func (h *Handler) ActivateJob(c *gin.Context) {
	userID := utils.GetCurrentUserID(c)
	if userID == nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Not authenticated", fmt.Errorf("not authenticated"))
		return
	}
	id, err := utils.GetIDParam(c, "id")
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}
	var req dto.JobStatusRequest
	if err := utils.BindAndValidate(c, &req); err != nil {
		utils.ErrorResponse(c, utils.StatusFromBindError(err), fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}
	logger.Debugf("Activate job initiated job_id[%d] activate[%t] user_id[%v]", id, req.Activate, userID)
	if err := h.etl.ActivateJob(c.Request.Context(), id, req, userID); err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, constants.ErrJobNotFound) {
			status = http.StatusNotFound
		}
		utils.ErrorResponse(c, status, fmt.Sprintf("failed to activate job: %s", err), err)
		return
	}
	action := "paused"
	if req.Activate {
		action = "resumed"
	}
	utils.SuccessResponse(c, fmt.Sprintf("job %d %s successfully", id, action), nil)
}

// @Summary Cancel running job
// @Tags Jobs
// @Description Request cancellation of a currently running job execution.
// @Param   projectid     path    string  true    "project id (default is 123)"
// @Param   id            path    int     true    "job id"
// @Success 200 {object} dto.JSONResponse "job cancel requested successfully"
// @Failure 400 {object} dto.Error400Response "failed to validate request"
// @Failure 401 {object} dto.Error401Response "unauthorized"
// @Failure 404 {object} dto.Error404Response "job not found"
// @Failure 500 {object} dto.Error500Response "failed to cancel job run"
// @Router /api/v1/project/{projectid}/jobs/{id}/cancel [get]
func (h *Handler) CancelJobRun(c *gin.Context) {
	projectID, err := utils.GetProjectID(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}
	id, err := utils.GetIDParam(c, "id")
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}
	logger.Debugf("Cancel job run initiated project_id[%s] job_id[%d]", projectID, id)
	if err := h.etl.CancelJobRun(c.Request.Context(), projectID, id); err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, constants.ErrJobNotFound) {
			status = http.StatusNotFound
		}
		utils.ErrorResponse(c, status, fmt.Sprintf("failed to cancel job run: %s", err), err)
		return
	}
	utils.SuccessResponse(c, fmt.Sprintf("job workflow cancel requested successfully for job_id[%d]", id), nil)
}

// @Summary Clear destination data
// @Tags Jobs
// @Description Initiate job to clear data in the destination associated with a job.
// @Param   projectid     path    string  true    "project id (default is 123)"
// @Param   id            path    int     true    "job id"
// @Success 200 {object} dto.JSONResponse "clear destination triggered successfully"
// @Failure 400 {object} dto.Error400Response "failed to validate request"
// @Failure 401 {object} dto.Error401Response "unauthorized"
// @Failure 404 {object} dto.Error404Response "job not found"
// @Failure 500 {object} dto.Error500Response "failed to trigger clear destination"
// @Router /api/v1/project/{projectid}/jobs/{id}/clear-destination [post]
func (h *Handler) ClearDestination(c *gin.Context) {
	projectID, err := utils.GetProjectID(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}
	id, err := utils.GetIDParam(c, "id")
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}
	logger.Debugf("Clear destination initiated project_id[%s] job_id[%d]", projectID, id)
	if err := h.etl.ClearDestination(c.Request.Context(), projectID, id, "", 0, true); err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, constants.ErrJobNotFound) {
			status = http.StatusNotFound
		}
		utils.ErrorResponse(c, status, fmt.Sprintf("failed to trigger clear destination: %s", err), err)
		return
	}
	utils.SuccessResponse(c, fmt.Sprintf("clear destination triggered successfully for job_id[%d]", id), nil)
}

// @Summary Get stream differences
// @Tags Jobs
// @Description Get difference between current streams.json and existing streams.json.
// @Param   projectid     path    string  true    "project id (default is 123)"
// @Param   id            path    int     true    "job id"
// @Param   body          body    dto.StreamDifferenceRequest true "stream difference data"
// @Success 200 {object} dto.JSONResponse{data=dto.StreamDifferenceResponse}
// @Failure 400 {object} dto.Error400Response "failed to validate request"
// @Failure 401 {object} dto.Error401Response "unauthorized"
// @Failure 404 {object} dto.Error404Response "job not found"
// @Failure 413 {object} dto.Error413Response "payload too large"
// @Failure 500 {object} dto.Error500Response "failed to get stream difference"
// @Router /api/v1/project/{projectid}/jobs/{id}/stream-difference [post]
func (h *Handler) GetStreamDifference(c *gin.Context) {
	projectID, err := utils.GetProjectID(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}
	id, err := utils.GetIDParam(c, "id")
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}
	var req dto.StreamDifferenceRequest
	if err := utils.BindAndValidate(c, &req); err != nil {
		utils.ErrorResponse(c, utils.StatusFromBindError(err), fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}
	logger.Debugf("Get stream difference initiated project_id[%s] job_id[%d]", projectID, id)
	diffStreams, err := h.etl.GetStreamDifference(c.Request.Context(), projectID, id, req)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, constants.ErrJobNotFound) {
			status = http.StatusNotFound
		}
		utils.ErrorResponse(c, status, fmt.Sprintf("failed to get stream difference: %s", err), err)
		return
	}
	utils.SuccessResponse(c, fmt.Sprintf("stream difference retrieved successfully for job_id[%d]", id), dto.StreamDifferenceResponse{
		DifferenceStreams: diffStreams,
	})
}

// @Summary Get clear destination status
// @Tags Jobs
// @Description Retrieve the current status of an ongoing clear destination job.
// @Param   projectid     path    string  true    "project id (default is 123)"
// @Param   id            path    int     true    "job id"
// @Success 200 {object} dto.JSONResponse{data=dto.ClearDestinationStatusResponse}
// @Failure 400 {object} dto.Error400Response "failed to validate request"
// @Failure 401 {object} dto.Error401Response "unauthorized"
// @Failure 404 {object} dto.Error404Response "job not found"
// @Failure 500 {object} dto.Error500Response "failed to get status"
// @Router /api/v1/project/{projectid}/jobs/{id}/clear-destination [get]
func (h *Handler) GetClearDestinationStatus(c *gin.Context) {
	projectID, err := utils.GetProjectID(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}
	jobID, err := utils.GetIDParam(c, "id")
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}
	logger.Debugf("Get clear destination status initiated project_id[%s] job_id[%d]", projectID, jobID)
	status, err := h.etl.GetClearDestinationStatus(c.Request.Context(), projectID, jobID)
	if err != nil {
		httpStatus := http.StatusInternalServerError
		if errors.Is(err, constants.ErrJobNotFound) {
			httpStatus = http.StatusNotFound
		}
		utils.ErrorResponse(c, httpStatus, fmt.Sprintf("failed to get clear destination status: %s", err), err)
		return
	}
	utils.SuccessResponse(c, fmt.Sprintf("clear destination status retrieved successfully for job_id[%d]", jobID), dto.ClearDestinationStatusResponse{Running: status})
}

// @Summary List job tasks
// @Tags Jobs
// @Description Retrieve a list of execution tasks associated with a specific job.
// @Param   projectid     path    string  true    "project id (default is 123)"
// @Param   id            path    int     true    "job id"
// @Success 200 {object} dto.JSONResponse{data=[]dto.JobTask}
// @Failure 400 {object} dto.Error400Response "failed to validate request"
// @Failure 401 {object} dto.Error401Response "unauthorized"
// @Failure 404 {object} dto.Error404Response "job not found"
// @Failure 500 {object} dto.Error500Response "failed to get job tasks"
// @Router /api/v1/project/{projectid}/jobs/{id}/tasks [get]
func (h *Handler) GetJobTasks(c *gin.Context) {
	projectID, err := utils.GetProjectID(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}
	id, err := utils.GetIDParam(c, "id")
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}
	logger.Debugf("Get job tasks initiated project_id[%s] job_id[%d]", projectID, id)
	tasks, err := h.etl.GetJobTasks(c.Request.Context(), projectID, id)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, constants.ErrJobNotFound) {
			status = http.StatusNotFound
		}
		utils.ErrorResponse(c, status, fmt.Sprintf("failed to get job tasks: %s", err), err)
		return
	}
	utils.SuccessResponse(c, fmt.Sprintf("job tasks listed successfully for job_id[%d]", id), tasks)
}

// @Summary Get task logs
// @Tags Jobs
// @Description Retrieves the execution logs for a specific task. The file path for the log must be obtained from the [Get Job Tasks](#/Jobs/get_api_v1_project__projectid__jobs__id__tasks) endpoint.
// @Param   projectid     path    string  true    "project id (default is 123)"
// @Param   id            path    int     true    "job id"
// @Param   taskid        path    string  true    "task id (defaults to 1)"
// @Param   body          body    dto.JobTaskRequest true "task log data"
// @Param   cursor        query   int     false   "log cursor"
// @Param   limit         query   int     false   "log limit"
// @Param   direction     query   string  false   "log direction"
// @Success 200 {object} dto.JSONResponse{data=dto.TaskLogsResponse}
// @Failure 400 {object} dto.Error400Response "failed to validate request"
// @Failure 401 {object} dto.Error401Response "unauthorized"
// @Failure 404 {object} dto.Error404Response "job not found"
// @Failure 413 {object} dto.Error413Response "payload too large"
// @Failure 500 {object} dto.Error500Response "failed to get task logs"
// @Router /api/v1/project/{projectid}/jobs/{id}/tasks/{taskid}/logs [post]
func (h *Handler) GetTaskLogs(c *gin.Context) {
	id, err := utils.GetIDParam(c, "id")
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}
	var req dto.JobTaskRequest
	if err := utils.BindAndValidate(c, &req); err != nil {
		utils.ErrorResponse(c, utils.StatusFromBindError(err), fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}
	logger.Debugf("Get task logs initiated job_id[%d] file_path[%s]", id, req.FilePath)

	cursor := constants.DefaultLogsCursor
	if raw := c.Query("cursor"); raw != "" {
		if parsed, err := strconv.ParseInt(raw, 10, 64); err == nil {
			cursor = parsed
		}
	}
	limit := constants.DefaultLogsLimit
	if raw := c.Query("limit"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil {
			limit = parsed
		}
	}
	direction := c.DefaultQuery("direction", constants.DefaultLogsDirection)

	logs, err := h.etl.GetTaskLogs(c.Request.Context(), id, req.FilePath, cursor, limit, direction)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, constants.ErrJobNotFound) {
			status = http.StatusNotFound
		}
		utils.ErrorResponse(c, status, fmt.Sprintf("failed to get task logs: %s", err), err)
		return
	}
	utils.SuccessResponse(c, fmt.Sprintf("task logs retrieved successfully for job_id[%d]", id), logs)
}

// @Summary Download task logs
// @Tags Jobs
// @Description Downloads the log file for a specific task. The file path required for the download must be obtained from the [Get Job Tasks](#/Jobs/get_api_v1_project__projectid__jobs__id__tasks) endpoint.
// @Param   projectid     path    string  true    "project id (default is 123)"
// @Param   id            path    int     true    "job id"
// @Param   file_path     query   string  true    "log file path"
// @Success 200 {file} file
// @Failure 400 {object} dto.Error400Response "failed to validate request"
// @Failure 401 {object} dto.Error401Response "unauthorized"
// @Failure 404 {object} dto.Error404Response "failed to prepare log archive"
// @Failure 500 {object} dto.Error500Response "internal server error"
// @Router /api/v1/project/{projectid}/jobs/{id}/logs/download [get]
func (h *Handler) DownloadTaskLogs(c *gin.Context) {
	id, err := utils.GetIDParam(c, "id")
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}
	filePath := c.Query("file_path")
	if filePath == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "file_path query parameter is required", nil)
		return
	}
	logger.Debugf("Download task logs initiated job_id[%d] file_path[%s]", id, filePath)
	filename, err := utils.GetLogArchiveFilename(id, filePath)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, fmt.Sprintf("failed to prepare log archive: %s", err), err)
		return
	}

	c.Header("Content-Type", "application/gzip")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
	c.Header("Cache-Control", "no-cache")
	c.Header("X-Content-Type-Options", "nosniff")
	c.Header("Access-Control-Expose-Headers", "Content-Disposition")
	c.Status(http.StatusOK)
	if err := h.etl.StreamLogArchive(id, filePath, c.Writer); err != nil {
		logger.Errorf("failed to stream log archive job_id[%d]: %s", id, err)
		return
	}
}

// @Summary (Internal) Update sync telemetry
// @Tags Internal
// @Description Internal callback to update sync telemetry data.
// @Param   body          body    dto.UpdateSyncTelemetryRequest true "telemetry data"
// @Success 200 {object} dto.JSONResponse "sync telemetry updated successfully"
// @Failure 400 {object} dto.Error400Response "failed to validate request"
// @Failure 413 {object} dto.Error413Response "payload too large"
// @Failure 500 {object} dto.Error500Response "internal server error"
// @Router /internal/worker/callback/sync-telemetry [post]
func (h *Handler) UpdateSyncTelemetry(c *gin.Context) {
	var req dto.UpdateSyncTelemetryRequest
	if err := utils.BindAndValidate(c, &req); err != nil {
		utils.ErrorResponse(c, utils.StatusFromBindError(err), fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}
	if req.JobID == 0 || req.WorkflowID == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "job_id and workflow_id are required", nil)
		return
	}
	logger.Debugf("Update sync telemetry callback initiated job_id[%d] workflow_id[%s] event[%s]", req.JobID, req.WorkflowID, req.Event)
	if err := h.etl.UpdateSyncTelemetry(c.Request.Context(), req); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to update sync telemetry", err)
		return
	}
	utils.SuccessResponse(c, fmt.Sprintf("sync telemetry updated successfully for job_id[%d] workflow_id[%s] event[%s]", req.JobID, req.WorkflowID, req.Event), nil)
}

// RecoverClearDestination handles recovery from stuck clear-destination workflows (internal use only)
// @Summary (Internal) Recover clear determination
// @Tags Internal
// @Description Internal recovery endpoint to cancel stuck clear-destination workflows and restore sync schedules.
// @Param   projectid     path    string  true    "project id (default is 123)"
// @Param   id            path    int     true    "job id"
// @Success 200 {object} dto.JSONResponse "successfully recovered"
// @Failure 400 {object} dto.Error400Response "failed to validate request"
// @Failure 404 {object} dto.Error404Response "job not found"
// @Failure 500 {object} dto.Error500Response "internal server error"
// @Router /internal/project/{projectid}/jobs/{id}/clear-destination/recover [post]
func (h *Handler) RecoverClearDestination(c *gin.Context) {
	projectID, err := utils.GetProjectID(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}
	jobID, err := utils.GetIDParam(c, "id")
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}
	logger.Debugf("Recover clear destination initiated project_id[%s] job_id[%d]", projectID, jobID)
	if err := h.etl.RecoverFromClearDestination(c.Request.Context(), projectID, jobID); err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, constants.ErrJobNotFound) {
			status = http.StatusNotFound
		}
		utils.ErrorResponse(c, status, fmt.Sprintf("failed to recover from clear-destination: %s", err), err)
		return
	}
	utils.SuccessResponse(c, fmt.Sprintf("successfully recovered from clear-destination and restored sync schedule for job_id[%d]", jobID), nil)
}

// @Summary (Internal) Update state file
// @Tags Internal
// @Description Internal endpoint to update the state file associated with a job.
// @Param   projectid     path    string  true    "project id (default is 123)"
// @Param   id            path    int     true    "job id"
// @Param   body          body    dto.UpdateStateFileRequest true "state file data"
// @Success 200 {object} dto.JSONResponse "state file updated successfully"
// @Failure 400 {object} dto.Error400Response "failed to validate request"
// @Failure 404 {object} dto.Error404Response "job not found"
// @Failure 413 {object} dto.Error413Response "payload too large"
// @Failure 500 {object} dto.Error500Response "internal server error"
// @Router /internal/project/{projectid}/jobs/{id}/statefile [put]
func (h *Handler) UpdateStateFile(c *gin.Context) {
	jobID, err := utils.GetIDParam(c, "id")
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}
	var req dto.UpdateStateFileRequest
	if err := utils.BindAndValidate(c, &req); err != nil {
		utils.ErrorResponse(c, utils.StatusFromBindError(err), fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}
	logger.Debugf("Update state file callback initiated job_id[%d]", jobID)
	if err := h.etl.UpdateStateFile(jobID, req.StateFile); err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, constants.ErrJobNotFound) {
			status = http.StatusNotFound
		}
		utils.ErrorResponse(c, status, fmt.Sprintf("failed to update state file: %s", err), err)
		return
	}
	utils.SuccessResponse(c, fmt.Sprintf("state file updated successfully for job_id[%d]", jobID), nil)
}

func validateJobDriverConfig(source, destination *dto.DriverConfig) error {
	if source != nil && source.ID == nil {
		if err := dto.ValidateSourceType(source.Type); err != nil {
			return err
		}
		if source.Name == "" || source.Version == "" || source.Config == "" {
			return fmt.Errorf("source name, version, and config are required when source id is not provided")
		}
	}
	if destination != nil && destination.ID == nil {
		if err := dto.ValidateDestinationType(destination.Type); err != nil {
			return err
		}
		if destination.Name == "" || destination.Version == "" || destination.Config == "" {
			return fmt.Errorf("destination name, version, and config are required when destination id is not provided")
		}
	}
	return nil
}
