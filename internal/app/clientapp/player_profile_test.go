package clientapp

import (
	"path/filepath"
	"testing"

	d2save "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/save"
)

func TestLoadPlayerProfileStartsEmptyAndRestoresPersistedSelection(t *testing.T) {
	path := filepath.Join(t.TempDir(), "profile.json")
	store, writablePath, err := loadPlayerProfile(path, nil)
	if err != nil {
		t.Fatal(err)
	}
	if writablePath != path || len(store.Characters()) != 0 {
		t.Fatalf("new profile store/path = %#v %q", store.Characters(), writablePath)
	}
	if err := store.Create(d2save.Character{ID: "hero", Name: "Hero"}); err != nil {
		t.Fatal(err)
	}
	if err := store.Select("hero"); err != nil {
		t.Fatal(err)
	}
	if err := d2save.WriteProfileFile(path, store.Profile()); err != nil {
		t.Fatal(err)
	}
	restored, _, err := loadPlayerProfile(path, nil)
	if err != nil {
		t.Fatal(err)
	}
	selected, ok := restored.Selected()
	if !ok || selected.Name != "Hero" {
		t.Fatalf("restored selection = %#v", selected)
	}
}

func TestDevelopmentFixturesCannotOverwritePlayerProfile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "profile.json")
	store, writablePath, err := loadPlayerProfile(path, []d2save.Character{{ID: "fixture"}})
	if err != nil {
		t.Fatal(err)
	}
	if writablePath != "" || store.Characters()[0].ID != "fixture" {
		t.Fatalf("fixture store/path = %#v %q", store.Characters(), writablePath)
	}
}
