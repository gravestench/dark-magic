package d2legacy

import (
	"context"
	"testing"

	"github.com/gravestench/dark-magic/internal/content"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	gamesession "github.com/gravestench/dark-magic/internal/game/session"
	"github.com/gravestench/dark-magic/internal/game/simulation"
)

func goldInitialData(amount any) map[string]any {
	return map[string]any{"d2legacy.interactions": map[string]any{
		"owner": "alice", "targets": []any{map[string]any{
			"id": "gold:fixture", "x": 10.0, "y": 12.0, "radius": 4.0,
			"act": 1, "level_id": 2, "selectable": true, "gold_amount": amount,
		}},
	}}
}

func startGoldFixture(t *testing.T, checkpoint *simulation.Checkpoint) (*gameecs.Engine, *gamesession.Session) {
	t.Helper()
	var engine *gameecs.Engine
	var restore []simulation.ParticipantState
	if checkpoint == nil {
		engine = gameecs.New()
	} else {
		var err error
		engine, err = gameecs.RestoreSnapshot(*checkpoint.Snapshot)
		if err != nil {
			t.Fatal(err)
		}
		restore = checkpoint.Participants
	}
	session, err := gamesession.New(engine, gamesession.Config{CheckpointInterval: 1})
	if err != nil {
		t.Fatal(err)
	}
	authority, err := StartWithConfig(t.Context(), content.D2Legacy(), fixtureRecords{}, engine, session,
		Config{Seed: 374, InitialData: goldInitialData(1000), Restore: restore})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = authority.Stop(context.Background())
		_ = session.Close()
		_ = engine.Close()
	})
	return engine, session
}

func assertGoldFixture(t *testing.T, engine *gameecs.Engine, amount, aliceGold int64, inactive bool) {
	t.Helper()
	target := interactionTargetEntity(t, engine, "gold:fixture")
	pile := componentSnapshot(t, engine, target, "d2legacy.gold.pile")
	if pile["amount"] != amount {
		t.Errorf("gold pile amount = %v, want %d", pile["amount"], amount)
	}
	inactiveStore := dynamicStore(t, engine, "d2legacy.world.inactive")
	_, found := inactiveStore.Get(target)
	if found != inactive {
		t.Errorf("gold pile inactive = %t, want %t", found, inactive)
	}
	layouts := dynamicStore(t, engine, "d2legacy.items.layout")
	for _, entity := range layouts.Entities() {
		value, _ := layouts.Get(entity)
		owner, _ := value.Get("owner")
		if owner == "alice" {
			gold, _ := value.Get("carried_gold")
			if gold != aliceGold {
				t.Errorf("alice carried gold = %v, want %d", gold, aliceGold)
			}
			return
		}
	}
	t.Fatal("alice gold layout missing")
}

func TestPartyGoldPickupRestoresBeforeAndAfterConsumption(t *testing.T) {
	engine, session := startGoldFixture(t, nil)
	if err := session.Submit(simulation.Command{
		Tick: 1, Player: "system", Authority: simulation.AuthoritySystem, Sequence: 1,
		Kind: "system.player.enter", Payload: generatedPlayerPayload(t, "hero-alice", "alice", 10, 12),
	}); err != nil {
		t.Fatal(err)
	}
	stepPartySession(t, session)
	assertGoldFixture(t, engine, 1000, 0, false)
	beforeReplay, err := session.Replay()
	if err != nil {
		t.Fatal(err)
	}
	before := beforeReplay.Checkpoints[len(beforeReplay.Checkpoints)-1]

	submitPartyCommand(t, session, 2, "alice", 1, "interaction.open", map[string]any{"target": "gold:fixture"})
	stepPartySession(t, session)
	assertGoldFixture(t, engine, 0, 1000, true)
	afterReplay, err := session.Replay()
	if err != nil {
		t.Fatal(err)
	}
	after := afterReplay.Checkpoints[len(afterReplay.Checkpoints)-1]

	replayedEngine, replayedSession := startGoldFixture(t, &before)
	submitPartyCommand(t, replayedSession, 2, "alice", 1, "interaction.open", map[string]any{"target": "gold:fixture"})
	stepPartySession(t, replayedSession)
	assertGoldFixture(t, replayedEngine, 0, 1000, true)
	replayed, err := replayedSession.Replay()
	if err != nil {
		t.Fatal(err)
	}
	if got := replayed.Checkpoints[len(replayed.Checkpoints)-1].Checksum; got != after.Checksum {
		t.Errorf("replayed pickup checksum = %s, want %s", got, after.Checksum)
	}

	restoredEngine, restoredSession := startGoldFixture(t, &after)
	assertGoldFixture(t, restoredEngine, 0, 1000, true)
	if err := restoredSession.Submit(simulation.Command{
		Tick: 3, Player: "system", Authority: simulation.AuthoritySystem, Sequence: 1,
		Kind: "system.player.enter", Payload: generatedPlayerPayload(t, "hero-bob", "bob", 30, 30),
	}); err != nil {
		t.Fatal(err)
	}
	stepPartySession(t, restoredSession)
	assertGoldFixture(t, restoredEngine, 0, 1000, true)
}

func TestPartyGoldRejectsMalformedTrustedPiles(t *testing.T) {
	for _, scenario := range []struct {
		name   string
		change func(map[string]any)
	}{
		{"zero amount", func(target map[string]any) { target["gold_amount"] = 0 }},
		{"oversized amount", func(target map[string]any) { target["gold_amount"] = int64(2147483648) }},
		{"fractional amount", func(target map[string]any) { target["gold_amount"] = 1.5 }},
		{"missing position", func(target map[string]any) { delete(target, "x") }},
		{"missing act", func(target map[string]any) { delete(target, "act") }},
		{"missing level", func(target map[string]any) { delete(target, "level_id") }},
		{"missing radius", func(target map[string]any) { delete(target, "radius") }},
		{"not selectable", func(target map[string]any) { target["selectable"] = false }},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			initial := goldInitialData(1000)
			initial["d2legacy.items"] = map[string]any{"owner": "alice", "items": []any{}}
			interactions := initial["d2legacy.interactions"].(map[string]any)
			target := interactions["targets"].([]any)[0].(map[string]any)
			scenario.change(target)
			engine := gameecs.New()
			defer func() { _ = engine.Close() }()
			session, err := gamesession.New(engine, gamesession.Config{CheckpointInterval: 1})
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = session.Close() }()
			authority, err := StartWithConfig(t.Context(), content.D2Legacy(), fixtureRecords{}, engine, session,
				Config{Seed: 374, InitialData: initial})
			if err == nil {
				_ = authority.Stop(context.Background())
				t.Fatal("malformed trusted gold pile was accepted")
			}
		})
	}
}
