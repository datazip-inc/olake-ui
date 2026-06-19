package utils

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strings"

	artifactregistry "cloud.google.com/go/artifactregistry/apiv1"
	"cloud.google.com/go/artifactregistry/apiv1/artifactregistrypb"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/datazip-inc/olake-ui/server/internal/appconfig"
	"github.com/datazip-inc/olake-ui/server/internal/constants"
	"github.com/datazip-inc/olake-ui/server/internal/utils/logger"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/spf13/viper"
	"golang.org/x/mod/semver"
	"google.golang.org/api/iterator"
)

// docker hub tags api url template
const dockerHubTagsURLTemplate = "https://hub.docker.com/v2/repositories/%s/tags/?page_size=100"
const ecrRepositoryPrivateRegex = `^(\d+)\.dkr\.ecr\.([a-z0-9-]+)\.amazonaws\.com(\.cn)?/(.+)$`
const ecrRepositoryPublicRegex = `^public\.ecr\.aws/(.+)$`
const gcrArtifactRegistryRepositoryRegex = `^([a-z][a-z0-9-]*)-docker\.pkg\.dev/([^/]+)/([^/]+)/(.+)$`

// DockerHubTag represents a single tag from Docker Hub API response
type DockerHubTag struct {
	Name string `json:"name"`
}

// DockerHubTagsResponse represents the response structure from Docker Hub tags API
type DockerHubTagsResponse struct {
	Results []DockerHubTag `json:"results"`
}

var defaultImages = []string{"olakego/source-mysql", "olakego/source-postgres", "olakego/source-oracle", "olakego/source-mongodb", "olakego/source-kafka", "olakego/source-s3", "olakego/source-db2", "olakego/source-mssql"}

// ignoredWorkerEnv is a map of environment variables that are ignored from the worker container.
var ignoredWorkerEnv = map[string]any{ // A map is chosen because it gives O(1) lookup time for key existence.
	"HOSTNAME":                nil,
	"PATH":                    nil,
	"PWD":                     nil,
	"HOME":                    nil,
	"SHLVL":                   nil,
	"TERM":                    nil,
	"PERSISTENT_DIR":          nil,
	"CONTAINER_REGISTRY_BASE": nil,
	"TEMPORAL_ADDRESS":        nil,
	"TEMPORAL_NAMESPACE":      nil,
	"TEMPORAL_ENABLE_TLS":     nil,
	"TEMPORAL_API_KEY":        nil,
	"TEMPORAL_EXTERNAL":       nil,
	"TEMPORAL_TASK_QUEUE":     nil,
	"OLAKE_SECRET_KEY":        nil,
	"_":                       nil,
	// Registry credentials/TLS settings must not leak into driver containers.
	constants.EnvRegistryUsername:      nil,
	constants.EnvRegistryPassword:      nil,
	constants.EnvRegistryInsecure:      nil,
	constants.EnvRegistryTLSSkipVerify: nil,
	constants.EnvRegistryCACert:        nil,
}

// GetWorkerEnvVars returns the environment variables from the worker container.
func GetWorkerEnvVars() map[string]string {
	vars := make(map[string]string)
	for _, entry := range os.Environ() {
		parts := strings.SplitN(entry, "=", 2)
		key := parts[0]
		if _, ignore := ignoredWorkerEnv[key]; ignore {
			continue
		}
		vars[key] = parts[1]
	}
	return vars
}

// GetDriverImageTags returns image tags from ECR, Artifact Registry, or Docker Hub with fallback to cached images
func GetDriverImageTags(ctx context.Context, imageName string, cachedTags bool) ([]string, string, error) {
	// TODO: make constants file and validate all env vars in start of server
	repositoryBase := appconfig.Load().ContainerRegistryBase
	if repositoryBase == "" {
		return nil, "", fmt.Errorf("failed to get CONTAINER_REGISTRY_BASE")
	}

	var err error
	var tags []string
	images := []string{imageName}
	if imageName == "" {
		images = defaultImages
	}
	driverImage := ""
	for _, imageName := range images {
		switch detectRegistryType(repositoryBase) {
		case registryECR:
			fullImage := fmt.Sprintf("%s/%s", strings.TrimSuffix(repositoryBase, "/"), imageName)
			tags, err = getECRImageTags(ctx, fullImage)
		case registryGCR:
			fullImage := fmt.Sprintf("%s/%s", strings.TrimSuffix(repositoryBase, "/"), imageName)
			tags, err = getGCRArtifactRegistryImageTags(ctx, fullImage)
		case registryDockerHub:
			tags, err = getDockerHubImageTags(ctx, imageName)
		default: // registryGeneric: any Docker Registry v2 compliant registry (Harbor, Nexus, Quay, GitLab, registry:2)
			fullImage := fmt.Sprintf("%s/%s", strings.TrimSuffix(repositoryBase, "/"), imageName)
			tags, err = getGenericRegistryImageTags(ctx, fullImage)
		}

		// Fallback to cached if online fetch fails or explicitly requested
		if err != nil && cachedTags {
			if constants.ExecutorEnvironment == "kubernetes" {
				logger.Warnf("failed to fetch image tags online for %s: %s. Cached fallback unavailable on Kubernetes (no Docker daemon)", imageName, err)
				continue
			}
			logger.Warnf("failed to fetch image tags online for %s: %s, falling back to cached tags", imageName, err)
			tags, err = fetchCachedImageTags(ctx, imageName, repositoryBase)
			if err != nil {
				return nil, "", fmt.Errorf("failed to fetch cached image tags for %s: %s", imageName, err)
			}
		}

		if len(tags) == 0 {
			// if no tags found continue
			continue
		}

		// TODO : return highest tag out of all sources (currently breaking loop once any tag found on any image)
		driverImage = imageName
		break
	}

	// Add custom version if provided
	if customVersion := GetCustomDriverVersion(); customVersion != "" {
		tags = append(tags, customVersion)
	}

	if len(tags) == 0 {
		return nil, "", fmt.Errorf("no tags found for image: %s", imageName)
	}
	driverImage = strings.TrimPrefix(driverImage, "olakego/source-")
	sort.Slice(tags, func(i, j int) bool { return semver.Compare(tags[i], tags[j]) > 0 }) // highest first
	return tags, driverImage, err
}

// getECRImageTags fetches tags from AWS ECR
func getECRImageTags(ctx context.Context, fullImageName string) ([]string, error) {
	accountID, region, repoName, err := ParseECRDetails(fullImageName)
	if err != nil {
		return nil, fmt.Errorf("invalid ECR URI: %s", err)
	}

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %s", err)
	}

	client := ecr.NewFromConfig(cfg)
	resp, err := client.DescribeImages(ctx, &ecr.DescribeImagesInput{
		RepositoryName: aws.String(repoName),
		RegistryId:     aws.String(accountID),
		MaxResults:     aws.Int32(100),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch ECR tags: %s", err)
	}

	var tags []string
	for i := range resp.ImageDetails {
		for _, tag := range resp.ImageDetails[i].ImageTags {
			if isValidTag(tag) {
				tags = append(tags, tag)
			}
		}
	}
	return tags, nil
}

// getDockerHubImageTags fetches tags from Docker Hub
func getDockerHubImageTags(ctx context.Context, imageName string) ([]string, error) {
	// Create a new HTTP request with context
	req, err := http.NewRequestWithContext(ctx, "GET",
		fmt.Sprintf(dockerHubTagsURLTemplate, imageName), http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %s", err)
	}
	// Make the HTTP request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch tags from Docker Hub: %s", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("docker hub api request failed with status code: %d", resp.StatusCode)
	}

	var responseData DockerHubTagsResponse
	if err := json.NewDecoder(resp.Body).Decode(&responseData); err != nil {
		return nil, fmt.Errorf("failed to decode Docker Hub response: %s", err)
	}

	var tags []string
	for _, tagData := range responseData.Results {
		if isValidTag(tagData.Name) {
			tags = append(tags, tagData.Name)
		}
	}
	return tags, nil
}

// registryType classifies the configured CONTAINER_REGISTRY_BASE so tag listing
// can be routed to the right client.
type registryType int

const (
	registryDockerHub registryType = iota // Docker Hub (default); listed via the Docker Hub web API
	registryECR                           // AWS ECR; listed via the AWS SDK using IAM
	registryGCR                           // Google Artifact Registry; listed via the GCP SDK using ADC
	registryGeneric                       // any other Docker Registry v2 compliant registry (Harbor, Nexus, Quay, GitLab, registry:2)
)

// detectRegistryType classifies the registry base. Empty and the Docker Hub hosts
// map to Docker Hub (preserving the shipped default registry-1.docker.io), so existing
// deployments are unaffected. Classification is based on the registry host only, so
// look-alike hosts (e.g. Azure's *.azurecr.io, which contains the substring "ecr")
// are not misrouted to the AWS ECR path.
func detectRegistryType(base string) registryType {
	base = strings.TrimSpace(strings.TrimSuffix(base, "/"))
	if base == "" {
		return registryDockerHub
	}
	host := base
	if i := strings.IndexByte(base, '/'); i >= 0 {
		host = base[:i]
	}
	switch {
	case host == "docker.io", host == "registry-1.docker.io", host == "index.docker.io":
		return registryDockerHub
	case isECRRegistry(host):
		return registryECR
	case isGCRArtifactRegistry(host):
		return registryGCR
	}
	return registryGeneric
}

// isECRRegistry reports whether the host is an AWS ECR registry: a private
// "<account>.dkr.ecr.<region>.amazonaws.com[.cn]" host, or "public.ecr.aws".
// This deliberately matches the AWS domain rather than the substring "ecr" so that
// other registries (e.g. *.azurecr.io) are not misclassified.
func isECRRegistry(host string) bool {
	return strings.HasPrefix(host, "public.ecr.aws") ||
		(strings.Contains(host, ".ecr.") && strings.Contains(host, "amazonaws.com"))
}

// getGenericRegistryImageTags lists tags from any Docker Registry HTTP API v2
// compliant registry (Harbor, Nexus, Quay, GitLab, registry:2). The underlying
// library transparently handles the bearer-token auth challenge and Link-header
// pagination.
func getGenericRegistryImageTags(ctx context.Context, fullImageName string) ([]string, error) {
	nameOpts := []name.Option{name.WeakValidation}
	if appconfig.Load().ContainerRegistryInsecure {
		nameOpts = append(nameOpts, name.Insecure) // allow plain-HTTP registries (e.g. host:5000)
	}
	repo, err := name.NewRepository(fullImageName, nameOpts...)
	if err != nil {
		return nil, fmt.Errorf("invalid registry image reference %q: %s", fullImageName, err)
	}

	tr, err := registryTransport()
	if err != nil {
		return nil, err
	}
	tags, err := remote.List(repo,
		remote.WithContext(ctx),
		registryAuthOption(),
		remote.WithTransport(tr),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list tags from registry %q: %s", repo.String(), err)
	}

	var valid []string
	for _, t := range tags {
		if isValidTag(t) {
			valid = append(valid, t)
		}
	}
	return valid, nil
}

// registryAuthOption uses explicit credentials when provided, otherwise falls back to
// the standard docker keychain (~/.docker/config.json via DOCKER_CONFIG), which itself
// resolves to anonymous when no credentials are configured.
func registryAuthOption() remote.Option {
	cfg := appconfig.Load()
	user := cfg.ContainerRegistryUsername
	pass := cfg.ContainerRegistryPassword
	if user != "" || pass != "" {
		return remote.WithAuth(&authn.Basic{Username: user, Password: pass})
	}
	return remote.WithAuthFromKeychain(authn.DefaultKeychain)
}

// registryTransport clones the default transport and applies optional TLS overrides
// for on-prem registries with self-signed certificates or a private CA.
func registryTransport() (http.RoundTripper, error) {
	cfg := appconfig.Load()
	base := http.DefaultTransport.(*http.Transport).Clone()
	tlsCfg := &tls.Config{InsecureSkipVerify: cfg.ContainerRegistryTLSSkipVerify} //nolint:gosec // opt-in via CONTAINER_REGISTRY_TLS_SKIP_VERIFY for self-signed on-prem registries
	if caPEM := cfg.ContainerRegistryCACert; caPEM != "" {
		// The value is the PEM CA bundle contents itself, which is convenient for
		// env-driven configs and Kubernetes Secrets (no file to mount).
		pool, _ := x509.SystemCertPool()
		if pool == nil {
			pool = x509.NewCertPool()
		}
		if !pool.AppendCertsFromPEM([]byte(caPEM)) {
			return nil, fmt.Errorf("no valid CA certificates found in %s (expected PEM contents)", constants.EnvRegistryCACert)
		}
		tlsCfg.RootCAs = pool
	}
	base.TLSClientConfig = tlsCfg
	return base, nil
}

// isGCRArtifactRegistry reports whether the registry base refers to Google Artifact Registry (*-docker.pkg.dev).
func isGCRArtifactRegistry(registryBase string) bool {
	return strings.Contains(registryBase, "docker.pkg.dev")
}

// getGCRArtifactRegistryImageTags fetches tags from Google Artifact Registry using the native SDK.
// Authentication is handled via Google Application Default Credentials.
func getGCRArtifactRegistryImageTags(ctx context.Context, fullImageName string) ([]string, error) {
	project, location, repository, packageName, err := ParseGCRArtifactRegistryDetails(fullImageName)
	if err != nil {
		return nil, fmt.Errorf("invalid Artifact Registry URI: %s", err)
	}

	client, err := artifactregistry.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create Artifact Registry client: %s", err)
	}
	defer client.Close()

	// Build the parent path for listing tags
	parent := fmt.Sprintf("projects/%s/locations/%s/repositories/%s/packages/%s", project, location, repository, packageName)

	req := &artifactregistrypb.ListTagsRequest{
		Parent: parent,
	}

	var tags []string
	it := client.ListTags(ctx, req)
	for {
		tag, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to fetch tags from Artifact Registry: %s", err)
		}

		// Extract tag name from the full resource name
		// Format: projects/{project}/locations/{location}/repositories/{repository}/packages/{package}/tags/{tag}
		parts := strings.Split(tag.Name, "/")
		if len(parts) > 0 {
			tagName := parts[len(parts)-1]
			if isValidTag(tagName) {
				tags = append(tags, tagName)
			}
		}
	}

	return tags, nil
}

// ParseGCRArtifactRegistryDetails extracts project, location, repository, and package name
// from an Artifact Registry Docker image URI.
//
// Example:
//
//	Input:  "us-docker.pkg.dev/my-project/my-repo/olakego/source-mysql:v1.0.0"
//	Output: project     = "my-project"
//	        location    = "us"
//	        repository  = "my-repo"
//	        packageName = "olakego/source-mysql"
//
// The package name is URL-encoded for the API (e.g., "olakego%2Fsource-mysql")
func ParseGCRArtifactRegistryDetails(fullImageName string) (project, location, repository, packageName string, err error) {
	// Remove tag if present
	imageRef := strings.SplitN(fullImageName, ":", 2)[0]

	// Format: {location}-docker.pkg.dev/{project}/{repository}/{package-path}
	arRe := regexp.MustCompile(gcrArtifactRegistryRepositoryRegex)
	if matches := arRe.FindStringSubmatch(imageRef); len(matches) == 5 {
		location = matches[1]
		project = matches[2]
		repository = matches[3]
		packagePath := matches[4]
		// URL encode the package path (forward slashes become %2F)
		packageName = strings.ReplaceAll(packagePath, "/", "%2F")
		return project, location, repository, packageName, nil
	}

	return "", "", "", "", fmt.Errorf("failed to parse Artifact Registry URI: %s", fullImageName)
}

// fetchCachedImageTags retrieves locally cached tags for an image
func fetchCachedImageTags(ctx context.Context, imageName, repositoryBase string) ([]string, error) {
	if detectRegistryType(repositoryBase) != registryDockerHub {
		// Non-Docker-Hub registries store images host-prefixed in `docker images`, e.g.
		// ECR:     "123456789012.dkr.ecr.us-west-2.amazonaws.com/olakego/source-mysql"
		// GCR:     "us-docker.pkg.dev/my-project/my-repo/olakego/source-mysql"
		// generic: "harbor.corp.com/olake/olakego/source-mysql"
		imageName = fmt.Sprintf("%s/%s", strings.TrimSuffix(repositoryBase, "/"), imageName)
	}

	images, err := GetCachedImages(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get cached images: %s", err)
	}

	tagsMap := make(map[string]struct{})
	for _, image := range images {
		if strings.HasPrefix(image, imageName) {
			parts := strings.Split(image, ":")
			if len(parts) != 2 || !isValidTag(parts[1]) {
				continue
			}
			tagsMap[parts[1]] = struct{}{}
		}
	}

	var tags []string
	for tag := range tagsMap {
		tags = append(tags, tag)
	}
	return tags, nil
}

// GetCachedImages retrieves locally cached Docker images
func GetCachedImages(ctx context.Context) ([]string, error) {
	cmd := exec.CommandContext(ctx, "docker", "images", "--format", "{{.Repository}}:{{.Tag}}")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list docker images: %s", err)
	}

	return strings.Split(strings.TrimSpace(string(output)), "\n"), nil
}

// ParseECRDetails extracts account ID, region, and repository name from ECR URI
// Example:
//
//	Input:  "123456789012.dkr.ecr.us-west-2.amazonaws.com/olakego/source-mysql:latest"
//	Output: accountID = "123456789012"
//	        region    = "us-west-2"
//	        repoName  = "olakego/source-mysql:latest"
func ParseECRDetails(fullImageName string) (accountID, region, repoName string, err error) {
	// handle private and public ecr and china ecr
	privateRe := regexp.MustCompile(ecrRepositoryPrivateRegex)
	publicRe := regexp.MustCompile(ecrRepositoryPublicRegex)

	if matches := privateRe.FindStringSubmatch(fullImageName); len(matches) == 5 {
		return matches[1], matches[2], matches[4], nil
	}

	if matches := publicRe.FindStringSubmatch(fullImageName); len(matches) == 2 {
		// Public ECR doesn’t have accountID/region
		return "public", "global", matches[1], nil
	}

	return "", "", "", fmt.Errorf("failed to parse ECR URI: %s", fullImageName)
}

// isValidTag centralizes tag filtering logic
func isValidTag(tag string) bool {
	return tag != "<none>" &&
		!strings.Contains(tag, "stag") &&
		!strings.Contains(tag, "latest") &&
		!strings.Contains(tag, "dev") &&
		tag >= "v0.1.0"
}

// --- developer utils ---

// CustomDriverVersion returns the custom driver version used to test olake with olake-ui.
// Note: This is only for development/testing purposes.
// When a custom version is set, semver-based compatibility checks will bypassed.
func GetCustomDriverVersion() string {
	if GetAppEnv() == constants.AppEnvDevelopment {
		return viper.GetString(constants.EnvCustomDriverImage)
	}
	return ""
}

// GetAppEnv returns the application environment in normalized format
// Supported values: development, production
func GetAppEnv() string {
	return NormalizeString(viper.GetString(constants.EnvAppEnvironment))
}
