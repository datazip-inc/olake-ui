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

// InitGitOps sets up the GitOps mode named by mode and runs it in the background.
func InitGitOps(ctx context.Context, mode, fileDir, stateFile string, app *services.AppService) error {
	if mode == "" {
		return nil
	}

	ctrl.SetLogger(logger.Logr())
	r, err := newRunner(mode, fileDir, stateFile, app)
	if err != nil {
		return fmt.Errorf("gitops %s mode: %w", mode, err)
	}
	go func() {
		if err := r.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
			logger.Errorf("GitOps %s mode stopped: %s", mode, err)
		}
	}()
	return nil
}

func newRunner(mode, fileDir, stateFile string, app *services.AppService) (runner, error) {
	switch mode {
	case ModeKubernetes:
		restConfig, err := rest.InClusterConfig()
		if err != nil {
			return nil, fmt.Errorf("load in-cluster config: %w", err)
		}
		ctrlr, err := Setup(app, restConfig)
		if err != nil {
			return nil, err
		}
		return ctrlr, nil
	case ModeFile:
		if fileDir == "" {
			return nil, fmt.Errorf("OLAKE_GITOPS_FILE_DIR is empty")
		}
		fw := newFileWatcher(fileDir, stateFile, app.ETL(), app.ETL().Temporal())
		if err := fw.setup(); err != nil {
			return nil, err
		}
		logger.Infof("GitOps file watcher watching %s", fileDir)
		return fw, nil
	default:
		return nil, fmt.Errorf("unknown GITOPS_MODE %s, want %s or %s", mode, ModeKubernetes, ModeFile)
	}
}
