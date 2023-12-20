package fileWatcher

import (
	"github.com/gravestench/servicemesh"
)

// these are static declarations that force a
// compile-time error if the service does not
// implement them.
var (
	_ servicemesh.Service   = &Service{} // implement in`service.go`
	_ servicemesh.HasLogger = &Service{} // implement in`service.go`
	_ IsFileWatcher         = &Service{} // implement in`service.go`
)

// this is an alias which can be used to make
// the dependency resolution methods of other
// services more coherent. It's just sugar.

type Dependency = IsFileWatcher

// Here is the declaration of our service as
// an interface. This is all the dependent services
// should know about this service.

type IsFileWatcher interface {
	AddWatcher(path string, f func(path string) error)
	WatchAndLoad(path string, f func(path string) error)
	CloseWatcher()
}

// TODO :: other services should be able to un-watch stuff, too...

// NeedsFileWatcher is an integration interface intended to be implemented by
// other services to integrate with and be maintained by this file watcher
// service.
type NeedsFileWatcher interface {
	FileHandlers() map[string]FileHandlerFunc
}

type FileHandlerFunc = func(path string) error
