package cofLoader

import (
	"github.com/gravestench/cof"
	"github.com/gravestench/servicemesh"
)

var (
	_ servicemesh.Service         = &Service{}
	_ servicemesh.HasLogger       = &Service{}
	_ servicemesh.HasDependencies = &Service{}
	_ LoadsDc6Files               = &Service{}
)

type Dependency = LoadsDc6Files

type LoadsDc6Files = interface {
	Load(filepath string) (*cof.COF, error)
	LoadAnimationData(filepath string) (*cof.AnimationData, error)
}
