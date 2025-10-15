package handlers

import (
	"context"
	"errors"
	"net/http"

	"github.com/beego/beego/v2/server/web"
	"github.com/datazip/olake-ui/server/internal/constants"
	"github.com/datazip/olake-ui/server/internal/dto"
	"github.com/datazip/olake-ui/server/internal/logger"
	"github.com/datazip/olake-ui/server/internal/models"
	"github.com/datazip/olake-ui/server/internal/telemetry"
	"github.com/datazip/olake-ui/server/utils"
)

type AuthHandler struct {
	web.Controller
}

// @router /login [post]
func (c *AuthHandler) Login() {
	var req dto.LoginRequest
	if err := UnmarshalAndValidate(c.Ctx.Input.RequestBody, &req); err != nil {
		respondWithError(&c.Controller, http.StatusBadRequest, constants.ValidationInvalidRequestFormat, err)
		return
	}

	logger.Info("Login initiated - username=%s", req.Username)

	user, err := AuthSvc().Login(context.Background(), req.Username, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, constants.ErrUserNotFound):
			respondWithError(&c.Controller, http.StatusUnauthorized, "user not found, sign up first", err)
		case errors.Is(err, constants.ErrInvalidCredentials):
			respondWithError(&c.Controller, http.StatusUnauthorized, "Invalid credentials", err)
		default:
			respondWithError(&c.Controller, http.StatusInternalServerError, "Login failed", err)
		}
		return
	}

	// check if session is enabled
	if web.BConfig.WebConfig.Session.SessionOn {
		_ = c.SetSession(constants.SessionUserID, user.ID)
	}

	utils.SuccessResponse(&c.Controller, map[string]interface{}{
		"username": user.Username,
	})
}

// @router /checkauth [get]
func (c *AuthHandler) CheckAuth() {
	userID := c.GetSession(constants.SessionUserID)
	if userID == nil {
		utils.ErrorResponse(&c.Controller, http.StatusUnauthorized, "Not authenticated")
		return
	}

	logger.Info("Check auth initiated - user_id=%v", userID)

	// Optional: Validate that the user still exists in the database
	if userIDInt, ok := userID.(int); ok {
		if err := AuthSvc().ValidateUser(userIDInt); err != nil {
			utils.ErrorResponse(&c.Controller, http.StatusUnauthorized, "Invalid session")
			return
		}
	}

	utils.SuccessResponse(&c.Controller, dto.LoginResponse{
		Message: "Authenticated",
		Success: true,
	})
}

// @router /logout [post]
func (c *AuthHandler) Logout() {
	userID := c.GetSession(constants.SessionUserID)
	logger.Info("Logout initiated - user_id=%v", userID)

	_ = c.DestroySession()
	utils.SuccessResponse(&c.Controller, dto.LoginResponse{
		Message: "Logged out successfully",
		Success: true,
	})
}

// @router /signup [post]
func (c *AuthHandler) Signup() {
	var req models.User
	if err := UnmarshalAndValidate(c.Ctx.Input.RequestBody, &req); err != nil {
		respondWithError(&c.Controller, http.StatusBadRequest, constants.ValidationInvalidRequestFormat, err)
		return
	}

	logger.Info("Signup initiated - username=%s email=%s", req.Username, req.Email)

	if err := AuthSvc().Signup(context.Background(), &req); err != nil {
		switch {
		case errors.Is(err, constants.ErrUserAlreadyExists):
			respondWithError(&c.Controller, http.StatusConflict, "Username already exists", err)
		case errors.Is(err, constants.ErrPasswordProcessing):
			respondWithError(&c.Controller, http.StatusInternalServerError, "Failed to process password", err)
		default:
			respondWithError(&c.Controller, http.StatusInternalServerError, "Failed to create user", err)
		}
		return
	}

	utils.SuccessResponse(&c.Controller, map[string]interface{}{
		"email":    req.Email,
		"username": req.Username,
	})
}

// @router /telemetry-id [get]
func (c *AuthHandler) GetTelemetryID() {
	logger.Info("Get telemetry ID initiated")

	telemetryID := telemetry.GetTelemetryUserID()
	utils.SuccessResponse(&c.Controller, map[string]interface{}{
		telemetry.TelemetryUserIDFile: string(telemetryID),
	})
}
