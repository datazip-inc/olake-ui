package optimization

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/datazip-inc/olake-ui/server/internal/httpserver/httputil"
	"github.com/datazip-inc/olake-ui/server/internal/models/dto"
)

// PiggyBacking forwards any /api/opt/v1/* request to optimization service.
// Returns standardized JSON response format for consistency with other APIs.
func (h *Handler) PiggyBacking(c *gin.Context) {
	req := c.Request

	var body json.RawMessage
	if req.ContentLength > 0 {
		raw, err := io.ReadAll(req.Body)
		if err != nil {
			httputil.ErrorResponse(c, http.StatusBadRequest, "failed to read request body", err)
			return
		}

		body = json.RawMessage(raw)
	}

	transformedPath := strings.Replace(req.URL.Path, "/api/opt/v1/", "/api/ams/v1/", 1)

	data, statusCode, headers, err := h.opt.ProxyWithHeaders(req.Context(), req.Method, transformedPath, req.URL.Query(), body)
	if err != nil {
		if statusCode == 0 {
			statusCode = http.StatusBadGateway
		}
		httputil.ErrorResponse(c, statusCode, "upstream request failed", err)
		return
	}

	// Check if response is a file download
	contentType := headers.Get("Content-Type")
	contentDisposition := headers.Get("Content-Disposition")

	isFileDownload := strings.Contains(contentDisposition, "attachment") ||
		strings.Contains(contentType, "application/octet-stream") ||
		strings.Contains(contentType, "application/x-tar") ||
		strings.Contains(contentType, "application/gzip")

	if isFileDownload {
		// Stream file directly without JSON wrapping
		if contentType != "" {
			c.Header("Content-Type", contentType)
		}
		if contentDisposition != "" {
			c.Header("Content-Disposition", contentDisposition)
		}
		c.Data(statusCode, contentType, data)
		return
	}

	// Parse upstream response to re-wrap in standard format
	var upstreamResponse interface{}
	if err := json.Unmarshal(data, &upstreamResponse); err != nil {
		httputil.ErrorResponse(c, http.StatusInternalServerError, "failed to parse upstream response", err)
		return
	}

	var optResp dto.OptimizationResponse
	if jsonErr := json.Unmarshal(data, &optResp); jsonErr == nil && optResp.Code != 0 && optResp.Code != 200 {
		httputil.ErrorResponse(c, optResp.Code, optResp.Message, nil)
		return
	}

	// TODO BEFORE MERGE
	c.JSON(statusCode, dto.JSONResponse{
		Success: true,
		Message: "request forwarded successfully",
		Data:    upstreamResponse,
	})
}
