package routes

import (
	"net/http"
	"strings"

	"github.com/beego/beego/v2/server/web"
	"github.com/beego/beego/v2/server/web/context"
	"github.com/datazip/olake-server/internal/handlers"
)

// Helper function to write default CORS headers
func writeDefaultCorsHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
	w.Header().Set("Access-Control-Allow-Headers", "Origin, Authorization, Content-Type, Accept")
	w.Header().Set("Access-Control-Expose-Headers", "Content-Length")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Max-Age", "86400") // 24 hours
}

// Helper function to extract token from request
func extractToken(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		// Typically "Bearer {token}"
		parts := strings.Split(authHeader, " ")
		if len(parts) == 2 {
			return parts[1]
		}
	}
	// Check for token in query parameters
	return r.URL.Query().Get("token")
}

// ErrResponse creates a standardized error response
func ErrResponse(message string, data interface{}) map[string]interface{} {
	return map[string]interface{}{
		"success": false,
		"message": message,
		"data":    data,
	}
}

// Custom CORS filter for Beego
func CustomCorsFilter(ctx *context.Context) {
	r := ctx.Request
	w := ctx.ResponseWriter

	// Set common CORS headers that apply to most responses
	writeDefaultCorsHeaders(w) // This sets Allow-Methods, Allow-Headers, Credentials, Max-Age etc.

	requestOrigin := r.Header.Get("Origin")

	// Handle specific paths that need dynamic origin reflection or wildcard
	// These include your auth paths and API paths.
	if strings.HasPrefix(r.URL.Path, "/api/v1/") ||
		r.URL.Path == "/login" ||
		r.URL.Path == "/signup" ||
		r.URL.Path == "/auth/check" {

		if requestOrigin != "" {
			// For these paths, reflect the request origin.
			// IMPORTANT: For production, you should validate `requestOrigin` against a list of
			// explicitly allowed origins from your environment variable (CORS_ALLOWED_ORIGINS)
			// instead of just reflecting any origin.
			// Example validation (conceptual, assuming `allowedOrigins` is populated from ENV):
			// isAllowed := false
			// for _, allowed := range allowedOrigins { // allowedOrigins should be populated from os.Getenv("CORS_ALLOWED_ORIGINS")
			//   if requestOrigin == allowed {
			//     isAllowed = true
			//     break
			//   }
			// }
			// if isAllowed {
			//   w.Header().Set("Access-Control-Allow-Origin", requestOrigin)
			// } else {
			//   // Origin not allowed, you might choose to not set ACAO or handle differently
			// }
			w.Header().Set("Access-Control-Allow-Origin", requestOrigin) // Simplified for now, directly reflects
		} else {
			// No Origin header present, usually for same-origin or server-to-server.
			// No ACAO needed, or could set to a default if your app logic expects one.
		}

		// Handle preflight OPTIONS requests for these paths
		if r.Method == http.MethodOptions {
			// Preflight request already has ACAO set (if origin was present),
			// and other headers from writeDefaultCorsHeaders.
			// Just return HTTP 200 OK.
			w.WriteHeader(http.StatusOK)
			return
		}

		// For actual requests (GET, POST etc.) on /api/v1/, you might have token logic
		if strings.HasPrefix(r.URL.Path, "/api/v1/") {
			_ = extractToken(r) // Assuming token extraction is relevant here
			// Further token validation might be handled by AuthMiddleware later
		}

	} else if strings.Contains(r.URL.Path, "/p/") ||
		strings.Contains(r.URL.Path, "/s/") ||
		strings.Contains(r.URL.Path, "/t/") {
		// Public endpoints that can allow any origin
		w.Header().Set("Access-Control-Allow-Origin", "*")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
	}
	// For any other paths not explicitly handled by the above conditions,
	// no specific Access-Control-Allow-Origin is set by this filter.
	// They will rely on default browser behavior (typically allowed for same-origin,
	// blocked for cross-origin unless specific headers are set by other means).
}

func Init() {
	// auth routes
	web.Router("/login", &handlers.AuthHandler{}, "post:Login")
	web.Router("/signup", &handlers.AuthHandler{}, "post:Signup")
	web.Router("/auth/check", &handlers.AuthHandler{}, "get:CheckAuth")

	// Apply custom CORS filter before router
	web.InsertFilter("*", web.BeforeRouter, CustomCorsFilter)

	// Then auth middleware
	web.InsertFilter("/api/v1/*", web.BeforeRouter, handlers.AuthMiddleware)

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
	web.Router("/api/v1/project/:projectid/jobs/:id/streams", &handlers.JobHandler{}, "get:GetJobStreams")
	web.Router("/api/v1/project/:projectid/jobs/:id/sync", &handlers.JobHandler{}, "post:SyncJob")
	web.Router("/api/v1/project/:projectid/jobs/:id/activate", &handlers.JobHandler{}, "post:ActivateJob")
	web.Router("/api/v1/project/:projectid/jobs/:id/tasks", &handlers.JobHandler{}, "get:GetJobTasks")
	web.Router("/api/v1/project/:projectid/jobs/:id/tasks/:taskid/logs", &handlers.JobHandler{}, "post:GetTaskLogs")
}
