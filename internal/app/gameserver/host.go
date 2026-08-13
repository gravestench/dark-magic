// Package gameserver composes one renderer-free authoritative game session.
package gameserver

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"strings"

	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	gamesession "github.com/gravestench/dark-magic/internal/game/session"
	"github.com/gravestench/dark-magic/internal/game/simulation"
	d2legacy "github.com/gravestench/dark-magic/internal/mod/d2legacy"
)

// Config contains the deterministic and compatibility inputs fixed when a
// standalone session is allocated.
type Config struct {
	Mode        Mode
	SessionID   string
	Seed        uint64
	Prediction  gamesession.PredictionTier
	Session     gamesession.Config
	InitialData map[string]any
}

type Mode string

const (
	ModeStandalone Mode = "standalone"
	ModeListen     Mode = "listen"
	ModeRealm      Mode = "realm"
)

// Host owns one authoritative ECS, session, and d2legacy runtime. Allocation
// uses the identity registered by that exact runtime, never a client claim.
type Host struct {
	Mode       Mode
	Engine     *gameecs.Engine
	Session    *gamesession.Session
	Authority  *d2legacy.Authority
	Allocation gamesession.Allocation
}

// Start creates the renderer-free composition used by cmd/server, realm-owned
// workers, and future in-process listen servers.
func Start(ctx context.Context, source fs.FS, records d2legacy.Records, config Config) (*Host, error) {
	if ctx == nil || source == nil || records == nil {
		return nil, errors.New("game server: context, content, and records are required")
	}
	if strings.TrimSpace(config.SessionID) == "" {
		return nil, errors.New("game server: session ID is required")
	}
	if config.Mode == "" {
		config.Mode = ModeStandalone
	}
	if config.Mode != ModeStandalone && config.Mode != ModeListen && config.Mode != ModeRealm {
		return nil, fmt.Errorf("game server: unknown hosting mode %q", config.Mode)
	}
	if err := gamesession.ValidatePredictionTier(config.Prediction); err != nil {
		return nil, err
	}
	engine := gameecs.New()
	session, err := gamesession.New(engine, config.Session)
	if err != nil {
		_ = engine.Close()
		return nil, err
	}
	authority, err := d2legacy.StartWithConfig(ctx, source, records, engine, session, d2legacy.Config{
		Seed: config.Seed, InitialData: config.InitialData,
	})
	if err != nil {
		_ = session.Close()
		_ = engine.Close()
		return nil, err
	}
	allocation, err := gamesession.Allocate(config.SessionID, authority.Identity, config.Prediction)
	if err != nil {
		_ = authority.Stop(context.Background())
		_ = session.Close()
		_ = engine.Close()
		return nil, fmt.Errorf("game server: allocate session: %w", err)
	}
	return &Host{Mode: config.Mode, Engine: engine, Session: session, Authority: authority, Allocation: allocation}, nil
}

// Admit validates a client against the exact runtime already running here.
func (host *Host) Admit(characterID string, client simulation.RuntimeIdentity) (gamesession.AdmissionToken, error) {
	if host == nil {
		return gamesession.AdmissionToken{}, errors.New("game server: nil host")
	}
	return host.Allocation.Admit(characterID, client)
}

// ValidateReconnect rejects stale tokens and clients whose runtime changed.
func (host *Host) ValidateReconnect(token gamesession.AdmissionToken, client simulation.RuntimeIdentity) error {
	if host == nil {
		return errors.New("game server: nil host")
	}
	return host.Allocation.ValidateReconnect(token, client)
}

// Close releases the mod before its session and ECS dependencies.
func (host *Host) Close(ctx context.Context) error {
	if host == nil {
		return nil
	}
	if ctx == nil {
		ctx = context.Background()
	}
	var authorityErr, sessionErr, engineErr error
	if host.Authority != nil {
		authorityErr = host.Authority.Stop(ctx)
	}
	if host.Session != nil {
		sessionErr = host.Session.Close()
	}
	if host.Engine != nil {
		engineErr = host.Engine.Close()
	}
	return errors.Join(authorityErr, sessionErr, engineErr)
}
