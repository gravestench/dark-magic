package raylibRenderer

import (
	"image"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/google/uuid"
	"github.com/gravestench/runtime"

	"github.com/gravestench/dark-magic/pkg/services/cacheManager"
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

	ZIndex() int
	SetZIndex(int)

	Position() (x, y int)
	SetPosition(x, y int)

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
}
