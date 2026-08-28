package gitops

import (
	"context"
	"errors"

	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
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
	res, err := fetchManaged(ctx, r.Client, req.NamespacedName)
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	return r.sync(ctx, res)
}

func (r *DestinationReconciler) sync(ctx context.Context, res *ResourceData) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	observedHash := ContentHash(res.Data)
	if skipReconcile(res.Annotations, observedHash) {
		return ctrl.Result{}, nil
	}

	createReq, err := ParseAndValidateDestination(res.Config())
	if err != nil {
		return failResource(ctx, r.Sink, res, NonRetryableError(err), observedHash)
	}
	userID, err := requireSpec(res.ProjectID(), res.UserID())
	if err != nil {
		return failResource(ctx, r.Sink, res, err, observedHash)
	}

	existing, err := r.ETL.GetDestinationByName(ctx, res.ProjectID(), createReq.Name)
	if err != nil && !errors.Is(err, constants.ErrDestinationNotFound) {
		logger.Error(err, "lookup destination failed")
		return requeueTransient(ctx, r.Sink, res, err, observedHash)
	}

	changed := existing != nil && !destinationMatches(existing, createReq)
	if existing == nil || changed {
		if err := testDestinationConnection(ctx, r.ETL, createReq.Type, createReq.Version, createReq.Config, "", ""); err != nil {
			logger.Error(err, "destination connection test failed")
			return failResource(ctx, r.Sink, res, NonRetryableError(err), observedHash)
		}
	}

	switch {
	case existing == nil:
		if err := r.ETL.CreateDestination(ctx, createReq, res.ProjectID(), &userID); err != nil {
			logger.Error(err, "create destination failed")
			return failResource(ctx, r.Sink, res, NonRetryableError(err), observedHash)
		}
		existing, err = r.ETL.GetDestinationByName(ctx, res.ProjectID(), createReq.Name)
		if err != nil {
			logger.Error(err, "reload destination after create failed")
			return requeueTransient(ctx, r.Sink, res, err, observedHash)
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
			return failResource(ctx, r.Sink, res, NonRetryableError(err), observedHash)
		}
	}

	if err := successResource(ctx, r.Sink, res, existing.ID, observedHash); err != nil {
		logger.Error(err, "update destination status failed")
		return requeueTransient(ctx, r.Sink, res, err, observedHash)
	}
	return ctrl.Result{}, nil
}

func (r *DestinationReconciler) Setup(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		Named("gitops-destination").
		For(&corev1.ConfigMap{}).
		Watches(&corev1.Secret{}, handler.EnqueueRequestsFromMapFunc(identityEnqueue)).
		WithEventFilter(kindPredicate(KindDestination)).
		Complete(r)
}

func destinationMatches(existing *models.Destination, req *dto.CreateDestinationRequest) bool {
	return existing.Name == req.Name &&
		existing.DestType == req.Type &&
		existing.Version == req.Version &&
		utils.EqualJSON(existing.Config, req.Config.String())
}
