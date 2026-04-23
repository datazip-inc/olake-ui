package utils

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/datazip-inc/olake-ui/server/internal/constants"
	"github.com/datazip-inc/olake-ui/server/internal/models/dto"
	"github.com/datazip-inc/olake-ui/server/internal/utils/logger"
)

func GetCurrentUserID(c *gin.Context) *int {
	raw, ok := c.Get(constants.ContextUserIDKey)
	if !ok {
		return nil
	}
	id, ok := raw.(int)
	if !ok || id == 0 {
		return nil
	}
	return &id
}

func GetProjectID(c *gin.Context) (string, error) {
	projectID := c.Param(constants.ProjectIDParam)
	if projectID == "" {
		return "", fmt.Errorf("project id is required")
	}
	return projectID, nil
}

func GetIDParam(c *gin.Context) (int, error) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return 0, fmt.Errorf("invalid id: %s", err)
	}
	return id, nil
}

func BindAndValidate(c *gin.Context, target interface{}) error {
	return c.ShouldBindJSON(target)
}

func StatusFromBindError(err error) int {
	var mbe *http.MaxBytesError
	if errors.As(err, &mbe) {
		return http.StatusRequestEntityTooLarge
	}
	if strings.Contains(strings.ToLower(err.Error()), "request body too large") {
		return http.StatusRequestEntityTooLarge
	}

	return http.StatusBadRequest
}

func SuccessResponse(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, dto.JSONResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func ErrorResponse(c *gin.Context, status int, message string, err error) {
	if err != nil {
		logger.Errorf("error in request %s: %s", c.Request.URL.Path, err)
	}
	c.JSON(status, dto.JSONResponse{
		Success: false,
		Message: message,
	})
}
