package hero

import (
	"github.com/gravestench/servicemesh"

	"github.com/gravestench/dark-magic/pkg/models"
	"github.com/gravestench/dark-magic/pkg/services/configManager"
	"github.com/gravestench/dark-magic/pkg/services/luaManager"
)

// these are static declarations that force a
// compile-time error if the service does not
// implement them.
var (
	_ servicemesh.Service            = &Service{} // implement in`service.go`
	_ servicemesh.HasLogger          = &Service{} // implement in`service.go`
	_ servicemesh.HasDependencies    = &Service{} // implement in`dependencies.go`
	_ luaManager.LuaPlugin           = &Service{} // implement in`lua_integration.go`
	_ configManager.HasConfiguration = &Service{} // implement in`lua_integration.go`
	_ ManagesHeroes                  = &Service{} // implement in`service.go`
)

// this is an alias which can be used to make
// the dependency resolution methods of other
// services more coherent. It's just sugar.

type Dependency = ManagesHeroes

// Here is the declaration of our service as
// an interface. This is all the dependent services
// should know about this service.

type ManagesHeroes interface {
	ReloadHeroes() error
	SaveHeroes() error
	CreateHero(name string, hero models.Hero) State
	GetHeroes() []State
	GetHeroByName(name string) *State
}
