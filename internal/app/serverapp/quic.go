// Package serverapp owns standalone game-server process composition below the
// deliberately thin cmd/server entry point.
package serverapp

import (
	"crypto/tls"
	"errors"
	"fmt"

	"github.com/gravestench/dark-magic/internal/app/gameserver"
	"github.com/gravestench/dark-magic/internal/app/gameserver/sessionquic"
	"github.com/gravestench/dark-magic/internal/app/realm"
	playeradapter "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/player"
	"github.com/gravestench/dark-magic/internal/modcache"
)

// QUICConfig groups the transport identity, admission policy, and optional
// session services that must agree before a public listener is opened.
type QUICConfig struct {
	Address, CertificatePath, PrivateKeyPath, AdmissionKeyPath, SessionID string
	RemoteProfile                                                         *RemoteProfileConfig
	ModCache                                                              *modcache.Store
	Tickets                                                               *gameserver.TicketAuthority
	RealmMemberships                                                      *realm.WorkerMemberships
}

// StartQUIC opens the session transport only when its complete TLS and
// admission configuration is present. Any failure after Listen closes the
// listener so callers never inherit a partially configured server.
func StartQUIC(config QUICConfig, host *gameserver.Host) (*sessionquic.Server, error) {
	if !quicConfigured(config) {
		return nil, nil
	}

	if err := validateQUICConfig(config); err != nil {
		return nil, err
	}

	certificate, err := tls.LoadX509KeyPair(config.CertificatePath, config.PrivateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("server: load QUIC certificate: %w", err)
	}

	authenticator, err := quicAuthenticator(config)
	if err != nil {
		return nil, err
	}

	endpoint, err := gameserver.NewEndpoint(host, authenticator, playeradapter.ProjectClientView)
	if err != nil {
		return nil, err
	}

	if err := configureEndpointLifecycle(endpoint, host, config.RealmMemberships); err != nil {
		return nil, err
	}

	server, err := sessionquic.Listen(
		config.Address,
		&tls.Config{Certificates: []tls.Certificate{certificate}},
		endpoint,
	)
	if err != nil {
		return nil, err
	}

	if err := configureQUICServices(server, host, authenticator, config); err != nil {
		_ = server.Close()
		return nil, err
	}

	return server, nil
}

// quicConfigured distinguishes an omitted optional transport from a partial
// configuration, which StartQUIC must reject rather than silently ignore.
func quicConfigured(config QUICConfig) bool {
	return config.Address != "" || config.CertificatePath != "" || config.PrivateKeyPath != "" ||
		config.AdmissionKeyPath != ""
}

// validateQUICConfig preserves the all-or-nothing transport identity contract
// before certificate files or listeners can create side effects.
func validateQUICConfig(config QUICConfig) error {
	if config.Address == "" || config.CertificatePath == "" || config.PrivateKeyPath == "" ||
		config.AdmissionKeyPath == "" {
		return errors.New("server: quic-listen, tls-cert, tls-key, and admission-key must be set together")
	}

	return nil
}

// quicAuthenticator reuses the Realm's ticket authority when supplied so the
// control and public transports revoke and authenticate the same tickets.
func quicAuthenticator(config QUICConfig) (*gameserver.TicketAuthority, error) {
	if config.Tickets != nil {
		return config.Tickets, nil
	}

	secret, err := ReadAdmissionKey(config.AdmissionKeyPath)
	if err != nil {
		return nil, err
	}

	return gameserver.NewTicketAuthority(secret, config.SessionID)
}

// configureEndpointLifecycle assigns mutually exclusive Realm and standalone
// departure ownership. Realm persistence must retain canonical state until its
// lease commits, while standalone sessions remove players on transport leave.
func configureEndpointLifecycle(
	endpoint *gameserver.Endpoint,
	host *gameserver.Host,
	memberships *realm.WorkerMemberships,
) error {
	// A missing HUD player means the snapshot is not ready yet, rather than a
	// protocol failure that should terminate the client session.
	endpoint.SetSnapshotPending(func(err error) bool { return errors.Is(err, playeradapter.ErrHUDPlayer) })

	if host.Mode == gameserver.ModeRealm {
		if memberships == nil {
			return errors.New("server: Realm QUIC requires shared worker memberships")
		}

		endpoint.SetLeave(func(principal gameserver.Principal) error {
			memberships.Expire(principal.PlayerID)
			return nil
		})
		endpoint.SetConnected(func(principal gameserver.Principal) {
			memberships.Connect(principal.PlayerID)
		})

		return nil
	}

	departures := &playeradapter.DepartureQueue{}

	endpoint.SetLeave(func(principal gameserver.Principal) error {
		return departures.Submit(host.Session, principal.PlayerID)
	})

	return nil
}

// configureQUICServices attaches package and profile services in their
// original order so an invalid recipe still fails before remote admission is
// constructed.
func configureQUICServices(
	server *sessionquic.Server,
	host *gameserver.Host,
	authenticator *gameserver.TicketAuthority,
	config QUICConfig,
) error {
	packages, err := NewPackageProvider(host.Allocation.Identity.Recipe, config.ModCache)
	if err != nil {
		return err
	}

	server.SetPackageProvider(packages)

	if config.RemoteProfile == nil {
		return nil
	}

	profiles, err := NewRemoteProfileAdmissions(host, authenticator, *config.RemoteProfile)
	if err != nil {
		return err
	}

	server.SetProfileAdmissions(profiles)

	return nil
}
