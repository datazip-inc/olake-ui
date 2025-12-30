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

const githubReleasesURLTemplate = "https://api.github.com/repos/datazip-inc/%s/releases?per_page=100"
const githubFeaturesJSONURL = "https://raw.githubusercontent.com/datazip-inc/olake-docs/master/features.json"

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

	resp, err := http.DefaultClient.Do(req)
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

	out := make([]*dto.ReleaseMetadataResponse, 0)

	for _, r := range rawReleases {
		if r.Draft || r.Prerelease {
			continue
		}

		version := normalizeVersion(releaseType, r.TagName)
		if version == "" {
			continue
		}

		out = append(out, &dto.ReleaseMetadataResponse{
			Version:     version,
			Description: r.Body,
			Date:        r.PublishedAt.Format(time.RFC3339),
			Link:        r.HTMLURL,
		})
	}

	// Sort by version (descending)
	sort.Slice(out, func(i, j int) bool {
		return semver.Compare(out[i].Version, out[j].Version) > 0
	})

	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}

	return out, nil
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

// MergeReleaseDescriptions merges secondary release notes into primary by version.
func MergeReleaseDescriptions(primary *dto.ReleaseTypeData, primaryTitle string, secondary *dto.ReleaseTypeData, secondaryTitle string) *dto.ReleaseTypeData {
	if primary == nil || secondary == nil {
		return primary
	}

	secondaryByVersion := make(map[string]*dto.ReleaseMetadataResponse)
	for _, r := range secondary.Releases {
		secondaryByVersion[r.Version] = r
	}

	for _, p := range primary.Releases {
		s, ok := secondaryByVersion[p.Version]
		if !ok {
			continue
		}

		if strings.TrimSpace(s.Description) == "" {
			continue
		}

		p.Description = fmt.Sprintf(
			"## %s\n%s\n\n## %s\n%s",
			primaryTitle,
			strings.TrimSpace(p.Description),
			secondaryTitle,
			strings.TrimSpace(s.Description),
		)
	}

	return primary
}

// FetchFeaturesJSON fetches the features.json file from GitHub.
func FetchFeaturesJSON(
	ctx context.Context,
) ([]*dto.ReleaseMetadataResponse, error) {
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

	resp, err := http.DefaultClient.Do(req)
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
