package etl

import (
	"github.com/gin-gonic/gin"

	"github.com/datazip-inc/olake-ui/server/internal/httpserver/httputil"
)

func getProjectID(c *gin.Context) (string, error) {
	return httputil.GetProjectID(c)
}

func getIDParam(c *gin.Context, key string) (int, error) {
	return httputil.GetIDParam(c, key)
}

func getCurrentUserID(c *gin.Context) *int {
	return httputil.UserID(c)
}

func bindAndValidate(c *gin.Context, target interface{}) error {
	return httputil.BindAndValidate(c, target)
}

func statusFromBindError(err error) int {
	return httputil.StatusFromBindError(err)
}

func successResponse(c *gin.Context, message string, data interface{}) {
	httputil.SuccessResponse(c, message, data)
}

func errorResponse(c *gin.Context, status int, message string, err error) {
	httputil.ErrorResponse(c, status, message, err)
}
