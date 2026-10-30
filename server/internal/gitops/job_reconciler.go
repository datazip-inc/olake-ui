package gitops

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/datazip-inc/olake-ui/server/internal/constants"
	"github.com/datazip-inc/olake-ui/server/internal/models"
	"github.com/datazip-inc/olake-ui/server/internal/models/dto"
	"github.com/datazip-inc/olake-ui/server/internal/types"
	"github.com/datazip-inc/olake-ui/server/internal/utils"
)

// jobService is the part of the ETL service the job reconciler calls.
type jobService interface {
	catalogService
	connectionService
	GetSourceByName(ctx context.Context, projectID, name string) (*models.Source, error)
	GetSourceByID(ctx context.Context, projectID string, id int) (*models.Source, error)
	GetDestinationByName(ctx context.Context, projectID, name string) (*models.Destination, error)
	GetDestinationByID(ctx context.Context, projectID string, id int) (*models.Destination, error)
	GetJobByName(ctx context.Context, projectID, name string) (*models.Job, error)
	CreateJob(ctx context.Context, req *dto.CreateJobRequest, projectID string, userID *int) error
	UpdateJob(ctx context.Context, req *dto.UpdateJobRequest, projectID string, jobID int, userID *int) error
	ActivateJob(ctx context.Context, jobID int, req dto.JobStatusRequest, userID *int) error
}

// JobReconciler applies a job manifest and its streams manifest to OLake.
type JobReconciler struct {
	ETL  jobService
	Sink StatusSink
}

/*
Job reconcile rules:

Validate / resolve:
  - Validate job spec; resolve source and destination from DB (Pending if missing).
  - Wait for the Streams ConfigMap that references this job.
  - The source must be at MinStreamsV2Version or later; the streams CM is selection-only.

Create / update:
  - Create or connector change: test source + destination connections first.
  - Create: CreateJob. Update: UpdateJob on any drift;
    stream difference runs on streams drift (clear destination).
  - Activation changes go through ActivateJob.

Streams persistence (see resolveCatalog):
  - The catalog comes from the CM on create or when the CM changed, else from the job.
  - Discover merges it with the source on connector or streams drift.

Why discover here:
  - UI: discover/merge runs first, then you edit streams.
  - GitOps: you edit the CM first; discover takes that as input and merges with
    the source (sync_new_columns, new tables in schema, etc.).
*/

// reconcileJob applies job res. streamsRes is its streams manifest, or nil when there is none yet.
func (r *JobReconciler) reconcileJob(ctx context.Context, res, streamsRes *ResourceData) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	projectID := res.ProjectID()

	refs := r.loadJobRefs(ctx, res)
	observed := reconcileHash(res, refs.source, refs.dest)
	if skipReconcile(res.Annotations, observed) && streamsSettled(streamsRes) {
		return ctrl.Result{}, nil
	}

	if refs.cfgErr != nil {
		return r.failJob(ctx, res, nil, NonRetryableError(refs.cfgErr), observed)
	}
	userID, err := requireSpec(projectID, res.UserID())
	if err != nil {
		return r.failJob(ctx, res, nil, err, observed)
	}

	if errors.Is(refs.sourceErr, constants.ErrSourceNotFound) {
		return waitResource(ctx, r.Sink, res, fmt.Sprintf("waiting for source %q", refs.cfg.Source), observed)
	}
	if refs.sourceErr != nil {
		logger.Error(refs.sourceErr, "lookup source failed")
		return requeueTransient(ctx, r.Sink, res, refs.sourceErr, observed)
	}

	if errors.Is(refs.destErr, constants.ErrDestinationNotFound) {
		return waitResource(ctx, r.Sink, res, fmt.Sprintf("waiting for destination %q", refs.cfg.Destination), observed)
	}
	if refs.destErr != nil {
		logger.Error(refs.destErr, "lookup destination failed")
		return requeueTransient(ctx, r.Sink, res, refs.destErr, observed)
	}
	jobCfg, source, dest := refs.cfg, refs.source, refs.dest

	if streamsRes == nil {
		return waitResource(ctx, r.Sink, res, fmt.Sprintf("waiting for Streams referencing job %q", res.Name), observed)
	}

	existingJob, err := r.ETL.GetJobByName(ctx, projectID, jobCfg.Name)
	if err != nil && !errors.Is(err, constants.ErrJobNotFound) {
		logger.Error(err, "lookup job failed")
		return requeueTransient(ctx, r.Sink, res, err, observed)
	}
	if existingJob != nil {
		res.bindEntity(existingJob.ID)
		streamsRes.bindEntity(existingJob.ID)
	}

	cmCatalog, err := parseStreamsCM(streamsRes.Config())
	if err != nil {
		return r.failJob(ctx, res, streamsRes, err, observed)
	}
	if !utils.SupportsStreamsV2(source.Version) {
		err := fmt.Errorf("GitOps requires source version %s or later (source %q is on %s); upgrade the source",
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
		req := jobCfg.createRequest(source.ID, source.Version, dest.ID, catalog)
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
		req := jobCfg.updateRequest(source.ID, source.Version, dest.ID, existingJob.Active, catalog, diffStreams)
		if err := r.ETL.UpdateJob(ctx, req, projectID, existingJob.ID, &userID); err != nil {
			// clear-destination ends on its own; retry the update after it instead of failing for this hash
			if errors.Is(err, constants.ErrClearDestinationRunning) {
				return waitResource(ctx, r.Sink, res, err.Error(), observed)
			}
			logger.Error(err, "update job failed")
			return r.failJob(ctx, res, streamsRes, NonRetryableError(err), observed)
		}
	}

	// CreateJob applies the activation; UpdateJob keeps the current one, so a change on an
	// existing job is applied through ActivateJob, which also pauses or resumes the schedule.
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

// jobRefs is the job config and the source/destination it points at, read once per
// reconcile so the hash and the reconcile see the same values. A nil field has its
// reason in the matching error.
type jobRefs struct {
	cfg    *JobConfig
	source *models.Source
	dest   *models.Destination

	cfgErr    error
	sourceErr error
	destErr   error
}

func (r *JobReconciler) loadJobRefs(ctx context.Context, res *ResourceData) jobRefs {
	var refs jobRefs
	refs.cfg, refs.cfgErr = ParseAndValidateJobConfig(res.Config())
	if refs.cfgErr != nil {
		return refs // no valid config, so nothing to resolve
	}
	refs.source, refs.sourceErr = r.resolveSource(ctx, res.ProjectID(), refs.cfg.Source)
	refs.dest, refs.destErr = r.resolveDestination(ctx, res.ProjectID(), refs.cfg.Destination)
	return refs
}

/*
reconcileHash → olake.io/observed-hash for skipReconcile.

Fingerprint: Job CM data + referenced source/destination from DB.
Streams are excluded (large catalogs); streamsSettled covers streams-only drift.

Why connectors are in the hash: a Failed job (e.g. bad credentials) must retry
when the Source/Destination CM is fixed, without requiring a Job CM edit.
*/
func reconcileHash(res *ResourceData, source *models.Source, dest *models.Destination) string {
	parts := map[string]string{}
	if source != nil {
		parts["src"] = strconv.Itoa(source.ID) + "\x00" + source.Type + "\x00" + source.Version + "\x00" + strconv.FormatInt(source.UpdatedAt.UnixNano(), 10)
	}
	if dest != nil {
		parts["dst"] = strconv.Itoa(dest.ID) + "\x00" + dest.DestType + "\x00" + dest.Version + "\x00" + strconv.FormatInt(dest.UpdatedAt.UnixNano(), 10)
	}
	connectors := ""
	if len(parts) > 0 {
		connectors = ContentHash(parts)
	}
	return ContentHash(map[string]string{
		"data":       ContentHash(res.Data),
		"connectors": connectors,
	})
}

func (r *JobReconciler) testConnectors(ctx context.Context, source *models.Source, dest *models.Destination) error {
	if err := testSourceConnection(ctx, r.ETL, source.ProjectID, source.Type, source.Version, dto.JSONConfig(source.Config)); err != nil {
		return err
	}
	return testDestinationConnection(ctx, r.ETL, dest.ProjectID, dest.DestType, dest.Version, dto.JSONConfig(dest.Config), source.Type, source.Version)
}

// streamsSettled: skip job reconcile when streams do not need another apply. Like skipReconcile,
// a Streams CM that failed for its current data is settled; it retries when its data, the Job CM
// or a connector changes. Retrying it on every status patch would loop on errors that differ per
// run, such as ones naming a Temporal workflow ID.
func streamsSettled(streams *ResourceData) bool {
	return streams != nil && skipReconcile(streams.Annotations, ContentHash(streams.Data))
}

// streamsBelongTo reports whether streams is the streams manifest of job: same namespace and
// project, and its job ref names the job or its entity ID.
func streamsBelongTo(streams, job *ResourceData) bool {
	projectID := job.ProjectID()
	return streams.Kind == KindStreams &&
		streams.Namespace == job.Namespace &&
		projectID != "" && streams.ProjectID() == projectID &&
		matchesNameOrID(streams.JobRef(), job.Name, job.EntityID())
}

// jobDependsOn reports whether a change to dep can change the reconcile result of job: dep is the
// job's streams manifest, or the source/destination its config refers to.
func jobDependsOn(job, dep *ResourceData) bool {
	switch dep.Kind {
	case KindStreams:
		return streamsBelongTo(dep, job)
	case KindSource, KindDestination:
		if dep.Namespace != job.Namespace {
			return false
		}
		cfg, err := ParseAndValidateJobConfig(job.Config())
		if err != nil {
			return false
		}
		ref := cfg.Source
		if dep.Kind == KindDestination {
			ref = cfg.Destination
		}
		// a job refers to the OLake entity name from the object's config, not the object's name
		return matchesNameOrID(ref, configName(dep.Config()), dep.EntityID())
	default:
		return false
	}
}

// catalogService is the part of the ETL service that discovers and diffs job catalogs.
type catalogService interface {
	DiscoverWithCatalog(ctx context.Context, projectID string, req *dto.StreamsRequest, stored types.StreamsCatalog) (types.StreamsCatalog, error)
	GetStreamDifference(ctx context.Context, projectID string, jobID int, req dto.StreamDifferenceRequest) (string, error)
	AwaitOperationOutput(ctx context.Context, operationID string) (map[string]interface{}, error)
}

// resolveCatalog returns the split catalog to store for the job. It starts from the Streams CM when
// the CM changed (or on create), otherwise from the job's stored catalog, and runs discover when
// the job's connectors or streams changed.
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

	// re-merge with the source only when an existing job's source/destination or streams changed
	needsDiscover := existingJob != nil && (drift.connectors || drift.streams)

	// the CM has no available streams, so take them from the job or, for a new job, a fresh discover
	if base.Available == "" {
		base.Available = stored.Available
		if base.Available == "" {
			fresh, err := svc.DiscoverWithCatalog(ctx, source.ProjectID, req, types.StreamsCatalog{})
			if err != nil {
				return types.StreamsCatalog{}, err
			}
			base.Available = fresh.Available
		}
		needsDiscover = true // merge the CM's selection with the available streams
	}

	if needsDiscover { // discover merges base with the source and writes the merged catalog
		discovered, err := svc.DiscoverWithCatalog(ctx, source.ProjectID, req, base)
		if err != nil {
			return types.StreamsCatalog{}, err
		}
		base = discovered
	}
	return types.StreamsCatalog{Available: base.Available, Selected: base.Selected}, nil
}

// streamDifferenceJSON diffs the job's stored catalog against the new one.
func streamDifferenceJSON(ctx context.Context, svc catalogService, job *models.Job, catalog types.StreamsCatalog) (string, error) {
	operationID, err := svc.GetStreamDifference(ctx, job.ProjectID, job.ID, dto.StreamDifferenceRequest{
		UpdatedAvailableStreamsConfig: catalog.Available,
		UpdatedSelectedStreamsConfig:  catalog.Selected,
	})
	if err != nil {
		return "", err
	}
	diffCatalog, err := svc.AwaitOperationOutput(ctx, operationID)
	if err != nil {
		return "", fmt.Errorf("failed to get stream difference: %w", err)
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

func (cfg *JobConfig) createRequest(sourceID int, sourceVersion string, destID int, catalog types.StreamsCatalog) *dto.CreateJobRequest {
	return &dto.CreateJobRequest{
		JobMetadata:            cfg.JobMetadata,
		AvailableStreamsConfig: catalog.Available,
		SelectedStreamsConfig:  catalog.Selected,
		Source:                 &dto.DriverConfig{ID: &sourceID, Version: sourceVersion},
		Destination:            &dto.DriverConfig{ID: &destID},
	}
}

// updateRequest keeps the job's current activation (active): UpdateJob stores it without pausing or
// resuming the schedule, so the reconciler changes activation through ActivateJob.
func (cfg *JobConfig) updateRequest(sourceID int, sourceVersion string, destID int, active bool, catalog types.StreamsCatalog, differenceStreams string) *dto.UpdateJobRequest {
	metadata := cfg.JobMetadata
	metadata.Activate = active
	return &dto.UpdateJobRequest{
		JobMetadata:            metadata,
		AvailableStreamsConfig: catalog.Available,
		SelectedStreamsConfig:  catalog.Selected,
		DifferenceStreams:      differenceStreams,
		Source:                 &dto.DriverConfig{ID: &sourceID, Version: sourceVersion},
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
