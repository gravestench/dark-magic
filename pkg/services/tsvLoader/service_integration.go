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
	Load(filepath string) ([]byte, error)
	Unmarshal(filepath string, destination any) error
}
