# MPQ Archive Loader & Decompression Service

The purpose of this [runtime](https://github.com/gravestench/runtime) service is
to provide a single service which loads and is able to decompress
files from the collection of MPQ archive files used by diablo.

## Dependencies

This service depends upon the [config file service](../configFile) and will
initialize its own default config.

There is an external dependency of this service, namely the presence of the
Diablo 2 MPQ files on your local filesystem. The MPQ files are:
* `patch_d2.mpq`
* `d2exp.mpq`
* `d2xmusic.mpq`
* `d2xtalk.mpq`
* `d2xvideo.mpq`
* `d2data.mpq`
* `d2char.mpq`
* `d2music.mpq`
* `d2sfx.mpq`
* `d2video.mpq`
* `d2speech.mpq`

These archives are searched in order, and the general idea is:
* patch archives are searched first
* the LOD expansion archives are searched second
* the vanilla d2 archives are searched last

For modding, obviously you would want to place your own patch MPQ file
at the top of the list so that it is checked for the presence of files first.

## Integration with other services

This service exports an integration interface `ReadsMpqArchives` with an alias
`Dependencncy` which are intended to be used by other services for dependency
resolution (see runtime.HasDependencies), and expose just the methods which
other services should use.

```golang
type Dependency = ReadsMpqArchives

type ReadsMpqArchives interface {
    RequiredArchivesLoaded() bool
    Archives() map[string]*mpq.MPQ
    AddArchive(filepath string) error
    RemoveArchive(filepath string) error
    Load(filepath string) (io.Reader, error)
}
```

Other services should use the `ReadsMpqArchives` or `Dependency` interfaces to resolve
their dependency on this service. 

The `RequiredArchivesLoaded` method is intended to be used to delay other 
dependent services from initializing in their respective dependency 
resolution `DependenciesResolved` method .