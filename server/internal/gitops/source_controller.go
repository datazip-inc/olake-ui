package gitops

import (
	"context"
	"errors"

	k8srecord "k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	"github.com/datazip-inc/olake-ui/server/internal/constants"
	olakev1 "github.com/datazip-inc/olake-ui/server/internal/gitops/api/v1"
	"github.com/datazip-inc/olake-ui/server/internal/models"
	"github.com/datazip-inc/olake-ui/server/internal/models/dto"
	"github.com/datazip-inc/olake-ui/server/internal/services/etl"
	"github.com/datazip-inc/olake-ui/server/internal/utils"
)

type SourceReconciler struct {
	client.Client
	ETL      *etl.Service
	Recorder k8srecord.EventRecorder
}

func (r *SourceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	sw := NewStatusWriter(r.Client, r.Recorder)

	var source olakev1.Source
	if err := r.Get(ctx, req.NamespacedName, &source); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if skipReconcile(source.Status.Phase, source.Status.ObservedGeneration, source.Generation) {
		return ctrl.Result{}, nil
	}

	_ = sw.SetStatusIfChanged(ctx, &source, olakev1.PhasePending, source.Generation, 0, "Reconciling to OLake", source.Status)

	createReq, err := ParseAndValidateSource(source.Spec.Config)
	if err != nil {
		return failCR(ctx, sw, &source, source.Generation, err)
	}
	if err := requireSpec(source.Spec.ProjectID, source.Spec.UserID); err != nil {
		return failCR(ctx, sw, &source, source.Generation, err)
	}

	existing, err := r.ETL.GetSourceByName(ctx, source.Spec.ProjectID, createReq.Name)
	if err != nil && !errors.Is(err, constants.ErrSourceNotFound) {
		logger.Error(err, "lookup source failed")
		return ctrl.Result{RequeueAfter: syncRequeueAfter}, nil
	}

	changed := existing != nil && !sourceMatches(existing, createReq)
	if existing == nil || changed {
		if err := testSourceConnection(ctx, r.ETL, createReq.Type, createReq.Version, createReq.Config); err != nil {
			logger.Error(err, "source connection test failed")
			return failCR(ctx, sw, &source, source.Generation, err)
		}
	}

	switch {
	case existing == nil:
		if err := r.ETL.CreateSource(ctx, createReq, source.Spec.ProjectID, &source.Spec.UserID); err != nil {
			logger.Error(err, "create source failed")
			return failCR(ctx, sw, &source, source.Generation, err)
		}
		existing, err = r.ETL.GetSourceByName(ctx, source.Spec.ProjectID, createReq.Name)
		if err != nil {
			logger.Error(err, "reload source after create failed")
			return ctrl.Result{RequeueAfter: syncRequeueAfter}, nil
		}
	case changed:
		updateReq := &dto.UpdateSourceRequest{
			Name:    createReq.Name,
			Type:    createReq.Type,
			Version: createReq.Version,
			Config:  createReq.Config,
		}
		if err := r.ETL.UpdateSource(ctx, source.Spec.ProjectID, existing.ID, updateReq, &source.Spec.UserID); err != nil {
			logger.Error(err, "update source failed")
			return failCR(ctx, sw, &source, source.Generation, err)
		}
	}

	if err := sw.SetStatus(ctx, &source, olakev1.PhaseReady, source.Generation, existing.ID, ""); err != nil {
		logger.Error(err, "update source status failed")
		return ctrl.Result{RequeueAfter: syncRequeueAfter}, nil
	}
	return ctrl.Result{}, nil
}

func (r *SourceReconciler) Setup(mgr ctrl.Manager) error {
	r.Recorder = mgr.GetEventRecorderFor("olake-source-controller")
	return ctrl.NewControllerManagedBy(mgr).
		For(&olakev1.Source{}).
		WithEventFilter(predicate.GenerationChangedPredicate{}).
		Complete(r)
}

func sourceMatches(existing *models.Source, req *dto.CreateSourceRequest) bool {
	return existing.Name == req.Name &&
		existing.Type == req.Type &&
		existing.Version == req.Version &&
		utils.EqualJSON(existing.Config, req.Config.String())
}
