package web_server

import (
	"github.com/gravestench/runtime"
	"github.com/gravestench/runtime/examples/services/config_file"
)

var (
	_ runtime.Service              = &Service{}
	_ runtime.HasLogger            = &Service{}
	_ config_file.HasDefaultConfig = &Service{}
	_ IsWebServer                  = &Service{}
)

type Dependency = IsWebServer

type IsWebServer interface {
	RestartServer()
	StartServer()
	StopServer()
}
