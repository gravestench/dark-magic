package assets

import (
	"image"
	"image/png"
	"os"
	"testing"

	"github.com/gravestench/dark-magic/internal/content"
	d2save "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/save"
)

func TestCacheRendersAllowlistedMPQContentToStableFiles(t *testing.T) {
	directory := os.Getenv("DARK_MAGIC_TEST_MPQ_DIRECTORY")
	if directory == "" {
		t.Skip("DARK_MAGIC_TEST_MPQ_DIRECTORY is required")
	}
	t.Setenv("MPQ_DIRECTORY", directory)
	source, err := content.FromEnvironment()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = source.Close() })
	cache, err := New(source, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}

	wantBounds := map[string]image.Rectangle{
		"background":     image.Rect(0, 0, 800, 600),
		"dialog":         image.Rect(0, 0, 340, 224),
		"textbox":        image.Rect(0, 0, 300, 26),
		"button":         image.Rect(0, 0, 272, 35),
		"button-pressed": image.Rect(0, 0, 272, 35),
	}
	for id := range cache.images {
		first, err := cache.renderImage(id)
		if err != nil {
			t.Fatalf("render %s: %v", id, err)
		}
		file, err := os.Open(first)
		if err != nil {
			t.Fatal(err)
		}
		decoded, decodeErr := png.Decode(file)
		_ = file.Close()
		if decodeErr != nil {
			t.Fatalf("decode %s: %v", id, decodeErr)
		}
		if decoded.Bounds().Empty() {
			t.Fatalf("decode %s: empty bounds", id)
		}
		if decoded.Bounds() != wantBounds[id] {
			t.Fatalf("decode %s bounds = %v, want %v", id, decoded.Bounds(), wantBounds[id])
		}
		second, err := cache.renderImage(id)
		if err != nil || second != first {
			t.Fatalf("unstable cache path for %s: %q then %q, %v", id, first, second, err)
		}
	}

	for id := range fontAllowlist {
		atlas, metadata, err := cache.renderFont(id)
		if err != nil {
			t.Fatalf("render font %s: %v", id, err)
		}
		for _, path := range []string{atlas, metadata} {
			if info, err := os.Stat(path); err != nil || info.Size() == 0 {
				t.Fatalf("font cache %s: info=%v error=%v", path, info, err)
			}
		}
	}
	if _, err := cache.renderImage("arbitrary-mpq-path"); !os.IsNotExist(err) {
		t.Fatalf("non-allowlisted asset error = %v", err)
	}
}

func TestImageRecipesFollowD2LegacyPresentationPalettes(t *testing.T) {
	specs, err := loadImageSpecs(content.D2Legacy())
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]string{
		"background": "data/global/Palette/sky/pal.dat",
		"dialog":     "data/global/Palette/units/pal.dat",
		"textbox":    "data/global/Palette/units/pal.dat",
		"button":     "data/global/Palette/units/pal.dat",
	}
	for id, palette := range want {
		if specs[id].palette != palette {
			t.Errorf("%s palette = %q, want %q", id, specs[id].palette, palette)
		}
	}
	if specs["button"].page != 0 || specs["button-pressed"].page != 1 {
		t.Fatalf("button states = up page %d, down page %d", specs["button"].page, specs["button-pressed"].page)
	}
}

func TestRosterCacheSharesIdenticalResolvedCharacterVisuals(t *testing.T) {
	directory := os.Getenv("DARK_MAGIC_TEST_MPQ_DIRECTORY")
	if directory == "" {
		t.Skip("DARK_MAGIC_TEST_MPQ_DIRECTORY is required")
	}
	t.Setenv("MPQ_DIRECTORY", directory)
	source, err := content.FromEnvironment()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = source.Close() })
	cache, err := New(source, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}

	first, err := cache.PrepareRoster(d2save.Character{ID: "one", Name: "One", Class: "Amazon", Level: 1})
	if err != nil {
		t.Fatal(err)
	}
	second, err := cache.PrepareRoster(d2save.Character{ID: "two", Name: "Two", Class: "Amazon", Level: 99})
	if err != nil {
		t.Fatal(err)
	}
	if first.ID != second.ID || first.Image != second.Image || first.Metadata != second.Metadata {
		t.Fatalf("identical visuals did not share cache: first=%#v second=%#v", first, second)
	}
	if len(first.Frames) < 2 || first.FrameDurationMS <= 0 {
		t.Fatalf("roster metadata = %#v", first)
	}
	file, err := os.Open(cache.directory + "/roster-" + first.ID + ".png")
	if err != nil {
		t.Fatal(err)
	}
	decoded, decodeErr := png.Decode(file)
	_ = file.Close()
	if decodeErr != nil {
		t.Fatalf("decode roster image: %v", decodeErr)
	}
	if decoded.Bounds().Empty() {
		t.Fatal("decoded roster image has empty bounds")
	}
}
