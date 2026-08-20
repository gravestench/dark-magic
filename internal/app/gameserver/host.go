// Package gameserver composes one renderer-free authoritative game session.
package gameserver

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"path"
	"strings"

	recordstore "github.com/gravestench/dark-magic/internal/game/data/store"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	gamesession "github.com/gravestench/dark-magic/internal/game/session"
	"github.com/gravestench/dark-magic/internal/game/simulation"
	d2legacy "github.com/gravestench/dark-magic/internal/mod/d2legacy"
	"github.com/gravestench/dark-magic/internal/modcache"
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
	Packages    simulation.RuntimePackageSet
	AssetSetID  string
	Content     fs.FS
	Mods        *modcache.ResolvedSet
}

// Mode describes which ownership layer is hosting the shared authoritative composition.
type Mode string

const (
	// ModeStandalone runs without a listening transport or Realm worker owner.
	ModeStandalone Mode = "standalone"
	// ModeListen exposes the shared authority through a locally owned network listener.
	ModeListen Mode = "listen"
	// ModeRealm gives Realm worker lifecycle code ownership of the shared authority.
	ModeRealm Mode = "realm"
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
	config, err := validateHostConfig(ctx, source, records, config)
	if err != nil {
		return nil, err
	}

	records, err = pinHostRecords(records, config.Content)
	if err != nil {
		return nil, err
	}

	engine := gameecs.New()

	session, err := gamesession.New(engine, config.Session)
	if err != nil {
		_ = engine.Close()

		return nil, err
	}

	d2config, err := buildAuthorityConfig(config, records)
	if err != nil {
		closeSessionFoundation(session, engine)

		return nil, err
	}

	authority, err := d2legacy.StartWithConfig(ctx, source, records, engine, session, d2config)
	if err != nil {
		closeSessionFoundation(session, engine)

		return nil, err
	}

	allocation, err := gamesession.Allocate(config.SessionID, authority.Identity, config.Prediction)
	if err != nil {
		_ = authority.Stop(context.Background())

		closeSessionFoundation(session, engine)

		return nil, fmt.Errorf("game server: allocate session: %w", err)
	}

	return &Host{Mode: config.Mode, Engine: engine, Session: session, Authority: authority, Allocation: allocation}, nil
}

// validateHostConfig normalizes the default mode while rejecting incomplete or unsupported composition inputs.
func validateHostConfig(
	ctx context.Context,
	source fs.FS,
	records d2legacy.Records,
	config Config,
) (Config, error) {
	if ctx == nil || source == nil || records == nil {
		return Config{}, errors.New("game server: context, content, and records are required")
	}

	if strings.TrimSpace(config.SessionID) == "" {
		return Config{}, errors.New("game server: session ID is required")
	}

	if config.Mode == "" {
		config.Mode = ModeStandalone
	}

	if config.Mode != ModeStandalone && config.Mode != ModeListen && config.Mode != ModeRealm {
		return Config{}, fmt.Errorf("game server: unknown hosting mode %q", config.Mode)
	}

	if err := gamesession.ValidatePredictionTier(config.Prediction); err != nil {
		return Config{}, err
	}

	return config, nil
}

// pinnableRecords is the optional content capability required to freeze authoritative table provenance.
type pinnableRecords interface {
	fs.FS
	List(string, string) ([]string, error)
	ResolveSource(string) (string, string, error)
}

// pinHostRecords prefers content-pinned records but tolerates distributions with no authoritative tables.
func pinHostRecords(records d2legacy.Records, content fs.FS) (d2legacy.Records, error) {
	pinnable, ok := content.(pinnableRecords)
	if !ok {
		return records, nil
	}

	pinned, _, err := recordstore.Pin(pinnable)
	if errors.Is(err, recordstore.ErrNoAuthoritativeTables) {
		return records, nil
	}

	if err != nil {
		return nil, fmt.Errorf("game server: pin authoritative records: %w", err)
	}

	return pinned, nil
}

// buildAuthorityConfig binds immutable data generation and extension sources to the runtime started by this host.
func buildAuthorityConfig(config Config, records d2legacy.Records) (d2legacy.Config, error) {
	generationID := hostGameDataGenerationID(config.AssetSetID, records)
	initialData := cloneInitialData(config.InitialData)
	initialData["engine.game_data_generation_id"] = generationID

	authorityConfig := d2legacy.Config{
		Seed:                 config.Seed,
		InitialData:          initialData,
		Packages:             config.Packages,
		AssetSetID:           config.AssetSetID,
		GameDataGenerationID: generationID,
	}
	if config.Mods == nil {
		return authorityConfig, nil
	}

	if config.Content == nil {
		return d2legacy.Config{}, errors.New("game server: resolved mods require package content")
	}

	authorityConfig.PackageContent = config.Content
	for _, pkg := range config.Mods.Extensions.Packages {
		extensionSource, err := fs.Sub(config.Content, path.Join("mods", pkg.Manifest.ID))
		if err != nil {
			return d2legacy.Config{}, fmt.Errorf("game server: resolve extension %q: %w", pkg.Manifest.ID, err)
		}

		authorityConfig.Extensions = append(authorityConfig.Extensions, d2legacy.Extension{
			Manifest: pkg.Manifest,
			Source:   extensionSource,
		})
	}

	return authorityConfig, nil
}

// hostGameDataGenerationID uses pinned-record provenance when available and otherwise derives it from the asset set.
func hostGameDataGenerationID(assetSetID string, records d2legacy.Records) string {
	if pinned, ok := records.(interface{ GenerationID() string }); ok && pinned.GenerationID() != "" {
		return pinned.GenerationID()
	}

	return simulation.GameDataGenerationIDForAssetSet(assetSetID)
}

// closeSessionFoundation unwinds resources in dependency order before authority ownership has been established.
func closeSessionFoundation(session *gamesession.Session, engine *gameecs.Engine) {
	_ = session.Close()
	_ = engine.Close()
}

// cloneInitialData prevents runtime bootstrap from mutating configuration retained by the caller.
func cloneInitialData(source map[string]any) map[string]any {
	result := make(map[string]any, len(source)+1)
	for key, value := range source {
		result[key] = value
	}

	return result
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
