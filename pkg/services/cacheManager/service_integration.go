package cacheManager

import (
	"github.com/gravestench/servicemesh"

	"github.com/gravestench/dark-magic/pkg/cache"
	"github.com/gravestench/dark-magic/pkg/services/lua"
	"github.com/gravestench/dark-magic/pkg/services/modalTui"
)

// these are static declarations that force a
// compile-time error if the service does not
// implement them.
var (
	_ servicemesh.Service                = &Service{} // implement in`service.go`
	_ servicemesh.HasLogger              = &Service{} // implement in`service.go`
	_ lua.UsesLuaEnvironment             = &Service{} // implement in`lua_integration.go`
	_ modalTui.HasModalTextUserInterface = &Service{} // implement in`service.go`
	_ IsCacheManager                     = &Service{} // implement in`service.go`
)

// this is an alias which can be used to make
// the dependency resolution methods of other
// services more coherent. It's just sugar.

type Dependency = IsCacheManager

// Here is the declaration of our service as
// an interface. This is all the dependent services
// should know about this service.

type IsCacheManager interface {
	FlushAllCaches()
}

// Here is the declaration of an integration interface
// other services can implement if they want a cache

type HasCache interface {
	CacheBudget() int
	FlushCache(newCache *cache.Cache)
}
