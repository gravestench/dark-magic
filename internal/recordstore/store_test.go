package recordstore

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/gravestench/dark-magic/internal/content"
)

func TestStoreLoadsCachesClonesAndInvalidatesTSV(t *testing.T) {
	t.Parallel()

	source := fstest.MapFS{"records.txt": &fstest.MapFile{Data: []byte("\ufeffcode\tname\na\tAlpha\n")}}
	store := New(source)
	rows, err := store.Load("records.txt")
	if err != nil {
		t.Fatal(err)
	}
	if rows[0]["code"] != "a" || rows[0]["name"] != "Alpha" || !store.Loaded("records.txt") {
		t.Fatalf("rows = %#v", rows)
	}
	rows[0]["name"] = "mutated"
	again, err := store.Load("records.txt")
	if err != nil {
		t.Fatal(err)
	}
	if again[0]["name"] != "Alpha" {
		t.Fatalf("cached row was mutated: %#v", again)
	}
	store.Invalidate("records.txt")
	if store.Loaded("records.txt") {
		t.Fatal("table remains loaded after invalidation")
	}
}

func TestStorePreservesDuplicateShippedColumns(t *testing.T) {
	t.Parallel()

	store := New(fstest.MapFS{"armor.txt": &fstest.MapFile{Data: []byte("code\tmindam\tmindam\ncap\t1\t2\n")}})
	rows, err := store.Load("armor.txt")
	if err != nil {
		t.Fatal(err)
	}
	if rows[0]["mindam"] != "1" || rows[0]["mindam#2"] != "2" {
		t.Fatalf("duplicate columns were not preserved deterministically: %#v", rows)
	}
}

func TestStorePreservesUnnamedShippedColumns(t *testing.T) {
	t.Parallel()

	store := New(fstest.MapFS{"weapons.txt": &fstest.MapFile{Data: []byte("code\t\ncax\tunused\n")}})
	rows, err := store.Load("weapons.txt")
	if err != nil {
		t.Fatal(err)
	}
	if rows[0]["code"] != "cax" || rows[0]["#unnamed-2"] != "unused" {
		t.Fatalf("unnamed column was not preserved deterministically: %#v", rows)
	}
}

func TestStoreLogsEachLoadedGenerationWithProvenance(t *testing.T) {
	t.Parallel()

	source, err := content.New(content.Layer{Name: "patch_d2.mpq", FS: fstest.MapFS{
		"data/global/excel/armor.txt": &fstest.MapFile{Data: []byte("code\ncap\n")},
	}})
	if err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	store := New(source)
	store.SetLogger(slog.New(slog.NewJSONHandler(&output, nil)))
	for range 2 {
		if _, err := store.Load("data/global/excel/armor.txt"); err != nil {
			t.Fatal(err)
		}
	}
	logged := output.String()
	for _, expected := range []string{`"msg":"loaded records"`, `"table":"data/global/excel/armor.txt"`, `"records":1`, `"source":"patch_d2.mpq"`, `"source_path":"data/global/excel/armor.txt"`} {
		if !strings.Contains(logged, expected) {
			t.Errorf("load log %q does not contain %q", logged, expected)
		}
	}
	if strings.Count(logged, `"msg":"loaded records"`) != 1 {
		t.Fatalf("cache hit produced another load event: %q", logged)
	}
}
