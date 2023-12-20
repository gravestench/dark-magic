package webServer

import (
	"net/http"

	"github.com/foomo/tlsconfig"
)

func (s *Service) initTlsServer() {
	tlsconf := tlsconfig.NewServerTLSConfig(tlsconfig.TLSModeServerStrict)

	g := s.config.Group("Web Server")

	// init server
	s.server = &http.Server{
		Addr:      ":" + g.GetString(keyPort),
		TLSConfig: tlsconf,
		Handler:   s.router.RouteRoot(),
	}

	// we throw away this error because it may just be that the
	// server is restarting for some normal reason, not that
	// anything crashed
	if err := s.server.ListenAndServeTLS(g.GetString(keyCertFilepath), g.GetString(keyKeyFilepath)); err != nil {
		s.log.Warn("TLS server not running", "error", err)
	}
}
