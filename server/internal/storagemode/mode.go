package storagemode

import (
	"strings"

	"github.com/datazip-inc/olake-ui/server/internal/constants"
	"github.com/spf13/viper"
)

// Get returns OLAKE_STORAGE_MODE from the environment, defaulting to nfs.
func Get() string {
	mode := strings.ToLower(strings.TrimSpace(viper.GetString(constants.EnvStorageMode)))
	if mode == "" {
		return constants.StorageModeNFS
	}
	return mode
}
