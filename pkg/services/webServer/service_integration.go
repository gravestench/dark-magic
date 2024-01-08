package webServer

import (
	"github.com/gravestench/servicemesh"

	"github.com/gravestench/dark-magic/pkg/services/configManager"
)

var (
	_ servicemesh.Service            = &Service{}
	_ servicemesh.HasLogger          = &Service{}
	_ configManager.HasConfiguration = &Service{}
	_ IsWebServer                    = &Service{}
)

type Dependency = IsWebServer

type IsWebServer interface {
	RestartServer()
	StartServer()
	StopServer()
}
