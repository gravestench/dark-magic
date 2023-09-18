# Map Generator Service

The purpose of this [runtime](https://github.com/gravestench/runtime) service is
to provide a way of generating "map" objects for diablo2, which are

## Dependencies

This service depends upon the [config file service](../configFile) and will
initialize its own default config.

## Integration with other services

This service exports an integration interface `GeneratesDiablo2Maps` with an alias
`Dependencncy` which are intended to be used by other services for dependency
resolution (see runtime.HasDependencies), and expose just the methods which
other services should use.

```golang
type Dependency = GeneratesDiablo2Maps

type GeneratesDiablo2Maps interface {
    Seed() uint64
    SetSeed(uint64)
    GenerateMap(act, difficulty uint) (models.Diablo2Map, error)
}
```

Other services should use the `GeneratesDiablo2Maps` or `Dependency` interfaces to resolve
their dependency on this service.