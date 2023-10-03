package raylibRenderer

import (
	"image"

	"github.com/gravestench/runtime"

	"github.com/gravestench/dark-magic/pkg/services/configFile"
	"github.com/gravestench/dark-magic/pkg/services/lua"
)

// these are static declarations that force a
// compile-time error if the service does not
// implement them.
var (
	_ runtime.Service             = &Service{} // implement in`service.go`
	_ runtime.HasLogger           = &Service{} // implement in`service.go`
	_ runtime.HasDependencies     = &Service{} // implement in`dependencies.go`
	_ lua.UsesLuaEnvironment      = &Service{} // implement in`lua_integration.go`
	_ configFile.HasDefaultConfig = &Service{} // implement in`lua_integration.go`
	_ IsRenderer                  = &Service{} // implement in`service.go`
)

// this is an alias which can be used to make
// the dependency resolution methods of other
// services more coherent. It's just sugar.

type Dependency = IsRenderer

// Here is the declaration of our service as
// an interface. This is all the dependent services
// should know about this service.

type IsRenderer interface {
	IsInit() bool
	SetWindowTitle(string)
	WindowSize() (width, height int)
	Resolution() (width, height int)
}

// IsRenderableLayer is an integration interface that other services
// can implement to be picked up by this raylib renderer service.
type IsRenderableLayer interface {
	LayerIndex() int
	Position() (x, y int)
	Rotation() (degrees float32)
	Scale() (scale float32)
	Opacity() (opcaity float64)
	LayerImage() image.Image
}
