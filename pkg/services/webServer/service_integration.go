package webServer

import (
	"github.com/gravestench/runtime"

	"github.com/gravestench/dark-magic/pkg/services/configFile"
)

var (
	_ runtime.Service             = &Service{}
	_ runtime.HasLogger           = &Service{}
	_ configFile.HasDefaultConfig = &Service{}
	_ IsWebServer                 = &Service{}
)

type Dependency = IsWebServer

type IsWebServer interface {
	RestartServer()
	StartServer()
	StopServer()
}
