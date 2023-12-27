package wavLoader

import (
	"github.com/gravestench/servicemesh"

	"github.com/gravestench/dark-magic/pkg/services/cacheManager"
	"github.com/gravestench/dark-magic/pkg/services/configFile"
)

type Dependency = LoadsWavFiles

type LoadsWavFiles interface {
	servicemesh.Service
	servicemesh.HasLogger
	servicemesh.HasDependencies
	configFile.HasDefaultConfig
	cacheManager.HasCache
	Load(filepath string) ([]byte, error)
}
