// Package secretref resolves secret references in connector configs.
//
// A reference is a whole JSON string value of the form ${<provider>:<path>[#<key>]}, for example
// "${aws-sm:olake/prod/pg#password}". The value is fetched from the provider and, when #<key> is
// given, read from that key of a JSON-object secret, keeping that key's JSON type.
//
// Providers are created on first use, so a cloud SDK client is only built when a config
// references that provider.
package secretref

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"regexp"
	"slices"
	"strings"
	"sync"
)

// ErrUnresolvable marks an error that will not go away on retry: a malformed reference, an
// unsupported provider, a missing secret or key, or access denied.
var ErrUnresolvable = errors.New("secret reference cannot be resolved")

func unresolvable(format string, args ...any) error {
	return fmt.Errorf("%w: %s", ErrUnresolvable, fmt.Sprintf(format, args...))
}

// Provider fetches the raw value of one secret. path is the part of the reference between the
// provider name and "#"; each provider defines its format.
type Provider interface {
	Fetch(ctx context.Context, path string) (string, error)
}

type newProviderFunc func(ctx context.Context, cfg *Config) (Provider, error)

var supportedProviders = map[string]newProviderFunc{
	"aws-sm":   newAWSProvider,
	"gcp-sm":   newGCPProvider,
	"azure-kv": newAzureProvider,
	"secrets":  newK8sProvider, // k8s secrets
	"env":      newEnvProvider, // environment variables for docker environment
}

type Config struct {
	AWS   AWSConfig
	GCP   GCPConfig
	Azure AzureConfig
	K8s   K8sConfig
}

var refPattern = regexp.MustCompile(`^\$\{([a-z][a-z0-9-]*):([^#}]+)(?:#([^}]*))?\}$`)

// Resolver is safe for concurrent use.
type Resolver struct {
	cfg Config

	mu        sync.Mutex
	providers map[string]Provider
}

func New(cfg *Config) *Resolver {
	return &Resolver{cfg: *cfg, providers: map[string]Provider{}}
}

// NewWithProviders is for tests: it resolves references without a cloud account.
func NewWithProviders(providers map[string]Provider) *Resolver {
	r := New(&Config{})
	maps.Copy(r.providers, providers)
	return r
}

// Resolve replaces every reference in the JSON config with its secret value and reports whether
// it found any. Errors name the reference, never a secret value.
func (r *Resolver) Resolve(ctx context.Context, config string) (string, bool, error) {
	if !strings.Contains(config, "${") {
		return config, false, nil
	}

	fields, err := decodeObject(config)
	if err != nil {
		return "", false, unresolvable("config is not a JSON object")
	}

	found := false
	if err := r.resolveFields(ctx, fields, &found); err != nil {
		return "", false, err
	}
	resolved, _ := json.Marshal(fields) // cannot fail: fields came from json decoding
	return string(resolved), found, nil
}

// resolveFields resolves references in fields and in nested objects, such as ssh_config.password
// in a source or writer.aws_secret_key in a destination.
func (r *Resolver) resolveFields(ctx context.Context, fields map[string]any, found *bool) error {
	for key, value := range fields {
		switch value := value.(type) {
		case map[string]any:
			if err := r.resolveFields(ctx, value, found); err != nil {
				return err
			}
		case string:
			resolved, err := r.resolveValue(ctx, value, found)
			if err != nil {
				return err
			}
			fields[key] = resolved
		}
	}
	return nil
}

// A #key value keeps its JSON type from the secret, so {"port":5432} gives the number 5432.
func (r *Resolver) resolveValue(ctx context.Context, ref string, found *bool) (any, error) {
	match := refPattern.FindStringSubmatch(ref)
	if match == nil {
		return ref, nil
	}
	providerName, path, key := match[1], match[2], match[3]
	provider, err := r.loadProvider(ctx, providerName)
	if err != nil {
		return nil, fmt.Errorf("resolve %s: %w", ref, err)
	}
	secret, err := provider.Fetch(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("resolve %s: %w", ref, err)
	}
	*found = true
	if !strings.Contains(ref, "#") {
		return secret, nil
	}
	value, err := readKey(secret, key)
	if err != nil {
		return nil, fmt.Errorf("resolve %s: %w", ref, err)
	}
	return value, nil
}

// decodeObject decodes a JSON object, keeping numbers exactly as written.
func decodeObject(data string) (map[string]any, error) {
	decoder := json.NewDecoder(strings.NewReader(data))
	decoder.UseNumber()
	var fields map[string]any
	err := decoder.Decode(&fields)
	return fields, err
}

// readKey reads key from a JSON-object secret. Errors never quote the secret.
func readKey(secret, key string) (any, error) {
	fields, err := decodeObject(secret)
	if err != nil {
		return nil, unresolvable("secret is not a JSON object, so #%s cannot be read", key)
	}
	value, ok := fields[key]
	if !ok {
		return nil, unresolvable("key %q not found in secret", key)
	}
	return value, nil
}

// loadProvider creates the provider on first use. A failed creation is not kept, so the next
// call tries again. Two callers may create it at once; either result is a working client.
func (r *Resolver) loadProvider(ctx context.Context, name string) (Provider, error) {
	r.mu.Lock()
	provider, ok := r.providers[name]
	r.mu.Unlock()
	if ok {
		return provider, nil
	}

	newProvider, ok := supportedProviders[name]
	if !ok {
		return nil, unresolvable("unsupported secret provider %q (supported: %s)", name,
			strings.Join(slices.Sorted(maps.Keys(supportedProviders)), ", "))
	}
	provider, err := newProvider(ctx, &r.cfg)
	if err != nil {
		return nil, fmt.Errorf("create %s provider: %w", name, err)
	}

	r.mu.Lock()
	r.providers[name] = provider
	r.mu.Unlock()
	return provider, nil
}
