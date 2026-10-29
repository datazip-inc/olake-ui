package gitops

import (
	"context"
	"fmt"
	"os"
	"strings"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/datazip-inc/olake-ui/server/internal/appconfig"
	"github.com/datazip-inc/olake-ui/server/internal/gitops/secretref"
	"github.com/datazip-inc/olake-ui/server/internal/services"
	"github.com/datazip-inc/olake-ui/server/internal/utils/logger"
)

var scheme = runtime.NewScheme()

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
}

// Controller runs GitOps in Kubernetes mode: it watches labelled ConfigMaps/Secrets and
// hands them to the reconcilers.
type Controller struct {
	mgr       ctrl.Manager
	app       *services.AppService
	namespace string
}

// Setup validates GitOps prerequisites synchronously.
func Setup(app *services.AppService, restConfig *rest.Config) (*Controller, error) {
	managed := labels.SelectorFromSet(labels.Set{LabelManaged: LabelManagedValue})

	cacheOpts := cache.Options{
		ByObject: map[client.Object]cache.ByObject{
			&corev1.ConfigMap{}: {Label: managed},
			&corev1.Secret{}:    {Label: managed},
		},
	}
	ns := strings.TrimSpace(os.Getenv("POD_NAMESPACE"))
	if ns == "" {
		return nil, fmt.Errorf("POD_NAMESPACE is required for in-cluster GitOps (namespace-scoped RBAC)")
	}
	cacheOpts.DefaultNamespaces = map[string]cache.Config{ns: {}}

	mgr, err := ctrl.NewManager(restConfig, ctrl.Options{
		Scheme: scheme,
		Cache:  cacheOpts,
	})
	if err != nil {
		return nil, fmt.Errorf("create controller manager: %w", err)
	}

	logger.Info("GitOps controller manager configured for labelled ConfigMap/Secret watch in namespace scope")
	return &Controller{mgr: mgr, app: app, namespace: ns}, nil
}

// Run registers reconcilers and blocks until ctx is cancelled.
func (c *Controller) Run(ctx context.Context) error {
	etlSvc := c.app.ETL()
	k8sClient := c.mgr.GetClient()
	sink := NewK8sSink(k8sClient, c.mgr.GetEventRecorderFor("olake-gitops"), etlSvc.Temporal())
	cfg := appconfig.Load()
	secrets := newSecretRefs(&cfg, secretref.K8sConfig{Client: k8sClient, Namespace: c.namespace})

	source := &SourceReconciler{ETL: etlSvc, Sink: sink, Secrets: secrets}
	if err := setupResourceController(c.mgr, "gitops-source", KindSource, &resourceController{client: k8sClient, reconcile: source.reconcile}); err != nil {
		return fmt.Errorf("setup source controller: %w", err)
	}

	dest := &DestinationReconciler{ETL: etlSvc, Sink: sink, Secrets: secrets}
	if err := setupResourceController(c.mgr, "gitops-destination", KindDestination, &resourceController{client: k8sClient, reconcile: dest.reconcile}); err != nil {
		return fmt.Errorf("setup destination controller: %w", err)
	}

	job := &jobController{client: k8sClient, job: &JobReconciler{ETL: etlSvc, Sink: sink}}
	if err := job.setup(c.mgr); err != nil {
		return fmt.Errorf("setup job controller: %w", err)
	}

	logger.Info("starting GitOps controller manager")
	return c.mgr.Start(ctx)
}

// resourceController loads the Source/Destination object a watch event names and reconciles it.
type resourceController struct {
	client    client.Client
	reconcile func(ctx context.Context, res *ResourceData) (ctrl.Result, error)
}

func (c *resourceController) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	res, err := fetchManaged(ctx, c.client, req.NamespacedName)
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	return c.reconcile(ctx, res)
}

func setupResourceController(mgr ctrl.Manager, name, kind string, rec reconcile.Reconciler) error {
	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&corev1.ConfigMap{}).
		Watches(&corev1.Secret{}, handler.EnqueueRequestsFromMapFunc(identityEnqueue)).
		WithEventFilter(kindPredicate(kind)).
		Complete(rec)
}

// jobController loads a Job ConfigMap and its Streams ConfigMap and reconciles them. It also
// re-queues jobs when the Source, Destination or Streams object they depend on changes.
type jobController struct {
	client client.Client
	job    *JobReconciler
}

func (c *jobController) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var cm corev1.ConfigMap
	if err := c.client.Get(ctx, req.NamespacedName, &cm); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	job := resourceFromCM(&cm)

	var streamsList corev1.ConfigMapList
	if err := c.client.List(ctx, &streamsList, client.InNamespace(job.Namespace), client.MatchingLabels(managedLabels(KindStreams))); err != nil {
		return ctrl.Result{}, fmt.Errorf("list streams: %w", err)
	}
	return c.job.reconcileJob(ctx, &job, findStreams(&streamsList, &job))
}

// findStreams returns the Streams ConfigMap of job, or nil when there is none yet.
func findStreams(list *corev1.ConfigMapList, job *ResourceData) *ResourceData {
	for i := range list.Items {
		streams := resourceFromCM(&list.Items[i])
		if streamsBelongTo(&streams, job) {
			return &streams
		}
	}
	return nil
}

// enqueueDependentJobs maps a Source, Destination or Streams object to the jobs that depend on it.
func (c *jobController) enqueueDependentJobs(ctx context.Context, obj client.Object) []reconcile.Request {
	dep, ok := resourceFromObject(obj)
	if !ok {
		return nil
	}
	var list corev1.ConfigMapList
	if err := c.client.List(ctx, &list, client.InNamespace(dep.Namespace), client.MatchingLabels(managedLabels(KindJob))); err != nil {
		log.FromContext(ctx).Error(err, "list jobs failed")
		return nil
	}
	var reqs []reconcile.Request
	for i := range list.Items {
		job := resourceFromCM(&list.Items[i])
		if jobDependsOn(&job, &dep) {
			reqs = append(reqs, reconcile.Request{NamespacedName: client.ObjectKeyFromObject(&list.Items[i])})
		}
	}
	return reqs
}

func (c *jobController) setup(mgr ctrl.Manager) error {
	dataChanged := predicate.ResourceVersionChangedPredicate{}
	dependents := handler.EnqueueRequestsFromMapFunc(c.enqueueDependentJobs)

	return ctrl.NewControllerManagedBy(mgr).
		Named("gitops-job").
		For(&corev1.ConfigMap{}, builder.WithPredicates(kindPredicate(KindJob))).
		Watches(&corev1.ConfigMap{}, dependents, builder.WithPredicates(kindPredicate(KindSource))).
		Watches(&corev1.ConfigMap{}, dependents, builder.WithPredicates(kindPredicate(KindDestination))).
		Watches(&corev1.ConfigMap{}, dependents, builder.WithPredicates(kindPredicate(KindStreams), dataChanged)).
		Watches(&corev1.Secret{}, dependents, builder.WithPredicates(kindPredicate(KindSource))).
		Watches(&corev1.Secret{}, dependents, builder.WithPredicates(kindPredicate(KindDestination))).
		Complete(c)
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
