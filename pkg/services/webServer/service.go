package webServer

import (
	"net/http"

	"github.com/rs/zerolog"

	"github.com/gravestench/runtime/pkg"

	"github.com/gravestench/dark-magic/pkg/services/configFile"
	"github.com/gravestench/dark-magic/pkg/services/webRouter"
)

type Service struct {
	log        *zerolog.Logger
	router     webRouter.Dependency
	cfgManager configFile.Dependency
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
