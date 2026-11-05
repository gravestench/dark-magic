package d2legacy

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"testing"

	"github.com/gravestench/akara"
	"github.com/gravestench/dark-magic/internal/content"
	recordstore "github.com/gravestench/dark-magic/internal/game/data/store"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	gamesession "github.com/gravestench/dark-magic/internal/game/session"
	"github.com/gravestench/dark-magic/internal/game/simulation"
	gameworld "github.com/gravestench/dark-magic/internal/game/world"
	entryworld "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/entryworld"
	playeradapter "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/player"
	d2save "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/save"
	worldobjects "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/worldobjects"
	recovered "github.com/gravestench/dark-magic/internal/mod/d2legacy/data/recovered"
)

func TestNaturalCombatLootPickup(t *testing.T) {
	for _, seed := range []uint64{10, 8, 0} {
		t.Run(fmt.Sprint(seed), func(t *testing.T) { proveNaturalLootPickup(t, seed) })
	}
}

func proveNaturalLootPickup(t *testing.T, seed uint64) {
	t.Helper()
	directory := os.Getenv("DARK_MAGIC_TEST_MPQ_DIRECTORY")
	if directory == "" {
		t.Skip("set DARK_MAGIC_TEST_MPQ_DIRECTORY to owned Diablo II assets")
	}
	t.Setenv("MPQ_DIRECTORY", directory)
	source, err := content.FromEnvironment(content.Layer{Name: "d2legacy", FS: content.D2Legacy()})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { source.Close() })
	records := recordstore.New(source)
	ratioRows, err := records.Load("data/global/excel/itemratio.txt")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("real ItemRatio rows: %v", ratioRows)
	// Map seed 1 retains the production entry recipe. Authority seeds select junk and implicit equipment TCs.
	recoveredData, err := recovered.New(source).Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	objects, err := worldobjects.New(recoveredData, records)
	if err != nil {
		t.Fatal(err)
	}
	prepared, err := entryworld.Build(t.Context(), source, content.D2Legacy(), records, objects, 1, 0)
	if err != nil {
		t.Fatal(err)
	}
	character := d2save.Character{ID: "hero", Name: "Hero", Class: "Amazon", Level: 1, Expansion: true,
		Stats: &d2save.Stats{Dexterity: 20, Vitality: 20, Health: 50, MaxHealth: 50, Mana: 20, MaxMana: 20, Stamina: 84, MaxStamina: 84}}
	start := func(character d2save.Character, owner string) (*gameecs.Engine, *gamesession.Session) {
		engine := gameecs.New()
		session, err := gamesession.New(engine, gamesession.Config{CheckpointInterval: 1})
		if err != nil {
			t.Fatal(err)
		}
		authority, err := StartWithConfig(t.Context(), content.D2Legacy(), records, engine, session, Config{Seed: seed, InitialData: prepared.InitialData("", false)})
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { authority.Stop(t.Context()); session.Close(); engine.Close() })
		if err := prepared.InstallCollision(t.Context(), authority.Runtime); err != nil {
			t.Fatal(err)
		}
		population, err := prepared.PopulationCommand(0)
		if err != nil {
			t.Fatal(err)
		}
		if err := session.Submit(population); err != nil {
			t.Fatal(err)
		}
		destination, err := prepared.Destination(2)
		if err != nil {
			t.Fatal(err)
		}
		admission, err := playeradapter.AdmissionCommand(character, owner, destination, "system", 1, 1, simulation.AuthoritySystem)
		if err != nil {
			t.Fatal(err)
		}
		if err := session.Submit(admission); err != nil {
			t.Fatal(err)
		}
		stepSession(t, session, 3)
		return engine, session
	}
	engine, session := start(character, "alice")
	sourceLayouts := dynamicStore(t, engine, "d2legacy.items.world_source")
	if sourceLayouts.Len() != 1 {
		t.Fatal("production bootstrap did not create one world source")
	}
	for index, invalidOwner := range []string{"", " \t"} {
		destination, err := prepared.Destination(2)
		if err != nil {
			t.Fatal(err)
		}
		command, err := playeradapter.AdmissionCommand(character, invalidOwner, destination, "invalid-admission", uint64(index+1), engine.Tick()+1, simulation.AuthoritySystem)
		if err == nil {
			err = session.Submit(command)
		}
		if err == nil {
			t.Fatalf("reserved empty source owner admitted as player: %q", invalidOwner)
		}
	}
	startingItems := dynamicStore(t, engine, "d2legacy.item.identity").Len()
	hero := playerEntity(t, engine, "alice")
	positions := dynamicStore(t, engine, "d2legacy.world.position")
	position := func(entity akara.Entity) gameworld.Point {
		component, ok := positions.Get(entity)
		if !ok {
			t.Fatalf("entity %d has no position", entity)
		}
		x, _ := component.Get("x")
		y, _ := component.Get("y")
		return gameworld.Point{X: x.(float64), Y: y.(float64)}
	}
	sequence := uint64(0)
	activeOwner := "alice"
	submit := func(kind string, payload any) {
		sequence++
		encoded, err := json.Marshal(payload)
		if err != nil {
			t.Fatal(err)
		}
		if err := session.Submit(simulation.Command{Tick: engine.Tick() + 1, Player: activeOwner, Authority: simulation.AuthorityPlayer, Sequence: sequence, Kind: kind, Payload: encoded}); err != nil {
			t.Fatal(err)
		}
	}
	walk := func(goal gameworld.Point, stop float64) {
		path, err := prepared.Worlds[2].FindPath(gameworld.PathRequest{Start: position(hero), Goal: goal, Radius: 1, StopRadius: stop})
		if err != nil {
			t.Fatal(err)
		}
		for _, point := range path {
			if math.Hypot(position(hero).X-goal.X, position(hero).Y-goal.Y) <= stop {
				break
			}
			submit("player.move", map[string]any{"x": 0, "y": 0, "running": false, "target": map[string]any{"x": point.X, "y": point.Y}})
			reached := false
			for range 30 {
				stepSession(t, session, 1)
				current := position(hero)
				if math.Hypot(current.X-point.X, current.Y-point.Y) < 0.6 || math.Hypot(current.X-goal.X, current.Y-goal.Y) <= stop {
					reached = true
					break
				}
			}
			if !reached {
				t.Fatalf("ordinary movement stalled: at %v goal %v", position(hero), point)
			}
		}
		current := position(hero)
		submit("player.move", map[string]any{"x": 0, "y": 0, "running": false, "target": map[string]any{"x": current.X, "y": current.Y}})
		stepSession(t, session, 1)
	}
	monsters := dynamicStore(t, engine, "d2legacy.monster.identity")
	inactive := dynamicStore(t, engine, "d2legacy.world.inactive")
	locations := dynamicStore(t, engine, "d2legacy.world.location")
	var monster akara.Entity
	shortest := math.Inf(1)
	for _, entity := range monsters.Entities() {
		if _, hidden := inactive.Get(entity); hidden {
			continue
		}
		location, ok := locations.Get(entity)
		if !ok {
			continue
		}
		level, _ := location.Get("level_id")
		if level != int64(2) {
			continue
		}
		path, err := prepared.Worlds[2].FindPath(gameworld.PathRequest{Start: position(hero), Goal: position(entity), Radius: 1, StopRadius: 1.5})
		if err == nil && float64(len(path)) < shortest {
			monster, shortest = entity, float64(len(path))
		}
	}
	if monster == 0 {
		t.Fatal("no naturally activated reachable Moor monster")
	}
	walk(position(monster), 3.5)
	selected := dynamicStore(t, engine, "d2legacy.world.selectable")
	target, _ := selected.Get(monster)
	targetID, _ := target.Get("id")
	deaths := dynamicStore(t, engine, "d2legacy.monster.death")
	initialDefinition, _ := monsters.Get(monster)
	initialFamily, _ := initialDefinition.Get("definition_id")
	initialTC, _ := initialDefinition.Get("treasure_class")
	t.Logf("ordinary combat target %v real TC=%v", initialFamily, initialTC)
	for range 100 {
		if _, dead := deaths.Get(monster); dead {
			break
		}
		goal := position(monster)
		submit("player.use_skill", map[string]any{"side": "left", "target_x": goal.X, "target_y": goal.Y, "target_id": targetID})
		stepSession(t, session, 20)
	}
	death, dead := deaths.Get(monster)
	if !dead {
		t.Fatal("natural monster survived ordinary attacks")
	}
	drops, _ := death.Get("drops")
	tc, _ := death.Get("treasure_class")
	definition, _ := monsters.Get(monster)
	family, _ := definition.Get("definition_id")
	monsterTC, _ := definition.Get("treasure_class")
	if tc != monsterTC {
		t.Fatalf("death treasure class = %v, monster has %v", tc, monsterTC)
	}
	heroSelected, _ := selected.Get(hero)
	heroID, _ := heroSelected.Get("id")
	credited, _ := death.Get("credited_id")
	if credited != heroID {
		t.Fatalf("natural kill credited to %v, want %v", credited, heroID)
	}
	for _, field := range []string{"game_player_count", "effective_player_count", "monster_player_count", "no_drop_player_count"} {
		count, _ := death.Get(field)
		if count != int64(1) {
			t.Fatalf("natural solo death %s = %v, want 1", field, count)
		}
	}
	t.Logf("mapSeed=1 authoritySeed=%d generated monster=%v TC=%v committed drops=%v", seed, family, tc, drops)
	identities := dynamicStore(t, engine, "d2legacy.item.identity")
	placements := dynamicStore(t, engine, "d2legacy.item.placement")
	var facts []map[string]any
	if err := json.Unmarshal([]byte(drops.(string)), &facts); err != nil {
		t.Fatal(err)
	}
	if seed == 0 {
		if len(facts) != 0 {
			t.Fatalf("empty-roll control produced committed drops: %v", drops)
		}
		if identities.Len() != startingItems {
			t.Fatal("natural empty-roll kill materialized an item")
		}
		stepSession(t, session, 5)
		if identities.Len() != startingItems {
			t.Fatal("empty roll later materialized an item")
		}
		return
	}
	if len(facts) != 1 {
		t.Fatalf("positive control needs exactly one committed drop: %v", drops)
	}
	deathTick, _ := death.Get("tick")
	expectedID := fmt.Sprintf("loot:%v:%d:1", deathTick, monster)
	var ground akara.Entity
	for _, entity := range identities.Entities() {
		identity, _ := identities.Get(entity)
		id, _ := identity.Get("id")
		if id == expectedID {
			if ground != 0 {
				t.Fatal("death materialized duplicate item identity")
			}
			ground = entity
		}
	}
	if ground == 0 {
		t.Fatalf("exact death item %q not materialized", expectedID)
	}
	originalIdentity, _ := identities.Get(ground)
	originalCode, _ := originalIdentity.Get("code")
	if originalCode != facts[0]["code"] {
		t.Fatal("materialized code differs from committed drop")
	}
	generation := dynamicStore(t, engine, "d2legacy.item.generation")
	originalGeneration, _ := generation.Get(ground)
	originalFacts, _ := originalGeneration.Get("facts")
	if originalFacts != drops {
		t.Fatal("materialization changed committed rolled facts")
	}
	placed, _ := placements.Get(ground)
	container, _ := placed.Get("container")
	if container != "world" {
		t.Fatal("natural item not in ordinary ground placement")
	}
	if seed == 10 && originalCode != "vps" {
		t.Fatalf("potion control code = %v", originalCode)
	}
	if seed == 8 {
		found := false
		for _, table := range []string{"weapons", "armor"} {
			rows, err := records.Load("data/global/excel/" + table + ".txt")
			if err != nil {
				t.Fatal(err)
			}
			for _, row := range rows {
				if row["code"] == originalCode {
					found = true
				}
			}
		}
		if !found || facts[0]["quality"] == "unresolved" {
			t.Fatalf("equipment base unresolved: %v", facts)
		}
	}
	target, ok := selected.Get(ground)
	if !ok {
		t.Fatal("natural ground item is not selectable")
	}
	itemTarget, _ := target.Get("id")
	checkpoint, err := session.CanonicalCheckpoint()
	if err != nil {
		t.Fatal(err)
	}
	worldJSON, err := playeradapter.ProjectWorldView("alice", checkpoint)
	if err != nil {
		t.Fatal(err)
	}
	var public playeradapter.WorldView
	if err := json.Unmarshal(worldJSON, &public); err != nil {
		t.Fatal(err)
	}
	visible := false
	for _, entity := range public.Entities {
		if entity.ID == itemTarget && entity.Kind == "item" {
			visible = true
		}
	}
	if !visible {
		t.Fatal("natural ground item missing from public world projection")
	}
	// Clear the killer's collision footprint through ordinary movement before the peer approaches.
	dropPosition := position(ground)
	clearX, clearY, found := prepared.Worlds[2].OpenPointNearSubtile(dropPosition.X+8, dropPosition.Y+8)
	if !found {
		t.Fatal("no observed open point for killer to clear pickup approach")
	}
	walk(gameworld.Point{X: clearX, Y: clearY}, 1)
	admissionSequence := uint64(1)
	admit := func(owner string, level int) {
		admissionSequence++
		destination, err := prepared.Destination(level)
		if err != nil {
			t.Fatal(err)
		}
		peer := character
		peer.ID = owner
		peer.Name = owner
		command, err := playeradapter.AdmissionCommand(peer, owner, destination, "system", admissionSequence, engine.Tick()+1, simulation.AuthoritySystem)
		if err != nil {
			t.Fatal(err)
		}
		if err := session.Submit(command); err != nil {
			t.Fatal(err)
		}
		stepSession(t, session, 2)
	}
	admit("bob", 2)
	admit("carol", 1)
	assertVisible := func(owner string, wanted bool) {
		checkpoint, err := session.CanonicalCheckpoint()
		if err != nil {
			t.Fatal(err)
		}
		raw, err := playeradapter.ProjectWorldView(owner, checkpoint)
		if err != nil {
			t.Fatal(err)
		}
		var view playeradapter.WorldView
		if err := json.Unmarshal(raw, &view); err != nil {
			t.Fatal(err)
		}
		found := false
		for _, value := range view.Entities {
			if value.ID == itemTarget {
				found = true
				if value.Kind != "item" {
					t.Fatal("wrong public item kind")
				}
			}
		}
		if found != wanted {
			t.Fatalf("owner %s item visibility=%v want %v", owner, found, wanted)
		}
	}
	assertVisible("alice", true)
	assertVisible("carol", false)
	hero = playerEntity(t, engine, "bob")
	activeOwner = "bob"
	sequence = 0
	walk(position(ground), 2.5)
	assertVisible("bob", true)
	assertVisible("alice", true)
	identityBefore, _ := identities.Get(ground)
	ownerBefore, _ := identityBefore.Get("owner")
	walk(position(ground), 2.5)
	// Departure removes player-owned possessions, but never the session's generated public item.
	leavePayload, err := json.Marshal(map[string]any{"player": "alice"})
	if err != nil {
		t.Fatal(err)
	}
	admissionSequence++
	if err := session.Submit(simulation.Command{Tick: engine.Tick() + 1, Player: "system", Authority: simulation.AuthoritySystem, Sequence: admissionSequence, Kind: "system.player.leave", Payload: leavePayload}); err != nil {
		t.Fatal(err)
	}
	stepSession(t, session, 2)
	if _, exists := identities.Get(ground); !exists {
		t.Fatal("killer departure deleted public generated item")
	}
	identityAfterLeave, _ := identities.Get(ground)
	ownerAfterLeave, _ := identityAfterLeave.Get("owner")
	if ownerAfterLeave != ownerBefore {
		t.Fatal("departure changed exact item lifetime owner")
	}
	if sourceLayouts.Len() != 1 {
		t.Fatal("departure duplicated or removed source layout")
	}
	assertVisible("bob", true)
	assertVisible("carol", false)
	recovery, err := session.RecoveryCheckpoint()
	if err != nil {
		t.Fatal(err)
	}
	restorePoint := recovery.State
	restoredEngine, err := gameecs.RestoreSnapshot(*restorePoint.Snapshot)
	if err != nil {
		t.Fatal(err)
	}
	restoredSession, err := gamesession.New(restoredEngine, gamesession.Config{CheckpointInterval: 1})
	if err != nil {
		t.Fatal(err)
	}
	restored, err := StartWithConfig(t.Context(), content.D2Legacy(), records, restoredEngine, restoredSession, Config{Seed: seed, InitialData: prepared.InitialData("", false), Restore: restorePoint.Participants})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = restored.Stop(t.Context()); _ = restoredSession.Close(); _ = restoredEngine.Close() })
	if err := prepared.InstallCollision(t.Context(), restored.Runtime); err != nil {
		t.Fatal(err)
	}
	if err := restoredSession.RestoreRecoveryCheckpoint(recovery); err != nil {
		t.Fatal(err)
	}
	if dynamicStore(t, restoredEngine, "d2legacy.items.world_source").Len() != 1 {
		t.Fatal("restored bootstrap duplicated world source")
	}
	sequence++
	pickupPayload, _ := json.Marshal(map[string]any{"target_id": itemTarget})
	pickupCommand := simulation.Command{Tick: engine.Tick() + 1, Player: "bob", Authority: simulation.AuthorityPlayer, Sequence: sequence, Kind: "item.pickup", Payload: pickupPayload}
	if err := session.Submit(pickupCommand); err != nil {
		t.Fatal(err)
	}
	if _, err := restoredSession.SubmitNetwork(pickupCommand); err != nil {
		t.Fatal(err)
	}

	stepSession(t, session, 2)
	stepSession(t, restoredSession, 2)
	if _, ok := selected.Get(ground); ok {
		t.Fatal("ordinary pickup retained ground selection")
	}
	sequence++
	pickupCommand.Tick = engine.Tick() + 1
	pickupCommand.Sequence = sequence
	if err := session.Submit(pickupCommand); err != nil {
		t.Fatal(err)
	}
	if _, err := restoredSession.SubmitNetwork(pickupCommand); err != nil {
		t.Fatal(err)
	}
	stepSession(t, session, 3)
	stepSession(t, restoredSession, 3)
	originalFinal, err := session.CanonicalCheckpoint()
	if err != nil {
		t.Fatal(err)
	}
	restoredFinal, err := restoredSession.CanonicalCheckpoint()
	if err != nil {
		t.Fatal(err)
	}
	if originalFinal.Checksum != restoredFinal.Checksum {
		t.Fatal("departure/pickup checkpoint continuation differs")
	}

	if identities.Len() != startingItems+1 {
		t.Fatal("pickup duplicated or lost item")
	}
	identityAfter, _ := identities.Get(ground)
	ownerAfter, _ := identityAfter.Get("owner")
	if ownerAfter == ownerBefore {
		t.Fatal("other admitted player's pickup did not transfer ownership")
	}
	ownerLayout := dynamicStore(t, engine, "d2legacy.items.layout")
	acquiredLayout, _ := ownerLayout.Get(ownerAfter.(akara.Entity))
	ownerID, _ := acquiredLayout.Get("owner")
	if ownerID != "bob" {
		t.Fatalf("pickup owner = %v", ownerID)
	}
	placementAfter, _ := placements.Get(ground)
	containerAfter, _ := placementAfter.Get("container")
	if containerAfter != "inventory" {
		t.Fatal("pickup did not place inventory item")
	}
	factsAfter, _ := originalGeneration.Get("facts")
	if factsAfter != originalFacts {
		t.Fatal("pickup changed committed facts")
	}
	assertVisible("bob", false)
	assertVisible("carol", false)
	t.Logf("natural exact death %d tick%v item%v code%v acquired once by bob", monster, deathTick, itemTarget, originalCode)
}
