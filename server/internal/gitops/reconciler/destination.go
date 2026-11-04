package reconciler

import (
	"context"

	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/datazip-inc/olake-ui/server/internal/constants"
	"github.com/datazip-inc/olake-ui/server/internal/models"
	"github.com/datazip-inc/olake-ui/server/internal/models/dto"
)

// destinationService is the part of the ETL service the destination reconciler calls.
type destinationService interface {
	connectionService
	GetDestinationByName(ctx context.Context, projectID, name string) (*models.Destination, error)
	CreateDestination(ctx context.Context, req *dto.CreateDestinationRequest, projectID string, userID *int) error
	UpdateDestination(ctx context.Context, id int, projectID string, req *dto.UpdateDestinationRequest, userID *int) error
}

// Destination applies a destination manifest to OLake, in every GitOps mode.
type Destination struct {
	ETL     destinationService
	Sink    StatusSink
	Secrets *SecretRefs // resolves secret references in the config
}

// Reconcile applies res and records its status through r.Sink.
func (r *Destination) Reconcile(ctx context.Context, res *ResourceData) (ctrl.Result, error) {
	return reconcileResource(ctx, r.Sink, r.Secrets, r, res)
}

func (r *Destination) kind() string { return KindDestination }

func (r *Destination) notFoundErr() error { return constants.ErrDestinationNotFound }

func (r *Destination) parse(config string) (*resourceSpec, error) {
	req, err := ParseAndValidateDestination(config)
	if err != nil {
		return nil, err
	}
	return &resourceSpec{Name: req.Name, Type: req.Type, Version: req.Version, Config: req.Config}, nil
}

func (r *Destination) get(ctx context.Context, projectID, name string) (*resourceSpec, error) {
	dest, err := r.ETL.GetDestinationByName(ctx, projectID, name)
	if err != nil {
		return nil, err
	}
	return &resourceSpec{ID: dest.ID, Name: dest.Name, Type: dest.DestType, Version: dest.Version, Config: dto.JSONConfig(dest.Config)}, nil
}

func (r *Destination) test(ctx context.Context, projectID string, spec *resourceSpec) error {
	return testDestinationConnection(ctx, r.ETL, projectID, spec.Type, spec.Version, spec.Config, "", "")
}

func (r *Destination) create(ctx context.Context, spec *resourceSpec, projectID string, userID *int) error {
	return r.ETL.CreateDestination(ctx, &dto.CreateDestinationRequest{
		Name:    spec.Name,
		Type:    spec.Type,
		Version: spec.Version,
		Config:  spec.Config,
	}, projectID, userID)
}

func (r *Destination) update(ctx context.Context, id int, spec *resourceSpec, projectID string, userID *int) error {
	return r.ETL.UpdateDestination(ctx, id, projectID, &dto.UpdateDestinationRequest{
		Name:    spec.Name,
		Type:    spec.Type,
		Version: spec.Version,
		Config:  spec.Config,
	}, userID)
}
