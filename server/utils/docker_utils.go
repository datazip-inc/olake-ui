package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"sort"
	"strings"
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

// return source tags, source, and error
func GetDriverImageTags(ctx context.Context, imageName string, cachedTags bool) ([]string, error) {
	fetchCachedImageTags := func(ctx context.Context, imageName string) ([]string, error) {
		imagePrefix := Ternary(imageName != "", fmt.Sprintf("%s:", imageName), "olakego/source-").(string)
		images, err := GetCachedImages(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get cached images: %s", err)
		}

		tagsMap := make(map[string]struct{})
		for _, image := range images {
			if strings.HasPrefix(image, imagePrefix) {
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

		// Sort tags in descending order
		sort.Slice(tags, func(i, j int) bool {
			return tags[i] > tags[j] // '>' for descending order
		})

		return tags, nil
	}

	fetchTagsFromDockerHub := func(ctx context.Context, imageName string) ([]string, error) {
		// use default postgres if empty
		imageName = Ternary(imageName == "", "olakego/source-postgres", imageName).(string)
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

	tags, err := fetchTagsFromDockerHub(ctx, imageName)
	if err != nil {
		if cachedTags {
			// check for cached images on local
			return fetchCachedImageTags(ctx, imageName)
		}
		return nil, err
	}
	return tags, nil
}

func GetCachedImages(ctx context.Context) ([]string, error) {
	cmd := exec.CommandContext(ctx, "docker", "images", "--format", "{{.Repository}}:{{.Tag}}")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list docker images: %s", err)
	}

	return strings.Split(strings.TrimSpace(string(output)), "\n"), nil
}

// isValidTag centralizes tag filtering logic
func isValidTag(tag string) bool {
	return tag != "<none>" &&
		!strings.Contains(tag, "stag") &&
		!strings.Contains(tag, "latest") &&
		!strings.Contains(tag, "dev") &&
		tag >= "v0.1.0"
}
