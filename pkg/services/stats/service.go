package stats

import (
	"fmt"

	"github.com/gravestench/runtime"
	"github.com/rs/zerolog"

	"github.com/gravestench/dark-magic/pkg/models"
	"github.com/gravestench/dark-magic/pkg/services/locale"
	"github.com/gravestench/dark-magic/pkg/services/recordManager"
	"github.com/gravestench/dark-magic/pkg/services/tblLoader"
)

// Service is responsible for creating stats
type Service struct {
	logger  *zerolog.Logger
	records recordManager.Dependency
	tables  tblLoader.Dependency
	locale  locale.Dependency
}

func (s *Service) Init(rt runtime.Runtime) {

}

func (s *Service) Name() string {
	return "Stats"
}

func (s *Service) BindLogger(logger *zerolog.Logger) {
	s.logger = logger
}

func (s *Service) Logger() *zerolog.Logger {
	return s.logger
}

// NewStat creates a stat instance with the given record and values
func (s *Service) NewStat(key string, values ...float64) *Stat {
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

	stat := &Stat{
		statService: s,
		statRecord:  record,
	}

	stat.init(values...) // init stat values, value types, and value combination rules

	return stat
}

// NewStatList creates a stat list
func (s *Service) NewStatList(stats ...Stat) *StatList {
	return &StatList{stats}
}

// NewStatValue creates a stat value of the given type
func (s *Service) NewStatValue(t StatNumberType, c ValueCombineType) *StatValue {
	sv := &StatValue{
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
	startingCharStats := s.records.CharStartingAttributes()
	classRecord := startingCharStats[heroIndex]
	str, _ := s.locale.GetTextByKey(classRecord.StrAllSkills)

	return str
}

func (s *Service) stringerClassOnly(sv StatValue) string {
	heroIndex := sv.Int()
	startingCharStats := s.records.CharStartingAttributes()
	classRecord := startingCharStats[heroIndex]
	str, _ := s.locale.GetTextByKey(classRecord.StrClassOnly)

	return str
}

func (s *Service) stringerSkillName(sv StatValue) string {
	const (
		skillNotFound = "{Monster not found!}"
	)

	skills := s.records.Skills()

	for _, skill := range skills {
		if skill.ID == sv.String() {
			return skill.SkillDesc
		}
	}

	return "{Skill not found!}"
}

func (s *Service) stringerMonsterName(sv StatValue) string {
	const (
		monsterNotFound = "{Monster not found!}"
	)

	for _, record := range s.records.MonsterStats() {
		if record.Id == sv.String() {
			return record.NameStr
		}
	}

	return monsterNotFound
}
