package clientapp

import (
	"context"
	"math"
	"os"
	"testing"
	"time"

	"github.com/gravestench/akara"
	"github.com/gravestench/dark-magic/internal/content"
	gameworld "github.com/gravestench/dark-magic/internal/game/world"
	"github.com/gravestench/dark-magic/internal/inputstate"
	"github.com/gravestench/dark-magic/internal/localization"
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
	definition, err := modruntime.LoadDefinition(t.Context(), app.scripts, app.options.Content, "components/d2legacy.lua")
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

// This is the complete offline admission seam exercised with production data:
// the frontend creates/selects a durable character, while the fixed-tick
// session—not Lua—creates its live ECS entity at the generated town anchor.
func TestCreatedCharacterEntersGeneratedActOneTown(t *testing.T) {
	if os.Getenv("MPQ_DIRECTORY") == "" {
		t.Skip("MPQ_DIRECTORY is not configured")
	}
	assets, err := content.FromEnvironment(content.Layer{Name: "d2legacy", FS: content.D2Legacy()})
	if err != nil {
		t.Fatal(err)
	}
	app := &application{
		options:    Options{Content: assets},
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

	character := d2save.Character{ID: "amazon-campfirehero", Name: "CampfireHero", Class: "Amazon", Level: 1, Expansion: true}
	if err := app.saves.Create(character); err != nil {
		t.Fatal(err)
	}
	if err := app.saves.Select(character.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := app.offlineSession.AdvanceWithSource(time.Second, app.commandSource); err != nil {
		t.Fatal(err)
	}
	monsters, found := akara.GetDynamicStore(app.entitySimulation.World(), "d2legacy.monster.identity")
	if !found || monsters.Len() == 0 {
		t.Fatal("real Blood Moor population produced no authoritative monsters")
	}
	locations, _ := akara.GetDynamicStore(app.entitySimulation.World(), "d2legacy.world.location")
	for _, entity := range monsters.Entities() {
		location, _ := locations.Get(entity)
		level, _ := location.Get("level_id")
		if level != int64(2) {
			t.Fatalf("Blood Moor monster level=%v", level)
		}
	}

	wantX, wantY, found := d2mapgen.ResolveTownEntry(t.Context(), app.options.Content, app.records, app.gameWorlds[1])
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
	assets, err := content.FromEnvironment(content.Layer{Name: "d2legacy", FS: content.D2Legacy()})
	if err != nil {
		t.Fatal(err)
	}
	app := &application{
		options: Options{
			Content:           assets,
			StartScene:        "combat_lab",
			FixtureCharacters: 1,
			FixtureWorldLevel: 2,
		},
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

	if _, err := app.offlineSession.AdvanceWithSource(time.Second, app.commandSource); err != nil {
		t.Fatal(err)
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
