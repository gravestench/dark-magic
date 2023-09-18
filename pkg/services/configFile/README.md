# Config File Manager
The purpose of this [runtime](https://github.com/gravestench/runtime) service is 
to provide a unified method of managing configuration files in a common root
directory, and to provide other services a way to integrate with it to manage 
their own config files.

## Events
This service uses the global runtime event bus. There is only a single event
`EventConfigChanged`, which is emitted when a config file is changed. The 
arguments that are passed with this event are the config group, the key, and 
the value. 

Other services can listen for this event to respond to their own 
config file being changed internally, or even externally if a file is edited
outside of the application.

Here is an example of binding to this event during the Init of a service:
```golang
func (s *Service) Init(rt runtime.R) {
    rt.Events().On(config_file.EventConfigChanged, func(...any){
	    // handler logic here	
    })
}
```

## Integration with other services

This service exports an integration interface `Manager` with an alias
`Dependencncy` which are intended to be used by other services for dependency
resolution (see runtime.HasDependencies), and expose just the methods which
other services should use.
```golang
// Manager represents something that manages configurations.
type Manager interface {
    GetPath(path string) string
    ConfigDirectory() string
    SetConfigDirectory(string) error
    Configs() map[string]*Config
    GetConfig(string) (*Config, error)
    CreateConfig(path string) (*Config, error)
    LoadConfig(string) (*Config, error)
    SaveConfig(string) error
}
```

Other services should use the `Manager` or `Dependency` interfaces to resolve
their dependency on this service.

_________________

This runtime service operates primarily by looking for other services which 
implement the following interfaces: 
```golang
// HasConfig represents a something with a configuration file path and retrieval methods.
type HasConfig interface {
    ConfigFileName() string   // ConfigFilePath returns the path to the configuration file.
}
```

```golang
// HasDefaultConfig represents something with a default configuration.
type HasDefaultConfig interface {
    HasConfig
    DefaultConfig() Config // DefaultConfig returns the default configuration.
}
````

If another service implements `HasConfig` it will be used to create an empty 
config (if it doesnt exist). The `Config` method should use a reference to the
`Manager` interface (obtained through implementing runtime.HasDependencies) to 
yield it's own config file.

Additionally, another service can implement `HasDefaultConfig` in order to 
declare what the default values of its own config file should be.
