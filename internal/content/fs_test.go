package content

import (
	"archive/zip"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
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

func TestShimContainsBoot(t *testing.T) {
	t.Parallel()

	data, err := fs.ReadFile(Shim(), "boot.lua")
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
}
