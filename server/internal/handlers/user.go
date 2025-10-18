package handlers

import (
	"fmt"
	"net/http"

	"errors"

	"github.com/datazip/olake-ui/server/internal/logger"
	"github.com/datazip/olake-ui/server/internal/models"
	"github.com/datazip/olake-ui/server/utils"
)

// @router /users [post]
func (h *Handler) CreateUser() {
	var req models.User
	if err := UnmarshalAndValidate(h.Ctx.Input.RequestBody, &req); err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusBadRequest, fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}

	if req.Username == "" || req.Email == "" || req.Password == "" {
		utils.ErrorResponse(&h.Controller, http.StatusBadRequest, fmt.Sprintf("failed to validate request: %s", errors.New("missing required user fields")), errors.New("missing required user fields"))
		return
	}

	logger.Info("Create user initiated username[%s] email[%s]", req.Username, req.Email)

	if err := h.svc.CreateUser(h.Ctx.Request.Context(), &req); err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusInternalServerError, fmt.Sprintf("failed to create user: %s", err), err)
		return
	}

	utils.SuccessResponse(&h.Controller, req)
}

// @router /users [get]
func (h *Handler) GetAllUsers() {
	logger.Info("Get all users initiated")

	users, err := h.svc.GetAllUsers(h.Ctx.Request.Context())
	if err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusInternalServerError, fmt.Sprintf("failed to get users: %s", err), err)
		return
	}

	utils.SuccessResponse(&h.Controller, users)
}

// @router /users/:id [put]
func (h *Handler) UpdateUser() {
	id, err := GetIDFromPath(&h.Controller)
	if err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusBadRequest, fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}

	var req models.User
	if err := UnmarshalAndValidate(h.Ctx.Input.RequestBody, &req); err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusBadRequest, fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}

	logger.Info("Update user initiated user_id[%d] username[%s]", id, req.Username)

	updatedUser, err := h.svc.UpdateUser(h.Ctx.Request.Context(), id, &req)
	if err != nil {
		if err.Error() == "user not found" {
			utils.ErrorResponse(&h.Controller, http.StatusNotFound, fmt.Sprintf("user not found: %s", err), err)
			return
		}
		utils.ErrorResponse(&h.Controller, http.StatusInternalServerError, fmt.Sprintf("failed to update user: %s", err), err)
		return
	}

	utils.SuccessResponse(&h.Controller, updatedUser)
}

// @router /users/:id [delete]
func (h *Handler) DeleteUser() {
	id, err := GetIDFromPath(&h.Controller)
	if err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusBadRequest, fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}

	logger.Info("Delete user initiated user_id[%d]", id)

	if err := h.svc.DeleteUser(h.Ctx.Request.Context(), id); err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusInternalServerError, fmt.Sprintf("failed to delete user: %s", err), err)
		return
	}

	utils.SuccessResponse(&h.Controller, nil)
}
