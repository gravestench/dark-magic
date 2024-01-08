package hero

import (
	"encoding/json"
	"fmt"

	"github.com/gravestench/dark-magic/pkg/services/configManager"
)

type Config map[string]*State

var _ configManager.HasConfiguration = &Service{}

func (s *Service) ConfigFileName() string {
	return "heroes.json"
}

func (s *Service) DefaultConfigData() (data []byte) {
	data, _ = json.MarshalIndent(Config{}, "", "\t")

	return data
}

func (s *Service) IngestConfig(handle *configManager.ConfigHandle) error {
	s.cfgHandle = handle // used for saving/loading elsewhere

	data, err := handle.Data()
	if err != nil {
		return fmt.Errorf("getting data from config handle: %v", err)
	}

	tmp := make(Config)

	if err = json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	s.Config = tmp

	return nil
}
