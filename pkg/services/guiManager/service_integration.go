package guiManager

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

//
//type Node interface {
//	element
//
//	Parent() Node
//	SetParent(Node)
//
//	Children() []Node
//	AddChild(child Node)
//	RemoveChild(child Node)
//	ContainsChild(child Node) bool
//
//	LayerIndexOf(child Node) int
//	SetLayerIndexOf(child Node, index int)
//
//	BringToTop(child Node)
//	BringToBottom(child Node)
//	Raise(child Node)
//	Lower(child Node)
//
//	HandleInput(InputState) (terminate bool)
//
//	Image() image.Image
//	SetImage(image.Image)
//
//	Update()
//	UpdateFunc() func()
//	SetUpdateFunc(func())
//
//	GetRelativePosition(target Node) (relativePosition image.Point, found bool)
//}
