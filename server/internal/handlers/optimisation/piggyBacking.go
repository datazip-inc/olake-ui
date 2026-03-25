package optimisation

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/datazip-inc/olake-ui/server/utils"
)

const maxProxyBodyBytes = 10 << 20 // 10 MiB

// forwardAMSHTTPStatus returns the upstream status when it is a normal HTTP client/server code; otherwise 502.
func forwardAMSHTTPStatus(status int) int {
	if status >= 100 && status <= 599 {
		return status
	}
	return http.StatusBadGateway
}

// PiggyBacking forwards any /api/opt/v1/* request to optimisation service.
// Returns standardized JSON response format for consistency with other APIs.
func (h *Handler) PiggyBacking() {
	req := h.Ctx.Request

	var body json.RawMessage
	if req.Body != nil {
		limited := http.MaxBytesReader(h.Ctx.ResponseWriter, req.Body, maxProxyBodyBytes)
		raw, err := io.ReadAll(limited)
		if err != nil {
			var maxErr *http.MaxBytesError
			if errors.As(err, &maxErr) {
				utils.ErrorResponse(&h.Controller, http.StatusRequestEntityTooLarge, "request body too large", err)
				return
			}
			utils.ErrorResponse(&h.Controller, http.StatusBadRequest, "failed to read request body", err)
			return
		}
		if len(raw) > 0 {
			body = json.RawMessage(raw)
		}
	}

	transformedPath := transformOptPathToAMS(req.URL.Path)

	httpStatus, data, err := h.opt.Proxy(req.Context(), req.Method, transformedPath, req.URL.Query(), body)
	if err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusBadGateway, "upstream request failed", err)
		return
	}

	if httpStatus >= 400 {
		msg := upstreamErrorMessage(data, "upstream request failed")
		utils.ErrorResponse(&h.Controller, forwardAMSHTTPStatus(httpStatus), msg, nil)
		return
	}

	// AMS may use HTTP 2xx with a JSON envelope where code != 200 indicates failure.
	var amsCode struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(data, &amsCode); err == nil && amsCode.Code != 0 && amsCode.Code != http.StatusOK {
		msg := amsCode.Message
		if msg == "" {
			msg = "upstream request failed"
		}
		utils.ErrorResponse(&h.Controller, forwardAMSHTTPStatus(amsCode.Code), msg, nil)
		return
	}

	if httpStatus == http.StatusNoContent {
		h.Ctx.Output.SetStatus(http.StatusNoContent)
		return
	}

	var upstreamResponse interface{}
	if err := json.Unmarshal(data, &upstreamResponse); err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusInternalServerError, "failed to parse upstream response", err)
		return
	}

	outStatus := httpStatus
	if outStatus < 200 || outStatus > 299 {
		outStatus = http.StatusOK
	}
	utils.RespondJSON(&h.Controller, outStatus, true, "request forwarded successfully", upstreamResponse)
}

func upstreamErrorMessage(data []byte, fallback string) string {
	var er struct {
		Message string `json:"message"`
	}
	if len(data) > 0 && json.Unmarshal(data, &er) == nil && er.Message != "" {
		return er.Message
	}
	return fallback
}

func transformOptPathToAMS(path string) string {
	const optPrefix = "/api/opt/v1/"
	const amsPrefix = "/api/ams/v1/"

	if len(path) >= len(optPrefix) && path[:len(optPrefix)] == optPrefix {
		return amsPrefix + path[len(optPrefix):]
	}
	return path
}
