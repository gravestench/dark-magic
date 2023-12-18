package modalTui

import (
	"github.com/charmbracelet/bubbletea"

	"github.com/gravestench/servicemesh"
)

// these are static declarations that force a
// compile-time error if the service does not
// implement them.
var (
	_ servicemesh.Service             = &Service{} // implement in`service.go`
	_ servicemesh.HasLogger           = &Service{} // implement in`service.go`
	_ servicemesh.HasDependencies     = &Service{} // implement in`service.go`
	_ servicemesh.HasGracefulShutdown = &Service{} // implement in`service.go`
	_ ManagesModalTextUserInterface   = &Service{} // implement in`service.go`
)

// this is an alias which can be used to make
// the dependency resolution methods of other
// services more coherent. It's just sugar.

type Dependency = ManagesModalTextUserInterface

type ManagesModalTextUserInterface interface {
	CurrentModalName() string
	GetModalNames() []string
	SwitchModal(name string) error
}

// HasModalTextUserInterface is an integration interface that another
// service can implement to be picked up and used by the Modal TUI service
type HasModalTextUserInterface interface {
	ModalTui() (name string, model tea.Model)
}
