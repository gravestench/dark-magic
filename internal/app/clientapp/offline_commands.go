package clientapp

import (
	"errors"

	"github.com/gravestench/dark-magic/internal/game/simulation"
	gameworld "github.com/gravestench/dark-magic/internal/game/world"
	d2movement "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/movement"
	gameplayer "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/player"
)

// registerOfflineCommands connects movement, entry, and intent command sources.
func (app *application) registerOfflineCommands() error {
	if err := app.registerOfflineMovementSystem(); err != nil {
		return err
	}

	movementSource, movement, err := app.createOfflineMovementSource()
	if err != nil {
		return err
	}

	entrySource, err := app.createOfflineEntrySource()
	if err != nil {
		return err
	}

	app.installOfflineCommandSource(entrySource, movementSource)
	app.playerControl = movement

	return nil
}

// registerOfflineMovementSystem installs collision-aware velocity movement.
func (app *application) registerOfflineMovementSystem() error {
	bloodMoor := app.gameWorlds[2]
	if bloodMoor == nil {
		return errors.New("register hostile simulation: Blood Moor world is unavailable")
	}

	components := gameworld.VelocityComponents{
		Position: "d2legacy.world.position",
		Velocity: "d2legacy.world.velocity",
		Collider: "d2legacy.world.collider",
	}
	err := gameworld.RegisterVelocityMovement(app.entitySimulation, bloodMoor, components)

	return wrap("register generic velocity movement", err)
}

// createOfflineMovementSource creates local input and selects its navigation map.
func (app *application) createOfflineMovementSource() (
	*d2movement.MovementSource,
	*d2movement.MovementController,
	error,
) {
	movement := &d2movement.MovementController{}

	source, err := d2movement.NewMovementSource(
		app.entitySimulation,
		app.inputState,
		"local-player",
		"game_world",
		movement,
	)
	if err != nil {
		return nil, nil, wrap("create offline movement source", err)
	}

	// Keep the source on app so network mode can reuse the same input controller.
	app.movementSource = source

	worldMap := app.gameWorlds[app.activeWorldLevel]
	if worldMap == nil {
		return nil, nil, errors.New("load offline entry map: session world is unavailable")
	}

	source.SetNavigation(worldMap)

	return source, movement, nil
}

// createOfflineEntrySource admits the local player at the trusted world spawn.
func (app *application) createOfflineEntrySource() (*gameplayer.EntrySource, error) {
	entryLevel := app.activeWorldLevel
	worldMap := app.gameWorlds[entryLevel]

	spawn, found := app.gameWorldSpawns[entryLevel]
	if !found {
		return nil, errors.New("create offline player entry source: world has no trusted spawn subtile")
	}

	request := app.gameWorldZones[entryLevel].Request()

	destination, err := gameplayer.NewDestination(
		spawn[0],
		spawn[1],
		float64(worldMap.WidthSubtiles),
		float64(worldMap.HeightSubtiles),
		int64(request.Act),
		int64(request.LevelID),
	)
	if err != nil {
		return nil, wrap("create Act I town admission destination", err)
	}

	entry, err := gameplayer.NewEntrySourceForDestination(
		app.entitySimulation,
		app.saves,
		"local-player",
		destination,
	)
	if err != nil {
		return nil, wrap("create offline player entry source", err)
	}

	return entry, nil
}

// installOfflineCommandSource sequences entry, movement, and intent commands.
func (app *application) installOfflineCommandSource(
	entry *gameplayer.EntrySource,
	movement *d2movement.MovementSource,
) {
	sequencer := simulation.NewLocalSequencer()
	app.commandSource = func(tick uint64) []simulation.Command {
		commands := entry.Commands(tick)
		commands = append(commands, movement.Commands(tick)...)
		commands = append(commands, app.commandIntentSource.Commands(tick)...)

		return sequencer.Assign(commands)
	}
}
