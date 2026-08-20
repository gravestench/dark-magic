package content_test

import (
	"image/color"
	"io/fs"
	"path/filepath"
	"testing"

	assetdecode "github.com/gravestench/dark-magic/internal/assets/decode"
	"github.com/gravestench/dark-magic/internal/content"
	"github.com/gravestench/dark-magic/internal/modcache"
	"github.com/gravestench/dc6"
	"github.com/yuin/gopher-lua/parse"
)

// TestDS1EditorPackageOwnsNamespacedLuaAndDC6Assets verifies the standalone package is self-contained and decodable.
func TestDS1EditorPackageOwnsNamespacedLuaAndDC6Assets(t *testing.T) {
	source := content.DS1Editor()
	manifest, err := modcache.ReadManifest(source)
	if err != nil {
		t.Fatal(err)
	}
	packageFS, err := modcache.NewPackageFS(manifest, source)
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{
		"mods/ds1editor/lua/ds1editor/screens/map_editor.lua",
		"mods/ds1editor/lua/ds1editor/ui/composition.lua",
		"darkmagic/ds1-editor/ui/assets.json",
		"darkmagic/ds1-editor/ui/palette.dat",
		"darkmagic/ds1-editor/ui/fonts/manifest.json",
		"darkmagic/ds1-editor/ui/fonts/large.tbl",
		"darkmagic/ds1-editor/ui/fonts/large.dc6",
		"darkmagic/ds1-editor/ui/fonts/medium.tbl",
		"darkmagic/ds1-editor/ui/fonts/medium.dc6",
		"darkmagic/ds1-editor/ui/fonts/small.tbl",
		"darkmagic/ds1-editor/ui/fonts/small.dc6",
		"darkmagic/ds1-editor/ui/fonts/very_small.tbl",
		"darkmagic/ds1-editor/ui/fonts/very_small.dc6",
	} {
		if _, err := fs.Stat(packageFS, path); err != nil {
			t.Fatalf("stat %s: %v", path, err)
		}
	}
	encoded, err := fs.ReadFile(packageFS, "darkmagic/ds1-editor/ui/authoring.dc6")
	if err != nil {
		t.Fatal(err)
	}
	asset, err := dc6.FromBytes(encoded)
	if err != nil {
		t.Fatal(err)
	}
	if len(asset.Directions) != 1 || len(asset.Directions[0].Frames) != 16 || asset.Directions[0].Frames[0].Width != 32 {
		t.Fatalf("authoring sheet shape = %#v", asset.Directions)
	}
	chromeData, err := fs.ReadFile(packageFS, "darkmagic/ds1-editor/ui/chrome.dc6")
	if err != nil {
		t.Fatal(err)
	}
	chrome, err := dc6.FromBytes(chromeData)
	if err != nil {
		t.Fatal(err)
	}
	if len(chrome.Directions) != 1 || len(chrome.Directions[0].Frames) != 16 ||
		chrome.Directions[0].Frames[0].Width != 16 {
		t.Fatalf("chrome sheet shape = %#v", chrome.Directions)
	}

	for _, name := range []string{"large", "medium", "small", "very_small"} {
		font, err := assetdecode.LoadBitmapFont(
			packageFS,
			"darkmagic/ds1-editor/ui/fonts/"+name+".tbl",
			"darkmagic/ds1-editor/ui/fonts/"+name+".dc6",
			"darkmagic/ds1-editor/ui/palette.dat",
		)
		if err != nil {
			t.Fatalf("load %s font: %v", name, err)
		}
		if _, err := font.Render("Map 012 ✓", color.White, 0, "left"); err != nil {
			t.Fatalf("render %s font: %v", name, err)
		}
	}

	err = fs.WalkDir(packageFS, "mods/ds1editor/lua", func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil || entry.IsDir() || filepath.Ext(path) != ".lua" {
			return walkErr
		}
		file, openErr := packageFS.Open(path)
		if openErr != nil {
			return openErr
		}
		_, parseErr := parse.Parse(file, path)
		closeErr := file.Close()
		if parseErr != nil {
			return parseErr
		}
		return closeErr
	})
	if err != nil {
		t.Fatalf("parse editor Lua: %v", err)
	}
}
