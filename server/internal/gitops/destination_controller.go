package gitops

import (
	"context"
	"errors"

	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/datazip-inc/olake-ui/server/internal/constants"
	"github.com/datazip-inc/olake-ui/server/internal/models"
	"github.com/datazip-inc/olake-ui/server/internal/models/dto"
	"github.com/datazip-inc/olake-ui/server/internal/services/etl"
	"github.com/datazip-inc/olake-ui/server/internal/utils"
)

type DestinationReconciler struct {
	client.Client
	ETL  *etl.Service
	Sink StatusSink
}

func (r *DestinationReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var cm corev1.ConfigMap
	if err := r.Get(ctx, req.NamespacedName, &cm); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	return r.sync(ctx, resourceFromCM(&cm))
}

func (r *DestinationReconciler) sync(ctx context.Context, res ResourceData) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	dataHash := ContentHash(res.Data)
	if skipReconcile(res.Annotations, dataHash) {
		return ctrl.Result{}, nil
	}

	createReq, err := ParseAndValidateDestination(res.Config())
	if err != nil {
		return failResource(ctx, r.Sink, res, Permanent(err))
	}
	userID, err := requireSpec(res.ProjectID(), res.UserID())
	if err != nil {
		return failResource(ctx, r.Sink, res, err)
	}

	existing, err := r.ETL.GetDestinationByName(ctx, res.ProjectID(), createReq.Name)
	if err != nil && !errors.Is(err, constants.ErrDestinationNotFound) {
		logger.Error(err, "lookup destination failed")
		return ctrl.Result{RequeueAfter: syncRequeueAfter}, nil
	}

	changed := existing != nil && !destinationMatches(existing, createReq)
	if existing == nil || changed {
		if err := testDestinationConnection(ctx, r.ETL, createReq.Type, createReq.Version, createReq.Config, "", ""); err != nil {
			logger.Error(err, "destination connection test failed")
			return failResource(ctx, r.Sink, res, Permanent(err))
		}
	}

	switch {
	case existing == nil:
		if err := r.ETL.CreateDestination(ctx, createReq, res.ProjectID(), &userID); err != nil {
			logger.Error(err, "create destination failed")
			return failResource(ctx, r.Sink, res, Permanent(err))
		}
		existing, err = r.ETL.GetDestinationByName(ctx, res.ProjectID(), createReq.Name)
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
		if err := r.ETL.UpdateDestination(ctx, existing.ID, res.ProjectID(), updateReq, &userID); err != nil {
			logger.Error(err, "update destination failed")
			return failResource(ctx, r.Sink, res, Permanent(err))
		}
	}

	if err := successResource(ctx, r.Sink, res, existing.ID); err != nil {
		logger.Error(err, "update destination status failed")
		return ctrl.Result{RequeueAfter: syncRequeueAfter}, nil
	}
	return ctrl.Result{}, nil
}

func (r *DestinationReconciler) Setup(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.ConfigMap{}).
		WithEventFilter(kindPredicate(KindDestination)).
		Complete(r)
}

func destinationMatches(existing *models.Destination, req *dto.CreateDestinationRequest) bool {
	return existing.Name == req.Name &&
		existing.DestType == req.Type &&
		existing.Version == req.Version &&
		utils.EqualJSON(existing.Config, req.Config.String())
}
