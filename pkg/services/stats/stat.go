package stats

import (
	"github.com/gravestench/dark-magic/pkg/models"
)

// Stat a generic interface for a stat. It is something which can be
// combined with other stats, holds one or more values, and handles the
// way that it is printed as a string
type Stat struct {
	statService *Service
	statRecord  *models.ItemStatCost
}

func (s *Stat) init(values ...float64) string {
	//TODO implement me
	panic("implement me")
}

func (s *Stat) Name() string {
	//TODO implement me
	panic("implement me")
}

func (s *Stat) Clone() Stat {
	//TODO implement me
	panic("implement me")
}

func (s *Stat) Copy(s2 Stat) Stat {
	//TODO implement me
	panic("implement me")
}

func (s *Stat) Combine(s2 Stat) (combined Stat, err error) {
	//TODO implement me
	panic("implement me")
}

func (s *Stat) String() string {
	//TODO implement me
	panic("implement me")
}

func (s *Stat) Values() []StatValue {
	//TODO implement me
	panic("implement me")
}

func (s *Stat) SetValues(value ...StatValue) {
	//TODO implement me
	panic("implement me")
}

func (s *Stat) Priority() int {
	//TODO implement me
	panic("implement me")
}
