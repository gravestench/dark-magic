package configManager

import (
	"fmt"
	"log"
	"path/filepath"
)

type Dependency = interface {
	ConfigDirectory() (string, error)
}

// HasConfiguration represents something that is managed by our
// config file manager
type HasConfiguration interface {
	ConfigFileName() string
	DefaultConfigData() []byte
	IngestConfig(handle *ConfigHandle) error
}

// InitConfiguration will take a HasConfiguration instance and init the config
// file with default data or load the existing data and pass it back to the
// HasConfiguration to do whatever it needs to do.
func (s *Service) InitConfiguration(c HasConfiguration) error {
	// always make sure our root directory is good
	if err := s.ensureExistingRootConfigDirectory(); err != nil {
		return fmt.Errorf("ensuring existing root config directory: %v", err)
	}

	// create a new config handle for the file within our config dir
	configPath := filepath.Join(s.RootDirectory, c.ConfigFileName())
	handle, err := s.newConfigHandle(configPath)
	if err != nil {
		return fmt.Errorf("getting config handle: %v", err)
	}

	// if there is no existing data in the file, use the defaults
	if existingData, errLoad := handle.Data(); errLoad == nil && len(existingData) == 0 {
		log.Printf("setting default values for config file %q", configPath)

		if errSetDefault := handle.SetData(c.DefaultConfigData()); errSetDefault != nil {
			return fmt.Errorf("applying default configuration data: %v", err)
		}
	}

	s.Logger().Info("ingesting config file", "path", configPath)

	// pass the config handle to the configurable so it can do whatever it needs to do
	if errIngest := c.IngestConfig(handle); errIngest != nil {
		return fmt.Errorf("ingesting config data for %q: %v", c.ConfigFileName(), errIngest)
	}

	return nil
}
