package docker

import (
	"github.com/datazip-inc/olake-ui/server/internal/database"
	"github.com/docker/docker/client"
)

// Constants
const (
	// File and directory permissions
	DefaultDirPermissions  = 0755
	DefaultFilePermissions = 0644

	// Directory paths
	DefaultConfigDir  = "/tmp/olake-config"
	ContainerMountDir = "/mnt/config"

	// Docker related
	// dockerImagePrefix = "olakego/source"

	// Environment variables
	envPersistentDir = "PERSISTENT_DIR"
)

// Command represents a Docker command type
type Command string

const (
	Discover Command = "discover"
	Spec     Command = "spec"
	Check    Command = "check"
	Sync     Command = "sync"
)

// FileConfig represents a configuration file to be written
type FileConfig struct {
	Name string
	Data string
}

// Runner is responsible for executing Docker commands
type Runner struct {
	WorkingDir   string
	anonymousID  string
	dockerClient *client.Client
	db           *database.Database
}
