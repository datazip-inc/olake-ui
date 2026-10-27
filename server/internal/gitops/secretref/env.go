package secretref

import (
	"context"
	"os"
	"strings"
)

// envPrefix limits ${env:...} to variables meant for it.
// TODO[BEFORE_MERGE]: check if we need this prefix check
const envPrefix = "GITOPS_SECRET_"

// envProvider reads olake-ui's environment, for Docker mode where there are no cluster Secrets.
type envProvider struct{}

func newEnvProvider(context.Context, Config) (Provider, error) {
	return envProvider{}, nil
}

func (envProvider) Fetch(_ context.Context, name string) (string, error) {
	if !strings.HasPrefix(name, envPrefix) {
		return "", unresolvable("env reference %q must start with %s", name, envPrefix)
	}
	value, ok := os.LookupEnv(name)
	if !ok {
		return "", unresolvable("environment variable %s is not set", name)
	}
	return value, nil
}
