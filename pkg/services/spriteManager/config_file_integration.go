package spriteManager

import (
	"encoding/json"
	"fmt"

	"github.com/gravestench/dark-magic/pkg/services/configManager"
)

type Config struct {
	Cache struct {
		BudgetMB int
	}
}

var _ configManager.HasConfiguration = &Service{}

func (s *Service) ConfigFileName() string {
	return "sprite_manager.json"
}

func (s *Service) DefaultConfigData() []byte {
	var cfg Config

	cfg.Cache.BudgetMB = 500

	data, _ := json.MarshalIndent(&cfg, "", "\t")

	return data
}

func (s *Service) IngestConfig(handle *configManager.ConfigHandle) error {
	data, err := handle.Data()
	if err != nil {
		return fmt.Errorf("getting config data: %v", err)
	}

	if err = json.Unmarshal(data, &s.config); err != nil {
		return fmt.Errorf("unmarshalling config data: %v", err)
	}

	return nil
}
