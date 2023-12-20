package webServer

import (
	"time"
)

func (s *Service) StartServer() {
	g := s.config.Group("Web Server")

	autocertEnabled := g.GetBool(keyAutocert)
	tlsEnabled := g.GetBool(keyTls)
	port := g.GetInt(keyPort)

	if port <= 1 {
		port = defaultPort
	}

	if tlsEnabled && autocertEnabled {
		s.log.Info("starting autocert HTTPS server", "port", port)
		go s.initAutocertTlsDebugServer()
		return
	} else if tlsEnabled && !autocertEnabled {
		s.log.Info("starting HTTPS server", "port", port)
		go s.initTlsServer()
		return
	}

	s.log.Info("starting HTTP server", "port", port)
	go s.initHttpServer()
}

func (s *Service) StopServer() {
	if s.server == nil {
		return
	}

	s.log.Warn("stopping server")

	_ = s.server.Close()
	s.server = nil
}

func (s *Service) RestartServer() {
	s.log.Warn("restarting server")
	time.Sleep(time.Second)
	s.StopServer()
	s.StartServer()
}
