package gitops

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/datazip-inc/olake-ui/server/internal/constants"
	"github.com/datazip-inc/olake-ui/server/internal/models"
	"github.com/datazip-inc/olake-ui/server/internal/models/dto"
	"github.com/datazip-inc/olake-ui/server/internal/services/etl"
	"github.com/datazip-inc/olake-ui/server/internal/utils"
)

type JobReconciler struct {
	client.Client
	ETL         *etl.Service
	Sink        StatusSink
	findStreams func(ctx context.Context, job *ResourceData, projectID string, entityID int) (*ResourceData, error)
}

func (r *JobReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var cm corev1.ConfigMap
	if err := r.Get(ctx, req.NamespacedName, &cm); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	res := resourceFromCM(&cm)
	return r.sync(ctx, &res)
}

/*
Job sync rules:

Validate / resolve:
  - Validate job spec; resolve source and destination from DB (Pending if missing).
  - Wait for the Streams ConfigMap that references this job.
  - Classify streams CM: legacy (streams[] + selected_streams) or split (selected_streams only).

Create / update:
  - Create: test source + destination connections, then CreateJob.
  - Update: if connectors or streams drifted, UpdateJob;
    stream difference runs on streams drift (clear destination).

Streams persistence:
  - Legacy create: store CM as streams_config.
  - Legacy update (connector/streams drift): discover → persist catalog as streams_config;
  - Split create: store CM as streams_config; run discover to fill selected_streams_config.
  - Split update (connector/streams drift): discover → persist catalog as streams_config + selected_streams_config.

Why discover here:
  - UI: discover/merge runs first, then you edit streams.
  - GitOps: you edit the CM first; discover takes that as input and merges with
    the source (sync_new_columns, new tables in schema, etc.).
*/
func (r *JobReconciler) sync(ctx context.Context, res *ResourceData) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	observed := r.reconcileHash(ctx, res)
	if skipReconcile(res.Annotations, observed) {
		settled, err := r.streamsSettled(ctx, res)
		if err != nil {
			logger.Error(err, "check streams drift failed")
		} else if settled {
			return ctrl.Result{}, nil
		}
	}

	jobCfg, err := ParseAndValidateJobConfig(res.Config())
	if err != nil {
		return r.failJob(ctx, res, nil, NonRetryableError(err), observed)
	}
	userID, err := requireSpec(res.ProjectID(), res.UserID())
	if err != nil {
		return r.failJob(ctx, res, nil, err, observed)
	}
	projectID := res.ProjectID()

	source, err := r.resolveSource(ctx, projectID, jobCfg.Source)
	if errors.Is(err, constants.ErrSourceNotFound) {
		return waitResource(ctx, r.Sink, res, fmt.Sprintf("waiting for source %q", jobCfg.Source), observed)
	}
	if err != nil {
		logger.Error(err, "lookup source failed")
		return requeueTransient(ctx, r.Sink, res, err, observed)
	}

	dest, err := r.resolveDestination(ctx, projectID, jobCfg.Destination)
	if errors.Is(err, constants.ErrDestinationNotFound) {
		return waitResource(ctx, r.Sink, res, fmt.Sprintf("waiting for destination %q", jobCfg.Destination), observed)
	}
	if err != nil {
		logger.Error(err, "lookup destination failed")
		return requeueTransient(ctx, r.Sink, res, err, observed)
	}

	streamsRes, err := r.findStreams(ctx, res, projectID, res.EntityID())
	if errors.Is(err, ErrStreamsNotFound) {
		return waitResource(ctx, r.Sink, res, fmt.Sprintf("waiting for Streams referencing job %q", res.Name), observed)
	}
	if err != nil {
		logger.Error(err, "find streams failed")
		return requeueTransient(ctx, r.Sink, res, err, observed)
	}

	existingJob, err := r.ETL.GetJobByName(ctx, projectID, jobCfg.Name)
	if err != nil && !errors.Is(err, constants.ErrJobNotFound) {
		logger.Error(err, "lookup job failed")
		return requeueTransient(ctx, r.Sink, res, err, observed)
	}

	streamsConfig := streamsRes.Config()
	// Streams format (legacy vs split) is chosen at job creation; do not switch afterwards.
	isSplitStream, err := classifyStreamsFormat(streamsConfig)
	if err != nil {
		return r.failJob(ctx, res, streamsRes, err, observed)
	}

	var drift jobDrift
	if existingJob != nil {
		drift = diffJob(existingJob, jobCfg, source.ID, dest.ID)
		drift.streams = !streamsCMApplied(streamsRes.Annotations, streamsRes.Data)
	}

	selectedStreamsConfig := ""
	// On update, default to the DB catalog so job-only edits (frequency, name, etc.)
	// never overwrite merged selected_columns. Discover replaces this when it runs.
	persistStreamsConfig := streamsConfig
	if existingJob != nil && existingJob.StreamsConfig != "" {
		persistStreamsConfig = existingJob.StreamsConfig
	}
	// Discover runs on:
	//   split create to populate selected_streams_config from source
	//   any update with connector or streams drift to merge new schema/columns
	needsDiscover := (existingJob == nil && isSplitStream) ||
		(existingJob != nil && (drift.connectors || drift.streams))
	if needsDiscover {
		streamsOverride := ""
		if drift.streams {
			streamsOverride = streamsConfig
		}
		catalog, err := r.discoverCatalog(ctx, source, jobCfg, existingJob, streamsOverride, isSplitStream)
		if err != nil {
			logger.Error(err, "discover schema failed")
			return r.failJob(ctx, res, streamsRes, NonRetryableError(err), observed)
		}
		if isSplitStream {
			selectedStreamsConfig = catalog.SelectedStreamsConfig
		}
		if existingJob != nil && catalog.StreamsConfig != "" {
			persistStreamsConfig = catalog.StreamsConfig
		}
	}

	if existingJob == nil || drift.connectors {
		if err := r.testConnectors(ctx, source, dest); err != nil {
			logger.Error(err, "job connection test failed")
			return r.failJob(ctx, res, streamsRes, NonRetryableError(err), observed)
		}
	}

	switch {
	case existingJob == nil:
		if err := r.ETL.CreateJob(ctx, jobCfg.createRequest(source.ID, dest.ID, persistStreamsConfig, selectedStreamsConfig), projectID, &userID); err != nil {
			logger.Error(err, "create job failed")
			return r.failJob(ctx, res, streamsRes, NonRetryableError(err), observed)
		}
		existingJob, err = r.ETL.GetJobByName(ctx, projectID, jobCfg.Name)
		if err != nil {
			logger.Error(err, "reload job after create failed")
			return requeueTransient(ctx, r.Sink, res, err, observed)
		}
	case drift.any():
		diffStreams := ""
		if drift.streams {
			diffStreams, err = r.streamDifferenceJSON(ctx, projectID, existingJob.ID, streamsConfig)
			if err != nil {
				logger.Error(err, "stream difference failed")
				return r.failJob(ctx, res, streamsRes, NonRetryableError(err), observed)
			}
		}
		if err := r.ETL.UpdateJob(ctx, jobCfg.updateRequest(source.ID, dest.ID, persistStreamsConfig, selectedStreamsConfig, diffStreams), projectID, existingJob.ID, &userID); err != nil {
			logger.Error(err, "update job failed")
			return r.failJob(ctx, res, streamsRes, NonRetryableError(err), observed)
		}
	}

	if err := successResource(ctx, r.Sink, res, existingJob.ID, observed); err != nil {
		logger.Error(err, "update job status failed")
		return requeueTransient(ctx, r.Sink, res, err, observed)
	}
	if err := successResource(ctx, r.Sink, streamsRes, existingJob.ID, ""); err != nil {
		logger.Error(err, "update streams status failed")
		return requeueTransient(ctx, r.Sink, res, err, observed)
	}
	return ctrl.Result{}, nil
}

func (r *JobReconciler) failJob(ctx context.Context, job, streams *ResourceData, err error, observedHash string) (ctrl.Result, error) {
	result, _ := failResource(ctx, r.Sink, job, err, observedHash)
	if streams != nil {
		_, _ = failResource(ctx, r.Sink, streams, err, "")
	}
	return result, nil
}

/*
reconcileHash → olake.io/observed-hash for skipReconcile.

Fingerprint: Job CM data + referenced source/destination from DB.
Streams are excluded (large catalogs); streamsSettled covers streams-only drift.

Why connectors are in the hash: a Failed job (e.g. bad credentials) must retry
when the Source/Destination CM is fixed, without requiring a Job CM edit.
*/
func (r *JobReconciler) reconcileHash(ctx context.Context, res *ResourceData) string {
	connectors := ""
	if jobCfg, err := ParseAndValidateJobConfig(res.Config()); err == nil && r.ETL != nil {
		parts := map[string]string{}
		projectID := res.ProjectID()
		if src, err := r.resolveSource(ctx, projectID, jobCfg.Source); err == nil && src != nil {
			parts["src"] = strconv.Itoa(src.ID) + "\x00" + src.Type + "\x00" + src.Version + "\x00" + src.Config
		}
		if dest, err := r.resolveDestination(ctx, projectID, jobCfg.Destination); err == nil && dest != nil {
			parts["dst"] = strconv.Itoa(dest.ID) + "\x00" + dest.DestType + "\x00" + dest.Version + "\x00" + dest.Config
		}
		if len(parts) > 0 {
			connectors = ContentHash(parts)
		}
	}
	return ContentHash(map[string]string{
		"data":       ContentHash(res.Data),
		"connectors": connectors,
	})
}

func (r *JobReconciler) testConnectors(ctx context.Context, source *models.Source, dest *models.Destination) error {
	if err := testSourceConnection(ctx, r.ETL, source.Type, source.Version, dto.JSONConfig(source.Config)); err != nil {
		return err
	}
	return testDestinationConnection(ctx, r.ETL, dest.DestType, dest.Version, dto.JSONConfig(dest.Config), source.Type, source.Version)
}

// streamsSettled: skip job reconcile when streams do not need another apply
func (r *JobReconciler) streamsSettled(ctx context.Context, job *ResourceData) (bool, error) {
	projectID := job.ProjectID()
	if projectID == "" {
		return false, nil
	}
	streamsRes, err := r.findStreams(ctx, job, projectID, job.EntityID())
	if errors.Is(err, ErrStreamsNotFound) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return streamsCMApplied(streamsRes.Annotations, streamsRes.Data), nil
}

func (r *JobReconciler) discoverCatalog(ctx context.Context, source *models.Source, jobCfg *JobConfig, existingJob *models.Job, streamsCatalogOverride string, splitStreams bool) (dto.CatalogResponse, error) {
	req := &dto.StreamsRequest{
		Name:    source.Name,
		Type:    source.Type,
		Version: source.Version,
		Config:  dto.JSONConfig(source.Config),
		JobID:   discoverJobID(existingJob),
		JobName: jobCfg.Name,
	}
	if jobCfg.AdvancedSettings != nil {
		req.MaxDiscoverThreads = jobCfg.AdvancedSettings.MaxDiscoverThreads
	}
	return r.ETL.GetSourceCatalog(ctx, req, splitStreams, streamsCatalogOverride)
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

func (r *JobReconciler) findStreamsInCluster(ctx context.Context, job *ResourceData, projectID string, entityID int) (*ResourceData, error) {
	var list corev1.ConfigMapList
	if err := r.List(ctx, &list, client.InNamespace(job.Namespace), client.MatchingLabels(managedLabels(KindStreams))); err != nil {
		return nil, err
	}
	for i := range list.Items {
		item := resourceFromCM(&list.Items[i])
		if item.ProjectID() != projectID {
			continue
		}
		if matchesNameOrID(item.JobRef(), job.Name, entityID) {
			cp := item
			return &cp, nil
		}
	}
	return nil, ErrStreamsNotFound
}

func (r *JobReconciler) enqueueJobsForSource(ctx context.Context, obj client.Object) []reconcile.Request {
	res, ok := resourceFromObject(obj)
	if !ok {
		return nil
	}
	return r.enqueueJobsReferencing(ctx, res.Namespace, func(cfg *JobConfig) bool {
		return matchesNameOrID(cfg.Source, res.Name, res.EntityID())
	})
}

func (r *JobReconciler) enqueueJobsForDestination(ctx context.Context, obj client.Object) []reconcile.Request {
	res, ok := resourceFromObject(obj)
	if !ok {
		return nil
	}
	return r.enqueueJobsReferencing(ctx, res.Namespace, func(cfg *JobConfig) bool {
		return matchesNameOrID(cfg.Destination, res.Name, res.EntityID())
	})
}

func (r *JobReconciler) enqueueJobsReferencing(ctx context.Context, namespace string, match func(*JobConfig) bool) []reconcile.Request {
	var list corev1.ConfigMapList
	if err := r.List(ctx, &list, client.InNamespace(namespace), client.MatchingLabels(managedLabels(KindJob))); err != nil {
		return nil
	}
	reqs := make([]reconcile.Request, 0, len(list.Items))
	for i := range list.Items {
		cm := &list.Items[i]
		res := resourceFromCM(cm)
		cfg, err := ParseAndValidateJobConfig(res.Config())
		if err != nil || !match(cfg) {
			continue
		}
		reqs = append(reqs, reconcile.Request{NamespacedName: client.ObjectKeyFromObject(cm)})
	}
	return reqs
}

func (r *JobReconciler) enqueueJobForStreams(ctx context.Context, obj client.Object) []reconcile.Request {
	cm, ok := obj.(*corev1.ConfigMap)
	if !ok {
		return nil
	}
	streams := resourceFromCM(cm)
	var list corev1.ConfigMapList
	if err := r.List(ctx, &list, client.InNamespace(streams.Namespace), client.MatchingLabels(managedLabels(KindJob))); err != nil {
		return nil
	}
	for i := range list.Items {
		job := resourceFromCM(&list.Items[i])
		if matchesNameOrID(streams.JobRef(), job.Name, job.EntityID()) {
			return []reconcile.Request{{NamespacedName: client.ObjectKeyFromObject(&list.Items[i])}}
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

func diffJob(existing *models.Job, cfg *JobConfig, sourceID, destID int) jobDrift {
	return jobDrift{
		connectors: existing.SourceID != sourceID || existing.DestID != destID,
		other: existing.Name != cfg.Name ||
			existing.Frequency != cfg.Frequency ||
			existing.Active != cfg.Activate ||
			!advancedSettingsEqual(existing.AdvancedSettings, cfg.AdvancedSettings),
	}
}

func (cfg *JobConfig) createRequest(sourceID, destID int, streamsConfig, selectedStreamsConfig string) *dto.CreateJobRequest {
	return &dto.CreateJobRequest{
		JobMetadata:           cfg.JobMetadata,
		StreamsConfig:         streamsConfig,
		SelectedStreamsConfig: selectedStreamsConfig,
		Source:                &dto.DriverConfig{ID: &sourceID},
		Destination:           &dto.DriverConfig{ID: &destID},
	}
}

func (cfg *JobConfig) updateRequest(sourceID, destID int, streamsConfig, selectedStreamsConfig, differenceStreams string) *dto.UpdateJobRequest {
	return &dto.UpdateJobRequest{
		JobMetadata:           cfg.JobMetadata,
		StreamsConfig:         streamsConfig,
		SelectedStreamsConfig: selectedStreamsConfig,
		DifferenceStreams:     differenceStreams,
		Source:                &dto.DriverConfig{ID: &sourceID},
		Destination:           &dto.DriverConfig{ID: &destID},
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
	r.findStreams = r.findStreamsInCluster
	dataChanged := predicate.ResourceVersionChangedPredicate{}

	return ctrl.NewControllerManagedBy(mgr).
		Named("gitops-job").
		For(&corev1.ConfigMap{}, builder.WithPredicates(kindPredicate(KindJob))).
		Watches(&corev1.ConfigMap{}, handler.EnqueueRequestsFromMapFunc(r.enqueueJobsForSource), builder.WithPredicates(kindPredicate(KindSource))).
		Watches(&corev1.ConfigMap{}, handler.EnqueueRequestsFromMapFunc(r.enqueueJobsForDestination), builder.WithPredicates(kindPredicate(KindDestination))).
		Watches(&corev1.ConfigMap{}, handler.EnqueueRequestsFromMapFunc(r.enqueueJobForStreams), builder.WithPredicates(kindPredicate(KindStreams), dataChanged)).
		Watches(&corev1.Secret{}, handler.EnqueueRequestsFromMapFunc(r.enqueueJobsForSource), builder.WithPredicates(kindPredicate(KindSource))).
		Watches(&corev1.Secret{}, handler.EnqueueRequestsFromMapFunc(r.enqueueJobsForDestination), builder.WithPredicates(kindPredicate(KindDestination))).
		Complete(r)
}
