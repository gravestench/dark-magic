package goscript

import (
	"github.com/gravestench/servicemesh"
)

// these are static declarations that force a
// compile-time error if the service does not
// implement them.
var (
	_ servicemesh.Service   = &Service{} // implement in`service.go`
	_ servicemesh.HasLogger = &Service{} // implement in`service.go`
)
