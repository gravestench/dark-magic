package webRouter

import (
	"github.com/gin-gonic/gin"

	"github.com/gravestench/servicemesh"

	"github.com/gravestench/dark-magic/pkg/services/configFile"
)

var (
	_ servicemesh.Service         = &Service{}
	_ servicemesh.HasLogger       = &Service{}
	_ servicemesh.HasDependencies = &Service{}
	_ configFile.HasDefaultConfig = &Service{}
	_ IsWebRouter                 = &Service{}
)

type Dependency = IsWebRouter

// IsWebRouter is just responsible for yielding the root route handler.
// Services will use this in order to set up their own routes.
type IsWebRouter interface {
	RouteRoot() *gin.Engine
	Reload()
}

// IsRouteInitializer is a type of service that will
// set up its own web routes using a base route group
type IsRouteInitializer interface {
	servicemesh.Service
	InitRoutes(*gin.RouterGroup)
}

// HasRouteSlug describes a service that has an identifier that is used
// as a prefix for its subroutes
type HasRouteSlug interface {
	Slug() string
}
