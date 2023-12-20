package webServer

import (
	"log/slog"
	"net/http"

	"github.com/gravestench/servicemesh"

	"github.com/gravestench/dark-magic/pkg/services/configFile"
	"github.com/gravestench/dark-magic/pkg/services/webRouter"
)

type Service struct {
	log        *slog.Logger
	router     webRouter.Dependency
	config     *configFile.Config
	server     *http.Server
	lastConfig string
}

func (s *Service) Init(mesh servicemesh.Mesh) {
	s.StartServer()
}

func (s *Service) SetLogger(l *slog.Logger) {
	s.log = l
}

func (s *Service) Logger() *slog.Logger {
	return s.log
}

func (s *Service) Name() string {
	return "Web Server"
}

func (s *Service) Ready() bool {
	if s.config == nil {
		return false
	}

	if s.router == nil {
		return false
	}

	return true
}
