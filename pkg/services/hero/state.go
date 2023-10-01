package hero

import (
	"github.com/gravestench/dark-magic/pkg/models"
)

type State struct {
	Name                  string
	Experience            int
	Class                 models.Hero `json:"Class,string"`
	Level                 int
	record                models.CharStats
	experienceProgression []experienceBreakpoint // comes from character class records
	Attributes            map[string]int
	Skills                []string
}

type experienceBreakpoint struct {
	Experience int
	Ratio      float32
}

func (s *State) updateLevel() int {
	if s.Level != 0 {
		return s.Level
	}

	level := 1

	for idx, progression := range s.experienceProgression {
		if s.Experience > progression.Experience {
			continue
		}

		level += idx

		break
	}

	s.Level = level

	return level
}

func (s *State) ExperienceRatio() float32 {
	s.updateLevel()
	return s.experienceProgression[s.Level-1].Ratio
}

func (s *State) SetExperience(set int) {
	s.Experience = set
	s.Level = 0
	s.updateLevel()
}

func (s *State) AddExperience(amount int) {
	ratio := s.ExperienceRatio()
	s.Level = 0

	s.Experience += int(ratio * float32(amount))
}

func (s *State) SubtractExperience(amount int) {
	s.AddExperience(-amount)
}

func (s *State) SetAttribute(key string, value int) {
	if s.Attributes == nil {
		s.Attributes = make(map[string]int)
	}

	s.Attributes[key] = value
}
