package handlers

import (
	"net/http"
	"strconv"

	"github.com/beego/beego/v2/server/web"
	"github.com/datazip/olake-server/utils"
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
