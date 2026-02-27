package utils

import (
	"context"
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
	"github.com/beego/beego/v2/server/web"
	"github.com/datazip-inc/olake-ui/server/internal/constants"
	"github.com/datazip-inc/olake-ui/server/utils/logger"
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
	"OLAKE_SECRET_KEY":        nil,
	"_":                       nil,
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
	repositoryBase, err := web.AppConfig.String(constants.ConfContainerRegistryBase)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get CONTAINER_REGISTRY_BASE: %s", err)
	}
	var tags []string
	images := []string{imageName}
	if imageName == "" {
		images = defaultImages
	}
	driverImage := ""
	for _, imageName := range images {
		if strings.Contains(repositoryBase, "ecr") {
			fullImage := fmt.Sprintf("%s/%s", repositoryBase, imageName)
			tags, err = getECRImageTags(ctx, fullImage)
		} else if isGCRArtifactRegistry(repositoryBase) {
			fullImage := fmt.Sprintf("%s/%s", repositoryBase, imageName)
			tags, err = getGCRArtifactRegistryImageTags(ctx, fullImage)
		} else {
			tags, err = getDockerHubImageTags(ctx, imageName)
		}

		// Fallback to cached if online fetch fails or explicitly requested
		if err != nil && cachedTags {
			if constants.ExecutorEnvironment == "kubernetes" {
				logger.Warn("failed to fetch image tags online for %s: %s. Cached fallback unavailable on Kubernetes (no Docker daemon)", imageName, err)
				continue
			}
			logger.Warn("failed to fetch image tags online for %s: %s, falling back to cached tags", imageName, err)
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
	resp, err := http.DefaultClient.Do(req) // #nosec G704 -- URL is built from a compile-time constant (dockerHubTagsURLTemplate)
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
	if strings.Contains(repositoryBase, "ecr") || isGCRArtifactRegistry(repositoryBase) {
		// after making it ecr, it will be like "123456789012.dkr.ecr.us-west-2.amazonaws.com/olakego/source-mysql"
		// from gcr, it will be like "us-docker.pkg.dev/my-project/my-repo/olakego/source-mysql"
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
