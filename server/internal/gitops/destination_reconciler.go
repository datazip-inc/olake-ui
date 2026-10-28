package gitops

import (
	"context"

	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/datazip-inc/olake-ui/server/internal/constants"
	"github.com/datazip-inc/olake-ui/server/internal/models/dto"
	"github.com/datazip-inc/olake-ui/server/internal/services/etl"
)

// DestinationReconciler applies a destination manifest to OLake.
type DestinationReconciler struct {
	ETL  *etl.Service
	Sink StatusSink
}

func (r *DestinationReconciler) reconcile(ctx context.Context, res *ResourceData) (ctrl.Result, error) {
	return reconcileResource(ctx, r.Sink, r, res)
}

func (r *DestinationReconciler) kind() string { return KindDestination }

func (r *DestinationReconciler) notFoundErr() error { return constants.ErrDestinationNotFound }

func (r *DestinationReconciler) parse(config string) (*resourceSpec, error) {
	req, err := ParseAndValidateDestination(config)
	if err != nil {
		return nil, err
	}
	return &resourceSpec{Name: req.Name, Type: req.Type, Version: req.Version, Config: req.Config}, nil
}

func (r *DestinationReconciler) get(ctx context.Context, projectID, name string) (*resourceSpec, error) {
	dest, err := r.ETL.GetDestinationByName(ctx, projectID, name)
	if err != nil {
		return nil, err
	}
	return &resourceSpec{ID: dest.ID, Name: dest.Name, Type: dest.DestType, Version: dest.Version, Config: dto.JSONConfig(dest.Config)}, nil
}

func (r *DestinationReconciler) test(ctx context.Context, projectID string, spec *resourceSpec) error {
	return testDestinationConnection(ctx, r.ETL, projectID, spec.Type, spec.Version, spec.Config, "", "")
}

func (r *DestinationReconciler) create(ctx context.Context, spec *resourceSpec, projectID string, userID *int) error {
	return r.ETL.CreateDestination(ctx, &dto.CreateDestinationRequest{
		Name:    spec.Name,
		Type:    spec.Type,
		Version: spec.Version,
		Config:  spec.Config,
	}, projectID, userID)
}

func (r *DestinationReconciler) update(ctx context.Context, id int, spec *resourceSpec, projectID string, userID *int) error {
	return r.ETL.UpdateDestination(ctx, id, projectID, &dto.UpdateDestinationRequest{
		Name:    spec.Name,
		Type:    spec.Type,
		Version: spec.Version,
		Config:  spec.Config,
	}, userID)
}
