// handlers/app.go
package handlers

import "github.com/datazip/olake-ui/server/internal/services"

var svc *services.Services

func InitHandlers(s *services.Services) {
	svc = s
}
