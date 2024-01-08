package webServer

import (
	"fmt"
	"net/http"

	"github.com/foomo/tlsconfig"
)

func (s *Service) initTlsServer() {
	tlsconf := tlsconfig.NewServerTLSConfig(tlsconfig.TLSModeServerStrict)

	// init server
	s.server = &http.Server{
		Addr:      fmt.Sprintf(":%d", s.config.Port),
		TLSConfig: tlsconf,
		Handler:   s.router.RouteRoot(),
	}

	// we throw away this error because it may just be that the
	// server is restarting for some normal reason, not that
	// anything crashed
	if err := s.server.ListenAndServeTLS(s.config.CertFilepath, s.config.KeyFilepath); err != nil {
		s.log.Warn("TLS server not running", "error", err)
	}
}
