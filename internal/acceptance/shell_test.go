package acceptance

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/gravestench/akara"
	"github.com/gravestench/dark-magic/internal/app/host"
	"github.com/gravestench/dark-magic/internal/audio"
	"github.com/gravestench/dark-magic/internal/content"
	"github.com/gravestench/dark-magic/internal/game/data/catalog"
	"github.com/gravestench/dark-magic/internal/game/data/store"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	gameplayer "github.com/gravestench/dark-magic/internal/game/player"
	gamesession "github.com/gravestench/dark-magic/internal/game/session"
	"github.com/gravestench/dark-magic/internal/game/simulation"
	"github.com/gravestench/dark-magic/internal/inputstate"
	"github.com/gravestench/dark-magic/internal/localization"
	"github.com/gravestench/dark-magic/internal/persistence"
	"github.com/gravestench/dark-magic/internal/preferences"
	"github.com/gravestench/dark-magic/internal/presentation/navigation"
	"github.com/gravestench/dark-magic/internal/presentation/render"
	"github.com/gravestench/dark-magic/internal/runtime/lua"
	"github.com/gravestench/dark-magic/internal/video"
)

func TestEmbeddedShimNavigationAndResourceLifetime(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	contentFS, err := content.New(content.Layer{Name: "darkmagic", FS: content.Shim()})
	if err != nil {
		t.Fatal(err)
	}
	runtime := modruntime.New()
	navigator := navigation.New()
	scenes := modruntime.NewScenes(runtime, navigator)
	var composer render.Composer
	var input inputstate.Store
	scenes.SetInputStore(&input)
	var mixer audio.Mixer
	saves := persistence.New(persistence.Character{ID: "hero", Name: "Hero", Class: "Amazon", Level: 1})
	entitySimulation := gameecs.New()
	authority, err := gamesession.New(entitySimulation, gamesession.Config{Step: time.Second})
	if err != nil {
		t.Fatal(err)
	}
	defer authority.Close()
	if err := gamesession.RegisterMovement(authority); err != nil {
		t.Fatal(err)
	}
	if err := gamesession.RegisterSkillAssignments(authority); err != nil {
		t.Fatal(err)
	}
	if err := gameplayer.Register(authority); err != nil {
		t.Fatal(err)
	}
	movementController := &gamesession.MovementController{}
	movementSource, err := gamesession.NewMovementSource(entitySimulation, &input, "local-player", "game_world", movementController)
	if err != nil {
		t.Fatal(err)
	}
	skillSource, err := gamesession.NewSkillSource(movementController, "local-player")
	if err != nil {
		t.Fatal(err)
	}
	entrySource, err := gameplayer.NewEntrySource(entitySimulation, saves, "local-player", 4096, 4096)
	if err != nil {
		t.Fatal(err)
	}
	commandSource := func(tick uint64) []simulation.Command {
		commands := append(entrySource.Commands(tick), movementSource.Commands(tick)...)
		return append(commands, skillSource.Commands(tick)...)
	}
	worldReady := make(chan struct{})
	loading := acceptanceLoadingCoordinatorWithWorld(worldReady)
	defer loading.Close()
	if err := runtime.RegisterInstaller(modruntime.ContentRequire(contentFS, "lua")); err != nil {
		t.Fatal(err)
	}
	for _, module := range []modruntime.Module{
		modruntime.AppModule("test", func() {}),
		modruntime.VFSModule(contentFS),
		modruntime.DataModule(contentFS),
		modruntime.InputModule(&input),
		modruntime.AudioModule(runtime, &mixer, contentFS, gamedata.New(recordstore.New(contentFS))),
		modruntime.SettingsModule(preferences.NewTransient(), &mixer),
		modruntime.VideoModule(runtime, video.Unavailable{}, contentFS),
		modruntime.LocaleModule(localization.New(contentFS, "English")),
		modruntime.RenderModule(runtime, &composer),
		modruntime.SaveModule(saves),
		modruntime.PlayerControlModule(movementController),
		modruntime.NewECSCapability(runtime, entitySimulation).Module(),
		modruntime.LoadingModule(loading),
		scenes.Module(),
	} {
		if err := runtime.RegisterModule(module); err != nil {
			t.Fatal(err)
		}
	}
	if err := runtime.Start(ctx); err != nil {
		t.Fatal(err)
	}
	defer runtime.Stop(ctx)
	components := host.NewManager()
	boot, err := modruntime.LoadDefinition(ctx, runtime, contentFS, "boot.lua")
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
	for range 2 {
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
	positions, found := akara.GetDynamicStore(entitySimulation.World(), "dm.world.position")
	if !found {
		t.Fatal("game world did not register position component")
	}
	players, found := akara.GetDynamicStore(entitySimulation.World(), "dm.world.player_control")
	if !found {
		t.Fatal("game world did not register player-control component")
	}
	heroes, err := entitySimulation.World().Subscribe(akara.All(positions, players))
	if err != nil {
		t.Fatal(err)
	}
	defer heroes.Close()
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
	input.Publish(inputstate.Frame{Actions: map[string]inputstate.ActionState{"right": {Down: true}, "toggle_run": {Pressed: true}}, Owner: inputstate.FocusOwner{Domain: inputstate.FocusScene, ID: "game_world"}})
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
	modes, found := akara.GetDynamicStore(entitySimulation.World(), "dm.player.movement_mode")
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
	if err := movementController.AssignSkill("right", 42); err != nil {
		t.Fatal(err)
	}
	input.Publish(inputstate.Frame{Owner: inputstate.FocusOwner{Domain: inputstate.FocusScene, ID: "game_world"}})
	if _, err := authority.AdvanceWithSource(time.Second, commandSource); err != nil {
		t.Fatal(err)
	}
	assignments, found := akara.GetDynamicStore(entitySimulation.World(), "dm.player.skill_assignment")
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

func publishAction(input *inputstate.Store, name string) {
	input.Publish(inputstate.Frame{Actions: map[string]inputstate.ActionState{name: {Pressed: true}}})
}

func assertStack(t *testing.T, navigator *navigation.Manager, want ...string) {
	t.Helper()
	if got := navigator.Stack(); !reflect.DeepEqual(got, want) {
		t.Fatalf("stack = %v, want %v", got, want)
	}
}

func assertNodes(t *testing.T, composer *render.Composer, want int) {
	t.Helper()
	if got := len(composer.Snapshot()); got != want {
		t.Fatalf("render node count = %d, want %d; nodes=%#v", got, want, composer.Snapshot())
	}
}
