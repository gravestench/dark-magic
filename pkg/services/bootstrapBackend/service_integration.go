package service_template

import (
	"github.com/gravestench/servicemesh"

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
	_ IsFooService                = &Service{} // implement in`service.go`
)

// this is an alias which can be used to make
// the dependency resolution methods of other
// services more coherent. It's just sugar.

type Dependency = IsFooService

// Here is the declaration of our service as
// an interface. This is all the dependent services
// should know about this service.

type IsFooService interface {
	Foo()
}
