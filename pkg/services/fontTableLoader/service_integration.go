package fontTableLoader

import (
	"github.com/gravestench/font_table"
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
	_ LoadsFontTableFiles         = &Service{}
)

type Dependency = LoadsFontTableFiles

type LoadsFontTableFiles = interface {
	Load(filepath string) (*font_table.Font, error)
}
