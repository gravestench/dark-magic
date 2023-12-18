package dt1Loader

import (
	"github.com/gravestench/dt1"
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
	_ LoadsDt1Files               = &Service{}
)

type Dependency = LoadsDt1Files

type LoadsDt1Files = interface {
	Load(filepath string) (*dt1.DT1, error)
}
