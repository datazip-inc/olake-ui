package handlers

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/datazip-inc/olake-ui/server/internal/auth"
	"github.com/datazip-inc/olake-ui/server/internal/constants"
	"github.com/datazip-inc/olake-ui/server/internal/models"
	"github.com/datazip-inc/olake-ui/server/internal/models/dto"
	"github.com/datazip-inc/olake-ui/server/utils/logger"
	"github.com/datazip-inc/olake-ui/server/utils/telemetry"
	"github.com/gin-gonic/gin"
)

// @Summary User login
// @Tags Authentication
// @Description Authenticate a user and create a new session.
// @Param   body          body    dto.LoginRequest true "login credentials"
// @Success 200 {object} dto.JSONResponse{data=dto.LoginResponse}
// @Failure 400 {object} dto.Error400Response "invalid request"
// @Failure 401 {object} dto.Error401Response "invalid credentials"
// @Failure 500 {object} dto.Error500Response "internal server error"
// @Router /login [post]
func (h *Handler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := bindAndValidate(c, &req); err != nil {
		errorResponse(c, http.StatusBadRequest, constants.ValidationInvalidRequestFormat, err)
		return
	}
	logger.Debugf("Login initiated username[%s]", req.Username)

	user, err := h.etl.Login(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		if errors.Is(err, constants.ErrUserNotFound) || errors.Is(err, constants.ErrInvalidCredentials) {
			errorResponse(c, http.StatusUnauthorized, fmt.Sprintf("Invalid credentials: %s", err), err)
			return
		}
		errorResponse(c, http.StatusInternalServerError, fmt.Sprintf("Login failed: %s", err), err)
		return
	}

	if err := h.sessions.SetUserSession(c, user.ID); err != nil {
		errorResponse(c, http.StatusInternalServerError, fmt.Sprintf("Failed to create session: %s", err), err)
		return
	}

	successResponse(c, "login successful", dto.LoginResponse{Username: user.Username})
}

// @Summary User signup
// @Tags Authentication
// @Description Register a new user account with the provided details.
// @Param   body          body    models.User true "user info"
// @Success 200 {object} dto.JSONResponse "user created successfully"
// @Failure 400 {object} dto.Error400Response "failed to validate request"
// @Failure 409 {object} dto.Error409Response "user already exists"
// @Failure 500 {object} dto.Error500Response "failed to create user"
// @Router /signup [post]
func (h *Handler) Signup(c *gin.Context) {
	var req models.User
	if err := bindAndValidate(c, &req); err != nil {
		errorResponse(c, http.StatusBadRequest, constants.ValidationInvalidRequestFormat, err)
		return
	}

	if err := h.etl.Signup(c.Request.Context(), &req); err != nil {
		switch {
		case errors.Is(err, constants.ErrUserAlreadyExists):
			errorResponse(c, http.StatusConflict, fmt.Sprintf("Username already exists: %s", err), err)
		case errors.Is(err, constants.ErrPasswordProcessing):
			errorResponse(c, http.StatusInternalServerError, fmt.Sprintf("Failed to process password: %s", err), err)
		default:
			errorResponse(c, http.StatusInternalServerError, fmt.Sprintf("failed to create user: %s", err), err)
		}
		return
	}

	// First user ever created becomes the global admin
	if h.enforcer != nil {
		if first, err := h.etl.IsFirstUser(); err == nil && first {
			if err := auth.AssignGlobalAdmin(h.enforcer, req.ID); err != nil {
				logger.Warnf("failed to assign global admin to first user: %s", err)
			}
		}
	}

	successResponse(c, "user created successfully", map[string]interface{}{
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
	userID, ok := h.sessions.GetUserID(c)
	if !ok {
		errorResponse(c, http.StatusUnauthorized, "Not authenticated", errors.New("not authenticated"))
		return
	}
	logger.Debugf("Check auth initiated user_id[%v]", userID)

	if err := h.etl.ValidateUser(userID); err != nil {
		errorResponse(c, http.StatusUnauthorized, fmt.Sprintf("Invalid session: %s", err), err)
		return
	}

	successResponse(c, "authenticated successfully", nil)
}

// @Summary Get telemetry ID
// @Tags Internal
// @Description Retrieve the unique telemetry identifier and current UI version.
// @Success 200 {object} dto.JSONResponse{data=dto.TelemetryIDResponse}
// @Failure 500 {object} dto.Error500Response "internal server error"
// @Router /telemetry-id [get]
func (h *Handler) TelemetryID(c *gin.Context) {
	logger.Info("Get telemetry ID initiated")
	successResponse(c, "telemetry ID fetched successfully", dto.TelemetryIDResponse{
		TelemetryUserID: telemetry.GetTelemetryUserID(),
		OlakeUIVersion:  telemetry.GetVersion(),
	})
}

func (h *Handler) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !h.sessions.enabled {
			c.Next()
			return
		}
		userID, ok := h.sessions.GetUserID(c) // was: _, ok
		if !ok {
			errorResponse(c, http.StatusUnauthorized, "Unauthorized, try login again", nil)
			c.Abort()
			return
		}
		c.Set(constants.UserIDContextKey, userID) // ← new line
		c.Next()
	}
}
