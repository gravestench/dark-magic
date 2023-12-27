package tsvLoader

import (
	"github.com/gravestench/servicemesh"
)

var (
	_ servicemesh.Service         = &Service{}
	_ servicemesh.HasLogger       = &Service{}
	_ servicemesh.HasDependencies = &Service{}
	_ LoadsTsvFiles               = &Service{}
)

type Dependency = LoadsTsvFiles

type LoadsTsvFiles = interface {
	servicemesh.Service
	Load(filepath string) ([]byte, error)
	Unmarshal(filepath string, destination any) error
}
