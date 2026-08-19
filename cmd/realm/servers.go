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

// realmServers keeps listener handles and terminal errors together so the runtime
// can coordinate first failure and close every surface it successfully opened.
type realmServers struct {
	public         *http.Server
	publicErrors   <-chan error
	operator       *http.Server
	operatorErrors <-chan error
}

// startRealmServers validates the private surface before opening public ingress,
// then shares one Realm TLS identity across both listeners. Partial startup is
// rolled back so callers never inherit an unmanaged public server.
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

// buildPublicHandler mounts browser account flows around the authenticated API
// without giving the portal a second control-plane instance or repository view.
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

// loadServerTLS loads the persistent Realm network identity whose fingerprint is
// pinned by clients and workers; regenerating it casually would break established trust.
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

// startPublicServer opens the player-facing ingress only after handler and TLS
// construction succeed, then reports the fingerprint operators must distribute.
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

// startOperatorServer creates a mutation-capable API only on an explicit loopback
// address and with a durable bearer token. Public binding is rejected before listen.
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

// serveTLS wraps an already-bound listener and captures its sole terminal error.
// Buffering prevents the serving goroutine from leaking while shutdown selects elsewhere.
func serveTLS(server *http.Server, listener net.Listener, config *tls.Config) <-chan error {
	errors := make(chan error, 1)
	go func() {
		errors <- server.Serve(tls.NewListener(listener, config))
	}()

	return errors
}

// validateOperatorConfig makes the private API all-or-nothing: a token without a
// listener is misleading, while a listener without a token would be unsafe.
func validateOperatorConfig(config realmConfig) error {
	if (config.operatorTokenFile == "") != (config.operatorListen == "") {
		return errors.New("operator-token-file and operator-listen must be set together")
	}

	return nil
}

// validateLoopbackAddress requires a literal loopback IP rather than a hostname.
// This avoids DNS or hosts-file changes silently widening the operator trust boundary.
func validateLoopbackAddress(address string) error {
	host, port, err := net.SplitHostPort(address)

	ip := net.ParseIP(host)
	if err != nil || port == "" || ip == nil || !ip.IsLoopback() {
		return errors.New("operator-listen must use an explicit loopback IP and port")
	}

	return nil
}
