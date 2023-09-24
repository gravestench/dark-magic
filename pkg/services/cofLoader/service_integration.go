package cofLoader

import (
	"github.com/gravestench/cof"
	"github.com/gravestench/runtime"
)

var (
	_ runtime.Service         = &Service{}
	_ runtime.HasLogger       = &Service{}
	_ runtime.HasDependencies = &Service{}
	_ LoadsDc6Files           = &Service{}
)

type Dependency = LoadsDc6Files

type LoadsDc6Files = interface {
	Load(filepath string) (*cof.COF, error)
	LoadAnimationData(filepath string) (*cof.AnimationData, error)
}
