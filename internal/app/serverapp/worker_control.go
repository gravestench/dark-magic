package serverapp

import (
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/gravestench/dark-magic/internal/app/gameserver"
	"github.com/gravestench/dark-magic/internal/app/realm"
	playeradapter "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/player"
)

// WorkerControlConfig describes the private Realm allocator channel. Its
// ticket authority must be the same instance used by the public QUIC endpoint.
type WorkerControlConfig struct {
	Address, CertificatePath, PrivateKeyPath, TokenPath string
	Tickets                                             *gameserver.TicketAuthority
	Destination                                         playeradapter.Destination
	Memberships                                         *realm.WorkerMemberships
	Drain                                               func()
}

// StartWorkerControl opens the private allocator listener only after its
// loopback boundary, TLS identity, token, and shared ticket authority validate.
func StartWorkerControl(config WorkerControlConfig, host *gameserver.Host) (*WorkerControlServer, error) {
	if !workerControlConfigured(config) {
		return nil, nil
	}

	if err := validateWorkerControlConfig(config, host); err != nil {
		return nil, err
	}

	certificate, fingerprint, err := loadWorkerControlIdentity(config)
	if err != nil {
		return nil, err
	}

	handler, err := newWorkerControlHandler(config, host)
	if err != nil {
		return nil, err
	}

	listener, err := net.Listen("tcp", config.Address)
	if err != nil {
		return nil, fmt.Errorf("server: listen for worker control: %w", err)
	}

	return newWorkerControlServer(listener, certificate, fingerprint, handler), nil
}

// workerControlConfigured distinguishes a disabled optional control plane from
// a partial configuration that could otherwise fail open.
func workerControlConfigured(config WorkerControlConfig) bool {
	return config.Address != "" || config.CertificatePath != "" || config.PrivateKeyPath != "" ||
		config.TokenPath != "" || config.Tickets != nil
}

// validateWorkerControlConfig enforces a numeric loopback address before any
// secret is read or listener is opened; DNS names are intentionally excluded.
func validateWorkerControlConfig(config WorkerControlConfig, host *gameserver.Host) error {
	if strings.TrimSpace(config.Address) == "" || config.CertificatePath == "" || config.PrivateKeyPath == "" ||
		config.TokenPath == "" || config.Tickets == nil || host == nil {
		return errors.New("server: worker control requires listen address, TLS identity, token, tickets, and host")
	}

	hostName, port, err := net.SplitHostPort(config.Address)

	ip := net.ParseIP(hostName)
	if err != nil || port == "" || ip == nil || !ip.IsLoopback() {
		return errors.New("server: worker control must listen on an explicit loopback host and port")
	}

	return nil
}

// loadWorkerControlIdentity loads the server certificate and derives the exact
// leaf fingerprint the allocator must pin before sending its bearer token.
func loadWorkerControlIdentity(config WorkerControlConfig) (tls.Certificate, string, error) {
	certificate, err := tls.LoadX509KeyPair(config.CertificatePath, config.PrivateKeyPath)
	if err != nil {
		return tls.Certificate{}, "", fmt.Errorf("server: load worker-control certificate: %w", err)
	}

	if len(certificate.Certificate) == 0 {
		return tls.Certificate{}, "", errors.New("server: worker-control certificate is empty")
	}

	fingerprintBytes := sha256.Sum256(certificate.Certificate[0])
	fingerprint := "sha256:" + hex.EncodeToString(fingerprintBytes[:])

	return certificate, fingerprint, nil
}

// newWorkerControlHandler composes the allocator API in security-sensitive
// order: token, worker membership, shared issuer, then authenticated handler.
func newWorkerControlHandler(config WorkerControlConfig, host *gameserver.Host) (http.Handler, error) {
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

	return realm.NewWorkerHTTPHandler(worker, tickets, string(token), config.Drain)
}
