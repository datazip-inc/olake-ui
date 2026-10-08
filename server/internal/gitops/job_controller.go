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
	"github.com/datazip-inc/olake-ui/server/internal/types"
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
	return r.reconcileJob(ctx, &res)
}

/*
Job reconcile rules:

Validate / resolve:
  - Validate job spec; resolve source and destination from DB (Pending if missing).
  - Wait for the Streams ConfigMap that references this job.
  - Classify streams CM: legacy (streams[] + selected_streams) or split (selected_streams only).

Create / update:
  - Create or connector change: test source + destination connections first.
  - Create: CreateJob. Update: UpdateJob on any drift;
    stream difference runs on streams drift (clear destination).
  - Activation changes go through ActivateJob.

Streams persistence (see resolveCatalog):
  - The catalog comes from the CM on create or when the CM changed, else from the job.
  - Discover merges it with the source on connector or streams drift.
  - A CM with streams[] is stored legacy, a selection-only CM split; a split job stays split.

Why discover here:
  - UI: discover/merge runs first, then you edit streams.
  - GitOps: you edit the CM first; discover takes that as input and merges with
    the source (sync_new_columns, new tables in schema, etc.).
*/
func (r *JobReconciler) reconcileJob(ctx context.Context, res *ResourceData) (ctrl.Result, error) {
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

	cmCatalog, err := parseStreamsCM(streamsRes.Config())
	if err != nil {
		return r.failJob(ctx, res, streamsRes, err, observed)
	}
	if cmCatalog.Streams == "" && !utils.SupportsStreamsV2(source.Version) {
		err := fmt.Errorf("selection-only streams config needs source version %s or later (source %q is on %s); add streams[] to the config or upgrade the source",
			constants.MinStreamsV2Version, source.Name, source.Version)
		return r.failJob(ctx, res, streamsRes, NonRetryableError(err), observed)
	}

	var drift jobDrift
	if existingJob != nil {
		drift = diffJob(existingJob, jobCfg, source.ID, dest.ID)
		drift.streams = !streamsCMApplied(streamsRes.Annotations, streamsRes.Data)
	}

	// test connections before discover: cheaper, and a bad credential fails with a clear error
	if existingJob == nil || drift.connectors {
		if err := r.testConnectors(ctx, source, dest); err != nil {
			logger.Error(err, "job connection test failed")
			if !errors.Is(err, ErrNonRetryable) {
				return requeueTransient(ctx, r.Sink, res, err, observed)
			}
			return r.failJob(ctx, res, streamsRes, err, observed)
		}
	}

	catalog, err := resolveCatalog(ctx, r.ETL, source, jobCfg, existingJob, cmCatalog, drift)
	if err != nil {
		logger.Error(err, "discover schema failed")
		if err = workflowError(err); !errors.Is(err, ErrNonRetryable) {
			return requeueTransient(ctx, r.Sink, res, err, observed)
		}
		return r.failJob(ctx, res, streamsRes, err, observed)
	}

	switch {
	case existingJob == nil:
		req := jobCfg.createRequest(source.ID, dest.ID, catalog)
		if err := r.ETL.CreateJob(ctx, req, projectID, &userID); err != nil {
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
			diffStreams, err = streamDifferenceJSON(ctx, r.ETL, existingJob, catalog)
			if err != nil {
				logger.Error(err, "stream difference failed")
				if err = workflowError(err); !errors.Is(err, ErrNonRetryable) {
					return requeueTransient(ctx, r.Sink, res, err, observed)
				}
				return r.failJob(ctx, res, streamsRes, err, observed)
			}
		}
		req := jobCfg.updateRequest(source.ID, dest.ID, existingJob.Active, catalog, diffStreams)
		if err := r.ETL.UpdateJob(ctx, req, projectID, existingJob.ID, &userID); err != nil {
			// clear-destination ends on its own; retry the update after it instead of failing for this hash
			if errors.Is(err, constants.ErrClearDestinationRunning) {
				return waitResource(ctx, r.Sink, res, err.Error(), observed)
			}
			logger.Error(err, "update job failed")
			return r.failJob(ctx, res, streamsRes, NonRetryableError(err), observed)
		}
	}

	// CreateJob always creates an active job and UpdateJob keeps the current activation, so a
	// change is applied through ActivateJob, which also pauses or resumes the schedule.
	if existingJob.Active != jobCfg.Activate {
		if err := r.ETL.ActivateJob(ctx, existingJob.ID, dto.JobStatusRequest{Activate: jobCfg.Activate}, &userID); err != nil {
			logger.Error(err, "set job activation failed")
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

// streamsSettled: skip job reconcile when streams do not need another apply. Like skipReconcile,
// a Streams CM that failed for its current data is settled; it retries when its data, the Job CM
// or a connector changes. Retrying it on every status patch would loop on errors that differ per
// run, such as ones naming a Temporal workflow ID.
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
	return skipReconcile(streamsRes.Annotations, ContentHash(streamsRes.Data)), nil
}

// catalogService is the part of the ETL service that discovers and diffs job catalogs.
type catalogService interface {
	DiscoverWithCatalog(ctx context.Context, req *dto.StreamsRequest, stored types.StreamsCatalog) (types.StreamsCatalog, error)
	GetStreamDifference(ctx context.Context, projectID string, jobID int, req dto.StreamDifferenceRequest) (map[string]interface{}, error)
}

// resolveCatalog returns the catalog to store for the job. It starts from the Streams CM when the
// CM changed (or on create), otherwise from the job's stored catalog, and runs discover when the
// job's connectors or streams changed. A CM with streams[] is stored legacy, a selection-only CM
// split; a job that is already split stays split.
func resolveCatalog(ctx context.Context, svc catalogService, source *models.Source, jobCfg *JobConfig, existingJob *models.Job, cmCatalog types.StreamsCatalog, drift jobDrift) (types.StreamsCatalog, error) {
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

	var stored types.StreamsCatalog
	if existingJob != nil {
		stored = existingJob.StreamsCatalog()
	}

	base := cmCatalog

	// no change in streams then use the stored catalog
	if existingJob != nil && !drift.streams {
		base = stored
	}
	// store split if the base is split or selection-only, or the job is already split; legacy otherwise
	split := base.Streams == "" || stored.IsSplit()

	// re-merge with the source only when an existing job's source/destination or streams changed
	needsDiscover := existingJob != nil && (drift.connectors || drift.streams)

	// selection-only CM: it has no available streams, so take them from the job or a fresh discover
	if base.Streams == "" && base.Available == "" {
		base.Available = stored.Available // reuse the stored available streams
		if base.Available == "" {         // new job or legacy job: discover available streams from scratch
			fresh, err := svc.DiscoverWithCatalog(ctx, req, types.StreamsCatalog{})
			if err != nil {
				return types.StreamsCatalog{}, err
			}
			// without available streams the merge below would run a fresh discover and drop the CM's selection
			if fresh.Available == "" {
				return types.StreamsCatalog{}, fmt.Errorf("discover wrote no available streams for source %q on %s", source.Name, source.Version)
			}
			base.Available = fresh.Available
		}
		needsDiscover = true // merge the CM's selection with the available streams
	}

	if needsDiscover { // discover merges base with the source and writes the merged catalog
		discovered, err := svc.DiscoverWithCatalog(ctx, req, base)
		if err != nil {
			return types.StreamsCatalog{}, err
		}
		base = discovered
	}

	if !split { // legacy catalog: store streams_config only
		return types.StreamsCatalog{Streams: base.Streams}, nil
	}
	if !base.IsSplit() { // the source did not write the split format (e.g. downgraded below MinStreamsV2Version)
		return types.StreamsCatalog{}, fmt.Errorf("job needs the split streams format, but source %q on %s did not write it", source.Name, source.Version)
	}
	return types.StreamsCatalog{Available: base.Available, Selected: base.Selected}, nil
}

// streamDifferenceJSON diffs the job's stored catalog against the new one. GetStreamDifference
// accepts each side in its own format, so a legacy job can be diffed against a split catalog.
func streamDifferenceJSON(ctx context.Context, svc catalogService, job *models.Job, catalog types.StreamsCatalog) (string, error) {
	if stored := job.StreamsCatalog(); stored.Streams == "" && !stored.IsSplit() {
		return "", nil
	}
	diffCatalog, err := svc.GetStreamDifference(ctx, job.ProjectID, job.ID, dto.StreamDifferenceRequest{
		UpdatedStreamsConfig:          catalog.Streams,
		UpdatedAvailableStreamsConfig: catalog.Available,
		UpdatedSelectedStreamsConfig:  catalog.Selected,
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
	// a job refers to the OLake source name from the object's config, not the object's name
	name := configName(res.Config())
	return r.enqueueJobsReferencing(ctx, res.Namespace, func(cfg *JobConfig) bool {
		return matchesNameOrID(cfg.Source, name, res.EntityID())
	})
}

func (r *JobReconciler) enqueueJobsForDestination(ctx context.Context, obj client.Object) []reconcile.Request {
	res, ok := resourceFromObject(obj)
	if !ok {
		return nil
	}
	// a job refers to the OLake destination name from the object's config, not the object's name
	name := configName(res.Config())
	return r.enqueueJobsReferencing(ctx, res.Namespace, func(cfg *JobConfig) bool {
		return matchesNameOrID(cfg.Destination, name, res.EntityID())
	})
}

func (r *JobReconciler) enqueueJobsReferencing(ctx context.Context, namespace string, match func(*JobConfig) bool) []reconcile.Request {
	var list corev1.ConfigMapList
	if err := r.List(ctx, &list, client.InNamespace(namespace), client.MatchingLabels(managedLabels(KindJob))); err != nil {
		log.FromContext(ctx).Error(err, "list jobs failed")
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
		log.FromContext(ctx).Error(err, "list jobs failed")
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
			!advancedSettingsEqual(existing.AdvancedSettings, cfg.AdvancedSettings),
	}
}

func (cfg *JobConfig) createRequest(sourceID, destID int, catalog types.StreamsCatalog) *dto.CreateJobRequest {
	return &dto.CreateJobRequest{
		JobMetadata:            cfg.JobMetadata,
		StreamsConfig:          catalog.Streams,
		AvailableStreamsConfig: catalog.Available,
		SelectedStreamsConfig:  catalog.Selected,
		Source:                 &dto.DriverConfig{ID: &sourceID},
		Destination:            &dto.DriverConfig{ID: &destID},
	}
}

// updateRequest keeps the job's current activation (active): UpdateJob stores it without pausing or
// resuming the schedule, so the reconciler changes activation through ActivateJob.
func (cfg *JobConfig) updateRequest(sourceID, destID int, active bool, catalog types.StreamsCatalog, differenceStreams string) *dto.UpdateJobRequest {
	metadata := cfg.JobMetadata
	metadata.Activate = active
	return &dto.UpdateJobRequest{
		JobMetadata:            metadata,
		StreamsConfig:          catalog.Streams,
		AvailableStreamsConfig: catalog.Available,
		SelectedStreamsConfig:  catalog.Selected,
		DifferenceStreams:      differenceStreams,
		Source:                 &dto.DriverConfig{ID: &sourceID},
		Destination:            &dto.DriverConfig{ID: &destID},
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
