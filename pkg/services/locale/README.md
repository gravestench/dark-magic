# Diablo 2 Locale Service

The purpose of this [Service Mesh](https://github.com/gravestench/servicemesh) service is
to provide a single service that is responsible for retrieving locale-specific
data from the MPQ files. This includes string translation, character sets, etc.

## Dependencies

This service has two required dependencies:

* [mpq loader service](../mpqLoader)
* [tbl loader service](../tblLoader)

## Integration with other services

This service exports an integration interface `LoadsStringTables` with an alias
`Dependencncy` which are intended to be used by other services for dependency
resolution (see servicemesh.HasDependencies), and expose just the methods which
other services should use.

```golang
type Dependency = LoadsStringTables

type LoadsStringTables interface {
    GetSupportedLanguages() []string
    GetSupportedCharsets() []string
    GetTextByKey(key string) (string, error)
}
```

Other services should use the `LoadsStringTables` or `Dependency` interfaces to resolve
their dependency on this service.

## Web router service integration
If the [web router service](../webRouter) is present at servicemesh, this service will 
register routes for retrieving data.

The route slug for this service is `locale`, so all routes defined will be under
that route group.

| route                | method | purpose                                                      |
|----------------------|--------|--------------------------------------------------------------|
| `locale/`            | GET    | yields the entire composite string translation table as json |
| `locale/lookup/:key` | GET    | yields a single string tarnslation for the given key         |
