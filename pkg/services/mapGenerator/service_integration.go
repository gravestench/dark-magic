package mapGenerator

import (
	"github.com/gravestench/servicemesh"

	"github.com/gravestench/dark-magic/pkg/services/webRouter"
)

var (
	_ servicemesh.Service          = &Service{}
	_ servicemesh.HasLogger        = &Service{}
	_ servicemesh.HasDependencies  = &Service{}
	_ webRouter.IsRouteInitializer = &Service{}
	_ GeneratesDiablo2Maps         = &Service{}
)

type Dependency = GeneratesDiablo2Maps

type GeneratesDiablo2Maps interface {
	Seed() uint64
	SetSeed(uint64)
	GenerateMap(act, difficulty uint) (*WorldMap, error)
}
