package gitops

import (
	"context"
	"errors"
	"fmt"
	"math/rand/v2"
	"strings"
	"time"

	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/datazip-inc/olake-ui/server/internal/appconfig"
	"github.com/datazip-inc/olake-ui/server/internal/gitops/secretref"
)

// secretRefs resolves secret references in source/destination configs, and re-checks resources
// that use them every refresh interval so a rotated secret reaches OLake without a change in git.
type secretRefs struct {
	resolver *secretref.Resolver
	// refresh is how often a Ready resource with references is re-checked; 0 turns this off.
	refresh time.Duration
	// encrypted is true when OLAKE_SECRET_KEY is set. Resolved values are stored in the
	// database, so references are refused without it rather than stored as plaintext.
	encrypted bool
}

func newSecretRefs(cfg appconfig.Config) *secretRefs {
	return &secretRefs{
		resolver: secretref.New(secretref.Config{
			AWS: secretref.AWSConfig{
				Region:          cfg.GitOpsSecretsAWSRegion,
				AccessKeyID:     cfg.GitOpsSecretsAWSAccessKeyID,
				SecretAccessKey: cfg.GitOpsSecretsAWSSecretKey,
			},
			GCP: secretref.GCPConfig{
				Project: cfg.GitOpsSecretsGCPProject,
			},
		}),
		refresh:   cfg.GitOpsSecretsRefreshInterval,
		encrypted: cfg.EncryptionKey != "",
	}
}

// mayRefresh is a cheap check of the raw config: false means it cannot hold a reference, so a
// synced resource can be skipped without parsing it.
func (s *secretRefs) mayRefresh(config string) bool {
	return s.refresh > 0 && strings.Contains(config, "${")
}

// resolve returns config with every reference replaced and whether it had any. Errors that will
// not go away on retry are marked non-retryable.
func (s *secretRefs) resolve(ctx context.Context, config string) (string, bool, error) {
	resolved, found, err := s.resolver.Resolve(ctx, config)
	if err != nil {
		if errors.Is(err, secretref.ErrUnresolvable) {
			return "", false, NonRetryableError(err)
		}
		return "", false, err
	}
	// TODO[BEFORE_MERGE]: check if OLAKE_SECRET_KEY is coupled with secret reference.
	if found && !s.encrypted {
		return "", false, NonRetryableError(fmt.Errorf("secret references require OLAKE_SECRET_KEY, so resolved secrets are encrypted in the database"))
	}
	return resolved, found, nil
}

// next schedules the next re-check of a resource with references, with up to 10% jitter so
// resources synced together do not call the secret managers at the same moment.
func (s *secretRefs) next(hasRefs bool) ctrl.Result {
	if !hasRefs || s.refresh <= 0 {
		return ctrl.Result{}
	}
	return ctrl.Result{RequeueAfter: s.refresh + rand.N(s.refresh/10)}
}
