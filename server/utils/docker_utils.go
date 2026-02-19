package utils

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/beego/beego/v2/server/web"
	"github.com/datazip-inc/olake-ui/server/internal/constants"
	"github.com/datazip-inc/olake-ui/server/utils/logger"
	"golang.org/x/mod/semver"
)

// docker hub tags api url template
const dockerHubTagsURLTemplate = "https://hub.docker.com/v2/repositories/%s/tags/?page_size=100"

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

// GetDriverImageTags returns image tags from ECR or Docker Hub with fallback to cached images
func GetDriverImageTags(ctx context.Context, imageName string, cachedTags bool) ([]string, string, error) {
	// TODO: make constants file and validate all env vars in start of server
	repositoryBase, err := web.AppConfig.String(constants.ConfContainerRegistryBase)
	fmt.Printf("[LOG-SAP]repositoryBase: %s", repositoryBase)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get CONTAINER_REGISTRY_BASE: %s", err)
	}
	var tags []string
	images := []string{imageName}
	if imageName == "" {
		images = defaultImages
	}
	fmt.Printf("[LOG-SAP]images to check: %v", images)
	driverImage := ""
	for _, imageName := range images {
		if strings.Contains(repositoryBase, "ecr") {
			fullImage := fmt.Sprintf("%s/%s", repositoryBase, imageName)
			fmt.Printf("[LOG-SAP]fetching tags for ECR image: %s", fullImage)
			tags, err = getECRImageTags(ctx, fullImage)
			fmt.Printf("[LOG-SAP]tags from ECR for %s: %v, err: %v", fullImage, tags, err)
		} else {
			tags, err = getDockerHubImageTags(ctx, imageName)
			fmt.Printf("[LOG-SAP]tags from Docker Hub for %s: %v, err: %v", imageName, tags, err)
		}

		// Fallback to cached if online fetch fails or explicitly requested
		if err != nil && cachedTags {
			logger.Warn("failed to fetch image tags online for %s: %s, falling back to cached tags", imageName, err)
			tags, err = fetchCachedImageTags(ctx, imageName, repositoryBase)
			fmt.Printf("[LOG-SAP]tags from cache for %s: %v, err: %v", imageName, tags, err)
			if err != nil {
				// If cached fetch also fails (e.g. no Docker daemon in EKS),
				// try Docker Hub API as a last resort for the base image name
				logger.Warn("cached image fetch failed for %s: %s, trying Docker Hub API as last resort", imageName, err)
				baseImageName := imageName
				// Strip any ECR prefix to get the base olakego/source-* name
				if strings.Contains(repositoryBase, "ecr") {
					parts := strings.Split(imageName, "/")
					if len(parts) >= 2 {
						baseImageName = strings.Join(parts[len(parts)-2:], "/")
					}
				}
				tags, err = getDockerHubImageTags(ctx, baseImageName)
				fmt.Printf("[LOG-SAP]tags from Docker Hub fallback for %s: %v, err: %v", baseImageName, tags, err)
				if err != nil {
					return nil, "", fmt.Errorf("failed to fetch image tags for %s from all sources: %s", imageName, err)
				}
			}
		}

		if len(tags) == 0 {
			// if no tags found continue
			continue
		}

		// TODO : return highest tag out of all sources (currently breaking loop once any tag found on any image)
		driverImage = imageName
		fmt.Printf("[LOG-SAP]selected driver image: %s with tags: %v", driverImage, tags)
		break
	}
	fmt.Printf("[LOG-SAP]final selected driver image: %s with tags: %v", driverImage, tags)
	if len(tags) == 0 {
		return nil, "", fmt.Errorf("no tags found for image: %s", imageName)
	}
	driverImage = strings.TrimPrefix(driverImage, "olakego/source-")
	sort.Slice(tags, func(i, j int) bool { return semver.Compare(tags[i], tags[j]) > 0 }) // highest first
	fmt.Printf("[LOG-SAP]sorted tags: %v", tags)
	fmt.Printf("[LOG-SAP]returning driver image: %s with tags: %v", driverImage, tags)
	return tags, driverImage, err
}

// getECRImageTags fetches tags from AWS ECR
func getECRImageTags(ctx context.Context, fullImageName string) ([]string, error) {
	accountID, region, repoName, err := ParseECRDetails(fullImageName)
	fmt.Printf("[LOG-SAP]Parsed ECR details - accountID: %s, region: %s, repoName: %s, err: %v", accountID, region, repoName, err)
	if err != nil {
		return nil, fmt.Errorf("invalid ECR URI: %s", err)
	}

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	fmt.Printf("[LOG-SAP]Loaded AWS config for region %s, err: %v", region, err)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %s", err)
	}

	client := ecr.NewFromConfig(cfg)
	resp, err := client.DescribeImages(ctx, &ecr.DescribeImagesInput{
		RepositoryName: aws.String(repoName),
		RegistryId:     aws.String(accountID),
		MaxResults:     aws.Int32(100),
	})
	fmt.Printf("[LOG-SAP]DescribeImages response for %s: %v, err: %v", fullImageName, resp, err)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch ECR tags: %s", err)
	}

	var tags []string
	for i := range resp.ImageDetails {
		fmt.Printf("[LOG-SAP]Processing image details for %s: %v", fullImageName, resp.ImageDetails[i])
		for _, tag := range resp.ImageDetails[i].ImageTags {
			if isValidTag(tag) {
				fmt.Printf("[LOG-SAP]Valid tag found for %s: %s", fullImageName, tag)
				tags = append(tags, tag)
			}
		}
	}
	fmt.Printf("[LOG-SAP]Valid tags from ECR for %s: %v", fullImageName, tags)
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

// fetchCachedImageTags retrieves locally cached tags for an image
func fetchCachedImageTags(ctx context.Context, imageName, repositoryBase string) ([]string, error) {
	if strings.Contains(repositoryBase, "ecr") {
		// after making it ecr, it will be like "123456789012.dkr.ecr.us-west-2.amazonaws.com/olakego/source-mysql"
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

// isDockerAvailable checks if the Docker daemon is reachable
// by verifying the Docker socket exists. In EKS/Kubernetes,
// the container runtime is containerd and there is no Docker socket.
func isDockerAvailable() bool {
	// Check common Docker socket paths
	for _, socketPath := range []string{"/var/run/docker.sock", "/run/docker.sock"} {
		if _, err := os.Stat(socketPath); err == nil {
			return true
		}
	}
	return false
}

// GetCachedImages retrieves locally cached Docker images.
// Returns an error if the Docker daemon is not available (e.g. in EKS).
func GetCachedImages(ctx context.Context) ([]string, error) {
	if !isDockerAvailable() {
		return nil, fmt.Errorf("docker daemon not available (no docker socket found) - this is expected in EKS/Kubernetes environments")
	}

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
	privateRe := regexp.MustCompile(`^(\d+)\.dkr\.ecr\.([a-z0-9-]+)\.amazonaws\.com(\.cn)?/(.+)$`)
	publicRe := regexp.MustCompile(`^public\.ecr\.aws/(.+)$`)

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

// DockerLoginECR logs in to an AWS ECR repository using the AWS SDK
func DockerLoginECR(ctx context.Context, region, registryID string) error {
	// Load AWS credentials & config
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return fmt.Errorf("failed to load AWS config: %s", err)
	}

	client := ecr.NewFromConfig(cfg)

	// Get ECR authorization token
	authResp, err := client.GetAuthorizationToken(ctx, &ecr.GetAuthorizationTokenInput{
		RegistryIds: []string{registryID},
	})
	if err != nil {
		return fmt.Errorf("failed to get ECR authorization token: %s", err)
	}

	if len(authResp.AuthorizationData) == 0 {
		return fmt.Errorf("no authorization data received from ECR")
	}

	authData := authResp.AuthorizationData[0]

	// Decode token
	decodedToken, err := base64.StdEncoding.DecodeString(aws.ToString(authData.AuthorizationToken))
	if err != nil {
		return fmt.Errorf("failed to decode authorization token: %s", err)
	}

	parts := strings.SplitN(string(decodedToken), ":", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid authorization token format")
	}
	username := parts[0]
	password := parts[1]
	registryURL := aws.ToString(authData.ProxyEndpoint) // e.g., https://678819669750.dkr.ecr.ap-south-1.amazonaws.com

	// Perform docker login
	cmd := exec.CommandContext(ctx, "docker", "login", "-u", username, "--password-stdin", registryURL)
	cmd.Stdin = strings.NewReader(password)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("docker login failed: %s\nOutput: %s", err, output)
	}

	return nil
}
