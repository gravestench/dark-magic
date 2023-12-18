# Font Table Loader Service
The purpose of this [servicemesh](https://github.com/gravestench/servicemesh) service is
to provide a single service that is responsible for loading font table files.

## Dependencies
This service has a single dependency on the MPQ loader service
* [mpq loader service](../mpqLoader)

## Integration with other services
This service exports an integration interface `LoadsFontTableFiles` with an alias
`Dependencncy` which are intended to be used by other services for dependency
resolution (see servicemesh.HasDependencies), and expose just the methods which
other services should use.
```golang
type Dependency = LoadsFontTableFiles

type LoadsFontTableFiles = interface {
    Load(filepath string) (*font_table.Font, error)
}

```

Other services should use the `LoadsFontTableFiles` or `Dependency` interfaces to resolve
their dependency on this service.
