package hero

import (
	"encoding/json"
	"fmt"

	"github.com/gravestench/dark-magic/pkg/models"
	"github.com/gravestench/dark-magic/pkg/services/configFile"
)

func (s *Service) ConfigFileName() string {
	return "heroes.json"
}

func (s *Service) LoadHeroes() error {
	cfg, err := s.config.LoadConfigWithFileName(s.ConfigFileName())
	if err != nil {
		return fmt.Errorf("loading config", "error", err)
	}

	if cfg == nil {
		cfg, err = s.config.CreateConfigWithFileName(s.ConfigFileName())
		if err != nil {
			return fmt.Errorf("creating config", "error", err)
		}
	}

	for _, group := range cfg.GroupKeys() {
		state := s.loadHeroStateFromConfigGroup(cfg.Group(group))
		existing := s.GetHeroByName(state.Name)

		if existing == nil {
			s.heroStates = append(s.heroStates, state)
			continue
		}

		existing = &state
	}

	return nil
}

func (s *Service) loadHeroStateFromConfigGroup(cfg configFile.Object) State {
	state := State{
		Name:       cfg.GetString("Name"),
		Experience: cfg.GetInt("Experience"),
		Class:      models.HeroFromString(cfg.GetString("Class")),
		Level:      cfg.GetInt("Level"),
		Skills:     cfg.GetStrings("Skills"),
	}

	state.Attributes = make(map[string]int)
	attr := cfg.Get("Attributes")
	data, _ := json.Marshal(attr)
	_ = json.Unmarshal(data, &state.Attributes)

	return state
}

func (s *Service) SaveHeroes() error {
	cfg, err := s.config.LoadConfigWithFileName(s.ConfigFileName())
	if err != nil {
		return fmt.Errorf("loading config", "error", err)
	}

	for _, heroState := range s.heroStates {
		g := cfg.Group(heroState.Name)

		g.Set("Name", heroState.Name)
		g.Set("Experience", heroState.Experience)
		g.Set("Class", heroState.Class.String())
		g.Set("Level", heroState.Level)
		g.Set("Skills", heroState.Skills)
		g.Set("Attributes", heroState.Attributes)
	}

	s.config.SaveConfigWithFileName(s.ConfigFileName())

	return nil
}
