package gpl_loader

import (
	"github.com/gravestench/gpl"

	"github.com/gravestench/runtime"
)

var (
	_ runtime.Service         = &Service{}
	_ runtime.HasLogger       = &Service{}
	_ runtime.HasDependencies = &Service{}
	_ LoadsGplFiles           = &Service{}
)

type Dependency = LoadsGplFiles

type LoadsGplFiles = interface {
	Load(filepath string) (*gpl.GPL, error)
}
