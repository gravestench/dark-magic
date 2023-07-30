package web_server

import (
	"time"
)

func (s *Service) StartServer() {
	cfg, err := s.Config()
	if err != nil {
		s.log.Fatal().Msgf("getting config: %v", err)
	}

	g := cfg.Group("Web Server")

	autocertEnabled := g.GetBool(keyAutocert)
	tlsEnabled := g.GetBool(keyTls)
	port := g.GetInt(keyPort)

	if port <= 1 {
		port = defaultPort
	}

	if tlsEnabled && autocertEnabled {
		s.log.Info().Msgf("starting autocert HTTPS server, listening on %d", port)
		go s.initAutocertTlsDebugServer()
		return
	} else if tlsEnabled && !autocertEnabled {
		s.log.Info().Msgf("starting HTTPS server, listening on %d", port)
		go s.initTlsServer()
		return
	}

	s.log.Info().Msgf("starting HTTP server, listening on %d", port)
	go s.initHttpServer()
}

func (s *Service) StopServer() {
	if s.server == nil {
		return
	}

	s.log.Warn().Msg("stopping server")

	_ = s.server.Close()
	s.server = nil
}

func (s *Service) RestartServer() {
	s.log.Warn().Msg("restarting server")
	time.Sleep(time.Second)
	s.StopServer()
	s.StartServer()
}
