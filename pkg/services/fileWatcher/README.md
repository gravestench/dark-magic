# File Watcher Service
The purpose of this [servicemesh](https://github.com/gravestench/servicemesh) service is to provide a single service
responsible watching for filesystem changes to files and calling handler
functions when those files change.


## Dependencies
This service has dependencies on all other Diablo2 file-loader services:
* [config file service](../configFile)

_______
This service exports an integration interface `IsFoo` with an alias 
`Dependencncy` which are intended to be used by other services for dependency
resolution (see servicemesh.HasDependencies), and expose just the methods which 
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