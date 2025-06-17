package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
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
		return nil, fmt.Errorf("failed to fetch tags: %s", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("docker hub api request failed with status code: %d", resp.StatusCode)
	}

	var responseData DockerHubTagsResponse
	if err := json.NewDecoder(resp.Body).Decode(&responseData); err != nil {
		return nil, fmt.Errorf("failed to decode response: %s", err)
	}

	tags := make([]string, 0, len(responseData.Results))
	for _, tagData := range responseData.Results {
		if !strings.Contains(tagData.Name, "stag") && !strings.Contains(tagData.Name, "latest") && !strings.Contains(tagData.Name, "dev") {
			tags = append(tags, tagData.Name)
		}
	}

	return tags, nil
}
