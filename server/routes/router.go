package routes

import (
	"os"
	"strings"

	"github.com/beego/beego/v2/server/web"
	"github.com/datazip/olake-frontend/server/internal/handlers"
)

func Init() {
	isDev := strings.ToLower(os.Getenv("IS_DEV")) == "true"
	if !isDev {
		// Only set static path in production mode
		web.SetStaticPath("", "/opt/frontend/dist") // Vite assets are in /assets
	}

	// Auth routes
	web.Router("/login", &handlers.AuthHandler{}, "post:Login")
	web.Router("/signup", &handlers.AuthHandler{}, "post:Signup")
	web.Router("/auth/check", &handlers.AuthHandler{}, "get:CheckAuth")
	web.Router("/telemetry-id", &handlers.AuthHandler{}, "get:GetTelemetryID")

	// Apply auth middleware to protected routes
	web.InsertFilter("/api/v1/*", web.BeforeRouter, handlers.AuthMiddleware)

	// API routes
	// User routes
	web.Router("/api/v1/users", &handlers.UserHandler{}, "post:CreateUser")
	web.Router("/api/v1/users", &handlers.UserHandler{}, "get:GetAllUsers")
	web.Router("/api/v1/users/:id", &handlers.UserHandler{}, "put:UpdateUser")
	web.Router("/api/v1/users/:id", &handlers.UserHandler{}, "delete:DeleteUser")

	// Source routes
	web.Router("/api/v1/project/:projectid/sources", &handlers.SourceHandler{}, "get:GetAllSources")
	web.Router("/api/v1/project/:projectid/sources", &handlers.SourceHandler{}, "post:CreateSource")
	web.Router("/api/v1/project/:projectid/sources/:id", &handlers.SourceHandler{}, "put:UpdateSource")
	web.Router("/api/v1/project/:projectid/sources/:id", &handlers.SourceHandler{}, "delete:DeleteSource")
	web.Router("/api/v1/project/:projectid/sources/test", &handlers.SourceHandler{}, "post:TestConnection")
	web.Router("/api/v1/project/:projectid/sources/streams", &handlers.SourceHandler{}, "post:GetSourceCatalog")
	web.Router("/api/v1/project/:projectid/sources/versions", &handlers.SourceHandler{}, "get:GetSourceVersions")
	web.Router("/api/v1/project/:projectid/sources/spec", &handlers.SourceHandler{}, "post:GetProjectSourceSpec")

	// Destination routes
	web.Router("/api/v1/project/:projectid/destinations", &handlers.DestHandler{}, "get:GetAllDestinations")
	web.Router("/api/v1/project/:projectid/destinations", &handlers.DestHandler{}, "post:CreateDestination")
	web.Router("/api/v1/project/:projectid/destinations/:id", &handlers.DestHandler{}, "put:UpdateDestination")
	web.Router("/api/v1/project/:projectid/destinations/:id", &handlers.DestHandler{}, "delete:DeleteDestination")
	web.Router("/api/v1/project/:projectid/destinations/test", &handlers.DestHandler{}, "post:TestConnection")
	web.Router("/api/v1/project/:projectid/destinations/versions", &handlers.DestHandler{}, "get:GetDestinationVersions")
	web.Router("/api/v1/project/:projectid/destinations/spec", &handlers.DestHandler{}, "post:GetDestinationSpec")

	// Job routes
	web.Router("/api/v1/project/:projectid/jobs", &handlers.JobHandler{}, "get:GetAllJobs")
	web.Router("/api/v1/project/:projectid/jobs", &handlers.JobHandler{}, "post:CreateJob")
	web.Router("/api/v1/project/:projectid/jobs/:id", &handlers.JobHandler{}, "put:UpdateJob")
	web.Router("/api/v1/project/:projectid/jobs/:id", &handlers.JobHandler{}, "delete:DeleteJob")
	web.Router("/api/v1/project/:projectid/jobs/:id/sync", &handlers.JobHandler{}, "post:SyncJob")
	web.Router("/api/v1/project/:projectid/jobs/:id/activate", &handlers.JobHandler{}, "post:ActivateJob")
	web.Router("/api/v1/project/:projectid/jobs/:id/tasks", &handlers.JobHandler{}, "get:GetJobTasks")
	web.Router("/api/v1/project/:projectid/jobs/:id/tasks/:taskid/logs", &handlers.JobHandler{}, "post:GetTaskLogs")

	// This should be the last route - catch all for frontend
	web.Router("/*", &handlers.FrontendHandler{})
}
