package reconciler

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/datazip-inc/olake-ui/server/internal/models/dto"
	"github.com/datazip-inc/olake-ui/server/internal/utils"
)

const (
	PhasePending = "Pending"
	PhaseReady   = "Ready"
	PhaseFailed  = "Failed"

	KindSource      = "source"
	KindDestination = "destination"
	KindJob         = "job"
	KindStreams     = "streams"

	// Status annotations (operator-owned). Git/Argo manifests should only set
	// metadata.labels and data — not these keys. Unlike a CRD status subresource,
	// ConfigMap has no API split between user fields and operator fields; if a
	// sync replaces metadata or clears annotations, skipReconcile sees a fresh
	// object and may re-run reconcile (see kubernetes/sink.go). Safe with normal
	// kubectl apply / Argo merge sync; risky with Replace, force SSA, or
	// manifests that set metadata.annotations explicitly (including {}).
	AnnotationPhase        = "olake.io/phase"
	AnnotationMessage      = "olake.io/message"
	AnnotationEntityID     = "olake.io/entity-id"
	AnnotationObservedHash = "olake.io/observed-hash"

	DataProjectID = "projectId"
	DataUserID    = "userId"
	DataConfig    = "config"
	DataJob       = "job"
)

// ResourceData is the view of an OLake-managed resource, whatever the GitOps mode stores it in.
// ObjectKind is set and read by the mode only: in Kubernetes mode it says whether a ConfigMap or a
// Secret backs the resource.
type ResourceData struct {
	Kind        string
	ObjectKind  string
	Name        string
	Namespace   string
	Data        map[string]string
	Annotations map[string]string
}

func (r *ResourceData) key() string {
	return r.Kind + "/" + r.Namespace + "/" + r.Name
}

func (r *ResourceData) ProjectID() string { return r.Data[DataProjectID] }
func (r *ResourceData) UserID() string    { return r.Data[DataUserID] }
func (r *ResourceData) Config() string    { return r.Data[DataConfig] }
func (r *ResourceData) JobRef() string    { return r.Data[DataJob] }

func (r *ResourceData) EntityID() int {
	n, _ := strconv.Atoi(r.Annotations[AnnotationEntityID])
	return n
}

// bindEntity records the OLake entity r manages once a reconcile has found it, so every status
// written afterwards carries its id: a Failed or Pending phase means the entity exists but the
// latest data is not applied, not that there is no entity.
func (r *ResourceData) bindEntity(id int) {
	if r.Annotations == nil {
		r.Annotations = map[string]string{}
	}
	r.Annotations[AnnotationEntityID] = strconv.Itoa(id)
}

// StatusSink writes phase and failure indicators. SetPhase always gets the observed hash to store;
// the reconcilers resolve it (observedHashForResource), so a sink never computes one.
type StatusSink interface {
	Indicator(ctx context.Context, r *ResourceData, action, errMsg string)
	SetPhase(ctx context.Context, r *ResourceData, phase, message, entityID, observedHash string) error
}

// parseResourceID treats a pure positive integer ref as a DB entity id.
func parseResourceID(ref string) (int, bool) {
	id, err := strconv.Atoi(ref)
	if err != nil || id <= 0 {
		return 0, false
	}
	return id, true
}

// configName returns the OLake entity name in a Source/Destination/Job data.config, or "" when
// the config is not valid JSON.
func configName(config string) string {
	var named struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal([]byte(config), &named); err != nil {
		return ""
	}
	return named.Name
}

// matchesNameOrID is true when ref is the resource name or its entity id string.
func matchesNameOrID(ref, name string, entityID int) bool {
	return ref == name || (entityID > 0 && ref == strconv.Itoa(entityID))
}

type resourceSpec struct {
	ID      int
	Name    string
	Type    string
	Version string
	Config  dto.JSONConfig
}

// resource is the source/destination wiring for the shared reconcile loop.
type resource interface {
	kind() string
	notFoundErr() error
	parse(config string) (*resourceSpec, error)
	get(ctx context.Context, projectID, name string) (*resourceSpec, error)
	test(ctx context.Context, projectID string, spec *resourceSpec) error
	create(ctx context.Context, spec *resourceSpec, projectID string, userID *int) error
	update(ctx context.Context, id int, spec *resourceSpec, projectID string, userID *int) error
}

func resourceMatches(existing, spec *resourceSpec) bool {
	return existing.Name == spec.Name &&
		existing.Type == spec.Type &&
		existing.Version == spec.Version &&
		utils.EqualJSON(existing.Config.String(), spec.Config.String())
}

func reconcileResource(ctx context.Context, sink StatusSink, secrets *SecretRefs, r resource, res *ResourceData) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	observedHash := ContentHash(res.Data)
	if skipResource(res, observedHash, secrets) {
		return ctrl.Result{}, nil
	}

	spec, err := r.parse(res.Config())
	if err != nil {
		return failResource(ctx, sink, res, NonRetryableError(err), observedHash)
	}
	userID, err := requireSpec(res.ProjectID(), res.UserID())
	if err != nil {
		return failResource(ctx, sink, res, err, observedHash)
	}

	resolved, hasRefs, err := secrets.resolve(ctx, spec.Config.String())
	if err != nil {
		logger.Error(err, "resolve "+r.kind()+" secret references failed")
		return failOrRequeue(ctx, sink, res, err, observedHash)
	}
	spec.Config = dto.JSONConfig(resolved)

	existing, err := r.get(ctx, res.ProjectID(), spec.Name)
	if err != nil && !errors.Is(err, r.notFoundErr()) {
		logger.Error(err, "lookup "+r.kind()+" failed")
		return requeueTransient(ctx, sink, res, err, observedHash)
	}
	if existing != nil {
		res.bindEntity(existing.ID)
	}

	changed := existing != nil && !resourceMatches(existing, spec)
	if existing == nil || changed {
		if err := r.test(ctx, res.ProjectID(), spec); err != nil {
			logger.Error(err, r.kind()+" connection test failed")
			return failOrRequeue(ctx, sink, res, err, observedHash)
		}
	}

	switch {
	case existing == nil:
		if err := r.create(ctx, spec, res.ProjectID(), &userID); err != nil {
			logger.Error(err, "create "+r.kind()+" failed")
			return failResource(ctx, sink, res, NonRetryableError(err), observedHash)
		}
		existing, err = r.get(ctx, res.ProjectID(), spec.Name)
		if err != nil {
			logger.Error(err, "reload "+r.kind()+" after create failed")
			return requeueTransient(ctx, sink, res, err, observedHash)
		}
	case changed:
		if err := r.update(ctx, existing.ID, spec, res.ProjectID(), &userID); err != nil {
			logger.Error(err, "update "+r.kind()+" failed")
			return failResource(ctx, sink, res, NonRetryableError(err), observedHash)
		}
	}

	if err := successResource(ctx, sink, res, existing.ID, observedHash); err != nil {
		logger.Error(err, "update "+r.kind()+" status failed")
		return requeueTransient(ctx, sink, res, err, observedHash)
	}
	return ctrl.Result{RequeueAfter: secrets.next(hasRefs)}, nil
}

// skipResource: true when a source/destination needs no reconcile. A synced one is skipped, except
// a Ready one whose config may hold secret references: it is re-checked every refresh interval so a
// rotated secret reaches OLake without a change in git. mayRefresh only looks for "${" in the raw
// config, so a synced resource without references is skipped without parsing or resolving it.
func skipResource(res *ResourceData, observedHash string, secrets *SecretRefs) bool {
	if !skipReconcile(res.Annotations, observedHash) {
		return false
	}
	return res.Annotations[AnnotationPhase] != PhaseReady || !secrets.mayRefresh(res.Config())
}
