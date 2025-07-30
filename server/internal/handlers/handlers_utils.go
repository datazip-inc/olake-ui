package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/beego/beego/v2/core/logs"
	"github.com/beego/beego/v2/server/web"
	"github.com/datazip/olake-ui/server/internal/constants"
	"github.com/datazip/olake-ui/server/utils"
)

// get id from path
func GetIDFromPath(c *web.Controller) int {
	idStr := c.Ctx.Input.Param(":id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid id")
		return 0
	}
	return id
}

// Helper to log and respond with error
func respondWithError(c *web.Controller, statusCode int, msg string, err error) {
	if err != nil {
		logs.Error("%s: %s", msg, err)
	}
	utils.ErrorResponse(c, statusCode, msg)
}

// Helper to extract user ID from session
func GetUserIDFromSession(c *web.Controller) *int {
	if sessionUserID := c.GetSession(constants.SessionUserID); sessionUserID != nil {
		if uid, ok := sessionUserID.(int); ok {
			return &uid
		}
	}
	return nil
}

// Helper to bind and validate JSON request
func bindJSON(c *web.Controller, target interface{}) error {
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, target); err != nil {
		return fmt.Errorf("invalid request format: %s", err)
	}
	return nil
}
