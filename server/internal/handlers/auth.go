package handlers

import (
	"errors"
	"net/http"

	"github.com/beego/beego/v2/server/web"
	"github.com/datazip/olake-ui/server/internal/constants"
	"github.com/datazip/olake-ui/server/internal/logger"
	"github.com/datazip/olake-ui/server/internal/models"
	"github.com/datazip/olake-ui/server/internal/models/dto"
	"github.com/datazip/olake-ui/server/internal/telemetry"
	"github.com/datazip/olake-ui/server/utils"
)

// @router /login [post]
func (h *Handler) Login() {
	var req dto.LoginRequest
	if err := UnmarshalAndValidate(h.Ctx.Input.RequestBody, &req); err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusBadRequest, constants.ValidationInvalidRequestFormat, err)
		return
	}

	logger.Debugf("Login initiated username[%s]", req.Username)

	user, err := h.svc.Login(h.Ctx.Request.Context(), req.Username, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, constants.ErrUserNotFound):
			utils.ErrorResponse(&h.Controller, http.StatusUnauthorized, "user not found, sign up first", err)
		case errors.Is(err, constants.ErrInvalidCredentials):
			utils.ErrorResponse(&h.Controller, http.StatusUnauthorized, "Invalid credentials", err)
		default:
			utils.ErrorResponse(&h.Controller, http.StatusInternalServerError, "Login failed", err)
		}
		return
	}

	// check if session is enabled
	if web.BConfig.WebConfig.Session.SessionOn {
		_ = h.SetSession(constants.SessionUserID, user.ID)
	}

	utils.SuccessResponse(&h.Controller, map[string]interface{}{
		"username": user.Username,
	})
}

// @router /checkauth [get]
func (h *Handler) CheckAuth() {
	userID := h.GetSession(constants.SessionUserID)
	if userID == nil {
		utils.ErrorResponse(&h.Controller, http.StatusUnauthorized, "Not authenticated", errors.New("not authenticated"))
		return
	}

	logger.Debugf("Check auth initiated user_id[%v]", userID)

	// Optional: Validate that the user still exists in the database
	if userIDInt, ok := userID.(int); ok {
		if err := h.svc.ValidateUser(userIDInt); err != nil {
			utils.ErrorResponse(&h.Controller, http.StatusUnauthorized, "Invalid session", err)
			return
		}
	}

	utils.SuccessResponse(&h.Controller, dto.LoginResponse{
		Message: "Authenticated",
		Success: true,
	})
}

// @router /logout [post]
func (h *Handler) Logout() {
	userID := h.GetSession(constants.SessionUserID)
	logger.Debugf("Logout initiated user_id[%v]", userID)

	err := h.DestroySession()
	if err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusInternalServerError, "Failed to destroy session", err)
		return
	}

	utils.SuccessResponse(&h.Controller, nil)
}

// @router /signup [post]
func (h *Handler) Signup() {
	var req models.User
	if err := UnmarshalAndValidate(h.Ctx.Input.RequestBody, &req); err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusBadRequest, constants.ValidationInvalidRequestFormat, err)
		return
	}

	logger.Debugf("Signup initiated username[%s] email[%s]", req.Username, req.Email)

	if err := h.svc.Signup(h.Ctx.Request.Context(), &req); err != nil {
		switch {
		case errors.Is(err, constants.ErrUserAlreadyExists):
			utils.ErrorResponse(&h.Controller, http.StatusConflict, "Username already exists", err)
		case errors.Is(err, constants.ErrPasswordProcessing):
			utils.ErrorResponse(&h.Controller, http.StatusInternalServerError, "Failed to process password", err)
		default:
			utils.ErrorResponse(&h.Controller, http.StatusInternalServerError, "Failed to create user", err)
		}
		return
	}

	utils.SuccessResponse(&h.Controller, map[string]interface{}{
		"email":    req.Email,
		"username": req.Username,
	})
}

// @router /telemetry-id [get]
func (h *Handler) GetTelemetryID() {
	logger.Infof("Get telemetry ID initiated")

	telemetryID := telemetry.GetTelemetryUserID()
	utils.SuccessResponse(&h.Controller, map[string]interface{}{
		telemetry.TelemetryUserIDFile: string(telemetryID),
	})
}
