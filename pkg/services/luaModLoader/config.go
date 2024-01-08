package luaModLoader

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gravestench/dark-magic/pkg/services/configManager"
)

const (
	defaultModDirectory = "dark-magic-mods"
)

type Config struct {
	ModDirectory string
	EnabledMods  map[string]bool
}

var _ configManager.HasConfiguration = &Service{}

func (s *Service) ConfigFileName() string {
	return "modloader.json"
}

func (s *Service) DefaultConfigData() (data []byte) {
	usrConfigDir, err := os.UserConfigDir()
	if err != nil {
		return nil
	}

	defaults := &Config{
		ModDirectory: filepath.Join(usrConfigDir, defaultModDirectory),
		EnabledMods: map[string]bool{
			"Dark Magic Terminal (0.0.1)": true,
		},
	}

	data, _ = json.MarshalIndent(defaults, "", "\t")

	return
}

func (s *Service) IngestConfig(handle *configManager.ConfigHandle) error {
	data, err := handle.Data()
	if err != nil {
		return fmt.Errorf("getting data from config handle: %v", err)
	}

	if err = json.Unmarshal(data, &s.Config); err != nil {
		return err
	}

	s.Config.ModDirectory, err = expandHomeDirectoryPath(s.Config.ModDirectory)
	if err != nil {
		return fmt.Errorf("expanding path: %v", err)
	}

	return nil
}

func expandHomeDirectoryPath(path string) (string, error) {
	if !strings.HasPrefix(path, "~") {
		return path, nil
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	path = strings.Replace(path, "~", homeDir, 1)

	return path, nil
}
