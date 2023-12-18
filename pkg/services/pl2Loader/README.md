# PL2 Loader Service
The purpose of this [servicemesh](https://github.com/gravestench/servicemesh) service is
to provide a single service that is responsible for loading PL2 palette transformation files.

The palette transformations are pre-computed indexes into the standard diablo2 
palettes which are used as a pseudo gamma/alpha/hue shifts for the DC6, DCC, DT1,
and DS1 graphics.

## Dependencies
This service has a single dependency on the MPQ loader service
* [mpq loader service](../mpqLoader)

## Integration with other services
This service exports an integration interface `LoadsPl2Files` with an alias
`Dependencncy` which are intended to be used by other services for dependency
resolution (see servicemesh.HasDependencies), and expose just the methods which
other services should use.
```golang
type Dependency = LoadsPl2Files

type LoadsPl2Files = interface {
    Load(filepath string) (*pl2.PL2, error)
}
```

Other services should use the `LoadsPl2Files` or `Dependency` interfaces to resolve
their dependency on this service.
