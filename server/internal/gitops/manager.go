package gitops

import (
	"context"
	"errors"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"

	olakev1 "github.com/datazip-inc/olake-ui/server/internal/gitops/api/v1"
	"github.com/datazip-inc/olake-ui/server/internal/services"
	"github.com/datazip-inc/olake-ui/server/internal/utils/logger"
)

var scheme = runtime.NewScheme()

func init() {
	// Registers Kubernetes built-in types (Pods, Secrets, ConfigMaps, Namespaces, etc.) into scheme
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	// Registers OLake custom types into scheme
	utilruntime.Must(olakev1.AddToScheme(scheme))
}

// Controller runs the embedded GitOps reconcilers.
type Controller struct {
	mgr ctrl.Manager
	app *services.AppService
}

// InitGitOps validates GitOps prerequisites when enabled and starts the controller
// manager in a background goroutine. Returns an error on synchronous setup failure.
func InitGitOps(ctx context.Context, enabled bool, app *services.AppService) error {
	if !enabled {
		return nil
	}

	ctrl, err := Setup(app)
	if err != nil {
		return err
	}

	go func() {
		if err := ctrl.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
			logger.Errorf("GitOps controller manager stopped: %s", err)
		}
	}()

	return nil
}

// Setup validates GitOps prerequisites synchronously. Fails fast on misconfiguration
// (in-cluster config, manager creation) before the HTTP server starts.
func Setup(app *services.AppService) (*Controller, error) {
	restConfig, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("in-cluster kubernetes config: %w", err)
	}

	ctrl.SetLogger(logger.Logr())

	// Cluster-scoped cache; requires GitOps ClusterRole (olake.io resources only).
	mgr, err := ctrl.NewManager(restConfig, ctrl.Options{
		Scheme: scheme,
	})
	if err != nil {
		return nil, fmt.Errorf("create controller manager: %w", err)
	}

	logger.Info("GitOps controller manager configured for cluster-wide watch")
	return &Controller{mgr: mgr, app: app}, nil
}

// Run registers reconcilers and blocks until ctx is cancelled. Call from a goroutine.
func (c *Controller) Run(ctx context.Context) error {
	etlSvc := c.app.ETL()

	if err := (&SourceReconciler{
		Client: c.mgr.GetClient(),
		ETL:    etlSvc,
	}).Setup(c.mgr); err != nil {
		return fmt.Errorf("setup source controller: %w", err)
	}

	if err := (&DestinationReconciler{
		Client: c.mgr.GetClient(),
		ETL:    etlSvc,
	}).Setup(c.mgr); err != nil {
		return fmt.Errorf("setup destination controller: %w", err)
	}

	if err := (&JobReconciler{
		Client: c.mgr.GetClient(),
		ETL:    etlSvc,
	}).Setup(c.mgr); err != nil {
		return fmt.Errorf("setup job controller: %w", err)
	}

	logger.Info("starting GitOps controller manager")
	return c.mgr.Start(ctx)
}
