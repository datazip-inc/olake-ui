package auth

import (
	"fmt"

	"github.com/casbin/casbin/v3"
	"github.com/casbin/casbin/v3/model"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"gorm.io/gorm"
)

func newModel() model.Model {
	m := model.NewModel()
	m.AddDef("r", "r", "sub, obj, act")
	m.AddDef("p", "p", "sub, obj, act")
	m.AddDef("g", "g", "_, _")
	m.AddDef("e", "e", "some(where (p.eft == allow))")
	m.AddDef("m", "m", `g(r.sub, p.sub) && keyMatch2(r.obj, p.obj) && regexMatch(r.act, p.act)`)
	return m
}

// InitCasbin wires up Casbin using the existing GORM connection.
// gorm-adapter auto-creates the casbin_rule table.
func InitCasbin(db *gorm.DB) (*casbin.Enforcer, error) {
	adapter, err := gormadapter.NewAdapterByDB(db)
	if err != nil {
		return nil, fmt.Errorf("casbin adapter: %w", err)
	}
	e, err := casbin.NewEnforcer(newModel(), adapter)
	if err != nil {
		return nil, fmt.Errorf("casbin enforcer: %w", err)
	}
	if err := e.LoadPolicy(); err != nil {
		return nil, fmt.Errorf("casbin load policy: %w", err)
	}
	return e, nil
}

// ProjectSubject returns "userID:projectID" — the per-project Casbin subject.
//
// This single key is what isolates projects from each other.
// User 42 in project "7" → subject "42:7".
// If they hit /project/9/..., middleware builds "42:9" → no role → 403.
func ProjectSubject(userID int, projectID string) string {
	return fmt.Sprintf("%d:%s", userID, projectID)
}

// AssignGlobalAdmin grants a user global admin — stored as g,"userID","admin".
// Called once for the first user after signup.
func AssignGlobalAdmin(e *casbin.Enforcer, userID int) error {
	if _, err := e.AddRoleForUser(fmt.Sprintf("%d", userID), RoleAdmin); err != nil {
		return err
	}
	return e.SavePolicy()
}

// IsGlobalAdmin checks if a user has the global admin role (not project-scoped).
func IsGlobalAdmin(e *casbin.Enforcer, userID int) bool {
	ok, _ := e.HasRoleForUser(fmt.Sprintf("%d", userID), RoleAdmin)
	return ok
}
