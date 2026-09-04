package utils

import (
	"fmt"
	"strings"

	"github.com/datazip-inc/olake-ui/server/internal/constants"
	"golang.org/x/mod/semver"
)

func CompareAtLeast(version, minVersion string) bool {
	version = strings.TrimSpace(version)
	if version == "" {
		return false
	}
	if !strings.HasPrefix(version, "v") {
		version = "v" + version
	}
	if !semver.IsValid(version) {
		return false
	}

	minVersion = strings.TrimSpace(minVersion)
	if !strings.HasPrefix(minVersion, "v") {
		minVersion = "v" + minVersion
	}
	if !semver.IsValid(minVersion) {
		return false
	}
	return semver.Compare(version, minVersion) >= 0
}

// SupportsDestinationDatabasePrefix reports whether the driver accepts --destination-database-prefix.
func SupportsDestinationDatabasePrefix(version string) bool {
	return GetCustomDriverVersion() != "" || CompareAtLeast(version, constants.DefaultSpecVersion)
}

// SupportsMaxDiscoverThreads reports whether the driver accepts --max-discover-threads.
func SupportsMaxDiscoverThreads(version string) bool {
	return CompareAtLeast(version, constants.DefaultMaxDiscoverThreadsVersion)
}

// SupportsSelectedStreamsFlag reports whether the driver accepts --selected_streams on discover/sync/clear.
func SupportsSelectedStreamsFlag(version string) bool {
	return GetCustomDriverVersion() != "" || CompareAtLeast(version, constants.MinSelectedStreamsSplitVersion)
}

// ResolveSpecVersion bumps the version to DefaultSpecVersion when below the minimum
// required for `spec`, unless a custom driver version override is set.
func ResolveSpecVersion(version string) string {
	if GetCustomDriverVersion() == "" && !CompareAtLeast(version, constants.DefaultSpecVersion) {
		return constants.DefaultSpecVersion
	}
	return version
}

func CheckClearDestinationCompatibility(sourceVersion string) error {
	if GetCustomDriverVersion() != "" || CompareAtLeast(sourceVersion, constants.DefaultClearDestinationVersion) {
		return nil
	}
	return fmt.Errorf(
		"source version %s is not supported for clear destination. please update the source version to %s or higher",
		sourceVersion,
		constants.DefaultClearDestinationVersion,
	)
}
