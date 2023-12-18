package configFile

import (
	"github.com/gravestench/servicemesh"
)

// Ensure that Service implements the required interfaces.
var (
	_ servicemesh.Service   = &Service{}
	_ servicemesh.HasLogger = &Service{}
	_ Manager               = &Service{}
)

// The following interfaces are to be used much like the service interfaces
// found inside of servicemesh/pkg. These can be used by other services to
// declare and resolve their dependencies to the service defined in this
// package.

// HasConfig represents a something with a configuration file path and retrieval methods.
type HasConfig interface {
	ConfigFileName() string // ConfigFilePath returns the path to the configuration file.
}

// HasDefaultConfig represents something with a default configuration.
type HasDefaultConfig interface {
	HasConfig
	DefaultConfig() Config // DefaultConfig returns the default configuration.
}

type Dependency = Manager

// Manager represents something that manages configurations.
type Manager interface {
	GetFilePath(path string) string
	ConfigDirectory() string                               // ConfigDirectory returns the directory path where configurations are stored.
	SetConfigDirectory(string) error                       // SetConfigDirectory sets the directory path for configurations.
	Configs() map[string]*Config                           // Configs returns all configurations stored in the service.
	GetConfigByFileName(name string) (*Config, error)      // GetConfig retrieves a configuration by its path.
	CreateConfigWithFileName(name string) (*Config, error) // CreateConfig creates a new configuration file at the specified path.
	LoadConfigWithFileName(name string) (*Config, error)   // LoadConfig loads a configuration from the specified path.
	SaveConfigWithFileName(name string) error              // SaveConfig saves a configuration to the specified path.
}
