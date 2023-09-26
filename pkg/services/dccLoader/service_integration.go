package dccLoader

import (
	"github.com/gravestench/dcc"

	"github.com/gravestench/runtime"

	"github.com/gravestench/dark-magic/pkg/services/cacheManager"
	"github.com/gravestench/dark-magic/pkg/services/configFile"
)

var (
	_ runtime.Service             = &Service{}
	_ runtime.HasLogger           = &Service{}
	_ runtime.HasDependencies     = &Service{}
	_ configFile.HasDefaultConfig = &Service{}
	_ cacheManager.HasCache       = &Service{}
	_ LoadsDccFiles               = &Service{}
)

type Dependency = LoadsDccFiles

type LoadsDccFiles = interface {
	Load(filepath string) (*dcc.DCC, error)
}
