package tsvLoader

import (
	"github.com/gravestench/runtime"
)

var (
	_ runtime.Service         = &Service{}
	_ runtime.HasLogger       = &Service{}
	_ runtime.HasDependencies = &Service{}
	_ LoadsTsvFiles           = &Service{}
)

type Dependency = LoadsTsvFiles

type LoadsTsvFiles = interface {
	Load(filepath string, destination any) error
}
