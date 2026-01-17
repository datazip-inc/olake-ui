package services

import (
	"context"
	"fmt"

	"github.com/datazip-inc/olake-ui/server/internal/constants"
	"github.com/datazip-inc/olake-ui/server/internal/models/dto"
	"github.com/datazip-inc/olake-ui/server/utils"
)

func (s *ETLService) GetAllReleasesResponse(
	ctx context.Context,
	limit int,
) (*dto.ReleasesResponse, error) {
	currentVersion := constants.AppVersion
	fetchedReleases := make(map[utils.ReleaseType][]*dto.ReleaseMetadataResponse)

	// fetch features
	features, err := utils.FetchFeaturesJSON(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch features.json: %s", err)
	}
	fetchedReleases[utils.ReleaseFeatures] = features

	// fetch releases
	for _, src := range utils.ReleaseSources {
		data, err := utils.FetchAndBuildReleaseMetadata(ctx, src.Repo, string(src.Type), limit)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch github releases for %s: %s", src.Type, err)
		}
		fetchedReleases[src.Type] = data
	}

	return utils.BuildReleasesResponse(currentVersion, fetchedReleases)
}
