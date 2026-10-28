package gitops

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"unicode"

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

	LabelManaged = "olake.io/managed"
	LabelKind    = "olake.io/kind"

	// Status annotations (operator-owned). Git/Argo manifests should only set
	// metadata.labels and data — not these keys. Unlike a CRD status subresource,
	// ConfigMap has no API split between user fields and operator fields; if a
	// sync replaces metadata or clears annotations, skipReconcile sees a fresh
	// object and may re-run reconcile (see k8s_sink.go). Safe with normal
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

	LabelManagedValue = "true"

	// ObjectKind identifies which Kubernetes type backs a ResourceData, so
	// StatusSink implementations know whether to Get/Patch a ConfigMap or a
	// Secret for the same managed resource.
	ObjectKindConfigMap = "ConfigMap"
	ObjectKindSecret    = "Secret"

	indicatorNameMax  = 63
	terminationLogMax = 4096
)

// ResourceData is the view of an OLake-managed resource. It may be backed by
// either a ConfigMap or a Secret (see ObjectKind) — Source and Destination
// accept either so credentials can live in a Secret instead of git-tracked
// ConfigMap data. Job and Streams are ConfigMap-only (no credentials)
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

// StatusSink writes phase and failure indicators.
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

// indicatorName returns "<name>-olake-<kind>" truncated to 63 chars.
func indicatorName(name, kind string) string {
	suffix := "-olake-" + kind
	n := sanitizeDNS1123(name)
	if n == "" {
		n = "unnamed"
	}
	if len(n)+len(suffix) > indicatorNameMax {
		n = strings.TrimRight(n[:indicatorNameMax-len(suffix)], "-")
	}
	return n + suffix
}

func sanitizeDNS1123(s string) string {
	s = strings.ToLower(s)
	var b strings.Builder
	prevDash := false
	for _, r := range s {
		ok := unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-'
		if !ok {
			r = '-'
		}
		if r == '-' {
			if prevDash {
				continue
			}
			prevDash = true
		} else {
			prevDash = false
		}
		b.WriteRune(r)
	}
	return strings.Trim(b.String(), "-")
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
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

func reconcileResource(ctx context.Context, sink StatusSink, r resource, res *ResourceData) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	observedHash := ContentHash(res.Data)
	if skipReconcile(res.Annotations, observedHash) {
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

	existing, err := r.get(ctx, res.ProjectID(), spec.Name)
	if err != nil && !errors.Is(err, r.notFoundErr()) {
		logger.Error(err, "lookup "+r.kind()+" failed")
		return requeueTransient(ctx, sink, res, err, observedHash)
	}

	changed := existing != nil && !resourceMatches(existing, spec)
	if existing == nil || changed {
		if err := r.test(ctx, res.ProjectID(), spec); err != nil {
			logger.Error(err, r.kind()+" connection test failed")
			if !errors.Is(err, ErrNonRetryable) {
				return requeueTransient(ctx, sink, res, err, observedHash)
			}
			return failResource(ctx, sink, res, err, observedHash)
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
	return ctrl.Result{}, nil
}
