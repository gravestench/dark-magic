package webServer

import (
	"github.com/gravestench/servicemesh"

	"github.com/gravestench/dark-magic/pkg/services/configFile"
)

var (
	_ servicemesh.Service         = &Service{}
	_ servicemesh.HasLogger       = &Service{}
	_ configFile.HasDefaultConfig = &Service{}
	_ IsWebServer                 = &Service{}
)

type Dependency = IsWebServer

type IsWebServer interface {
	RestartServer()
	StartServer()
	StopServer()
}
