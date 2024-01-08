package hero

import (
	"encoding/json"
	"fmt"
)

func (s *Service) ReloadHeroes() error {
	if s.cfgHandle == nil { // populated by config manager service
		return fmt.Errorf("no config handle")
	}

	s.Config = make(Config)

	data, err := s.cfgHandle.Data()
	if err != nil {
		return fmt.Errorf("getting data from config handle: %v", err)
	}

	if err = json.Unmarshal(data, &s.Config); err != nil {
		return fmt.Errorf("marshalling data: %v", err)
	}

	return nil
}

func (s *Service) SaveHeroes() error {
	data, err := json.MarshalIndent(s.Config, "", "\t")
	if err != nil {
		return fmt.Errorf("marshaling hero data: %v", err)
	}

	if s.cfgHandle == nil { // populated by config manager service
		return fmt.Errorf("no config handle")
	}

	if err = s.cfgHandle.SetData(data); err != nil {
		return fmt.Errorf("setting data for config: %v", err)
	}

	return nil
}
