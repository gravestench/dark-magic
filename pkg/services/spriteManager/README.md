# Sprite Manager Service

The purpose of this [runtime](https://github.com/gravestench/runtime) service is to use the DC6, DCC,
DS1, and DT1 files to generate graphical sprites in PNG and GIF format.

## Dependencies

This service has dependencies on the following services:

* [config file manager](../configFile)
* [dc6 loader](../dc6Loader)
* [dcc loader](../dccLoader)
* [ds1 loader](../ds1Loader)
* [dt1 loader](../dt1Loader)
* [pl2 loader](../pl2Loader)
* [mpq loader](../mpqLoader)

## Integration with other services

This service integrates with the following services:

* [web router](../webRouter)
* [cache manager](../cacheManager)

The integrations are optional, meaning that the web router and cache
manager can be safely omitted from the runtime.

_______
This service exports an integration interface `IsSpriteLoader` with an alias
`Dependencncy` which are intended to be used by other services for dependency
resolution (see runtime.HasDependencies), and expose just the methods which
other services should use.

```golang
type Dependency = IsSpriteLoader

type IsSpriteLoader interface {
    LoadDc6ToPngSpriteAtlas(pathDC6 string, pathPL2 string) ([]byte, error)
    LoadDc6ToAnimatedGif(pathDC6 string, pathPL2 string) ([]byte, error)
    
    LoadDccToPngSpriteAtlas(pathDC6 string, pathPL2 string) ([]byte, error)
    LoadDccToAnimatedGif(pathDC6 string, pathPL2 string) ([]byte, error)
    
    LoadDt1ToPngSpriteAtlas(pathDC6 string, pathPL2 string) ([]byte, error)
    
    LoadDs1ToPngSpriteAtlas(pathDC6 string, pathPL2 string) ([]byte, error)
}
```

## Web router service integration

If the [web router service](../webRouter) is present at runtime, this service will
register routes for flushing all caches and for getting cache statistics.

The route slug for this service is `sprite`, so all routes defined will be under
that route group.

| route                       | method | purpose                                                                                                                                                    |
|-----------------------------|--------|------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `sprite/png/:palette/*path` | GET    | generates and serves a png image (with packed JSON sprite atlas information after the PNG data) using the specified filepath (supports dcc, dc6, dt1, ds1) |
| `sprite/gif/:palette/*path` | GET    | generates and serves and animated gif from the specified filepath (supports dcc, dc6)                                                                      |
