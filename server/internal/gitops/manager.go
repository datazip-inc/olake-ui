package gitops

import (
	"context"
	"errors"
	"fmt"

	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/datazip-inc/olake-ui/server/internal/services"
	"github.com/datazip-inc/olake-ui/server/internal/utils/logger"
)

// GITOPS_MODE values.
const (
	ModeKubernetes = "kubernetes" // labelled ConfigMaps/Secrets in the pod's namespace
	ModeFile       = "file"       // JSON manifests in OLAKE_GITOPS_FILE_DIR
)

type runner interface {
	Run(ctx context.Context) error
}

// Config selects and configures the GitOps mode.
type Config struct {
	Mode      string // GITOPS_MODE; empty disables GitOps
	Namespace string // Kubernetes mode: the pod's namespace, which it watches
	FileDir   string // file mode: the manifest directory
	StateFile string // file mode: the status file
}

// InitGitOps sets up the GitOps mode named by cfg.Mode and runs it in the background.
func InitGitOps(ctx context.Context, cfg Config, app *services.AppService) error {
	if cfg.Mode == "" {
		return nil
	}

	ctrl.SetLogger(logger.Logr())
	r, err := newRunner(cfg, app)
	if err != nil {
		return fmt.Errorf("gitops %s mode: %w", cfg.Mode, err)
	}
	go func() {
		if err := r.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
			logger.Errorf("GitOps %s mode stopped: %s", cfg.Mode, err)
		}
	}()
	return nil
}

func newRunner(cfg Config, app *services.AppService) (runner, error) {
	switch cfg.Mode {
	case ModeKubernetes:
		restConfig, err := rest.InClusterConfig()
		if err != nil {
			return nil, fmt.Errorf("load in-cluster config: %w", err)
		}
		ctrlr, err := Setup(app, restConfig, cfg.Namespace)
		if err != nil {
			return nil, err
		}
		return ctrlr, nil
	case ModeFile:
		if cfg.FileDir == "" {
			return nil, fmt.Errorf("OLAKE_GITOPS_FILE_DIR is empty")
		}
		fw := newFileWatcher(cfg.FileDir, cfg.StateFile, app.ETL(), app.ETL().Temporal())
		if err := fw.setup(); err != nil {
			return nil, err
		}
		logger.Infof("GitOps file watcher watching %s", cfg.FileDir)
		return fw, nil
	default:
		return nil, fmt.Errorf("unknown GITOPS_MODE %s, want %s or %s", cfg.Mode, ModeKubernetes, ModeFile)
	}
}
