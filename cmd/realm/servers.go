package main

import (
	"crypto/tls"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"path/filepath"
	"time"

	"github.com/gravestench/dark-magic/internal/app/networktrust"
	"github.com/gravestench/dark-magic/internal/app/realm"
	"github.com/gravestench/dark-magic/internal/app/realmportal"
	portalassets "github.com/gravestench/dark-magic/internal/app/realmportal/assets"
)

// realmServers owns the public and optional private HTTP server lifecycles.
type realmServers struct {
	public         *http.Server
	publicErrors   <-chan error
	operator       *http.Server
	operatorErrors <-chan error
}

// startRealmServers creates the portal, TLS identity, and configured listeners.
func startRealmServers(
	control *realm.ControlPlane,
	assets *portalassets.Cache,
	directory string,
	config realmConfig,
) (realmServers, error) {
	if err := validateOperatorConfig(config); err != nil {
		return realmServers{}, err
	}

	handler, err := buildPublicHandler(control, assets)
	if err != nil {
		return realmServers{}, err
	}

	serverTLS, fingerprint, err := loadServerTLS(directory)
	if err != nil {
		return realmServers{}, err
	}

	servers, err := startPublicServer(handler, serverTLS, fingerprint, control, config.listenAddress)
	if err != nil {
		return realmServers{}, err
	}

	servers.operator, servers.operatorErrors, err = startOperatorServer(
		control,
		serverTLS,
		fingerprint,
		config,
	)
	if err != nil {
		_ = servers.public.Close()
		return realmServers{}, err
	}

	return servers, nil
}

// buildPublicHandler combines the authenticated API and browser portal.
func buildPublicHandler(
	control *realm.ControlPlane,
	assets *portalassets.Cache,
) (http.Handler, error) {
	apiHandler, err := realm.NewHTTPHandler(control)
	if err != nil {
		return nil, fmt.Errorf("build Realm API: %w", err)
	}

	handler, err := realmportal.NewHandler(control, apiHandler, assets)
	if err != nil {
		return nil, fmt.Errorf("build Realm portal: %w", err)
	}

	return handler, nil
}

// loadServerTLS loads or creates the Realm network identity.
func loadServerTLS(directory string) (*tls.Config, string, error) {
	trust, err := networktrust.New(filepath.Join(directory, "network"))
	if err != nil {
		return nil, "", fmt.Errorf("load Realm network identity: %w", err)
	}

	serverTLS, _, fingerprint, err := trust.HostTLS()
	if err != nil {
		return nil, "", fmt.Errorf("load Realm TLS identity: %w", err)
	}

	return serverTLS, fingerprint, nil
}

// startPublicServer begins serving the authenticated Realm API and portal.
func startPublicServer(
	handler http.Handler,
	serverTLS *tls.Config,
	fingerprint string,
	control *realm.ControlPlane,
	address string,
) (realmServers, error) {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return realmServers{}, fmt.Errorf("listen for Realm clients: %w", err)
	}

	server := &http.Server{
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       90 * time.Second,
	}
	errors := serveTLS(server, listener, serverTLS)
	slog.Info(
		"serving authenticated Realm API",
		"address", listener.Addr(),
		"tls_fingerprint", fingerprint,
		"version", control.Version(),
	)

	return realmServers{public: server, publicErrors: errors}, nil
}

// startOperatorServer begins the optional loopback-only operator API.
func startOperatorServer(
	control *realm.ControlPlane,
	serverTLS *tls.Config,
	fingerprint string,
	config realmConfig,
) (*http.Server, <-chan error, error) {
	if config.operatorListen == "" {
		return nil, nil, nil
	}

	if err := validateLoopbackAddress(config.operatorListen); err != nil {
		return nil, nil, err
	}

	token, err := realm.LoadOrCreateOperatorToken(config.operatorTokenFile)
	if err != nil {
		return nil, nil, fmt.Errorf("load Realm operator credential: %w", err)
	}

	handler, err := realm.NewOperatorHTTPHandler(control, token)
	if err != nil {
		return nil, nil, fmt.Errorf("build Realm operator API: %w", err)
	}

	listener, err := net.Listen("tcp", config.operatorListen)
	if err != nil {
		return nil, nil, fmt.Errorf("listen for Realm operators: %w", err)
	}

	server := &http.Server{
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       30 * time.Second,
	}
	errors := serveTLS(server, listener, serverTLS)
	slog.Info(
		"serving private Realm operator API",
		"address", listener.Addr(),
		"tls_fingerprint", fingerprint,
	)

	return server, errors, nil
}

// serveTLS starts one HTTP server and returns its buffered error channel.
func serveTLS(server *http.Server, listener net.Listener, config *tls.Config) <-chan error {
	errors := make(chan error, 1)
	go func() {
		errors <- server.Serve(tls.NewListener(listener, config))
	}()

	return errors
}

// validateOperatorConfig requires both private endpoint settings or neither.
func validateOperatorConfig(config realmConfig) error {
	if (config.operatorTokenFile == "") != (config.operatorListen == "") {
		return errors.New("operator-token-file and operator-listen must be set together")
	}

	return nil
}

// validateLoopbackAddress prevents the private operator API from binding publicly.
func validateLoopbackAddress(address string) error {
	host, port, err := net.SplitHostPort(address)

	ip := net.ParseIP(host)
	if err != nil || port == "" || ip == nil || !ip.IsLoopback() {
		return errors.New("operator-listen must use an explicit loopback IP and port")
	}

	return nil
}
