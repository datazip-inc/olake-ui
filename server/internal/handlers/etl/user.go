package etlhandlers

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/datazip-inc/olake-ui/server/internal/models"
	"github.com/datazip-inc/olake-ui/server/utils"
	"github.com/datazip-inc/olake-ui/server/utils/logger"
)

// @Summary Create a new user
// @Tags Users
// @Description Create a new user record with the provided details.
// @Param   body    body    models.User true    "user info"
// @Success 200 {object} dto.JSONResponse "user created successfully"
// @Failure 400 {object} dto.Error400Response "failed to validate request"
// @Failure 401 {object} dto.Error401Response "unauthorized"
// @Failure 500 {object} dto.Error500Response "failed to create user"
// @Router /api/v1/users [post]
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

// @Summary List all users
// @Tags Users
// @Description Retrieve a list of all registered users.
// @Success 200 {array}  dto.JSONResponse{data=models.User}
// @Failure 400 {object} dto.Error400Response "failed to validate request"
// @Failure 401 {object} dto.Error401Response "unauthorized"
// @Failure 500 {object} dto.Error500Response "failed to get users"
// @Router /api/v1/users [get]
func (h *Handler) GetAllUsers() {
	logger.Info("Get all users initiated")

	users, err := h.etl.GetAllUsers(h.Ctx.Request.Context())
	if err != nil {
		utils.ErrorResponse(&h.Controller, http.StatusInternalServerError, fmt.Sprintf("failed to get users: %s", err), err)
		return
	}

	utils.SuccessResponse(&h.Controller, "users listed successfully", users)
}

// @Summary Update user details
// @Tags Users
// @Description Update the details of an existing user identified by their unique ID.
// @Param   id      path    int true    "user id"
// @Param   body    body    models.User true    "user info"
// @Success 200 {object} dto.JSONResponse{data=models.User}
// @Failure 400 {object} dto.Error400Response "failed to validate request"
// @Failure 401 {object} dto.Error401Response "unauthorized"
// @Failure 500 {object} dto.Error500Response "failed to update user"
// @Router /api/v1/users/{id} [put]
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

// @Summary Delete a user
// @Tags Users
// @Description Permanently remove a user record identified by their unique ID.
// @Param   id      path    int true    "user id"
// @Success 200 {object} dto.JSONResponse "user deleted successfully"
// @Failure 400 {object} dto.Error400Response "failed to validate request"
// @Failure 401 {object} dto.Error401Response "unauthorized"
// @Failure 500 {object} dto.Error500Response "failed to delete user"
// @Router /api/v1/users/{id} [delete]
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
