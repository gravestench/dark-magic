package tweens

import (
	"encoding/json"
	"fmt"

	"github.com/gravestench/dark-magic/pkg/services/configManager"
)

type Config struct {
	TickRate int
}

var _ configManager.HasConfiguration = &Service{}

func (s *Service) ConfigFileName() string {
	return "tweens.json"
}

func (s *Service) DefaultConfigData() (data []byte) {
	data, _ = json.MarshalIndent(&Config{
		TickRate: 24,
	}, "", "\t")

	return data
}

func (s *Service) IngestConfig(handle *configManager.ConfigHandle) error {
	data, err := handle.Data()
	if err != nil {
		return fmt.Errorf("getting data from config handle: %v", err)
	}

	if err = json.Unmarshal(data, &s.Config); err != nil {
		return err
	}

	if s.Config.TickRate <= 0 {
		s.Config.TickRate = 24
	}

	return nil
}
