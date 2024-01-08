package hero

import (
	"sort"

	"github.com/gravestench/servicemesh"

	"github.com/gravestench/dark-magic/pkg/models"
	"github.com/gravestench/dark-magic/pkg/services/common"
	"github.com/gravestench/dark-magic/pkg/services/configManager"
	"github.com/gravestench/dark-magic/pkg/services/recordManager"
)

type Service struct {
	common.Service
	Config
	cfgHandle      *configManager.ConfigHandle
	records        recordManager.Dependency
	expBreakpoints map[models.Hero][]ExperienceBreakpoint
}

func (s *Service) Init(mesh servicemesh.Mesh) {
	s.expBreakpoints = make(map[models.Hero][]ExperienceBreakpoint)
	s.loadExperienceBreakpoints()
}

func (s *Service) OnServiceMeshShutdownInitiated() {
	// we want to save the heroes when the app shuts down
	if err := s.SaveHeroes(); err != nil {
		s.Logger().Error("saving heroes", "error", err)
	}
}

func (s *Service) Name() string {
	return "Hero Manager"
}

func (s *Service) Ready() bool {
	if s.records == nil {
		return false
	}

	if s.Config == nil {
		return false
	}

	return true
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

	return state
}

func (s *Service) GetHeroes() (heroes []State) {
	for _, state := range s.Config {
		heroes = append(heroes, *state)
	}

	sort.Slice(heroes, func(i, j int) bool {
		return heroes[i].Name < heroes[j].Name
	})

	return
}

func (s *Service) GetHeroByName(name string) *State {
	for _, state := range s.Config {
		if state.Name == name {
			return state
		}
	}

	return nil
}
