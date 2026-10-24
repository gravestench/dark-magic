package d2legacy

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"github.com/gravestench/akara"
	"github.com/gravestench/dark-magic/internal/content"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	gamesession "github.com/gravestench/dark-magic/internal/game/session"
	"github.com/gravestench/dark-magic/internal/game/simulation"
	gameworld "github.com/gravestench/dark-magic/internal/game/world"
	modruntime "github.com/gravestench/dark-magic/internal/runtime/lua"
)

// coldPopulationRecords copies the baseline records so Cold Plains has a distinct three-member family.
func coldPopulationRecords() fixtureRecords {
	r := difficultyMonsterRecords()
	area := map[string]string{}
	for k, v := range r["data/global/excel/levels.txt"][0] {
		area[k] = v
	}
	area["Id"] = "3"
	area["mon1"] = "cold"
	area["nmon1"] = "cold"
	r["data/global/excel/levels.txt"] = append(r["data/global/excel/levels.txt"], area)
	stats := map[string]string{}
	for k, v := range r["data/global/excel/monstats.txt"][0] {
		stats[k] = v
	}
	stats["Id"] = "cold"
	stats["BaseId"] = "cold"
	stats["AI"] = "fallen"
	stats["aidist"] = "0"
	stats["MinGrp"] = "3"
	stats["MaxGrp"] = "3"
	r["data/global/excel/monstats.txt"] = append(r["data/global/excel/monstats.txt"], stats)
	second := map[string]string{}
	for k, v := range r["data/global/excel/monstats2.txt"][0] {
		second[k] = v
	}
	second["Id"] = "cold"
	r["data/global/excel/monstats2.txt"] = append(r["data/global/excel/monstats2.txt"], second)
	return r
}

// coldPopulationSession constructs or restores authority with shared collision geometry and owned cleanup.
func coldPopulationSession(
	t *testing.T,
	difficulty int,
	records fixtureRecords,
	checkpoint *simulation.Checkpoint,
) (*Authority, *gameecs.Engine, *gamesession.Session) {
	t.Helper()
	engine := gameecs.New()
	var restore []simulation.ParticipantState
	if checkpoint != nil {
		var err error
		_ = engine.Close()
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
	authority, err := StartWithConfig(
		t.Context(),
		content.D2Legacy(),
		records,
		engine,
		session,
		coldPopulationConfig(difficulty, restore),
	)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = authority.Stop(t.Context())
		_ = session.Close()
		_ = engine.Close()
	})
	geometry, err := gameworld.NewOpenMap(100, 100)
	if err != nil {
		t.Fatal(err)
	}
	for _, level := range []int{2, 3} {
		if err := modruntime.SetWorldMapForLevel(
			t.Context(),
			authority.Runtime,
			"d2legacy.gameplay.systems.init",
			"set_collision",
			level,
			geometry,
		); err != nil {
			t.Fatal(err)
		}
	}
	return authority, engine, session
}

// coldPopulationConfig keeps seed and immutable rules identical across original and restored sessions.
func coldPopulationConfig(difficulty int, restore []simulation.ParticipantState) Config {
	return Config{
		Seed:    314,
		Restore: restore,
		InitialData: map[string]any{
			"d2legacy.game_rules": map[string]any{
				"target": "lod-1.14d", "expansion": true, "difficulty": difficulty,
				"hardcore": false, "ladder": false, "maximum_players": 2,
			},
		},
	}
}

// coldZone provides identical room IDs across levels to expose accidental cross-level matching.
func coldZone(level, difficulty int) map[string]any {
	return map[string]any{
		"act": 1, "level_id": level, "difficulty": difficulty,
		"links": []any{},
		"rooms": []any{map[string]any{
			"id": "same", "populate": true,
			"x": 0, "y": 0, "width": 50, "height": 50,
			"points": []any{
				map[string]any{"x": 10, "y": 10},
				map[string]any{"x": 15, "y": 10},
				map[string]any{"x": 20, "y": 10},
			},
		}},
	}
}

// coldBootstrap submits the complete zone envelope through privileged command admission.
func coldBootstrap(t *testing.T, s *gamesession.Session, tick uint64, zones ...map[string]any) {
	t.Helper()
	payload, err := json.Marshal(map[string]any{"zones": zones})
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Submit(simulation.Command{
		Tick:      tick,
		Player:    "population",
		Authority: simulation.AuthoritySystem,
		Sequence:  tick,
		Kind:      "system.population.bootstrap",
		Payload:   payload,
	}); err != nil {
		t.Fatal(err)
	}
}

// coldEnter places a player in the requested level without changing the common player fixture.
func coldEnter(t *testing.T, s *gamesession.Session, level int, player string, seq uint64) {
	t.Helper()
	var payload map[string]any
	if err := json.Unmarshal(generatedPlayerPayload(t, "hero-"+player, player, 1, 1), &payload); err != nil {
		t.Fatal(err)
	}
	payload["level_id"] = level
	data, _ := json.Marshal(payload)
	if err := s.Submit(simulation.Command{
		Tick:      1,
		Player:    "system",
		Authority: simulation.AuthoritySystem,
		Sequence:  seq,
		Kind:      "system.player.enter",
		Payload:   data,
	}); err != nil {
		t.Fatal(err)
	}
}

// coldPlan decodes the authoritative envelope for level-scoped assertions.
func coldPlan(t *testing.T, a *Authority) []map[string]any {
	t.Helper()
	v, _ := a.State.Read("d2legacy.population.plan")
	var plan struct {
		Zones []map[string]any `json:"zones"`
	}
	if err := json.Unmarshal(v.Data, &plan); err != nil {
		t.Fatal(err)
	}
	return plan.Zones
}

// TestColdPlainsPopulationAdmitsBothLevelsAtomically checks ordered admission and distinct families at every
// difficulty.
func TestColdPlainsPopulationAdmitsBothLevelsAtomically(t *testing.T) {
	for difficulty := 0; difficulty < 3; difficulty++ {
		t.Run(fmt.Sprint(difficulty), func(t *testing.T) {
			a, e, s := coldPopulationSession(t, difficulty, coldPopulationRecords(), nil)
			coldBootstrap(t, s, 1, coldZone(2, difficulty), coldZone(3, difficulty))
			coldEnter(t, s, 2, "alice", 1)
			coldEnter(t, s, 3, "bob", 2)
			stepPartySession(t, s)
			zones := coldPlan(t, a)
			if len(zones) != 2 || zones[0]["level_id"] != float64(2) || zones[1]["level_id"] != float64(3) {
				t.Fatalf("zones=%#v", zones)
			}
			for level, want := range map[int]string{2: "fallen", 3: "cold"} {
				id := fmt.Sprintf("level:%d:room:same:monster:1", level)
				entity := monsterEntity(t, e, id)
				identities, _ := akara.GetDynamicStore(e.World(), "d2legacy.monster.identity")
				identity, _ := identities.Get(entity)
				if got, _ := identity.Get("definition_id"); got != want {
					t.Fatalf("definition=%v want=%s", got, want)
				}
			}
		})
	}
}

// TestColdPlainsPopulationRejectsDuplicateAndOutOfOrderZones requires invalid envelopes to preserve plan and ECS
// state.
func TestColdPlainsPopulationRejectsDuplicateAndOutOfOrderZones(t *testing.T) {
	for _, levels := range [][]int{{2, 2}, {3, 2}, {3}} {
		t.Run(fmt.Sprint(levels), func(t *testing.T) {
			a, e, s := coldPopulationSession(t, 0, coldPopulationRecords(), nil)
			before, _ := a.State.Read("d2legacy.population.plan")
			zones := []any{}
			for _, level := range levels {
				zones = append(zones, coldZone(level, 0))
			}
			payload, _ := json.Marshal(map[string]any{"zones": zones})
			if err := s.Submit(simulation.Command{
				Tick:      1,
				Player:    "population",
				Authority: simulation.AuthoritySystem,
				Sequence:  1,
				Kind:      "system.population.bootstrap",
				Payload:   payload,
			}); err == nil {
				t.Fatal("invalid zones admitted")
			}
			after, _ := a.State.Read("d2legacy.population.plan")
			if !reflect.DeepEqual(before, after) || monsterCount(e) != 0 {
				t.Fatal("invalid admission mutated plan/ECS")
			}
		})
	}
}

// TestColdPlainsPopulationRejectsSecondBootstrapWithoutMutation protects installed authority from a second
// bootstrap.
func TestColdPlainsPopulationRejectsSecondBootstrapWithoutMutation(t *testing.T) {
	a, e, s := coldPopulationSession(t, 0, coldPopulationRecords(), nil)
	coldBootstrap(t, s, 1, coldZone(2, 0), coldZone(3, 0))
	coldEnter(t, s, 3, "bob", 1)
	stepPartySession(t, s)
	before, _ := a.State.Read("d2legacy.population.plan")
	beforeECS, err := e.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	coldBootstrap(t, s, 2, coldZone(2, 0), coldZone(3, 0))
	if err := s.Step(); err == nil {
		t.Fatal("second bootstrap admitted")
	}
	after, _ := a.State.Read("d2legacy.population.plan")
	afterECS, err := e.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(before, after) || !reflect.DeepEqual(beforeECS, afterECS) {
		t.Fatal("second admission mutated plan/ECS")
	}
}

// TestColdPlainsPopulationSparseAndEmptyCandidates covers sparse selection and intentional zero-candidate rooms.
func TestColdPlainsPopulationSparseAndEmptyCandidates(t *testing.T) {
	for difficulty := 0; difficulty < 3; difficulty++ {
		for _, empty := range []bool{false, true} {
			t.Run(fmt.Sprintf("difficulty=%d/empty=%v", difficulty, empty), func(t *testing.T) {
				r := coldPopulationRecords()
				area := r["data/global/excel/levels.txt"][1]
				area["NumMon"] = "3"
				delete(area, "mon1")
				delete(area, "nmon1")
				area["mon3"] = "cold"
				area["nmon3"] = "cold"
				if empty {
					area["NumMon"] = "0"
				}
				_, e, s := coldPopulationSession(t, difficulty, r, nil)
				coldBootstrap(t, s, 1, coldZone(2, difficulty), coldZone(3, difficulty))
				coldEnter(t, s, 3, "bob", 1)
				stepPartySession(t, s)
				want := 3
				if empty {
					want = 0
				}
				if got := monsterCount(e); got != want {
					t.Fatalf("monsters=%d want=%d", got, want)
				}
			})
		}
	}
}

// TestColdPlainsPopulationRoomIDsAreLevelScoped checks activation and residency when levels reuse a room ID.
func TestColdPlainsPopulationRoomIDsAreLevelScoped(t *testing.T) {
	a, e, s := coldPopulationSession(t, 0, coldPopulationRecords(), nil)
	coldBootstrap(t, s, 1, coldZone(2, 0), coldZone(3, 0))
	coldEnter(t, s, 3, "bob", 1)
	stepPartySession(t, s)
	zones := coldPlan(t, a)
	if len(zones) != 2 {
		t.Fatalf("zones=%#v", zones)
	}
	if zones[0]["created"] != float64(0) || zones[1]["created"] != float64(3) {
		t.Fatalf("cross-level activation=%#v", zones)
	}
	monsterEntity(t, e, "level:3:room:same:monster:1")
	for _, level := range []int{2, 3} {
		value, err := modruntime.Call(
			t.Context(),
			a.Runtime,
			"d2legacy.bootstrap.population",
			"resident_at",
			"item",
			level,
			1,
			1,
		)
		if err != nil {
			t.Fatal(err)
		}
		resident, ok := value.(map[string]any)
		if !ok || resident["level_id"] != float64(level) || resident["room_id"] != "same" {
			t.Fatalf("level %d residency=%#v", level, value)
		}
	}
}

// TestColdPlainsPopulationSkipsBlockedOrOccupiedFootprints checks terrain, room edges, and live or queued
// occupancy.
func TestColdPlainsPopulationSkipsBlockedOrOccupiedFootprints(t *testing.T) {
	for _, scenario := range []string{
		"blocked-footprint",
		"empty-points",
		"room-edge",
		"occupied",
		"unattached-collider",
		"queued",
		"mixed",
	} {
		t.Run(scenario, func(t *testing.T) {
			a, e, s := coldPopulationSession(t, 0, coldPopulationRecords(), nil)
			zone := coldZone(3, 0)
			room := zone["rooms"].([]any)[0].(map[string]any)

			// Arrange candidate points and obstacles for this placement contract.
			switch scenario {
			case "empty-points":
				room["points"] = []any{}
			case "blocked-footprint":
				blocked, err := gameworld.NewOpenMap(100, 100)
				if err != nil {
					t.Fatal(err)
				}
				reference := gameworld.TileReference{Path: "synthetic"}
				for i := range reference.SubTileFlags {
					if i != 20 {
						reference.SubTileFlags[i].BlockWalk = true
					}
				}
				blocked.ReplaceFloor(2, 2, gameworld.TileIdentity{}, reference)
				blocked.RebuildFlags()
				room["points"] = []any{map[string]any{"x": 10, "y": 10}}
				if err := modruntime.SetWorldMapForLevel(
					t.Context(),
					a.Runtime,
					"d2legacy.gameplay.systems.init",
					"set_collision",
					3,
					blocked,
				); err != nil {
					t.Fatal(err)
				}
			case "unattached-collider":
				entity := e.World().MustCreateEntity()
				for name, values := range map[string]map[string]any{
					"d2legacy.world.position": {"x": 10.0, "y": 10.0},
					"d2legacy.world.location": {"act": int64(1), "level_id": int64(3)},
					"d2legacy.world.collider": {"radius": 1.0},
				} {
					store, _ := akara.GetDynamicStore(e.World(), name)
					if _, err := store.Set(entity, values); err != nil {
						t.Fatal(err)
					}
				}
				room["points"] = []any{map[string]any{"x": 10, "y": 10}}
			case "mixed":
				room["points"] = []any{
					map[string]any{"x": 0, "y": 0},
					map[string]any{"x": 10, "y": 10},
					map[string]any{"x": 10.1, "y": 10},
					map[string]any{"x": 15, "y": 10},
					map[string]any{"x": 20, "y": 10},
				}
			case "room-edge":
				room["points"] = []any{map[string]any{"x": 0, "y": 0}}
			case "occupied":
				room["points"] = []any{map[string]any{"x": 1, "y": 1}}
			case "queued":
				room["points"] = []any{
					map[string]any{"x": 10, "y": 10},
					map[string]any{"x": 10.1, "y": 10},
					map[string]any{"x": 10.2, "y": 10},
				}
			}

			// Activate the room through the ordinary population lifecycle.
			coldBootstrap(t, s, 1, coldZone(2, 0), zone)
			coldEnter(t, s, 3, "bob", 1)
			stepPartySession(t, s)

			// Verify both the accepted count and the exact surviving mixed points.
			want := 0
			if scenario == "queued" {
				want = 1
			}
			if scenario == "mixed" {
				want = 3
				positions, _ := akara.GetDynamicStore(e.World(), "d2legacy.world.position")
				for member, x := range []float64{10, 15, 20} {
					entity := monsterEntity(t, e, fmt.Sprintf("level:3:room:same:monster:%d", member+1))
					point, _ := positions.Get(entity)
					gotX, _ := point.Get("x")
					gotY, _ := point.Get("y")
					if gotX != x || gotY != float64(10) {
						t.Fatalf("member %d point=(%v,%v)", member+1, gotX, gotY)
					}
				}
			}
			if got := monsterCount(e); got != want {
				t.Fatalf("monsters=%d want=%d", got, want)
			}
		})
	}
}

// Visiting level 3 must not consume any of the established Blood Moor streams.
func TestColdPlainsPopulationKeepsLevelTwoRandomSequence(t *testing.T) {
	var previous map[string]map[string]any
	var draws []uint64
	for _, multi := range []bool{false, true} {
		a, e, s := coldPopulationSession(t, 0, coldPopulationRecords(), nil)
		zone := coldZone(2, 0)
		if multi {
			coldBootstrap(t, s, 1, zone, coldZone(3, 0))
		} else {
			payload, _ := json.Marshal(zone)
			if err := s.Submit(simulation.Command{
				Tick:      1,
				Player:    "population",
				Authority: simulation.AuthoritySystem,
				Sequence:  1,
				Kind:      "system.population.bootstrap",
				Payload:   payload,
			}); err != nil {
				t.Fatal(err)
			}
		}
		coldEnter(t, s, 2, "alice", 1)
		coldEnter(t, s, 3, "bob", 2)
		stepPartySession(t, s)
		got := map[string]map[string]any{}
		entity := monsterEntity(t, e, "level:2:room:same:monster:1")
		for _, name := range []string{
			"d2legacy.monster.identity",
			"d2legacy.monster.stats",
			"d2legacy.world.position",
			"d2legacy.world.location",
		} {
			store, _ := akara.GetDynamicStore(e.World(), name)
			component, _ := store.Get(entity)
			snapshot, err := component.Snapshot()
			if err != nil {
				t.Fatal(err)
			}
			got[name] = snapshot
		}
		current := []uint64{}
		for _, stream := range []string{
			"d2legacy.population.density",
			"d2legacy.population.family",
			"d2legacy.population.group",
			"d2legacy.population.seed",
			"d2legacy.monster.spawn.life",
		} {
			v, err := a.Random.Uint64n(stream, 1000000)
			if err != nil {
				t.Fatal(err)
			}
			current = append(current, v)
		}
		if multi && (!reflect.DeepEqual(previous, got) || !reflect.DeepEqual(draws, current)) {
			t.Fatalf("Cold Plains changed Blood Moor spawn or later draws: %v / %v", draws, current)
		}
		previous, draws = got, current
	}
}

// TestColdPlainsPopulationReplayMatchesContinuation compares restored commands and checksums with uninterrupted
// execution.
func TestColdPlainsPopulationReplayMatchesContinuation(t *testing.T) {
	_, _, s := coldPopulationSession(t, 0, coldPopulationRecords(), nil)
	coldBootstrap(t, s, 1, coldZone(2, 0), coldZone(3, 0))
	coldEnter(t, s, 3, "bob", 1)
	stepPartySession(t, s)
	before, err := s.Replay()
	if err != nil {
		t.Fatal(err)
	}
	checkpoint := before.Checkpoints[len(before.Checkpoints)-1]
	_, _, restored := coldPopulationSession(t, 0, coldPopulationRecords(), &checkpoint)
	command := simulation.Command{
		Tick:      2,
		Player:    "bob",
		Authority: simulation.AuthorityPlayer,
		Sequence:  1,
		Kind:      "player.move",
		Payload:   json.RawMessage(`{"x":1,"y":0,"running":false}`),
	}
	for _, session := range []*gamesession.Session{s, restored} {
		if err := session.Submit(command); err != nil {
			t.Fatal(err)
		}
		stepSession(t, session, 3)
	}
	original, err := s.Replay()
	if err != nil {
		t.Fatal(err)
	}
	continued, err := restored.Replay()
	if err != nil {
		t.Fatal(err)
	}
	var suffix []simulation.Command
	for _, c := range original.Commands {
		if c.Tick > checkpoint.Tick {
			suffix = append(suffix, c)
		}
	}
	if !reflect.DeepEqual(suffix, continued.Commands) {
		t.Fatalf("continuation commands differ: %#v / %#v", suffix, continued.Commands)
	}
	originalChecksum := original.Checkpoints[len(original.Checkpoints)-1].Checksum
	continuedChecksum := continued.Checkpoints[len(continued.Checkpoints)-1].Checksum
	if originalChecksum != continuedChecksum {
		t.Fatal("restored continuation checksum differs")
	}
	verifyColdPopulationReplay(t, original)
	verifyColdPopulationReplay(t, continued)
}

// VerifyReplay restores participant bytes before constructing Lua systems. These
// test adapters retain those bytes until StartWithConfig performs real recovery.
type coldReplayParticipant struct {
	id       string
	initial  []byte
	delegate simulation.StateParticipant
}

// StateID preserves the participant identity used to match restored authority stores.
func (p *coldReplayParticipant) StateID() string { return p.id }

// RestoreState copies replay bytes until authority construction can restore the owning store.
func (p *coldReplayParticipant) RestoreState(data []byte) error {
	p.initial = append([]byte(nil), data...)
	return nil
}

// SnapshotState delegates snapshots after authority construction has bound the owning participant.
func (p *coldReplayParticipant) SnapshotState() ([]byte, error) { return p.delegate.SnapshotState() }

// verifyColdPopulationReplay reconstructs the production authority participants before verifying deterministic
// replay.
func verifyColdPopulationReplay(t *testing.T, replay simulation.Replay) {
	t.Helper()
	var authority *Authority
	var session *gamesession.Session
	var pending []*coldReplayParticipant
	var participants []simulation.StateParticipant
	for _, v := range replay.InitialParticipants {
		p := &coldReplayParticipant{id: v.ID}
		pending = append(pending, p)
		participants = append(participants, p)
	}
	prepare := func(engine *gameecs.Engine) error {
		var err error
		session, err = gamesession.New(engine, gamesession.Config{})
		if err != nil {
			return err
		}
		restore := []simulation.ParticipantState{}
		for _, p := range pending {
			restore = append(restore, simulation.ParticipantState{ID: p.id, Data: p.initial})
		}
		authority, err = StartWithConfig(
			t.Context(),
			content.D2Legacy(),
			coldPopulationRecords(),
			engine,
			session,
			coldPopulationConfig(0, restore),
		)
		if err != nil {
			return err
		}
		identity, err := simulation.NewIdentityParticipant(authority.Identity)
		if err != nil {
			return err
		}
		for _, p := range pending {
			for _, delegate := range []simulation.StateParticipant{identity, authority.State, authority.Random} {
				if delegate.StateID() == p.id {
					p.delegate = delegate
				}
			}
		}
		geometry, err := gameworld.NewOpenMap(100, 100)
		if err != nil {
			return err
		}
		for _, level := range []int{2, 3} {
			if err := modruntime.SetWorldMapForLevel(
				t.Context(),
				authority.Runtime,
				"d2legacy.gameplay.systems.init",
				"set_collision",
				level,
				geometry,
			); err != nil {
				return err
			}
		}
		return nil
	}
	apply := func(_ *gameecs.Engine, command simulation.Command) error {
		modules := map[string]string{
			"system.population.bootstrap": "d2legacy.bootstrap.population",
			"system.player.enter":         "d2legacy.commands.enter_player",
			"player.move":                 "d2legacy.commands.move_player",
		}
		var payload map[string]any
		if err := json.Unmarshal(command.Payload, &payload); err != nil {
			return err
		}
		_, err := modruntime.Call(t.Context(), authority.Runtime, modules[command.Kind], "apply", map[string]any{
			"tick":      command.Tick,
			"player":    command.Player,
			"authority": string(command.Authority),
			"sequence":  command.Sequence,
			"kind":      command.Kind,
			"payload":   payload,
		})
		return err
	}
	defer func() {
		if authority != nil {
			_ = authority.Stop(t.Context())
		}
		if session != nil {
			_ = session.Close()
		}
	}()
	if err := simulation.VerifyReplay(replay, prepare, apply, participants...); err != nil {
		t.Fatal(err)
	}
}

// TestColdPlainsPopulationRejectsMalformedSecondZoneAtomically requires validation of every zone before any
// mutation.
func TestColdPlainsPopulationRejectsMalformedSecondZoneAtomically(t *testing.T) {
	for _, kind := range []string{"point", "duplicate-point", "link", "difficulty", "room"} {
		t.Run(kind, func(t *testing.T) {
			a, e, s := coldPopulationSession(t, 0, coldPopulationRecords(), nil)
			zone := coldZone(3, 0)
			room := zone["rooms"].([]any)[0].(map[string]any)
			switch kind {
			case "point":
				room["points"] = []any{map[string]any{"x": 100, "y": 10}}
			case "duplicate-point":
				room["points"] = []any{map[string]any{"x": 10, "y": 10}, map[string]any{"x": 10, "y": 10}}
			case "link":
				zone["links"] = []any{map[string]any{"from": "same", "to": "missing"}}
			case "difficulty":
				zone["difficulty"] = 1
			case "room":
				zone["rooms"] = append(zone["rooms"].([]any), room)
			}
			before, _ := a.State.Read("d2legacy.population.plan")
			payload, _ := json.Marshal(map[string]any{"zones": []any{coldZone(2, 0), zone}})
			if err := s.Submit(simulation.Command{
				Tick:      1,
				Player:    "population",
				Authority: simulation.AuthoritySystem,
				Sequence:  1,
				Kind:      "system.population.bootstrap",
				Payload:   payload,
			}); err == nil {
				t.Fatal("malformed second zone admitted")
			}
			after, _ := a.State.Read("d2legacy.population.plan")
			if !reflect.DeepEqual(before, after) || monsterCount(e) != 0 {
				t.Fatal("malformed second zone mutated authority")
			}
		})
	}
}

// Returning residents must reserve their footprints before any fresh room is
// materialized, regardless of which room appears first in the admitted plan.
func TestColdPlainsPopulationReturningRoomReservesFootprintsBeforeFreshRoom(t *testing.T) {
	for _, freshFirst := range []bool{false, true} {
		t.Run(fmt.Sprintf("fresh_first=%v", freshFirst), func(t *testing.T) {
			_, engine, session := coldPopulationSession(t, 0, coldPopulationRecords(), nil)
			returning := map[string]any{
				"id": "returning", "populate": true, "x": 0, "y": 0, "width": 20, "height": 20,
				"points": []any{map[string]any{"x": 19.4, "y": 10}},
			}
			fresh := map[string]any{
				"id": "fresh", "populate": true, "x": 20, "y": 0, "width": 20, "height": 20,
				"points": []any{map[string]any{"x": 20.5, "y": 10}, map[string]any{"x": 25, "y": 10}},
			}
			rooms := []any{returning, fresh}
			if freshFirst {
				rooms = []any{fresh, returning}
			}
			rooms = append(rooms,
				map[string]any{"id": "bridge", "populate": false, "x": 0, "y": 20, "width": 40, "height": 20, "points": []any{}},
				map[string]any{"id": "away", "populate": false, "x": 60, "y": 0, "width": 20, "height": 20, "points": []any{}},
			)
			zone := coldZone(3, 0)
			zone["rooms"] = rooms
			zone["links"] = []any{
				map[string]any{"from": "returning", "to": "bridge"},
				map[string]any{"from": "fresh", "to": "bridge"},
			}
			coldBootstrap(t, session, 1, coldZone(2, 0), zone)
			coldEnter(t, session, 3, "bob", 1)
			stepPartySession(t, session)
			const returningID = "level:3:room:returning:monster:1"
			original := monsterEntity(t, engine, returningID)
			positions, _ := akara.GetDynamicStore(engine.World(), "d2legacy.world.position")
			// Ordinary movement may leave a resident close enough to the shared edge
			// that its footprint reaches into its neighbor while its center stays here.
			setEntityPosition(t, positions, original, 19.9, 10)
			player := playerEntity(t, engine, "bob")
			setEntityPosition(t, positions, player, 61, 1)
			stepPartySession(t, session)
			inactive, _ := akara.GetDynamicStore(engine.World(), "d2legacy.world.inactive")
			if _, found := inactive.Get(original); !found {
				t.Fatal("returning resident did not deactivate")
			}
			if got := monsterCount(engine); got != 1 {
				t.Fatalf("fresh room populated before visit: monsters=%d", got)
			}
			setEntityPosition(t, positions, player, 1, 21)
			stepPartySession(t, session)
			if _, found := inactive.Get(original); found {
				t.Fatal("returning resident remained inactive")
			}
			if got := monsterCount(engine); got != 2 {
				t.Fatalf("returning footprint did not suppress overlapping candidate: monsters=%d want=2", got)
			}
			if got := monsterEntity(t, engine, returningID); got != original {
				t.Fatal("returning resident was recreated")
			}
			created := monsterEntity(t, engine, "level:3:room:fresh:monster:1")
			position, _ := positions.Get(created)
			x, _ := position.Get("x")
			y, _ := position.Get("y")
			if x != float64(25) || y != float64(10) {
				t.Fatalf("fresh resident point=(%v,%v), want (25,10)", x, y)
			}
		})
	}
}

// Admission preserves legacy plans and rejects invalid envelopes before mutation.
func TestStonyFieldPopulationAdmission(t *testing.T) {
	for _, levels := range [][]int{{2, 3}, {2, 3, 4}, {2, 4, 3}, {2, 3, 4, 4}, {2, 4}} {
		t.Run(fmt.Sprint(levels), func(t *testing.T) {
			a, e, s := coldPopulationSession(t, 0, coldPopulationRecords(), nil)
			before, _ := a.State.Read("d2legacy.population.plan")
			beforeECS, err := e.Snapshot()
			if err != nil {
				t.Fatal(err)
			}
			zones := []map[string]any{}
			for _, level := range levels {
				zones = append(zones, coldZone(level, 0))
			}
			payload, _ := json.Marshal(map[string]any{"zones": zones})
			err = s.Submit(simulation.Command{Tick: 1, Player: "population", Authority: simulation.AuthoritySystem,
				Sequence: 1, Kind: "system.population.bootstrap", Payload: payload})
			valid := reflect.DeepEqual(levels, []int{2, 3}) || reflect.DeepEqual(levels, []int{2, 3, 4})
			if valid {
				if err != nil {
					t.Fatal(err)
				}
				stepPartySession(t, s)
				installed := coldPlan(t, a)
				if len(installed) != len(levels) {
					t.Fatalf("installed zones=%#v", installed)
				}
				for i, level := range levels {
					if installed[i]["level_id"] != float64(level) {
						t.Fatalf("installed zone %d=%#v", i, installed[i])
					}
				}
				return
			}
			if err == nil {
				t.Fatal("invalid zone order admitted")
			}
			after, _ := a.State.Read("d2legacy.population.plan")
			afterECS, err := e.Snapshot()
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(before, after) || !reflect.DeepEqual(beforeECS, afterECS) {
				t.Fatal("invalid admission mutated plan or ECS")
			}
		})
	}
}

// Level 4 retains residents without loading monster definitions or advancing any random stream.
func TestStonyFieldPopulationItemLifecycleWithoutRandomDraws(t *testing.T) {
	for _, definition := range []string{"missing-level", "invalid-monster", "valid-monster"} {
		t.Run(definition, func(t *testing.T) {
			records := coldPopulationRecords()
			if definition != "missing-level" {
				area := map[string]string{}
				for key, value := range records["data/global/excel/levels.txt"][0] {
					area[key] = value
				}
				area["Id"] = "4"
				if definition == "invalid-monster" {
					area["mon1"] = "missing"
				}
				records["data/global/excel/levels.txt"] = append(records["data/global/excel/levels.txt"], area)
			}
			a, e, s := coldPopulationSession(t, 0, records, nil)
			item := e.World().MustCreateEntity()
			for name, values := range map[string]map[string]any{
				"d2legacy.world.position":    {"x": 10.0, "y": 10.0},
				"d2legacy.world.location":    {"act": int64(1), "level_id": int64(4)},
				"d2legacy.world.room_attach": {"id": "stony-item"},
				"d2legacy.world.selectable":  {"id": "stony-item", "kind": "item", "radius": 1.0},
			} {
				store, _ := akara.GetDynamicStore(e.World(), name)
				if _, err := store.Set(item, values); err != nil {
					t.Fatal(err)
				}
			}
			zone := coldZone(4, 0)
			zone["rooms"] = append(zone["rooms"].([]any), map[string]any{
				"id": "away", "populate": true, "x": 60, "y": 0, "width": 30, "height": 50,
				"points": []any{map[string]any{"x": 70, "y": 10}},
			})
			beforeRandom, err := a.Random.SnapshotState()
			if err != nil {
				t.Fatal(err)
			}
			coldBootstrap(t, s, 1, coldZone(2, 0), coldZone(3, 0), zone)
			coldEnter(t, s, 4, "alice", 1)
			stepPartySession(t, s)
			residents, _ := akara.GetDynamicStore(e.World(), "d2legacy.world.room_resident")
			resident, found := residents.Get(item)
			if !found {
				t.Fatal("preexisting item was not attached")
			}
			gotResident, err := resident.Snapshot()
			if err != nil {
				t.Fatal(err)
			}
			wantResident := map[string]any{"id": "stony-item", "level_id": int64(4), "room_id": "same"}
			if !reflect.DeepEqual(gotResident, wantResident) {
				t.Fatalf("resident=%#v", gotResident)
			}
			requests, _ := akara.GetDynamicStore(e.World(), "d2legacy.world.room_attach")
			if _, found := requests.Get(item); found {
				t.Fatal("resolved attachment request remains")
			}
			value, err := modruntime.Call(t.Context(), a.Runtime, "d2legacy.bootstrap.population", "resident_at", "later-item", 4, 10, 10)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(value, map[string]any{"id": "later-item", "level_id": float64(4), "room_id": "same"}) {
				t.Fatalf("resident_at=%#v", value)
			}
			positions, _ := akara.GetDynamicStore(e.World(), "d2legacy.world.position")
			inactive, _ := akara.GetDynamicStore(e.World(), "d2legacy.world.inactive")
			player := playerEntity(t, e, "alice")
			for _, visit := range []struct {
				x      float64
				active bool
			}{{1, true}, {61, false}, {1, true}} {
				setEntityPosition(t, positions, player, visit.x, 1)
				stepPartySession(t, s)
				if _, found := inactive.Get(item); found == visit.active {
					t.Fatalf("item inactive=%v, want active=%v", found, visit.active)
				}
				room := coldPlan(t, a)[2]["rooms"].([]any)[0].(map[string]any)
				if room["active"] != visit.active || room["activated"] != true {
					t.Fatalf("room lifecycle=%#v", room)
				}
				if got := monsterCount(e); got != 0 {
					t.Fatalf("level 4 created %d monsters", got)
				}
				afterRandom, err := a.Random.SnapshotState()
				if err != nil {
					t.Fatal(err)
				}
				if !reflect.DeepEqual(beforeRandom, afterRandom) {
					t.Fatalf("level 4 advanced random streams: before=%s after=%s", beforeRandom, afterRandom)
				}
				current, found := residents.Get(item)
				if !found {
					t.Fatal("item identity lost on return")
				}
				currentResident, err := current.Snapshot()
				if err != nil || !reflect.DeepEqual(gotResident, currentResident) {
					t.Fatalf("item residency changed: %#v, %v", currentResident, err)
				}
			}
		})
	}
}

// Level 4 must preserve legacy monster outcomes and the entire random continuation state.
func TestStonyFieldPopulationPreservesLevelTwoAndThree(t *testing.T) {
	for difficulty := 0; difficulty < 3; difficulty++ {
		t.Run(fmt.Sprint(difficulty), func(t *testing.T) {
			var previous map[string]map[string]map[string]any
			var previousRandom []byte
			for _, stony := range []bool{false, true} {
				a, e, s := coldPopulationSession(t, difficulty, coldPopulationRecords(), nil)
				zones := []map[string]any{coldZone(2, difficulty), coldZone(3, difficulty)}
				if stony {
					zones = append(zones, coldZone(4, difficulty))
				}
				coldBootstrap(t, s, 1, zones...)
				coldEnter(t, s, 2, "alice", 1)
				coldEnter(t, s, 3, "bob", 2)
				stepPartySession(t, s)
				got := map[string]map[string]map[string]any{}
				for _, entry := range []struct{ level, members int }{{2, 1}, {3, 3}} {
					for member := 1; member <= entry.members; member++ {
						id := fmt.Sprintf("level:%d:room:same:monster:%d", entry.level, member)
						entity := monsterEntity(t, e, id)
						got[id] = map[string]map[string]any{}
						for _, name := range []string{"d2legacy.monster.identity", "d2legacy.monster.stats", "d2legacy.world.position", "d2legacy.world.location"} {
							store, _ := akara.GetDynamicStore(e.World(), name)
							component, _ := store.Get(entity)
							snapshot, err := component.Snapshot()
							if err != nil {
								t.Fatal(err)
							}
							got[id][name] = snapshot
						}
					}
				}
				if monsterCount(e) != 4 {
					t.Fatalf("monsters=%d, want 4", monsterCount(e))
				}
				random, err := a.Random.SnapshotState()
				if err != nil {
					t.Fatal(err)
				}
				if stony && (!reflect.DeepEqual(previous, got) || !reflect.DeepEqual(previousRandom, random)) {
					t.Fatal("level 4 changed legacy monsters or random continuation")
				}
				previous, previousRandom = got, random
			}
		})
	}
}
