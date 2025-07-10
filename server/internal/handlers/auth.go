package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/beego/beego/v2/server/web"

	"github.com/datazip/olake-frontend/server/internal/constants"
	"github.com/datazip/olake-frontend/server/internal/models"
	"github.com/datazip/olake-frontend/server/internal/services"
	"github.com/datazip/olake-frontend/server/internal/telemetry"
	"github.com/datazip/olake-frontend/server/utils"
)

type AuthHandler struct {
	web.Controller
	authService *services.AuthService
}

func (c *AuthHandler) Prepare() {
	c.authService = services.NewAuthService()
}

// @router /login [post]
func (c *AuthHandler) Login() {
	var req models.LoginRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, "Invalid request format")
		return
	}

	user, err := c.authService.Login(req.Username, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrUserNotFound):
			utils.ErrorResponse(&c.Controller, http.StatusUnauthorized, "user not found, sign up first")
		case errors.Is(err, services.ErrInvalidCredentials):
			utils.ErrorResponse(&c.Controller, http.StatusUnauthorized, "Invalid credentials")
		default:
			utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, "Login failed")
		}
		return
	}

	// check if session is enabled
	if web.BConfig.WebConfig.Session.SessionOn {
		_ = c.SetSession(constants.SessionUserID, user.ID)
	}

	telemetry.TrackUserLogin(c.Ctx.Request.Context(), user)

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

	// Optional: Validate that the user still exists in the database
	if userIDInt, ok := userID.(int); ok {
		if err := c.authService.ValidateUser(userIDInt); err != nil {
			utils.ErrorResponse(&c.Controller, http.StatusUnauthorized, "Invalid session")
			return
		}
	}

	utils.SuccessResponse(&c.Controller, models.LoginResponse{
		Message: "Authenticated",
		Success: true,
	})
}

// @router /logout [post]
func (c *AuthHandler) Logout() {
	_ = c.DestroySession()
	utils.SuccessResponse(&c.Controller, models.LoginResponse{
		Message: "Logged out successfully",
		Success: true,
	})
}

// @router /signup [post]
func (c *AuthHandler) Signup() {
	var req models.User
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, "Invalid request format")
		return
	}

	if err := c.authService.Signup(&req); err != nil {
		switch {
		case errors.Is(err, services.ErrUserAlreadyExists):
			utils.ErrorResponse(&c.Controller, http.StatusConflict, "Username already exists")
		case errors.Is(err, services.ErrPasswordProcessing):
			utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, "Failed to process password")
		default:
			utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, "Failed to create user")
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
	telemetryID := telemetry.GetTelemetryUserID()
	utils.SuccessResponse(&c.Controller, map[string]interface{}{
		telemetry.TelemetryUserIDFile: string(telemetryID),
	})
}
