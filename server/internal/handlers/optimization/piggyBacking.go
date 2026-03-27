package optimization

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/datazip-inc/olake-ui/server/utils"
)

// PiggyBacking forwards any /api/opt/v1/* request to optimization service.
// Returns standardized JSON response format for consistency with other APIs.
func (h *Handler) PiggyBacking() {
	req := h.Ctx.Request

	var body json.RawMessage
	if req.ContentLength > 0 {
		raw, err := io.ReadAll(req.Body)
		if err != nil {
			utils.ErrorResponse(&h.Controller, http.StatusBadRequest, "failed to read request body", err)
			return
		}

		body = json.RawMessage(raw)
	}

	transformedPath := transformOptPathToAMS(req.URL.Path)

	data, headers, err := h.opt.ProxyWithHeaders(req.Context(), req.Method, transformedPath, req.URL.Query(), body)
	if err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusBadGateway, "upstream request failed", err)
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
			h.Ctx.Output.Header("Content-Type", contentType)
		}
		if contentDisposition != "" {
			h.Ctx.Output.Header("Content-Disposition", contentDisposition)
		}
		h.Ctx.Output.Body(data)
		return
	}

	// Parse upstream response to re-wrap in standard format
	var upstreamResponse interface{}
	if err := json.Unmarshal(data, &upstreamResponse); err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusInternalServerError, "failed to parse upstream response", err)
		return
	}

	utils.SuccessResponse(&h.Controller, "request forwarded successfully", upstreamResponse)
}

func transformOptPathToAMS(path string) string {
	const optPrefix = "/api/opt/v1/"
	const amsPrefix = "/api/ams/v1/"

	if len(path) >= len(optPrefix) && path[:len(optPrefix)] == optPrefix {
		return amsPrefix + path[len(optPrefix):]
	}
	return path
}
