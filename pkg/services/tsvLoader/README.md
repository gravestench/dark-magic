# TSV File Loader Service
The purpose of this [Service Mesh](https://github.com/gravestench/servicemesh) service is
to provide a single service that is responsible for loading TSV (tab-separated 
value) files, which are identical to CSV files save for the delimiter character.

All of the diablo 2 game data from the MPQ's (like monster stats, map-gen info, 
etc.) are stored in these spreadsheets. The record manager uses this service to
marshall all of the spreadsheets into array of their corresponding models from 
`pkg/models/`.

## Dependencies
This service has a single dependency on the MPQ loader service
* [mpq loader service](../mpqLoader)

## Integration with other services
This service exports an integration interface `LoadsTsvFiles` with an alias
`Dependencncy` which are intended to be used by other services for dependency
resolution (see servicemesh.HasDependencies), and expose just the methods which
other services should use.
```golang
type Dependency = LoadsTsvFiles

type LoadsTsvFiles = interface {
    Load(filepath string, destination any) error
}

```

Other services should use the `LoadsTsvFiles` or `Dependency` interfaces to resolve
their dependency on this service.
