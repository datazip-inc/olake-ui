package utils

import (
	"strings"

	"github.com/datazip-inc/olake-ui/server/internal/constants"
	"golang.org/x/mod/semver"
)

// SupportsSchemaSplit reports whether the source connector version accepts the
// --schema catalog split flag.
func SupportsSchemaSplit(version string) bool {
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
	return semver.Compare(version, constants.MinSchemaSplitVersion) >= 0
}

// UseSchemaSplit reports whether schema_config should be mounted and --schema passed.
func UseSchemaSplit(schemaConfig, version string) bool {
	return strings.TrimSpace(schemaConfig) != "" && SupportsSchemaSplit(version)
}
