package gitops

import (
	"context"

	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/datazip-inc/olake-ui/server/internal/constants"
	"github.com/datazip-inc/olake-ui/server/internal/models/dto"
	"github.com/datazip-inc/olake-ui/server/internal/services/etl"
)

// SourceReconciler applies a source manifest to OLake, in every GitOps mode.
type SourceReconciler struct {
	ETL     *etl.Service
	Sink    StatusSink
	Secrets *secretRefs
}

func (r *SourceReconciler) reconcile(ctx context.Context, res *ResourceData) (ctrl.Result, error) {
	return reconcileResource(ctx, r.Sink, r.Secrets, r, res)
}

func (r *SourceReconciler) kind() string { return KindSource }

func (r *SourceReconciler) notFoundErr() error { return constants.ErrSourceNotFound }

func (r *SourceReconciler) parse(config string) (*resourceSpec, error) {
	req, err := ParseAndValidateSource(config)
	if err != nil {
		return nil, err
	}
	return &resourceSpec{Name: req.Name, Type: req.Type, Version: req.Version, Config: req.Config}, nil
}

func (r *SourceReconciler) get(ctx context.Context, projectID, name string) (*resourceSpec, error) {
	src, err := r.ETL.GetSourceByName(ctx, projectID, name)
	if err != nil {
		return nil, err
	}
	return &resourceSpec{ID: src.ID, Name: src.Name, Type: src.Type, Version: src.Version, Config: dto.JSONConfig(src.Config)}, nil
}

func (r *SourceReconciler) test(ctx context.Context, projectID string, spec *resourceSpec) error {
	return testSourceConnection(ctx, r.ETL, projectID, spec.Type, spec.Version, spec.Config)
}

func (r *SourceReconciler) create(ctx context.Context, spec *resourceSpec, projectID string, userID *int) error {
	return r.ETL.CreateSource(ctx, &dto.CreateSourceRequest{
		Name:    spec.Name,
		Type:    spec.Type,
		Version: spec.Version,
		Config:  spec.Config,
	}, projectID, userID)
}

func (r *SourceReconciler) update(ctx context.Context, id int, spec *resourceSpec, projectID string, userID *int) error {
	return r.ETL.UpdateSource(ctx, projectID, id, &dto.UpdateSourceRequest{
		Name:    spec.Name,
		Type:    spec.Type,
		Version: spec.Version,
		Config:  spec.Config,
	}, userID)
}
