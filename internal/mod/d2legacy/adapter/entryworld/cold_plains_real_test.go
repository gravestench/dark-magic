package entryworld

import (
	"bytes"
	"io/fs"
	"os"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"github.com/gravestench/dark-magic/internal/content"
	recordstore "github.com/gravestench/dark-magic/internal/game/data/store"
	"github.com/gravestench/dark-magic/internal/game/worldgen"
	d2mapgen "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/mapgen"
)

// No object metadata is needed to verify topology preparation against owned terrain and collision.
type entryWorldObjectResolver struct{}

// Terrain-only preparation can leave static object resolution absent without changing collision.
func (entryWorldObjectResolver) ResolveStaticObject(int, int) (int, string, bool) {
	return 0, "", false
}

// Dynamic object lookup is likewise irrelevant to the seam and topology assertions.
func (entryWorldObjectResolver) ResolveDynamicObject(int, int) (string, bool) {
	return "", false
}

// realEntryWorldContent opts into owned MPQs only when the local path is supplied.
func realEntryWorldContent(t *testing.T) *content.FS {
	t.Helper()
	directory := os.Getenv("DARK_MAGIC_TEST_MPQ_DIRECTORY")
	if directory == "" {
		t.Skip("set DARK_MAGIC_TEST_MPQ_DIRECTORY to an owned Diablo II MPQ directory")
	}
	t.Setenv("MPQ_DIRECTORY", directory)
	source, err := content.FromEnvironment(content.Layer{Name: "d2legacy", FS: content.D2Legacy()})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := source.Close(); err != nil {
			t.Errorf("close owned content: %v", err)
		}
	})
	return source
}

// The public Build boundary must publish all five materialized zones and six directed transitions together.
func TestBuildColdPlainsReal(t *testing.T) {
	source := realEntryWorldContent(t)
	started := time.Now()
	prepared, err := Build(t.Context(), source, content.D2Legacy(), recordstore.New(source),
		entryWorldObjectResolver{}, 17, 0)
	t.Logf("owned-asset five-zone preparation took %s", time.Since(started))
	if err != nil {
		t.Fatal(err)
	}
	if len(prepared.Worlds) != 5 || len(prepared.Zones) != 5 || len(prepared.Seams) != 3 {
		t.Fatalf("prepared zones/maps/seams = %d/%d/%d",
			len(prepared.Zones), len(prepared.Worlds), len(prepared.Seams))
	}
	for _, level := range []int{1, 2, 3, 4} {
		if _, err := prepared.Destination(level); err != nil {
			t.Fatalf("level %d has no trusted destination: %v", level, err)
		}
	}
	population, err := prepared.PopulationData(0)
	if err != nil {
		t.Fatal(err)
	}
	zones := population["zones"].([]any)
	if len(zones) != 3 || zones[0].(map[string]any)["level_id"] != float64(2) ||
		zones[1].(map[string]any)["level_id"] != float64(3) ||
		zones[2].(map[string]any)["level_id"] != float64(4) {
		t.Fatalf("owned population zones=%#v", zones)
	}
	populated := 0
	for _, raw := range zones[1].(map[string]any)["rooms"].([]any) {
		room := raw.(map[string]any)
		if room["populate"] == true {
			populated++
		}
		seen := map[[2]float64]bool{}
		for _, rawPoint := range room["points"].([]any) {
			point := rawPoint.(map[string]any)
			x, y := point["x"].(float64), point["y"].(float64)
			key := [2]float64{x, y}
			flags, ok := prepared.Worlds[3].FlagsAt(int(x), int(y))
			if seen[key] || !ok || flags.Blocked() ||
				x < room["x"].(float64) || y < room["y"].(float64) ||
				x >= room["x"].(float64)+room["width"].(float64) ||
				y >= room["y"].(float64)+room["height"].(float64) {
				t.Fatalf("invalid owned level-3 point=%v", key)
			}
			seen[key] = true
		}
	}
	if populated == 0 {
		t.Fatal("owned Cold Plains has no population rooms")
	}
	transitions := TransitionData(prepared.Seams)
	if count := len(transitions["seams"].([]any)); count != 6 {
		t.Fatalf("directed transitions = %d, want six", count)
	}
}

// A pre-v2 checkpoint must remain playable with its original two-zone topology and identical snapshot identity.
func TestRestorePreparedV1KeepsTwoZonesReal(t *testing.T) {
	source := realEntryWorldContent(t)
	records := recordstore.New(source)
	generated, err := d2mapgen.GenerateEntryWorld(t.Context(), content.D2Legacy(), records, 17, 0)
	if err != nil {
		t.Fatal(err)
	}
	legacySeam := generated.Seams[0]
	legacySeam.TriggerRadius = 0
	legacySeam.MaxArrivalDistance = 0
	legacySeam.ArrivalInset = 0
	legacySeam.FirstWarpID = 0
	legacySeam.SecondWarpID = 0
	legacySeam.FirstWarpRole = ""
	legacySeam.SecondWarpRole = ""
	legacy := d2mapgen.EntryWorld{
		Zones: map[int]*worldgen.Zone{1: generated.Zones[1], 2: generated.Zones[2]},
		Seams: []d2mapgen.SeamSpec{legacySeam}, Version: 1,
	}
	snapshot, err := legacy.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	prepared, err := RestorePrepared(t.Context(), source, content.D2Legacy(), records,
		entryWorldObjectResolver{}, snapshot)
	if err != nil {
		t.Fatal(err)
	}

	if len(prepared.Worlds) != 2 || len(prepared.Zones) != 2 || len(prepared.Seams) != 1 ||
		prepared.Worlds[3] != nil || prepared.Worlds[4] != nil {
		t.Fatalf("v1 restore gained a wilderness: maps=%d zones=%d seams=%d",
			len(prepared.Worlds), len(prepared.Zones), len(prepared.Seams))
	}
	population, err := prepared.PopulationData(0)
	if err != nil {
		t.Fatal(err)
	}
	zones := population["zones"].([]any)
	if len(zones) != 1 || zones[0].(map[string]any)["level_id"] != float64(2) {
		t.Fatalf("v1 population zones=%#v", zones)
	}
	restoredSnapshot, err := prepared.Topology.Snapshot()
	if err != nil || !bytes.Equal(restoredSnapshot, snapshot) {
		t.Fatalf("v1 prepared topology changed snapshot bytes: %v", err)
	}
	wantChecksum, err := legacy.Checksum()
	if err != nil {
		t.Fatal(err)
	}
	gotChecksum, err := prepared.Topology.Checksum()
	if err != nil || gotChecksum != wantChecksum {
		t.Fatalf("v1 prepared topology checksum=%s, want %s: %v", gotChecksum, wantChecksum, err)
	}
	if count := len(TransitionData(prepared.Seams)["seams"].([]any)); count != 2 {
		t.Fatalf("v1 directed transitions = %d, want two", count)
	}
}

// Remove the destination level at the records boundary to prove Build does not return a partial world.
type recordsWithoutColdPlains struct{ Records }

// Filtering only Levels.txt preserves the rest of the owned fixture while making level 3 unavailable.
func (records recordsWithoutColdPlains) Load(path string) ([]map[string]string, error) {
	rows, err := records.Records.Load(path)
	if err != nil || path != "data/global/excel/Levels.txt" {
		return rows, err
	}
	filtered := make([]map[string]string, 0, len(rows))
	for _, row := range rows {
		if row["Id"] != "3" {
			filtered = append(filtered, row)
		}
	}
	return filtered, nil
}

// sourceWithOverride alters only the Lua policy under test; owned game data stays in its original source.
type sourceWithOverride struct {
	fs.FS
	override fstest.MapFS
}

// Only the named policy file is overridden; every other read uses the embedded source.
func (source sourceWithOverride) Open(name string) (fs.File, error) {
	if file, err := source.override.Open(name); err == nil {
		return file, nil
	}
	return source.FS.Open(name)
}

// Both a missing destination and an unsafe collision arrival must abort the
// public preparation boundary without returning a partial playable world.
func TestBuildFailsBeforePublicationOnMissingOrBlockedLevelThree(t *testing.T) {
	source := realEntryWorldContent(t)
	records := recordstore.New(source)
	if prepared, err := Build(t.Context(), source, content.D2Legacy(),
		recordsWithoutColdPlains{records}, entryWorldObjectResolver{}, 17, 0); err == nil || prepared != nil {
		t.Fatalf("missing level 3: prepared=%#v err=%v", prepared, err)
	} else if !strings.Contains(err.Error(), "Cold Plains is absent") {
		t.Fatalf("missing level 3 failed for an unrelated reason: %v", err)
	}
	const path = "lua/d2legacy/mapgen/entry_world.lua"
	policy, err := fs.ReadFile(content.D2Legacy(), path)
	if err != nil {
		t.Fatal(err)
	}
	changed := strings.Replace(string(policy), "arrival_inset = 6", "arrival_inset = 100", 1)
	if changed == string(policy) {
		t.Fatal("test could not inject an unsafe arrival inset")
	}
	policySource := sourceWithOverride{
		FS:       content.D2Legacy(),
		override: fstest.MapFS{path: &fstest.MapFile{Data: []byte(changed)}},
	}
	if prepared, err := Build(t.Context(), source, policySource, records,
		entryWorldObjectResolver{}, 17, 0); err == nil || prepared != nil {
		t.Fatalf("unsafe level-3 arrival: prepared=%#v err=%v", prepared, err)
	} else if !strings.Contains(err.Error(), "collision search bound") {
		t.Fatalf("unsafe arrival failed for an unrelated reason: %v", err)
	}
}
