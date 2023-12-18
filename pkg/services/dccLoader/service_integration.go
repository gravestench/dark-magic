package dccLoader

import (
	"github.com/gravestench/dcc"

	"github.com/gravestench/servicemesh"

	"github.com/gravestench/dark-magic/pkg/services/cacheManager"
	"github.com/gravestench/dark-magic/pkg/services/configFile"
)

var (
	_ servicemesh.Service         = &Service{}
	_ servicemesh.HasLogger       = &Service{}
	_ servicemesh.HasDependencies = &Service{}
	_ configFile.HasDefaultConfig = &Service{}
	_ cacheManager.HasCache       = &Service{}
	_ LoadsDccFiles               = &Service{}
)

type Dependency = LoadsDccFiles

type LoadsDccFiles = interface {
	Load(filepath string) (*dcc.DCC, error)
}
