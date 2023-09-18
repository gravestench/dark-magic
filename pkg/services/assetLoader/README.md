# Unified Asset Loader
The purpose of this [runtime](https://github.com/gravestench/runtime) service is to provide a single dependency
which has all file loading methods available for use by other high-level
D2 runtime services.

## Dependencies
This service has dependencies on all other Diablo2 file-loader services:
* [dc6 loader service](../dc6Loader)
* [dcc loader service](../dccLoader)
* [ds1 loader service](../ds1Loader)
* [dt1 loader service](../dt1Loader)
* [fontTable loader service](../fontTableLoader)
* [gpl loader service](../gplLoader)
* [mpq loader service](../mpqLoader)
* [pl2 loader service](../pl2Loader)
* [tbl loader service](../tblLoader)
* [tsv loader service](../tsvLoader)
* [wav loader service](../wavLoader)



## Integration with other services
This service integrates with the following services:
* [lua service](../lua)
* [web router service](../web_router)

The integration is optional; if neither are added to the runtime then the 
integration methods will never be called.

This service exports an integration interface `LoadsDiabloFiles` with an alias 
`Dependencncy` which are intended to be used by other services for dependency
resolution (see runtime.HasDependencies), and expose just the methods which 
other services should use.
```golang
type LoadsDiabloFiles interface {
    Load(filepath string) (io.Reader, error)
    LoadDc6(filepath string) (*dc6.DC6, error)
    LoadDcc(filepath string) (*dcc.DCC, error)
    LoadDs1(filepath string) (*ds1.DS1, error)
    LoadDt1(filepath string) (*dt1.DT1, error)
    LoadGpl(filepath string) (*gpl.GPL, error)
    LoadPl2(filepath string) (*pl2.PL2, error)
    LoadTbl(filepath string) (tbl.TextTable, error)
    LoadTsv(filepath string, destination any) error
    LoadWav(filepath string) ([]byte, error)
}
```

## Lua service integration
A global `assets` variable is exported to lua. At the time of writing, there is 
a single `load` function which loads a file from the mpq loader and yields the 
file data as an array of bytes. 

Here is an example of its usage:
```lua
data = assets.load("/data/global/ui/Loading/loadingscreen.dc6")
```

## Web router service integration
If the [web router service](../web_router) is present at runtime, this service will
register routes for retrieving data.

The route slug for this service is `asset`, so all routes defined will be under 
that route group.

| route         | method | purpose                                             |
|---------------|--------|-----------------------------------------------------|
| `asset/*path`  | GET    | yields the file for the given `path` from the mpq's |
