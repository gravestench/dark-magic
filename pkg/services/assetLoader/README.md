# Unified Asset Loader

The purpose of this [runtime](https://github.com/gravestench/runtime) service is to provide a single dependency
which can has all file loading methods available for use by other high-level
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



## Integrations

This service integrates with the following services:
* [lua service](../lua) - exports a global `assets` table with loader functions for all of the loaders
* [web router service](../web_router) - defines web routes for yielding files from the mpqs

This integration is optional, if neither are added to the runtime then the 
integration methods will never be called.

## Lua service integration
A global `assets` variable is exported to lua. At the time of writing, there is 
a single `load` function which loads a file from the mpq loader and yields the 
file data as an array of bytes. 

Here is an example of its usage:
```lua
data = assets.load("/data/global/ui/Loading/loadingscreen.dc6")
```

## Web router service integration

The route slug for this service is `asset`, so all routes defined will be under 
that route group.

| route         | method | purpose                                             |
|---------------|--------|-----------------------------------------------------|
| `asset/*path`  | GET    | yields the file for the given `path` from the mpq's |
