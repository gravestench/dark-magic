package stats

import (
	"github.com/gravestench/runtime"
)

var (
	_ runtime.Service         = &Service{}
	_ runtime.HasLogger       = &Service{}
	_ runtime.HasDependencies = &Service{}
	_ ManagesStats            = &Service{}
)

type Dependency = ManagesStats

type ManagesStats interface {
	NewStat(key string, values ...float64) *Stat
	NewStatList(stats ...Stat) *StatList
	NewStatValue(t StatNumberType, c ValueCombineType) *StatValue
}
