package luaModLoader

import (
	"testing"
	"testing/fstest"
)

func testMod(name, version string, enabled bool, requires ...string) discoveredMod {
	return discoveredMod{
		filesystem: fstest.MapFS{"manifest.json": &fstest.MapFile{}},
		manifest:   &Manifest{Name: name, Version: version, Enabled: enabled, Requires: requires},
	}
}

func modIDs(mods []discoveredMod) []string {
	result := make([]string, len(mods))
	for idx := range mods {
		result[idx] = mods[idx].manifest.ID()
	}
	return result
}

func TestOrderModsIsStableAndHonorsModGlobals(t *testing.T) {
	mods := []discoveredMod{
		testMod("Zulu", "1.0", true, "api.mods.alpha10"),
		testMod("Alpha", "1.0", true),
		testMod("Middle", "1.0", true, "api.renderer"),
	}
	ordered, err := orderMods(mods)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"Alpha (1.0)", "Middle (1.0)", "Zulu (1.0)"}
	got := modIDs(ordered)
	for idx := range want {
		if got[idx] != want[idx] {
			t.Fatalf("order %v, want %v", got, want)
		}
	}
}

func TestOrderModsRejectsMissingMod(t *testing.T) {
	_, err := orderMods([]discoveredMod{testMod("Consumer", "1.0", true, "api.mods.missing10")})
	if err == nil {
		t.Fatal("expected missing dependency error")
	}
}

func TestOrderModsRejectsCycle(t *testing.T) {
	mods := []discoveredMod{
		testMod("Alpha", "1.0", true, "api.mods.beta10"),
		testMod("Beta", "1.0", true, "api.mods.alpha10"),
	}
	if _, err := orderMods(mods); err == nil {
		t.Fatal("expected dependency cycle error")
	}
}

func TestOrderModsIgnoresRequirementsOfDisabledMods(t *testing.T) {
	mods := []discoveredMod{testMod("Disabled", "1.0", false, "api.mods.missing10")}
	ordered, err := orderMods(mods)
	if err != nil {
		t.Fatal(err)
	}
	if len(ordered) != 1 {
		t.Fatalf("ordered %d mods, want 1", len(ordered))
	}
}
