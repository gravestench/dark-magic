package hero

import (
	"fmt"

	"github.com/gravestench/runtime"
	"github.com/rs/zerolog"

	"github.com/gravestench/dark-magic/pkg/models"
	"github.com/gravestench/dark-magic/pkg/services/configFile"
	"github.com/gravestench/dark-magic/pkg/services/recordManager"
)

type Service struct {
	logger *zerolog.Logger

	config  configFile.Dependency
	records recordManager.Dependency

	expBreakpoints map[models.Hero][]experienceBreakpoint
	heroStates     map[string]State
}

func (s *Service) Init(rt runtime.Runtime) {
	s.expBreakpoints = make(map[models.Hero][]experienceBreakpoint)

	s.loadExperienceBreakpoints()
}

func (s *Service) loadExperienceBreakpoints() {
	const divisor = 1024

	for idx, record := range s.records.ExperienceBreakpoints() {
		if idx == 0 {
			// the records file has the max level as the first record...
			continue
		}

		s.expBreakpoints[models.HeroBarbarian] = append(s.expBreakpoints[models.HeroBarbarian], experienceBreakpoint{
			Experience: record.Barbarian,
			Ratio:      float32(record.ExpRatio) / divisor,
		})

		s.expBreakpoints[models.HeroNecromancer] = append(s.expBreakpoints[models.HeroNecromancer], experienceBreakpoint{
			Experience: record.Necromancer,
			Ratio:      float32(record.ExpRatio) / divisor,
		})

		s.expBreakpoints[models.HeroPaladin] = append(s.expBreakpoints[models.HeroPaladin], experienceBreakpoint{
			Experience: record.Paladin,
			Ratio:      float32(record.ExpRatio) / divisor,
		})

		s.expBreakpoints[models.HeroAssassin] = append(s.expBreakpoints[models.HeroAssassin], experienceBreakpoint{
			Experience: record.Assassin,
			Ratio:      float32(record.ExpRatio) / divisor,
		})

		s.expBreakpoints[models.HeroSorceress] = append(s.expBreakpoints[models.HeroSorceress], experienceBreakpoint{
			Experience: record.Sorceress,
			Ratio:      float32(record.ExpRatio) / divisor,
		})

		s.expBreakpoints[models.HeroAmazon] = append(s.expBreakpoints[models.HeroAmazon], experienceBreakpoint{
			Experience: record.Amazon,
			Ratio:      float32(record.ExpRatio) / divisor,
		})

		s.expBreakpoints[models.HeroDruid] = append(s.expBreakpoints[models.HeroDruid], experienceBreakpoint{
			Experience: record.Druid,
			Ratio:      float32(record.ExpRatio) / divisor,
		})

	}
}

func (s *Service) Name() string {
	return "Character Generator"
}

// the following methods are boilerplate, but they are used
// by the runtime to enforce a standard logging format.

func (s *Service) BindLogger(logger *zerolog.Logger) {
	s.logger = logger
}

func (s *Service) Logger() *zerolog.Logger {
	return s.logger
}

func (s *Service) LoadHeros() error {
	cfg, err := s.config.LoadConfigWithFileName(s.ConfigFileName())
	if err != nil {
		return fmt.Errorf("loading config: %v", err)
	}

	for _, group := range cfg.GroupKeys() {
		s.heroStates[group] = s.loadHeroStateFromConfigGroup(cfg.Group(group))
	}

	return nil
}

func (s *Service) loadHeroStateFromConfigGroup(cfg configFile.Object) State {
	return State{
		Name:       cfg.GetString("name"),
		experience: cfg.GetInt("Experience"),
		Type:       models.Hero(cfg.GetInt("class")),
	}
}

func (s *Service) SaveHeros() error {
	//TODO implement me
	panic("implement me")
}

func (s *Service) CreateHero(name string, hero models.Hero) State {
	state := State{
		Name:                  name,
		experience:            0,
		Type:                  hero,
		experienceProgression: s.expBreakpoints[hero],
	}

	return state
}

func (s *Service) GetHeros() map[models.Hero]map[string]*State {
	//TODO implement me
	panic("implement me")
}

func (s *Service) GetHero(t models.Hero, name string) (*State, error) {
	//TODO implement me
	panic("implement me")
}
