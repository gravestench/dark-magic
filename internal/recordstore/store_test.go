package recordstore

import (
	"testing"
	"testing/fstest"
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
