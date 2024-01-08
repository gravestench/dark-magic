package hero

import (
	"github.com/gravestench/dark-magic/pkg/models"
)

func (s *Service) loadExperienceBreakpoints() {
	const divisor = 1024

	for idx, record := range s.records.ExperienceBreakpoints() {
		if idx == 0 {
			// the records file has the max level as the first record...
			continue
		}

		s.expBreakpoints[models.HeroBarbarian] = append(s.expBreakpoints[models.HeroBarbarian], ExperienceBreakpoint{
			Experience: record.Barbarian,
			Ratio:      float32(record.ExpRatio) / divisor,
		})

		s.expBreakpoints[models.HeroNecromancer] = append(s.expBreakpoints[models.HeroNecromancer], ExperienceBreakpoint{
			Experience: record.Necromancer,
			Ratio:      float32(record.ExpRatio) / divisor,
		})

		s.expBreakpoints[models.HeroPaladin] = append(s.expBreakpoints[models.HeroPaladin], ExperienceBreakpoint{
			Experience: record.Paladin,
			Ratio:      float32(record.ExpRatio) / divisor,
		})

		s.expBreakpoints[models.HeroAssassin] = append(s.expBreakpoints[models.HeroAssassin], ExperienceBreakpoint{
			Experience: record.Assassin,
			Ratio:      float32(record.ExpRatio) / divisor,
		})

		s.expBreakpoints[models.HeroSorceress] = append(s.expBreakpoints[models.HeroSorceress], ExperienceBreakpoint{
			Experience: record.Sorceress,
			Ratio:      float32(record.ExpRatio) / divisor,
		})

		s.expBreakpoints[models.HeroAmazon] = append(s.expBreakpoints[models.HeroAmazon], ExperienceBreakpoint{
			Experience: record.Amazon,
			Ratio:      float32(record.ExpRatio) / divisor,
		})

		s.expBreakpoints[models.HeroDruid] = append(s.expBreakpoints[models.HeroDruid], ExperienceBreakpoint{
			Experience: record.Druid,
			Ratio:      float32(record.ExpRatio) / divisor,
		})

	}
}
