package guiManager

import (
	"image"

	"github.com/google/uuid"
	"github.com/gravestench/runtime"

	"github.com/gravestench/dark-magic/pkg/services/lua"
	"github.com/gravestench/dark-magic/pkg/services/raylibRenderer"
)

// these are static declarations that force a
// compile-time error if the service does not
// implement them.
var (
	_ runtime.Service                  = &Service{} // implement in`service.go`
	_ runtime.HasLogger                = &Service{} // implement in`service.go`
	_ runtime.HasDependencies          = &Service{} // implement in`dependencies.go`
	_ lua.UsesLuaEnvironment           = &Service{} // implement in`lua_integration.go`
	_ raylibRenderer.IsRenderableLayer = &Service{} // implement in`lua_integration.go`
	_ ManagesGui                       = &Service{} // implement in`service.go`
)

// this is an alias which can be used to make
// the dependency resolution methods of other
// services more coherent. It's just sugar.

type Dependency = ManagesGui

// Here is the declaration of our service as
// an interface. This is all the dependent services
// should know about this service.

type ManagesGui interface {
	NewNode() Node
	Nodes() []Node
	Update()
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
	inputHandler

	UUID() uuid.UUID

	Enable()
	Disable()
	IsEnabled() bool

	Position() (point image.Point)
	SetPosition(point image.Point)
}

type Node interface {
	element

	Parent() Node
	SetParent(Node)

	Children() []Node
	AddChild(child Node)
	RemoveChild(child Node)

	LayerIndexOf(child Node) int
	SetLayerIndexOf(child Node, index int)

	BringToTop(child Node)
	BringToBottom(child Node)
	Raise(child Node)
	Lower(child Node)

	HandleInput(InputState) (terminate bool)

	Image() image.Image
	ImageFunc() func() image.Image
	SetImageFunc(func() image.Image)

	Update()
	UpdateFunc() func()
	SetUpdateFunc(func())

	GetRelativePosition(target Node) (relativePosition image.Point, found bool)
}
