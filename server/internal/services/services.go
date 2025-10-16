// services/services.go
package services

import "fmt"

type Services struct {
	Source      *SourceService
	Job         *JobService
	Destination *DestinationService
	User        *UserService
	Auth        *AuthService
}

func InitServices() (*Services, error) {
	sourceService, err := NewSourceService()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize SourceService - error=%v", err)
	}

	jobService, err := NewJobService()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize JobService - error=%v", err)
	}

	destinationService, err := NewDestinationService()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize DestinationService - error=%v", err)
	}

	userService := NewUserService()
	authService := NewAuthService()

	return &Services{
		Source:      sourceService,
		Job:         jobService,
		Destination: destinationService,
		User:        userService,
		Auth:        authService,
	}, nil
}
