package services

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/beego/beego/v2/client/orm"
	"github.com/datazip-inc/olake-ui/server/internal/models"
	"github.com/datazip-inc/olake-ui/server/internal/models/dto"
	"github.com/datazip-inc/olake-ui/server/internal/services/temporal"
	"github.com/datazip-inc/olake-ui/server/utils/logger"
)

const declarativeManagedBy = "cli_apply"

type cliBundleFiles struct {
	BundleName string
	Source     []byte
	Dest       []byte
	Streams    []byte
	State      []byte
	Overlay    dto.ApplyCLIBundleOverlay
}

type desiredCLIBundle struct {
	BundleName    string
	ApplyIdentity string
	Source        dto.DriverConfig
	Destination   dto.DriverConfig
	JobName       string
	Frequency     string
	Activate      bool
	StreamsConfig string
	StateFile     *string
}

type applyResourcePlan[T any] struct {
	Current *T
	Action  dto.ApplyPlanAction
	Name    string
	Fields  []string
}

type applyStatePlan struct {
	Action dto.ApplyPlanAction
	Fields []string
}

type cliBundleApplyPlan struct {
	BundleName string
	Effective  dto.ApplyCLIBundleEffective
	Source     applyResourcePlan[models.Source]
	Dest       applyResourcePlan[models.Destination]
	Job        applyResourcePlan[models.Job]
	State      applyStatePlan
}

func (s *ETLService) ApplyCLIBundle(
	ctx context.Context,
	projectID string,
	fileName string,
	archiveData []byte,
	opts dto.ApplyCLIBundleOptions,
	userID *int,
) (*dto.ApplyCLIBundleResponse, error) {
	if userID == nil {
		return nil, fmt.Errorf("user id is required")
	}

	files, err := extractCLIBundleFiles(fileName, archiveData)
	if err != nil {
		return nil, err
	}

	desired, err := buildDesiredCLIBundle(files)
	if err != nil {
		return nil, err
	}

	plan, err := s.buildCLIBundlePlan(projectID, desired)
	if err != nil {
		return nil, err
	}

	if opts.DryRun {
		resp := plan.toResponse(true, opts.Prune)
		return &resp, nil
	}

	if err := s.executeCLIBundlePlan(ctx, projectID, desired, &plan, userID); err != nil {
		return nil, err
	}

	resp := plan.toResponse(false, opts.Prune)
	return &resp, nil
}

func extractCLIBundleFiles(fileName string, archiveData []byte) (*cliBundleFiles, error) {
	bundleName := inferBundleName(fileName)
	files := map[string][]byte{}
	roots := map[string]struct{}{}

	loadFile := func(name string, reader io.Reader) error {
		baseName := path.Base(name)
		if baseName == "." || baseName == "/" || baseName == "" {
			return nil
		}
		switch baseName {
		case "source.json", "destination.json", "streams.json", "state.json", "olake-ui.json":
		default:
			return nil
		}

		if _, exists := files[baseName]; exists {
			return fmt.Errorf("duplicate file %q found in bundle", baseName)
		}

		content, err := io.ReadAll(reader)
		if err != nil {
			return fmt.Errorf("failed to read %s: %s", baseName, err)
		}
		files[baseName] = content

		cleanName := strings.TrimPrefix(path.Clean(name), "./")
		if cleanName != "." && cleanName != baseName {
			root := strings.Split(cleanName, "/")[0]
			if root != "" && root != baseName {
				roots[root] = struct{}{}
			}
		}

		return nil
	}

	switch {
	case strings.HasSuffix(strings.ToLower(fileName), ".zip"):
		reader, err := zip.NewReader(bytes.NewReader(archiveData), int64(len(archiveData)))
		if err != nil {
			return nil, fmt.Errorf("failed to open zip bundle: %s", err)
		}

		for _, file := range reader.File {
			if file.FileInfo().IsDir() {
				continue
			}

			rc, err := file.Open()
			if err != nil {
				return nil, fmt.Errorf("failed to open %s in zip bundle: %s", file.Name, err)
			}
			err = loadFile(file.Name, rc)
			rc.Close()
			if err != nil {
				return nil, err
			}
		}
	case strings.HasSuffix(strings.ToLower(fileName), ".tar.gz"), strings.HasSuffix(strings.ToLower(fileName), ".tgz"):
		gzipReader, err := gzip.NewReader(bytes.NewReader(archiveData))
		if err != nil {
			return nil, fmt.Errorf("failed to open tar.gz bundle: %s", err)
		}
		defer gzipReader.Close()

		tarReader := tar.NewReader(gzipReader)
		for {
			header, err := tarReader.Next()
			if err == io.EOF {
				break
			}
			if err != nil {
				return nil, fmt.Errorf("failed to read tar bundle: %s", err)
			}
			if header.FileInfo().IsDir() {
				continue
			}
			if err := loadFile(header.Name, tarReader); err != nil {
				return nil, err
			}
		}
	default:
		return nil, fmt.Errorf("unsupported bundle format for %q, use .zip or .tar.gz", fileName)
	}

	required := []string{"source.json", "destination.json", "streams.json"}
	for _, name := range required {
		if _, ok := files[name]; !ok {
			return nil, fmt.Errorf("bundle is missing required file %q", name)
		}
	}

	if len(roots) == 1 {
		for root := range roots {
			bundleName = root
		}
	}

	result := &cliBundleFiles{
		BundleName: bundleName,
		Source:     files["source.json"],
		Dest:       files["destination.json"],
		Streams:    files["streams.json"],
		State:      files["state.json"],
	}

	if overlayBytes, ok := files["olake-ui.json"]; ok {
		if err := json.Unmarshal(overlayBytes, &result.Overlay); err != nil {
			return nil, fmt.Errorf("failed to parse olake-ui.json: %s", err)
		}
	}

	return result, nil
}

func buildDesiredCLIBundle(files *cliBundleFiles) (*desiredCLIBundle, error) {
	sourceConfig, err := canonicalizeJSON(files.Source)
	if err != nil {
		return nil, fmt.Errorf("failed to parse source.json: %s", err)
	}

	destConfig, err := canonicalizeJSON(files.Dest)
	if err != nil {
		return nil, fmt.Errorf("failed to parse destination.json: %s", err)
	}

	streamsConfig, err := canonicalizeJSON(files.Streams)
	if err != nil {
		return nil, fmt.Errorf("failed to parse streams.json: %s", err)
	}

	applyIdentity := strings.TrimSpace(files.Overlay.ApplyIdentity)
	if applyIdentity == "" {
		applyIdentity = sanitizeApplyIdentity(files.BundleName)
	}

	sourceType := strings.ToLower(strings.TrimSpace(files.Overlay.SourceType))
	sourceVersion := strings.TrimSpace(files.Overlay.SourceVersion)
	destType := strings.ToLower(strings.TrimSpace(files.Overlay.DestinationType))
	destVersion := strings.TrimSpace(files.Overlay.DestinationVersion)

	if sourceType == "" || sourceVersion == "" || destVersion == "" {
		return nil, fmt.Errorf("olake-ui.json must include source_type, source_version, and destination_version for server-side apply")
	}

	if destType == "" {
		inferredDestType, err := inferDestinationType(destConfig)
		if err != nil {
			return nil, err
		}
		destType = inferredDestType
	}

	if err := dto.ValidateSourceType(sourceType); err != nil {
		return nil, err
	}
	if err := dto.ValidateDestinationType(destType); err != nil {
		return nil, err
	}

	jobName := strings.TrimSpace(files.Overlay.JobName)
	if jobName == "" {
		jobName = files.BundleName
	}

	sourceName := strings.TrimSpace(files.Overlay.SourceName)
	if sourceName == "" {
		sourceName = fmt.Sprintf("%s-source", jobName)
	}

	destName := strings.TrimSpace(files.Overlay.DestinationName)
	if destName == "" {
		destName = fmt.Sprintf("%s-destination", jobName)
	}

	activate := false
	if files.Overlay.Activate != nil {
		activate = *files.Overlay.Activate
	}

	var stateFile *string
	if len(files.State) > 0 {
		stateConfig, err := canonicalizeJSON(files.State)
		if err != nil {
			return nil, fmt.Errorf("failed to parse state.json: %s", err)
		}
		stateFile = &stateConfig
	}

	return &desiredCLIBundle{
		BundleName:    files.BundleName,
		ApplyIdentity: applyIdentity,
		JobName:       jobName,
		Frequency:     strings.TrimSpace(files.Overlay.Frequency),
		Activate:      activate,
		StreamsConfig: streamsConfig,
		StateFile:     stateFile,
		Source: dto.DriverConfig{
			Name:    sourceName,
			Type:    sourceType,
			Version: sourceVersion,
			Config:  sourceConfig,
		},
		Destination: dto.DriverConfig{
			Name:    destName,
			Type:    destType,
			Version: destVersion,
			Config:  destConfig,
		},
	}, nil
}

func (s *ETLService) buildCLIBundlePlan(projectID string, desired *desiredCLIBundle) (cliBundleApplyPlan, error) {
	plan := cliBundleApplyPlan{
		BundleName: desired.BundleName,
		Effective: dto.ApplyCLIBundleEffective{
			ApplyIdentity:      desired.ApplyIdentity,
			JobName:            desired.JobName,
			SourceName:         desired.Source.Name,
			SourceType:         desired.Source.Type,
			SourceVersion:      desired.Source.Version,
			DestinationName:    desired.Destination.Name,
			DestinationType:    desired.Destination.Type,
			DestinationVersion: desired.Destination.Version,
			Frequency:          desired.Frequency,
			Activate:           desired.Activate,
		},
	}

	sourceCurrent, err := s.resolveSourceForApply(projectID, desired)
	if err != nil {
		return plan, err
	}
	plan.Source = applyResourcePlan[models.Source]{
		Current: sourceCurrent,
		Name:    desired.Source.Name,
	}
	plan.Source.Fields = diffSourceApply(sourceCurrent, desired)
	plan.Source.Action = actionFromDiff(sourceCurrent == nil, plan.Source.Fields)

	destCurrent, err := s.resolveDestinationForApply(projectID, desired)
	if err != nil {
		return plan, err
	}
	plan.Dest = applyResourcePlan[models.Destination]{
		Current: destCurrent,
		Name:    desired.Destination.Name,
	}
	plan.Dest.Fields = diffDestinationApply(destCurrent, desired)
	plan.Dest.Action = actionFromDiff(destCurrent == nil, plan.Dest.Fields)

	jobCurrent, err := s.resolveJobForApply(projectID, desired)
	if err != nil {
		return plan, err
	}
	plan.Job = applyResourcePlan[models.Job]{
		Current: jobCurrent,
		Name:    desired.JobName,
	}
	plan.Job.Fields = diffJobApply(jobCurrent, desired)
	plan.Job.Action = actionFromDiff(jobCurrent == nil, plan.Job.Fields)
	plan.State = buildStatePlan(jobCurrent, desired.StateFile)

	return plan, nil
}

func (s *ETLService) executeCLIBundlePlan(
	ctx context.Context,
	projectID string,
	desired *desiredCLIBundle,
	plan *cliBundleApplyPlan,
	userID *int,
) error {
	user := &models.User{ID: *userID}

	needsJobReconcile := plan.Job.Current != nil && (plan.Source.Action != dto.ApplyPlanUnchanged ||
		plan.Dest.Action != dto.ApplyPlanUnchanged ||
		plan.Job.Action != dto.ApplyPlanUnchanged ||
		plan.State.Action == dto.ApplyPlanUpdated ||
		plan.State.Action == dto.ApplyPlanCreated)

	if needsJobReconcile {
		clearRunning, _, err := isWorkflowRunning(ctx, s.temporal, projectID, plan.Job.Current.ID, temporal.ClearDestination)
		if err != nil {
			return fmt.Errorf("failed to check if clear-destination is running: %s", err)
		}
		if clearRunning {
			return fmt.Errorf("clear-destination is in progress, cannot apply bundle")
		}

		if err := cancelAllJobWorkflows(ctx, s.temporal, []*models.Job{plan.Job.Current}, projectID); err != nil {
			return fmt.Errorf("failed to cancel running syncs: %s", err)
		}
	}

	sourceModel, err := s.persistSourceFromApply(projectID, desired, plan.Source, user)
	if err != nil {
		return err
	}
	plan.Source.Current = sourceModel

	destModel, err := s.persistDestinationFromApply(projectID, desired, plan.Dest, user)
	if err != nil {
		return err
	}
	plan.Dest.Current = destModel

	jobModel, err := s.persistJobFromApply(ctx, projectID, desired, plan, sourceModel, destModel, user)
	if err != nil {
		return err
	}
	plan.Job.Current = jobModel

	return nil
}

func (s *ETLService) persistSourceFromApply(
	projectID string,
	desired *desiredCLIBundle,
	plan applyResourcePlan[models.Source],
	user *models.User,
) (*models.Source, error) {
	if plan.Current == nil {
		source := &models.Source{
			Name:      desired.Source.Name,
			Type:      desired.Source.Type,
			Version:   desired.Source.Version,
			Config:    desired.Source.Config,
			ProjectID: projectID,
			ManagedBy: declarativeManagedBy,
			ApplyID:   desired.ApplyIdentity,
			CreatedBy: user,
			UpdatedBy: user,
		}
		if err := s.db.CreateSource(source); err != nil {
			return nil, fmt.Errorf("failed to create source: %s", err)
		}
		return s.db.GetSourceByID(source.ID)
	}

	if plan.Action == dto.ApplyPlanUnchanged {
		return plan.Current, nil
	}

	source := plan.Current
	source.Name = desired.Source.Name
	source.Type = desired.Source.Type
	source.Version = desired.Source.Version
	source.Config = desired.Source.Config
	source.ManagedBy = declarativeManagedBy
	source.ApplyID = desired.ApplyIdentity
	source.UpdatedBy = user

	if err := s.db.UpdateSource(source); err != nil {
		return nil, fmt.Errorf("failed to update source: %s", err)
	}
	return s.db.GetSourceByID(source.ID)
}

func (s *ETLService) persistDestinationFromApply(
	projectID string,
	desired *desiredCLIBundle,
	plan applyResourcePlan[models.Destination],
	user *models.User,
) (*models.Destination, error) {
	if plan.Current == nil {
		dest := &models.Destination{
			Name:      desired.Destination.Name,
			DestType:  desired.Destination.Type,
			Version:   desired.Destination.Version,
			Config:    desired.Destination.Config,
			ProjectID: projectID,
			ManagedBy: declarativeManagedBy,
			ApplyID:   desired.ApplyIdentity,
			CreatedBy: user,
			UpdatedBy: user,
		}
		if err := s.db.CreateDestination(dest); err != nil {
			return nil, fmt.Errorf("failed to create destination: %s", err)
		}
		return s.db.GetDestinationByID(dest.ID)
	}

	if plan.Action == dto.ApplyPlanUnchanged {
		return plan.Current, nil
	}

	dest := plan.Current
	dest.Name = desired.Destination.Name
	dest.DestType = desired.Destination.Type
	dest.Version = desired.Destination.Version
	dest.Config = desired.Destination.Config
	dest.ManagedBy = declarativeManagedBy
	dest.ApplyID = desired.ApplyIdentity
	dest.UpdatedBy = user

	if err := s.db.UpdateDestination(dest); err != nil {
		return nil, fmt.Errorf("failed to update destination: %s", err)
	}
	return s.db.GetDestinationByID(dest.ID)
}

func (s *ETLService) persistJobFromApply(
	ctx context.Context,
	projectID string,
	desired *desiredCLIBundle,
	plan *cliBundleApplyPlan,
	source *models.Source,
	dest *models.Destination,
	user *models.User,
) (*models.Job, error) {
	initialState := "{}"
	if desired.StateFile != nil {
		initialState = *desired.StateFile
	}

	if plan.Job.Current == nil {
		job := &models.Job{
			Name:          desired.JobName,
			SourceID:      source,
			DestID:        dest,
			Active:        desired.Activate,
			Frequency:     desired.Frequency,
			ManagedBy:     declarativeManagedBy,
			ApplyID:       desired.ApplyIdentity,
			StreamsConfig: desired.StreamsConfig,
			State:         initialState,
			ProjectID:     projectID,
			CreatedBy:     user,
			UpdatedBy:     user,
		}

		if err := s.db.CreateJob(job); err != nil {
			return nil, fmt.Errorf("failed to create job: %s", err)
		}

		if err := s.temporal.CreateSchedule(ctx, job); err != nil {
			if deleteErr := s.db.DeleteJob(job.ID); deleteErr != nil {
				logger.Errorf("failed to rollback created job[%d]: %s", job.ID, deleteErr)
			}
			return nil, fmt.Errorf("failed to create temporal schedule: %s", err)
		}

		if !desired.Activate {
			if err := s.temporal.PauseSchedule(ctx, projectID, job.ID); err != nil {
				return nil, fmt.Errorf("failed to pause created job schedule: %s", err)
			}
		}

		return s.db.GetJobByID(job.ID, true)
	}

	job := plan.Job.Current
	jobChanged := plan.Job.Action != dto.ApplyPlanUnchanged || plan.State.Action == dto.ApplyPlanUpdated || plan.State.Action == dto.ApplyPlanCreated

	if jobChanged {
		updateParams := orm.Params{
			"name":           desired.JobName,
			"source_id":      source.ID,
			"dest_id":        dest.ID,
			"active":         desired.Activate,
			"frequency":      desired.Frequency,
			"managed_by":     declarativeManagedBy,
			"apply_id":       desired.ApplyIdentity,
			"streams_config": desired.StreamsConfig,
			"updated_by_id":  user.ID,
		}
		if desired.StateFile != nil {
			updateParams["state"] = *desired.StateFile
		}
		if err := s.db.UpdateJob(job.ID, updateParams); err != nil {
			return nil, fmt.Errorf("failed to update job: %s", err)
		}
	}

	needsScheduleRefresh := jobChanged ||
		plan.Source.Action != dto.ApplyPlanUnchanged ||
		plan.Dest.Action != dto.ApplyPlanUnchanged

	persistedJob, err := s.db.GetJobByID(job.ID, true)
	if err != nil {
		return nil, fmt.Errorf("failed to reload job: %s", err)
	}

	if needsScheduleRefresh {
		persistedJob.SourceID = source
		persistedJob.DestID = dest
		if err := s.temporal.EnsureSchedule(ctx, persistedJob); err != nil {
			return nil, fmt.Errorf("failed to reconcile temporal schedule: %s", err)
		}
	}

	if desired.Activate != job.Active {
		if desired.Activate {
			if err := s.temporal.ResumeSchedule(ctx, projectID, job.ID); err != nil {
				return nil, fmt.Errorf("failed to resume job schedule: %s", err)
			}
		} else {
			if err := s.temporal.PauseSchedule(ctx, projectID, job.ID); err != nil {
				return nil, fmt.Errorf("failed to pause job schedule: %s", err)
			}
		}
	}

	return s.db.GetJobByID(job.ID, true)
}

func (s *ETLService) resolveSourceForApply(projectID string, desired *desiredCLIBundle) (*models.Source, error) {
	source, err := s.db.GetSourceByProjectIDAndApplyID(projectID, desired.ApplyIdentity)
	if err != nil {
		return nil, err
	}
	if source != nil {
		return source, nil
	}
	return s.db.GetSourceByProjectIDAndName(projectID, desired.Source.Name)
}

func (s *ETLService) resolveDestinationForApply(projectID string, desired *desiredCLIBundle) (*models.Destination, error) {
	dest, err := s.db.GetDestinationByProjectIDAndApplyID(projectID, desired.ApplyIdentity)
	if err != nil {
		return nil, err
	}
	if dest != nil {
		return dest, nil
	}
	return s.db.GetDestinationByProjectIDAndName(projectID, desired.Destination.Name)
}

func (s *ETLService) resolveJobForApply(projectID string, desired *desiredCLIBundle) (*models.Job, error) {
	job, err := s.db.GetJobByProjectIDAndApplyID(projectID, desired.ApplyIdentity, true)
	if err != nil {
		return nil, err
	}
	if job != nil {
		return job, nil
	}
	return s.db.GetJobByProjectIDAndName(projectID, desired.JobName, true)
}

func diffSourceApply(current *models.Source, desired *desiredCLIBundle) []string {
	if current == nil {
		return []string{"name", "type", "version", "config", "managed_by", "apply_id"}
	}

	fields := make([]string, 0, 6)
	if current.Name != desired.Source.Name {
		fields = append(fields, "name")
	}
	if current.Type != desired.Source.Type {
		fields = append(fields, "type")
	}
	if current.Version != desired.Source.Version {
		fields = append(fields, "version")
	}
	if !sameCanonicalJSON(current.Config, desired.Source.Config) {
		fields = append(fields, "config")
	}
	if current.ManagedBy != declarativeManagedBy {
		fields = append(fields, "managed_by")
	}
	if current.ApplyID != desired.ApplyIdentity {
		fields = append(fields, "apply_id")
	}
	return fields
}

func diffDestinationApply(current *models.Destination, desired *desiredCLIBundle) []string {
	if current == nil {
		return []string{"name", "type", "version", "config", "managed_by", "apply_id"}
	}

	fields := make([]string, 0, 6)
	if current.Name != desired.Destination.Name {
		fields = append(fields, "name")
	}
	if current.DestType != desired.Destination.Type {
		fields = append(fields, "type")
	}
	if current.Version != desired.Destination.Version {
		fields = append(fields, "version")
	}
	if !sameCanonicalJSON(current.Config, desired.Destination.Config) {
		fields = append(fields, "config")
	}
	if current.ManagedBy != declarativeManagedBy {
		fields = append(fields, "managed_by")
	}
	if current.ApplyID != desired.ApplyIdentity {
		fields = append(fields, "apply_id")
	}
	return fields
}

func diffJobApply(current *models.Job, desired *desiredCLIBundle) []string {
	if current == nil {
		return []string{"name", "source", "destination", "frequency", "activate", "streams_config", "managed_by", "apply_id"}
	}

	fields := make([]string, 0, 8)
	if current.Name != desired.JobName {
		fields = append(fields, "name")
	}
	if current.SourceID == nil || current.SourceID.Name != desired.Source.Name {
		fields = append(fields, "source")
	}
	if current.DestID == nil || current.DestID.Name != desired.Destination.Name {
		fields = append(fields, "destination")
	}
	if current.Frequency != desired.Frequency {
		fields = append(fields, "frequency")
	}
	if current.Active != desired.Activate {
		fields = append(fields, "activate")
	}
	if !sameCanonicalJSON(current.StreamsConfig, desired.StreamsConfig) {
		fields = append(fields, "streams_config")
	}
	if current.ManagedBy != declarativeManagedBy {
		fields = append(fields, "managed_by")
	}
	if current.ApplyID != desired.ApplyIdentity {
		fields = append(fields, "apply_id")
	}
	return fields
}

func buildStatePlan(current *models.Job, desiredState *string) applyStatePlan {
	if desiredState == nil {
		return applyStatePlan{Action: dto.ApplyPlanPreserved}
	}

	if current == nil {
		return applyStatePlan{Action: dto.ApplyPlanCreated, Fields: []string{"state"}}
	}

	if sameCanonicalJSON(current.State, *desiredState) {
		return applyStatePlan{Action: dto.ApplyPlanUnchanged}
	}

	return applyStatePlan{Action: dto.ApplyPlanUpdated, Fields: []string{"state"}}
}

func actionFromDiff(isCreate bool, fields []string) dto.ApplyPlanAction {
	if isCreate {
		return dto.ApplyPlanCreated
	}
	if len(fields) == 0 {
		return dto.ApplyPlanUnchanged
	}
	return dto.ApplyPlanUpdated
}

func (p cliBundleApplyPlan) toResponse(dryRun, prune bool) dto.ApplyCLIBundleResponse {
	resp := dto.ApplyCLIBundleResponse{
		DryRun:    dryRun,
		Prune:     prune,
		Bundle:    p.BundleName,
		Effective: p.Effective,
		Source: dto.ApplyCLIBundleResourcePlan{
			Action: p.Source.Action,
			Name:   p.Source.Name,
			Fields: p.Source.Fields,
		},
		Dest: dto.ApplyCLIBundleResourcePlan{
			Action: p.Dest.Action,
			Name:   p.Dest.Name,
			Fields: p.Dest.Fields,
		},
		Job: dto.ApplyCLIBundleResourcePlan{
			Action: p.Job.Action,
			Name:   p.Job.Name,
			Fields: p.Job.Fields,
		},
		State: dto.ApplyCLIBundleStatePlan{
			Action: p.State.Action,
			Fields: p.State.Fields,
		},
	}

	if p.Source.Current != nil {
		resp.Source.ID = &p.Source.Current.ID
	}
	if p.Dest.Current != nil {
		resp.Dest.ID = &p.Dest.Current.ID
	}
	if p.Job.Current != nil {
		resp.Job.ID = &p.Job.Current.ID
	}

	return resp
}

func inferDestinationType(destConfig string) (string, error) {
	var config map[string]interface{}
	if err := json.Unmarshal([]byte(destConfig), &config); err != nil {
		return "", fmt.Errorf("failed to parse destination.json for type inference: %s", err)
	}

	typeValue, ok := config["type"].(string)
	if !ok || strings.TrimSpace(typeValue) == "" {
		return "", fmt.Errorf("olake-ui.json must include destination_type when destination.json does not expose a top-level type")
	}

	return strings.ToLower(strings.TrimSpace(typeValue)), nil
}

func inferBundleName(fileName string) string {
	base := filepath.Base(fileName)
	for _, suffix := range []string{".tar.gz", ".tgz", ".zip"} {
		if strings.HasSuffix(strings.ToLower(base), suffix) {
			return strings.TrimSuffix(base, suffix)
		}
	}
	return strings.TrimSuffix(base, filepath.Ext(base))
}

func sanitizeApplyIdentity(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	replacer := strings.NewReplacer(" ", "-", "/", "-", "\\", "-", "_", "-", ".", "-")
	value = replacer.Replace(value)

	var builder strings.Builder
	lastDash := false
	for _, char := range value {
		if (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') {
			builder.WriteRune(char)
			lastDash = false
			continue
		}
		if !lastDash {
			builder.WriteRune('-')
			lastDash = true
		}
	}

	result := strings.Trim(builder.String(), "-")
	if result == "" {
		return "cli-bundle"
	}
	return result
}

func canonicalizeJSON(raw []byte) (string, error) {
	var payload interface{}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	if err := decoder.Decode(&payload); err != nil {
		return "", err
	}

	normalized := normalizeJSON(payload)
	data, err := json.Marshal(normalized)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func sameCanonicalJSON(left, right string) bool {
	leftCanonical, err := canonicalizeJSON([]byte(left))
	if err != nil {
		return strings.TrimSpace(left) == strings.TrimSpace(right)
	}

	rightCanonical, err := canonicalizeJSON([]byte(right))
	if err != nil {
		return strings.TrimSpace(left) == strings.TrimSpace(right)
	}

	return leftCanonical == rightCanonical
}

func normalizeJSON(value interface{}) interface{} {
	switch typed := value.(type) {
	case map[string]interface{}:
		keys := make([]string, 0, len(typed))
		for key := range typed {
			keys = append(keys, key)
		}
		sort.Strings(keys)

		normalized := make(map[string]interface{}, len(typed))
		for _, key := range keys {
			normalized[key] = normalizeJSON(typed[key])
		}
		return normalized
	case []interface{}:
		normalized := make([]interface{}, len(typed))
		for index, item := range typed {
			normalized[index] = normalizeJSON(item)
		}
		return normalized
	default:
		return typed
	}
}
