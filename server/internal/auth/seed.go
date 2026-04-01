package auth

import "github.com/casbin/casbin/v3"

// Role constants — referenced by middleware and future management APIs.
const (
	RoleReader = "reader"
	RoleWriter = "writer"
	RoleAdmin  = "admin"
)

// defaultPolicies covers every URL depth under /api/v1/project/:projectid/
// Depth 1 → /project/:p/:r               (list, create)
// Depth 2 → /project/:p/:r/:id           (get, update, delete, test, spec)
// Depth 3 → /project/:p/:r/:id/:a        (sync, activate, tasks, cancel)
// Depth 4 → /project/:p/:r/:id/:a/:s/:sa (tasks/:taskid/logs)
//
// New routes added to routes/route.go under /project/:projectid are
// automatically covered — no changes here needed.
var defaultPolicies = [][]string{
	{RoleReader, "/api/v1/project/:p/:r", "GET"},
	{RoleWriter, "/api/v1/project/:p/:r", "(GET)|(POST)|(PUT)|(DELETE)"},

	{RoleReader, "/api/v1/project/:p/:r/:id", "GET"},
	{RoleWriter, "/api/v1/project/:p/:r/:id", "(GET)|(POST)|(PUT)|(DELETE)"},

	{RoleReader, "/api/v1/project/:p/:r/:id/:a", "GET"},
	{RoleWriter, "/api/v1/project/:p/:r/:id/:a", "(GET)|(POST)|(PUT)|(DELETE)"},

	{RoleReader, "/api/v1/project/:p/:r/:id/:a/:s/:sa", "GET"},
	{RoleWriter, "/api/v1/project/:p/:r/:id/:a/:s/:sa", "(GET)|(POST)|(PUT)|(DELETE)"},
}

// SeedDefaultRoles inserts reader/writer policies on every startup.
// Idempotent — skips rows that already exist.
func SeedDefaultRoles(e *casbin.Enforcer) error {
	for _, p := range defaultPolicies {
		exists, err := e.HasPolicy(p[0], p[1], p[2])
		if err != nil {
			return err
		}
		if !exists {
			if _, err := e.AddPolicy(p[0], p[1], p[2]); err != nil {
				return err
			}
		}
	}
	return e.SavePolicy()
}
