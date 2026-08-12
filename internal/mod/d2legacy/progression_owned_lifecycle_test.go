package d2legacy

import (
	"context"
	"testing"

	"github.com/gravestench/akara"
	"github.com/gravestench/dark-magic/internal/content"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	gamesession "github.com/gravestench/dark-magic/internal/game/session"
	"github.com/gravestench/dark-magic/internal/game/simulation"
	lua "github.com/yuin/gopher-lua"
)

// TestCharacterLevelAndOwnedUnitExpirationRestoreIdentically closes M21.14.10
// with production Lua composition on both sides of a checkpoint boundary.
func TestCharacterLevelAndOwnedUnitExpirationRestoreIdentically(t *testing.T) {
	ctx := context.Background()
	records := fixtureRecords{}
	engine := gameecs.New()
	session, err := gamesession.New(engine, gamesession.Config{CheckpointInterval: 1})
	if err != nil {
		t.Fatal(err)
	}
	authority, err := StartWithConfig(ctx, content.D2Legacy(), records, engine, session, Config{Seed: 42})
	if err != nil {
		t.Fatal(err)
	}

	payload := []byte(`{"character_id":"hero","player":"alice","name":"Hero","class":"Amazon",` +
		`"level":1,"experience":5,"dexterity":20,"defense":0,"health":50,"max_health":50,` +
		`"mana":20,"max_mana":20,"expansion":true,"hardcore":false,"cof":"","palette":"units",` +
		`"direction":0,"mode":"NU","x":10,"y":12,"world_width":100,"world_height":80,` +
		`"act":1,"level_id":1,"skills":[]}`)
	if err := session.Submit(simulation.Command{Tick: 1, Player: "system", Authority: simulation.AuthoritySystem,
		Sequence: 1, Kind: "system.player.enter", Payload: payload}); err != nil {
		t.Fatal(err)
	}
	if err := session.Step(); err != nil {
		t.Fatal(err)
	}
	// Materialize a temporary owned unit using the production attach command.
	err = authority.Runtime.Run(ctx, func(state *lua.LState) error {
		return state.DoString(`
local ecs=require("engine.ecs/v1")
ecs.create({["d2legacy.world.selectable"]={id="monster:wolf",kind="friendly",label="Wolf",owner="alice",radius=1,priority=1}})
`)
	})
	if err != nil {
		t.Fatal(err)
	}
	attach := []byte(`{"unit_id":"monster:wolf","owner_id":"player:alice","ultimate_owner_id":"player:alice",` +
		`"expires_tick":3,"category":{"id":"wolf","group":1,"base_max":1,"replacement":"replace_oldest"}}`)
	if err := session.Submit(simulation.Command{Tick: 2, Player: "system", Authority: simulation.AuthoritySystem,
		Sequence: 2, Kind: "system.owned_unit.attach", Payload: attach}); err != nil {
		t.Fatal(err)
	}
	if err := session.Step(); err != nil {
		t.Fatal(err)
	}
	replay, err := session.Replay()
	if err != nil {
		t.Fatal(err)
	}
	checkpoint := replay.Checkpoints[len(replay.Checkpoints)-1]
	if err := session.Step(); err != nil {
		t.Fatal(err)
	}
	originalReplay, _ := session.Replay()
	original := originalReplay.Checkpoints[len(originalReplay.Checkpoints)-1]
	assertLevelAndExpiredUnit(t, engine)
	_ = authority.Stop(ctx)
	_ = session.Close()
	_ = engine.Close()

	restoredEngine, err := gameecs.RestoreSnapshot(*checkpoint.Snapshot)
	if err != nil {
		t.Fatal(err)
	}
	defer restoredEngine.Close()
	restoredSession, err := gamesession.New(restoredEngine, gamesession.Config{CheckpointInterval: 1})
	if err != nil {
		t.Fatal(err)
	}
	defer restoredSession.Close()
	restored, err := StartWithConfig(ctx, content.D2Legacy(), records, restoredEngine, restoredSession,
		Config{Seed: 42, Restore: checkpoint.Participants})
	if err != nil {
		t.Fatal(err)
	}
	defer restored.Stop(ctx)
	if err := restoredSession.Step(); err != nil {
		t.Fatal(err)
	}
	continuedReplay, _ := restoredSession.Replay()
	continued := continuedReplay.Checkpoints[len(continuedReplay.Checkpoints)-1]
	if continued.Checksum != original.Checksum {
		t.Fatalf("restored progression/lifecycle checksum = %s, want %s", continued.Checksum, original.Checksum)
	}
	assertLevelAndExpiredUnit(t, restoredEngine)
}

func assertLevelAndExpiredUnit(t *testing.T, engine *gameecs.Engine) {
	t.Helper()
	progress, _ := akara.GetDynamicStore(engine.World(), "d2legacy.player.progress")
	value, _ := progress.Get(progress.Entities()[0])
	level, _ := value.Get("level")
	if level != int64(2) {
		t.Fatalf("player level = %v, want 2", level)
	}
	owned, _ := akara.GetDynamicStore(engine.World(), "d2legacy.owned_unit")
	if owned.Len() != 0 {
		t.Fatalf("owned units after expiration = %d, want 0", owned.Len())
	}
}
