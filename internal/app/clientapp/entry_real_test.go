package clientapp

import (
	"os"
	"testing"
	"time"

	"github.com/gravestench/akara"
	"github.com/gravestench/dark-magic/internal/content"
	gameworld "github.com/gravestench/dark-magic/internal/game/world"
	"github.com/gravestench/dark-magic/internal/inputstate"
	"github.com/gravestench/dark-magic/internal/localization"
	"github.com/gravestench/dark-magic/internal/presentation/maprender"
)

// This is the complete offline admission seam exercised with production data:
// the frontend creates/selects a durable character, while the fixed-tick
// session—not Lua—creates its live ECS entity at the generated town anchor.
func TestCreatedCharacterEntersGeneratedActOneTown(t *testing.T) {
	if os.Getenv("MPQ_DIRECTORY") == "" {
		t.Skip("MPQ_DIRECTORY is not configured")
	}
	assets, err := content.FromEnvironment()
	if err != nil {
		t.Fatal(err)
	}
	app := &application{
		options:    Options{Content: assets},
		inputState: &inputstate.Store{},
		locale:     localization.New(assets, "English"),
	}
	if err := app.loadGameCatalogs(); err != nil {
		t.Fatal(err)
	}
	if err := app.buildOfflineSession(); err != nil {
		t.Fatal(err)
	}
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

	character, err := app.saves.CreateNamedWithOptions("CampfireHero", "Amazon", true, false)
	if err != nil {
		t.Fatal(err)
	}
	if err := app.saves.Select(character.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := app.offlineSession.AdvanceWithSource(time.Second, app.commandSource); err != nil {
		t.Fatal(err)
	}

	wantX, wantY, found := app.gameWorlds[1].ActOneTownEntry()
	if !found {
		t.Fatal("generated town has no campfire entry anchor")
	}
	identities, ok := akara.GetDynamicStore(app.entitySimulation.World(), "dm.player.identity")
	if !ok {
		t.Fatal("player identity store was not admitted")
	}
	positions, _ := akara.GetDynamicStore(app.entitySimulation.World(), "dm.world.position")
	locations, _ := akara.GetDynamicStore(app.entitySimulation.World(), "dm.world.location")
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
