package content

import (
	"archive/zip"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"testing/fstest"
)

func TestLayerPriorityProvenanceAndEnumeration(t *testing.T) {
	t.Parallel()

	base := fstest.MapFS{
		"data/shared.txt": &fstest.MapFile{Data: []byte("base")},
		"data/base.txt":   &fstest.MapFile{Data: []byte("base only")},
	}
	shim := fstest.MapFS{
		"data/shared.txt": &fstest.MapFile{Data: []byte("shim")},
		"data/shim.txt":   &fstest.MapFile{Data: []byte("shim only")},
	}
	mods := fstest.MapFS{
		"data/shared.txt": &fstest.MapFile{Data: []byte("mod")},
		"data/mod.txt":    &fstest.MapFile{Data: []byte("mod only")},
	}
	contentFS, err := New(
		Layer{Name: "user-mods", FS: mods},
		Layer{Name: "darkmagic", FS: shim},
		Layer{Name: "d2data", FS: base},
	)
	if err != nil {
		t.Fatal(err)
	}
	data, err := fs.ReadFile(contentFS, `\data\shared.txt`)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "mod" {
		t.Fatalf("shared data = %q", data)
	}
	source, err := contentFS.Resolve("/data/shared.txt")
	if err != nil {
		t.Fatal(err)
	}
	if source.Layer != "user-mods" || source.Path != "data/shared.txt" {
		t.Fatalf("source = %#v", source)
	}
	entries, err := contentFS.ReadDir("data")
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, entry := range entries {
		names = append(names, entry.Name())
	}
	want := []string{"base.txt", "mod.txt", "shared.txt", "shim.txt"}
	if !reflect.DeepEqual(names, want) {
		t.Fatalf("entries = %v, want %v", names, want)
	}
}

func TestMountAndUnmount(t *testing.T) {
	t.Parallel()

	contentFS, err := New(Layer{Name: "base", FS: fstest.MapFS{"value": &fstest.MapFile{Data: []byte("base")}}})
	if err != nil {
		t.Fatal(err)
	}
	if err := contentFS.MountFirst(Layer{Name: "mod", FS: fstest.MapFS{"value": &fstest.MapFile{Data: []byte("mod")}}}); err != nil {
		t.Fatal(err)
	}
	if got := contentFS.Layers(); !reflect.DeepEqual(got, []string{"mod", "base"}) {
		t.Fatalf("layers = %v", got)
	}
	if !contentFS.Unmount("mod") {
		t.Fatal("expected mod to be unmounted")
	}
	data, err := fs.ReadFile(contentFS, "value")
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "base" {
		t.Fatalf("value = %q", data)
	}
}

func TestZIPAndDirectoryNormalizePaths(t *testing.T) {
	t.Parallel()

	directory := t.TempDir()
	zipPath := filepath.Join(directory, "shim.zip")
	file, err := os.Create(zipPath)
	if err != nil {
		t.Fatal(err)
	}
	writer := zip.NewWriter(file)
	entry, err := writer.Create("scripts/boot.lua")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := entry.Write([]byte("return true")); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}

	zipFS, err := ZIP(zipPath)
	if err != nil {
		t.Fatal(err)
	}
	defer Close(zipFS)
	data, err := fs.ReadFile(zipFS, `\scripts\boot.lua`)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "return true" {
		t.Fatalf("zip data = %q", data)
	}
}

func TestD2LegacyContainsBoot(t *testing.T) {
	t.Parallel()

	data, err := fs.ReadFile(D2Legacy(), "boot.lua")
	if err != nil {
		t.Fatal(err)
	}
	if len(data) == 0 {
		t.Fatal("empty boot.lua")
	}
}

func TestNormalizeRejectsTraversal(t *testing.T) {
	t.Parallel()

	for _, name := range []string{"../secret", "data/../../secret"} {
		if _, err := Normalize(name); err == nil {
			t.Fatalf("Normalize(%q) succeeded", name)
		}
	}
}

func TestExistsWalkAndInvalidation(t *testing.T) {
	contentFS, err := New(Layer{Name: "base", FS: fstest.MapFS{
		"components/a.lua": &fstest.MapFile{Data: []byte("return {}")},
		"components/b.lua": &fstest.MapFile{Data: []byte("return {}")},
	}})
	if err != nil {
		t.Fatal(err)
	}
	if !contentFS.Exists(`\components\a.lua`) || contentFS.Exists("missing") {
		t.Fatal("unexpected existence result")
	}
	var names []string
	if err := contentFS.Walk("components", func(name string, _ fs.DirEntry, err error) error {
		if err == nil {
			names = append(names, name)
		}
		return err
	}); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(names, []string{"components", "components/a.lua", "components/b.lua"}) {
		t.Fatalf("walk = %v", names)
	}
	changes, cancel := contentFS.Subscribe(1)
	defer cancel()
	change, err := contentFS.Invalidate(`\components\a.lua`)
	if err != nil {
		t.Fatal(err)
	}
	if observed := <-changes; observed != change || observed.Generation != 1 || observed.Path != "components/a.lua" {
		t.Fatalf("change = %#v observed = %#v", change, observed)
	}
	if contentFS.Generation() != change.Generation {
		t.Fatalf("generation = %d", contentFS.Generation())
	}
}

func TestListCombinesDirectoryAndFlatArchiveIndexes(t *testing.T) {
	t.Parallel()

	directory := fstest.MapFS{
		"data/global/tiles/local.ds1":  &fstest.MapFile{},
		"data/global/tiles/shared.dt1": &fstest.MapFile{},
	}
	archive := &testListedFS{
		FS: fstest.MapFS{
			"data/global/tiles/shared.dt1":  &fstest.MapFile{},
			"data/global/tiles/archive.dt1": &fstest.MapFile{},
		},
		paths: []string{
			`data\global\tiles\shared.dt1`,
			`data\global\tiles\archive.dt1`,
			`data\global\other\ignored.dt1`,
		},
	}
	contentFS, err := New(
		Layer{Name: "directory", FS: directory},
		Layer{Name: "archive", FS: archive},
	)
	if err != nil {
		t.Fatal(err)
	}
	got, err := contentFS.List(`data\global\tiles`, ".DT1")
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"data/global/tiles/archive.dt1", "data/global/tiles/shared.dt1"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("list = %v, want %v", got, want)
	}
}

type testListedFS struct {
	fs.FS
	paths []string
}

func (f *testListedFS) Paths() []string { return append([]string(nil), f.paths...) }

func TestFromEnvironmentAppliesConfiguredModPriority(t *testing.T) {
	mods := t.TempDir()
	if err := os.WriteFile(filepath.Join(mods, "boot.lua"), []byte("mod boot"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("DARK_MAGIC_TEST_MODS", mods)
	t.Setenv("DARK_MAGIC_MOD_DIRECTORY", "$DARK_MAGIC_TEST_MODS")
	t.Setenv("MPQ_DIRECTORY", "")
	contentFS, err := FromEnvironment()
	if err != nil {
		t.Fatal(err)
	}
	if got := contentFS.Layers(); !reflect.DeepEqual(got, []string{"user-mods", "d2legacy"}) {
		t.Fatalf("layers = %v", got)
	}
	data, err := fs.ReadFile(contentFS, "boot.lua")
	if err != nil || string(data) != "mod boot" {
		t.Fatalf("boot = %q, %v", data, err)
	}
}

func TestFromEnvironmentMountsMultipleMPQDirectoriesInOrder(t *testing.T) {
	first, second := t.TempDir(), t.TempDir()
	if err := os.WriteFile(filepath.Join(first, "shared.gpl"), []byte("first"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(second, "shared.gpl"), []byte("second"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(second, "second-only.gpl"), []byte("mounted"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("DARK_MAGIC_MOD_DIRECTORY", "")
	t.Setenv("DARK_MAGIC_FIRST_CONTENT", first)
	t.Setenv("DARK_MAGIC_SECOND_CONTENT", second)
	t.Setenv("MPQ_DIRECTORY", "$DARK_MAGIC_FIRST_CONTENT, $DARK_MAGIC_SECOND_CONTENT")
	contentFS, err := FromEnvironment()
	if err != nil {
		t.Fatal(err)
	}
	if got := contentFS.Layers(); !reflect.DeepEqual(got, []string{"d2legacy", "mpq-0-directory", "mpq-1-directory"}) {
		t.Fatalf("layers = %v", got)
	}
	shared, err := fs.ReadFile(contentFS, "shared.gpl")
	if err != nil || string(shared) != "first" {
		t.Fatalf("shared = %q, %v", shared, err)
	}
	secondOnly, err := fs.ReadFile(contentFS, "second-only.gpl")
	if err != nil || string(secondOnly) != "mounted" {
		t.Fatalf("second-only = %q, %v", secondOnly, err)
	}
}

func TestFromEnvironmentRejectsEmptyMPQDirectoryEntry(t *testing.T) {
	t.Setenv("DARK_MAGIC_MOD_DIRECTORY", "")
	t.Setenv("MPQ_DIRECTORY", t.TempDir()+",")
	if _, err := FromEnvironment(); err == nil || !strings.Contains(err.Error(), "entry 2 is empty") {
		t.Fatalf("error = %v", err)
	}
}

func TestStandardMPQOrderStartsWithLegacyPatchArchive(t *testing.T) {
	if len(standardMPQNames) == 0 || standardMPQNames[0] != "patch_d2.mpq" {
		t.Fatalf("archive priority = %v", standardMPQNames)
	}
}

func TestFromEnvironmentListsRealMPQMapAssets(t *testing.T) {
	directory := os.Getenv("DARK_MAGIC_TEST_MPQ_DIRECTORY")
	if directory == "" {
		t.Skip("set DARK_MAGIC_TEST_MPQ_DIRECTORY to a Diablo II MPQ directory")
	}
	t.Setenv("DARK_MAGIC_MOD_DIRECTORY", "")
	t.Setenv("MPQ_DIRECTORY", directory)
	contentFS, err := FromEnvironment()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		for _, layer := range contentFS.snapshot() {
			_ = Close(layer.FS)
		}
	}()
	if _, err := fs.ReadFile(contentFS, "data/global/excel/sets.txt"); err != nil {
		t.Fatalf("open patch table sets.txt: %v", err)
	}

	for _, suffix := range []string{".dt1", ".ds1"} {
		paths, listErr := contentFS.List("data/global/tiles", suffix)
		if listErr != nil {
			t.Fatalf("list %s: %v", suffix, listErr)
		}
		if len(paths) == 0 {
			t.Fatalf("no %s assets found in mounted MPQs", suffix)
		}
		t.Logf("found %d %s assets", len(paths), suffix)
	}
}
