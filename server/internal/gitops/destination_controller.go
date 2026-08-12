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

type DestinationReconciler struct {
	client.Client
	ETL      *etl.Service
	Recorder k8srecord.EventRecorder
}

func (r *DestinationReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	sw := NewStatusWriter(r.Client, r.Recorder)

	var destination olakev1.Destination
	if err := r.Get(ctx, req.NamespacedName, &destination); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if skipReconcile(destination.Status.Phase, destination.Status.ObservedGeneration, destination.Generation) {
		return ctrl.Result{}, nil
	}

	_ = sw.SetStatusIfChanged(ctx, &destination, olakev1.PhasePending, destination.Generation, 0, "Reconciling to OLake", destination.Status)

	createReq, err := ParseAndValidateDestination(destination.Spec.Config)
	if err != nil {
		return failCR(ctx, sw, &destination, destination.Generation, err)
	}
	if err := requireSpec(destination.Spec.ProjectID, destination.Spec.UserID); err != nil {
		return failCR(ctx, sw, &destination, destination.Generation, err)
	}

	existing, err := r.ETL.GetDestinationByName(ctx, destination.Spec.ProjectID, createReq.Name)
	if err != nil && !errors.Is(err, constants.ErrDestinationNotFound) {
		logger.Error(err, "lookup destination failed")
		return ctrl.Result{RequeueAfter: syncRequeueAfter}, nil
	}

	changed := existing != nil && !destinationMatches(existing, createReq)
	if existing == nil || changed {
		if err := testDestinationConnection(ctx, r.ETL, createReq.Type, createReq.Version, createReq.Config, "", ""); err != nil {
			logger.Error(err, "destination connection test failed")
			return failCR(ctx, sw, &destination, destination.Generation, err)
		}
	}

	switch {
	case existing == nil:
		if err := r.ETL.CreateDestination(ctx, createReq, destination.Spec.ProjectID, &destination.Spec.UserID); err != nil {
			logger.Error(err, "create destination failed")
			return failCR(ctx, sw, &destination, destination.Generation, err)
		}
		existing, err = r.ETL.GetDestinationByName(ctx, destination.Spec.ProjectID, createReq.Name)
		if err != nil {
			logger.Error(err, "reload destination after create failed")
			return ctrl.Result{RequeueAfter: syncRequeueAfter}, nil
		}
	case changed:
		updateReq := &dto.UpdateDestinationRequest{
			Name:    createReq.Name,
			Type:    createReq.Type,
			Version: createReq.Version,
			Config:  createReq.Config,
		}
		if err := r.ETL.UpdateDestination(ctx, existing.ID, destination.Spec.ProjectID, updateReq, &destination.Spec.UserID); err != nil {
			logger.Error(err, "update destination failed")
			return failCR(ctx, sw, &destination, destination.Generation, err)
		}
	}

	if err := sw.SetStatus(ctx, &destination, olakev1.PhaseReady, destination.Generation, existing.ID, ""); err != nil {
		logger.Error(err, "update destination status failed")
		return ctrl.Result{RequeueAfter: syncRequeueAfter}, nil
	}
	return ctrl.Result{}, nil
}

func (r *DestinationReconciler) Setup(mgr ctrl.Manager) error {
	r.Recorder = mgr.GetEventRecorderFor("olake-destination-controller")
	return ctrl.NewControllerManagedBy(mgr).
		For(&olakev1.Destination{}).
		WithEventFilter(predicate.GenerationChangedPredicate{}).
		Complete(r)
}

func destinationMatches(existing *models.Destination, req *dto.CreateDestinationRequest) bool {
	return existing.Name == req.Name &&
		existing.DestType == req.Type &&
		existing.Version == req.Version &&
		utils.EqualJSON(existing.Config, req.Config.String())
}
