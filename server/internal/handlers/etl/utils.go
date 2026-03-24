package etl

import (
	"fmt"
	"strconv"

	"github.com/beego/beego/v2/server/web"

	"github.com/datazip-inc/olake-ui/server/internal/constants"
)

// get id from path
func GetIDFromPath(c *web.Controller) (int, error) {
	idStr := c.Ctx.Input.Param(":id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return 0, fmt.Errorf("invalid id: %s", err)
	}
	return id, nil
}

// get id from path
func GetProjectIDFromPath(c *web.Controller) (string, error) {
	projectID := c.Ctx.Input.Param(":projectid")
	if projectID == "" {
		return "", fmt.Errorf("project id is required")
	}
	return projectID, nil
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
