# Hero Manager Service
The purpose of this [servicemesh](https://github.com/gravestench/servicemesh) service is
to provide a means of managing heroes (player characters) for the game. This
service uses the record manager to load up various records for character
starting attributes items, graphois, etc.


## Dependencies
This service has dependencies on the following services:
* [config file manager](../configFile)
* [record manager](../recordManager)


## Integration with other services
This service integrates with the following services:
* [buzz](.)
* [quzz](.)

* describe where the integration is optonal or not ...

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