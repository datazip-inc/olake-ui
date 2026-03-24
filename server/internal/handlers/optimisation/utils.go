package optimisation

import "github.com/datazip-inc/olake-ui/server/utils"

func (h *Handler) bindJSON(dst interface{}) bool {
	if err := h.Ctx.BindJSON(dst); err != nil {
		utils.ErrorResponse(&h.Controller, badRequestStatusCode, "invalid request body", err)
		return false
	}
	return true
}
