package clientapp

import (
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/gravestench/akara"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	"github.com/gravestench/dark-magic/internal/game/simulation"
	playeradapter "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/player"
)

// Match public identity and position directly to the immutable authoritative ECS snapshot.
func assertStonyPopulationView(t *testing.T, checkpoint simulation.Checkpoint, view playeradapter.WorldView) {
	t.Helper()

	engine, err := gameecs.RestoreSnapshot(*checkpoint.Snapshot)
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = engine.Close() }()

	identities, ok := akara.GetDynamicStore(engine.World(), "d2legacy.monster.identity")
	if !ok {
		t.Fatal("missing authoritative monsters")
	}

	positions, ok := akara.GetDynamicStore(engine.World(), "d2legacy.world.position")
	if !ok {
		t.Fatal("missing authoritative positions")
	}

	count := 0

	for _, visible := range view.Entities {
		if visible.Kind != "hostile" || visible.LevelID != 4 {
			continue
		}
		if !strings.HasPrefix(visible.SpawnID, "level:4:room:") {
			t.Fatalf("unqualified identity %q", visible.SpawnID)
		}
		if visible.DefinitionID != "skeleton1" && visible.DefinitionID != "zombie2" && visible.DefinitionID != "crownest1" {
			t.Fatalf("unexpected family %s", visible.DefinitionID)
		}

		found := false

		for _, entity := range identities.Entities() {
			identity, _ := identities.Get(entity)
			id, _ := identity.Get("spawn_id")
			if id != visible.SpawnID {
				continue
			}

			definition, _ := identity.Get("definition_id")
			x, y := dynamicPosition(positions, entity)
			if definition != visible.DefinitionID || visible.Position.X != x || visible.Position.Y != y {
				t.Fatal("public identity/position differs from authority")
			}

			found = true
			break
		}

		if !found {
			t.Fatal("public monster absent from authority")
		}

		count++
	}

	if count == 0 {
		t.Fatal("Stony Field ordinary population absent from public view")
	}
}

// Project the offline checkpoint through the same public adapter used by connected sessions.
func assertStonyPopulationProjection(t *testing.T, f *realD2LegacyFixture) {
	t.Helper()

	checkpoint, err := f.app.offlineSession.CanonicalCheckpoint()
	if err != nil {
		t.Fatal(err)
	}

	assertStonyPopulationView(t, checkpoint, coldPopulationWorldView(t, "local-player", checkpoint))
}

// Reliable lifecycle and datagram positions must agree with the same acknowledged host tick.
func assertConnectedStonyPopulation(t *testing.T, f *realD2LegacyFixture) map[string]akara.Entity {
	t.Helper()

	for range 100 {
		time.Sleep(40 * time.Millisecond)
		f.advanceGame(t, 1)

		presentation := f.app.network.client.PresentationSnapshot()
		if presentation.Events.Tick != presentation.HUD.Tick || presentation.World.Tick != presentation.HUD.Tick {
			continue
		}

		replay, err := f.app.network.host.Session.Replay()
		if err != nil {
			t.Fatal(err)
		}

		for _, checkpoint := range replay.Checkpoints {
			if checkpoint.Tick != presentation.HUD.Tick {
				continue
			}

			want := coldPopulationWorldView(t, f.app.presentationPlayerID(), checkpoint)
			if !reflect.DeepEqual(publicColdPopulation(want), publicColdPopulation(presentation.World)) {
				t.Fatal("connected population differs from acknowledged authority")
			}

			assertStonyPopulationView(t, checkpoint, presentation.World)
			return checkpointPopulationIDs(t, checkpoint)
		}
	}

	t.Fatal("no acknowledged population checkpoint")
	return nil
}
