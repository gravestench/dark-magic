package webServer

import (
	"fmt"
	"net/http"
)

func (s *Service) initHttpServer() {
	g := s.config.Group("Web Server")

	s.server = &http.Server{
		Addr:    fmt.Sprintf(":%v", g.GetInt(keyPort)),
		Handler: s.router.RouteRoot(),
	}

	if err := s.server.ListenAndServe(); err != nil {
		s.log.Warn("http server not running", "error", err)
	}
}
