package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"sort"
	"strings"

	"github.com/beego/beego/v2/core/logs"
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

// GetDockerHubTags fetches all valid tags for a given Docker image from Docker Hub or falls back to local images
func GetDockerHubTags(ctx context.Context, imageName string) ([]string, error) {
	tags, err := fetchTagsFromDockerHub(imageName)
	if err != nil {
		logs.Debug("Warning: %s. Falling back to local images.", err)
		return getLocalImageTags(ctx, imageName)
	}
	return tags, nil
}

// fetchTagsFromDockerHub tries to fetch tags from Docker Hub
func fetchTagsFromDockerHub(imageName string) ([]string, error) {
	resp, err := http.Get(fmt.Sprintf(dockerHubTagsURLTemplate, imageName))
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

// getLocalImageTags fetches all valid local tags for a given image
func getLocalImageTags(ctx context.Context, imageName string) ([]string, error) {
	images, err := getLocalDockerData(ctx, imageName+":")
	if err != nil {
		return nil, err
	}

	var tags []string
	for _, image := range images {
		// Expected format: imageName:tag
		parts := strings.Split(image, ":")
		if len(parts) != 2 {
			continue
		}
		tag := parts[1]
		if isValidTag(tag) {
			tags = append(tags, tag)
		}
	}

	// Sort tags in descending order
	sort.Slice(tags, func(i, j int) bool {
		return tags[i] > tags[j] // '>' for descending order
	})

	return tags, nil
}

// GetAvailableDriversVersions returns the driver with the highest available version locally
func GetAvailableDriversVersions(ctx context.Context) (string, string) {
	tags, err := GetDockerHubTags(ctx, "olakego/source-postgres")
	if err == nil && len(tags) > 0 {
		return "postgres", tags[0]
	}
	logs.Debug("Falling back to local source images due to error: %s", err)
	images, err := getLocalDockerData(ctx, "olakego/source-")
	if err != nil {
		logs.Debug("Failed to fetch local source images: %s", err)
		return "", ""
	}

	result := make(map[string]string)

	for _, image := range images {
		// Expected format: olakego/source-postgres:1.2.3
		parts := strings.Split(strings.TrimPrefix(image, "olakego/source-"), ":")
		if len(parts) != 2 {
			continue
		}
		driver := parts[0]
		version := parts[1]

		if !isValidTag(version) {
			continue
		}

		// Keep the highest version per driver
		if existing, ok := result[driver]; !ok || version > existing {
			result[driver] = version
		}
	}

	// Find the driver with the highest version overall
	var maxDriver, maxVersion string
	for driver, version := range result {
		if version > maxVersion {
			maxDriver = driver
			maxVersion = version
		}
	}

	return maxDriver, maxVersion
}

// getLocalDockerData returns filtered docker images output based on provided prefix
func getLocalDockerData(ctx context.Context, filterPrefix string) ([]string, error) {
	cmd := exec.CommandContext(ctx, "docker", "images", "--format", "{{.Repository}}:{{.Tag}}")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list docker images: %s", err)
	}

	var results []string
	for _, line := range strings.Split(strings.TrimSpace(string(output)), "\n") {
		if strings.HasPrefix(line, filterPrefix) {
			results = append(results, line)
		}
	}
	return results, nil
}

// isValidTag centralizes tag filtering logic
func isValidTag(tag string) bool {
	return tag != "<none>" &&
		!strings.Contains(tag, "stag") &&
		!strings.Contains(tag, "latest") &&
		!strings.Contains(tag, "dev") &&
		tag >= "v0.1.0"
}
