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

// ReleaseType represents the type of release source.
type ReleaseType string

const (
	ReleaseOlakeUI   ReleaseType = "olake_ui"
	ReleaseWorker    ReleaseType = "worker"
	ReleaseOlakeHelm ReleaseType = "olake_helm"
	ReleaseOlake     ReleaseType = "olake"
	ReleaseFeatures  ReleaseType = "features"
)

// GithubReleaseSource represents a GitHub repository source for fetching releases.
// It defines the release type, repository name, and whether to filter only newer releases.
type GithubReleaseSource struct {
	Type              ReleaseType
	Repo              string
	OnlyNewerReleases bool
}

// ReleaseSources defines the list of GitHub repositories to fetch releases from.
var ReleaseSources = []GithubReleaseSource{
	{Type: ReleaseOlakeUI, Repo: "olake-ui", OnlyNewerReleases: true},
	{Type: ReleaseWorker, Repo: "olake-helm", OnlyNewerReleases: true},
	{Type: ReleaseOlakeHelm, Repo: "olake-helm"},
	{Type: ReleaseOlake, Repo: "olake", OnlyNewerReleases: true},
}

// BuildReleasesResponse builds the final releases response from fetched data.
// currentVersion is used for olake-ui and worker releases.
// olakeSourceVersion is used for olake source releases (minimum version from database).
func BuildReleasesResponse(currentVersion, olakeSourceVersion string, fetched map[ReleaseType][]*dto.ReleaseMetadataResponse) (*dto.ReleasesResponse, error) {
	resp := &dto.ReleasesResponse{}

	var (
		uiData     *dto.ReleaseTypeData
		workerData *dto.ReleaseTypeData
	)

	features, ok := fetched[ReleaseFeatures]
	if !ok || features == nil {
		return nil, fmt.Errorf("features data is missing")
	}
	resp.Features = &dto.ReleaseTypeData{
		Releases: features,
	}

	for _, src := range ReleaseSources {
		raw, ok := fetched[src.Type]
		if !ok || raw == nil {
			return nil, fmt.Errorf("release data missing for %s", src.Type)
		}

		releases := make([]*dto.ReleaseMetadataResponse, 0)

		// Determine which version to compare against
		compareVersion := Ternary(src.Type == ReleaseOlake, olakeSourceVersion, currentVersion).(string)

		for _, release := range raw {
			comparison := semver.Compare(release.Version, compareVersion)

			// Skip older releases if required
			if src.OnlyNewerReleases && comparison <= 0 {
				continue
			}

			tags := []string{}
			if src.OnlyNewerReleases && comparison > 0 {
				tags = append(tags, "new-release")
			}

			releases = append(releases, &dto.ReleaseMetadataResponse{
				Version:     release.Version,
				Description: release.Description,
				Tags:        tags,
				Date:        release.Date,
				Link:        release.Link,
			})
		}

		data := &dto.ReleaseTypeData{
			Releases: releases,
		}

		if src.OnlyNewerReleases {
			data.CurrentVersion = Ternary(src.Type == ReleaseOlake, olakeSourceVersion, currentVersion).(string)
		}

		switch src.Type {
		case ReleaseOlakeUI:
			uiData = data
		case ReleaseWorker:
			workerData = data
		case ReleaseOlakeHelm:
			resp.OlakeHelm = data
		case ReleaseOlake:
			resp.Olake = data
		}
	}

	resp.OlakeUIWorker = MergeReleaseDescriptions(uiData, "OLake UI", workerData, "OLake Worker")

	return resp, nil
}

// MergeReleaseDescriptions merges secondary release notes into primary by version.
func MergeReleaseDescriptions(primary *dto.ReleaseTypeData, primaryTitle string, secondary *dto.ReleaseTypeData, secondaryTitle string) *dto.ReleaseTypeData {
	if primary == nil || secondary == nil {
		return primary
	}

	secondaryByVersion := make(map[string]*dto.ReleaseMetadataResponse)
	for _, release := range secondary.Releases {
		secondaryByVersion[release.Version] = release
	}

	for _, primaryRelease := range primary.Releases {
		secondaryRelease, ok := secondaryByVersion[primaryRelease.Version]
		if !ok {
			continue
		}

		if strings.TrimSpace(secondaryRelease.Description) == "" {
			continue
		}

		primaryRelease.Description = fmt.Sprintf(
			"## %s\n%s\n\n## %s\n%s",
			primaryTitle,
			strings.TrimSpace(primaryRelease.Description),
			secondaryTitle,
			strings.TrimSpace(secondaryRelease.Description),
		)
	}

	return primary
}

// FetchAndBuildReleaseMetadata fetches GitHub releases and converts them into
// normalized, sorted ReleaseMetadataResponse objects.
func FetchAndBuildReleaseMetadata(ctx context.Context, repo, releaseType string, limit int) ([]*dto.ReleaseMetadataResponse, error) {
	url := fmt.Sprintf(githubReleasesURLTemplate, repo)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
	if err != nil {
		return nil, err
	}

	resp, err := httpClient.Do(req) // #nosec G704 -- URL is built from a compile-time constant (githubReleasesURLTemplate)
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

	resp, err := httpClient.Do(req) // #nosec G704 -- URL is a compile-time constant (githubFeaturesJSONURL)
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
