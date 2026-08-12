package d2legacy

import (
	"context"
	"fmt"
	"io/fs"

	"github.com/gravestench/dark-magic/internal/app/host"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	gamesession "github.com/gravestench/dark-magic/internal/game/session"
	"github.com/gravestench/dark-magic/internal/game/simulation"
	adaptercatalog "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/catalog"
	"github.com/gravestench/dark-magic/internal/mod/d2legacy/data/recovered"
	modruntime "github.com/gravestench/dark-magic/internal/runtime/lua"
)

// Records is the policy-neutral table reader required by d2legacy. The engine
// decides how bytes are mounted and decoded; Lua decides what the rows mean.
type Records interface {
	Load(string) ([]map[string]string, error)
	Invalidate(string)
	Loaded(string) bool
}

// Authority is one renderer-free authoritative mod instance. Client and server
// composition roots use the same object so neither can quietly acquire a
// different set of capabilities, streams, or lifecycle rules.
type Authority struct {
	Runtime  *modruntime.Runtime
	State    *simulation.StateStore
	Random   *simulation.RandomStreams
	Identity simulation.RuntimeIdentity

	component host.Component
}

// Config describes the deterministic inputs needed to start d2legacy. Restore
// contains opaque participant snapshots from a replay or checkpoint. They are
// applied before registration, so the session records the restored state as its
// starting point instead of briefly observing fresh mod state.
type Config struct {
	Seed    uint64
	Restore []simulation.ParticipantState
	// InitialData contains immutable import/bootstrap values. It is deliberately
	// absent from mutable runtime APIs after d2legacy materializes its own state.
	InitialData map[string]any
}

func Start(ctx context.Context, source fs.FS, records Records, engine *gameecs.Engine, session *gamesession.Session, seed uint64) (*Authority, error) {
	return StartWithConfig(ctx, source, records, engine, session, Config{Seed: seed})
}

// StartWithConfig constructs the same renderer-free authority used by clients,
// servers, replay verification, and restored sessions.
func StartWithConfig(ctx context.Context, source fs.FS, records Records, engine *gameecs.Engine, session *gamesession.Session, config Config) (*Authority, error) {
	if source == nil || records == nil || engine == nil || session == nil {
		return nil, fmt.Errorf("d2legacy: content, records, engine, and session are required")
	}
	identity, err := Identity(source, config.InitialData)
	if err != nil {
		return nil, err
	}
	streams, err := NewRandomStreams(config.Seed)
	if err != nil {
		return nil, err
	}
	result := &Authority{Runtime: modruntime.New(), State: simulation.NewStateStore(), Random: streams, Identity: identity}
	identityState, err := simulation.NewIdentityParticipant(identity)
	if err != nil {
		return nil, err
	}
	participants := map[string]simulation.StateParticipant{
		identityState.StateID(): identityState,
		streams.StateID():       streams,
	}
	var restoredState []byte
	for _, restored := range config.Restore {
		if restored.ID == result.State.StateID() {
			restoredState = append([]byte(nil), restored.Data...)
			continue
		}
		participant, found := participants[restored.ID]
		if !found {
			return nil, fmt.Errorf("d2legacy: unknown restored participant %q", restored.ID)
		}
		if err := participant.RestoreState(restored.Data); err != nil {
			return nil, fmt.Errorf("d2legacy: restore participant %q: %w", restored.ID, err)
		}
		delete(participants, restored.ID)
	}
	if len(config.Restore) > 0 && (len(participants) > 0 || restoredState == nil) {
		return nil, fmt.Errorf("d2legacy: restored state is missing %d participants", len(participants))
	}
	if err := ConfigureRuntime(result.Runtime, source, records, engine, session, result.State, streams, config.InitialData); err != nil {
		return nil, err
	}
	if err := result.Runtime.Start(ctx); err != nil {
		return nil, err
	}
	definition, err := modruntime.LoadDefinition(ctx, result.Runtime, source, "components/d2legacy.lua")
	if err != nil {
		_ = result.Runtime.Stop(context.Background())
		return nil, err
	}
	component, err := definition.Managed().New(ctx)
	if err == nil {
		err = component.Start(ctx)
	}
	if err != nil {
		_ = result.Runtime.Stop(context.Background())
		return nil, err
	}
	// Lua has now declared every durable store and its schema. Restoring before
	// this point would compare the checkpoint against an empty registry.
	if restoredState != nil {
		if err := result.State.RestoreState(restoredState); err != nil {
			_ = component.Stop(context.Background())
			_ = result.Runtime.Stop(context.Background())
			return nil, fmt.Errorf("d2legacy: restore participant %q: %w", result.State.StateID(), err)
		}
	}
	if err := session.RegisterAuthoritativeRuntime(identity, result.State, streams); err != nil {
		_ = component.Stop(context.Background())
		_ = result.Runtime.Stop(context.Background())
		return nil, err
	}
	result.component = component
	return result, nil
}

// ConfigureRuntime installs the one canonical authoritative d2legacy capability
// set into a stopped Lua runtime. The interactive client adds presentation
// capabilities to this same runtime; headless servers add nothing. Keeping the
// authority wiring here prevents the two hosts from quietly drifting apart.
func ConfigureRuntime(runtime *modruntime.Runtime, source fs.FS, records Records, engine *gameecs.Engine, session *gamesession.Session, state *simulation.StateStore, random *simulation.RandomStreams, initial map[string]any) error {
	if runtime == nil || source == nil || records == nil || engine == nil || session == nil || state == nil || random == nil {
		return fmt.Errorf("d2legacy: complete runtime dependencies are required")
	}
	if err := runtime.RegisterInstaller(modruntime.ContentRequire(source, "lua")); err != nil {
		return err
	}
	recoveredCatalog := recovered.New(source)
	for _, module := range []modruntime.Module{
		modruntime.DeterministicModule(), modruntime.WorldgenModule(),
		modruntime.RecordsModule(records), modruntime.AuthorityStateModule(state), modruntime.AuthorityRandomModule(random),
		modruntime.AuthorityCommandModule(runtime, session), modruntime.InitialDataModule(initial),
		modruntime.NewECSCapability(runtime, engine).Module(),
	} {
		if err := runtime.RegisterModule(module); err != nil {
			return err
		}
	}
	// Interactive clients install locale-aware catalogs first; renderer-free
	// servers receive these policy-neutral defaults instead.
	for _, module := range []modruntime.Module{
		adaptercatalog.QuestModule(recoveredCatalog), adaptercatalog.MapModule(recoveredCatalog),
	} {
		if err := runtime.RegisterModuleDefault(module); err != nil {
			return err
		}
	}
	return nil
}

func (authority *Authority) Stop(ctx context.Context) error {
	if authority == nil {
		return nil
	}
	var componentErr error
	if authority.component != nil {
		componentErr = authority.component.Stop(ctx)
	}
	runtimeErr := authority.Runtime.Stop(ctx)
	if componentErr != nil {
		return componentErr
	}
	return runtimeErr
}
