package routes

import (
	"github.com/beego/beego/v2/server/web"
	"github.com/beego/beego/v2/server/web/filter/cors"
	"github.com/datazip/olake-server/internal/handlers"
)

func Init() {
	// auth routes
	web.Router("/login", &handlers.AuthHandler{}, "post:Login")
	web.Router("/signup", &handlers.AuthHandler{}, "post:Signup")
	web.Router("/auth/check", &handlers.AuthHandler{}, "get:CheckAuth")

	// Then CORS
	web.InsertFilter("*", web.BeforeRouter, cors.Allow(&cors.Options{
		AllowOrigins:     []string{"http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"*"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// Then auth middleware
	web.InsertFilter("/api/v1/*", web.BeforeRouter, handlers.AuthMiddleware)

	// User routes
	web.Router("/api/v1/users", &handlers.UserHandler{}, "post:CreateUser")
	web.Router("/api/v1/users", &handlers.UserHandler{}, "get:GetAllUsers")
	web.Router("/api/v1/users/:id", &handlers.UserHandler{}, "put:UpdateUser")
	web.Router("/api/v1/users/:id", &handlers.UserHandler{}, "delete:DeleteUser")

	// Source routes
	web.Router("/api/v1/sources", &handlers.SourceHandler{}, "get:GetAllSources")
	web.Router("/api/v1/sources", &handlers.SourceHandler{}, "post:CreateSource")
	web.Router("/api/v1/sources/:id", &handlers.SourceHandler{}, "put:UpdateSource")
	web.Router("/api/v1/sources/:id", &handlers.SourceHandler{}, "delete:DeleteSource")

	// Destination routes
	web.Router("/api/v1/destinations", &handlers.DestHandler{}, "get:GetAllDestinations")
	web.Router("/api/v1/destinations", &handlers.DestHandler{}, "post:CreateDestination")
	web.Router("/api/v1/destinations/:id", &handlers.DestHandler{}, "put:UpdateDestination")
	web.Router("/api/v1/destinations/:id", &handlers.DestHandler{}, "delete:DeleteDestination")

	// Job routes
	web.Router("/api/v1/jobs", &handlers.JobHandler{}, "get:GetAllJobs")
	web.Router("/api/v1/jobs", &handlers.JobHandler{}, "post:CreateJob")
	web.Router("/api/v1/jobs/:id", &handlers.JobHandler{}, "put:UpdateJob")
	web.Router("/api/v1/jobs/:id", &handlers.JobHandler{}, "delete:DeleteJob")
}
