package database

import (
	"fmt"
	"strings"

	"github.com/casbin/casbin/v3"
	"github.com/datazip-inc/olake-ui/server/internal/auth"
	"github.com/datazip-inc/olake-ui/server/internal/models"
	"gorm.io/gorm/clause"
)

type RBACStore struct {
	db       *Database
	enforcer *casbin.Enforcer
}

func NewRBACStore(db *Database, e *casbin.Enforcer) *RBACStore {
	return &RBACStore{db: db, enforcer: e}
}

// AssignProjectRole adds a user to a project with the given role.
func (s *RBACStore) AssignProjectRole(projectID string, userID, assignedByID int, role string) error {
	if err := s.validateRole(role); err != nil {
		return err
	}
	row := models.ProjectUserRole{
		ProjectID: projectID,
		UserID:    userID,
		Role:      role,
		InvitedBy: &assignedByID,
	}
	if err := s.db.conn.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "project_id"}, {Name: "user_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"role", "invited_by"}),
	}).Create(&row).Error; err != nil {
		return err
	}
	return s.syncCasbinRole(projectID, userID, role)
}

// UpdateProjectRole changes an existing member's role.
func (s *RBACStore) UpdateProjectRole(projectID string, userID int, newRole string) error {
	if err := s.validateRole(newRole); err != nil {
		return err
	}
	if err := s.db.conn.Model(&models.ProjectUserRole{}).
		Where("project_id = ? AND user_id = ?", projectID, userID).
		Update("role", newRole).Error; err != nil {
		return err
	}
	return s.syncCasbinRole(projectID, userID, newRole)
}

// RemoveFromProject removes a user from a project.
func (s *RBACStore) RemoveFromProject(projectID string, userID int) error {
	if err := s.db.conn.
		Where("project_id = ? AND user_id = ?", projectID, userID).
		Delete(&models.ProjectUserRole{}).Error; err != nil {
		return err
	}
	subject := auth.ProjectSubject(userID, projectID)
	if _, err := s.enforcer.DeleteRolesForUser(subject); err != nil {
		return err
	}
	return s.enforcer.SavePolicy()
}

// ListProjectMembers returns all members and their roles.
func (s *RBACStore) ListProjectMembers(projectID string) ([]models.ProjectUserRole, error) {
	var members []models.ProjectUserRole
	err := s.db.conn.Where("project_id = ?", projectID).Find(&members).Error
	return members, err
}

// GetUserByEmail looks up a user by email for the invite flow.
func (s *RBACStore) GetUserByEmail(email string) (*models.User, error) {
	var user models.User
	err := s.db.conn.Where("email = ?", email).First(&user).Error
	return &user, err
}

func (s *RBACStore) syncCasbinRole(projectID string, userID int, role string) error {
	subject := auth.ProjectSubject(userID, projectID)
	if _, err := s.enforcer.DeleteRolesForUser(subject); err != nil {
		return err
	}
	if _, err := s.enforcer.AddRoleForUser(subject, role); err != nil {
		return err
	}
	return s.enforcer.SavePolicy()
}

func (s *RBACStore) validateRole(role string) error {
	if role != auth.RoleReader && role != auth.RoleWriter {
		return fmt.Errorf("invalid role %q: must be %q or %q", role, auth.RoleReader, auth.RoleWriter)
	}
	return nil
}

func splitPerm(perm string) (obj, act string, err error) {
	idx := strings.IndexByte(perm, ':')
	if idx < 0 {
		return "", "", fmt.Errorf("invalid permission %q: expected resource:action", perm)
	}
	return perm[:idx], perm[idx+1:], nil
}
