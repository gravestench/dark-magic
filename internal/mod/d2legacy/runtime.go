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

// Extension couples a locked manifest to the source tree that defines its
// managed components. Startup validates both together before activating any
// extension-owned authority entrypoint.
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

// Start constructs a vanilla authority from the minimal deterministic seed.
// It delegates to StartWithConfig so simple callers cannot drift from the
// checkpoint-aware production startup sequence.
func Start(
	ctx context.Context,
	source fs.FS,
	records Records,
	engine *gameecs.Engine,
	session *gamesession.Session,
	seed uint64,
) (*Authority, error) {
	return StartWithConfig(ctx, source, records, engine, session, Config{Seed: seed})
}

// StartWithConfig constructs the same renderer-free authority used by clients,
// servers, replay verification, and restored sessions.
func StartWithConfig(
	ctx context.Context,
	source fs.FS,
	records Records,
	engine *gameecs.Engine,
	session *gamesession.Session,
	config Config,
) (*Authority, error) {
	if source == nil || records == nil || engine == nil || session == nil {
		return nil, fmt.Errorf("d2legacy: content, records, engine, and session are required")
	}

	config = normalizeRuntimeConfig(config, records)

	identity, err := runtimeIdentity(source, config)
	if err != nil {
		return nil, err
	}

	streams, err := NewRandomStreams(config.Seed)
	if err != nil {
		return nil, err
	}

	result := &Authority{
		Runtime:  modruntime.New(),
		State:    simulation.NewStateStore(),
		Random:   streams,
		Identity: identity,
	}
	if err := configureExecutionBudget(result.Runtime, config); err != nil {
		return nil, err
	}

	restoredState, err := restoreFoundationalParticipants(result, config.Restore)
	if err != nil {
		return nil, err
	}

	if err := registerPackageInstaller(result.Runtime, config); err != nil {
		return nil, err
	}

	if err := ConfigureRuntime(
		result.Runtime,
		source,
		records,
		engine,
		session,
		result.State,
		streams,
		config.InitialData,
	); err != nil {
		return nil, err
	}

	if err := validateCapabilityIdentity(identity, result.Runtime.ModuleNames()); err != nil {
		return nil, err
	}

	if err := result.Runtime.Start(ctx); err != nil {
		return nil, err
	}

	definitions, desired, err := loadAuthorityDefinitions(ctx, result.Runtime, source, config.Extensions)
	if err != nil {
		_ = result.Runtime.Stop(context.Background())
		return nil, err
	}

	manager := host.NewManager()
	for _, candidate := range definitions {
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

// normalizeRuntimeConfig fills storage-derived identity fields before hashing
// the recipe. The Config value is copied, so callers do not observe defaults
// being written into their reusable configuration.
func normalizeRuntimeConfig(config Config, records Records) Config {
	if config.AssetSetID == "" {
		config.AssetSetID = simulation.EmptyAssetSetID
	}

	if config.GameDataGenerationID == "" {
		if pinned, ok := records.(interface{ GenerationID() string }); ok {
			config.GameDataGenerationID = pinned.GenerationID()
		}
	}

	if config.Packages.Base.ID != "" && config.GameDataGenerationID == "" {
		config.GameDataGenerationID = simulation.GameDataGenerationIDForAssetSet(config.AssetSetID)
	}

	return config
}

// runtimeIdentity selects the legacy shorthand only when no explicit package
// lock was supplied. Production package locks therefore verify both package and
// authoritative data generations before any runtime state is created.
func runtimeIdentity(source fs.FS, config Config) (simulation.RuntimeIdentity, error) {
	if config.Packages.Base.ID == "" {
		return Identity(source, config.InitialData)
	}

	return IdentityForPackagesAndData(
		source,
		config.Packages,
		config.AssetSetID,
		config.GameDataGenerationID,
		config.InitialData,
	)
}

// configureExecutionBudget preserves the explicit offline escape hatch while
// leaving the runtime's safe default intact when no override is requested.
func configureExecutionBudget(runtime *modruntime.Runtime, config Config) error {
	if config.DisableExecutionBudget {
		return runtime.SetExecutionBudget(0)
	}

	if config.ExecutionBudget > 0 {
		return runtime.SetExecutionBudget(config.ExecutionBudget)
	}

	return nil
}

// restoreFoundationalParticipants restores identity and random streams before
// Lua starts. StateStore bytes are copied and deferred because Lua must first
// declare every durable store and schema referenced by that checkpoint.
func restoreFoundationalParticipants(
	authority *Authority,
	restoredParticipants []simulation.ParticipantState,
) ([]byte, error) {
	identityState, err := simulation.NewIdentityParticipant(authority.Identity)
	if err != nil {
		return nil, err
	}

	pending := map[string]simulation.StateParticipant{
		identityState.StateID():    identityState,
		authority.Random.StateID(): authority.Random,
	}

	var restoredState []byte

	for _, restored := range restoredParticipants {
		if restored.ID == authority.State.StateID() {
			// Own the checkpoint bytes independently of caller buffers because the
			// actual state restoration happens later in the startup sequence.
			restoredState = append([]byte(nil), restored.Data...)
			continue
		}

		participant, found := pending[restored.ID]
		if !found {
			return nil, fmt.Errorf("d2legacy: unknown restored participant %q", restored.ID)
		}

		if err := participant.RestoreState(restored.Data); err != nil {
			return nil, fmt.Errorf("d2legacy: restore participant %q: %w", restored.ID, err)
		}

		delete(pending, restored.ID)
	}

	if len(restoredParticipants) > 0 && (len(pending) > 0 || restoredState == nil) {
		return nil, fmt.Errorf("d2legacy: restored state is missing %d participants", len(pending))
	}

	return restoredState, nil
}

// registerPackageInstaller exposes only the locked base and extension package
// namespaces. Skipping registration for embedded tests preserves their direct
// content loader and avoids broadening the production require contract.
func registerPackageInstaller(runtime *modruntime.Runtime, config Config) error {
	if config.PackageContent == nil {
		return nil
	}

	packageIDs := []string{"d2legacy"}
	for _, extension := range config.Extensions {
		packageIDs = append(packageIDs, extension.Manifest.ID)
	}

	return runtime.RegisterInstaller(modruntime.PackageRequire(config.PackageContent, packageIDs))
}

// loadAuthorityDefinitions discovers and validates the base and extension
// component graph before the host manager starts anything. Returning the full
// graph and desired set together prevents partially validated extensions from
// becoming observable.
func loadAuthorityDefinitions(
	ctx context.Context,
	runtime *modruntime.Runtime,
	source fs.FS,
	extensions []Extension,
) ([]modruntime.Definition, map[string]bool, error) {
	base, err := modruntime.LoadDefinition(ctx, runtime, source, "components/d2legacy.lua")
	if err != nil {
		return nil, nil, err
	}

	definitions := []modruntime.Definition{base}
	clientEntrypoints := []string{"d2legacy.boot"}
	authorityEntrypoints := []string{base.ID}
	desired := map[string]bool{base.ID: true}

	for _, extension := range extensions {
		extensionDefinitions, err := loadExtensionDefinitions(ctx, runtime, extension)
		if err != nil {
			return nil, nil, err
		}

		for _, id := range extension.Manifest.Entrypoints.AuthorityComponents {
			desired[id] = true
		}

		definitions = append(definitions, extensionDefinitions...)
		clientEntrypoints = append(
			clientEntrypoints,
			extension.Manifest.Entrypoints.ClientComponents...,
		)
		authorityEntrypoints = append(
			authorityEntrypoints,
			extension.Manifest.Entrypoints.AuthorityComponents...,
		)
	}

	if err := modruntime.ValidateDefinitionDomains(
		definitions,
		clientEntrypoints,
		authorityEntrypoints,
	); err != nil {
		return nil, nil, err
	}

	return definitions, desired, nil
}

// loadExtensionDefinitions checks ownership, declared dependencies, and both
// entrypoint lists for one extension. This keeps package metadata authoritative
// instead of trusting arbitrary component files found in its source tree.
func loadExtensionDefinitions(
	ctx context.Context,
	runtime *modruntime.Runtime,
	extension Extension,
) ([]modruntime.Definition, error) {
	definitions, err := modruntime.DiscoverOwnedDefinitions(
		ctx,
		runtime,
		extension.Source,
		extension.Manifest.ID,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"d2legacy: discover authority extension %q: %w",
			extension.Manifest.ID,
			err,
		)
	}

	dependencies := make([]string, len(extension.Manifest.Dependencies))
	for index, dependency := range extension.Manifest.Dependencies {
		dependencies[index] = dependency.ID
	}

	if err := modruntime.ValidateDefinitionDependencies(
		definitions,
		extension.Manifest.ID,
		dependencies,
	); err != nil {
		return nil, err
	}

	if err := modruntime.ValidateDefinitionEntrypoints(
		definitions,
		extension.Manifest.Entrypoints.ClientComponents,
		extension.Manifest.Entrypoints.AuthorityComponents,
	); err != nil {
		return nil, err
	}

	return definitions, nil
}

// validateCapabilityIdentity proves that the runtime's installed module set
// exactly matches the capability versions hashed into the session identity.
// Exact equality rejects both missing capabilities and undeclared additions.
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
func ConfigureRuntime(
	runtime *modruntime.Runtime,
	source fs.FS,
	records Records,
	engine *gameecs.Engine,
	session *gamesession.Session,
	state *simulation.StateStore,
	random *simulation.RandomStreams,
	initial map[string]any,
	supplied ...*modruntime.ECSCapability,
) error {
	missingRuntimeDependency := runtime == nil || source == nil || records == nil || engine == nil

	missingAuthorityDependency := session == nil || state == nil || random == nil
	if missingRuntimeDependency || missingAuthorityDependency {
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
func ConfigureModuleRuntime(
	runtime *modruntime.Runtime,
	source fs.FS,
	records Records,
	random *simulation.RandomStreams,
	initial map[string]any,
) error {
	if runtime == nil || source == nil || records == nil || random == nil {
		return fmt.Errorf("d2legacy: complete module runtime dependencies are required")
	}

	if err := runtime.RegisterInstallerDefault(modruntime.ContentRequire(source, "lua")); err != nil {
		return err
	}

	var animationSource interface {
		Read(string) ([]byte, error)
	}

	animationSource, _ = records.(interface {
		Read(string) ([]byte, error)
	})
	// The interactive client installs a presentation-profile-aware data module
	// before authority composition. Headless hosts need this policy-neutral
	// fallback, but must not replace or duplicate the richer client capability.
	if err := runtime.RegisterModuleDefault(modruntime.DataModule(source)); err != nil {
		return err
	}

	movementCatalog, err := adaptermovement.LoadCatalog(records)
	if err != nil {
		return err
	}

	for _, module := range []modruntime.Module{
		modruntime.DeterministicModule(), modruntime.WorldgenModule(),
		modruntime.RecordsModule(records), modruntime.AnimDataModule(animationSource),
		modruntime.AuthorityRandomModule(random),
		modruntime.InitialDataModule(initial), adaptermovement.RulesModule(movementCatalog),
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
func ConfigureECSRuntime(
	runtime *modruntime.Runtime,
	engine *gameecs.Engine,
	supplied ...*modruntime.ECSCapability,
) error {
	if runtime == nil || engine == nil {
		return fmt.Errorf("d2legacy: runtime and ECS engine are required")
	}

	capability := modruntime.NewECSCapability(runtime, engine)
	if len(supplied) > 0 && supplied[0] != nil {
		capability = supplied[0]
	}

	return runtime.RegisterModule(capability.Module())
}

// Stop deactivates managed components before stopping Lua so component cleanup
// can still call runtime-backed hooks. The first component error wins, but Lua
// is always stopped to avoid leaking the runtime after partial cleanup.
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
