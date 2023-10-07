# Template service
The purpose of this [runtime](https://github.com/gravestench/runtime) service is to provide
a modal user interface, meaning that only one UI (mode) is rendering at any given time.


## Dependencies
This service has dependencies on the following services:
* [gui manager](../guiManager)
* [renderer](../raylibRenderer)
* [input](../input)


## Integration with other services
This service integrates with the following services:
* [lua](../lua)

Lua integration is optional, meaning that the lua service can be omitted at runtime.

_______
This service exports an integration interface `IsFoo` with an alias 
`Dependencncy` which are intended to be used by other services for dependency
resolution (see runtime.HasDependencies), and expose just the methods which 
other services should use.
```golang
type Dependency = Foo

type IsFoo interface {
    Foo()
}
```

## Lua service integration
Describe how this service integrates with the lua service (this is just an example).

You should show an example of the lua API usage:
```lua
data = assets.load("/data/global/ui/Loading/loadingscreen.dc6")
```