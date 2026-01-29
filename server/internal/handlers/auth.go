package handlers

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/beego/beego/v2/server/web"
	"github.com/datazip-inc/olake-ui/server/internal/constants"
	"github.com/datazip-inc/olake-ui/server/internal/models"
	"github.com/datazip-inc/olake-ui/server/internal/models/dto"
	"github.com/datazip-inc/olake-ui/server/utils"
	"github.com/datazip-inc/olake-ui/server/utils/logger"
	"github.com/datazip-inc/olake-ui/server/utils/telemetry"
)

// @Title Login
// @Tags Authentication
// @Description Authenticate a user and create a new session.
// @Param   body          body    dto.LoginRequest true "login credentials"
// @Success 200 {object} dto.JSONResponse{data=dto.LoginResponse}
// @Failure 400 {object} dto.Error400Response "invalid request"
// @Failure 401 {object} dto.Error401Response "invalid credentials"
// @Failure 500 {object} dto.Error500Response "internal server error"
// @Router /login [post]
func (h *Handler) Login() {
	var req dto.LoginRequest
	if err := UnmarshalAndValidate(h.Ctx.Input.RequestBody, &req); err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusBadRequest, constants.ValidationInvalidRequestFormat, err)
		return
	}

	logger.Debugf("Login initiated username[%s]", req.Username)

	user, err := h.etl.Login(h.Ctx.Request.Context(), req.Username, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, constants.ErrUserNotFound):
			utils.ErrorResponse(&h.Controller, http.StatusUnauthorized, fmt.Sprintf("user not found, sign up first: %s", err), err)
		case errors.Is(err, constants.ErrInvalidCredentials):
			utils.ErrorResponse(&h.Controller, http.StatusUnauthorized, fmt.Sprintf("Invalid credentials: %s", err), err)
		default:
			utils.ErrorResponse(&h.Controller, http.StatusInternalServerError, fmt.Sprintf("Login failed: %s", err), err)
		}
		return
	}

	// check if session is enabled
	if web.BConfig.WebConfig.Session.SessionOn {
		_ = h.SetSession(constants.SessionUserID, user.ID)
	}

	utils.SuccessResponse(&h.Controller, "login successful", dto.LoginResponse{
		Username: user.Username,
	})
}

// @Title CheckAuth
// @Tags Authentication
// @Description Verify if the current user session is active and valid.
// @Success 200 {object} dto.JSONResponse
// @Failure 401 {object} dto.Error401Response "Not authenticated"
// @Router /auth/check [get]
func (h *Handler) CheckAuth() {
	userID := h.GetSession(constants.SessionUserID)
	if userID == nil {
		utils.ErrorResponse(&h.Controller, http.StatusUnauthorized, "Not authenticated", errors.New("not authenticated"))
		return
	}

	logger.Debugf("Check auth initiated user_id[%v]", userID)

	// Optional: Validate that the user still exists in the database
	if userIDInt, ok := userID.(int); ok {
		if err := h.etl.ValidateUser(userIDInt); err != nil {
			utils.ErrorResponse(&h.Controller, http.StatusUnauthorized, fmt.Sprintf("Invalid session: %s", err), err)
			return
		}
	}

	utils.SuccessResponse(&h.Controller, "authenticated successfully", nil)
}

func (h *Handler) Logout() {
	userID := h.GetSession(constants.SessionUserID)
	logger.Debugf("Logout initiated user_id[%v]", userID)

	err := h.DestroySession()
	if err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusInternalServerError, fmt.Sprintf("Failed to destroy session: %s", err), err)
		return
	}

	utils.SuccessResponse(&h.Controller, "logout successful", nil)
}

// @Title Signup
// @Tags Authentication
// @Description Register a new user account with the provided details.
// @Param   body          body    models.User true "user info"
// @Success 200 {object} dto.JSONResponse "user created successfully"
// @Failure 400 {object} dto.Error400Response "failed to validate request"
// @Failure 409 {object} dto.Error409Response "user already exists"
// @Failure 500 {object} dto.Error500Response "failed to create user"
// @Router /signup [post]
func (h *Handler) Signup() {
	var req models.User
	if err := UnmarshalAndValidate(h.Ctx.Input.RequestBody, &req); err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusBadRequest, constants.ValidationInvalidRequestFormat, err)
		return
	}

	if err := h.etl.Signup(h.Ctx.Request.Context(), &req); err != nil {
		switch {
		case errors.Is(err, constants.ErrUserAlreadyExists):
			utils.ErrorResponse(&h.Controller, http.StatusConflict, fmt.Sprintf("Username already exists: %s", err), err)
		case errors.Is(err, constants.ErrPasswordProcessing):
			utils.ErrorResponse(&h.Controller, http.StatusInternalServerError, fmt.Sprintf("Failed to process password: %s", err), err)
		default:
			utils.ErrorResponse(&h.Controller, http.StatusInternalServerError, fmt.Sprintf("failed to create user: %s", err), err)
		}
		return
	}

	utils.SuccessResponse(&h.Controller, "user created successfully", map[string]interface{}{
		"email":    req.Email,
		"username": req.Username,
	})
}

// @Title GetTelemetryID
// @Tags Internal
// @Description Retrieve the unique telemetry identifier for the current installation.
// @Success 200 {object} dto.JSONResponse{data=dto.TelemetryIDResponse}
// @Failure 500 {object} dto.Error500Response "internal server error"
// @Router /telemetry-id [get]
func (h *Handler) GetTelemetryID() {
	logger.Info("Get telemetry ID initiated")

	telemetryID := telemetry.GetTelemetryUserID()
	version := telemetry.GetVersion()
	utils.SuccessResponse(&h.Controller, "telemetry ID fetched successfully", dto.TelemetryIDResponse{
		TelemetryUserID: telemetryID,
		OlakeUIVersion:  version,
	})
}
