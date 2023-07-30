package map_generator

import (
	"github.com/gravestench/runtime"

	"github.com/gravestench/dark-magic/pkg/models"
)

var (
	_ runtime.Service         = &Service{}
	_ runtime.HasLogger       = &Service{}
	_ runtime.HasDependencies = &Service{}
	_ GeneratesDiablo2Maps    = &Service{}
)

type Dependency = GeneratesDiablo2Maps

type GeneratesDiablo2Maps interface {
	Seed() uint64
	SetSeed(uint64)
	GenerateMap(act, difficulty uint) (models.Diablo2Map, error)
}
