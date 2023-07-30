package dcc_loader

import (
	"github.com/gravestench/dcc"

	"github.com/gravestench/runtime"
)

var (
	_ runtime.Service         = &Service{}
	_ runtime.HasLogger       = &Service{}
	_ runtime.HasDependencies = &Service{}
	_ LoadsDccFiles           = &Service{}
)

type Dependency = LoadsDccFiles

type LoadsDccFiles = interface {
	Load(filepath string) (*dcc.DCC, error)
}
