package raylibRenderer

import (
	"image"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/google/uuid"
	"github.com/gravestench/servicemesh"

	"github.com/gravestench/dark-magic/pkg/services/cacheManager"
	"github.com/gravestench/dark-magic/pkg/services/configFile"
)

// these are static declarations that force a
// compile-time error if the service does not
// implement them.
var (
	_ servicemesh.Service         = &Service{} // implement in`service.go`
	_ servicemesh.HasLogger       = &Service{} // implement in`service.go`
	_ servicemesh.HasDependencies = &Service{} // implement in`dependencies.go`
	_ configFile.HasDefaultConfig = &Service{} // implement in`lua_integration.go`
	_ cacheManager.HasCache       = &Service{} // implement in`lua_integration.go`
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
	ManagesWindow
	ManagesCameras
	ProvidesTextures
	ProvidesRenderables
}

type ManagesWindow interface {
	SetWindowTitle(string)
	WindowSize() (width, height int)
	Resolution() (width, height int)
}

type ManagesCameras interface {
	GetDefaultCamera() *rl.Camera2D
	AddCamera(name string) *rl.Camera2D
	GetCamera(name string) *rl.Camera2D
	RemoveCamera(name string)
}

type ProvidesTextures interface {
	GetTexture(uuid.UUID, image.Image) (texture rl.Texture2D, isNew bool)
}

type ProvidesRenderables interface {
	NewRenderable() Renderable
}

// Renderable is a thing that the renderer provides to other services, which
// encapsulates the necessary behavior of something that can be rendered
type Renderable interface {
	UUID() uuid.UUID
	dirty() bool

	hasChildren
	hasUpdate
	hasMatrix
	hasOrigin

	ZIndex() float32
	SetZIndex(float32)

	Position() (x, y float32)
	SetPosition(x, y float32)

	Rotation() (degrees float32)
	SetRotation(degrees float32)

	Scale() (scale float32)
	SetScale(scale float32)

	Opacity() (opacity float32)
	SetOpacity(opacity float32)

	BlendMode() (mode rl.BlendMode)
	SetBlendMode(mode rl.BlendMode)

	Texture() rl.Texture2D
	SetTexture(rl.Texture2D)

	Image() image.Image
	SetImage(image.Image)

	Enable()
	Disable()
	IsEnabled() bool
}

type hasUpdate interface {
	update()
	OnUpdate(func())
}

type hasOrigin interface {
	Origin() rl.Vector2
	SetOrigin(x, y float64)
}

type hasChildren interface {
	SetParent(Renderable)
	addChild(Renderable)
	removeChild(Renderable)
	Children() []Renderable
}

type hasMatrix interface {
	WorldMatrix() rl.Matrix
	LocalMatrix() rl.Matrix
	UpdateWorldMatrix(parent rl.Matrix)
}
