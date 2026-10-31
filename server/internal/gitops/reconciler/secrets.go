package reconciler

import (
	"context"
	"errors"
	"math/rand/v2"
	"strings"
	"time"

	"github.com/datazip-inc/olake-ui/server/internal/appconfig"
	"github.com/datazip-inc/olake-ui/server/internal/gitops/secretref"
	"github.com/datazip-inc/olake-ui/server/internal/utils/logger"
)

// minSecretRefresh is the shortest refresh interval used. A value without a unit, such as
// GITOPS_SECRETS_REFRESH_INTERVAL=60, is parsed as nanoseconds and would call the secret managers
// in a tight loop.
const minSecretRefresh = time.Minute

// SecretRefs resolves secret references in source/destination configs, and re-checks resources
// that use them every refresh interval so a rotated secret reaches OLake without a change in git.
// Each GitOps mode builds one with NewSecretRefs and gives it to the Source and Destination reconcilers.
type SecretRefs struct {
	resolver *secretref.Resolver
	// refresh is how often a Ready resource with references is re-checked; 0 turns this off.
	refresh time.Duration
}

// NewSecretRefs builds the resolver from the GITOPS_SECRETS_* settings. k8s is the cluster that
// ${secrets:...} references read from; its zero value refuses them (file mode without a cluster).
func NewSecretRefs(k8s secretref.K8sConfig) *SecretRefs {
	cfg := appconfig.Load()
	return &SecretRefs{
		resolver: secretref.New(&secretref.Config{
			AWS: secretref.AWSConfig{
				Region:          cfg.GitOpsSecretsAWSRegion,
				AccessKeyID:     cfg.GitOpsSecretsAWSAccessKeyID,
				SecretAccessKey: cfg.GitOpsSecretsAWSSecretKey,
			},
			GCP: secretref.GCPConfig{
				Project: cfg.GitOpsSecretsGCPProject,
			},
			Azure: secretref.AzureConfig{
				TenantID:     cfg.GitOpsSecretsAzureTenantID,
				ClientID:     cfg.GitOpsSecretsAzureClientID,
				ClientSecret: cfg.GitOpsSecretsAzureClientSecret,
			},
			K8s: k8s,
		}),
		refresh: refreshInterval(cfg.GitOpsSecretsRefreshInterval),
	}
}

// refreshInterval raises a GITOPS_SECRETS_REFRESH_INTERVAL below minSecretRefresh to it; 0 stays 0.
func refreshInterval(refresh time.Duration) time.Duration {
	if refresh > 0 && refresh < minSecretRefresh {
		logger.Warnf("GITOPS_SECRETS_REFRESH_INTERVAL %s is below %s (a value needs a unit, such as 15m); using %s", refresh, minSecretRefresh, minSecretRefresh)
		return minSecretRefresh
	}
	return refresh
}

// mayRefresh is a cheap check of the raw config: false means it cannot hold a reference, so a
// synced resource can be skipped without parsing it.
func (s *SecretRefs) mayRefresh(config string) bool {
	return s.refresh > 0 && strings.Contains(config, "${")
}

// resolve returns config with every reference replaced and whether it had any. Errors that will
// not go away on retry are marked non-retryable.
func (s *SecretRefs) resolve(ctx context.Context, config string) (string, bool, error) {
	resolved, found, err := s.resolver.Resolve(ctx, config)
	if err != nil {
		if errors.Is(err, secretref.ErrUnresolvable) {
			return "", false, NonRetryableError(err)
		}
		return "", false, err
	}
	return resolved, found, nil
}

// next is the delay to the next re-check of a resource with references, with up to 10% jitter so
// resources synced together do not call the secret managers at the same moment. 0 means no re-check.
func (s *SecretRefs) next(hasRefs bool) time.Duration {
	if !hasRefs || s.refresh <= 0 {
		return 0
	}
	return s.refresh + rand.N(s.refresh/10) // #nosec G404 -- jitter only spreads load, it needs no crypto randomness
}
