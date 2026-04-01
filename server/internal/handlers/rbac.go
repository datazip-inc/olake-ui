package handlers

import (
	"net/http"
	"strconv"

	"github.com/datazip-inc/olake-ui/server/internal/auth"
	"github.com/datazip-inc/olake-ui/server/internal/constants"
	"github.com/datazip-inc/olake-ui/server/internal/models/dto"
	"github.com/gin-gonic/gin"
)

// requireAdmin is a guard used by all member-management handlers.
func (h *Handler) requireAdmin(c *gin.Context) (int, bool) {
	userIDVal, _ := c.Get(constants.UserIDContextKey)
	userID := userIDVal.(int)
	if !auth.IsGlobalAdmin(h.enforcer, userID) {
		errorResponse(c, http.StatusForbidden, "only admin can manage project members", nil)
		return 0, false
	}
	return userID, true
}

// AssignRole godoc
// POST /api/v1/project/:projectid/members
// Admin assigns reader or writer role to a user in the project.
func (h *Handler) AssignRole(c *gin.Context) {
	adminID, ok := h.requireAdmin(c)
	if !ok {
		return
	}
	projectID, err := getProjectID(c)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error(), err)
		return
	}
	var req dto.AssignRoleRequest
	if err := bindAndValidate(c, &req); err != nil {
		errorResponse(c, http.StatusBadRequest, constants.ValidationInvalidRequestFormat, err)
		return
	}
	if err := h.etl.RBAC.AssignRole(projectID, req.UserID, adminID, req.Role); err != nil {
		errorResponse(c, http.StatusInternalServerError, err.Error(), err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "message": "role assigned"})
}

// UpdateMemberRole godoc
// PUT /api/v1/project/:projectid/members/:userid
// Admin updates the role of an existing project member.
func (h *Handler) UpdateMemberRole(c *gin.Context) {
	if _, ok := h.requireAdmin(c); !ok {
		return
	}
	projectID, err := getProjectID(c)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error(), err)
		return
	}
	userID, err := strconv.Atoi(c.Param("userid"))
	if err != nil {
		errorResponse(c, http.StatusBadRequest, "invalid user id", err)
		return
	}
	var req dto.UpdateRoleRequest
	if err := bindAndValidate(c, &req); err != nil {
		errorResponse(c, http.StatusBadRequest, constants.ValidationInvalidRequestFormat, err)
		return
	}
	if err := h.etl.RBAC.UpdateRole(projectID, userID, req.Role); err != nil {
		errorResponse(c, http.StatusInternalServerError, err.Error(), err)
		return
	}
	successResponse(c, "role updated", nil)
}

// RemoveMember godoc
// DELETE /api/v1/project/:projectid/members/:userid
// Admin removes a user from the project.
func (h *Handler) RemoveMember(c *gin.Context) {
	if _, ok := h.requireAdmin(c); !ok {
		return
	}
	projectID, err := getProjectID(c)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error(), err)
		return
	}
	userID, err := strconv.Atoi(c.Param("userid"))
	if err != nil {
		errorResponse(c, http.StatusBadRequest, "invalid user id", err)
		return
	}
	if err := h.etl.RBAC.RemoveUser(projectID, userID); err != nil {
		errorResponse(c, http.StatusInternalServerError, err.Error(), err)
		return
	}
	successResponse(c, "user removed from project", nil)
}

// ListMembers godoc
// GET /api/v1/project/:projectid/members
// Admin lists all users and their roles in the project.
func (h *Handler) ListMembers(c *gin.Context) {
	if _, ok := h.requireAdmin(c); !ok {
		return
	}
	projectID, err := getProjectID(c)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error(), err)
		return
	}
	members, err := h.etl.RBAC.ListMembers(projectID)
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, err.Error(), err)
		return
	}
	successResponse(c, "members fetched", members)
}
