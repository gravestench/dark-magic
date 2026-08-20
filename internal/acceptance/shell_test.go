package acceptance

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/gravestench/akara"
	"github.com/gravestench/dark-magic/internal/app/host"
	"github.com/gravestench/dark-magic/internal/audio"
	"github.com/gravestench/dark-magic/internal/content"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	gamesession "github.com/gravestench/dark-magic/internal/game/session"
	"github.com/gravestench/dark-magic/internal/game/simulation"
	"github.com/gravestench/dark-magic/internal/inputstate"
	"github.com/gravestench/dark-magic/internal/localization"
	d2legacy "github.com/gravestench/dark-magic/internal/mod/d2legacy"
	d2movement "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/movement"
	gameplayer "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/player"
	d2save "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/save"
	"github.com/gravestench/dark-magic/internal/modcache"
	"github.com/gravestench/dark-magic/internal/preferences"
	"github.com/gravestench/dark-magic/internal/presentation/navigation"
	"github.com/gravestench/dark-magic/internal/presentation/render"
	"github.com/gravestench/dark-magic/internal/runtime/lua"
	"github.com/gravestench/dark-magic/internal/video"
)

// TestEmbeddedD2LegacyNavigationAndResourceLifetime exercises the assembled shell and verifies owned resources close.
func TestEmbeddedD2LegacyNavigationAndResourceLifetime(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	d2legacySource := content.D2Legacy()

	builtin, err := modcache.DescribeBuiltin(d2legacySource)
	if err != nil {
		t.Fatal(err)
	}

	packageFS, err := modcache.NewPackageFS(builtin.Manifest, d2legacySource)
	if err != nil {
		t.Fatal(err)
	}

	contentFS, err := content.New(content.Layer{Name: "builtin:d2legacy", FS: packageFS})
	if err != nil {
		t.Fatal(err)
	}

	runtime := modruntime.New()
	// This broad navigation acceptance validates the assembled shell and
	// resource lifetime, not the default one-second Lua watchdog. Race
	// instrumentation can make the initial all-screen module graph exceed that
	// budget even with serialized tests, so keep a bounded but proportional
	// startup allowance for this fixture.
	if err := runtime.SetExecutionBudget(5 * time.Second); err != nil {
		t.Fatal(err)
	}

	navigator := navigation.New()
	scenes := modruntime.NewScenes(runtime, navigator)

	var (
		composer render.Composer
		input    inputstate.Store
	)
	scenes.SetInputStore(&input)

	var mixer audio.Mixer

	network := &shellNetworkController{}
	saves := d2save.New(d2save.Character{ID: "hero", Name: "Hero", Class: "Amazon", Level: 1,
		Stats: &d2save.Stats{Vitality: 20, Health: 50, MaxHealth: 50, Mana: 15, MaxMana: 15, Stamina: 84, MaxStamina: 84}})
	entitySimulation := gameecs.New()

	authority, err := gamesession.New(entitySimulation, gamesession.Config{Step: time.Second})
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		if err := authority.Close(); err != nil {
			t.Errorf("close authoritative session: %v", err)
		}
	}()

	mod, err := d2legacy.Start(ctx, d2legacySource, shellD2Records{}, entitySimulation, authority, 7)
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		if err := mod.Stop(ctx); err != nil {
			t.Errorf("stop d2legacy mod: %v", err)
		}
	}()

	movementController := &d2movement.MovementController{}

	movementSource, err := d2movement.NewMovementSource(
		entitySimulation,
		&input,
		"local-player",
		"game_world",
		movementController,
	)
	if err != nil {
		t.Fatal(err)
	}

	intentController := &gamesession.IntentController{}

	intentSource, err := gamesession.NewIntentSource(intentController, "local-player")
	if err != nil {
		t.Fatal(err)
	}

	entrySource, err := gameplayer.NewEntrySource(entitySimulation, saves, "local-player", 4096, 4096)
	if err != nil {
		t.Fatal(err)
	}

	sequencer := simulation.NewLocalSequencer()
	commandSource := func(tick uint64) []simulation.Command {
		commands := append(entrySource.Commands(tick), movementSource.Commands(tick)...)
		commands = append(commands, intentSource.Commands(tick)...)

		return sequencer.Assign(commands)
	}
	worldReady := make(chan struct{})

	loading := acceptanceLoadingCoordinatorWithWorld(worldReady)
	defer loading.Close()

	if err := runtime.RegisterInstaller(modruntime.PackageRequire(contentFS, []string{"d2legacy"})); err != nil {
		t.Fatal(err)
	}

	for _, module := range []modruntime.Module{
		modruntime.AppModule("test", func() {}),
		modruntime.VFSModule(contentFS),
		modruntime.DataModule(contentFS),
		modruntime.RecordsModule(shellD2Records{}),
		modruntime.CommandIntentModule(intentController),
		modruntime.InputModule(&input),
		modruntime.AudioModule(runtime, &mixer, contentFS),
		modruntime.SettingsModule(preferences.NewTransient(), &mixer),
		modruntime.VideoModule(runtime, video.Unavailable{}, contentFS),
		modruntime.LocaleModule(localization.New(contentFS, "English")),
		modruntime.RenderModule(runtime, &composer),
		d2save.Module(saves),
		modruntime.PlayerControlModule(movementController),
		modruntime.NewECSCapability(runtime, entitySimulation).Module(),
		modruntime.LoadingModule(loading),
		modruntime.NetworkModule(network),
		modruntime.RealmModule(nil),
		scenes.Module(),
	} {
		if err := runtime.RegisterModule(module); err != nil {
			t.Fatal(err)
		}
	}

	if err := runtime.Start(ctx); err != nil {
		t.Fatal(err)
	}

	defer func() {
		if err := runtime.Stop(ctx); err != nil {
			t.Errorf("stop shell runtime: %v", err)
		}
	}()

	components := host.NewManager()

	boot, err := modruntime.LoadDefinition(ctx, runtime, d2legacySource, "boot.lua")
	if err != nil {
		t.Fatal(err)
	}

	if err := components.Register(boot.Managed()); err != nil {
		t.Fatal(err)
	}

	if err := components.Enable(ctx, boot.ID); err != nil {
		t.Fatal(err)
	}

	if err := scenes.Flush(ctx); err != nil {
		t.Fatal(err)
	}

	assertStack(t, navigator, "loading")
	assertNodes(t, &composer, 2)

	publishAction(&input, "skip")

	if err := scenes.Update(ctx, time.Second/60); err != nil {
		t.Fatal(err)
	}

	assertStack(t, navigator, "title")
	assertNodes(t, &composer, 2)

	input.Publish(inputstate.Frame{})

	if err := scenes.Update(ctx, time.Second/60); err != nil {
		t.Fatal(err)
	}

	publishAction(&input, "skip")

	if err := scenes.Update(ctx, time.Second/60); err != nil {
		t.Fatal(err)
	}

	assertStack(t, navigator, "main_menu")
	assertNodes(t, &composer, 2)
	input.Publish(inputstate.Frame{})

	if err := scenes.Update(ctx, time.Second/60); err != nil {
		t.Fatal(err)
	}
	// Realm and its compact Gateway selector now sit between Single Player and
	// Other Multiplayer, so traverse all three rows before exercising TCP/IP.
	for range 3 {
		publishAction(&input, "down")

		if err := scenes.Update(ctx, time.Second/60); err != nil {
			t.Fatal(err)
		}

		input.Publish(inputstate.Frame{})

		if err := scenes.Update(ctx, time.Second/60); err != nil {
			t.Fatal(err)
		}
	}

	publishAction(&input, "confirm")

	if err := scenes.Update(ctx, time.Second/60); err != nil {
		t.Fatal(err)
	}

	assertStack(t, navigator, "tcpip")
	input.Publish(inputstate.Frame{})

	if err := scenes.Update(ctx, time.Second/60); err != nil {
		t.Fatal(err)
	}

	publishAction(&input, "down")

	if err := scenes.Update(ctx, time.Second/60); err != nil {
		t.Fatal(err)
	}

	input.Publish(inputstate.Frame{})

	if err := scenes.Update(ctx, time.Second/60); err != nil {
		t.Fatal(err)
	}

	publishAction(&input, "confirm")

	if err := scenes.Update(ctx, time.Second/60); err != nil {
		t.Fatal(err)
	}

	assertStack(t, navigator, "tcpip")
	input.Publish(inputstate.Frame{Text: "127.0.0.1"})

	if err := scenes.Update(ctx, time.Second/60); err != nil {
		t.Fatal(err)
	}

	input.Publish(inputstate.Frame{})

	if err := scenes.Update(ctx, time.Second/60); err != nil {
		t.Fatal(err)
	}

	publishAction(&input, "down")

	if err := scenes.Update(ctx, time.Second/60); err != nil {
		t.Fatal(err)
	}

	input.Publish(inputstate.Frame{})

	if err := scenes.Update(ctx, time.Second/60); err != nil {
		t.Fatal(err)
	}

	publishAction(&input, "confirm")

	if err := scenes.Update(ctx, time.Second/60); err != nil {
		t.Fatal(err)
	}

	if network.joined != "127.0.0.1" {
		t.Fatalf("join address = %q, want 127.0.0.1", network.joined)
	}

	assertStack(t, navigator, "tcpip")
	// The rest of this broad navigation acceptance continues through the
	// ordinary offline flow; clear the fixture-only rejected join state first.
	network.joined, network.phase = "", "frontend"

	publishAction(&input, "cancel")

	if err := scenes.Update(ctx, time.Second/60); err != nil {
		t.Fatal(err)
	}

	assertStack(t, navigator, "tcpip")
	input.Publish(inputstate.Frame{})

	if err := scenes.Update(ctx, time.Second/60); err != nil {
		t.Fatal(err)
	}

	publishAction(&input, "cancel")

	if err := scenes.Update(ctx, time.Second/60); err != nil {
		t.Fatal(err)
	}

	assertStack(t, navigator, "main_menu")

	// Credits are authored as a real scene rather than a placeholder route. In
	// the embedded/headless stack the localized MPQ payload is unavailable, so
	// this also exercises its explicit fallback copy and return navigation.
	input.Publish(inputstate.Frame{})

	if err := scenes.Update(ctx, time.Second/60); err != nil {
		t.Fatal(err)
	}

	for range 4 {
		publishAction(&input, "down")

		if err := scenes.Update(ctx, time.Second/60); err != nil {
			t.Fatal(err)
		}

		input.Publish(inputstate.Frame{})

		if err := scenes.Update(ctx, time.Second/60); err != nil {
			t.Fatal(err)
		}
	}

	publishAction(&input, "confirm")

	if err := scenes.Update(ctx, time.Second/60); err != nil {
		t.Fatal(err)
	}

	assertStack(t, navigator, "credits")
	input.Publish(inputstate.Frame{})

	if err := scenes.Update(ctx, time.Second/60); err != nil {
		t.Fatal(err)
	}

	publishAction(&input, "confirm")

	if err := scenes.Update(ctx, time.Second/60); err != nil {
		t.Fatal(err)
	}

	publishAction(&input, "cancel")

	if err := scenes.Update(ctx, time.Second/60); err != nil {
		t.Fatal(err)
	}

	assertStack(t, navigator, "main_menu")

	input.Publish(inputstate.Frame{})

	if err := scenes.Update(ctx, time.Second/60); err != nil {
		t.Fatal(err)
	}

	for range 5 {
		publishAction(&input, "down")

		if err := scenes.Update(ctx, time.Second/60); err != nil {
			t.Fatal(err)
		}

		input.Publish(inputstate.Frame{})

		if err := scenes.Update(ctx, time.Second/60); err != nil {
			t.Fatal(err)
		}
	}

	publishAction(&input, "confirm")

	if err := scenes.Update(ctx, time.Second/60); err != nil {
		t.Fatal(err)
	}

	assertStack(t, navigator, "cinematics")
	input.Publish(inputstate.Frame{})

	if err := scenes.Update(ctx, time.Second/60); err != nil {
		t.Fatal(err)
	}

	publishAction(&input, "cancel")

	if err := scenes.Update(ctx, time.Second/60); err != nil {
		t.Fatal(err)
	}

	assertStack(t, navigator, "main_menu")

	input.Publish(inputstate.Frame{})

	if err := scenes.Update(ctx, time.Second/60); err != nil {
		t.Fatal(err)
	}

	publishAction(&input, "confirm")

	if err := scenes.Update(ctx, time.Second/60); err != nil {
		t.Fatal(err)
	}

	assertStack(t, navigator, "character_select")
	assertNodes(t, &composer, 2)

	input.Publish(inputstate.Frame{})

	if err := scenes.Update(ctx, time.Second/60); err != nil {
		t.Fatal(err)
	}

	publishAction(&input, "confirm")

	if err := scenes.Update(ctx, time.Second/60); err != nil {
		t.Fatal(err)
	}
	// A single activation selects the row. Move through the page scrollbar and
	// footer controls to the explicit OK button before launching.
	for range 5 {
		input.Publish(inputstate.Frame{})

		if err := scenes.Update(ctx, time.Second/60); err != nil {
			t.Fatal(err)
		}

		publishAction(&input, "down")

		if err := scenes.Update(ctx, time.Second/60); err != nil {
			t.Fatal(err)
		}
	}

	input.Publish(inputstate.Frame{})

	if err := scenes.Update(ctx, time.Second/60); err != nil {
		t.Fatal(err)
	}

	publishAction(&input, "confirm")

	if err := scenes.Update(ctx, time.Second/60); err != nil {
		t.Fatal(err)
	}

	assertStack(t, navigator, "game_loading")
	input.Publish(inputstate.Frame{})

	if err := scenes.Update(ctx, 2*time.Second); err != nil {
		t.Fatal(err)
	}

	assertStack(t, navigator, "game_loading")
	close(worldReady)

	for frame := 0; frame < 120 && reflect.DeepEqual(navigator.Stack(), []string{"game_loading"}); frame++ {
		time.Sleep(time.Millisecond)

		if err := scenes.Update(ctx, time.Second/60); err != nil {
			t.Fatal(err)
		}
	}

	assertStack(t, navigator, "game_world")
	// The compatibility hero rectangle is gone. Headless metadata-only saves
	// retain the world root and cursor; native asset-backed runs add the HUD and
	// an appearance node only when authoritative presentation data exists.
	assertNodes(t, &composer, 2)

	if selected, ok := saves.Selected(); !ok || selected.ID != "hero" {
		t.Fatalf("selected character = %#v, %v", selected, ok)
	}

	// Side-panel hotkeys share the same slot-aware operation as mini-panel
	// buttons: opposite sides coexist, same-side panels replace, a non-top side
	// can toggle closed, and a full overlay evicts both sides.
	for _, action := range []string{"inventory", "character"} {
		publishAction(&input, action)

		if err := scenes.Update(ctx, time.Second/60); err != nil {
			t.Fatal(err)
		}

		input.Publish(inputstate.Frame{})

		if err := scenes.Update(ctx, time.Second/60); err != nil {
			t.Fatal(err)
		}
	}

	assertStack(t, navigator, "game_world", "inventory", "character")
	publishAction(&input, "skills")

	if err := scenes.Update(ctx, time.Second/60); err != nil {
		t.Fatal(err)
	}

	assertStack(t, navigator, "game_world", "character", "skills")
	input.Publish(inputstate.Frame{})

	if err := scenes.Update(ctx, time.Second/60); err != nil {
		t.Fatal(err)
	}

	publishAction(&input, "character")

	if err := scenes.Update(ctx, time.Second/60); err != nil {
		t.Fatal(err)
	}

	assertStack(t, navigator, "game_world", "skills")
	input.Publish(inputstate.Frame{})

	if err := scenes.Update(ctx, time.Second/60); err != nil {
		t.Fatal(err)
	}

	publishAction(&input, "help")

	if err := scenes.Update(ctx, time.Second/60); err != nil {
		t.Fatal(err)
	}

	assertStack(t, navigator, "game_world", "help")
	input.Publish(inputstate.Frame{})

	if err := scenes.Update(ctx, time.Second/60); err != nil {
		t.Fatal(err)
	}

	publishAction(&input, "help")

	if err := scenes.Update(ctx, time.Second/60); err != nil {
		t.Fatal(err)
	}

	assertStack(t, navigator, "game_world")

	for _, overlay := range []string{"inventory", "character", "skills", "automap", "options", "pause"} {
		input.Publish(inputstate.Frame{})

		if err := scenes.Update(ctx, time.Second/60); err != nil {
			t.Fatal(err)
		}

		publishAction(&input, overlay)

		if err := scenes.Update(ctx, time.Second/60); err != nil {
			t.Fatal(err)
		}

		assertStack(t, navigator, "game_world", overlay)

		wantNodes := 3
		if overlay == "options" || overlay == "pause" {
			// Escape overlays keep their coordinate root at the viewport origin
			// and center a separate full-screen dimming backdrop below the menu.
			wantNodes++
		}

		assertNodes(t, &composer, wantNodes)
		input.Publish(inputstate.Frame{})

		if err := scenes.Update(ctx, time.Second/60); err != nil {
			t.Fatal(err)
		}

		publishAction(&input, "cancel")

		if err := scenes.Update(ctx, time.Second/60); err != nil {
			t.Fatal(err)
		}

		assertStack(t, navigator, "game_world")
		assertNodes(t, &composer, 2)
	}

	for cycle := 0; cycle < 50; cycle++ {
		input.Publish(inputstate.Frame{})

		if err := scenes.Update(ctx, time.Second/60); err != nil {
			t.Fatal(err)
		}

		publishAction(&input, "inventory")

		if err := scenes.Update(ctx, time.Second/60); err != nil {
			t.Fatal(err)
		}

		input.Publish(inputstate.Frame{})

		if err := scenes.Update(ctx, time.Second/60); err != nil {
			t.Fatal(err)
		}

		publishAction(&input, "cancel")

		if err := scenes.Update(ctx, time.Second/60); err != nil {
			t.Fatal(err)
		}
	}

	if diagnostics := composer.Diagnostics(); diagnostics.ActiveNodes != 3 || diagnostics.NodeSlots > 6 {
		t.Fatalf("composer diagnostics after rapid transitions = %#v", diagnostics)
	}

	if _, err := authority.AdvanceWithSource(time.Second, commandSource); err != nil {
		t.Fatal(err)
	}

	positions, found := akara.GetDynamicStore(entitySimulation.World(), "d2legacy.world.position")
	if !found {
		t.Fatal("game world did not register position component")
	}

	players, found := akara.GetDynamicStore(entitySimulation.World(), "d2legacy.world.player_control")
	if !found {
		t.Fatal("game world did not register player-control component")
	}

	heroes, err := entitySimulation.World().Subscribe(akara.All(positions, players))
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		if err := heroes.Close(); err != nil {
			t.Errorf("close hero query: %v", err)
		}
	}()

	entities := heroes.Entities()
	if len(entities) != 1 {
		t.Fatalf("player-controlled entities = %v", entities)
	}

	position, found := positions.Get(entities[0])
	if !found {
		t.Fatal("hero position is missing")
	}

	beforeValue, err := position.Get("x")
	if err != nil {
		t.Fatal(err)
	}

	before := beforeValue.(float64)

	input.Publish(inputstate.Frame{
		Actions: map[string]inputstate.ActionState{
			"right":      {Down: true},
			"toggle_run": {Pressed: true},
		},
		Owner: inputstate.FocusOwner{Domain: inputstate.FocusScene, ID: "game_world"},
	})

	if _, err := authority.AdvanceWithSource(time.Second, commandSource); err != nil {
		t.Fatal(err)
	}

	afterValue, err := position.Get("x")
	if err != nil {
		t.Fatal(err)
	}

	if after := afterValue.(float64); after <= before {
		t.Fatalf("hero did not move: %v -> %v", before, after)
	}

	modes, found := akara.GetDynamicStore(entitySimulation.World(), "d2legacy.player.movement_mode")
	if !found {
		t.Fatal("game world did not register movement-mode component")
	}

	mode, found := modes.Get(entities[0])
	if !found {
		t.Fatal("hero movement mode is missing")
	}

	if running, err := mode.Get("running"); err != nil || running != true {
		t.Fatalf("authoritative movement mode = %v, %v", running, err)
	}

	if err := intentController.Submit("player.assign_skills", map[string]any{"right": 42}); err != nil {
		t.Fatal(err)
	}

	input.Publish(inputstate.Frame{Owner: inputstate.FocusOwner{Domain: inputstate.FocusScene, ID: "game_world"}})

	if _, err := authority.AdvanceWithSource(time.Second, commandSource); err != nil {
		t.Fatal(err)
	}

	assignments, found := akara.GetDynamicStore(entitySimulation.World(), "d2legacy.player.skill_assignment")
	if !found {
		t.Fatal("game world did not register skill-assignment component")
	}

	assignment, found := assignments.Get(entities[0])
	if !found {
		t.Fatal("hero skill assignment is missing")
	}

	if right, err := assignment.Get("right"); err != nil || right != int64(42) {
		t.Fatalf("authoritative right skill = %v, %v", right, err)
	}

	if err := scenes.Close(ctx); err != nil {
		t.Fatal(err)
	}

	if err := components.Disable(ctx, boot.ID); err != nil {
		t.Fatal(err)
	}

	assertNodes(t, &composer, 0)

	if diagnostics := composer.Diagnostics(); diagnostics.ActiveNodes != 0 || diagnostics.ActiveResources != 0 {
		t.Fatalf("composer leaked resources: %#v", diagnostics)
	}

	if diagnostics := mixer.Diagnostics(); diagnostics.Active != 0 {
		t.Fatalf("mixer leaked sounds: %#v", diagnostics)
	}
}

type shellNetworkController struct {
	joined string
	phase  string
}

// Host accepts the shell's host action without external networking in this fixture.
func (*shellNetworkController) Host() error { return nil }

// StartSelected records the local-session phase exposed back to Lua navigation.
func (controller *shellNetworkController) StartSelected() error {
	controller.phase = "local"
	return nil
}

// Cancel is inert because this fixture has no asynchronous connection attempt.
func (*shellNetworkController) Cancel() {}

// Join records the attempted address and returns the stable failure rendered by the shell.
func (controller *shellNetworkController) Join(address string) error {
	controller.joined = address
	controller.phase = "failed"

	return errors.New("TEST CONNECTION FAILED")
}

// Status translates fixture state into the map-shaped contract consumed by Lua.
func (controller *shellNetworkController) Status() map[string]any {
	phase := controller.phase
	if phase == "" {
		phase = "frontend"
	}

	status := map[string]any{"phase": phase}
	if phase == "failed" {
		status["error"] = "TEST CONNECTION FAILED"
	}

	return status
}

type shellD2Records struct{}

// Invalidate is inert because fixture rows never change during the acceptance flow.
func (shellD2Records) Invalidate(string) {}

// Loaded reports fixture tables as immediately available so this test avoids I/O loading.
func (shellD2Records) Loaded(string) bool { return true }

// Load returns the minimal authored records needed to cross from the shell into gameplay.
func (shellD2Records) Load(path string) ([]map[string]string, error) {
	switch path {
	case "data/global/excel/charstats.txt":
		return []map[string]string{{
			"class": "Amazon", "StartSkill": "Test Skill", "WalkVelocity": "6", "RunVelocity": "9",
			"vit": "20", "stamina": "84", "RunDrain": "20", "StaminaPerLevel": "4", "StaminaPerVitality": "4",
		}}, nil
	case "data/global/excel/skilldesc.txt":
		return []map[string]string{
			{"skilldesc": "firebolt", "ListRow": "1", "IconCel": "0"},
			{"skilldesc": "test", "ListRow": "0", "IconCel": "0"},
		}, nil
	case "data/global/excel/skills.txt":
		return []map[string]string{
			{
				"Id": "36", "skill": "Fire Bolt", "skilldesc": "firebolt", "leftskill": "1",
				"general": "0", "passive": "0", "srvmissile": "firebolt", "etype": "fire",
				"interrupt": "1", "srvstfunc": "", "srvdofunc": "", "mana": "5", "manashift": "7",
				"emin": "3", "emax": "6", "HitShift": "8",
			},
			{"Id": "42", "skill": "Test Skill", "skilldesc": "test", "leftskill": "1", "general": "1", "passive": "0"},
		}, nil
	case "data/global/excel/Missiles.txt":
		return []map[string]string{{
			"Missile": "firebolt", "Skill": "Fire Bolt", "pSrvDoFunc": "1", "CollideType": "3",
			"CollideKill": "1", "Vel": "20", "Range": "40", "Size": "2", "CelFile": "firebolt",
			"AnimSpeed": "16", "NumDirections": "16", "LoopAnim": "1",
		}}, nil
	}

	return nil, nil
}

// publishAction emits one edge-triggered frame through the same input path as the shell.
func publishAction(input *inputstate.Store, name string) {
	input.Publish(inputstate.Frame{Actions: map[string]inputstate.ActionState{name: {Pressed: true}}})
}

// assertStack compares the entire navigation stack because overlay order changes input ownership.
func assertStack(t *testing.T, navigator *navigation.Manager, want ...string) {
	t.Helper()

	if got := navigator.Stack(); !reflect.DeepEqual(got, want) {
		t.Fatalf("stack = %v, want %v", got, want)
	}
}

// assertNodes catches leaked scene render resources immediately after the responsible transition.
func assertNodes(t *testing.T, composer *render.Composer, want int) {
	t.Helper()

	if got := len(composer.Snapshot()); got != want {
		t.Fatalf("render node count = %d, want %d; nodes=%#v", got, want, composer.Snapshot())
	}
}
