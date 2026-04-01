package services

import (
	"github.com/datazip-inc/olake-ui/server/internal/database"
	"github.com/datazip-inc/olake-ui/server/internal/models"
)

type RBACService struct {
	store *database.RBACStore
}

func NewRBACService(store *database.RBACStore) *RBACService {
	return &RBACService{store: store}
}

func (s *RBACService) AssignRole(projectID string, userID, assignedByID int, role string) error {
	return s.store.AssignProjectRole(projectID, userID, assignedByID, role)
}

func (s *RBACService) UpdateRole(projectID string, userID int, role string) error {
	return s.store.UpdateProjectRole(projectID, userID, role)
}

func (s *RBACService) RemoveUser(projectID string, userID int) error {
	return s.store.RemoveFromProject(projectID, userID)
}

func (s *RBACService) ListMembers(projectID string) ([]models.ProjectUserRole, error) {
	return s.store.ListProjectMembers(projectID)
}
