package reconciler

import (
	"context"

	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/datazip-inc/olake-ui/server/internal/constants"
	"github.com/datazip-inc/olake-ui/server/internal/models"
	"github.com/datazip-inc/olake-ui/server/internal/models/dto"
)

// sourceService is the part of the ETL service the source reconciler calls.
type sourceService interface {
	connectionService
	GetSourceByName(ctx context.Context, projectID, name string) (*models.Source, error)
	CreateSource(ctx context.Context, req *dto.CreateSourceRequest, projectID string, userID *int) error
	UpdateSource(ctx context.Context, projectID string, id int, req *dto.UpdateSourceRequest, userID *int) error
}

// Source applies a source manifest to OLake, in every GitOps mode.
type Source struct {
	ETL  sourceService
	Sink StatusSink
}

// Reconcile applies res and records its status through r.Sink.
func (r *Source) Reconcile(ctx context.Context, res *ResourceData) (ctrl.Result, error) {
	return reconcileResource(ctx, r.Sink, r, res)
}

func (r *Source) kind() string { return KindSource }

func (r *Source) notFoundErr() error { return constants.ErrSourceNotFound }

func (r *Source) parse(config string) (*resourceSpec, error) {
	req, err := ParseAndValidateSource(config)
	if err != nil {
		return nil, err
	}
	return &resourceSpec{Name: req.Name, Type: req.Type, Version: req.Version, Config: req.Config}, nil
}

func (r *Source) get(ctx context.Context, projectID, name string) (*resourceSpec, error) {
	src, err := r.ETL.GetSourceByName(ctx, projectID, name)
	if err != nil {
		return nil, err
	}
	return &resourceSpec{ID: src.ID, Name: src.Name, Type: src.Type, Version: src.Version, Config: dto.JSONConfig(src.Config)}, nil
}

func (r *Source) test(ctx context.Context, projectID string, spec *resourceSpec) error {
	return testSourceConnection(ctx, r.ETL, projectID, spec.Type, spec.Version, spec.Config)
}

func (r *Source) create(ctx context.Context, spec *resourceSpec, projectID string, userID *int) error {
	return r.ETL.CreateSource(ctx, &dto.CreateSourceRequest{
		Name:    spec.Name,
		Type:    spec.Type,
		Version: spec.Version,
		Config:  spec.Config,
	}, projectID, userID)
}

func (r *Source) update(ctx context.Context, id int, spec *resourceSpec, projectID string, userID *int) error {
	return r.ETL.UpdateSource(ctx, projectID, id, &dto.UpdateSourceRequest{
		Name:    spec.Name,
		Type:    spec.Type,
		Version: spec.Version,
		Config:  spec.Config,
	}, userID)
}
