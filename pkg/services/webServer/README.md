# Web Router Service
The purpose of this [runtime](https://github.com/gravestench/runtime) service is
to provide a locally hosted web server that integrates with the web router 
service. 

The web router service integrates with other runtime services to expose 
routes which are used for retrieving data from the MPQ's, or marshalled 
record models.

## Dependencies
This service has dependencies on the the following services:
* [config file manager](../configFile)
* [web router service](../webRouter)

## Integration with other services
This service exports an integration interface `IsWebServer` with an alias
`Dependencncy` which are intended to be used by other services for dependency
resolution (see runtime.HasDependencies), and expose just the methods which
other services should use.
```golang
type Dependency = IsWebServer

type IsWebServer interface {
    RestartServer()
    StartServer()
    StopServer()
}
```

Other services should use the `IsWebServer` or `Dependency` interfaces to resolve
their dependency on this service.
