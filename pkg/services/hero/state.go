package hero

import (
	"github.com/gravestench/dark-magic/pkg/models"
)

type State struct {
	Name                  string
	experience            int
	Type                  models.Hero
	level                 int
	experienceProgression []experienceBreakpoint // comes from character class records
}

type experienceBreakpoint struct {
	Experience int
	Ratio      float32
}

func (s *State) Level() int {
	if s.level != 0 {
		return s.level
	}

	level := 1

	for idx, progression := range s.experienceProgression {
		if s.experience > progression.Experience {
			continue
		}

		level += idx

		break
	}

	s.level = level

	return level
}

func (s *State) ExperienceRatio() float32 {
	return s.experienceProgression[s.Level()-1].Ratio
}

func (s *State) SetExperience(set int) {
	s.experience = set
	s.level = 0
}

func (s *State) AddExperience(amount int) {
	ratio := s.ExperienceRatio()
	s.level = 0

	s.experience += int(ratio * float32(amount))
}

func (s *State) SubtractExperience(amount int) {
	s.AddExperience(-amount)
}
