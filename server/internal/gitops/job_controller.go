package gitops

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	k8srecord "k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/datazip-inc/olake-ui/server/internal/constants"
	olakev1 "github.com/datazip-inc/olake-ui/server/internal/gitops/api/v1"
	"github.com/datazip-inc/olake-ui/server/internal/models"
	"github.com/datazip-inc/olake-ui/server/internal/models/dto"
	"github.com/datazip-inc/olake-ui/server/internal/services/etl"
	"github.com/datazip-inc/olake-ui/server/internal/utils"
)

type JobReconciler struct {
	client.Client
	ETL      *etl.Service
	Recorder k8srecord.EventRecorder
}

func (r *JobReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	sw := NewStatusWriter(r.Client, r.Recorder)

	var jobCR olakev1.Job
	if err := r.Get(ctx, req.NamespacedName, &jobCR); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Job generation does not bump when a linked Streams CR changes, so Ready
	// jobs still need a streams-drift check. Failed jobs are terminal.
	if skipReconcile(jobCR.Status.Phase, jobCR.Status.ObservedGeneration, jobCR.Generation) {
		if jobCR.Status.Phase == olakev1.PhaseFailed {
			return ctrl.Result{}, nil
		}
		settled, err := r.streamsSettled(ctx, &jobCR)
		if err != nil {
			logger.Error(err, "check streams drift failed")
		} else if settled {
			return ctrl.Result{}, nil
		}
	}

	jobCfg, err := ParseAndValidateJobConfig(jobCR.Spec.Config)
	if err != nil {
		return r.fail(ctx, sw, &jobCR, nil, err)
	}
	if err := requireSpec(jobCR.Spec.ProjectID, jobCR.Spec.UserID); err != nil {
		return r.fail(ctx, sw, &jobCR, nil, err)
	}
	projectID := jobCR.Spec.ProjectID
	userID := jobCR.Spec.UserID

	source, err := r.resolveSource(ctx, projectID, jobCfg.Source)
	if errors.Is(err, constants.ErrSourceNotFound) {
		return r.wait(ctx, sw, &jobCR, fmt.Sprintf("waiting for source %q", jobCfg.Source))
	}
	if err != nil {
		logger.Error(err, "lookup source failed")
		return ctrl.Result{RequeueAfter: syncRequeueAfter}, nil
	}

	dest, err := r.resolveDestination(ctx, projectID, jobCfg.Destination)
	if errors.Is(err, constants.ErrDestinationNotFound) {
		return r.wait(ctx, sw, &jobCR, fmt.Sprintf("waiting for destination %q", jobCfg.Destination))
	}
	if err != nil {
		logger.Error(err, "lookup destination failed")
		return ctrl.Result{RequeueAfter: syncRequeueAfter}, nil
	}

	streamsCR, err := r.findStreamsForJob(ctx, &jobCR, projectID)
	if err != nil {
		logger.Error(err, "find streams CR failed")
		return ctrl.Result{RequeueAfter: syncRequeueAfter}, nil
	}
	if streamsCR == nil {
		return r.wait(ctx, sw, &jobCR, fmt.Sprintf("waiting for Streams CR referencing job %q", jobCR.Name))
	}

	existingJob, err := r.ETL.GetJobByName(ctx, projectID, jobCfg.Name)
	if err != nil && !errors.Is(err, constants.ErrJobNotFound) {
		logger.Error(err, "lookup job failed")
		return ctrl.Result{RequeueAfter: syncRequeueAfter}, nil
	}

	streamsConfig := streamsCR.Spec.Config
	var drift jobDrift
	if existingJob != nil {
		drift = diffJob(existingJob, jobCfg, source.ID, dest.ID, streamsConfig)
	}

	if existingJob == nil || drift.connectors {
		if err := r.testConnectors(ctx, source, dest); err != nil {
			logger.Error(err, "job connection test failed")
			return r.fail(ctx, sw, &jobCR, streamsCR, err)
		}
	}

	switch {
	case existingJob == nil:
		if err := r.ETL.CreateJob(ctx, jobCfg.createRequest(source.ID, dest.ID, streamsConfig), projectID, &userID); err != nil {
			logger.Error(err, "create job failed")
			return r.fail(ctx, sw, &jobCR, streamsCR, err)
		}
		existingJob, err = r.ETL.GetJobByName(ctx, projectID, jobCfg.Name)
		if err != nil {
			logger.Error(err, "reload job after create failed")
			return ctrl.Result{RequeueAfter: syncRequeueAfter}, nil
		}
	case drift.any():
		diffStreams := ""
		if drift.streams {
			diffStreams, err = r.streamDifferenceJSON(ctx, projectID, existingJob.ID, streamsConfig)
			if err != nil {
				logger.Error(err, "stream difference failed")
				return r.fail(ctx, sw, &jobCR, streamsCR, err)
			}
		}
		if err := r.ETL.UpdateJob(ctx, jobCfg.updateRequest(source.ID, dest.ID, streamsConfig, diffStreams), projectID, existingJob.ID, &userID); err != nil {
			logger.Error(err, "update job failed")
			return r.fail(ctx, sw, &jobCR, streamsCR, err)
		}
	}

	if err := sw.SetStatus(ctx, &jobCR, olakev1.PhaseReady, jobCR.Generation, existingJob.ID, ""); err != nil {
		logger.Error(err, "update job status failed")
		return ctrl.Result{RequeueAfter: syncRequeueAfter}, nil
	}
	if err := r.Get(ctx, client.ObjectKeyFromObject(streamsCR), streamsCR); err == nil {
		_ = sw.SetStatus(ctx, streamsCR, olakev1.PhaseReady, streamsCR.Generation, existingJob.ID, "")
	}
	return ctrl.Result{}, nil
}

func (r *JobReconciler) wait(ctx context.Context, sw StatusWriter, job *olakev1.Job, msg string) (ctrl.Result, error) {
	_ = sw.SetStatus(ctx, job, olakev1.PhasePending, job.Generation, 0, msg)
	return ctrl.Result{RequeueAfter: syncRequeueAfter}, nil
}

func (r *JobReconciler) fail(ctx context.Context, sw StatusWriter, job *olakev1.Job, streams *olakev1.Streams, err error) (ctrl.Result, error) {
	_ = sw.SetStatus(ctx, job, olakev1.PhaseFailed, job.Generation, 0, err.Error())
	if streams != nil {
		_ = r.updateStreamsStatus(ctx, sw, streams, olakev1.PhaseFailed, err.Error(), streams.Generation)
	}
	return ctrl.Result{}, nil
}

func (r *JobReconciler) testConnectors(ctx context.Context, source *models.Source, dest *models.Destination) error {
	if err := testSourceConnection(ctx, r.ETL, source.Type, source.Version, dto.JSONConfig(source.Config)); err != nil {
		return err
	}
	return testDestinationConnection(ctx, r.ETL, dest.DestType, dest.Version, dto.JSONConfig(dest.Config), source.Type, source.Version)
}

// streamsSettled reports whether a Ready job's catalog still matches the Streams CR.
func (r *JobReconciler) streamsSettled(ctx context.Context, job *olakev1.Job) (bool, error) {
	projectID := job.Spec.ProjectID
	if projectID == "" {
		return false, nil
	}
	jobCfg, err := ParseAndValidateJobConfig(job.Spec.Config)
	if err != nil {
		return false, nil
	}
	streamsCR, err := r.findStreamsForJob(ctx, job, projectID)
	if err != nil || streamsCR == nil {
		return false, err
	}
	existingJob, err := r.ETL.GetJobByName(ctx, projectID, jobCfg.Name)
	if err != nil {
		return false, nil
	}
	return utils.EqualJSON(existingJob.StreamsConfig, streamsCR.Spec.Config), nil
}

func (r *JobReconciler) streamDifferenceJSON(ctx context.Context, projectID string, jobID int, streamsConfig string) (string, error) {
	diffCatalog, err := r.ETL.GetStreamDifference(ctx, projectID, jobID, dto.StreamDifferenceRequest{
		UpdatedStreamsConfig: streamsConfig,
	})
	if err != nil {
		return "", err
	}
	diffBytes, err := json.Marshal(diffCatalog)
	if err != nil {
		return "", err
	}
	return string(diffBytes), nil
}

func (r *JobReconciler) findStreamsForJob(ctx context.Context, job *olakev1.Job, projectID string) (*olakev1.Streams, error) {
	var list olakev1.StreamsList
	if err := r.List(ctx, &list, client.InNamespace(job.Namespace)); err != nil {
		return nil, err
	}
	for i := range list.Items {
		item := &list.Items[i]
		if item.Spec.ProjectID != projectID {
			continue
		}
		if matchesNameOrID(item.Spec.Job, job.Name, job.Status.EntityID) {
			return item, nil
		}
	}
	return nil, nil
}

func (r *JobReconciler) updateStreamsStatus(ctx context.Context, sw StatusWriter, streams *olakev1.Streams, phase, message string, generation int64) error {
	key := client.ObjectKeyFromObject(streams)
	if err := r.Get(ctx, key, streams); err != nil {
		return err
	}
	return sw.SetStatus(ctx, streams, phase, generation, 0, message)
}

func (r *JobReconciler) enqueueJobsForSource(ctx context.Context, obj client.Object) []reconcile.Request {
	src, ok := obj.(*olakev1.Source)
	if !ok {
		return nil
	}
	return r.enqueueJobsReferencing(ctx, src.Namespace, func(cfg *GitOpsJobConfig) bool {
		return matchesNameOrID(cfg.Source, src.Name, src.Status.EntityID)
	})
}

func (r *JobReconciler) enqueueJobsForDestination(ctx context.Context, obj client.Object) []reconcile.Request {
	dest, ok := obj.(*olakev1.Destination)
	if !ok {
		return nil
	}
	return r.enqueueJobsReferencing(ctx, dest.Namespace, func(cfg *GitOpsJobConfig) bool {
		return matchesNameOrID(cfg.Destination, dest.Name, dest.Status.EntityID)
	})
}

func (r *JobReconciler) enqueueJobsReferencing(ctx context.Context, namespace string, match func(*GitOpsJobConfig) bool) []reconcile.Request {
	var jobList olakev1.JobList
	if err := r.List(ctx, &jobList, client.InNamespace(namespace)); err != nil {
		return nil
	}
	reqs := make([]reconcile.Request, 0, len(jobList.Items))
	for i := range jobList.Items {
		job := &jobList.Items[i]
		cfg, err := ParseAndValidateJobConfig(job.Spec.Config)
		if err != nil || !match(cfg) {
			continue
		}
		reqs = append(reqs, reconcile.Request{NamespacedName: client.ObjectKeyFromObject(job)})
	}
	return reqs
}

func (r *JobReconciler) enqueueJobForStreams(ctx context.Context, obj client.Object) []reconcile.Request {
	streams, ok := obj.(*olakev1.Streams)
	if !ok {
		return nil
	}
	var jobList olakev1.JobList
	if err := r.List(ctx, &jobList, client.InNamespace(streams.Namespace)); err != nil {
		return nil
	}
	for i := range jobList.Items {
		job := &jobList.Items[i]
		if matchesNameOrID(streams.Spec.Job, job.Name, job.Status.EntityID) {
			return []reconcile.Request{{NamespacedName: client.ObjectKeyFromObject(job)}}
		}
	}
	return nil
}

type jobDrift struct {
	connectors bool
	streams    bool
	other      bool
}

func (d jobDrift) any() bool { return d.connectors || d.streams || d.other }

func diffJob(existing *models.Job, cfg *GitOpsJobConfig, sourceID, destID int, streamsConfig string) jobDrift {
	return jobDrift{
		connectors: existing.SourceID != sourceID || existing.DestID != destID,
		streams:    !utils.EqualJSON(existing.StreamsConfig, streamsConfig),
		other: existing.Name != cfg.Name ||
			existing.Frequency != cfg.Frequency ||
			existing.Active != cfg.Activate ||
			!advancedSettingsEqual(existing.AdvancedSettings, cfg.AdvancedSettings),
	}
}

func (cfg *GitOpsJobConfig) createRequest(sourceID, destID int, streamsConfig string) *dto.CreateJobRequest {
	return &dto.CreateJobRequest{
		Name:             cfg.Name,
		Frequency:        cfg.Frequency,
		StreamsConfig:    streamsConfig,
		Activate:         cfg.Activate,
		Source:           &dto.DriverConfig{ID: &sourceID},
		Destination:      &dto.DriverConfig{ID: &destID},
		AdvancedSettings: cfg.AdvancedSettings,
	}
}

func (cfg *GitOpsJobConfig) updateRequest(sourceID, destID int, streamsConfig, differenceStreams string) *dto.UpdateJobRequest {
	return &dto.UpdateJobRequest{
		Name:              cfg.Name,
		Frequency:         cfg.Frequency,
		StreamsConfig:     streamsConfig,
		DifferenceStreams: differenceStreams,
		Activate:          cfg.Activate,
		Source:            &dto.DriverConfig{ID: &sourceID},
		Destination:       &dto.DriverConfig{ID: &destID},
		AdvancedSettings:  cfg.AdvancedSettings,
	}
}

func advancedSettingsEqual(stored *string, cfg *dto.AdvancedSettings) bool {
	if cfg == nil {
		return stored == nil || *stored == ""
	}
	if stored == nil {
		return false
	}
	encoded, err := json.Marshal(cfg)
	if err != nil {
		return false
	}
	return utils.EqualJSON(*stored, string(encoded))
}

func (r *JobReconciler) Setup(mgr ctrl.Manager) error {
	r.Recorder = mgr.GetEventRecorderFor("olake-job-controller")
	specChanged := predicate.GenerationChangedPredicate{}

	// Source/Destination watches omit GenerationChangedPredicate so a status
	// flip to Ready unblocks Jobs waiting on those entities. Streams still
	// filter on spec generation — that is the catalog the job must pick up.
	return ctrl.NewControllerManagedBy(mgr).
		For(&olakev1.Job{}, builder.WithPredicates(specChanged)).
		Watches(&olakev1.Source{}, handler.EnqueueRequestsFromMapFunc(r.enqueueJobsForSource)).
		Watches(&olakev1.Destination{}, handler.EnqueueRequestsFromMapFunc(r.enqueueJobsForDestination)).
		Watches(&olakev1.Streams{}, handler.EnqueueRequestsFromMapFunc(r.enqueueJobForStreams), builder.WithPredicates(specChanged)).
		Complete(r)
}

func (r *JobReconciler) resolveSource(ctx context.Context, projectID, ref string) (*models.Source, error) {
	if id, ok := parseResourceID(ref); ok {
		return r.ETL.GetSourceByID(ctx, projectID, id)
	}
	return r.ETL.GetSourceByName(ctx, projectID, ref)
}

func (r *JobReconciler) resolveDestination(ctx context.Context, projectID, ref string) (*models.Destination, error) {
	if id, ok := parseResourceID(ref); ok {
		return r.ETL.GetDestinationByID(ctx, projectID, id)
	}
	return r.ETL.GetDestinationByName(ctx, projectID, ref)
}

func parseResourceID(ref string) (int, bool) {
	id, err := strconv.Atoi(ref)
	if err != nil || id <= 0 {
		return 0, false
	}
	return id, true
}

// matchesNameOrID reports whether a CR field (source, destination, or job) matches
// the given resource by metadata.name or by status.entityId as a numeric string.
func matchesNameOrID(ref, name string, entityID int) bool {
	return ref == name || (entityID > 0 && ref == strconv.Itoa(entityID))
}
