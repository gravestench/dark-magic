package clientapp

import (
	"errors"
	"reflect"
	"testing"

	"github.com/gravestench/dark-magic/internal/content"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	gamesession "github.com/gravestench/dark-magic/internal/game/session"
	"github.com/gravestench/dark-magic/internal/game/simulation"
	gameworld "github.com/gravestench/dark-magic/internal/game/world"
	"github.com/gravestench/dark-magic/internal/mod/d2legacy"
)

// TestSessionInitialDataUsesSelectedDifficulty protects the offline and hosted
// game rules from silently reverting to Normal, which would also change identity.
func TestSessionInitialDataUsesSelectedDifficulty(t *testing.T) {
	t.Parallel()

	for _, difficulty := range []int{0, 1, 2} {
		app := &application{options: Options{GameDifficulty: difficulty}}

		rules, ok := app.sessionInitialData()["d2legacy.game_rules"].(map[string]any)
		if !ok {
			t.Fatal("session initial data has no d2legacy.game_rules map")
		}

		if rules["difficulty"] != difficulty {
			t.Errorf("game_rules difficulty = %v, want %d", rules["difficulty"], difficulty)
		}
	}
}

// TestSessionInitialDataPublishesGeneratedSpawnGeometry keeps portal placement on generated facts.
func TestSessionInitialDataPublishesGeneratedSpawnGeometry(t *testing.T) {
	t.Parallel()

	app := &application{
		gameWorldSpawns: map[int][2]float64{1: {10, 12}, 2: {20, 24}, 3: {30, 36}},
		gameWorlds: map[int]*gameworld.Map{
			1: {WidthSubtiles: 100, HeightSubtiles: 80},
			2: {WidthSubtiles: 200, HeightSubtiles: 160},
		},
	}
	want := map[string]any{
		"1": map[string]any{"x": float64(10), "y": float64(12), "width": float64(100), "height": float64(80)},
		"2": map[string]any{"x": float64(20), "y": float64(24), "width": float64(200), "height": float64(160)},
	}

	got := app.sessionInitialData()["d2legacy.world_spawns"]
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("spawn geometry = %#v, want %#v", got, want)
	}
}

// Topology identity must enter the same initial-data digest used by offline authority and connected peers.
func TestSessionInitialDataIncludesEntryWorldChecksum(t *testing.T) {
	app := &application{entryWorldChecksum: "three-zone-topology"}
	if got := app.sessionInitialData()["d2legacy.entry_world_checksum"]; got != "three-zone-topology" {
		t.Fatalf("entry-world checksum = %v", got)
	}
}

// Omitting the native version must reject old recovery before any target state changes.
func TestSessionInitialDataVelocityRuntimeIdentityRecovery(t *testing.T) {
	data := (&application{}).sessionInitialData()
	if data["engine.world.velocity_movement_version"] != "level-aware-v1" {
		t.Fatal("native movement policy is absent from session configuration")
	}
	current, err := d2legacy.Identity(content.D2Legacy(), data)
	if err != nil {
		t.Fatal(err)
	}
	delete(data, "engine.world.velocity_movement_version")
	old, err := d2legacy.Identity(content.D2Legacy(), data)
	if err != nil {
		t.Fatal(err)
	}
	if old.Recipe.ConfigurationHash == current.Recipe.ConfigurationHash {
		t.Fatal("old movement policy retained configuration identity")
	}
	build := func(identity simulation.RuntimeIdentity) *gamesession.Session {
		s, err := gamesession.New(gameecs.New(), gamesession.Config{})
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() {
			if err := s.Close(); err != nil {
				t.Error(err)
			}
		})
		if err := s.RegisterAuthoritativeRuntime(identity, simulation.NewStateStore()); err != nil {
			t.Fatal(err)
		}
		return s
	}
	source, target := build(current), build(current)
	if err := source.Step(); err != nil {
		t.Fatal(err)
	}
	recovery, err := source.RecoveryCheckpoint()
	if err != nil {
		t.Fatal(err)
	}
	if err := target.RestoreRecoveryCheckpoint(recovery); err != nil {
		t.Fatalf("current identity rejected: %v", err)
	}
	oldSource := build(old)
	oldRecovery, err := oldSource.RecoveryCheckpoint()
	if err != nil {
		t.Fatal(err)
	}
	before, err := target.RecoveryCheckpoint()
	if err != nil {
		t.Fatal(err)
	}
	if err := target.RestoreRecoveryCheckpoint(oldRecovery); !errors.Is(err, simulation.ErrAuthoritativeState) {
		t.Fatalf("old identity recovery = %v", err)
	}
	after, err := target.RecoveryCheckpoint()
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(before, after) {
		t.Fatal("rejected old identity mutated target")
	}
}
