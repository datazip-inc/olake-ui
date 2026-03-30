package etl

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/datazip-inc/olake-ui/server/internal/models/dto"
	"github.com/datazip-inc/olake-ui/server/utils/logger"
)

func getProjectID(c *gin.Context) (string, error) {
	projectID := c.Param("projectid")
	if projectID == "" {
		return "", fmt.Errorf("project id is required")
	}
	return projectID, nil
}

func getIDParam(c *gin.Context, key string) (int, error) {
	id, err := strconv.Atoi(c.Param(key))
	if err != nil {
		return 0, fmt.Errorf("invalid %s: %s", key, err)
	}
	return id, nil
}

func getCurrentUserID(c *gin.Context, sessions *.SessionStore) *int {
	userID, ok := sessions.GetUserID(c)
	if !ok {
		return nil
	}
	return &userID
}

func bindAndValidate(c *gin.Context, target interface{}) error {
	return c.ShouldBindJSON(target)
}

func statusFromBindError(err error) int {
	var mbe *http.MaxBytesError
	if errors.As(err, &mbe) {
		return http.StatusRequestEntityTooLarge
	}
	if strings.Contains(strings.ToLower(err.Error()), "request body too large") {
		return http.StatusRequestEntityTooLarge
	}

	return http.StatusBadRequest
}

func successResponse(c *gin.Context, message string, data interface{}) {
	c.JSON(200, dto.JSONResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func errorResponse(c *gin.Context, status int, message string, err error) {
	if err != nil {
		logger.Errorf("error in request %s: %s", c.Request.URL.Path, err)
	}
	c.JSON(status, dto.JSONResponse{
		Success: false,
		Message: message,
	})
}
