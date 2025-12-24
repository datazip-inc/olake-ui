package services

import (
	"context"

	"github.com/spf13/viper"

	"github.com/datazip-inc/olake-ui/server/internal/models/dto"
	"github.com/datazip-inc/olake-ui/server/utils"
)

var dockerRepos = map[string]string{
	"olake_ui_worker": "olakego/ui", // UI + worker
	"olake_helm":      "olakego/olake-helm",
	"olake":           "olakego/olake",
	"worker":          "olakego/worker",
}

func (s *ETLService) GetAllReleasesResponse(
	ctx context.Context,
	limit int,
) (*dto.ReleasesResponse, error) {
	currentVersion := viper.GetString("APP_VERSION")
	if currentVersion == "" {
		currentVersion = "v0.0.0"
	}

	resp := &dto.ReleasesResponse{}

	for releaseType, repo := range dockerRepos {
		// Skip "worker" as it's not part of the response structure
		if releaseType == "worker" {
			continue
		}

		onlyNewerVersions := releaseType == "olake_ui_worker"

		releaseData, err := utils.GetReleaseDataForType(ctx, repo, releaseType, currentVersion, limit, onlyNewerVersions)
		if err != nil || releaseData == nil {
			continue
		}

		// Map release type to response field
		switch releaseType {
		case "olake_ui_worker":
			resp.OlakeUIWorker = releaseData
		case "olake_helm":
			resp.OlakeHelm = releaseData
		case "olake":
			resp.Olake = releaseData
		}
	}

	return resp, nil
}
