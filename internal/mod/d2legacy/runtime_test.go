package d2legacy

import (
	"context"
	"reflect"
	"testing"

	"github.com/gravestench/dark-magic/internal/content"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	gamesession "github.com/gravestench/dark-magic/internal/game/session"
	"github.com/gravestench/dark-magic/internal/game/simulation"
)

type fixtureRecords struct{}

func (fixtureRecords) Invalidate(string)  {}
func (fixtureRecords) Loaded(string) bool { return true }
func (fixtureRecords) Load(path string) ([]map[string]string, error) {
	if path == "data/global/excel/skills.txt" {
		return []map[string]string{{"Id": "36", "skill": "Fire Bolt", "srvmissile": "firebolt", "etype": "fire", "interrupt": "1", "srvstfunc": "", "srvdofunc": "", "mana": "5", "manashift": "7", "emin": "3", "emax": "6", "HitShift": "8"}}, nil
	}
	return []map[string]string{{"Missile": "firebolt", "Skill": "Fire Bolt", "pSrvDoFunc": "1", "CollideType": "3", "CollideKill": "1", "Vel": "20", "Range": "40", "Size": "2", "CelFile": "firebolt", "AnimSpeed": "16", "NumDirections": "16", "LoopAnim": "1"}}, nil
}

func TestAuthorityRestoresAllDeterministicParticipantsBeforeFirstTick(t *testing.T) {
	ctx := context.Background()
	engine := gameecs.New()
	session, err := gamesession.New(engine, gamesession.Config{CheckpointInterval: 1})
	if err != nil {
		t.Fatal(err)
	}
	authority, err := Start(ctx, content.D2Legacy(), fixtureRecords{}, engine, session, 7)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := authority.Random.Uint64n("d2legacy.combat.fire_bolt.damage", 100); err != nil {
		t.Fatal(err)
	}
	if err := session.Step(); err != nil {
		t.Fatal(err)
	}
	replay, err := session.Replay()
	if err != nil {
		t.Fatal(err)
	}
	if err := authority.Stop(ctx); err != nil {
		t.Fatal(err)
	}
	session.Close()

	checkpoint := replay.Checkpoints[0]
	restoredEngine, err := gameecs.RestoreSnapshot(*checkpoint.Snapshot)
	if err != nil {
		t.Fatal(err)
	}
	restoredSession, err := gamesession.New(restoredEngine, gamesession.Config{CheckpointInterval: 1})
	if err != nil {
		t.Fatal(err)
	}
	defer restoredSession.Close()
	restored, err := StartWithConfig(ctx, content.D2Legacy(), fixtureRecords{}, restoredEngine, restoredSession, Config{Seed: 7, Restore: checkpoint.Participants})
	if err != nil {
		t.Fatal(err)
	}
	defer restored.Stop(ctx)
	restoredReplay, err := restoredSession.Replay()
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(checkpoint.Participants, restoredReplay.InitialParticipants) {
		t.Fatalf("restored participants differ\nwant: %#v\n got: %#v", checkpoint.Participants, restoredReplay.InitialParticipants)
	}
	assertParticipantIDs(t, restoredReplay.InitialParticipants,
		"engine.authoritative_rng/v1", "engine.authoritative_runtime/v1", "engine.authoritative_state/v1")
}

func assertParticipantIDs(t *testing.T, states []simulation.ParticipantState, expected ...string) {
	t.Helper()
	if len(states) != len(expected) {
		t.Fatalf("participant count = %d, want %d", len(states), len(expected))
	}
	for index, id := range expected {
		if states[index].ID != id {
			t.Fatalf("participant %d = %q, want %q", index, states[index].ID, id)
		}
	}
}

func TestAuthorityBootsWithoutClientOrRenderer(t *testing.T) {
	engine := gameecs.New()
	session, err := gamesession.New(engine, gamesession.Config{})
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()
	authority, err := Start(context.Background(), content.D2Legacy(), fixtureRecords{}, engine, session, 7)
	if err != nil {
		t.Fatal(err)
	}
	if err := authority.Stop(context.Background()); err != nil {
		t.Fatal(err)
	}
}
