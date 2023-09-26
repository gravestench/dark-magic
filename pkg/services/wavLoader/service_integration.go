package wavLoader

import (
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
	_ LoadsWavFiles               = &Service{}
)

type Dependency = LoadsWavFiles

type LoadsWavFiles = interface {
	Load(filepath string) ([]byte, error)
}
