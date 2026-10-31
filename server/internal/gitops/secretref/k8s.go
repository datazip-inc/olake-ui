package secretref

import (
	"context"
	"encoding/json"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// K8sConfig configures the secrets provider, which reads Secrets in olake-ui's namespace. Only
// Secrets that match Selector (olake.io/managed=true) are readable, so a reference cannot read any
// other Secret in the namespace, whatever Client returns. It is empty in file mode.
type K8sConfig struct {
	Client    client.Reader
	Namespace string
	Selector  labels.Selector
}

// k8sProvider returns a Secret's data as a JSON object, so ${secrets:<name>#<key>} reads one key.
type k8sProvider struct {
	K8sConfig
}

func newK8sProvider(_ context.Context, cfg *Config) (Provider, error) {
	if cfg.K8s.Client == nil || cfg.K8s.Selector == nil {
		return nil, unresolvable("secrets references need GITOPS_MODE=kubernetes; in file mode use ${env:...}")
	}
	return &k8sProvider{cfg.K8s}, nil
}

func (p *k8sProvider) Fetch(ctx context.Context, name string) (string, error) {
	var secret corev1.Secret
	err := p.Client.Get(ctx, client.ObjectKey{Namespace: p.Namespace, Name: name}, &secret)
	if apierrors.IsNotFound(err) || (err == nil && !p.Selector.Matches(labels.Set(secret.Labels))) {
		return "", unresolvable("Secret %q not found in namespace %s, or not labelled %s", name, p.Namespace, p.Selector)
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
