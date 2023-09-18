# DT1 Loader Service
The purpose of this [runtime](https://github.com/gravestench/runtime) service is
to provide a single service that is responsible for loading DT1 map tileset files.

## Dependencies
This service has a single dependency on the MPQ loader service
* [mpq loader service](../mpqLoader)

## Integration with other services
This service exports an integration interface `LoadsDt1Files` with an alias
`Dependencncy` which are intended to be used by other services for dependency
resolution (see runtime.HasDependencies), and expose just the methods which
other services should use.
```golang
type Dependency = LoadsDt1Files

type LoadsDc6Files = interface {
    Load(filepath string) (*dt1.DT1, error)
}
```

Other services should use the `LoadsDt1Files` or `Dependency` interfaces to resolve
their dependency on this service.
