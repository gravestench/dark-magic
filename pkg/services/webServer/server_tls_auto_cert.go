package webServer

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/foomo/simplecert"
	"github.com/foomo/tlsconfig"
)

func (s *Service) initAutocertTlsDebugServer() {
	g := s.config.Group("Web Server")

	certConfig := simplecert.Default
	certConfig.UpdateHosts = false
	certConfig.Domains = []string{"foobar.com"}
	certConfig.CacheDir = "simplecert"
	certConfig.Local = true
	certReloader, err := simplecert.Init(certConfig, func() {})
	if err != nil {
		s.log.Error("simplecert init failed", "error", err)
		panic(err)
	}

	//// redirect HTTP to HTTPS
	//s.log.Info("starting HTTP Listener on Port 80")
	//go http.ListenAndServe(":80", http.HandlerFunc(simplecert.Redirect))

	// init strict tlsConfig with certReloader
	tlsconf := tlsconfig.NewServerTLSConfig(tlsconfig.TLSModeServerStrict)

	// now set GetCertificate to the reloaders GetCertificateFunc to enable hot reload
	tlsconf.GetCertificate = certReloader.GetCertificateFunc()

	// init server
	s.server = &http.Server{
		Addr:      fmt.Sprintf(":%d", g.GetInt(keyPort)),
		TLSConfig: tlsconf,
		Handler:   s.router.RouteRoot(),
	}

	s.log.Info("serving: https://" + certConfig.Domains[0])

	_ = s.server.ListenAndServeTLS("", "")
}

func (s *Service) initAutocertTlsProductionServer() {
	g := s.config.Group("Web Server")

	var (
		// the structure that handles reloading the certificate
		certReloader *simplecert.CertReloader
		numRenews    int
		ctx, cancel  = context.WithCancel(context.Background())

		// init strict tlsConfig (this will enforce the use of modern TLS configurations)
		// you could use a less strict configuration if you have a customer facing web application that has visitors with old browsers
		tlsConf = tlsconfig.NewServerTLSConfig(tlsconfig.TLSModeServerStrict)

		// a simple constructor for a http.Server with our Handler
		makeServer = func() *http.Server {
			return &http.Server{
				Addr:      fmt.Sprintf(":%v", g.GetString(keyPort)),
				Handler:   s.router.RouteRoot(),
				TLSConfig: tlsConf,
			}
		}

		// init server
		srv = makeServer()

		// init simplecert configuration
		certConfig = simplecert.Default

		serve = func(ctx context.Context, srv *http.Server) {

			// lets go
			go func() {
				if err := srv.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
					s.log.Error("listen: %+s\n", err)
					panic(err)
				}
			}()

			s.log.Info("server started")
			<-ctx.Done()
			s.log.Info("server stopped")

			ctxShutDown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer func() {
				cancel()
			}()

			err := srv.Shutdown(ctxShutDown)
			if err == http.ErrServerClosed {
				s.log.Info("server exited properly")
			} else if err != nil {
				s.log.Info("server encountered an error on exit", "error", err)
			}
		}
	)

	// configure
	certConfig.Domains = []string{"foobar.com", "www.foobar.com"}
	certConfig.CacheDir = "letsencrypt"
	certConfig.SSLEmail = "bruh@foobar.com"

	// disable HTTP challenges - we will only use the TLS challenge for this example.
	certConfig.HTTPAddress = ""

	// this function will be called just before certificate renewal starts and is used to gracefully stop the service
	// (we need to temporarily free port 443 in order to complete the TLS challenge)
	certConfig.WillRenewCertificate = func() {
		// stop server
		cancel()
	}

	// this function will be called after the certificate has been renewed, and is used to restart your service.
	certConfig.DidRenewCertificate = func() {

		numRenews++

		// restart server: both context and server instance need to be recreated!
		ctx, cancel = context.WithCancel(context.Background())
		srv = makeServer()

		// force reload the updated cert from disk
		certReloader.ReloadNow()

		// here we go again
		go serve(ctx, srv)
	}

	// init simplecert configuration
	// this will block initially until the certificate has been obtained for the first time.
	// on subsequent runs, simplecert will load the certificate from the cache directory on disk.
	certReloader, err := simplecert.Init(certConfig, func() {})
	if err != nil {
		s.log.Error("simplecert init failed", "error", err)
		panic(err)
	}

	//// redirect HTTP to HTTPS
	//s.Msgf("starting HTTP Listener on Port 80")
	//go http.ListenAndServe(":80", http.HandlerFunc(simplecert.Redirect))

	// enable hot reload
	tlsConf.GetCertificate = certReloader.GetCertificateFunc()

	// start serving
	s.log.Info(fmt.Sprintf("will serve at: https://%s", certConfig.Domains[0]))
	serve(ctx, srv)
}
