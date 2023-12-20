package hero

import (
	"log/slog"

	"github.com/gravestench/servicemesh"

	"github.com/gravestench/dark-magic/pkg/models"
	"github.com/gravestench/dark-magic/pkg/services/configFile"
	"github.com/gravestench/dark-magic/pkg/services/recordManager"
)

type Service struct {
	logger *slog.Logger

	config  *configFile.Config
	records recordManager.Dependency

	expBreakpoints map[models.Hero][]experienceBreakpoint
	heroStates     []State
}

func (s *Service) Init(mesh servicemesh.Mesh) {
	s.expBreakpoints = make(map[models.Hero][]experienceBreakpoint)
	s.heroStates = make([]State, 0)

	if err := s.LoadHeroes(); err != nil {
		s.logger.Error("loading heroes from config", "error", err)
	}

	s.loadExperienceBreakpoints()
}

func (s *Service) Name() string {
	return "Hero Manager"
}

func (s *Service) Ready() bool {
	if s.records == nil {
		return false
	}

	if s.config == nil {
		return false
	}

	return true
}

func (s *Service) SetLogger(logger *slog.Logger) {
	s.logger = logger
}

func (s *Service) Logger() *slog.Logger {
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

func (s *Service) GetHeroes() []State {
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
