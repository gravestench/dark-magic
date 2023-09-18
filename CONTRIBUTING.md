## Conventions

### [Service Template](./internal/service_template)
Here is an outline of a typical service package:
```text
pkg/services/foo
├── README.md
├── dependencies.go
├── lua_integration.go
├── service.go
└── service_integration.go
```

For brevity, we have included only a single service integration 
file (other than the one for this service). If there are any interfaces
that are being implemented which provide integration with another service,
then those methods belong in a file which denotes said integration. 

For instance, if we were integrating with the web router service, we would
also add a `web_router_integration.go` file with the respective integration
methods.
_____________
#### Step 1: Declare the interface of your service
When creating a new service, start with the `service_integration.go` file. This 
file serves to declare what interfaces the service will implement, as well as
declare an interface for use as a dependency resolution interface when other 
services depend upon it.
```golang
// these are static declarations that force a 
// compile-time error if the service does not 
// implement them. 
var (
	_ runtime.Service         = &Service{}
	_ runtime.HasLogger       = &Service{}
	_ runtime.HasDependencies = &Service{}
	_ lua.UsesLuaEnvironment  = &Service{}
	_ IsFooService        = &Service{}
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
```

_____________
#### Step 2: write your concrete implementation
Next, implement the concrete service in `service.go`. The service should
always be called `Service` so that when importing there is no stutter in 
the name being imported (eg `foo.Service`).

Here are some main points:
* The `runtime.Service` and `runtime.HasLogger` methods belong in `service.go`.
* The methods that are described by this service's integration interface are defined here

_____________
#### Step 3: Implement other intergration interfaces in separate files
For example:
* methods for runtime dependency resolution belong in `dependencies.go`
* methods for integrating with the lua service belong in `lua_integration.go`
* methods for integrating with the web router belong in `web_router_integration.go`

_____________
#### Step 4: Write your readme
in `README.md`, declare the following:
* purpose of the service
* dependencies on other services
* how other services can integrate with this service
* how this service integrates with other services
