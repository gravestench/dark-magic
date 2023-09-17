package stats

import (
	"fmt"

	"github.com/gravestench/runtime"

	"github.com/gravestench/dark-magic/pkg/models"
	"github.com/gravestench/dark-magic/pkg/services/loaders/tblLoader"
	"github.com/gravestench/dark-magic/pkg/services/record_manager"
)

type recipe interface {
	runtime.Service
	runtime.HasDependencies
}

// Service is responsible for creating stats
type Service struct {
	records record_manager.Dependency
	tables  tblLoader.Dependency
}

func (s *Service) Init(rt runtime.Runtime) {

}

func (s *Service) Name() string {
	return "Stats"
}

func (s *Service) DependenciesResolved() bool {
	if s.records == nil {
		return false
	}

	if s.tables == nil {
		return false
	}

	return true
}

func (s *Service) ResolveDependencies(rt runtime.Runtime) {
	for _, service := range rt.Services() {
		switch candidate := service.(type) {
		case record_manager.Dependency:
			s.records = candidate
		case tblLoader.Dependency:
			s.tables = candidate
		}
	}
}

// NewStat creates a stat instance with the given record and values
func (s *Service) NewStat(key string, values ...float64) Stat {
	var record *models.ItemStatCost

	for _, candidate := range s.records.ItemStatCost() {
		if candidate.Stat != key {
			continue
		}

		record = &candidate
		break
	}

	if record == nil {
		return nil
	}

	stat := &diablo2Stat{
		factory: s,
		record:  record,
	}

	stat.init(values...) // init stat values, value types, and value combination rules

	return stat
}

// NewStatList creates a stat list
func (s *Service) NewStatList(stats ...Stat) StatList {
	return &Diablo2StatList{stats}
}

// NewValue creates a stat value of the given type
func (s *Service) NewValue(t StatNumberType, c ValueCombineType) StatValue {
	sv := &Diablo2StatValue{
		numberType:  t,
		combineType: c,
	}

	switch t {
	case StatValueFloat:
		sv.stringerFn = s.stringerUnsignedFloat
	case StatValueInt:
		sv.stringerFn = s.stringerUnsignedInt
	default:
		sv.stringerFn = s.stringerEmpty
	}

	return sv
}

const (
	monsterNotFound = "{Monster not found!}"
)

func (s *Service) getHeroMap() []models.Hero {
	return []models.Hero{
		models.HeroAmazon,
		models.HeroSorceress,
		models.HeroNecromancer,
		models.HeroPaladin,
		models.HeroBarbarian,
		models.HeroDruid,
		models.HeroAssassin,
	}
}

func (s *Service) stringerUnsignedInt(sv StatValue) string {
	return fmt.Sprintf("%d", sv.Int())
}

func (s *Service) stringerUnsignedFloat(sv StatValue) string {
	return fmt.Sprintf("%.2f", sv.Float())
}

func (s *Service) stringerEmpty(_ StatValue) string {
	return ""
}

func (s *Service) stringerIntSigned(sv StatValue) string {
	return fmt.Sprintf("%+d", sv.Int())
}

func (s *Service) stringerIntPercentageSigned(sv StatValue) string {
	return fmt.Sprintf("%+d%%", sv.Int())
}

func (s *Service) stringerIntPercentageUnsigned(sv StatValue) string {
	return fmt.Sprintf("%d%%", sv.Int())
}

func (s *Service) stringerClassAllSkills(sv StatValue) string {
	heroIndex := sv.Int()

	heroMap := s.getHeroMap()
	classRecord := s.records.CharStartingAttributes().Stats[heroMap[heroIndex]]

	return s.records.TranslateString(classRecord.SkillStrAll)
}

func (s *Service) stringerClassOnly(sv StatValue) string {
	heroMap := s.getHeroMap()
	heroIndex := sv.Int()
	classRecord := s.records.Character.Stats[heroMap[heroIndex]]
	classOnlyKey := classRecord.SkillStrClassOnly

	return s.records.TranslateString(classOnlyKey)
}

func (s *Service) stringerSkillName(sv StatValue) string {
	skillRecord := s.records.Skill.Details[sv.Int()]
	return skillRecord.Skill
}

func (s *Service) stringerMonsterName(sv StatValue) string {
	for key := range s.records.Monster.Stats {
		if s.records.Monster.Stats[key].ID == sv.Int() {
			return s.records.Monster.Stats[key].NameString
		}
	}

	return monsterNotFound
}
