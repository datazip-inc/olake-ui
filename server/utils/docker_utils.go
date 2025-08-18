package utils

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"regexp"
	"sort"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/beego/beego/v2/core/logs"
	"github.com/beego/beego/v2/server/web"
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

var defaultImages = []string{"olakego/source-mysql", "olakego/source-postgres", "olakego/source-oracle", "olakego/source-mongodb"}

// GetDriverImageTags returns image tags from ECR or Docker Hub with fallback to cached images
func GetDriverImageTags(ctx context.Context, imageName string, cachedTags bool) ([]string, string, error) {
	// TODO: make constants file and validate all env vars in start of server
	repositoryBase, err := web.AppConfig.String("CONTAINER_REGISTRY_BASE")
	if err != nil {
		return nil, "", fmt.Errorf("failed to get CONTAINER_REGISTRY_BASE: %v", err)
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
		} else {
			tags, err = getDockerHubImageTags(ctx, imageName)
		}

		// Fallback to cached if online fetch fails or explicitly requested
		if err != nil && cachedTags {
			logs.Warn("failed to fetch image tags online for %s: %s, falling back to cached tags", imageName, err)
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
	for _, imageDetail := range resp.ImageDetails {
		for _, tag := range imageDetail.ImageTags {
			if isValidTag(tag) {
				tags = append(tags, tag)
			}
		}
	}

	sort.Slice(tags, func(i, j int) bool { return tags[i] > tags[j] })
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
	sort.Slice(tags, func(i, j int) bool { return tags[i] > tags[j] })
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
	sort.Slice(tags, func(i, j int) bool { return tags[i] > tags[j] })
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
