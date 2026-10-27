package mapgen

import (
	"testing"

	"github.com/gravestench/dark-magic/internal/content"
)

type entryRecords map[string][]map[string]string

// Load returns deterministic in-memory record rows so the test exercises production Lua policy without external data.
func (records entryRecords) Load(path string) ([]map[string]string, error) {
	return records[path], nil
}

// Invalidate is intentionally inert because test records never change during a generation.
func (entryRecords) Invalidate(string) {}

// Loaded reports all fixture tables as available, matching the eager in-memory test boundary.
func (entryRecords) Loaded(string) bool { return true }

// TestLuaSelectedEntryWorldRestoresToSameTopologyChecksum verifies that policy-selected recipes and seams survive the
// durable snapshot boundary exactly, including difficulty and canonical checksum identity.
func TestLuaSelectedEntryWorldRestoresToSameTopologyChecksum(t *testing.T) {
	generated, err := RestoreEntryWorld(legacyEntryWorldBytes(t, 2))
	if err != nil {
		t.Fatal(err)
	}

	if generated.Seams[0].FirstLevel != 1 || generated.Seams[0].SecondLevel != 2 ||
		generated.Seams[0].FirstDirection != "east" || generated.Seams[0].SecondDirection != "west" {
		t.Fatalf("Lua-selected entry seam = %#v", generated.Seams[0])
	}

	if generated.Zones[1].Request().Difficulty != 0 || generated.Zones[2].Request().Difficulty != 0 {
		t.Fatalf(
			"entry-world difficulty = town %d, wilderness %d",
			generated.Zones[1].Request().Difficulty,
			generated.Zones[2].Request().Difficulty,
		)
	}

	want, err := generated.Checksum()
	if err != nil {
		t.Fatal(err)
	}

	snapshot, err := generated.Snapshot()
	if err != nil {
		t.Fatal(err)
	}

	restored, err := RestoreEntryWorld(snapshot)
	if err != nil {
		t.Fatal(err)
	}

	got, err := restored.Checksum()
	if err != nil {
		t.Fatal(err)
	}

	if got != want {
		t.Fatalf("restored entry topology checksum = %s, want %s", got, want)
	}
}

// TestGenerateEntryWorldRejectsInvalidDifficulty verifies that unsupported modes fail before starting the Lua runtime
// or reading recovered records.
func TestGenerateEntryWorldRejectsInvalidDifficulty(t *testing.T) {
	if _, err := GenerateEntryWorld(t.Context(), content.D2Legacy(), entryRecords{}, 42, 3); err == nil {
		t.Fatal("expected invalid difficulty to be rejected")
	}
}
