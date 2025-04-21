package routers

import (
	"olake_backend/controllers"
	// "olake_backend/middleware"

	"github.com/beego/beego/v2/server/web"
	"github.com/beego/beego/v2/server/web/filter/cors"
)

func init() {
	// Public routes must be first
	web.Router("/login", &controllers.LoginController{}, "post:Login")
	web.Router("/signup", &controllers.UserController{}, "post:Signup")
	web.Router("/auth/check", &controllers.LoginController{}, "get:CheckAuth")

	// Then CORS
	web.InsertFilter("*", web.BeforeRouter, cors.Allow(&cors.Options{
		AllowOrigins:     []string{"http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// Then auth middleware
	// web.InsertFilter("/*", web.BeforeRouter, middleware.AuthMiddleware)
	
	// Source routes
	web.Router("/sources", &controllers.SourceController{}, "post:CreateSource")
	web.Router("/sources", &controllers.SourceController{}, "get:GetAllSources")
	web.Router("/sources/:id", &controllers.SourceController{}, "put:UpdateSource")
	web.Router("/sources/:id", &controllers.SourceController{}, "delete:DeleteSource")

	// Destination routes
	web.Router("/destinations", &controllers.DestinationController{}, "post:CreateDestination")
	web.Router("/destinations", &controllers.DestinationController{}, "get:GetAllDestinations")
	web.Router("/destinations/:id", &controllers.DestinationController{}, "put:UpdateDestination")
	web.Router("/destinations/:id", &controllers.DestinationController{}, "delete:DeleteDestination")

	// Job routes
	web.Router("/jobs", &controllers.JobController{}, "post:CreateJob")
	web.Router("/jobs", &controllers.JobController{}, "get:GetAllJobs")
	web.Router("/jobs/:id", &controllers.JobController{}, "put:UpdateJob")
	web.Router("/jobs/:id", &controllers.JobController{}, "delete:DeleteJob")

	// User routes
	web.Router("/users", &controllers.UserController{}, "post:CreateUser")
	web.Router("/users", &controllers.UserController{}, "get:GetAllUsers")
	web.Router("/users/:id", &controllers.UserController{}, "put:UpdateUser")
	web.Router("/users/:id", &controllers.UserController{}, "delete:DeleteUser")

	// Protected routes last
	web.Router("/data", &controllers.MainController{})
	// web.Router("/api/sources/insert-test", &controllers.SourceController{}, "post:InsertTestData")
}
