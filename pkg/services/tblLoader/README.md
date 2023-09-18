# TBL (Text-table) Loader Service
The purpose of this [runtime](https://github.com/gravestench/runtime) service is
to provide a single service that is responsible for loading text-table files, 
which are used to store character and language translation data for all of the
user-facing strings used in the game.

## Dependencies
This service has a single dependency on the MPQ loader service
* [mpq loader service](../mpqLoader)

## Integration with other services
This service exports an integration interface `LoadsTblFiles` with an alias
`Dependencncy` which are intended to be used by other services for dependency
resolution (see runtime.HasDependencies), and expose just the methods which
other services should use.
```golang
type Dependency = LoadsTblFiles

type LoadsTblFiles = interface {
    Load(filepath string) (tbl.TextTable, error)
}
```

Other services should use the `LoadsTblFiles` or `Dependency` interfaces to resolve
their dependency on this service.
