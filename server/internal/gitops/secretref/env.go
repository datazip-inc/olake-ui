package secretref

import (
	"context"
	"os"
)

// envProvider reads any variable in olake-ui's environment
type envProvider struct{}

func newEnvProvider(context.Context, *Config) (Provider, error) {
	return envProvider{}, nil
}

func (envProvider) Fetch(_ context.Context, name string) (string, error) {
	value, ok := os.LookupEnv(name)
	if !ok {
		return "", unresolvable("environment variable %s is not set", name)
	}
	return value, nil
}
