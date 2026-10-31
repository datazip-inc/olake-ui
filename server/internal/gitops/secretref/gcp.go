package secretref

import (
	"context"
	"fmt"
	"strings"

	"cloud.google.com/go/compute/metadata"
	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GCPConfig configures the gcp-sm provider. Credentials come from Application Default
// Credentials (GKE Workload Identity, attached service account, GOOGLE_APPLICATION_CREDENTIALS).
type GCPConfig struct {
	Project string
}

// gcpProvider reads GCP Secret Manager. path is a secret id in the default project, or
// projects/<project>/secrets/<id>[/versions/<version>]. The latest version is read unless one is given.
type gcpProvider struct {
	client  *secretmanager.Client
	project string
}

func newGCPProvider(ctx context.Context, cfg *Config) (Provider, error) {
	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("create Secret Manager client: %w", err)
	}
	project := cfg.GCP.Project
	if project == "" && metadata.OnGCE() {
		// On GKE or GCE the metadata server knows the project the workload runs in. A failed lookup
		// fails the provider creation, which is retried, rather than caching a provider without one.
		project, err = metadata.ProjectIDWithContext(ctx)
		if err != nil {
			_ = client.Close()
			return nil, fmt.Errorf("read GCP project from the metadata server: %w", err)
		}
	}
	return &gcpProvider{client: client, project: project}, nil
}

func (p *gcpProvider) Fetch(ctx context.Context, path string) (string, error) {
	name, err := gcpVersionName(path, p.project)
	if err != nil {
		return "", err
	}
	resp, err := p.client.AccessSecretVersion(ctx, &secretmanagerpb.AccessSecretVersionRequest{Name: name})
	if err != nil {
		return "", classifyGCPError(err)
	}
	return string(resp.GetPayload().GetData()), nil
}

// gcpVersionName returns the full secret version name for path.
func gcpVersionName(path, project string) (string, error) {
	if !strings.HasPrefix(path, "projects/") {
		if project == "" {
			return "", unresolvable("GCP project is not set: set GITOPS_SECRETS_GCP_PROJECT or use projects/<project>/secrets/%s", path)
		}
		path = fmt.Sprintf("projects/%s/secrets/%s", project, path)
	}
	if !strings.Contains(path, "/versions/") {
		path += "/versions/latest"
	}
	return path, nil
}

// classifyGCPError marks errors that will not go away on retry as unresolvable. The client has
// already retried Unavailable and similar codes.
func classifyGCPError(err error) error {
	switch status.Code(err) {
	case codes.NotFound, codes.PermissionDenied, codes.Unauthenticated, codes.InvalidArgument, codes.FailedPrecondition:
		return fmt.Errorf("%w: %w", ErrUnresolvable, err)
	default:
		return err
	}
}
