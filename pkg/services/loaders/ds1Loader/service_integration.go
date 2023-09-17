package ds1Loader

import (
	"github.com/gravestench/ds1"

	"github.com/gravestench/runtime"
)

var (
	_ runtime.Service         = &Service{}
	_ runtime.HasLogger       = &Service{}
	_ runtime.HasDependencies = &Service{}
	_ LoadsDs1Files           = &Service{}
)

type Dependency = LoadsDs1Files

type LoadsDs1Files = interface {
	Load(filepath string) (*ds1.DS1, error)
}
