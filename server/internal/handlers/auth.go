package handlers

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/beego/beego/v2/core/logs"
	"github.com/beego/beego/v2/server/web"
	"golang.org/x/crypto/bcrypt"

	"github.com/datazip/olake-frontend/server/internal/constants"
	"github.com/datazip/olake-frontend/server/internal/database"
	"github.com/datazip/olake-frontend/server/internal/models"
	"github.com/datazip/olake-frontend/server/internal/telemetry"
	telemetryutils "github.com/datazip/olake-frontend/server/internal/telemetry/utils"
	"github.com/datazip/olake-frontend/server/utils"
)

type AuthHandler struct {
	web.Controller
	userORM *database.UserORM
}

func (c *AuthHandler) Prepare() {
	c.userORM = database.NewUserORM()
}

// @router /login [post]
func (c *AuthHandler) Login() {
	var req models.LoginRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, "Invalid request format")
		return
	}

	user, err := c.userORM.FindByUsername(req.Username)
	if err != nil {
		ErrorResponse := "Invalid credentials"
		if strings.Contains(err.Error(), "no row found") {
			ErrorResponse = "user not found, sign up first"
		}
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, ErrorResponse)
		return
	}

	if err := c.userORM.ComparePassword(user.Password, req.Password); err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, "Invalid credentials")
		return
	}

	// check if session is enabled
	if web.BConfig.WebConfig.Session.SessionOn {
		_ = c.SetSession(constants.SessionUserID, user.ID)
	}

	go func() {
		if err := telemetry.TrackUserLogin(c.Ctx.Request.Context(), user.ID, user.Email, user.Username); err != nil {
			logs.Error("Failed to track user login: %v", err)
		}
	}()

	utils.SuccessResponse(&c.Controller, map[string]interface{}{
		"username": user.Username,
	})
}

// @router /checkauth [get]
func (c *AuthHandler) CheckAuth() {
	if userID := c.GetSession(constants.SessionUserID); userID == nil {
		utils.ErrorResponse(&c.Controller, http.StatusUnauthorized, "Not authenticated")
		return
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

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, "Failed to process password")
		return
	}
	req.Password = string(hashedPassword)

	if err := c.userORM.Create(&req); err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusConflict, "Username already exists")
		return
	}

	utils.SuccessResponse(&c.Controller, map[string]interface{}{
		"email":    req.Email,
		"username": req.Username,
	})
}

// @router /telemetry-id [get]
func (c *AuthHandler) GetTelemetryID() {
	// read from /tmp/olake/telemetry_id
	telemetryID, err := os.ReadFile(filepath.Join(os.TempDir(), "olake", telemetryutils.TelemetryAnonymousIDFile))
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, "Failed to retrieve telemetry ID")
		return
	}

	utils.SuccessResponse(&c.Controller, map[string]interface{}{
		telemetryutils.TelemetryAnonymousIDFile: string(telemetryID),
	})
}
