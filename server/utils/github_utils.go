package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/datazip-inc/olake-ui/server/internal/models/dto"
	"github.com/datazip-inc/olake-ui/server/utils/logger"
	"golang.org/x/mod/semver"
)

const (
	githubAPITimeout = 10 * time.Second

	githubReleasesURLTemplate = "https://api.github.com/repos/datazip-inc/%s/releases?per_page=100"
	githubFeaturesJSONURL     = "https://raw.githubusercontent.com/datazip-inc/olake-docs/master/features.json"
)

var httpClient = &http.Client{
	Timeout: githubAPITimeout,
}

// GitHubRelease represents a release from GitHub API response.
type GitHubRelease struct {
	TagName     string    `json:"tag_name"`
	Name        string    `json:"name"`
	Body        string    `json:"body"`
	PublishedAt time.Time `json:"published_at"`
	HTMLURL     string    `json:"html_url"`
	Prerelease  bool      `json:"prerelease"`
	Draft       bool      `json:"draft"`
}

// FetchAndBuildReleaseMetadata fetches GitHub releases and converts them into
// normalized, sorted ReleaseMetadataResponse objects.
func FetchAndBuildReleaseMetadata(ctx context.Context, repo, releaseType string, limit int) ([]*dto.ReleaseMetadataResponse, error) {
	url := fmt.Sprintf(githubReleasesURLTemplate, repo)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
	if err != nil {
		return nil, err
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github api returned status %d", resp.StatusCode)
	}

	var rawReleases []GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&rawReleases); err != nil {
		return nil, err
	}

	releases := make([]*dto.ReleaseMetadataResponse, 0)

	for _, release := range rawReleases {
		if release.Draft || release.Prerelease {
			continue
		}

		version := normalizeVersion(releaseType, release.TagName)
		if version == "" {
			continue
		}

		releases = append(releases, &dto.ReleaseMetadataResponse{
			Version:     version,
			Description: release.Body,
			Date:        release.PublishedAt.Format(time.RFC3339),
			Link:        release.HTMLURL,
		})
	}

	// Sort by version (descending)
	sort.Slice(releases, func(i, j int) bool {
		return semver.Compare(releases[i].Version, releases[j].Version) > 0
	})

	if limit > 0 && len(releases) > limit {
		releases = releases[:limit]
	}

	return releases, nil
}

// normalizeVersion converts GitHub tag names into semver-compatible versions.
func normalizeVersion(releaseType, tag string) string {
	switch releaseType {
	// olake-helm has tag format like olake-X.X.X
	case "olake_helm":
		if !strings.HasPrefix(tag, "olake-") {
			return ""
		}
		return "v" + strings.TrimPrefix(tag, "olake-")

	default:
		if !strings.HasPrefix(tag, "v") {
			return ""
		}
		return tag
	}
}

// FetchFeaturesJSON fetches the features.json file from GitHub.
func FetchFeaturesJSON(ctx context.Context) ([]*dto.ReleaseMetadataResponse, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		githubFeaturesJSONURL,
		http.NoBody,
	)
	if err != nil {
		logger.Warnf("failed to create request for features.json: %s", err)
		return []*dto.ReleaseMetadataResponse{}, nil
	}

	// GitHub API requires User-Agent header
	req.Header.Set("User-Agent", "olake-ui-server")

	resp, err := httpClient.Do(req)
	if err != nil {
		logger.Warnf("failed to fetch features.json: %s", err)
		return []*dto.ReleaseMetadataResponse{}, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.Warnf("features.json returned status %d", resp.StatusCode)
		return []*dto.ReleaseMetadataResponse{}, nil
	}

	var releasesArray []*dto.ReleaseMetadataResponse
	if err := json.NewDecoder(resp.Body).Decode(&releasesArray); err != nil {
		// don't fail the request if the features.json is not found
		logger.Warnf("failed to decode features.json: %s", err)
		return []*dto.ReleaseMetadataResponse{}, nil
	}

	if releasesArray == nil {
		releasesArray = []*dto.ReleaseMetadataResponse{}
	}

	return releasesArray, nil
}
