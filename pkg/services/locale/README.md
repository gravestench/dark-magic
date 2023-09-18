# GPL Palette Loader Service
The purpose of this [runtime](https://github.com/gravestench/runtime) service is
to provide a single service that is responsible for loading GPL palette files.

## Dependencies
This service has a single dependency on the MPQ loader service
* [mpq loader service](../mpqLoader)

## Integration with other services
This service exports an integration interface `LoadsGplFiles` with an alias
`Dependencncy` which are intended to be used by other services for dependency
resolution (see runtime.HasDependencies), and expose just the methods which
other services should use.
```golang
type Dependency = LoadsGplFiles

type LoadsGplFiles = interface {
    Load(filepath string) (*gpl.GPL, error)
}
```

Other services should use the `LoadsGplFiles` or `Dependency` interfaces to resolve
their dependency on this service.