package secretref

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azsecrets"
)

// AzureConfig configures the azure-kv provider. Credentials come from the default chain (AKS
// Workload Identity, managed identity, env) unless all three client-secret fields are set.
type AzureConfig struct {
	TenantID     string
	ClientID     string
	ClientSecret string
}

// azureProvider reads Azure Key Vault. path is <vault-name>/<secret-name>[/<version>]; the latest
// version is read unless one is given.
type azureProvider struct {
	credential azcore.TokenCredential

	mu      sync.Mutex
	clients map[string]*azsecrets.Client // different vaults have different endpoints, so one client per vault
}

func newAzureProvider(_ context.Context, cfg Config) (Provider, error) {
	var credential azcore.TokenCredential
	var err error
	// for key based auth, all three must be set; otherwise use the managed identity
	if cfg.Azure.TenantID != "" && cfg.Azure.ClientID != "" && cfg.Azure.ClientSecret != "" {
		credential, err = azidentity.NewClientSecretCredential(cfg.Azure.TenantID, cfg.Azure.ClientID, cfg.Azure.ClientSecret, nil)
	} else {
		credential, err = azidentity.NewDefaultAzureCredential(nil)
	}
	if err != nil {
		return nil, fmt.Errorf("create Azure credential: %w", err)
	}
	return &azureProvider{credential: credential, clients: map[string]*azsecrets.Client{}}, nil
}

func (p *azureProvider) Fetch(ctx context.Context, path string) (string, error) {
	vault, name, version, err := azureSecretName(path)
	if err != nil {
		return "", err
	}
	client, err := p.client(vault)
	if err != nil {
		return "", err
	}
	resp, err := client.GetSecret(ctx, name, version, nil)
	if err != nil {
		return "", classifyAzureError(err)
	}
	if resp.Value == nil {
		return "", nil
	}
	return *resp.Value, nil
}

// client returns the client for a vault, creating it on first use. Clients share one credential.
func (p *azureProvider) client(vault string) (*azsecrets.Client, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if client, ok := p.clients[vault]; ok {
		return client, nil
	}
	client, err := azsecrets.NewClient(fmt.Sprintf("https://%s.vault.azure.net/", vault), p.credential, nil)
	if err != nil {
		return nil, fmt.Errorf("create Key Vault client for %s: %w", vault, err)
	}
	p.clients[vault] = client
	return client, nil
}

// azureSecretName splits path into vault, secret name and version ("" for the latest).
func azureSecretName(path string) (vault, name, version string, err error) {
	parts := strings.SplitN(path, "/", 3)
	if len(parts) < 2 {
		return "", "", "", unresolvable("%q is not <vault-name>/<secret-name>[/<version>]", path)
	}
	if len(parts) == 3 {
		version = parts[2]
	}
	return parts[0], parts[1], version, nil
}

// classifyAzureError marks a 4xx response as unresolvable (not found, forbidden, disabled secret),
// except throttling. Everything else is retried: 5xx, and network or token failures.
func classifyAzureError(err error) error {
	var respErr *azcore.ResponseError
	if errors.As(err, &respErr) && respErr.StatusCode >= 400 && respErr.StatusCode < 500 &&
		respErr.StatusCode != http.StatusTooManyRequests {
		return fmt.Errorf("%w: %w", ErrUnresolvable, err)
	}
	return err
}
