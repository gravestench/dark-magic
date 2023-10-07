package modalGameUI

import (
	"github.com/gravestench/runtime"

	"github.com/gravestench/dark-magic/pkg/services/lua"
	"github.com/gravestench/dark-magic/pkg/services/raylibRenderer"
)

// these are static declarations that force a
// compile-time error if the service does not
// implement them.
var (
	_ runtime.Service         = &Service{} // implement in`service.go`
	_ runtime.HasLogger       = &Service{} // implement in`service.go`
	_ runtime.HasDependencies = &Service{} // implement in`dependencies.go`
	_ lua.UsesLuaEnvironment  = &Service{} // implement in`lua_integration.go`
	_ ManagesModalGameUI      = &Service{} // implement in`service.go`
)

// this is an alias which can be used to make
// the dependency resolution methods of other
// services more coherent. It's just sugar.

type Dependency = ManagesModalGameUI

// Here is the declaration of our service as
// an interface. This is all the dependent services
// should know about this service.

type ManagesModalGameUI interface {
	Modes() []string
	Mode() string
	SetMode(string)
}

// ModalGameUI is an intergartion interface that other services can implement to be
// picked up by this service.
type ModalGameUI interface {
	Mode() string
	Renderable() raylibRenderer.Renderable
	Update()
	OnUpdate(func())
}
