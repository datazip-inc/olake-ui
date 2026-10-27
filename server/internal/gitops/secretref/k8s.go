package secretref

import (
	"context"
	"encoding/json"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// K8sConfig configures the secrets provider, which reads Secrets in olake-ui's namespace. The
// client is the GitOps cache, which holds only Secrets labelled olake.io/managed=true, so a
// reference cannot read any other Secret in the namespace. It is empty in Docker mode.
type K8sConfig struct {
	Client    client.Reader
	Namespace string
}

// k8sProvider returns a Secret's data as a JSON object, so ${secrets:<name>#<key>} reads one key.
type k8sProvider struct {
	K8sConfig
}

func newK8sProvider(_ context.Context, cfg Config) (Provider, error) {
	if cfg.K8s.Client == nil {
		return nil, unresolvable("secrets references need Kubernetes mode; in Docker mode use ${env:...}")
	}
	return &k8sProvider{cfg.K8s}, nil
}

func (p *k8sProvider) Fetch(ctx context.Context, name string) (string, error) {
	var secret corev1.Secret
	err := p.Client.Get(ctx, client.ObjectKey{Namespace: p.Namespace, Name: name}, &secret)
	if apierrors.IsNotFound(err) {
		return "", unresolvable("Secret %q not found in namespace %s, or not labelled olake.io/managed=true", name, p.Namespace)
	}
	if err != nil {
		return "", err
	}
	data := make(map[string]string, len(secret.Data))
	for key, value := range secret.Data {
		data[key] = string(value)
	}
	encoded, _ := json.Marshal(data) // cannot fail: a map of strings
	return string(encoded), nil
}
