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
// production world. This acceptance path proves its default Fire Bolt still
// crosses ordinary assignment, mana, cast, missile, and combat authority.
func TestSpellLabCastsProductionFireBolt(t *testing.T) {
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
	if learned.Len() != 15 {
		t.Fatalf("Spell Lab learned skills = %d, want 15 exact-ID behaviors", learned.Len())
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
	for _, id := range []int64{0, 36, 40, 45, 47, 48, 52, 54, 55, 66, 72, 98, 100, 104, 108} {
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
	for range 2 {
		if _, err := app.offlineSession.AdvanceWithSource(time.Second/25, app.commandSource); err != nil {
			t.Fatal(err)
		}
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
	vitals, _ := akara.GetDynamicStore(app.entitySimulation.World(), "d2legacy.player.vitals")
	playerVitals, _ := vitals.Get(player)
	mana, _ := playerVitals.Get("mana")
	if mana != int64(4093) {
		t.Fatalf("Spell Lab Fire Bolt mana = %v, want 4093 after one level-20 cast", mana)
	}
	if after, alive := stats.Get(target); alive {
		afterHealth, _ := after.Get("health")
		if afterHealth.(int64) >= beforeHealth.(int64) {
			t.Fatalf("Spell Lab Fire Bolt left target health unchanged at %v", afterHealth)
		}
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
