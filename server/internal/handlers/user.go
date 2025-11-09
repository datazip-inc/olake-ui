package handlers

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/datazip-inc/olake-ui/server/internal/models"
	"github.com/datazip-inc/olake-ui/server/utils"
	"github.com/datazip-inc/olake-ui/server/utils/logger"
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

	logger.Infof("Create user initiated username[%s] email[%s]", req.Username, req.Email)

	if err := h.etl.CreateUser(h.Ctx.Request.Context(), &req); err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusInternalServerError, fmt.Sprintf("failed to create user: %s", err), err)
		return
	}

	utils.SuccessResponse(&h.Controller, "user created successfully", req)
}

// @router /users [get]
func (h *Handler) GetAllUsers() {
	logger.Info("Get all users initiated")

	users, err := h.etl.GetAllUsers(h.Ctx.Request.Context())
	if err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusInternalServerError, fmt.Sprintf("failed to get users: %s", err), err)
		return
	}

	utils.SuccessResponse(&h.Controller, "users listed successfully", users)
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

	logger.Infof("Update user initiated user_id[%d] username[%s]", id, req.Username)

	updatedUser, err := h.etl.UpdateUser(h.Ctx.Request.Context(), id, &req)
	if err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusInternalServerError, fmt.Sprintf("failed to update user: %s", err), err)
		return
	}

	utils.SuccessResponse(&h.Controller, "user updated successfully", updatedUser)
}

// @router /users/:id [delete]
func (h *Handler) DeleteUser() {
	id, err := GetIDFromPath(&h.Controller)
	if err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusBadRequest, fmt.Sprintf("failed to validate request: %s", err), err)
		return
	}

	logger.Infof("Delete user initiated user_id[%d]", id)

	if err := h.etl.DeleteUser(h.Ctx.Request.Context(), id); err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusInternalServerError, fmt.Sprintf("failed to delete user: %s", err), err)
		return
	}

	utils.SuccessResponse(&h.Controller, "user deleted successfully", nil)
}
