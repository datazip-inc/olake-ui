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

	resp := &dto.ReleasesResponse{}
	currentVersion := constants.AppVersion

	var (
		uiData     *dto.ReleaseTypeData
		workerData *dto.ReleaseTypeData
	)

	for _, src := range releaseSources {

		rawReleases, err := utils.FetchGitHubReleases(ctx, src.Repo)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch github releases for %s: %s", src.Type, err)
		}

		metadata := utils.BuildReleaseMetadata(rawReleases, string(src.Type), limit)

		releases := make([]*dto.ReleaseMetadataResponse, 0)

		for _, m := range metadata {
			cmp := semver.Compare(m.Version, currentVersion)

			// Skip older releases
			if src.OnlyNewer && cmp <= 0 {
				continue
			}

			var tags []string
			if src.OnlyNewer && cmp > 0 {
				tags = append(tags, "new-release")
			}

			releases = append(releases, &dto.ReleaseMetadataResponse{
				Version:     m.Version,
				Description: m.Description,
				Tags:        tags,
				Date:        m.Date,
				Link:        m.Link,
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

	resp.OlakeUIWorker = utils.MergeReleaseDescriptions(uiData, "OLake UI", workerData, "Worker")

	return resp, nil
}
