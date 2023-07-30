package web_server

import (
	"net/http"

	"github.com/rs/zerolog"

	"github.com/gravestench/runtime/examples/services/config_file"
	"github.com/gravestench/runtime/examples/services/web_router"
	"github.com/gravestench/runtime/pkg"
)

type Service struct {
	log        *zerolog.Logger
	router     web_router.Dependency
	cfgManager config_file.Dependency
	server     *http.Server
	lastConfig string
}

func (s *Service) Init(rt pkg.IsRuntime) {
	s.StartServer()
}

func (s *Service) BindLogger(l *zerolog.Logger) {
	s.log = l
}

func (s *Service) Logger() *zerolog.Logger {
	return s.log
}

func (s *Service) Name() string {
	return "Web Server"
}
