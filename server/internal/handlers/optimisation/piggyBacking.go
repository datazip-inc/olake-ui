package optimisation

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// forwards any /api/ams/v1/* request to optimisation,
func (h *Handler) PiggyBacking() {
	req := h.Ctx.Request

	var body json.RawMessage
	if req.ContentLength > 0 {
		raw, err := io.ReadAll(req.Body)
		if err != nil {
			h.Ctx.Output.SetStatus(http.StatusBadRequest)
			_, _ = h.Ctx.ResponseWriter.Write([]byte(fmt.Sprintf(`{"error":"failed to read request body: %s"}`, err)))
			return
		}
		
		body = json.RawMessage(raw)
	}

	data, err := h.opt.Proxy(req.Context(), req.Method, req.URL.Path, req.URL.Query(), body)
	if err != nil {
		h.Ctx.Output.SetStatus(http.StatusBadGateway)
		_, _ = h.Ctx.ResponseWriter.Write([]byte(fmt.Sprintf(`{"error":"upstream request failed: %s"}`, err)))
		return
	}

	h.Ctx.Output.Header("Content-Type", "application/json")
	h.Ctx.Output.SetStatus(http.StatusOK)
	_, _ = h.Ctx.ResponseWriter.Write(data)
}
