package secretref

import (
	"context"
	"os"
)

// envProvider reads any variable in olake-ui's environment
type envProvider struct{}

func newEnvProvider(_ context.Context, cfg *Config) (Provider, error) {
	if cfg.K8s.Client != nil {
		return nil, unresolvable("env references are not allowed with OLAKE_GITOPS_MODE=kubernetes; use ${secrets:<name>#<key>}")
	}
	return envProvider{}, nil
}

func (envProvider) Fetch(_ context.Context, name string) (string, error) {
	value, ok := os.LookupEnv(name)
	if !ok {
		return "", unresolvable("environment variable %s is not set", name)
	}
	return value, nil
}
