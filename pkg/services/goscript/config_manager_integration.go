package goscript

import (
	"encoding/json"
	"fmt"

	"github.com/gravestench/dark-magic/pkg/services/configManager"
)

type Config struct {
	InitScriptPath string
}

var _ configManager.HasConfiguration = &Service{}

func (s *Service) ConfigFileName() string {
	return "goscript_environment.json"
}

func (s *Service) DefaultConfigData() (data []byte) {
	defaults := &Config{
		InitScriptPath: "init.go",
	}

	data, _ = json.MarshalIndent(defaults, "", "\t")

	return data
}

func (s *Service) IngestConfig(handle *configManager.ConfigHandle) error {
	data, err := handle.Data()
	if err != nil {
		return fmt.Errorf("getting data from config handle: %v", err)
	}

	var tmp Config

	if err = json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	s.Config = &tmp

	return nil
}
