package d2legacy

import (
	"encoding/json"
	"testing"

	"github.com/gravestench/dark-magic/internal/content"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	gamesession "github.com/gravestench/dark-magic/internal/game/session"
	"github.com/gravestench/dark-magic/internal/game/simulation"
	playeradapter "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/player"
)

type denQuestFixture struct {
	engine    *gameecs.Engine
	session   *gamesession.Session
	sequences map[string]uint64
}

// The small synthetic geometry exercises the production population, travel,
// combat, death and interaction owners without administrative quest completion.
func startDenQuest(t *testing.T, checkpoint *simulation.Checkpoint) *denQuestFixture {
	t.Helper()
	records := generatedHostileRecords()
	records["data/global/excel/levels.txt"] = append(records["data/global/excel/levels.txt"], map[string]string{"Id": "8", "MonDen": "100000", "NumMon": "1", "mon1": "fallen"})
	initial := map[string]any{"d2legacy.interactions": map[string]any{"owner": "alice", "targets": []any{
		map[string]any{"id": "npc:akara", "npc": "akara", "x": 1, "y": 1, "radius": 4, "act": 1, "level_id": 1, "selectable": true},
	}}, "d2legacy.world_transitions": map[string]any{"seams": []any{
		map[string]any{"source_level": 1, "destination_level": 8, "source_x": 9, "source_y": 1, "arrival_x": 1, "arrival_y": 1, "world_width": 100, "world_height": 100},
		map[string]any{"source_level": 8, "destination_level": 1, "source_x": 1, "source_y": 9, "arrival_x": 1, "arrival_y": 1, "world_width": 100, "world_height": 100},
	}}}
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
	return &denQuestFixture{engine: engine, session: session, sequences: map[string]uint64{}}
}

func (f *denQuestFixture) command(t *testing.T, kind string, payload any, owner string) {
	t.Helper()
	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	f.sequences[owner]++
	command := simulation.Command{Tick: f.engine.Tick() + 1, Player: owner, Authority: simulation.AuthorityPlayer, Sequence: f.sequences[owner], Kind: kind, Payload: data}
	if owner == "system" {
		command.Authority = simulation.AuthoritySystem
	}
	if err := f.session.Submit(command); err != nil {
		t.Fatal(err)
	}
	stepSession(t, f.session, 1)
}

func (f *denQuestFixture) checkpoint(t *testing.T) simulation.Checkpoint {
	t.Helper()
	replay, err := f.session.Replay()
	if err != nil {
		t.Fatal(err)
	}
	return replay.Checkpoints[len(replay.Checkpoints)-1]
}

func (f *denQuestFixture) enter(t *testing.T, owner string) {
	t.Helper()
	var payload map[string]any
	if err := json.Unmarshal(generatedPlayerPayload(t, owner, owner, 1, 1), &payload); err != nil {
		t.Fatal(err)
	}
	payload["level_id"] = 1
	if owner != "alice" {
		payload["x"], payload["y"] = 80, 80
	}
	f.command(t, "system.player.enter", payload, "system")
}

func (f *denQuestFixture) move(t *testing.T, x, y float64) {
	t.Helper()
	f.command(t, "player.move", map[string]any{"x": 0, "y": 0, "running": false, "target": map[string]any{"x": x, "y": y}}, "alice")
	stepSession(t, f.session, 80)
}

func (f *denQuestFixture) quest(t *testing.T, remaining, unsettled int64, complete, rewarded bool) {
	t.Helper()
	q := componentSnapshot(t, f.engine, playerEntity(t, f.engine, "alice"), "d2legacy.quest.progress")
	if q["remaining"] != remaining || q["unsettled_rooms"] != unsettled || q["completed"] != complete || q["rewarded"] != rewarded || q["population_available"] != true {
		t.Fatalf("quest=%v, want remaining=%d unsettled=%d complete=%v rewarded=%v", q, remaining, unsettled, complete, rewarded)
	}
}

func (f *denQuestFixture) killActive(t *testing.T) {
	t.Helper()
	identities := dynamicStore(t, f.engine, "d2legacy.monster.identity")
	for _, entity := range identities.Entities() {
		if _, inactive := dynamicStore(t, f.engine, "d2legacy.world.inactive").Get(entity); inactive {
			continue
		}
		id := componentSnapshot(t, f.engine, entity, "d2legacy.monster.identity")["spawn_id"].(string)
		for range 100 {
			if _, dead := dynamicStore(t, f.engine, "d2legacy.monster.death").Get(entity); dead {
				break
			}
			p := componentSnapshot(t, f.engine, entity, "d2legacy.world.position")
			f.command(t, "player.use_skill", map[string]any{"side": "left", "target_id": "monster:" + id, "target_x": p["x"], "target_y": p["y"]}, "alice")
			stepSession(t, f.session, 12)
		}
		if _, dead := dynamicStore(t, f.engine, "d2legacy.monster.death").Get(entity); !dead {
			for _, event := range dynamicStore(t, f.engine, "d2legacy.combat.event").Entities() {
				t.Log(componentSnapshot(t, f.engine, event, "d2legacy.combat.event"))
			}
			t.Fatal("ordinary attack did not kill resident", id,
				componentSnapshot(t, f.engine, entity, "d2legacy.monster.stats"), componentSnapshot(t, f.engine, entity, "d2legacy.world.position"),
				componentSnapshot(t, f.engine, playerEntity(t, f.engine, "alice"), "d2legacy.world.position"), componentSnapshot(t, f.engine, playerEntity(t, f.engine, "alice"), "d2legacy.player.vitals"))
		}
	}
}

func TestDenQuestNaturalClearReturnRewardAndCheckpoint(t *testing.T) {
	f := startDenQuest(t, nil)
	area := map[string]any{"act": 1, "level_id": 8, "difficulty": 0, "links": []any{}, "rooms": []any{
		map[string]any{"id": "entrance", "populate": true, "x": 0, "y": 0, "width": 20, "height": 20, "points": []any{map[string]any{"x": 4, "y": 1}}},
		map[string]any{"id": "far", "populate": true, "x": 30, "y": 0, "width": 20, "height": 20, "points": []any{map[string]any{"x": 34, "y": 1}}},
	}}
	f.command(t, "system.population.bootstrap", map[string]any{"act": 1, "level_id": 2, "difficulty": 0, "rooms": []any{}, "links": []any{}, "additional_zones": []any{area}}, "system")
	f.enter(t, "alice")
	f.enter(t, "bob") // Town-only bystander must not acquire quest credit.
	f.move(t, 9, 1)
	if got := componentSnapshot(t, f.engine, playerEntity(t, f.engine, "alice"), "d2legacy.world.location")["level_id"]; got != int64(8) {
		t.Fatal("ordinary entry failed", got)
	}
	f.quest(t, 1, 1, false, false)
	f.killActive(t)
	f.quest(t, 0, 1, false, false)
	f.move(t, 31, 1)
	f.quest(t, 1, 0, false, false)
	f.move(t, 1, 9)
	f.quest(t, 1, 0, false, false) // An unloaded area still owns a living resident.
	// A duplicated death notification is not a second kill or proof that the
	// still-living inactive resident has died.
	events := dynamicStore(t, f.engine, "d2legacy.monster.death_event")
	for _, entity := range events.Entities() {
		values := componentSnapshot(t, f.engine, entity, "d2legacy.monster.death_event")
		if values["kind"] == "monster_quest_kill" {
			duplicate, err := f.engine.World().CreateEntity()
			if err != nil {
				t.Fatal(err)
			}
			if _, err := events.Set(duplicate, values); err != nil {
				t.Fatal(err)
			}
			break
		}
	}
	stepSession(t, f.session, 1)
	f.quest(t, 1, 0, false, false)
	beforeKill := f.checkpoint(t)
	restored := startDenQuest(t, &beforeKill)
	finish := func(f *denQuestFixture) {
		f.move(t, 9, 1)
		f.move(t, 31, 1)
		f.killActive(t)
		f.quest(t, 0, 0, true, false)
		stepSession(t, f.session, 3) // Reobserving retained death facts is idempotent.
		if _, found := dynamicStore(t, f.engine, "d2legacy.quest.progress").Get(playerEntity(t, f.engine, "bob")); found {
			t.Fatal("town bystander received Den credit")
		}
		f.move(t, 1, 9)
	}
	finish(f)
	finish(restored)
	assertEqualLatestChecksums(t, f.session, restored.session, "Den final kill and walking return")
	beforeClaim := f.checkpoint(t)
	claimRestored := startDenQuest(t, &beforeClaim)
	claim := func(f *denQuestFixture) {
		before := componentSnapshot(t, f.engine, playerEntity(t, f.engine, "alice"), "d2legacy.player.progress")["unspent_skill_points"].(int64)
		f.command(t, "interaction.open", map[string]any{"target": "npc:akara"}, "alice")
		f.quest(t, 0, 0, true, true)
		f.command(t, "interaction.open", map[string]any{"target": "npc:akara"}, "alice")
		f.command(t, "quest.claim_reward", map[string]any{}, "alice")
		if got := componentSnapshot(t, f.engine, playerEntity(t, f.engine, "alice"), "d2legacy.player.progress")["unspent_skill_points"]; got != before+1 {
			t.Fatal("reward delta is not exactly one", got, before)
		}
	}
	claim(f)
	claim(claimRestored)
	assertEqualLatestChecksums(t, f.session, claimRestored.session, "Den pending reward continuation")
	afterClaim := f.checkpoint(t)
	spent := startDenQuest(t, &afterClaim)
	for _, candidate := range []*denQuestFixture{f, spent} {
		candidate.command(t, "quest.claim_reward", map[string]any{}, "alice")
		candidate.quest(t, 0, 0, true, true)
	}
	assertEqualLatestChecksums(t, f.session, spent.session, "Den spent reward continuation")
	view, err := playeradapter.ProjectPrivateView("alice", f.checkpoint(t))
	if err != nil || view.Quest == nil || !view.Quest.Rewarded {
		t.Fatal("owner quest projection", view.Quest, err)
	}
	other, err := playeradapter.ProjectPrivateView("bob", f.checkpoint(t))
	if err != nil || other.Quest != nil {
		t.Fatal("private progress leaked", other.Quest, err)
	}
	// A first entry after clear does not establish participation.
	f.enter(t, "late")
	late := playerEntity(t, f.engine, "late")
	locations := dynamicStore(t, f.engine, "d2legacy.world.location")
	location, _ := locations.Get(late)
	if err := location.Set("level_id", int64(8)); err != nil {
		t.Fatal(err)
	}
	stepSession(t, f.session, 1)
	if _, found := dynamicStore(t, f.engine, "d2legacy.quest.progress").Get(late); found {
		t.Fatal("post-clear first entrant received credit")
	}
}

func TestDenQuestSettledEmptyRoomAndMissingResident(t *testing.T) {
	f := startDenQuest(t, nil)
	f.command(t, "system.population.bootstrap", map[string]any{"act": 1, "level_id": 8, "difficulty": 0, "links": []any{}, "rooms": []any{
		map[string]any{"id": "live", "populate": true, "x": 0, "y": 0, "width": 20, "height": 20, "points": []any{map[string]any{"x": 4, "y": 1}}},
		map[string]any{"id": "empty", "populate": false, "x": 30, "y": 0, "width": 20, "height": 20, "points": []any{}},
	}}, "system")
	f.enter(t, "alice")
	f.move(t, 9, 1)
	f.killActive(t)
	f.quest(t, 0, 1, false, false)
	checkpoint := f.checkpoint(t)
	missing := startDenQuest(t, &checkpoint)
	store := dynamicStore(t, missing.engine, "d2legacy.world.room_resident")
	for _, entity := range dynamicStore(t, missing.engine, "d2legacy.monster.identity").Entities() {
		store.Remove(entity)
	}
	missing.move(t, 31, 1)
	q := componentSnapshot(t, missing.engine, playerEntity(t, missing.engine, "alice"), "d2legacy.quest.progress")
	if q["population_available"] != false || q["completed"] != false {
		t.Fatal("missing resident was treated as death", q)
	}
	f.move(t, 31, 1)
	f.quest(t, 0, 0, true, false)
}
