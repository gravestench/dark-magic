package clientapp

import (
	"testing"

	"github.com/gravestench/akara"
	gameworld "github.com/gravestench/dark-magic/internal/game/world"
	d2mapgen "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/mapgen"
	d2save "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/save"
	"github.com/gravestench/dark-magic/internal/presentation/maprender"
)

// TestCreatedCharacterEntersGeneratedActOneTown exercises the complete offline admission seam.
func TestCreatedCharacterEntersGeneratedActOneTown(t *testing.T) {
	fixture := newRealD2LegacyFixture(t, realD2LegacyFixtureConfig{})

	assertBloodMoorPresentation(t, fixture)
	character := createSelectedAcceptanceCharacter(t, fixture)
	fixture.advanceOffline(t, 10)
	assertCharacterEnteredTown(t, fixture, character)
}

// assertBloodMoorPresentation verifies generated paths and bounded chunk residency.
func assertBloodMoorPresentation(t *testing.T, fixture *realD2LegacyFixture) {
	t.Helper()

	app := fixture.app

	chunks, err := maprender.Index(
		fixture.options.Content,
		app.gameWorlds[2],
		"data/global/palette/ACT1/pal.pl2",
		maprender.DefaultChunkSize,
	)
	if err != nil {
		t.Fatal(err)
	}

	if len(chunks.Chunks) == 0 {
		t.Fatal("Blood Moor produced no indexed presentation chunks")
	}

	assertBloodMoorDirtPath(t, fixture)
	assertLazyChunkResidency(t, chunks)
}

// assertBloodMoorDirtPath verifies that semantic routes became non-default floor tiles.
func assertBloodMoorDirtPath(t *testing.T, fixture *realD2LegacyFixture) {
	t.Helper()

	app := fixture.app

	pathTiles := make(map[[2]int]bool)
	for _, tile := range app.gameWorldZones[2].Paths() {
		pathTiles[[2]int{tile.X, tile.Y}] = true
	}

	realizedPathFloors := 0

	for _, tile := range app.gameWorlds[2].Tiles {
		isDirtPath := tile.Layer == gameworld.LayerFloor &&
			tile.Identity.MainIndex == 0 &&
			tile.Identity.SubIndex != 0 &&
			pathTiles[[2]int{tile.X, tile.Y}]
		if isDirtPath {
			realizedPathFloors++
		}
	}

	if realizedPathFloors == 0 {
		t.Fatal("Blood Moor semantic route produced no realized dirt-path floors")
	}
}

// assertLazyChunkResidency guards against eager full-map RGBA composition.
func assertLazyChunkResidency(t *testing.T, chunks *maprender.Set) {
	t.Helper()

	for index, chunk := range chunks.Chunks {
		if chunk.Pixels != nil {
			t.Fatalf("indexed chunk %d eagerly retained expanded pixels", index)
		}
	}

	weight := 0

	visibleSample := min(16, len(chunks.Chunks))
	for index := range visibleSample {
		chunk, err := chunks.Materialize(index)
		if err != nil {
			t.Fatal(err)
		}

		weight += chunk.Pixels.Bounds().Dx() * chunk.Pixels.Bounds().Dy() * 4
	}

	const maximumWorldChunkBytes = 16 * 1024 * 1024
	if weight > maximumWorldChunkBytes {
		t.Fatalf(
			"Blood Moor sample residency = %d MiB across %d chunks, want <= 16 MiB",
			weight/(1024*1024),
			visibleSample,
		)
	}
}

// createSelectedAcceptanceCharacter creates the durable save admitted by the session.
func createSelectedAcceptanceCharacter(
	t *testing.T,
	fixture *realD2LegacyFixture,
) d2save.Character {
	t.Helper()

	character := d2save.Character{
		ID:        "amazon-campfirehero",
		Name:      "CampfireHero",
		Class:     "Amazon",
		Level:     1,
		Expansion: true,
		Stats:     &d2save.Stats{Vitality: 20},
	}
	if err := fixture.app.saves.Create(character); err != nil {
		t.Fatal(err)
	}

	if err := fixture.app.saves.Select(character.ID); err != nil {
		t.Fatal(err)
	}

	return character
}

// assertCharacterEnteredTown verifies session admission at the generated campfire anchor.
func assertCharacterEnteredTown(
	t *testing.T,
	fixture *realD2LegacyFixture,
	character d2save.Character,
) {
	t.Helper()

	app := fixture.app

	d2legacySource, err := app.modSource("d2legacy")
	if err != nil {
		t.Fatal(err)
	}

	wantX, wantY, found := d2mapgen.ResolveTownEntry(
		t.Context(),
		d2legacySource,
		app.records,
		app.gameWorlds[1],
	)
	if !found {
		t.Fatal("generated town has no campfire entry anchor")
	}

	identities, ok := akara.GetDynamicStore(app.entitySimulation.World(), "d2legacy.player.identity")
	if !ok {
		t.Fatal("player identity store was not admitted")
	}

	positions, _ := akara.GetDynamicStore(app.entitySimulation.World(), "d2legacy.world.position")
	locations, _ := akara.GetDynamicStore(app.entitySimulation.World(), "d2legacy.world.location")

	for _, entity := range identities.Entities() {
		identity, _ := identities.Get(entity)

		id, _ := identity.Get("character_id")
		if id != character.ID {
			continue
		}

		assertTownPosition(t, positions, locations, entity, wantX, wantY)

		return
	}

	t.Fatal("created and selected character was not admitted to the session")
}

// assertTownPosition checks one admitted entity's generated coordinates and level identity.
func assertTownPosition(
	t *testing.T,
	positions *akara.DynamicStore,
	locations *akara.DynamicStore,
	entity akara.Entity,
	wantX float64,
	wantY float64,
) {
	t.Helper()

	position, _ := positions.Get(entity)
	x, _ := position.Get("x")
	y, _ := position.Get("y")
	location, _ := locations.Get(entity)
	act, _ := location.Get("act")
	level, _ := location.Get("level_id")

	if x != wantX || y != wantY || act != int64(1) || level != int64(1) {
		t.Fatalf(
			"entry = (%v,%v) act=%v level=%v, want (%v,%v) act=1 level=1",
			x,
			y,
			act,
			level,
			wantX,
			wantY,
		)
	}
}
