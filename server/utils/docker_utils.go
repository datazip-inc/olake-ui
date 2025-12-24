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
	"github.com/datazip-inc/olake-ui/server/internal/models/dto"
	"github.com/datazip-inc/olake-ui/server/utils/logger"
	"golang.org/x/mod/semver"
)

// docker hub tags api url template
const dockerHubTagsURLTemplate = "https://hub.docker.com/v2/repositories/%s/tags/?page_size=100"

// docker hub registry manifest url template
const dockerHubManifestURLTemplate = "https://registry-1.docker.io/v2/%s/manifests/%s"

// docker hub registry blob url template
const dockerHubBlobURLTemplate = "https://registry-1.docker.io/v2/%s/blobs/%s"

// docker hub auth url template
const dockerHubAuthURLTemplate = "https://auth.docker.io/token?service=registry.docker.io&scope=repository:%s:pull"

// github releases url template
const githubReleasesURLTemplate = "https://github.com/datazip-inc/%s/releases/tag/%s"

// DockerHubTag represents a single tag from Docker Hub API response
type DockerHubTag struct {
	Name string `json:"name"`
}

// DockerHubTagsResponse represents the response structure from Docker Hub tags API
type DockerHubTagsResponse struct {
	Results []DockerHubTag `json:"results"`
}

var defaultImages = []string{"olakego/source-mysql", "olakego/source-postgres", "olakego/source-oracle", "olakego/source-mongodb", "olakego/source-kafka"}

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
		} else {
			tags, err = getDockerHubImageTags(ctx, imageName)
		}

		// Fallback to cached if online fetch fails or explicitly requested
		if err != nil && cachedTags {
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
	for _, imageDetail := range resp.ImageDetails {
		for _, tag := range imageDetail.ImageTags {
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

// GetReleaseDataForType fetches release data for a specific Docker repository type
func GetReleaseDataForType(
	ctx context.Context,
	repo string,
	releaseType string,
	currentVersion string,
	limit int,
	onlyNewerVersions bool,
) (*dto.ReleaseTypeData, error) {
	// Fetch token once and reuse for all registry API calls
	token, err := getDockerHubToken(ctx, repo)
	if err != nil {
		return nil, fmt.Errorf("failed to get auth token: %w", err)
	}

	tags, err := getDockerHubImageTags(ctx, repo)
	if err != nil {
		return nil, err
	}

	sort.Slice(tags, func(i, j int) bool {
		return semver.Compare(tags[i], tags[j]) > 0
	})

	releases := make([]*dto.ReleaseMetadataResponse, 0)

	for _, tag := range tags {
		if limit > 0 && len(releases) >= limit {
			break
		}

		// For olake_ui_worker → only versions >= current
		if onlyNewerVersions {
			if semver.Compare(tag, currentVersion) < 0 {
				continue
			}
		}

		info, err := fetchDockerHubReleaseMetadata(ctx, repo, tag, releaseType, currentVersion, token)
		if err != nil {
			continue // skip bad tags, don't fail whole endpoint
		}

		releases = append(releases, info)
	}

	if len(releases) == 0 {
		return nil, fmt.Errorf("no releases found")
	}

	releaseData := &dto.ReleaseTypeData{
		Releases: releases,
	}

	// Only set CurrentVersion for olake_ui_worker
	if onlyNewerVersions {
		releaseData.CurrentVersion = currentVersion
	}

	return releaseData, nil
}

// FetchDockerHubReleaseMetadata fetches release metadata from Docker Hub registry
func fetchDockerHubReleaseMetadata(
	ctx context.Context,
	repo string,
	tag string,
	releaseType string,
	currentVersion string,
	token string,
) (*dto.ReleaseMetadataResponse, error) {
	configDigest, err := getConfigDigestFromManifest(ctx, repo, tag, token)
	if err != nil {
		return nil, fmt.Errorf("failed to get config digest: %w", err)
	}

	configBlob, err := fetchConfigBlob(ctx, repo, configDigest, token)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch config blob: %w", err)
	}

	description, err := extractDescription(configBlob.Config.Labels)
	if err != nil {
		return nil, fmt.Errorf("failed to extract description: %s", err)
	}

	tags := extractReleaseTags(configBlob.Config.Labels, releaseType, tag, currentVersion)

	repoName := repo
	if idx := strings.LastIndex(repo, "/"); idx >= 0 {
		repoName = repo[idx+1:]
	}

	return &dto.ReleaseMetadataResponse{
		Version:     tag,
		Description: description,
		Tags:        tags,
		Date:        configBlob.Created,
		Link:        fmt.Sprintf(githubReleasesURLTemplate, repoName, tag),
	}, nil
}

// getDockerHubToken fetches an authentication token from Docker Hub for a repository
func getDockerHubToken(ctx context.Context, repo string) (string, error) {
	authURL := fmt.Sprintf(dockerHubAuthURLTemplate, repo)
	req, err := http.NewRequestWithContext(ctx, "GET", authURL, http.NoBody)
	if err != nil {
		return "", fmt.Errorf("create auth request: %s", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("fetch auth token: %s", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("auth request failed with status: %d", resp.StatusCode)
	}

	var authResponse struct {
		Token string `json:"token"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&authResponse); err != nil {
		return "", fmt.Errorf("decode auth response: %s", err)
	}

	if authResponse.Token == "" {
		return "", fmt.Errorf("auth token is empty")
	}

	return authResponse.Token, nil
}

// getConfigDigestFromManifest fetches manifest list and returns config digest from first platform manifest
func getConfigDigestFromManifest(ctx context.Context, repo, tag, token string) (string, error) {
	// Fetch manifest list
	manifestListURL := fmt.Sprintf(dockerHubManifestURLTemplate, repo, tag)
	req, err := http.NewRequestWithContext(ctx, "GET", manifestListURL, http.NoBody)
	if err != nil {
		return "", fmt.Errorf("create manifest list request: %s", err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("fetch manifest list: %s", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("manifest list request failed with status: %d", resp.StatusCode)
	}

	var manifestList struct {
		Manifests []struct {
			Digest string `json:"digest"`
		} `json:"manifests"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&manifestList); err != nil {
		return "", fmt.Errorf("decode manifest list: %s", err)
	}

	if len(manifestList.Manifests) == 0 {
		return "", fmt.Errorf("manifest list is empty")
	}

	// Use first manifest (config metadata is same across platforms for multi-arch images)
	firstManifestDigest := manifestList.Manifests[0].Digest

	// Fetch platform-specific manifest
	platformManifestURL := fmt.Sprintf(dockerHubManifestURLTemplate, repo, firstManifestDigest)
	platformReq, err := http.NewRequestWithContext(ctx, "GET", platformManifestURL, http.NoBody)
	if err != nil {
		return "", fmt.Errorf("create platform manifest request: %s", err)
	}
	platformReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	platformResp, err := http.DefaultClient.Do(platformReq)
	if err != nil {
		return "", fmt.Errorf("fetch platform manifest: %s", err)
	}
	defer platformResp.Body.Close()

	if platformResp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("platform manifest request failed with status: %d", platformResp.StatusCode)
	}

	var platformManifest struct {
		Config struct {
			Digest string `json:"digest"`
		} `json:"config"`
	}

	if err := json.NewDecoder(platformResp.Body).Decode(&platformManifest); err != nil {
		return "", fmt.Errorf("decode platform manifest: %s", err)
	}

	if platformManifest.Config.Digest == "" {
		return "", fmt.Errorf("platform manifest config digest is empty")
	}

	return platformManifest.Config.Digest, nil
}

// configBlob represents a Docker image configuration blob
type configBlob struct {
	Created string           `json:"created"`
	Config  configBlobConfig `json:"config"`
}

type configBlobConfig struct {
	Labels map[string]string `json:"Labels"`
}

// fetchConfigBlob fetches the configuration blob from Docker Hub
func fetchConfigBlob(ctx context.Context, repo, digest, token string) (*configBlob, error) {
	url := fmt.Sprintf(dockerHubBlobURLTemplate, repo, digest)

	req, err := http.NewRequestWithContext(ctx, "GET", url, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("create request: %s", err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute request: %s", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var blob configBlob
	if err := json.NewDecoder(resp.Body).Decode(&blob); err != nil {
		return nil, fmt.Errorf("decode response: %s", err)
	}

	return &blob, nil
}

// extractDescription extracts and processes description from image labels.
// Description is expected to be base64 encoded to support multi-line markdown content.
// Returns an error if the description cannot be decoded from base64.
func extractDescription(labels map[string]string) (string, error) {
	if labels == nil {
		return "No description available", nil
	}

	desc := labels["description"]
	if desc == "" {
		return "No description available", nil
	}

	// Decode base64 encoded description (required for multi-line markdown support)
	decoded, err := base64.StdEncoding.DecodeString(desc)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64 description: %s", err)
	}

	return string(decoded), nil
}

// extractReleaseTags extracts release tags from image labels and merges with computed tags.
// Reads comma-separated tags from release-tags label
// and adds "new-release" tag for olake_ui_worker versions greater than current.
func extractReleaseTags(labels map[string]string, releaseType, tag, currentVersion string) []string {
	tags := []string{}

	// Extract tags from Docker label (comma-separated)
	if labels != nil {
		if releaseTagsStr := labels["release-tags"]; releaseTagsStr != "" {
			releaseTags := strings.Split(releaseTagsStr, ",")
			for _, t := range releaseTags {
				trimmed := strings.TrimSpace(t)
				if trimmed != "" {
					tags = append(tags, trimmed)
				}
			}
		}
	}

	// Add "new-release" tag for olake_ui_worker versions greater than current
	if releaseType == "olake_ui_worker" && semver.Compare(tag, currentVersion) > 0 {
		// Check if "new-release" already exists to avoid duplicates
		hasNewRelease := false
		for _, t := range tags {
			if t == "new-release" {
				hasNewRelease = true
				break
			}
		}
		if !hasNewRelease {
			tags = append(tags, "new-release")
		}
	}

	return tags
}
