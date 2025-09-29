package handlers

import (
	"context"
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/beego/beego/v2/server/web"

	"github.com/datazip/olake-ui/server/internal/constants"
	"github.com/datazip/olake-ui/server/internal/dto"
	"github.com/datazip/olake-ui/server/internal/models"
	"github.com/datazip/olake-ui/server/internal/services"
	"github.com/datazip/olake-ui/server/internal/telemetry"
	"github.com/datazip/olake-ui/server/utils"
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
	var req dto.LoginRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		respondWithError(&c.Controller, http.StatusBadRequest, "Invalid request format", err)
		return
	}
	if err := dto.Validate(&req); err != nil {
		respondWithError(&c.Controller, http.StatusBadRequest, "Invalid request format", err)
		return
	}

	user, err := c.authService.Login(context.Background(), req.Username, req.Password)
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

	// Optional: Validate that the user still exists in the database
	if userIDInt, ok := userID.(int); ok {
		if err := c.authService.ValidateUser(userIDInt); err != nil {
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
	_ = c.DestroySession()
	utils.SuccessResponse(&c.Controller, dto.LoginResponse{
		Message: "Logged out successfully",
		Success: true,
	})
}

// @router /signup [post]
func (c *AuthHandler) Signup() {
	var req models.User
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		respondWithError(&c.Controller, http.StatusBadRequest, "Invalid request format", err)
		return
	}
	if err := dto.Validate(&req); err != nil {
		respondWithError(&c.Controller, http.StatusBadRequest, "Invalid request format", err)
		return
	}

	if err := c.authService.Signup(context.Background(), &req); err != nil {
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
	telemetryID := telemetry.GetTelemetryUserID()
	utils.SuccessResponse(&c.Controller, map[string]interface{}{
		telemetry.TelemetryUserIDFile: string(telemetryID),
	})
}
