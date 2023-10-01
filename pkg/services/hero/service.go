package hero

import (
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
	heroStates     []State
}

func (s *Service) Init(rt runtime.Runtime) {
	s.expBreakpoints = make(map[models.Hero][]experienceBreakpoint)
	s.heroStates = make([]State, 0)
	s.LoadHeros()

	s.loadExperienceBreakpoints()
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

func (s *Service) CreateHero(name string, hero models.Hero) State {
	state := State{
		Name:                  name,
		Experience:            0,
		Class:                 hero,
		experienceProgression: s.expBreakpoints[hero],
	}

	for _, record := range s.records.CharStartingAttributes() {
		if record.Class == hero.String() {
			state.record = record
		}
	}

	state.SetAttribute("strength", state.record.Strength)
	state.SetAttribute("dexterity", state.record.Dexterity)
	state.SetAttribute("vitality", state.record.Vitality)
	state.SetAttribute("energy", state.record.Intelligence)

	for _, skillID := range []string{
		state.record.Skill1,
		state.record.Skill2,
		state.record.Skill3,
		state.record.Skill4,
		state.record.Skill5,
		state.record.Skill6,
		state.record.Skill7,
		state.record.Skill8,
		state.record.Skill9,
		state.record.Skill10,
	} {
		if skillID == "" {
			continue
		}

		state.Skills = append(state.Skills, skillID)
	}

	s.heroStates = append(s.heroStates, state)

	return state
}

func (s *Service) GetHeros() []State {
	return s.heroStates
}

func (s *Service) GetHeroByName(name string) *State {
	for _, state := range s.heroStates {
		if state.Name == name {
			return &state
		}
	}

	return nil
}
