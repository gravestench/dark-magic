package dc6Loader

import (
	dc6 "github.com/gravestench/dc6/pkg"

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
	_ LoadsDc6Files               = &Service{}
)

type Dependency = LoadsDc6Files

type LoadsDc6Files = interface {
	Load(filepath string) (*dc6.DC6, error)
}
