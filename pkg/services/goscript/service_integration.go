package goscript

import (
	"github.com/gravestench/runtime"
)

// these are static declarations that force a
// compile-time error if the service does not
// implement them.
var (
	_ runtime.Service   = &Service{} // implement in`service.go`
	_ runtime.HasLogger = &Service{} // implement in`service.go`
)
