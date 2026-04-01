package routes

import (
	"github.com/datazip-inc/olake-ui/server/internal/handlers"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(engine *gin.Engine, h *handlers.Handler, sessionsEnabled bool) {
	engine.POST("/login", h.Login)
	engine.POST("/signup", h.Signup)
	engine.GET("/auth/check", h.CheckAuth)
	engine.GET("/telemetry-id", h.TelemetryID)
	engine.GET("/swagger/*any", h.ServeSwagger)

	v1 := engine.Group("/api/v1")
	v1.Use(h.AuthMiddleware())

	// Non-project routes — no RBAC
	v1.POST("/users", h.CreateUser)
	v1.GET("/users", h.GetAllUsers)
	v1.PUT("/users/:id", h.UpdateUser)
	v1.DELETE("/users/:id", h.DeleteUser)
	v1.GET("/platform/releases", h.GetReleaseUpdates)

	// Project-scoped routes — RBAC enforced via Casbin
	project := v1.Group("/project/:projectid")
	project.Use(handlers.NewRBACMiddleware(h.Enforcer(), sessionsEnabled))

	// Member management — admin only (enforced in handler via requireAdmin)
	project.POST("/members", h.AssignRole)
	project.GET("/members", h.ListMembers)
	project.PUT("/members/:userid", h.UpdateMemberRole)
	project.DELETE("/members/:userid", h.RemoveMember)

	project.GET("/sources", h.ListSources)
	project.POST("/sources", h.CreateSource)
	project.GET("/sources/:id", h.GetSource)
	project.PUT("/sources/:id", h.UpdateSource)
	project.DELETE("/sources/:id", h.DeleteSource)
	project.POST("/sources/test", h.TestSourceConnection)
	project.POST("/sources/streams", h.GetSourceCatalog)
	project.GET("/sources/versions", h.GetSourceVersions)
	project.POST("/sources/spec", h.GetSourceSpec)

	project.GET("/destinations", h.ListDestinations)
	project.POST("/destinations", h.CreateDestination)
	project.GET("/destinations/:id", h.GetDestination)
	project.PUT("/destinations/:id", h.UpdateDestination)
	project.DELETE("/destinations/:id", h.DeleteDestination)
	project.POST("/destinations/test", h.TestDestinationConnection)
	project.GET("/destinations/versions", h.GetDestinationVersions)
	project.POST("/destinations/spec", h.GetDestinationSpec)

	project.GET("/jobs", h.ListJobs)
	project.POST("/jobs", h.CreateJob)
	project.GET("/jobs/:id", h.GetJob)
	project.PUT("/jobs/:id", h.UpdateJob)
	project.DELETE("/jobs/:id", h.DeleteJob)
	project.POST("/jobs/:id/sync", h.SyncJob)
	project.POST("/jobs/:id/activate", h.ActivateJob)
	project.GET("/jobs/:id/tasks", h.GetJobTasks)
	project.GET("/jobs/:id/cancel", h.CancelJobRun)
	project.POST("/jobs/:id/tasks/:taskid/logs", h.GetTaskLogs)
	project.GET("/jobs/:id/logs/download", h.DownloadTaskLogs)
	project.POST("/jobs/:id/clear-destination", h.ClearDestination)
	project.GET("/jobs/:id/clear-destination", h.GetClearDestinationStatus)
	project.POST("/jobs/:id/stream-difference", h.GetStreamDifference)

	project.PUT("/settings", h.UpsertProjectSettings)
	project.GET("/settings", h.GetProjectSettings)
	project.POST("/check-unique", h.CheckUniqueName)

	// Internal worker routes — no auth
	engine.POST("/internal/worker/callback/sync-telemetry", h.UpdateSyncTelemetry)
	engine.POST("/internal/project/:projectid/jobs/:id/clear-destination/recover", h.RecoverClearDestination)
	engine.PUT("/internal/project/:projectid/jobs/:id/statefile", h.UpdateStateFile)
}
