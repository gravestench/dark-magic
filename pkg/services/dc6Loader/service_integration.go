package dc6Loader

import (
	dc6 "github.com/gravestench/dc6/pkg"

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
	_ LoadsDc6Files               = &Service{}
)

type Dependency = LoadsDc6Files

type LoadsDc6Files = interface {
	Load(filepath string) (*dc6.DC6, error)
}
