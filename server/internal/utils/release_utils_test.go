package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/datazip-inc/olake-ui/server/internal/models/dto"
)

// OLake Fusion releases are listed without newer-release filtering: the deployed
// Fusion version is unknown to the releases handler, so every fetched release is
// returned and no "new-release" tag is attached.
func TestBuildReleasesResponseIncludesFusionReleases(t *testing.T) {
	release := func(version string) *dto.ReleaseMetadataResponse {
		return &dto.ReleaseMetadataResponse{
			Version:     version,
			Description: "notes for " + version,
			Date:        "2026-01-17T10:00:00Z",
			Link:        "https://github.com/datazip-inc/olake-fusion/releases/tag/" + version,
		}
	}

	fetched := map[ReleaseType][]*dto.ReleaseMetadataResponse{
		ReleaseFeatures: {
			{Description: "feature", Date: "2026-01-17T10:00:00Z"},
		},
		ReleaseOlakeUI:     {release("v9.9.9")},
		ReleaseWorker:      {release("v9.9.9")},
		ReleaseOlakeHelm:   {release("v0.0.1")},
		ReleaseOlake:       {release("v9.9.9")},
		ReleaseOlakeFusion: {release("v0.2.2"), release("v0.2.1"), release("v0.2.0")},
	}

	resp, err := BuildReleasesResponse("v1.0.0", "v0.0.1", fetched)
	require.NoError(t, err)
	require.NotNil(t, resp.OlakeFusion, "fusion releases must be part of the response")

	assert.Len(t, resp.OlakeFusion.Releases, 3, "fusion releases are not filtered against the UI version")
	for _, r := range resp.OlakeFusion.Releases {
		assert.NotContains(t, r.Tags, "new-release", "fusion releases carry no new-release tag without a known deployed version")
	}
	assert.Empty(t, resp.OlakeFusion.CurrentVersion)
}

// The fusion source must resolve to the olake-fusion repository with no
// newer-release filtering.
func TestReleaseSourcesIncludesFusion(t *testing.T) {
	var fusion *GithubReleaseSource
	for i, src := range ReleaseSources {
		if src.Type == ReleaseOlakeFusion {
			fusion = &ReleaseSources[i]
		}
	}
	require.NotNil(t, fusion, "olake_fusion must be one of the release sources")
	assert.Equal(t, "olake-fusion", fusion.Repo)
	assert.False(t, fusion.OnlyNewerReleases)
}
