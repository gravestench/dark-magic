package clientapp

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/gravestench/akara"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	gameworld "github.com/gravestench/dark-magic/internal/game/world"
	"github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/entryworld"
	d2mapgen "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/mapgen"
	playeradapter "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/player"
	d2save "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/save"
	"github.com/gravestench/dark-magic/internal/presentation/maprender"
)

// TestCreatedCharacterEntersGeneratedActOneTown proves a durable selected save enters the real
// generated town while the paired wilderness remains renderable and memory-bounded.
func TestCreatedCharacterEntersGeneratedActOneTown(t *testing.T) {
	fixture := newRealD2LegacyFixture(t, realD2LegacyFixtureConfig{})

	assertBloodMoorPresentation(t, fixture)
	character := createSelectedAcceptanceCharacter(t, fixture)
	fixture.advanceOffline(t, 10)
	assertCharacterEnteredTown(t, fixture, character)
}

func TestOwnedTownNPCsProjectVisibleConnectedMirrors(t *testing.T) {
	fixture := newRealD2LegacyFixture(t, realD2LegacyFixtureConfig{})
	createSelectedAcceptanceCharacter(t, fixture)
	fixture.advanceOffline(t, 10)

	var x, y float64

	found := false

	for _, object := range fixture.app.gameWorlds[1].Objects {
		if object.Resolved && object.Type == gameworld.ObjectTypeDynamic && object.Class == "akara" {
			x, y = float64(object.X), float64(object.Y)
			found = true
		}
	}

	if !found {
		t.Fatal("owned town has no decoded Akara")
	}

	checkpoint, err := fixture.app.offlineSession.CanonicalCheckpoint()
	if err != nil {
		t.Fatal(err)
	}

	payload, err := playeradapter.ProjectWorldView("local-player", checkpoint)
	if err != nil {
		t.Fatal(err)
	}

	var view playeradapter.WorldView
	if err := json.Unmarshal(payload, &view); err != nil {
		t.Fatal(err)
	}

	var guide *playeradapter.WorldEntity

	for index := range view.Entities {
		if view.Entities[index].ID == "act1-akara" {
			guide = &view.Entities[index]
			break
		}
	}

	if guide == nil || guide.Position.X != x || guide.Position.Y != y || guide.Radius != 4 ||
		guide.Token != "PS" || guide.Mode != "NU" || guide.WeaponClass != "HTH" ||
		guide.Components != "TR=LIT" || guide.OverlayHeight != 2 {
		t.Fatalf("owned Akara connected projection = %#v, authored point = (%v,%v)", guide, x, y)
	}

	engine := gameecs.New()

	t.Cleanup(func() { _ = engine.Close() })
	registerRemoteViewSchemas(t, engine)

	client := &application{clientSimulation: engine}
	location := playeradapter.HUDLocation{Act: 1, LevelID: 1}

	if err := client.syncRemoteMirrors([]playeradapter.WorldEntity{*guide}, location); err != nil {
		t.Fatal(err)
	}

	appearances, _ := akara.GetDynamicStore(engine.World(), "d2legacy.monster.appearance")

	appearance, found := appearances.Get(client.remoteMirrors[guide.ID])
	if !found {
		t.Fatal("connected Akara mirror has no appearance")
	}

	token, _ := appearance.Get("token")
	if token != "PS" {
		t.Fatalf("connected Akara mirror token = %v", token)
	}

	rows, err := fixture.app.records.Load("data/global/excel/monstats.txt")
	if err != nil {
		t.Fatal(err)
	}

	wanted := map[string]string{"Akara": "", "Charsi": "", "Gheed": "", "Kashya": "", "Warriv": ""}

	classes := map[string]bool{"akara": true, "charsi": true, "gheed": true, "kashya": true, "warriv1": true}
	for _, row := range rows {
		if classes[row["Id"]] && row["npc"] == "1" {
			wanted[row["NameStr"]] = row["Code"]
		}
	}

	var residents []playeradapter.WorldEntity

	seen := map[string]bool{}

	for _, entity := range view.Entities {
		if token, ok := wanted[entity.Label]; ok {
			if seen[entity.Label] {
				t.Fatalf("duplicate town NPC %s", entity.Label)
			}

			if entity.Token != token || entity.Mode != "NU" || entity.WeaponClass != "HTH" || entity.Components == "" {
				t.Fatalf("town NPC appearance = %#v", entity)
			}

			residents = append(residents, entity)
			seen[entity.Label] = true
		}
	}

	if len(seen) != len(wanted) {
		t.Fatalf("town NPCs = %v, want %v", seen, wanted)
	}

	if err := client.syncRemoteMirrors(residents, location); err != nil {
		t.Fatal(err)
	}

	if len(client.remoteMirrors) != 5 {
		t.Fatalf("connected town has %d mirrors, want 5", len(client.remoteMirrors))
	}

	for _, resident := range residents {
		mirror, found := appearances.Get(client.remoteMirrors[resident.ID])
		if !found {
			t.Fatalf("connected %s mirror has no appearance", resident.Label)
		}

		if token, _ := mirror.Get("token"); token != resident.Token {
			t.Fatalf("connected %s token = %v, want %s", resident.Label, token, resident.Token)
		}
	}
}

func TestOwnedTownLayoutsPreserveNPCPlacementFacts(t *testing.T) {
	fixture := newRealD2LegacyFixture(t, realD2LegacyFixtureConfig{})
	for _, direction := range []string{"N", "E", "S", "W"} {
		t.Run(direction, func(t *testing.T) {
			// This checks placement transport only; the admission test covers actor appearance.
			path := fmt.Sprintf("data/global/tiles/act1/town/town%s1.ds1", direction)

			town, err := gameworld.Load(fixture.options.Content, path, nil, fixture.app.worldObjectResolver)
			if err != nil {
				t.Fatal(err)
			}

			data := entryworld.InteractionData(map[int]*gameworld.Map{1: town}, nil, "local-player", "")

			wanted := map[string]bool{"akara": true, "charsi": true, "gheed": true, "kashya": true, "warriv1": true}
			for _, object := range town.Objects {
				if !wanted[object.Class] {
					continue
				}

				matches := 0

				for _, value := range data["targets"].([]any) {
					target := value.(map[string]any)

					matchesClass := target["class"] == object.Class || (object.Class == "akara" && target["id"] == "act1-akara")
					if !matchesClass {
						continue
					}

					matches++

					if target["x"] != float64(object.X) || target["y"] != float64(object.Y) || target["level_id"] != float64(1) {
						t.Fatalf("%s placement = %#v, authored = %#v", object.Class, target, object)
					}
				}

				if matches != 1 {
					t.Fatalf("%s has %d interaction targets, want 1", object.Class, matches)
				}

				delete(wanted, object.Class)
			}

			if len(wanted) != 0 {
				t.Fatalf("town layout is missing %v", wanted)
			}
		})
	}
}

// assertBloodMoorPresentation checks both semantic world realization and renderer resource policy,
// catching worlds that generate correctly but cannot be presented safely.
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

// assertBloodMoorDirtPath requires generated route semantics to survive tile realization. Merely
// retaining path metadata would not prove players can visually identify the route.
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

// assertLazyChunkResidency proves indexing does not retain expanded RGBA maps and bounds a visible
// materialization sample, protecting production-sized wilderness from runaway memory use.
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

// createSelectedAcceptanceCharacter uses the real save store because admission must consume durable
// selected identity, not a test-only entity injected directly into ECS.
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

// assertCharacterEnteredTown independently resolves the generated campfire anchor and finds the
// admitted entity by durable character ID, proving generation and admission agree on placement.
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

// assertTownPosition verifies coordinates and world identity together; matching coordinates in the
// wrong act or level would still be an invalid admission.
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
