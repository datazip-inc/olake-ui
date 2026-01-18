package utils

import (
	"fmt"
	"strings"

	"github.com/datazip-inc/olake-ui/server/internal/models/dto"
	"golang.org/x/mod/semver"
)

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
