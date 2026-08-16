package serverapp

import (
	"context"
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/gravestench/dark-magic/internal/app/gameserver"
	"github.com/gravestench/dark-magic/internal/app/realm"
	playeradapter "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/player"
)

type WorkerControlConfig struct {
	Address, CertificatePath, PrivateKeyPath, TokenPath string
	Tickets                                             *gameserver.TicketAuthority
	Destination                                         playeradapter.Destination
	Memberships                                         *realm.WorkerMemberships
	Drain                                               func()
}

// WorkerControlServer owns the private HTTPS listener used by a Realm process
// allocator. It is not a public client or administration endpoint.
type WorkerControlServer struct {
	listener    net.Listener
	server      *http.Server
	fingerprint string
}

func StartWorkerControl(config WorkerControlConfig, host *gameserver.Host) (*WorkerControlServer, error) {
	configured := config.Address != "" || config.CertificatePath != "" || config.PrivateKeyPath != "" || config.TokenPath != "" || config.Tickets != nil
	if !configured {
		return nil, nil
	}
	if strings.TrimSpace(config.Address) == "" || config.CertificatePath == "" || config.PrivateKeyPath == "" || config.TokenPath == "" || config.Tickets == nil || host == nil {
		return nil, errors.New("server: worker control requires listen address, TLS identity, token, tickets, and host")
	}
	hostName, port, err := net.SplitHostPort(config.Address)
	if err != nil || port == "" || net.ParseIP(hostName) == nil || !net.ParseIP(hostName).IsLoopback() {
		return nil, errors.New("server: worker control must listen on an explicit loopback host and port")
	}
	certificate, err := tls.LoadX509KeyPair(config.CertificatePath, config.PrivateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("server: load worker-control certificate: %w", err)
	}
	if len(certificate.Certificate) == 0 {
		return nil, errors.New("server: worker-control certificate is empty")
	}
	fingerprintBytes := sha256.Sum256(certificate.Certificate[0])
	fingerprint := "sha256:" + hex.EncodeToString(fingerprintBytes[:])
	token, err := ReadAdmissionKey(config.TokenPath)
	if err != nil {
		return nil, fmt.Errorf("server: read worker-control token: %w", err)
	}
	worker, err := realm.NewInProcessWorkerWithMemberships(host, config.Destination, config.Memberships)
	if err != nil {
		return nil, err
	}
	tickets, err := realm.NewLocalTicketIssuer(config.Tickets)
	if err != nil {
		return nil, err
	}
	handler, err := realm.NewWorkerHTTPHandler(worker, tickets, string(token), config.Drain)
	if err != nil {
		return nil, err
	}
	listener, err := net.Listen("tcp", config.Address)
	if err != nil {
		return nil, fmt.Errorf("server: listen for worker control: %w", err)
	}
	tlsListener := tls.NewListener(listener, &tls.Config{Certificates: []tls.Certificate{certificate}, MinVersion: tls.VersionTLS13})
	return &WorkerControlServer{listener: tlsListener, fingerprint: fingerprint, server: &http.Server{Handler: handler,
		ReadHeaderTimeout: 5 * time.Second, IdleTimeout: 30 * time.Second, MaxHeaderBytes: 8 << 10}}, nil
}

func (server *WorkerControlServer) TLSFingerprint() string {
	if server == nil {
		return ""
	}
	return server.fingerprint
}

func (server *WorkerControlServer) Addr() net.Addr {
	if server == nil || server.listener == nil {
		return nil
	}
	return server.listener.Addr()
}

func (server *WorkerControlServer) Serve(ctx context.Context) error {
	if server == nil || server.listener == nil || server.server == nil || ctx == nil {
		return errors.New("server: invalid worker-control server")
	}
	shutdownDone := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
			shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
			_ = server.server.Shutdown(shutdownCtx)
			cancel()
		case <-shutdownDone:
		}
	}()
	err := server.server.Serve(server.listener)
	close(shutdownDone)
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

func (server *WorkerControlServer) Close(ctx context.Context) error {
	if server == nil || server.server == nil {
		return nil
	}
	if ctx == nil {
		ctx = context.Background()
	}
	return server.server.Shutdown(ctx)
}
