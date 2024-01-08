package webRouter

import (
	"encoding/json"
	"fmt"

	"github.com/gravestench/dark-magic/pkg/services/configManager"
)

var _ configManager.HasConfiguration = &Service{}

type Config struct {
	Gin struct {
		Debug bool
	}
}

func (s *Service) ConfigFileName() string {
	return "web_router.json"
}

func (s *Service) DefaultConfigData() []byte {
	var cfg Config

	cfg.Gin.Debug = true

	data, _ := json.MarshalIndent(&cfg, "", "\t")

	return data
}

func (s *Service) IngestConfig(config *configManager.ConfigHandle) error {
	data, err := config.Data()
	if err != nil {
		return fmt.Errorf("getting config data: %v", err)
	}

	if err = json.Unmarshal(data, &s.config); err != nil {
		return fmt.Errorf("unmarshalling config: %v", err)
	}

	return nil
}
