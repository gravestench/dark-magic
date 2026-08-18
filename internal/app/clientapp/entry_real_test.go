package clientapp

import (
	"context"
	"math"
	"os"
	"testing"
	"time"

	"github.com/gravestench/akara"
	"github.com/gravestench/dark-magic/internal/content"
	"github.com/gravestench/dark-magic/internal/distribution"
	gameworld "github.com/gravestench/dark-magic/internal/game/world"
	"github.com/gravestench/dark-magic/internal/inputstate"
	"github.com/gravestench/dark-magic/internal/localization"
	entryworld "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/entryworld"
	d2mapgen "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/mapgen"
	d2save "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/save"
	"github.com/gravestench/dark-magic/internal/presentation/maprender"
	modruntime "github.com/gravestench/dark-magic/internal/runtime/lua"
)

func startTestD2LegacyAuthority(t *testing.T, app *application) {
	t.Helper()
	if err := app.scripts.Start(t.Context()); err != nil {
		t.Fatal(err)
	}
	source, err := app.modSource("d2legacy")
	if err != nil {
		t.Fatal(err)
	}
	definition, err := modruntime.LoadDefinition(t.Context(), app.scripts, source, "components/d2legacy.lua")
	if err != nil {
		t.Fatal(err)
	}
	component, err := definition.Managed().New(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if err := component.Start(t.Context()); err != nil {
		t.Fatal(err)
	}
	if err := app.queueEntryPopulation(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = component.Stop(context.Background())
		_ = app.scripts.Stop(context.Background())
	})
}

func realD2LegacyOptions(t *testing.T) Options {
	t.Helper()
	mods, err := distribution.PrepareMods("none")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = mods.Close() })
	assets, err := content.FromEnvironment(mods.Layers...)
	if err != nil {
		t.Fatal(err)
	}
	assetSetID, err := content.AssetSetIdentityFromEnvironment()
	if err != nil {
		t.Fatal(err)
	}
	return Options{Content: assets, Mods: &mods.Resolved, Packages: mods.Packages, AssetSetID: assetSetID}
}

// This is the complete offline admission seam exercised with production data:
// the frontend creates/selects a durable character, while the fixed-tick
// session—not Lua—creates its live ECS entity at the generated town anchor.
func TestCreatedCharacterEntersGeneratedActOneTown(t *testing.T) {
	if os.Getenv("MPQ_DIRECTORY") == "" {
		t.Skip("MPQ_DIRECTORY is not configured")
	}
	options := realD2LegacyOptions(t)
	assets := options.Content
	app := &application{
		options:    options,
		inputState: &inputstate.Store{},
		locale:     localization.New(assets, "English"),
		scripts:    modruntime.New(),
	}
	if err := app.loadGameCatalogs(); err != nil {
		t.Fatal(err)
	}
	if err := app.buildOfflineSession(); err != nil {
		t.Fatal(err)
	}
	startTestD2LegacyAuthority(t, app)
	chunks, err := maprender.Index(assets, app.gameWorlds[2], "data/global/palette/ACT1/pal.pl2", maprender.DefaultChunkSize)
	if err != nil {
		t.Fatal(err)
	}
	if len(chunks.Chunks) == 0 {
		t.Fatal("Blood Moor produced no indexed presentation chunks")
	}
	pathTiles := make(map[[2]int]bool)
	for _, tile := range app.gameWorldZones[2].Paths() {
		pathTiles[[2]int{tile.X, tile.Y}] = true
	}
	realizedPathFloors := 0
	for _, tile := range app.gameWorlds[2].Tiles {
		if tile.Layer == gameworld.LayerFloor && tile.Identity.MainIndex == 0 && tile.Identity.SubIndex != 0 && pathTiles[[2]int{tile.X, tile.Y}] {
			realizedPathFloors++
		}
	}
	if realizedPathFloors == 0 {
		t.Fatal("Blood Moor semantic route produced no realized dirt-path floors")
	}
	for index, chunk := range chunks.Chunks {
		if chunk.Pixels != nil {
			t.Fatalf("indexed chunk %d eagerly retained expanded pixels", index)
		}
	}
	weight := 0
	visibleSample := min(16, len(chunks.Chunks))
	for index := 0; index < visibleSample; index++ {
		chunk, materializeErr := chunks.Materialize(index)
		if materializeErr != nil {
			t.Fatal(materializeErr)
		}
		weight += chunk.Pixels.Bounds().Dx() * chunk.Pixels.Bounds().Dy() * 4
	}
	// Camera residency owns only a nearby working set. This catches accidental
	// return to full-map RGBA composition without coupling the test to a window.
	const maximumWorldChunkBytes = 16 * 1024 * 1024
	if weight > maximumWorldChunkBytes {
		t.Fatalf("Blood Moor sample residency = %d MiB across %d chunks, want <= 16 MiB", weight/(1024*1024), visibleSample)
	}
	t.Cleanup(func() {
		app.loading.Close()
		_ = app.offlineSession.Close()
		_ = app.entitySimulation.Close()
		_ = content.Close(assets)
	})

	character := d2save.Character{
		ID: "amazon-campfirehero", Name: "CampfireHero", Class: "Amazon", Level: 1, Expansion: true,
		Stats: &d2save.Stats{Vitality: 20},
	}
	if err := app.saves.Create(character); err != nil {
		t.Fatal(err)
	}
	if err := app.saves.Select(character.ID); err != nil {
		t.Fatal(err)
	}
	for range 10 {
		if _, err := app.offlineSession.AdvanceWithSource(time.Second/25, app.commandSource); err != nil {
			t.Fatal(err)
		}
	}
	locations, _ := akara.GetDynamicStore(app.entitySimulation.World(), "d2legacy.world.location")

	d2legacySource, err := app.modSource("d2legacy")
	if err != nil {
		t.Fatal(err)
	}
	wantX, wantY, found := d2mapgen.ResolveTownEntry(t.Context(), d2legacySource, app.records, app.gameWorlds[1])
	if !found {
		t.Fatal("generated town has no campfire entry anchor")
	}
	identities, ok := akara.GetDynamicStore(app.entitySimulation.World(), "d2legacy.player.identity")
	if !ok {
		t.Fatal("player identity store was not admitted")
	}
	positions, _ := akara.GetDynamicStore(app.entitySimulation.World(), "d2legacy.world.position")
	for _, entity := range identities.Entities() {
		identity, _ := identities.Get(entity)
		id, _ := identity.Get("character_id")
		if id != character.ID {
			continue
		}
		position, _ := positions.Get(entity)
		x, _ := position.Get("x")
		y, _ := position.Get("y")
		location, _ := locations.Get(entity)
		act, _ := location.Get("act")
		level, _ := location.Get("level_id")
		if x != wantX || y != wantY || act != int64(1) || level != int64(1) {
			t.Fatalf("entry = (%v,%v) act=%v level=%v, want (%v,%v) act=1 level=1", x, y, act, level, wantX, wantY)
		}
		return
	}
	t.Fatal("created and selected character was not admitted to the session")
}

// Combat Lab is intentionally just another client of the production Blood
// Moor session. This protects its startup recipe from silently returning to a
// presentation-only actor or an unselected character fixture.
func TestCombatLabFixtureEntersBloodMoor(t *testing.T) {
	if os.Getenv("MPQ_DIRECTORY") == "" {
		t.Skip("MPQ_DIRECTORY is not configured")
	}
	options := realD2LegacyOptions(t)
	assets := options.Content
	options.StartScene = "combat_lab"
	options.FixtureCharacters = 1
	options.FixtureWorldLevel = 2
	app := &application{
		options:    options,
		inputState: &inputstate.Store{},
		locale:     localization.New(assets, "English"),
		scripts:    modruntime.New(),
	}
	if err := app.loadGameCatalogs(); err != nil {
		t.Fatal(err)
	}
	if err := app.buildOfflineSession(); err != nil {
		t.Fatal(err)
	}
	startTestD2LegacyAuthority(t, app)
	t.Cleanup(func() {
		app.loading.Close()
		_ = app.offlineSession.Close()
		_ = app.entitySimulation.Close()
		_ = content.Close(assets)
	})

	for range 10 {
		if _, err := app.offlineSession.AdvanceWithSource(time.Second/25, app.commandSource); err != nil {
			t.Fatal(err)
		}
	}
	identities, ok := akara.GetDynamicStore(app.entitySimulation.World(), "d2legacy.player.identity")
	if !ok {
		t.Fatal("Combat Lab admitted no player identity store")
	}
	if identities.Len() != 1 {
		t.Fatalf("Combat Lab admitted players = %d, want 1", identities.Len())
	}
	locations, ok := akara.GetDynamicStore(app.entitySimulation.World(), "d2legacy.world.location")
	if !ok {
		t.Fatal("Combat Lab has no authoritative world locations")
	}
	for _, entity := range identities.Entities() {
		location, found := locations.Get(entity)
		if !found {
			t.Fatal("Combat Lab player has no authoritative location")
		}
		level, _ := location.Get("level_id")
		if level != int64(2) {
			t.Fatalf("Combat Lab player level = %v, want Blood Moor level 2", level)
		}
	}
	monsters, ok := akara.GetDynamicStore(app.entitySimulation.World(), "d2legacy.monster.identity")
	if !ok || monsters.Len() == 0 {
		t.Fatal("Combat Lab admitted no production Blood Moor hostiles")
	}
	positions, ok := akara.GetDynamicStore(app.entitySimulation.World(), "d2legacy.world.position")
	if !ok {
		t.Fatal("Combat Lab has no authoritative positions")
	}
	playerEntity := identities.Entities()[0]
	playerPosition, _ := positions.Get(playerEntity)
	playerX, _ := playerPosition.Get("x")
	playerY, _ := playerPosition.Get("y")
	nearby := false
	var nearbyMonster akara.Entity
	for _, monster := range monsters.Entities() {
		position, found := positions.Get(monster)
		if !found {
			continue
		}
		x, _ := position.Get("x")
		y, _ := position.Get("y")
		if math.Hypot(x.(float64)-playerX.(float64), y.(float64)-playerY.(float64)) <= 14 {
			nearby = true
			nearbyMonster = monster
			break
		}
	}
	if !nearby {
		t.Fatal("Combat Lab placed no hostile within its visible encounter radius")
	}
	selectables, ok := akara.GetDynamicStore(app.entitySimulation.World(), "d2legacy.world.selectable")
	if !ok {
		t.Fatal("Combat Lab has no authoritative selectable targets")
	}
	selected, _ := selectables.Get(nearbyMonster)
	targetID, _ := selected.Get("id")
	targetPosition, _ := positions.Get(nearbyMonster)
	// This package-level acceptance test does not load the Lua movement
	// integrator used by the running client. Move the already-validated nearby
	// target into footprint range so this section isolates the native
	// admission -> left assignment -> skill -> attack -> damage pipeline.
	if err := targetPosition.Set("x", playerX.(float64)+2.4); err != nil {
		t.Fatal(err)
	}
	if err := targetPosition.Set("y", playerY.(float64)); err != nil {
		t.Fatal(err)
	}
	targetX, _ := targetPosition.Get("x")
	targetY, _ := targetPosition.Get("y")
	stats, ok := akara.GetDynamicStore(app.entitySimulation.World(), "d2legacy.monster.stats")
	if !ok {
		t.Fatal("Combat Lab has no authoritative monster stats")
	}
	before, _ := stats.Get(nearbyMonster)
	beforeHealth, _ := before.Get("health")
	if err := app.commandIntents.Submit("player.use_skill", map[string]any{
		"side": "left", "target_x": targetX, "target_y": targetY, "target_id": targetID,
	}); err != nil {
		t.Fatal(err)
	}
	// Advance in host-sized slices: the session intentionally caps catch-up per
	// frame, so one giant duration is not equivalent to three seconds of play.
	for range 20 {
		if _, err := app.offlineSession.AdvanceWithSource(time.Second/25, app.commandSource); err != nil {
			t.Fatal(err)
		}
	}
	after, alive := stats.Get(nearbyMonster)
	if alive {
		afterHealth, _ := after.Get("health")
		if afterHealth.(int64) >= beforeHealth.(int64) {
			playerPosition, _ = positions.Get(playerEntity)
			playerX, _ = playerPosition.Get("x")
			playerY, _ = playerPosition.Get("y")
			targetPosition, _ = positions.Get(nearbyMonster)
			targetX, _ = targetPosition.Get("x")
			targetY, _ = targetPosition.Get("y")
			approaches, _ := akara.GetDynamicStore(app.entitySimulation.World(), "d2legacy.combat.attack_approach")
			animations, _ := akara.GetDynamicStore(app.entitySimulation.World(), "d2legacy.combat.attack_animation")
			events, _ := akara.GetDynamicStore(app.entitySimulation.World(), "d2legacy.combat.event")
			t.Fatalf("Combat Lab basic attack left monster health unchanged at %v; player=(%.1f,%.1f) target=(%.1f,%.1f) approaches=%d animations=%d events=%d", afterHealth, playerX, playerY, targetX, targetY, approaches.Len(), animations.Len(), events.Len())
		}
	}
}

// Spell Lab adds only ephemeral fixture learning and a read-only legend to the
// production world. This acceptance path proves a straight missile and the
// complete golem family cross their shared authoritative behavior systems.
func TestSpellLabCastsProductionSkillFamilies(t *testing.T) {
	if os.Getenv("MPQ_DIRECTORY") == "" {
		t.Skip("MPQ_DIRECTORY is not configured")
	}
	options := realD2LegacyOptions(t)
	options.StartScene = "spell_lab"
	options = applyDevelopmentSceneDefaults(options)
	app := &application{
		options: options, inputState: &inputstate.Store{},
		locale: localization.New(options.Content, "English"), scripts: modruntime.New(),
	}
	if err := app.loadGameCatalogs(); err != nil {
		t.Fatal(err)
	}
	if err := app.buildOfflineSession(); err != nil {
		t.Fatal(err)
	}
	startTestD2LegacyAuthority(t, app)
	t.Cleanup(func() {
		app.loading.Close()
		_ = app.offlineSession.Close()
		_ = app.entitySimulation.Close()
		_ = content.Close(options.Content)
	})
	for range 10 {
		if _, err := app.offlineSession.AdvanceWithSource(time.Second/25, app.commandSource); err != nil {
			t.Fatal(err)
		}
	}

	identities, ok := akara.GetDynamicStore(app.entitySimulation.World(), "d2legacy.player.identity")
	if !ok {
		t.Fatal("Spell Lab has no authoritative player identities")
	}
	if identities.Len() != 1 {
		t.Fatalf("Spell Lab admitted players = %d, want 1", identities.Len())
	}
	player := identities.Entities()[0]
	learned, ok := akara.GetDynamicStore(app.entitySimulation.World(), "d2legacy.player.learned_skill")
	if !ok {
		t.Fatal("Spell Lab has no authoritative learned skills")
	}
	if learned.Len() != 31 {
		t.Fatalf("Spell Lab learned skills = %d, want 31 exact-ID behaviors", learned.Len())
	}
	learnedIDs := map[int64]bool{}
	for _, entity := range learned.Entities() {
		value, _ := learned.Get(entity)
		owner, _ := value.Get("owner")
		if owner == player {
			id, _ := value.Get("skill_id")
			level, _ := value.Get("level")
			learnedIDs[id.(int64)] = level == int64(20)
		}
	}
	for _, id := range []int64{0, 36, 40, 45, 47, 48, 52, 54, 55, 66, 70, 72, 75, 80, 85, 90, 94, 95, 98, 99, 100, 103, 104, 105, 108, 109, 110, 115, 120, 124, 125} {
		if !learnedIDs[id] {
			t.Fatalf("Spell Lab skill %d is missing or not level 20: %#v", id, learnedIDs)
		}
	}
	assignments, _ := akara.GetDynamicStore(app.entitySimulation.World(), "d2legacy.player.skill_assignment")
	assignment, _ := assignments.Get(player)
	left, _ := assignment.Get("left")
	right, _ := assignment.Get("right")
	if left != int64(36) || right != int64(66) {
		t.Fatalf("Spell Lab assignments = left %v right %v, want Fire Bolt/Amplify Damage", left, right)
	}

	monsters, ok := akara.GetDynamicStore(app.entitySimulation.World(), "d2legacy.monster.identity")
	if !ok || monsters.Len() == 0 {
		spawn := app.gameWorldSpawns[2]
		room, found := entryworld.RoomIDAt(app.gameWorldZones[2], spawn[0], spawn[1])
		plan, installed := app.authoritativeState.Read("d2legacy.population.plan")
		t.Fatalf("Spell Lab admitted no nearby Blood Moor hostiles; spawn=(%.1f,%.1f) room=%q found=%v plan_registered=%v plan=%s", spawn[0], spawn[1], room, found, installed, plan.Data)
	}
	positions, _ := akara.GetDynamicStore(app.entitySimulation.World(), "d2legacy.world.position")
	playerPosition, _ := positions.Get(player)
	playerX, _ := playerPosition.Get("x")
	playerY, _ := playerPosition.Get("y")
	stats, ok := akara.GetDynamicStore(app.entitySimulation.World(), "d2legacy.monster.stats")
	if !ok {
		t.Fatal("Spell Lab has no authoritative monster stats")
	}
	selectables, ok := akara.GetDynamicStore(app.entitySimulation.World(), "d2legacy.world.selectable")
	if !ok {
		t.Fatal("Spell Lab has no authoritative selectable targets")
	}
	target := monsters.Entities()[0]
	targetPosition, _ := positions.Get(target)
	// Isolate the production cast -> projectile -> contact path from the
	// independently covered monster locomotion path. The running lab keeps its
	// naturally placed encounter; only this acceptance target is made stationary
	// and placed directly in Fire Bolt's line of travel.
	if err := targetPosition.Set("x", playerX.(float64)+6); err != nil {
		t.Fatal(err)
	}
	if err := targetPosition.Set("y", playerY); err != nil {
		t.Fatal(err)
	}
	targetX, _ := targetPosition.Get("x")
	targetY, _ := targetPosition.Get("y")
	targetSelectable, _ := selectables.Get(target)
	targetID, _ := targetSelectable.Get("id")
	before, _ := stats.Get(target)
	beforeHealth, _ := before.Get("health")
	if err := app.commandIntents.Submit("player.use_skill", map[string]any{
		"side": "left", "target_x": targetX, "target_y": targetY, "target_id": targetID,
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := app.offlineSession.AdvanceWithSource(time.Second/25, app.commandSource); err != nil {
		t.Fatal(err)
	}
	vitals, _ := akara.GetDynamicStore(app.entitySimulation.World(), "d2legacy.player.vitals")
	playerVitals, _ := vitals.Get(player)
	immediateManaRaw, _ := playerVitals.Get("mana_raw")
	maximumManaRaw, _ := playerVitals.Get("max_mana_raw")
	if immediateManaRaw.(int64) >= maximumManaRaw.(int64) {
		t.Fatalf("Spell Lab Fire Bolt did not spend mana at cast start: raw=%v max=%v", immediateManaRaw, maximumManaRaw)
	}
	if _, err := app.offlineSession.AdvanceWithSource(time.Second/25, app.commandSource); err != nil {
		t.Fatal(err)
	}
	animations, _ := akara.GetDynamicStore(app.entitySimulation.World(), "d2legacy.player.animation")
	animation, _ := animations.Get(player)
	mode, _ := animation.Get("mode")
	startTick, _ := animation.Get("start_tick")
	casts, _ := akara.GetDynamicStore(app.entitySimulation.World(), "d2legacy.skill.cast")
	cast, active := casts.Get(player)
	if mode != "SC" || !active {
		t.Fatalf("Spell Lab Fire Bolt action = mode %v active %v, want SC/true", mode, active)
	}
	effectTick, _ := cast.Get("effect_tick")
	completeTick, _ := cast.Get("complete_tick")
	if effectTick.(int64)-startTick.(int64) != 7 || completeTick.(int64)-startTick.(int64) != 14 {
		t.Fatalf("Spell Lab Fire Bolt SC timing = start %v effect %v complete %v, want +7/+14", startTick, effectTick, completeTick)
	}
	for range 18 {
		if _, err := app.offlineSession.AdvanceWithSource(time.Second/25, app.commandSource); err != nil {
			t.Fatal(err)
		}
	}
	animation, _ = animations.Get(player)
	mode, _ = animation.Get("mode")
	if mode != "NU" {
		t.Fatalf("Spell Lab Fire Bolt animation after completion = %v, want NU", mode)
	}
	playerVitals, _ = vitals.Get(player)
	mana, _ := playerVitals.Get("mana")
	manaRaw, _ := playerVitals.Get("mana_raw")
	if mana != int64(4096) || manaRaw != maximumManaRaw {
		t.Fatalf("Spell Lab Fire Bolt mana recovery = %v raw=%v, want full %v after cast completion", mana, manaRaw, maximumManaRaw)
	}
	if after, alive := stats.Get(target); alive {
		afterHealth, _ := after.Get("health")
		if afterHealth.(int64) >= beforeHealth.(int64) {
			t.Fatalf("Spell Lab Fire Bolt left target health unchanged at %v", afterHealth)
		}
	}

	// Exercise the complete golem family through the same assignment/cast path.
	// Each new member replaces the prior member of PetType `golem`; Iron Golem
	// additionally proves that an identified ground item is consumed only by a
	// successful cast and leaves durable provenance on the summon.
	advance := func(frames int) {
		t.Helper()
		for range frames {
			if _, err := app.offlineSession.AdvanceWithSource(time.Second/25, app.commandSource); err != nil {
				t.Fatal(err)
			}
		}
	}
	castGolem := func(skillID int64, targetID string) {
		t.Helper()
		if err := app.commandIntents.Submit("player.assign_skills", map[string]any{"right": skillID}); err != nil {
			t.Fatal(err)
		}
		advance(1)
		if err := app.commandIntents.Submit("player.use_skill", map[string]any{
			"side": "right", "target_x": playerX.(float64) + 2, "target_y": playerY.(float64), "target_id": targetID,
		}); err != nil {
			t.Fatal(err)
		}
		advance(1)
		if skillID == 90 {
			activeCasts, _ := akara.GetDynamicStore(app.entitySimulation.World(), "d2legacy.skill.cast")
			if _, active := activeCasts.Get(player); !active {
				vitals, _ := akara.GetDynamicStore(app.entitySimulation.World(), "d2legacy.player.vitals")
				playerVitals, _ := vitals.Get(player)
				manaRaw, _ := playerVitals.Get("mana_raw")
				assignments, _ := akara.GetDynamicStore(app.entitySimulation.World(), "d2legacy.player.skill_assignment")
				assignment, _ := assignments.Get(player)
				rightSkill, _ := assignment.Get("right")
				items, _ := akara.GetDynamicStore(app.entitySimulation.World(), "d2legacy.item.identity")
				placements, _ := akara.GetDynamicStore(app.entitySimulation.World(), "d2legacy.item.placement")
				positions, _ := akara.GetDynamicStore(app.entitySimulation.World(), "d2legacy.world.position")
				locations, _ := akara.GetDynamicStore(app.entitySimulation.World(), "d2legacy.world.location")
				inactive, _ := akara.GetDynamicStore(app.entitySimulation.World(), "d2legacy.world.inactive")
				var targetTypes, identified, placement, position, location any
				var unavailable bool
				for _, entity := range items.Entities() {
					item, _ := items.Get(entity)
					id, _ := item.Get("id")
					if id == targetID {
						targetTypes, _ = item.Get("item_types")
						identified, _ = item.Get("identified")
						if component, found := placements.Get(entity); found {
							placement, _ = component.Snapshot()
						}
						if component, found := positions.Get(entity); found {
							position, _ = component.Snapshot()
						}
						if component, found := locations.Get(entity); found {
							location, _ = component.Snapshot()
						}
						_, unavailable = inactive.Get(entity)
					}
				}
				playerLocation, _ := locations.Get(player)
				playerLocationSnapshot, _ := playerLocation.Snapshot()
				t.Fatalf("Spell Lab Iron Golem request failed preflight before mana payment; mana_raw=%v right_skill=%v item types=%v identified=%v placement=%v position=%v location=%v player_location=%v inactive=%v",
					manaRaw, rightSkill, targetTypes, identified, placement, position, location, playerLocationSnapshot, unavailable)
			}
		}
		advance(17)
	}
	assertOnlyGolem := func(definition string) akara.Entity {
		t.Helper()
		owned, ok := akara.GetDynamicStore(app.entitySimulation.World(), "d2legacy.owned_unit")
		if !ok || owned.Len() != 1 {
			t.Fatalf("Spell Lab owned golems = %d, want one", owned.Len())
		}
		entity := owned.Entities()[0]
		monster, found := monsters.Get(entity)
		if !found {
			t.Fatalf("Spell Lab golem %d has no monster identity", entity)
		}
		actual, _ := monster.Get("definition_id")
		if actual != definition {
			events, _ := akara.GetDynamicStore(app.entitySimulation.World(), "d2legacy.skill.summon_event")
			var outcome, reason any
			for _, eventEntity := range events.Entities() {
				event, _ := events.Get(eventEntity)
				kind, _ := event.Get("kind")
				if kind == "golem_summon" {
					outcome, _ = event.Get("outcome")
					reason, _ = event.Get("reason")
				}
			}
			t.Fatalf("Spell Lab golem definition = %v, want %s; last outcome=%v reason=%v", actual, definition, outcome, reason)
		}
		return entity
	}
	createMeleeEvent := func(attackerID, defenderID string, damageRaw int64) {
		t.Helper()
		entity := app.entitySimulation.World().MustCreateEntity()
		events, _ := akara.GetDynamicStore(app.entitySimulation.World(), "d2legacy.combat.melee_event")
		if _, err := events.Set(entity, map[string]any{
			"kind": "hit_resolved", "tick": int64(0), "attacker_id": attackerID,
			"target_id": defenderID, "hit": true, "damage_raw": damageRaw,
			"remaining_health_raw": int64(1), "hand": "rarm", "attack_rating": int64(1),
			"defense": int64(0), "hit_chance": int64(95), "outcome": "hit",
		}); err != nil {
			t.Fatal(err)
		}
	}
	castGolem(75, "")
	clay := assertOnlyGolem("claygolem")
	claySelectable, _ := selectables.Get(clay)
	clayID, _ := claySelectable.Get("id")
	createMeleeEvent(targetID.(string), clayID.(string), 256)
	advance(2)
	states, _ := akara.GetDynamicStore(app.entitySimulation.World(), "d2legacy.state.instance")
	statSources, _ := akara.GetDynamicStore(app.entitySimulation.World(), "d2legacy.stat.source")
	claySlow, velocitySlow, attackSlow := false, false, false
	for _, entity := range states.Entities() {
		state, _ := states.Get(entity)
		stateTarget, _ := state.Get("target")
		stateID, _ := state.Get("state_id")
		if stateTarget == target && stateID == "slowed" {
			claySlow = true
		}
	}
	for _, entity := range statSources.Entities() {
		source, _ := statSources.Get(entity)
		sourceTarget, _ := source.Get("target")
		stat, _ := source.Get("stat")
		value, _ := source.Get("value")
		if sourceTarget == target && value.(int64) < 0 {
			velocitySlow = velocitySlow || stat == "velocitypercent"
			attackSlow = attackSlow || stat == "attackrate"
		}
	}
	if !claySlow || !velocitySlow || !attackSlow {
		t.Fatalf("Spell Lab Clay Golem melee reaction state=%v velocity=%v attack=%v", claySlow, velocitySlow, attackSlow)
	}
	castGolem(85, "")
	blood := assertOnlyGolem("bloodgolem")
	reactions, _ := akara.GetDynamicStore(app.entitySimulation.World(), "d2legacy.combat.reactive_effect")
	if _, found := reactions.Get(blood); !found {
		t.Fatal("Spell Lab Blood Golem has no record-derived reactive effect")
	}
	bloodStats, _ := stats.Get(blood)
	bloodMaximum, _ := bloodStats.Get("max_health")
	if err := bloodStats.Set("health", bloodMaximum.(int64)-20*256); err != nil {
		t.Fatal(err)
	}
	playerVitals, _ = vitals.Get(player)
	playerMaximum, _ := playerVitals.Get("max_health")
	if err := playerVitals.Set("health", playerMaximum.(int64)-10); err != nil {
		t.Fatal(err)
	}
	bloodSelectable, _ := selectables.Get(blood)
	bloodID, _ := bloodSelectable.Get("id")
	createMeleeEvent(bloodID.(string), targetID.(string), 10*256)
	advance(1)
	bloodHealth, _ := bloodStats.Get("health")
	playerHealth, _ := playerVitals.Get("health")
	if bloodHealth.(int64) <= bloodMaximum.(int64)-20*256 || playerHealth.(int64) <= playerMaximum.(int64)-10 {
		t.Fatalf("Spell Lab Blood Golem did not split stolen life: golem=%v owner=%v", bloodHealth, playerHealth)
	}
	beforeOwnerTransfer := bloodHealth.(int64)
	if err := playerVitals.Set("health", playerHealth.(int64)+4); err != nil {
		t.Fatal(err)
	}
	advance(1)
	bloodHealth, _ = bloodStats.Get("health")
	if bloodHealth.(int64)-beforeOwnerTransfer != 256 {
		t.Fatalf("Spell Lab Blood Golem owner-healing transfer = %d raw, want 256", bloodHealth.(int64)-beforeOwnerTransfer)
	}
	castGolem(94, "")
	fire := assertOnlyGolem("firegolem")
	grants, _ := akara.GetDynamicStore(app.entitySimulation.World(), "d2legacy.monster.granted_skill")
	grant, found := grants.Get(fire)
	if !found {
		t.Fatal("Spell Lab Fire Golem has no granted Holy Fire fact")
	}
	grantedName, _ := grant.Get("skill")
	grantedLevel, _ := grant.Get("level")
	if grantedName != "holy fire" || grantedLevel != int64(27) {
		t.Fatalf("Spell Lab Fire Golem grant = %v level %v, want holy fire/27", grantedName, grantedLevel)
	}
	periodicDamage, _ := akara.GetDynamicStore(app.entitySimulation.World(), "d2legacy.combat.periodic_damage")
	periodic, found := periodicDamage.Get(fire)
	if !found {
		t.Fatal("Spell Lab Fire Golem has no decoded periodic-damage schedule")
	}
	period, _ := periodic.Get("period_ticks")
	channel, _ := periodic.Get("channel")
	if period != int64(50) || channel != "fire" {
		t.Fatalf("Spell Lab Fire Golem periodic damage = period %v channel %v", period, channel)
	}
	fireSelectable, _ := selectables.Get(fire)
	fireID, _ := fireSelectable.Get("id")
	targetStats, _ := stats.Get(target)
	targetMaximum, _ := targetStats.Get("max_health")
	if err := targetStats.Set("health", targetMaximum); err != nil {
		t.Fatal(err)
	}
	advance(51)
	periodicEvents, _ := akara.GetDynamicStore(app.entitySimulation.World(), "d2legacy.combat.event")
	firePulse := false
	for _, entity := range periodicEvents.Entities() {
		event, _ := periodicEvents.Get(entity)
		sourceKind, _ := event.Get("source_kind")
		attackerID, _ := event.Get("attacker_id")
		if sourceKind == "periodic_damage" && attackerID == fireID {
			firePulse = true
		}
	}
	if !firePulse {
		t.Fatal("Spell Lab Fire Golem emitted no Holy Fire damage event")
	}
	if err := app.commandIntents.Submit("item.move", map[string]any{
		"item_id":     "fixture-short-sword",
		"destination": map[string]any{"container": "world", "x": playerX.(float64), "y": playerY.(float64)},
	}); err != nil {
		t.Fatal(err)
	}
	advance(2)
	items, _ := akara.GetDynamicStore(app.entitySimulation.World(), "d2legacy.item.identity")
	placements, _ := akara.GetDynamicStore(app.entitySimulation.World(), "d2legacy.item.placement")
	inactiveItems, _ := akara.GetDynamicStore(app.entitySimulation.World(), "d2legacy.world.inactive")
	var ironItem akara.Entity
	var itemMinimum, itemMaximum int64
	for _, entity := range items.Entities() {
		item, _ := items.Get(entity)
		id, _ := item.Get("id")
		if id == "fixture-short-sword" {
			ironItem = entity
			itemTypes, _ := item.Get("item_types")
			materialFlags, _ := item.Get("material_flags")
			identified, _ := item.Get("identified")
			placement, _ := placements.Get(entity)
			container, _ := placement.Get("container")
			if itemTypes == "" || materialFlags.(int64)&2 == 0 || identified != true || container != "world" {
				t.Fatalf("Spell Lab Iron target types=%v material=%v identified=%v container=%v",
					itemTypes, materialFlags, identified, container)
			}
			itemMelee, _ := akara.GetDynamicStore(app.entitySimulation.World(), "d2legacy.item.melee")
			melee, _ := itemMelee.Get(entity)
			minimum, _ := melee.Get("physical_min")
			maximum, _ := melee.Get("physical_max")
			itemMinimum, itemMaximum = minimum.(int64), maximum.(int64)
			if _, inactive := inactiveItems.Get(entity); inactive {
				t.Fatal("Spell Lab Iron target item became inactive in the player's current room")
			}
			break
		}
	}
	if ironItem == 0 {
		t.Fatal("Spell Lab Iron target item is missing before cast")
	}
	// Enhanced Damage is local to the consumed weapon. Iron Golem must fold it
	// into the weapon's own damage before transferring the remaining property
	// sources; treating it as a whole-monster percentage would also multiply
	// the golem's intrinsic damage.
	modifierEntity := app.entitySimulation.World().MustCreateEntity()
	itemModifiers, _ := akara.GetDynamicStore(app.entitySimulation.World(), "d2legacy.item.stat_modifier")
	if _, err := itemModifiers.Set(modifierEntity, map[string]any{
		"item": ironItem, "source_id": "acceptance-enhanced-damage", "source_kind": "affix",
		"stat": "damagepercent", "operation": "local_percent", "value": int64(50), "order": int64(10),
	}); err != nil {
		t.Fatal(err)
	}
	// Start an otherwise-valid Iron Golem cast, then remove its item before the
	// SC action event. The effect-time revalidation must preserve the item and
	// existing Fire Golem while retaining the already-paid mana transaction.
	playerVitals, _ = vitals.Get(player)
	beforeInvalidatedMana, _ := playerVitals.Get("mana_raw")
	if err := app.commandIntents.Submit("player.assign_skills", map[string]any{"right": int64(90)}); err != nil {
		t.Fatal(err)
	}
	advance(1)
	if err := app.commandIntents.Submit("player.use_skill", map[string]any{
		"side": "right", "target_x": playerX.(float64) + 2, "target_y": playerY.(float64),
		"target_id": "fixture-short-sword",
	}); err != nil {
		t.Fatal(err)
	}
	advance(1)
	if err := app.commandIntents.Submit("item.move", map[string]any{
		"item_id": "fixture-short-sword", "destination": map[string]any{"container": "inventory", "x": 8, "y": 0},
	}); err != nil {
		t.Fatal(err)
	}
	advance(17)
	playerVitals, _ = vitals.Get(player)
	afterInvalidatedMana, _ := playerVitals.Get("mana_raw")
	if afterInvalidatedMana.(int64) >= beforeInvalidatedMana.(int64) {
		t.Fatal("Spell Lab invalidated Iron Golem cast did not retain its paid mana cost")
	}
	assertOnlyGolem("firegolem")
	if _, found := items.Get(ironItem); !found {
		t.Fatal("Spell Lab invalidated Iron Golem cast consumed its target item")
	}
	summonEvents, _ := akara.GetDynamicStore(app.entitySimulation.World(), "d2legacy.skill.summon_event")
	invalidated := false
	for _, entity := range summonEvents.Entities() {
		event, _ := summonEvents.Get(entity)
		outcome, _ := event.Get("outcome")
		reason, _ := event.Get("reason")
		if outcome == "invalidated" && reason == "item_not_on_ground" {
			invalidated = true
		}
	}
	if !invalidated {
		t.Fatal("Spell Lab emitted no item_not_on_ground result for invalidated Iron Golem cast")
	}
	if err := app.commandIntents.Submit("item.move", map[string]any{
		"item_id":     "fixture-short-sword",
		"destination": map[string]any{"container": "world", "x": playerX.(float64), "y": playerY.(float64)},
	}); err != nil {
		t.Fatal(err)
	}
	advance(2)
	castGolem(90, "fixture-short-sword")
	iron := assertOnlyGolem("irongolem")
	provenance, _ := akara.GetDynamicStore(app.entitySimulation.World(), "d2legacy.summon.item_provenance")
	itemSource, found := provenance.Get(iron)
	if !found {
		t.Fatal("Spell Lab Iron Golem has no consumed-item provenance")
	}
	consumedID, _ := itemSource.Get("item_id")
	identified, _ := itemSource.Get("identified")
	resolvedMinimum, _ := itemSource.Get("resolved_weapon_minimum_raw")
	resolvedMaximum, _ := itemSource.Get("resolved_weapon_maximum_raw")
	if consumedID != "fixture-short-sword" || identified != true {
		t.Fatalf("Spell Lab Iron Golem provenance = %v identified=%v", consumedID, identified)
	}
	if resolvedMinimum != itemMinimum*150/100 || resolvedMaximum != itemMaximum*150/100 {
		t.Fatalf("Spell Lab Iron Golem local Enhanced Damage = %v-%v, want %d-%d",
			resolvedMinimum, resolvedMaximum, itemMinimum*150/100, itemMaximum*150/100)
	}
	if _, found := itemModifiers.Get(modifierEntity); found {
		t.Fatal("Spell Lab Iron Golem retained the consumed item's local modifier entity")
	}
	ironStats, _ := stats.Get(iron)
	ironMinimum, _ := ironStats.Get("physical_min")
	ironMaximum, _ := ironStats.Get("physical_max")
	if ironMinimum.(int64) <= itemMinimum || ironMaximum.(int64) <= itemMaximum {
		t.Fatalf("Spell Lab Iron Golem base-item damage = %v-%v, consumed item %d-%d",
			ironMinimum, ironMaximum, itemMinimum, itemMaximum)
	}
}

// Warp Lab must remain a thin instrument over the production client path. This
// acceptance test starts it exactly as the CLI does, admits a real fixture,
// drives the shared movement mailbox, crosses the generated Act I seam, and
// observes the application swap its active world from authoritative location.
func TestWarpLabUsesProductionMovementAndTransition(t *testing.T) {
	if os.Getenv("MPQ_DIRECTORY") == "" {
		t.Skip("MPQ_DIRECTORY is not configured")
	}
	options := realD2LegacyOptions(t)
	options.StartScene = "warp_lab"
	options = applyDevelopmentSceneDefaults(options)
	app := &application{
		options:    options,
		inputState: &inputstate.Store{},
		locale:     localization.New(options.Content, "English"),
		scripts:    modruntime.New(),
	}
	if err := app.loadGameCatalogs(); err != nil {
		t.Fatal(err)
	}
	if err := app.buildOfflineSession(); err != nil {
		t.Fatal(err)
	}
	startTestD2LegacyAuthority(t, app)
	t.Cleanup(func() {
		app.loading.Close()
		_ = app.offlineSession.Close()
		_ = app.entitySimulation.Close()
		_ = content.Close(options.Content)
	})

	if !shouldActivateDevelopmentSession(app.options) {
		t.Fatal("Warp Lab direct start did not request offline-session activation")
	}
	if err := app.network.StartSelected(); err != nil {
		t.Fatal(err)
	}
	for range 10 {
		if err := app.advanceGame(time.Second / 25); err != nil {
			t.Fatal(err)
		}
	}

	controls, ok := akara.GetDynamicStore(app.entitySimulation.World(), "d2legacy.world.player_control")
	if !ok || controls.Len() != 1 {
		t.Fatalf("Warp Lab admitted players = %d, want 1", controls.Len())
	}
	locations, ok := akara.GetDynamicStore(app.entitySimulation.World(), "d2legacy.world.location")
	if !ok {
		t.Fatal("Warp Lab has no authoritative world locations")
	}
	player := controls.Entities()[0]
	location, found := locations.Get(player)
	if !found {
		t.Fatal("Warp Lab player has no authoritative location")
	}
	level, _ := location.Get("level_id")
	if level != int64(app.transitionSeam.Town.LevelID) {
		t.Fatalf("Warp Lab entry level = %v, want town %d", level, app.transitionSeam.Town.LevelID)
	}
	warps, ok := akara.GetDynamicStore(app.entitySimulation.World(), "d2legacy.world.warp")
	residents, residentsOK := akara.GetDynamicStore(app.entitySimulation.World(), "d2legacy.world.room_resident")
	positions, positionsOK := akara.GetDynamicStore(app.entitySimulation.World(), "d2legacy.world.position")
	selectables, selectablesOK := akara.GetDynamicStore(app.entitySimulation.World(), "d2legacy.world.selectable")
	if !ok || !residentsOK || !positionsOK || !selectablesOK {
		t.Fatal("Warp Lab is missing authoritative warp presentation stores")
	}
	if warps.Len() != 2 {
		t.Fatalf("Warp Lab authoritative endpoints = %d, want paired production entities", warps.Len())
	}
	if residents.Len() < 2 {
		t.Fatalf("Warp Lab room residents = %d, want both endpoints", residents.Len())
	}
	var portalID string
	var portalX, portalY float64
	for _, entity := range warps.Entities() {
		if _, present := residents.Get(entity); !present {
			t.Fatalf("Warp Lab endpoint %d has no room residency", entity)
		}
		portalLocation, present := locations.Get(entity)
		if !present {
			continue
		}
		portalLevel, _ := portalLocation.Get("level_id")
		if portalLevel != int64(app.transitionSeam.Town.LevelID) {
			continue
		}
		portalPosition, _ := positions.Get(entity)
		portalXValue, _ := portalPosition.Get("x")
		portalYValue, _ := portalPosition.Get("y")
		portalX, portalY = portalXValue.(float64), portalYValue.(float64)
		selectable, _ := selectables.Get(entity)
		portalIDValue, _ := selectable.Get("id")
		portalID = portalIDValue.(string)
		break
	}
	if portalID == "" {
		t.Fatal("Warp Lab created no town-side warp endpoint")
	}
	// A predicted client can briefly believe it is in range before authority.
	// The real point-click command must reject that stale attempt without
	// terminating the game session or moving the player through the warp.
	if err := app.commandIntents.Submit("interaction.open", map[string]any{
		"at": true, "x": portalX, "y": portalY,
	}); err != nil {
		t.Fatal(err)
	}
	if err := app.advanceGame(time.Second / 25); err != nil {
		t.Fatalf("stale Warp Lab interaction terminated the session: %v", err)
	}
	level, _ = location.Get("level_id")
	if level != int64(app.transitionSeam.Town.LevelID) {
		t.Fatalf("out-of-range Warp Lab interaction changed level to %v", level)
	}
	if err := app.playerControl.SetMoveTargetWithRadius(portalX, portalY, 3.5); err != nil {
		t.Fatal(err)
	}
	for range 250 {
		if err := app.advanceGame(time.Second / 25); err != nil {
			t.Fatal(err)
		}
		if !app.playerControl.HasMoveTarget() {
			break
		}
	}
	if app.playerControl.HasMoveTarget() {
		t.Fatal("Warp Lab player never reached the town-side warp")
	}
	// Exercise the same point-based API used by presentation input after the
	// authoritative movement mailbox reports route completion.
	if err := app.commandIntents.Submit("interaction.open", map[string]any{
		"at": true, "x": portalX, "y": portalY,
	}); err != nil {
		t.Fatal(err)
	}
	for range 5 {
		if err := app.advanceGame(time.Second / 25); err != nil {
			t.Fatal(err)
		}
	}
	level, _ = location.Get("level_id")
	if level != int64(app.transitionSeam.Wilderness.LevelID) {
		t.Fatalf("Warp Lab player did not operate paired warp; final level = %v", level)
	}
	if app.activeWorldLevel != app.transitionSeam.Wilderness.LevelID {
		t.Fatalf("active presentation world = %d, want %d", app.activeWorldLevel,
			app.transitionSeam.Wilderness.LevelID)
	}

	var returnX, returnY float64
	for _, entity := range warps.Entities() {
		portalLocation, present := locations.Get(entity)
		if !present {
			continue
		}
		portalLevel, _ := portalLocation.Get("level_id")
		if portalLevel != int64(app.transitionSeam.Wilderness.LevelID) {
			continue
		}
		portalPosition, _ := positions.Get(entity)
		returnXValue, _ := portalPosition.Get("x")
		returnYValue, _ := portalPosition.Get("y")
		returnX, returnY = returnXValue.(float64), returnYValue.(float64)
		break
	}
	if returnX == 0 && returnY == 0 {
		t.Fatal("Warp Lab created no wilderness-side warp endpoint")
	}
	if err := app.playerControl.SetMoveTargetWithRadius(returnX, returnY, 3.5); err != nil {
		t.Fatal(err)
	}
	for range 250 {
		if err := app.advanceGame(time.Second / 25); err != nil {
			t.Fatal(err)
		}
		if !app.playerControl.HasMoveTarget() {
			break
		}
	}
	if app.playerControl.HasMoveTarget() {
		t.Fatal("Warp Lab player never reached the wilderness-side warp")
	}
	// Model a route sample that remains queued while authority commits the
	// interaction. World-relative intent from level 2 must be invalidated when
	// presentation/navigation follows the player back to level 1; otherwise it
	// can be replanned in town and keep ownership ahead of the next click.
	if err := app.playerControl.SetMoveTarget(returnX+8, returnY); err != nil {
		t.Fatal(err)
	}
	if err := app.commandIntents.Submit("interaction.open", map[string]any{
		"at": true, "x": returnX, "y": returnY,
	}); err != nil {
		t.Fatal(err)
	}
	for range 5 {
		if err := app.advanceGame(time.Second / 25); err != nil {
			t.Fatal(err)
		}
	}
	level, _ = location.Get("level_id")
	if level != int64(app.transitionSeam.Town.LevelID) || app.activeWorldLevel != app.transitionSeam.Town.LevelID {
		t.Fatalf("Warp Lab return left authority/presentation at %v/%d", level, app.activeWorldLevel)
	}
	if app.playerControl.HasMoveTarget() {
		t.Fatal("Warp Lab return retained a route target from the previous world")
	}

	playerPosition, _ := positions.Get(player)
	startXValue, _ := playerPosition.Get("x")
	startYValue, _ := playerPosition.Get("y")
	startX, startY := startXValue.(float64), startYValue.(float64)
	town := app.gameWorlds[app.transitionSeam.Town.LevelID]
	goalX, goalY, found := town.OpenPointNearSubtileForRadius(startX-6, startY, 1)
	if !found {
		t.Fatal("Warp Lab return has no footprint-safe locomotion target")
	}
	if _, err := town.FindPath(gameworld.PathRequest{
		Start: gameworld.Point{X: startX, Y: startY}, Goal: gameworld.Point{X: goalX, Y: goalY}, Radius: 1,
	}); err != nil {
		t.Fatalf("Warp Lab return position cannot start production locomotion: %v", err)
	}
	if err := app.playerControl.SetMoveTarget(goalX, goalY); err != nil {
		t.Fatal(err)
	}
	moved := false
	for range 100 {
		if err := app.advanceGame(time.Second / 25); err != nil {
			t.Fatal(err)
		}
		xValue, _ := playerPosition.Get("x")
		yValue, _ := playerPosition.Get("y")
		if math.Hypot(xValue.(float64)-startX, yValue.(float64)-startY) > 0.5 {
			moved = true
			break
		}
	}
	if !moved {
		t.Fatal("Warp Lab locomotion did not resume after the return warp")
	}
}
