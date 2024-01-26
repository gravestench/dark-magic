package gui

import (
	"image"

	"github.com/google/uuid"
	"github.com/gravestench/servicemesh"

	"github.com/gravestench/dark-magic/pkg/services/lua"
	"github.com/gravestench/dark-magic/pkg/services/raylibRenderer"
)

// these are static declarations that force a
// compile-time error if the service does not
// implement them.
var (
	_ servicemesh.Service         = &Service{} // implement in`service.go`
	_ servicemesh.HasLogger       = &Service{} // implement in`service.go`
	_ servicemesh.HasDependencies = &Service{} // implement in`dependencies.go`
	_ lua.UsesLuaEnvironment      = &Service{} // implement in`lua_integration.go`
	_ ManagesGui                  = &Service{} // implement in`service.go`
)

// this is an alias which can be used to make
// the dependency resolution methods of other
// services more coherent. It's just sugar.

type Dependency = ManagesGui

// Here is the declaration of our service as
// an interface. This is all the dependent services
// should know about this service.

type ManagesGui interface {
	servicemesh.Service
	managesTileSprites
	managesAnimations
}

type Key int
type ModifierKey int
type MouseButton int

type InputState struct {
	Keys         []Key
	ModifierKeys []ModifierKey
	MouseButtons []MouseButton
	MouseCursor  image.Point
}

type inputHandler interface {
	InputVector() InputState
	SetInputVector(InputState)
	ClearInputVectors()

	Handler() func()
	SetHandler(func())
}

type element interface {
	raylibRenderer.Renderable
	inputHandler

	UUID() uuid.UUID

	Enable()
	Disable()
	IsEnabled() bool
}

type managesTileSprites interface {
	NewTileSprite(dc6Path, palettePath string, gridWidth, gridHeight int) (ts *TileSprite, err error)
}

type managesAnimations interface {
	NewAnimationDC6(dc6Path, pl2Path string) (*AnimationDC6, error)
}
