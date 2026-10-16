package storagemode

import (
	"github.com/datazip-inc/olake-ui/server/internal/appconfig"
	"github.com/datazip-inc/olake-ui/server/internal/constants"
)

// Get returns OLAKE_STORAGE_MODE from config, defaulting to nfs.
func Get() string {
	mode := appconfig.Load().OlakeStorageMode
	if mode == "" {
		return constants.StorageModeNFS
	}
	return mode
}
