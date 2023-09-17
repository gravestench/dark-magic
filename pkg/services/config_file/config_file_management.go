package config_file

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// GetPath returns the absolute path for a given incoming file
func (s *Service) GetPath(name string) string {
	p, _ := filepath.Abs(expandHomeDirectory(filepath.Join(s.ConfigDirectory(), name)))
	return p
}

// ConfigDirectory returns the directory path where the service's configurations are stored.
// If the directory is not set, it returns a default.
func (s *Service) ConfigDirectory() string {
	if s.RootDirectory == "" {
		s.RootDirectory = defaultConfigDir
	}

	return s.RootDirectory
}

// GetConfig retrieves a configuration by its path from the service's internal map.
// It locks the service's mutex to ensure safe concurrent access.
func (s *Service) GetConfig(path string) (*Config, error) {
	s.mux.Lock()
	defer s.mux.Unlock()

	return s.getConfigUnsafe(path)
}

// getConfigUnsafe retrieves a configuration by its path from the service's internal map.
// If the configuration does not exist, it attempts to load it or create a new one.
// It returns the configuration and any error that occurred.
func (s *Service) getConfigUnsafe(path string) (cfg *Config, err error) {
	if existing, exists := s.configs[path]; exists {
		return existing, nil
	}

	errs := make([]string, 0)

	if cfg, err = s.loadConfigUnsafe(path); err == nil && cfg != nil {
		return cfg, nil
	} else if err != nil {
		errs = append(errs, err.Error())
	}

	if cfg, err = s.createConfigUnsafe(path); err == nil {
		return cfg, nil
	} else {
		errs = append(errs, err.Error())
	}

	strErrors := fmt.Sprintf("[%s]", strings.Join(errs, "]["))
	return nil, fmt.Errorf("getting config: %s", strErrors)
}

// SetConfigDirectory sets the directory path where the service's configurations should be stored.
// It locks the service's mutex to ensure safe concurrent access.
func (s *Service) SetConfigDirectory(dir string) error {
	s.mux.Lock()
	defer s.mux.Unlock()

	if exists, err := directoryExists(dir); !exists || err != nil {
		return fmt.Errorf("bad config RootDirectory: %v", dir)
	}

	s.RootDirectory = dir

	return nil
}

// Configs returns a map of all configurations stored in the service.
func (s *Service) Configs() map[string]*Config {
	return s.configs
}

// CreateConfig creates a new configuration file at the specified path.
// It locks the service's mutex to ensure safe concurrent access.
func (s *Service) CreateConfig(path string) (*Config, error) {
	s.mux.Lock()
	defer s.mux.Unlock()

	return s.createConfigUnsafe(path)
}

// createConfigUnsafe creates a new configuration file at the specified path.
// If the directory doesn't exist or the configuration already exists, it returns an error.
// It returns the newly created configuration and any error that occurred.
func (s *Service) createConfigUnsafe(path string) (*Config, error) {
	if s.configs == nil {
		s.configs = make(map[string]*Config)
	}

	path = expandHomeDirectory(prefixIfPathRelative(s.ConfigDirectory(), path))

	// check if RootDirectory exists
	if exists, err := directoryExistsForFile(path); !exists || err != nil {
		base := filepath.Dir(path)
		if err = os.MkdirAll(base, 0755); err != nil {
			return nil, fmt.Errorf("bad config file path (directory doesnt exist, couldnt create): %v", path)
		}
	}

	// check if we already have an existing config
	if _, exists := s.configs[path]; exists {
		return nil, fmt.Errorf("config already exists")
	}

	// try to create a new file
	file, err := os.Create(path)
	if err != nil {
		return nil, fmt.Errorf("creating file %q: %v", path, err)
	}
	defer file.Close()

	// write a new config to the file
	config := newConfig()
	data, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("unmarshalling new config: %v", err)
	}
	if _, err = file.Write(data); err != nil {
		return nil, fmt.Errorf("writing to new config file: %v", err)
	}

	// keep track of the config
	s.configs[path] = config

	return s.configs[path], nil
}

// LoadConfig loads a configuration from the specified path.
// It locks the service's mutex to ensure safe concurrent access.
func (s *Service) LoadConfig(path string) (*Config, error) {
	s.mux.Lock()
	defer s.mux.Unlock()

	return s.loadConfigUnsafe(path)
}

// loadConfigUnsafe loads a configuration from the specified path.
// If the configuration is not found, it returns an error.
// It returns the loaded configuration and any error that occurred.
func (s *Service) loadConfigUnsafe(path string) (*Config, error) {
	if s.configs == nil {
		s.configs = make(map[string]*Config)
	}

	path = prefixIfPathRelative(s.ConfigDirectory(), path)

	if exists, err := pathExists(path); !exists || err != nil {
		return nil, err
	}

	fileData, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	cfg := newConfig()
	if err = cfg.Unmarshal(fileData); err != nil {
		return nil, err
	}

	s.configs[path] = cfg

	return s.configs[path], nil
}

// SaveConfig saves a configuration to the specified path.
// It locks the service's mutex to ensure safe concurrent access.
func (s *Service) SaveConfig(path string) error {
	s.mux.Lock()
	defer s.mux.Unlock()

	return s.saveConfigUnsafe(path)
}

// saveConfigUnsafe saves a configuration to the specified path.
// If the configuration is not found, it returns an error.
// It returns any error that occurred during the saving process.
func (s *Service) saveConfigUnsafe(path string) error {
	path = prefixIfPathRelative(s.ConfigDirectory(), path)

	cfg, err := s.getConfigUnsafe(path)
	if err != nil {
		return fmt.Errorf("getting config %q: %v", path, err)
	}

	fileData, err := cfg.Marshal()
	if err != nil {
		return fmt.Errorf("marshalling config data: %q", cfg)
	}

	err = os.WriteFile(path, fileData, 0644)
	if err != nil {
		return fmt.Errorf("writing config data to file: %q", cfg)
	}

	return nil
}
