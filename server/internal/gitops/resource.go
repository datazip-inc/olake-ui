package gitops

import (
	"context"
	"strconv"
	"strings"
	"unicode"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
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

// ResourceData is the view of an OLake-managed resource.
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
	SpawnIndicator(ctx context.Context, r *ResourceData, errMsg string) error
	DeleteIndicator(ctx context.Context, r *ResourceData) error
	SetPhase(ctx context.Context, r *ResourceData, phase, message, entityID, observedHash string) error
}

func resourceFromCM(cm *corev1.ConfigMap) ResourceData {
	ann := cm.Annotations
	if ann == nil {
		ann = map[string]string{}
	}
	data := cm.Data
	if data == nil {
		data = map[string]string{}
	}
	return ResourceData{
		Kind:        cm.Labels[LabelKind],
		ObjectKind:  ObjectKindConfigMap,
		Name:        cm.Name,
		Namespace:   cm.Namespace,
		Data:        data,
		Annotations: ann,
	}
}

// resourceFromSecret mirrors resourceFromCM for Secret-backed Source/Destination
// resources. Secret.Data values are already base64-decoded []byte by client-go.
func resourceFromSecret(secret *corev1.Secret) ResourceData {
	ann := secret.Annotations
	if ann == nil {
		ann = map[string]string{}
	}
	data := make(map[string]string, len(secret.Data))
	for k, v := range secret.Data {
		data[k] = string(v)
	}
	return ResourceData{
		Kind:        secret.Labels[LabelKind],
		ObjectKind:  ObjectKindSecret,
		Name:        secret.Name,
		Namespace:   secret.Namespace,
		Data:        data,
		Annotations: ann,
	}
}

// resourceFromObject builds a ResourceData from whichever managed type triggered
// a watch event. ok is false for any other object type.
func resourceFromObject(obj client.Object) (ResourceData, bool) {
	switch o := obj.(type) {
	case *corev1.ConfigMap:
		return resourceFromCM(o), true
	case *corev1.Secret:
		return resourceFromSecret(o), true
	default:
		return ResourceData{}, false
	}
}

// newManagedObject returns an empty client.Object of the given ObjectKind, for
// Get/Patch calls that must target the correct Kubernetes type.
func newManagedObject(objectKind string) client.Object {
	if objectKind == ObjectKindSecret {
		return &corev1.Secret{}
	}
	return &corev1.ConfigMap{}
}

// identityEnqueue re-queues the very object that triggered the watch
func identityEnqueue(_ context.Context, obj client.Object) []reconcile.Request {
	return []reconcile.Request{{NamespacedName: client.ObjectKeyFromObject(obj)}}
}

// fetchManaged loads a Source/Destination resource that may be backed by
// either a ConfigMap or a Secret with the same name. A ConfigMap takes
// precedence if both somehow exist. Returns a NotFound error if neither exists.
func fetchManaged(ctx context.Context, c client.Client, key client.ObjectKey) (*ResourceData, error) {
	var cm corev1.ConfigMap
	err := c.Get(ctx, key, &cm)
	if err == nil {
		res := resourceFromCM(&cm)
		return &res, nil
	}
	if !apierrors.IsNotFound(err) {
		return nil, err
	}

	var secret corev1.Secret
	if err := c.Get(ctx, key, &secret); err != nil {
		return nil, err
	}
	res := resourceFromSecret(&secret)
	return &res, nil
}

func kindPredicate(kind string) predicate.Predicate {
	return predicate.NewPredicateFuncs(func(obj client.Object) bool {
		labels := obj.GetLabels()
		return labels[LabelManaged] == LabelManagedValue && labels[LabelKind] == kind
	})
}

func managedLabels(kind string) map[string]string {
	return map[string]string{
		LabelManaged: LabelManagedValue,
		LabelKind:    kind,
	}
}

// parseResourceID treats a pure positive integer ref as a DB entity id.
func parseResourceID(ref string) (int, bool) {
	id, err := strconv.Atoi(ref)
	if err != nil || id <= 0 {
		return 0, false
	}
	return id, true
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
