package services

import (
	"context"
	"fmt"

	"golang.org/x/mod/semver"

	"github.com/datazip-inc/olake-ui/server/internal/constants"
	"github.com/datazip-inc/olake-ui/server/internal/models/dto"
	"github.com/datazip-inc/olake-ui/server/utils"
)

type ReleaseType string

const (
	ReleaseOlakeUI   ReleaseType = "olake_ui"
	ReleaseWorker    ReleaseType = "worker"
	ReleaseOlakeHelm ReleaseType = "olake_helm"
	ReleaseOlake     ReleaseType = "olake"
	ReleaseFeatures  ReleaseType = "features"
)

type GithubReleaseSource struct {
	Type      ReleaseType
	Repo      string
	OnlyNewer bool
}

var releaseSources = []GithubReleaseSource{
	{Type: ReleaseOlakeUI, Repo: "olake-ui", OnlyNewer: true},
	{Type: ReleaseWorker, Repo: "olake-helm", OnlyNewer: true},
	{Type: ReleaseOlakeHelm, Repo: "olake-helm"},
	{Type: ReleaseOlake, Repo: "olake"},
}

func (s *ETLService) GetAllReleasesResponse(
	ctx context.Context,
	limit int,
) (*dto.ReleasesResponse, error) {
	currentVersion := constants.AppVersion
	fetchedReleases := make(map[ReleaseType][]*dto.ReleaseMetadataResponse)

	// fetch features
	features, err := utils.FetchFeaturesJSON(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch features.json: %s", err)
	}
	fetchedReleases[ReleaseFeatures] = features

	// fetch releases
	for _, src := range releaseSources {
		data, err := utils.FetchAndBuildReleaseMetadata(ctx, src.Repo, string(src.Type), limit)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch github releases for %s: %s", src.Type, err)
		}
		fetchedReleases[src.Type] = data
	}

	return buildReleasesResponse(currentVersion, fetchedReleases)
}

// build releases response from fetched data
func buildReleasesResponse(currentVersion string, fetched map[ReleaseType][]*dto.ReleaseMetadataResponse) (*dto.ReleasesResponse, error) {
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

	for _, src := range releaseSources {
		raw, ok := fetched[src.Type]
		if !ok || raw == nil {
			return nil, fmt.Errorf("release data missing for %s", src.Type)
		}

		releases := make([]*dto.ReleaseMetadataResponse, 0)

		for _, r := range raw {
			cmp := semver.Compare(r.Version, currentVersion)

			// Skip older releases if required
			if src.OnlyNewer && cmp <= 0 {
				continue
			}

			tags := []string{}
			if src.OnlyNewer && cmp > 0 {
				tags = append(tags, "new-release")
			}

			releases = append(releases, &dto.ReleaseMetadataResponse{
				Version:     r.Version,
				Description: r.Description,
				Tags:        tags,
				Date:        r.Date,
				Link:        r.Link,
			})
		}

		data := &dto.ReleaseTypeData{
			Releases: releases,
		}

		if src.OnlyNewer {
			data.CurrentVersion = currentVersion
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

	resp.OlakeUIWorker = utils.MergeReleaseDescriptions(uiData, "OLake UI", workerData, "OLake Worker")

	return resp, nil
}
