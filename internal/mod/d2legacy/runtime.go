package d2legacy

import (
	"context"
	"fmt"
	"io/fs"
	"time"

	"github.com/gravestench/dark-magic/internal/app/host"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	gamesession "github.com/gravestench/dark-magic/internal/game/session"
	"github.com/gravestench/dark-magic/internal/game/simulation"
	adaptercatalog "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/catalog"
	adaptermovement "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/movement"
	"github.com/gravestench/dark-magic/internal/mod/d2legacy/data/recovered"
	"github.com/gravestench/dark-magic/internal/modcache"
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

	components *host.Manager
}

type Extension struct {
	Manifest modcache.Manifest
	Source   fs.FS
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
	// Packages is the storage-neutral built-in-plus-extension recipe supplied by
	// product composition. Tests and bounded tools may omit it to derive the
	// vanilla built-in package identity directly from source.
	Packages simulation.RuntimePackageSet
	// AssetSetID is the path-independent digest of external game data mounted
	// by product composition. An empty value is valid only for embedded tests.
	AssetSetID string
	// GameDataGenerationID identifies the immutable authoritative record view.
	// When omitted, StartWithConfig reads it from a pinned Records implementation.
	GameDataGenerationID string
	// PackageContent and Extensions are supplied together by production
	// composition so private Lua namespaces and authoritative entrypoints from
	// every locked extension run on the actual authority.
	PackageContent fs.FS
	Extensions     []Extension
	// ExecutionBudget overrides the Lua runtime invocation budget when positive.
	// DisableExecutionBudget is reserved for bounded offline tools and large
	// deterministic test vectors; interactive and network hosts should keep a
	// finite budget.
	ExecutionBudget        time.Duration
	DisableExecutionBudget bool
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
	var identity simulation.RuntimeIdentity
	var err error
	if config.AssetSetID == "" {
		config.AssetSetID = simulation.EmptyAssetSetID
	}
	if config.GameDataGenerationID == "" {
		if pinned, ok := records.(interface{ GenerationID() string }); ok {
			config.GameDataGenerationID = pinned.GenerationID()
		}
	}
	if config.Packages.Base.ID == "" {
		identity, err = Identity(source, config.InitialData)
	} else {
		if config.GameDataGenerationID == "" {
			config.GameDataGenerationID = simulation.GameDataGenerationIDForAssetSet(config.AssetSetID)
		}
		identity, err = IdentityForPackagesAndData(source, config.Packages, config.AssetSetID, config.GameDataGenerationID, config.InitialData)
	}
	if err != nil {
		return nil, err
	}
	streams, err := NewRandomStreams(config.Seed)
	if err != nil {
		return nil, err
	}
	result := &Authority{Runtime: modruntime.New(), State: simulation.NewStateStore(), Random: streams, Identity: identity}
	if config.DisableExecutionBudget {
		if err := result.Runtime.SetExecutionBudget(0); err != nil {
			return nil, err
		}
	} else if config.ExecutionBudget > 0 {
		if err := result.Runtime.SetExecutionBudget(config.ExecutionBudget); err != nil {
			return nil, err
		}
	}
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
	if config.PackageContent != nil {
		ids := []string{"d2legacy"}
		for _, extension := range config.Extensions {
			ids = append(ids, extension.Manifest.ID)
		}
		if err := result.Runtime.RegisterInstaller(modruntime.PackageRequire(config.PackageContent, ids)); err != nil {
			return nil, err
		}
	}
	if err := ConfigureRuntime(result.Runtime, source, records, engine, session, result.State, streams, config.InitialData); err != nil {
		return nil, err
	}
	if err := validateCapabilityIdentity(identity, result.Runtime.ModuleNames()); err != nil {
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
	allDefinitions := []modruntime.Definition{definition}
	clientEntrypoints := []string{"d2legacy.boot"}
	authorityEntrypoints := []string{definition.ID}
	desired := map[string]bool{definition.ID: true}
	for _, extension := range config.Extensions {
		definitions, discoverErr := modruntime.DiscoverOwnedDefinitions(ctx, result.Runtime, extension.Source, extension.Manifest.ID)
		if discoverErr != nil {
			_ = result.Runtime.Stop(context.Background())
			return nil, fmt.Errorf("d2legacy: discover authority extension %q: %w", extension.Manifest.ID, discoverErr)
		}
		dependencies := make([]string, len(extension.Manifest.Dependencies))
		for index, dependency := range extension.Manifest.Dependencies {
			dependencies[index] = dependency.ID
		}
		if dependencyErr := modruntime.ValidateDefinitionDependencies(definitions, extension.Manifest.ID, dependencies); dependencyErr != nil {
			_ = result.Runtime.Stop(context.Background())
			return nil, dependencyErr
		}
		if entrypointErr := modruntime.ValidateDefinitionEntrypoints(definitions,
			extension.Manifest.Entrypoints.ClientComponents, extension.Manifest.Entrypoints.AuthorityComponents); entrypointErr != nil {
			_ = result.Runtime.Stop(context.Background())
			return nil, entrypointErr
		}
		for _, id := range extension.Manifest.Entrypoints.AuthorityComponents {
			desired[id] = true
		}
		allDefinitions = append(allDefinitions, definitions...)
		clientEntrypoints = append(clientEntrypoints, extension.Manifest.Entrypoints.ClientComponents...)
		authorityEntrypoints = append(authorityEntrypoints, extension.Manifest.Entrypoints.AuthorityComponents...)
	}
	if err := modruntime.ValidateDefinitionDomains(allDefinitions, clientEntrypoints, authorityEntrypoints); err != nil {
		_ = result.Runtime.Stop(context.Background())
		return nil, err
	}
	manager := host.NewManager()
	for _, candidate := range allDefinitions {
		if err := manager.Register(candidate.Managed()); err != nil {
			_ = result.Runtime.Stop(context.Background())
			return nil, err
		}
	}
	if err := manager.ApplyDesired(ctx, desired); err != nil {
		_ = result.Runtime.Stop(context.Background())
		return nil, err
	}
	// Lua has now declared every durable store and its schema. Restoring before
	// this point would compare the checkpoint against an empty registry.
	if restoredState != nil {
		if err := result.State.RestoreState(restoredState); err != nil {
			_ = manager.ApplyDesired(context.Background(), map[string]bool{})
			_ = result.Runtime.Stop(context.Background())
			return nil, fmt.Errorf("d2legacy: restore participant %q: %w", result.State.StateID(), err)
		}
	}
	if err := session.RegisterAuthoritativeRuntime(identity, result.State, streams); err != nil {
		_ = manager.ApplyDesired(context.Background(), map[string]bool{})
		_ = result.Runtime.Stop(context.Background())
		return nil, err
	}
	result.components = manager
	return result, nil
}

func validateCapabilityIdentity(identity simulation.RuntimeIdentity, moduleNames []string) error {
	expected := make(map[string]bool, len(identity.Recipe.CapabilityVersions))
	for name, version := range identity.Recipe.CapabilityVersions {
		expected[name+"/"+version] = true
	}
	actual := make(map[string]bool, len(moduleNames))
	for _, name := range moduleNames {
		actual[name] = true
	}
	for name := range expected {
		if !actual[name] {
			return fmt.Errorf("d2legacy: runtime identity names unavailable authoritative capability %q", name)
		}
	}
	for name := range actual {
		if !expected[name] {
			return fmt.Errorf("d2legacy: authoritative capability %q is absent from runtime identity", name)
		}
	}
	return nil
}

// ConfigureRuntime installs the one canonical authoritative d2legacy capability
// set into a stopped Lua runtime. The interactive client adds presentation
// capabilities to this same runtime; headless servers add nothing. Keeping the
// authority wiring here prevents the two hosts from quietly drifting apart.
func ConfigureRuntime(runtime *modruntime.Runtime, source fs.FS, records Records, engine *gameecs.Engine, session *gamesession.Session, state *simulation.StateStore, random *simulation.RandomStreams, initial map[string]any, supplied ...*modruntime.ECSCapability) error {
	if runtime == nil || source == nil || records == nil || engine == nil || session == nil || state == nil || random == nil {
		return fmt.Errorf("d2legacy: complete runtime dependencies are required")
	}
	if err := ConfigureModuleRuntime(runtime, source, records, random, initial); err != nil {
		return err
	}
	recoveredCatalog := recovered.New(source)
	for _, module := range []modruntime.Module{
		modruntime.AuthorityStateModule(state), modruntime.AuthorityCommandModule(runtime, session),
	} {
		if err := runtime.RegisterModule(module); err != nil {
			return err
		}
	}
	if err := ConfigureECSRuntime(runtime, engine, supplied...); err != nil {
		return err
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

// ConfigureModuleRuntime installs the production modules available to policy
// modules that do not need mutable authority or ECS access. Test profiles use
// this function so capability names and implementations cannot drift from the
// production authority composition.
func ConfigureModuleRuntime(runtime *modruntime.Runtime, source fs.FS, records Records, random *simulation.RandomStreams, initial map[string]any) error {
	if runtime == nil || source == nil || records == nil || random == nil {
		return fmt.Errorf("d2legacy: complete module runtime dependencies are required")
	}
	if err := runtime.RegisterInstallerDefault(modruntime.ContentRequire(source, "lua")); err != nil {
		return err
	}
	for _, module := range []modruntime.Module{
		modruntime.DeterministicModule(), modruntime.WorldgenModule(),
		modruntime.RecordsModule(records), modruntime.AuthorityRandomModule(random),
		modruntime.InitialDataModule(initial), adaptermovement.RulesModule(),
	} {
		if err := runtime.RegisterModule(module); err != nil {
			return err
		}
	}
	return nil
}

// ConfigureECSRuntime adds the same ECS capability used by production. It is
// separate from ConfigureModuleRuntime so narrow tests can prove their declared
// capability boundary without maintaining a parallel list of implementations.
func ConfigureECSRuntime(runtime *modruntime.Runtime, engine *gameecs.Engine, supplied ...*modruntime.ECSCapability) error {
	if runtime == nil || engine == nil {
		return fmt.Errorf("d2legacy: runtime and ECS engine are required")
	}
	capability := modruntime.NewECSCapability(runtime, engine)
	if len(supplied) > 0 && supplied[0] != nil {
		capability = supplied[0]
	}
	return runtime.RegisterModule(capability.Module())
}

func (authority *Authority) Stop(ctx context.Context) error {
	if authority == nil {
		return nil
	}
	var componentErr error
	if authority.components != nil {
		componentErr = authority.components.ApplyDesired(ctx, map[string]bool{})
	}
	runtimeErr := authority.Runtime.Stop(ctx)
	if componentErr != nil {
		return componentErr
	}
	return runtimeErr
}
