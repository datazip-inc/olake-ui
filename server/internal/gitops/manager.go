package gitops

import (
	"context"
	"errors"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/datazip-inc/olake-ui/server/internal/services"
	"github.com/datazip-inc/olake-ui/server/internal/utils/logger"
)

var scheme = runtime.NewScheme()

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
}

// Controller runs the embedded GitOps reconcilers.
type Controller struct {
	mgr ctrl.Manager
	app *services.AppService
}

// InitGitOps starts GitOps when enabled. In-cluster uses the ConfigMap
// controller-runtime manager. Outside the cluster, GITOPS_FILE_DIR must be set.
func InitGitOps(ctx context.Context, enabled bool, fileDir string, app *services.AppService) error {
	if !enabled {
		return nil
	}

	restConfig, err := rest.InClusterConfig()
	if err == nil {
		ctrlr, setupErr := Setup(app, restConfig)
		if setupErr != nil {
			return setupErr
		}
		go func() {
			if runErr := ctrlr.Run(ctx); runErr != nil && !errors.Is(runErr, context.Canceled) {
				logger.Errorf("GitOps controller manager stopped: %s", runErr)
			}
		}()
		return nil
	}

	if fileDir == "" {
		return fmt.Errorf("gitops enabled but not in-cluster and GITOPS_FILE_DIR is empty: %w", err)
	}
	return startFileMode(ctx, fileDir, app)
}

func startFileMode(ctx context.Context, fileDir string, app *services.AppService) error {
	fw := newFileWatcher(fileDir, app.ETL(), app.ETL().Temporal())
	go func() {
		if err := fw.Start(ctx); err != nil && !errors.Is(err, context.Canceled) {
			logger.Errorf("GitOps file watcher stopped: %s", err)
		}
	}()
	logger.Infof("GitOps file watcher watching %s", fileDir)
	return nil
}

// Setup validates GitOps prerequisites synchronously.
func Setup(app *services.AppService, restConfig *rest.Config) (*Controller, error) {
	ctrl.SetLogger(logger.Logr())

	managed := labels.SelectorFromSet(labels.Set{LabelManaged: LabelManagedValue})

	mgr, err := ctrl.NewManager(restConfig, ctrl.Options{
		Scheme: scheme,
		Cache: cache.Options{
			ByObject: map[client.Object]cache.ByObject{
				&corev1.ConfigMap{}: {Label: managed},
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("create controller manager: %w", err)
	}

	logger.Info("GitOps controller manager configured for labelled ConfigMap watch")
	return &Controller{mgr: mgr, app: app}, nil
}

// Run registers reconcilers and blocks until ctx is cancelled.
func (c *Controller) Run(ctx context.Context) error {
	etlSvc := c.app.ETL()
	sink := NewK8sSink(c.mgr.GetClient(), c.mgr.GetEventRecorderFor("olake-gitops"), etlSvc.Temporal())

	if err := (&SourceReconciler{
		Client: c.mgr.GetClient(),
		ETL:    etlSvc,
		Sink:   sink,
	}).Setup(c.mgr); err != nil {
		return fmt.Errorf("setup source controller: %w", err)
	}

	if err := (&DestinationReconciler{
		Client: c.mgr.GetClient(),
		ETL:    etlSvc,
		Sink:   sink,
	}).Setup(c.mgr); err != nil {
		return fmt.Errorf("setup destination controller: %w", err)
	}

	if err := (&JobReconciler{
		Client: c.mgr.GetClient(),
		ETL:    etlSvc,
		Sink:   sink,
	}).Setup(c.mgr); err != nil {
		return fmt.Errorf("setup job controller: %w", err)
	}

	logger.Info("starting GitOps controller manager")
	return c.mgr.Start(ctx)
}
