package routers

import (
	"github.com/beego/beego/v2/server/web"
	"github.com/beego/beego/v2/server/web/filter/cors"
	"github.com/datazip/olake-server/internal/handlers"
)

func Init() {
	// auth routes
	web.Router("/login", &controllers.AuthController{}, "post:Login")
	web.Router("/signup", &controllers.AuthController{}, "post:Signup")
	web.Router("/auth/check", &controllers.AuthController{}, "get:CheckAuth")

	// Then CORS
	web.InsertFilter("*", web.BeforeRouter, cors.Allow(&cors.Options{
		AllowOrigins:     []string{"http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"*"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// Then auth middleware
	web.InsertFilter("/api/v1/*", web.BeforeRouter, controllers.AuthMiddleware)

	// User routes
	web.Router("/api/v1/users", &controllers.UserController{}, "post:CreateUser")
	web.Router("/api/v1/users", &controllers.UserController{}, "get:GetAllUsers")
	web.Router("/api/v1/users/:id", &controllers.UserController{}, "put:UpdateUser")
	web.Router("/api/v1/users/:id", &controllers.UserController{}, "delete:DeleteUser")
	// // Source routes
	// web.Router("/api/v1/sources", &controllers.SourceController{}, "post:CreateSource")
	// web.Router("/api/v1/sources", &controllers.SourceController{}, "get:GetAllSources")
	// web.Router("/api/v1/sources/:id", &controllers.SourceController{}, "put:UpdateSource")
	// web.Router("/api/v1/sources/:id", &controllers.SourceController{}, "delete:DeleteSource")

	// // Destination routes
	// web.Router("/api/v1/destinations", &controllers.DestinationController{}, "post:CreateDestination")
	// web.Router("/api/v1/destinations", &controllers.DestinationController{}, "get:GetAllDestinations")
	// web.Router("/api/v1/destinations/:id", &controllers.DestinationController{}, "put:UpdateDestination")
	// web.Router("/api/v1/destinations/:id", &controllers.DestinationController{}, "delete:DeleteDestination")

	// // Job routes
	// web.Router("/api/v1/jobs", &controllers.JobController{}, "post:CreateJob")
	// web.Router("/api/v1/jobs", &controllers.JobController{}, "get:GetAllJobs")
	// web.Router("/api/v1/jobs/:id", &controllers.JobController{}, "put:UpdateJob")
	// web.Router("/api/v1/jobs/:id", &controllers.JobController{}, "delete:DeleteJob")
}
