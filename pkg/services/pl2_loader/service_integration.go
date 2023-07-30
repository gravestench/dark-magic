package pl2_loader

import (
	"github.com/gravestench/pl2"

	"github.com/gravestench/runtime"
)

var (
	_ runtime.Service         = &Service{}
	_ runtime.HasLogger       = &Service{}
	_ runtime.HasDependencies = &Service{}
	_ LoadsPl2Files           = &Service{}
)

type Dependency = LoadsPl2Files

type LoadsPl2Files = interface {
	Load(filepath string) (*pl2.PL2, error)
}
