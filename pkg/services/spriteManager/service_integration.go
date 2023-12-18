package spriteManager

import (
	"github.com/gravestench/servicemesh"

	"github.com/gravestench/dark-magic/pkg/services/configFile"
	"github.com/gravestench/dark-magic/pkg/services/lua"
)

var (
	_ servicemesh.Service         = &Service{}
	_ servicemesh.HasLogger       = &Service{}
	_ servicemesh.HasDependencies = &Service{}
	_ configFile.HasDefaultConfig = &Service{}
	_ lua.UsesLuaEnvironment      = &Service{}
	_ IsSpriteLoader              = &Service{}
)

// this is an alias which can be used to make
// the dependency resolution methods of other
// services more coherent. It's just sugar.

type Dependency = IsSpriteLoader

// Here is the declaration of our service as
// an interface. This is all the dependent services
// should know about this service.

type IsSpriteLoader interface {
	LoadDc6ToPngSpriteAtlas(pathDC6 string, pathPL2 string) ([]byte, error)
	LoadDc6ToAnimatedGif(pathDC6 string, pathPL2 string) ([]byte, error)

	LoadDccToPngSpriteAtlas(pathDC6 string, pathPL2 string) ([]byte, error)
	LoadDccToAnimatedGif(pathDC6 string, pathPL2 string) ([]byte, error)

	LoadDt1ToPngSpriteAtlas(pathDC6 string, pathPL2 string) ([]byte, error)

	LoadDs1ToPngSpriteAtlas(pathDC6 string, pathPL2 string) ([]byte, error)
}
