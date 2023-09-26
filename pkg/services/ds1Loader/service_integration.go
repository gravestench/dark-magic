package ds1Loader

import (
	"github.com/gravestench/ds1"

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
	_ LoadsDs1Files               = &Service{}
)

type Dependency = LoadsDs1Files

type LoadsDs1Files = interface {
	Load(filepath string) (*ds1.DS1, error)
}
