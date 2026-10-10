package recordstore

import (
	"testing"
	"testing/fstest"

	"github.com/gravestench/dark-magic/internal/content"
)

func TestLoadRowsKeepsOrdinalAndWinningSourceSeparateFromCells(t *testing.T) {
	const path = "data/global/excel/Levels.txt"
	source, err := content.New(content.Layer{Name: "patch", FS: fstest.MapFS{
		path: &fstest.MapFile{Data: []byte("Id\tName\n42\tFirst\n7\tSecond\n")},
	}})
	if err != nil {
		t.Fatal(err)
	}
	pinned, _, err := Pin(source)
	if err != nil {
		t.Fatal(err)
	}

	rows, err := pinned.LoadRows("DATA/GLOBAL/EXCEL/LEVELS.TXT")
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 2 || rows[0].Ordinal != 0 || rows[1].Ordinal != 1 ||
		rows[0].Values["Id"] != "42" || rows[1].Values["Id"] != "7" ||
		rows[0].Table != path || rows[0].Provenance.Layer != "patch" || rows[0].Provenance.Path != path {
		t.Fatalf("rows = %#v", rows)
	}

	rows[0].Values["Id"] = "changed"
	again, err := pinned.LoadRows(path)
	if err != nil || again[0].Values["Id"] != "42" {
		t.Fatalf("cached rows changed: %#v, %v", again, err)
	}
}
