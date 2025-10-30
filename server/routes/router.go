package routes

import (
	"net/http"

	"github.com/beego/beego/v2/server/web"
	"github.com/beego/beego/v2/server/web/context"
	"github.com/datazip-inc/olake-ui/server/internal/constants"
	"github.com/datazip-inc/olake-ui/server/internal/handlers"
	"github.com/datazip-inc/olake-ui/server/internal/middleware"
)

// writeDefaultCorsHeaders sets common CORS headers
func writeDefaultCorsHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
	w.Header().Set("Access-Control-Allow-Headers", "Origin, Authorization, Content-Type, Accept")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Max-Age", "86400")
}

// CustomCorsFilter handles CORS for different route patterns
func CustomCorsFilter(ctx *context.Context) {
	r := ctx.Request
	w := ctx.ResponseWriter
	writeDefaultCorsHeaders(w)
	requestOrigin := r.Header.Get("Origin")
	// API and auth routes - reflect origin
	w.Header().Set("Access-Control-Allow-Origin", requestOrigin)
	// Handle preflight OPTIONS requests
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
}

func Init(h *handlers.Handler) {
	if runmode, err := web.AppConfig.String(constants.ConfRunMode); err == nil && runmode == "localdev" {
		web.InsertFilter("*", web.BeforeRouter, CustomCorsFilter)
	} else {
		// Serve static frontend files
		web.SetStaticPath("", "/opt/frontend/dist") // Vite assets are in /assets

		// Serve index.html for React frontend
		web.Router("/*", h, "get:ServeFrontend") // any other frontend route
	}

	// Apply auth middleware to protected routes
	web.InsertFilter("/api/v1/*", web.BeforeRouter, middleware.AuthMiddleware)
	// Auth routes
	web.Router("/login", h, "post:Login")
	web.Router("/signup", h, "post:Signup")
	web.Router("/auth/check", h, "get:CheckAuth")
	web.Router("/telemetry-id", h, "get:GetTelemetryID")

	// User routes
	web.Router("/api/v1/users", h, "post:CreateUser")
	web.Router("/api/v1/users", h, "get:GetAllUsers")
	web.Router("/api/v1/users/:id", h, "put:UpdateUser")
	web.Router("/api/v1/users/:id", h, "delete:DeleteUser")

	// Source routes
	web.Router("/api/v1/project/:projectid/sources", h, "get:ListSources")
	web.Router("/api/v1/project/:projectid/sources", h, "post:CreateSource")
	web.Router("/api/v1/project/:projectid/sources/:id", h, "put:UpdateSource")
	web.Router("/api/v1/project/:projectid/sources/:id", h, "delete:DeleteSource")
	web.Router("/api/v1/project/:projectid/sources/test", h, "post:TestSourceConnection")
	web.Router("/api/v1/project/:projectid/sources/streams", h, "post:GetSourceCatalog")
	web.Router("/api/v1/project/:projectid/sources/versions", h, "get:GetSourceVersions")
	web.Router("/api/v1/project/:projectid/sources/spec", h, "post:GetProjectSourceSpec")

	// Destination routes
	web.Router("/api/v1/project/:projectid/destinations", h, "get:ListDestinations")
	web.Router("/api/v1/project/:projectid/destinations", h, "post:CreateDestination")
	web.Router("/api/v1/project/:projectid/destinations/:id", h, "put:UpdateDestination")
	web.Router("/api/v1/project/:projectid/destinations/:id", h, "delete:DeleteDestination")
	web.Router("/api/v1/project/:projectid/destinations/test", h, "post:TestDestinationConnection")
	web.Router("/api/v1/project/:projectid/destinations/versions", h, "get:GetDestinationVersions")
	web.Router("/api/v1/project/:projectid/destinations/spec", h, "post:GetDestinationSpec")

	// Job routes
	web.Router("/api/v1/project/:projectid/jobs", h, "get:ListJobs")
	web.Router("/api/v1/project/:projectid/jobs", h, "post:CreateJob")
	web.Router("/api/v1/project/:projectid/jobs/:id", h, "put:UpdateJob")
	web.Router("/api/v1/project/:projectid/jobs/:id", h, "delete:DeleteJob")
	web.Router("/api/v1/project/:projectid/jobs/:id/sync", h, "post:SyncJob")
	web.Router("/api/v1/project/:projectid/jobs/:id/activate", h, "post:ActivateJob")
	web.Router("/api/v1/project/:projectid/jobs/:id/tasks", h, "get:GetJobTasks")
	web.Router("/api/v1/project/:projectid/jobs/:id/cancel", h, "get:CancelJobRun")
	web.Router("/api/v1/project/:projectid/jobs/:id/tasks/:taskid/logs", h, "post:GetTaskLogs")
	web.Router("/api/v1/project/:projectid/jobs/check-unique", h, "post:CheckUniqueJobName")

	// worker callback routes
	web.Router("/internal/worker/callback/sync-telemetry", h, "post:UpdateSyncTelemetry")
}
