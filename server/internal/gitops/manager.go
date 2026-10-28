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
)

type runner interface {
	Run(ctx context.Context) error
}

// InitGitOps sets up the GitOps mode named by mode and runs it in the background.
func InitGitOps(ctx context.Context, mode string, app *services.AppService) error {
	if mode == "" {
		return nil
	}

	ctrl.SetLogger(logger.Logr())
	r, err := newRunner(mode, app)
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

func newRunner(mode string, app *services.AppService) (runner, error) {
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
	default:
		return nil, fmt.Errorf("unknown GITOPS_MODE %s, want %s", mode, ModeKubernetes)
	}
}
