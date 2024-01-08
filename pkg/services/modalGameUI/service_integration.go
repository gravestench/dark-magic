package modalGameUI

import (
	"github.com/gravestench/servicemesh"

	"github.com/gravestench/dark-magic/pkg/services/luaManager"
	"github.com/gravestench/dark-magic/pkg/services/raylibRenderer"
)

// these are static declarations that force a
// compile-time error if the service does not
// implement them.
var (
	_ servicemesh.Service         = &Service{}
	_ servicemesh.HasLogger       = &Service{}
	_ servicemesh.HasDependencies = &Service{}
	_ luaManager.LuaPlugin        = &Service{}
	_ ManagesModalGameUI          = &Service{}
)

// this is an alias which can be used to make
// the dependency resolution methods of other
// services more coherent. It's just sugar.

type Dependency = ManagesModalGameUI

// Here is the declaration of our service as
// an interface. This is all the dependent services
// should know about this service.

type ManagesModalGameUI interface {
	//AddMode(ModalGameUI)
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
}
