package web_server

import (
	"fmt"
	"net/http"
)

func (s *Service) initHttpServer() {
	cfg, err := s.Config()
	if err != nil {
		s.log.Fatal().Msgf("getting config: %v", err)
	}

	g := cfg.Group("Web Server")

	s.server = &http.Server{
		Addr:    fmt.Sprintf(":%v", g.GetInt(keyPort)),
		Handler: s.router.RouteRoot(),
	}

	if err := s.server.ListenAndServe(); err != nil {
		s.log.Warn().Msgf("http server not running: %v", err)
	}
}
