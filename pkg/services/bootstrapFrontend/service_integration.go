package bootstrap_backend

import (
	"github.com/gravestench/servicemesh"
)

var (
	_ servicemesh.Service   = &Service{} // implement in`service.go`
	_ servicemesh.HasLogger = &Service{} // implement in`service.go`
)
