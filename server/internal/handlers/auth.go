package handlers

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/datazip-inc/olake-ui/server/internal/constants"
	"github.com/datazip-inc/olake-ui/server/internal/httpserver/httpx"
	"github.com/datazip-inc/olake-ui/server/internal/models/dto"
	"github.com/datazip-inc/olake-ui/server/utils/logger"
	"github.com/datazip-inc/olake-ui/server/utils/telemetry"
)

// @Summary User login
// @Tags Authentication
// @Description Authenticate a user and create a new session.
// @Param   body          body    dto.LoginRequest true "login credentials"
// @Success 200 {object} dto.JSONResponse{data=dto.LoginResponse}
// @Failure 400 {object} dto.Error400Response "invalid request"
// @Failure 401 {object} dto.Error401Response "invalid credentials"
// @Failure 413 {object} dto.Error413Response "payload too large"
// @Failure 500 {object} dto.Error500Response "internal server error"
// @Router /login [post]
func (h *Handler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := httpx.BindAndValidate(c, &req); err != nil {
		httpx.ErrorResponse(c, httpx.StatusFromBindError(err), constants.ValidationInvalidRequestFormat, err)
		return
	}
	logger.Debugf("Login initiated username[%s]", req.Username)

	user, err := h.appSvc.ETL().Login(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		if errors.Is(err, constants.ErrUserNotFound) || errors.Is(err, constants.ErrInvalidCredentials) {
			httpx.ErrorResponse(c, http.StatusUnauthorized, "Invalid credentials", err)
			return
		}
		httpx.ErrorResponse(c, http.StatusInternalServerError, fmt.Sprintf("Login failed: %s", err), err)
		return
	}

	if err := h.sessions.SetUserSession(c, user.ID); err != nil {
		httpx.ErrorResponse(c, http.StatusInternalServerError, fmt.Sprintf("Failed to create session: %s", err), err)
		return
	}

	httpx.SuccessResponse(c, "login successful", dto.LoginResponse{Username: user.Username})
}

// @Summary User signup
// @Tags Authentication
// @Description Register a new user account with the provided details.
// @Param   body          body    dto.CreateUserRequest true "user info"
// @Success 200 {object} dto.JSONResponse "user created successfully"
// @Failure 400 {object} dto.Error400Response "failed to validate request"
// @Failure 409 {object} dto.Error409Response "user already exists"
// @Failure 413 {object} dto.Error413Response "payload too large"
// @Failure 500 {object} dto.Error500Response "failed to create user"
// @Router /signup [post]
func (h *Handler) Signup(c *gin.Context) {
	var req dto.CreateUserRequest
	if err := httpx.BindAndValidate(c, &req); err != nil {
		httpx.ErrorResponse(c, httpx.StatusFromBindError(err), constants.ValidationInvalidRequestFormat, err)
		return
	}

	logger.Debugf("Signup initiated username[%s] email[%s]", req.Username, req.Email)
	if err := h.appSvc.ETL().Signup(c.Request.Context(), &req); err != nil {
		switch {
		case errors.Is(err, constants.ErrUserAlreadyExists):
			httpx.ErrorResponse(c, http.StatusConflict, fmt.Sprintf("Username already exists: %s", err), err)
		case errors.Is(err, constants.ErrPasswordProcessing):
			httpx.ErrorResponse(c, http.StatusInternalServerError, fmt.Sprintf("Failed to process password: %s", err), err)
		default:
			httpx.ErrorResponse(c, http.StatusInternalServerError, fmt.Sprintf("failed to create user: %s", err), err)
		}
		return
	}

	httpx.SuccessResponse(c, "user created successfully", map[string]interface{}{
		"email":    req.Email,
		"username": req.Username,
	})
}

// @Summary Check authentication status
// @Tags Authentication
// @Description Verify if the current user session is active and valid.
// @Success 200 {object} dto.JSONResponse
// @Failure 401 {object} dto.Error401Response "Not authenticated"
// @Router /auth/check [get]
func (h *Handler) CheckAuth(c *gin.Context) {
	userID := httpx.GetCurrentUserID(c)
	if userID == nil {
		httpx.ErrorResponse(c, http.StatusUnauthorized, "Not authenticated", errors.New("not authenticated"))
		return
	}
	logger.Debugf("Check auth initiated user_id[%v]", *userID)

	if err := h.appSvc.ETL().ValidateUser(*userID); err != nil {
		httpx.ErrorResponse(c, http.StatusUnauthorized, fmt.Sprintf("Invalid session: %s", err), err)
		return
	}

	httpx.SuccessResponse(c, "authenticated successfully", nil)
}

// @Summary Get telemetry ID
// @Tags Internal
// @Description Retrieve the unique telemetry identifier and current UI version.
// @Success 200 {object} dto.JSONResponse{data=dto.TelemetryIDResponse}
// @Failure 500 {object} dto.Error500Response "internal server error"
// @Router /telemetry-id [get]
func (h *Handler) TelemetryID(c *gin.Context) {
	logger.Info("Get telemetry ID initiated")
	httpx.SuccessResponse(c, "telemetry ID fetched successfully", dto.TelemetryIDResponse{
		TelemetryUserID: telemetry.GetTelemetryUserID(),
		OlakeUIVersion:  telemetry.GetVersion(),
	})
}
