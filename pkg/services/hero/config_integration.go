package hero

import (
	"encoding/json"

	"github.com/gravestench/dark-magic/pkg/models"
	"github.com/gravestench/dark-magic/pkg/services/configFile"
)

var _ configFile.HasConfig = &Service{}

func (s *Service) ConfigFileName() string {
	return "heroes.json"
}

func (s *Service) LoadConfig(config *configFile.Config) {
	s.config = config
}

func (s *Service) LoadHeroes() error {
	for _, group := range s.config.GroupKeys() {
		state := s.loadHeroStateFromConfigGroup(s.config.Group(group))
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
	for _, heroState := range s.heroStates {
		g := s.config.Group(heroState.Name)

		g.Set("Name", heroState.Name)
		g.Set("Experience", heroState.Experience)
		g.Set("Class", heroState.Class.String())
		g.Set("Level", heroState.Level)
		g.Set("Skills", heroState.Skills)
		g.Set("Attributes", heroState.Attributes)
	}

	return s.config.Save()
}
