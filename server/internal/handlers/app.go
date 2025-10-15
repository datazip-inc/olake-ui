package handlers

import "github.com/datazip/olake-ui/server/internal/services"

type AppDeps struct {
	SourceService      *services.SourceService
	JobService         *services.JobService
	DestinationService *services.DestinationService
	UserService        *services.UserService
	AuthService        *services.AuthService
}

var app AppDeps

func InitApp(d AppDeps) {
	app = d
}

func SourceSvc() *services.SourceService {
	return app.SourceService
}

func JobSvc() *services.JobService {
	return app.JobService
}

func DestSvc() *services.DestinationService {
	return app.DestinationService
}

func UserSvc() *services.UserService {
	return app.UserService
}

func AuthSvc() *services.AuthService {
	return app.AuthService
}
