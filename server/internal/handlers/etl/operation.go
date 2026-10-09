package etl

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	services "github.com/datazip-inc/olake-ui/server/internal/services/etl"
	"github.com/datazip-inc/olake-ui/server/internal/utils"
)

// operationErrorStatus maps an operations-API error onto an HTTP status.
//
// A forbidden operation is reported as 404 rather than 403: operation IDs are the only
// thing tying an operation to a project, so confirming that someone else's ID exists
// would leak more than refusing it does.
func operationErrorStatus(err error) int {
	switch {
	case errors.Is(err, services.ErrOperationNotFound), errors.Is(err, services.ErrOperationForbidden):
		return http.StatusNotFound
	case errors.Is(err, services.ErrOperationNotDone):
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}

// parseWaitParam reads the optional long-poll duration, in seconds.
func parseWaitParam(c *gin.Context) (time.Duration, error) {
	raw := strings.TrimSpace(c.Query("wait"))
	if raw == "" {
		return 0, nil
	}
	seconds, err := strconv.Atoi(raw)
	if err != nil || seconds < 0 {
		return 0, fmt.Errorf("wait must be a non-negative number of seconds")
	}
	return time.Duration(seconds) * time.Second, nil
}

// @Summary Get async operation status
// @Tags Operations
// @Description Report whether a long-running operation is still running, and if not, how it ended. With ?wait=N the request is held server-side for up to N seconds (capped by OPERATION_MAX_WAIT) and returns as soon as the operation closes.
// @Param   projectid     path    string  true    "project id (default is 123)"
// @Param   operationid   path    string  true    "operation id returned by the submitting endpoint"
// @Param   wait          query   int     false   "seconds to long-poll before answering"
// @Success 200 {object} dto.JSONResponse{data=dto.OperationStatusResponse}
// @Failure 400 {object} dto.Error400Response "failed to validate request"
// @Failure 401 {object} dto.Error401Response "unauthorized"
// @Failure 404 {object} dto.Error404Response "operation not found"
// @Failure 500 {object} dto.Error500Response "failed to get operation status"
// @Router /api/v1/project/{projectid}/operations/{operationid} [get]
func (h *Handler) GetOperationStatus(c *gin.Context) {
	projectID, err := utils.GetProjectID(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}
	operationID := c.Param("operationid")
	wait, err := parseWaitParam(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}

	status, err := h.etl.GetOperationStatus(c.Request.Context(), projectID, operationID, wait)
	if err != nil {
		utils.ErrorResponse(c, operationErrorStatus(err), fmt.Sprintf("failed to get operation status: %s", err), err)
		return
	}
	utils.SuccessResponse(c, fmt.Sprintf("operation %s status fetched successfully", operationID), status)
}

// @Summary Get async operation result
// @Tags Operations
// @Description Fetch a finished operation's payload, shaped exactly as the submitting endpoint used to return it synchronously. Returns 409 while the operation is still running.
// @Param   projectid     path    string  true    "project id (default is 123)"
// @Param   operationid   path    string  true    "operation id returned by the submitting endpoint"
// @Success 200 {object} dto.JSONResponse{data=object}
// @Failure 400 {object} dto.Error400Response "failed to validate request"
// @Failure 401 {object} dto.Error401Response "unauthorized"
// @Failure 404 {object} dto.Error404Response "operation not found"
// @Failure 409 {object} dto.Error409Response "operation has not finished"
// @Failure 500 {object} dto.Error500Response "operation failed"
// @Router /api/v1/project/{projectid}/operations/{operationid}/result [get]
func (h *Handler) GetOperationResult(c *gin.Context) {
	projectID, err := utils.GetProjectID(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}
	operationID := c.Param("operationid")

	result, err := h.etl.GetOperationResult(c.Request.Context(), projectID, operationID)
	if err != nil {
		utils.ErrorResponse(c, operationErrorStatus(err), fmt.Sprintf("failed to get operation result: %s", err), err)
		return
	}
	utils.SuccessResponse(c, fmt.Sprintf("operation %s result fetched successfully", operationID), result)
}

// @Summary Cancel an async operation
// @Tags Operations
// @Description Stop a running operation and the connector container behind it.
// @Param   projectid     path    string  true    "project id (default is 123)"
// @Param   operationid   path    string  true    "operation id returned by the submitting endpoint"
// @Success 200 {object} dto.JSONResponse
// @Failure 400 {object} dto.Error400Response "failed to validate request"
// @Failure 401 {object} dto.Error401Response "unauthorized"
// @Failure 404 {object} dto.Error404Response "operation not found"
// @Failure 500 {object} dto.Error500Response "failed to cancel operation"
// @Router /api/v1/project/{projectid}/operations/{operationid} [delete]
func (h *Handler) CancelOperation(c *gin.Context) {
	projectID, err := utils.GetProjectID(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}
	operationID := c.Param("operationid")

	if err := h.etl.CancelOperation(c.Request.Context(), projectID, operationID); err != nil {
		utils.ErrorResponse(c, operationErrorStatus(err), fmt.Sprintf("failed to cancel operation: %s", err), err)
		return
	}
	utils.SuccessResponse(c, fmt.Sprintf("operation %s cancelled successfully", operationID), nil)
}
