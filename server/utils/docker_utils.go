package utils

import (
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
func GetDockerHubTags(imageName string) ([]string, error) {
	tags, err := fetchTagsFromDockerHub(imageName)
	if err != nil {
		logs.Debug("Warning: %s. Falling back to local images.", err)
		return getLocalImageTags(imageName)
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
		return nil, fmt.Errorf("Docker Hub API returned status %d", resp.StatusCode)
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
func getLocalImageTags(imageName string) ([]string, error) {
	cmd := exec.Command("docker", "images", imageName, "--format", "{{.Tag}}")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list tags for image %s: %v", imageName, err)
	}
	var tags []string
	for _, tag := range strings.Split(strings.TrimSpace(string(output)), "\n") {
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
func GetAvailableDriversVersions() (string, string) {
	drivers := []string{"postgres", "mysql", "oracle", "mongodb"}
	result := make(map[string]string)

	for _, driver := range drivers {
		imageName := fmt.Sprintf("olakego/source-%s", driver)
		versions, err := getLocalImageTags(imageName)
		if err != nil || len(versions) == 0 {
			continue
		}

		result[driver] = versions[len(versions)-1]
	}

	var maxDriver, maxVersion string
	for driver, version := range result {
		if version > maxVersion {
			maxDriver = driver
			maxVersion = version
		}
	}

	return maxDriver, maxVersion
}

// isValidTag centralizes tag filtering logic
func isValidTag(tag string) bool {
	return tag != "<none>" &&
		!strings.Contains(tag, "stag") &&
		!strings.Contains(tag, "latest") &&
		!strings.Contains(tag, "dev") &&
		tag >= "v0.1.0"
}
