package handlers

import (
	"context"
	"net/http"

	"github.com/beego/beego/v2/server/web"
	"github.com/datazip/olake-ui/server/internal/constants"
	"github.com/datazip/olake-ui/server/internal/logger"
	"github.com/datazip/olake-ui/server/internal/models"
	"github.com/datazip/olake-ui/server/utils"
)

type UserHandler struct {
	web.Controller
}

// @router /users [post]
func (c *UserHandler) CreateUser() {
	var req models.User
	if err := UnmarshalAndValidate(c.Ctx.Input.RequestBody, &req); err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, constants.ValidationInvalidRequestFormat, err)
		return
	}

	logger.Info("Create user initiated - username=%s email=%s", req.Username, req.Email)

	if err := svc.User.CreateUser(context.Background(), &req); err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, "Failed to create user", err)
		return
	}

	utils.SuccessResponse(&c.Controller, req)
}

// @router /users [get]
func (c *UserHandler) GetAllUsers() {
	logger.Info("Get all users initiated")

	users, err := svc.User.GetAllUsers(context.Background())
	if err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, "Failed to get users", err)
		return
	}

	utils.SuccessResponse(&c.Controller, users)
}

// @router /users/:id [put]
func (c *UserHandler) UpdateUser() {
	id := GetIDFromPath(&c.Controller)

	var req models.User
	if err := UnmarshalAndValidate(c.Ctx.Input.RequestBody, &req); err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusBadRequest, constants.ValidationInvalidRequestFormat, err)
		return
	}

	logger.Info("Update user initiated - user_id=%d username=%s", id, req.Username)

	updatedUser, err := svc.User.UpdateUser(context.Background(), id, &req)
	if err != nil {
		if err.Error() == "user not found" {
			utils.ErrorResponse(&c.Controller, http.StatusNotFound, "User not found", err)
			return
		}
		utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, "Failed to update user", err)
		return
	}

	utils.SuccessResponse(&c.Controller, updatedUser)
}

// @router /users/:id [delete]
func (c *UserHandler) DeleteUser() {
	id := GetIDFromPath(&c.Controller)
	logger.Info("Delete user initiated - user_id=%d", id)

	if err := svc.User.DeleteUser(context.Background(), id); err != nil {
		utils.ErrorResponse(&c.Controller, http.StatusInternalServerError, "Failed to delete user", err)
		return
	}

	utils.SuccessResponse(&c.Controller, nil)
}
