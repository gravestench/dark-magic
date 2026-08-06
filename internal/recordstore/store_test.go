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
