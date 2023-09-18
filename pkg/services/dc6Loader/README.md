# DC6 Loader Service
The purpose of this [runtime](https://github.com/gravestench/runtime) service is
to provide a single service that is responsible for loading DC6 image files.

## Dependencies
This service has a single dependency on the MPQ loader service
* [mpq loader service](../mpqLoader)

## Integration with other services
This service exports an integration interface `LoadsDc6Files` with an alias
`Dependencncy` which are intended to be used by other services for dependency
resolution (see runtime.HasDependencies), and expose just the methods which
other services should use.
```golang
type Dependency = LoadsDc6Files

type LoadsDc6Files = interface {
    Load(filepath string) (*dc6.DC6, error)
}
```

Other services should use the `LoadsDc6Files` or `Dependency` interfaces to resolve
their dependency on this service.
