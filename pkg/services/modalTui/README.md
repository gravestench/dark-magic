# Modal Text User Interface
The purpose of this [servicemesh](https://github.com/gravestench/servicemesh) service is to provide a modal text 
user interface. This means that a text-based user interface will be
shown on standard output. The normal servicemesh logging output will be 
redirected to a file, while `stdout` will be used to display the TUI.

This service will use other services that implement an integration interface
(defined below), which has a "modal" ui. A "modal" ui is a UI that has a single
current mode, and one or more possible modes to switch between. Only a single
modal will be shown at a time.

This service implements the TUI using the [bubbletea](https://github.com/charmbracelet/bubbletea) library.

## Dependencies
This service depends upon the [config file manager](../configFile) service.


## Integration with other services
This service exports an integration interface `ManagesModalTextUserInterface` 
with an alias `Dependencncy` which are intended to be used by other services 
for dependency resolution (see servicemesh.HasDependencies), and expose just the 
methods which other services should use.
```golang
type Dependency = ManagesModalTextUserInterface

type ManagesModalTextUserInterface interface {
    GetModalNames() []string
    SwitchModal(name string) error
}
```

There is also an integration `HasModalTextUserInterface` interface that other 
services can implement in order to be picked up and used as a modal TUI.
```golang
type HasModalTextUserInterface interface {
    ModalTui() (name string, model tea.Model)
}
```