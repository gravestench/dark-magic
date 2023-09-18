# Web Router Service
The purpose of this [runtime](https://github.com/gravestench/runtime) service is
to provide a way for other services to expose web routes on a locally hosted 
web server. 

At the time of writing, many of the exposed routes are used for
debugging purposes like extracting files from the MPQ's or yielding the
marshalled records in json format.

## Dependencies
This service has a single dependency on the config manager service
* [config file manager](../config_file)

It should be noted that while this service may operate without the
[web server service](../web_server), it is only used by said service. The 
absence of the web server renders this service inert.

## Integration with other services
This service exports an integration interface `IsWebRouter` with an alias
`Dependencncy` which are intended to be used by other services for dependency
resolution (see runtime.HasDependencies), and expose just the methods which
other services should use.
```golang
type Dependency = IsWebRouter

// IsWebRouter is just responsible for yielding the root route handler.
// Services will use this in order to set up their own routes.
type IsWebRouter interface {
    RouteRoot() *gin.Engine
    Reload()
}
```

Other services should use the `IsWebRouter` or `Dependency` interfaces to resolve
their dependency on this service.

_____________

The following interfaces can be optionally implemented by other runtime services
to expose web routes on the locally hosted web server.

```golang
// HasRouteSlug describes a service that has an identifier that is used
// as a prefix for its subroutes
type HasRouteSlug interface {
    Slug() string
}
```
```golang
// IsRouteInitializer is a type of service that will
// set up its own web routes using a base route group
type IsRouteInitializer interface {
    runtime.Service
    InitRoutes(*gin.RouterGroup)
}
```
