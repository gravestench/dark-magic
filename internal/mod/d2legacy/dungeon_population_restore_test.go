package d2legacy

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/gravestench/dark-magic/internal/content"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	gamesession "github.com/gravestench/dark-magic/internal/game/session"
	"github.com/gravestench/dark-magic/internal/game/simulation"
)

func TestDungeonCombatAndBothAreaResidentsSurviveReturnAndCheckpoint(t *testing.T) {
	records := generatedHostileRecords()
	records["data/global/excel/levels.txt"] = append(records["data/global/excel/levels.txt"], map[string]string{"Id": "8", "MonDen": "100000", "NumMon": "1", "mon1": "fallen"})
	records["data/global/excel/monstats.txt"][0]["minHP"] = "12"
	records["data/global/excel/monstats.txt"][0]["maxHP"] = "12"
	initial := map[string]any{"d2legacy.world_transitions": map[string]any{"seams": []any{
		map[string]any{"source_level": 2, "destination_level": 8, "source_x": 9, "source_y": 1, "arrival_x": 1, "arrival_y": 1, "world_width": 100, "world_height": 100},
		map[string]any{"source_level": 8, "destination_level": 2, "source_x": 1, "source_y": 9, "arrival_x": 1, "arrival_y": 1, "world_width": 100, "world_height": 100},
	}}}
	start := func(checkpoint *simulation.Checkpoint) (*Authority, *gameecs.Engine, *gamesession.Session) {
		t.Helper()
		engine := gameecs.New()
		var restore []simulation.ParticipantState
		if checkpoint != nil {
			_ = engine.Close()
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
		authority, err := StartWithConfig(t.Context(), content.D2Legacy(), records, engine, session, Config{Seed: 314, InitialData: initial, Restore: restore})
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = authority.Stop(t.Context()); _ = session.Close(); _ = engine.Close() })
		return authority, engine, session
	}
	authority, engine, session := start(nil)
	sequences := map[*gamesession.Session]uint64{}
	submit := func(session *gamesession.Session, engine *gameecs.Engine, kind string, payload any, system bool) {
		t.Helper()
		encoded, err := json.Marshal(payload)
		if err != nil {
			t.Fatal(err)
		}
		sequence := sequences[session] + 1
		command := simulation.Command{Tick: engine.Tick() + 1, Player: "alice", Authority: simulation.AuthorityPlayer, Sequence: sequence, Kind: kind, Payload: encoded}
		if system {
			command.Player, command.Authority = "system", simulation.AuthoritySystem
		}
		if system {
			command.Sequence = engine.Tick() + 1
		} else {
			sequences[session] = sequence
		}
		if err := session.Submit(command); err != nil {
			t.Fatal(err)
		}
	}
	area := func(level int) map[string]any {
		return map[string]any{"act": 1, "level_id": level, "difficulty": 0, "links": []any{}, "rooms": []any{
			map[string]any{"id": "a", "populate": true, "x": 0, "y": 0, "width": 30, "height": 30, "points": []any{map[string]any{"x": 4, "y": 1, "clearance": 2}}, "eligible_tiles": 100},
			map[string]any{"id": "unvisited", "populate": true, "x": 60, "y": 0, "width": 30, "height": 30, "points": []any{map[string]any{"x": 64, "y": 1}}},
		}}
	}
	population := area(2)
	population["rooms"].([]any)[0].(map[string]any)["points"] = []any{map[string]any{"x": 20, "y": 1}}
	population["additional_zones"] = []any{area(8)}
	submit(session, engine, "system.population.bootstrap", population, true)
	stepSession(t, session, 1)
	submit(session, engine, "system.player.enter", json.RawMessage(generatedPlayerPayload(t, "hero", "alice", 1, 1)), true)
	stepSession(t, session, 1)
	moor := monsterEntity(t, engine, "level:2:room:a:monster:1")
	moorStats := componentSnapshot(t, engine, moor, "d2legacy.monster.stats")
	travel := func(session *gamesession.Session, engine *gameecs.Engine, level int) {
		t.Helper()
		x, y := 1, 0
		if level == 2 {
			x, y = 0, 1
		}
		submit(session, engine, "player.move", map[string]any{"x": x, "y": y, "running": false}, false)
		for range 100 {
			stepSession(t, session, 1)
			if componentSnapshot(t, engine, playerEntity(t, engine, "alice"), "d2legacy.world.location")["level_id"] == int64(level) {
				stepSession(t, session, 1)
				return
			}
		}
		t.Fatalf("did not walk to level %d: position=%v location=%v", level, componentSnapshot(t, engine, playerEntity(t, engine, "alice"), "d2legacy.world.position"), componentSnapshot(t, engine, playerEntity(t, engine, "alice"), "d2legacy.world.location"))
	}
	travel(session, engine, 8)
	const denID = "level:8:room:a:monster:1"
	den := monsterEntity(t, engine, denID)
	initialHealth := componentSnapshot(t, engine, den, "d2legacy.monster.stats")["health"].(int64)
	assertResidentActivation(t, engine, moor, false, true)
	attack := func(session *gamesession.Session, engine *gameecs.Engine) {
		t.Helper()
		submit(session, engine, "player.use_skill", map[string]any{"side": "left", "target_id": "monster:" + denID, "target_x": 4, "target_y": 1}, false)
		stepSession(t, session, 1)
	}
	for range 100 {
		attack(session, engine)
		if componentSnapshot(t, engine, den, "d2legacy.monster.stats")["health"].(int64) < initialHealth {
			break
		}
	}
	wounded := componentSnapshot(t, engine, den, "d2legacy.monster.stats")["health"].(int64)
	if wounded <= 0 || wounded >= initialHealth {
		t.Fatalf("normal attack did not leave a wounded resident: %d", wounded)
	}
	replay, err := session.Replay()
	if err != nil {
		t.Fatal(err)
	}
	checkpoint := replay.Checkpoints[len(replay.Checkpoints)-1]
	restored, restoredEngine, restoredSession := start(&checkpoint)
	finish := func(authority *Authority, engine *gameecs.Engine, session *gamesession.Session) {
		t.Helper()
		den := monsterEntity(t, engine, denID)
		travel(session, engine, 2)
		assertResidentActivation(t, engine, den, false, true)
		inactiveHealth := componentSnapshot(t, engine, den, "d2legacy.monster.stats")["health"].(int64)
		if inactiveHealth <= 0 || inactiveHealth > wounded {
			t.Fatalf("wound lost across departure: %v", inactiveHealth)
		}
		stepSession(t, session, 5)
		if got := componentSnapshot(t, engine, den, "d2legacy.monster.stats")["health"]; got != inactiveHealth {
			t.Fatalf("inactive resident health changed: %v", got)
		}
		if got := componentSnapshot(t, engine, monsterEntity(t, engine, "level:2:room:a:monster:1"), "d2legacy.monster.stats"); !reflect.DeepEqual(got, moorStats) {
			t.Fatal("Den activation reset Moor")
		}
		travel(session, engine, 8)
		for range 200 {
			attack(session, engine)
			if componentSnapshot(t, engine, den, "d2legacy.monster.stats")["health"].(int64) <= 0 {
				break
			}
		}
		death := componentSnapshot(t, engine, den, "d2legacy.monster.death")
		if death["credited_id"] != "player:alice" {
			t.Fatal("normal death lost player credit")
		}
		progress := componentSnapshot(t, engine, playerEntity(t, engine, "alice"), "d2legacy.player.progress")
		if progress["experience"] != int64(5) {
			t.Fatalf("XP=%v, want 5", progress["experience"])
		}
		travel(session, engine, 2)
		assertCorpseActivation(t, engine, den, false)
		travel(session, engine, 8)
		assertCorpseActivation(t, engine, den, true)
		stepSession(t, session, 5)
		if got := componentSnapshot(t, engine, den, "d2legacy.monster.death"); !reflect.DeepEqual(got, death) {
			t.Fatal("death committed again")
		}
		if got := componentSnapshot(t, engine, playerEntity(t, engine, "alice"), "d2legacy.player.progress"); !reflect.DeepEqual(got, progress) {
			t.Fatal("XP repeated")
		}
		questKills := 0
		events := dynamicStore(t, engine, "d2legacy.monster.death_event")
		for _, entity := range events.Entities() {
			event := componentSnapshot(t, engine, entity, "d2legacy.monster.death_event")
			if event["kind"] == "monster_quest_kill" && event["monster_id"] == denID {
				questKills++
			}
		}
		if questKills != 1 {
			t.Fatalf("Den quest-kill facts=%d, want 1", questKills)
		}
		resident := componentSnapshot(t, engine, den, "d2legacy.world.room_resident")
		if resident["id"] != denID || resident["level_id"] != int64(8) || resident["room_id"] != "a" {
			t.Fatal("canonical death/resident mapping lost")
		}
		registered, found := authority.State.Read("d2legacy.population.plan")
		if !found {
			t.Fatal("population state missing")
		}
		var plan struct {
			Installed  bool `json:"installed"`
			Additional []struct {
				Level      int `json:"level_id"`
				Difficulty int `json:"difficulty"`
				Rooms      []struct {
					ID        string `json:"id"`
					Activated bool   `json:"activated"`
				} `json:"rooms"`
			} `json:"additional_zones"`
		}
		if err := json.Unmarshal(registered.Data, &plan); err != nil {
			t.Fatal(err)
		}
		if !plan.Installed || len(plan.Additional) != 1 || plan.Additional[0].Level != 8 || plan.Additional[0].Difficulty != 0 || len(plan.Additional[0].Rooms) != 2 {
			t.Fatal("complete Den domain lost")
		}
		if plan.Additional[0].Rooms[1].Activated {
			t.Fatal("unvisited room falsely settled")
		}
	}
	finish(authority, engine, session)
	finish(restored, restoredEngine, restoredSession)
	assertEqualLatestChecksums(t, session, restoredSession, "multi-area midfight continuation")
}
