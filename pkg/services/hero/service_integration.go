package hero

import (
	"github.com/gravestench/servicemesh"

	"github.com/gravestench/dark-magic/pkg/models"
	"github.com/gravestench/dark-magic/pkg/services/configFile"
	"github.com/gravestench/dark-magic/pkg/services/lua"
)

// these are static declarations that force a
// compile-time error if the service does not
// implement them.
var (
	_ servicemesh.Service         = &Service{} // implement in`service.go`
	_ servicemesh.HasLogger       = &Service{} // implement in`service.go`
	_ servicemesh.HasDependencies = &Service{} // implement in`dependencies.go`
	_ lua.UsesLuaEnvironment      = &Service{} // implement in`lua_integration.go`
	_ configFile.HasConfig        = &Service{} // implement in`lua_integration.go`
	_ ManagesHeroes               = &Service{} // implement in`service.go`
)

// this is an alias which can be used to make
// the dependency resolution methods of other
// services more coherent. It's just sugar.

type Dependency = ManagesHeroes

// Here is the declaration of our service as
// an interface. This is all the dependent services
// should know about this service.

type ManagesHeroes interface {
	LoadHeroes() error
	SaveHeroes() error
	CreateHero(name string, hero models.Hero) State
	GetHeroes() []State
	GetHeroByName(name string) *State
}
