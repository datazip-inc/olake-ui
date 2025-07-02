package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"strings"

	"github.com/beego/beego/v2/core/logs"
)

// DockerHubTag represents a single tag from Docker Hub API response
type DockerHubTag struct {
	Name string `json:"name"`
}

// DockerHubTagsResponse represents the response structure from Docker Hub tags API
type DockerHubTagsResponse struct {
	Results []DockerHubTag `json:"results"`
}

// GetDockerHubTags fetches all tags for a given Docker image from Docker Hub
// imageName should be in the format "namespace/repository" (e.g., "library/nginx")
func GetDockerHubTags(imageName string) ([]string, error) {
	url := fmt.Sprintf("https://hub.docker.com/v2/repositories/%s/tags/?page_size=100", imageName)
	resp, err := http.Get(url)
	if err != nil {
		logs.Debug("Warning: Failed to fetch tags from Docker Hub, falling back to local images: %v\n", err)
		return getLocalImageTags(imageName)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logs.Debug("Warning: Docker Hub API returned status %d, falling back to local images\n", resp.StatusCode)
		return getLocalImageTags(imageName)
	}

	var responseData DockerHubTagsResponse
	if err := json.NewDecoder(resp.Body).Decode(&responseData); err != nil {
		logs.Debug("Warning: Failed to decode Docker Hub response, falling back to local images: %v\n", err)
		return getLocalImageTags(imageName)
	}

	tags := make([]string, 0, len(responseData.Results))
	for _, tagData := range responseData.Results {
		if !strings.Contains(tagData.Name, "stag") && !strings.Contains(tagData.Name, "latest") && !strings.Contains(tagData.Name, "dev") && tagData.Name >= "v0.1.0" {
			tags = append(tags, tagData.Name)
		}
	}

	return tags, nil
}

func getLocalImageTags(imageName string) ([]string, error) {
	cmd := exec.Command("docker", "images", "--format", "{{.Repository}}:{{.Tag}}")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list local docker images: %v", err)
	}

	lines := strings.Split(string(output), "\n")
	tags := make([]string, 0)

	for _, line := range lines {
		if strings.HasPrefix(line, imageName+":") {
			tag := strings.TrimPrefix(line, imageName+":")
			if tag == "<none>" {
				continue
			}
			if !strings.Contains(tag, "stag") && !strings.Contains(tag, "latest") && !strings.Contains(tag, "dev") && tag >= "v0.1.0" {
				tags = append(tags, tag)
			}
		}
	}

	return tags, nil
}
