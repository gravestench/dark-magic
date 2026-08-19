package clientapp

import (
	"io/fs"

	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	gamesession "github.com/gravestench/dark-magic/internal/game/session"
	"github.com/gravestench/dark-magic/internal/game/simulation"
	d2legacymod "github.com/gravestench/dark-magic/internal/mod/d2legacy"
	d2save "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/save"
	darkpaths "github.com/gravestench/dark-magic/internal/paths"
	modruntime "github.com/gravestench/dark-magic/internal/runtime/lua"
)

// buildOfflineSession composes durable player state, authoritative simulation,
// command sources, and loading prerequisites before any gameplay scene can open.
func (app *application) buildOfflineSession() error {
	if err := app.loadOfflinePlayer(); err != nil {
		return err
	}

	if err := app.createOfflineAuthority(); err != nil {
		return err
	}

	if err := app.registerOfflineCommands(); err != nil {
		return err
	}

	return app.buildLoadingCoordinator()
}

// loadOfflinePlayer establishes the single save store shared by offline play and
// connection controllers, then applies direct-start fixtures only when explicitly requested.
func (app *application) loadOfflinePlayer() error {
	app.configuredMods = cloneRuntimePackages(app.options.Packages)
	fixtures := developmentCharactersForScene(app.options.StartScene, app.options.FixtureCharacters)

	profilePath, err := darkpaths.ExpandHost(app.options.PlayerProfilePath)
	if err != nil {
		return wrap("expand player profile path", err)
	}

	app.saves, app.playerProfilePath, err = loadPlayerProfile(profilePath, fixtures)
	if err != nil {
		return err
	}

	// Both controllers share the loaded save store and trust configuration.
	app.network = newNetworkController(app)
	app.realm = newRealmController(app)

	return app.selectDevelopmentFixture(fixtures)
}

// selectDevelopmentFixture bypasses frontend selection only for known gameplay
// fixtures; ordinary startup must leave roster choice to the player.
func (app *application) selectDevelopmentFixture(fixtures []d2save.Character) error {
	if len(fixtures) == 0 || !fixtureNeedsSelection(app.options.StartScene) {
		return nil
	}

	return wrap("select development fixture", app.saves.Select(fixtures[0].ID))
}

// createOfflineAuthority creates the canonical ECS/session before registering
// deterministic state and random streams that participate in runtime identity.
func (app *application) createOfflineAuthority() error {
	app.entitySimulation = gameecs.New()
	if err := app.buildEntryWorld(); err != nil {
		return err
	}

	session, err := gamesession.New(app.entitySimulation, gamesession.Config{})
	if err != nil {
		_ = app.entitySimulation.Close()
		return wrap("create offline game session", err)
	}

	app.offlineSession = session

	// The state store and random streams are part of the authority's identity.
	app.authoritativeState = simulation.NewStateStore()

	app.authoritativeRandom, err = d2legacymod.NewRandomStreams(0)
	if err != nil {
		return wrap("register d2legacy random streams", err)
	}

	return app.configureOfflineRuntime(session)
}

// configureOfflineRuntime computes identity from packages, assets, records, and
// bootstrap data before exposing authority to Lua, preventing incompatible recomposition.
func (app *application) configureOfflineRuntime(session *gamesession.Session) error {
	initialData := app.sessionInitialData()

	d2legacySource, err := app.modSource("d2legacy")
	if err != nil {
		return wrap("resolve d2legacy package", err)
	}

	identity, err := d2legacymod.IdentityForPackagesAndData(
		d2legacySource,
		app.options.Packages,
		app.options.AssetSetID,
		app.gameDataGenerationID(),
		initialData,
	)
	if err != nil {
		return wrap("identify d2legacy mod", err)
	}

	if err := session.RegisterAuthoritativeRuntime(
		identity,
		app.authoritativeState,
		app.authoritativeRandom,
	); err != nil {
		return wrap("register d2legacy authoritative runtime", err)
	}

	if err := app.createCommandIntents(); err != nil {
		return err
	}

	// Capabilities and package namespaces must exist before ConfigureRuntime.
	app.ecsCapability = modruntime.NewECSCapability(app.scripts, app.entitySimulation)
	if err := app.configurePackageRegistry(); err != nil {
		return err
	}

	return app.configureD2LegacyRuntime(d2legacySource, initialData)
}

// createCommandIntents keeps generic actions separate from fixed-rate movement;
// merging their sequencing would let empty action batches consume movement ticks.
func (app *application) createCommandIntents() error {
	app.commandIntents = &gamesession.IntentController{}
	source, err := gamesession.NewIntentSource(app.commandIntents, "local-player")
	app.commandIntentSource = source

	return wrap("create local command intent source", err)
}

// configurePackageRegistry permits require only for resolved package namespaces
// and records digests used to invalidate precisely the modules changed by recomposition.
func (app *application) configurePackageRegistry() error {
	if app.options.Mods == nil {
		return nil
	}

	packages := app.options.Mods.Packages()

	packageIDs := make([]string, len(packages))
	for index, pkg := range packages {
		packageIDs[index] = pkg.Manifest.ID
	}

	app.packageRegistry = modruntime.NewPackageRegistry(packageIDs)

	installer := modruntime.PackageRequireRegistry(app.options.Content, app.packageRegistry)
	if err := app.scripts.RegisterInstaller(installer); err != nil {
		return wrap("register package Lua namespaces", err)
	}

	// Digests let network recomposition replace only packages that changed.
	app.packageDigests = make(map[string]string, len(packages))
	for _, pkg := range packages {
		app.packageDigests[pkg.Manifest.ID] = pkg.Descriptor.Digest
	}

	return nil
}

// configureD2LegacyRuntime is the final authority handoff to d2legacy Lua. Every
// service passed here is already identity-pinned and owned by application.
func (app *application) configureD2LegacyRuntime(source fs.FS, initialData map[string]any) error {
	err := d2legacymod.ConfigureRuntime(
		app.scripts,
		source,
		app.records,
		app.entitySimulation,
		app.offlineSession,
		app.authoritativeState,
		app.authoritativeRandom,
		initialData,
		app.ecsCapability,
	)

	return wrap("configure canonical d2legacy runtime", err)
}
