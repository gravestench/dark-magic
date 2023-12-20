# COF Loader Service
The purpose of this [Service Mesh](https://github.com/gravestench/servicemesh) service is
to provide a single service that is responsible for loading COF animation files.
COF likely stands for "Composite Object File", which is coherent considering that
all of the character animations are composed of many separate dcc images.

## Dependencies
This service has a single dependency on the MPQ loader service
* [mpq loader service](../mpqLoader)

## Integration with other services
This service exports an integration interface `LoadsCofFiles` with an alias
`Dependencncy` which are intended to be used by other services for dependency
resolution (see servicemesh.HasDependencies), and expose just the methods which
other services should use.
```golang
type Dependency = LoadsCofFiles

type LoadsCofFiles = interface {
    Load(filepath string) (*cof.COF, error)
    LoadAnimationData(filepath string) (*cof.AnimationData, error)
}
```

Other services should use the `LoadsCofFiles` or `Dependency` interfaces to resolve
their dependency on this service.
