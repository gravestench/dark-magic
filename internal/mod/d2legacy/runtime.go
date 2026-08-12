package d2legacy

import (
	"context"
	"fmt"
	"io/fs"

	"github.com/gravestench/dark-magic/internal/app/host"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	gamesession "github.com/gravestench/dark-magic/internal/game/session"
	"github.com/gravestench/dark-magic/internal/game/simulation"
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
	Runtime *modruntime.Runtime
	State   *simulation.StateStore
	Random  *simulation.RandomStreams

	component host.Component
}

// Config describes the deterministic inputs needed to start d2legacy. Restore
// contains opaque participant snapshots from a replay or checkpoint. They are
// applied before registration, so the session records the restored state as its
// starting point instead of briefly observing fresh mod state.
type Config struct {
	Seed    uint64
	Restore []simulation.ParticipantState
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
	identity, err := Identity(source)
	if err != nil {
		return nil, err
	}
	streams, err := NewRandomStreams(config.Seed)
	if err != nil {
		return nil, err
	}
	result := &Authority{Runtime: modruntime.New(), State: simulation.NewStateStore(), Random: streams}
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
	if err := result.Runtime.RegisterInstaller(modruntime.ContentRequire(source, "lua")); err != nil {
		return nil, err
	}
	for _, module := range []modruntime.Module{
		modruntime.RecordsModule(records), modruntime.AuthorityStateModule(result.State),
		modruntime.AuthorityRandomModule(streams), modruntime.AuthorityCommandModule(result.Runtime, session),
		modruntime.NewECSCapability(result.Runtime, engine).Module(),
	} {
		if err := result.Runtime.RegisterModule(module); err != nil {
			return nil, err
		}
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
