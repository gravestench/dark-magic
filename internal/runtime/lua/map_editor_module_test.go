package modruntime

import (
	"bytes"
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"

	"github.com/gravestench/dark-magic/internal/mapeditor"
	lua "github.com/yuin/gopher-lua"
)

// TestMapEditorScreenLuaParses catches syntax drift across every namespaced editor script.
func TestMapEditorScreenLuaParses(t *testing.T) {
	state := lua.NewState()
	defer state.Close()
	root := "../../content/ds1editor/lua"
	if err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil || entry.IsDir() || filepath.Ext(path) != ".lua" {
			return walkErr
		}
		source, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if bytes.Contains(source, []byte("set_scale(math.min")) {
			t.Errorf("%s passes one computed argument to set_scale; pass explicit X and Y values", path)
		}
		if bytes.Contains(source, []byte("text.set(self.inspector,")) {
			t.Errorf("%s writes through the retired inspector node instead of the active selected view", path)
		}
		if bytes.Contains(source, []byte("left=x - 512, top=y - 1024")) {
			t.Errorf("%s restored map-wide dirty rectangles for one-cell edits", path)
		}
		function, err := state.Load(bytes.NewReader(source), path)
		if err != nil {
			return err
		}
		if function == nil {
			t.Fatalf("%s compiled without a function", path)
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
}

// TestMapEditorModulePaintsAndPreviewsUnsavedDocument exercises the Lua-to-codec-to-world boundary in memory.
func TestMapEditorModulePaintsAndPreviewsUnsavedDocument(t *testing.T) {
	document, err := mapeditor.New(mapeditor.NewConfig{Width: 2, Height: 2})
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := document.Encode()
	if err != nil {
		t.Fatal(err)
	}
	runtime := New()
	source := fstest.MapFS{"map.ds1": &fstest.MapFile{Data: encoded}}
	if err := runtime.RegisterModule(MapEditorModule(source, nil)); err != nil {
		t.Fatal(err)
	}
	if err := runtime.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	defer runtime.Stop(context.Background())
	script := fstest.MapFS{"test.lua": &fstest.MapFile{Data: []byte(`
local editor = require("engine.map_editor/v1")
local opened, err = editor.open("map.ds1")
assert(opened and not err and opened.dirty == false)
assert(editor.begin_stroke("floor", 0, {style=7, sequence=11, orientation=0}))
assert(editor.paint(1, 1))
assert(editor.end_stroke())
local cell = editor.cell(1, 1)
assert(cell.floors[1].style == 7 and cell.floors[1].sequence == 11)
local preview = editor.preview()
assert(preview:dimensions().width_tiles == 2)
assert(editor.undo())
assert(editor.redo())
`)}}
	if err := runtime.Execute(context.Background(), script, "test.lua"); err != nil {
		t.Fatal(err)
	}
}
