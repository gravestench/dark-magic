package assets

import (
	"image"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/gravestench/dark-magic/internal/content"
	d2save "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/save"
)

// TestCacheRendersAllowlistedMPQContentToStableFiles verifies the full archive-to-cache path. It protects both the
// reviewed allowlist boundary and the stable filenames that browser aliases depend on across repeated requests.
func TestCacheRendersAllowlistedMPQContentToStableFiles(t *testing.T) {
	cache := newArchiveBackedCache(t)

	assertAllowlistedImagesRenderStably(t, cache)
	assertAllowlistedFontsRender(t, cache)

	if _, err := cache.renderImage("arbitrary-mpq-path"); !os.IsNotExist(err) {
		t.Fatalf("non-allowlisted asset error = %v", err)
	}
}

// newArchiveBackedCache opens the optional real-game fixture and gives ownership to the test. Centralizing cleanup
// ensures every integration test closes the MPQ source even when construction or a later assertion fails.
func newArchiveBackedCache(t *testing.T) *Cache {
	t.Helper()

	directory := os.Getenv("DARK_MAGIC_TEST_MPQ_DIRECTORY")
	if directory == "" {
		t.Skip("DARK_MAGIC_TEST_MPQ_DIRECTORY is required")
	}

	t.Setenv("MPQ_DIRECTORY", directory)

	source, err := content.FromEnvironment()
	if err != nil {
		t.Fatal(err)
	}
	// The content source may own archive file descriptors, so bind its lifetime to the test before constructing cache.
	t.Cleanup(func() { _ = source.Close() })

	cache, err := New(source, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}

	return cache
}

// assertAllowlistedImagesRenderStably checks dimensions and repeated cache paths for every reviewed image recipe. The
// exact dimensions catch accidentally selected pages or palettes that a simple non-empty decode would miss.
func assertAllowlistedImagesRenderStably(t *testing.T, cache *Cache) {
	t.Helper()

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

		decoded := decodeCachedPNG(t, id, first)
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
}

// assertAllowlistedFontsRender verifies that every font recipe publishes both halves of its cache contract. Metadata
// without an atlas, or an atlas without metadata, is unusable to the browser even when the existing file is non-empty.
func assertAllowlistedFontsRender(t *testing.T, cache *Cache) {
	t.Helper()

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
}

// decodeCachedPNG opens and closes one rendered image around png.Decode. Keeping file ownership here prevents repeated
// assertions from leaking descriptors when an integration fixture contains many images.
func decodeCachedPNG(t *testing.T, id, name string) image.Image {
	t.Helper()

	file, err := os.Open(name)
	if err != nil {
		t.Fatal(err)
	}

	decoded, decodeErr := png.Decode(file)
	closeErr := file.Close()

	if decodeErr != nil {
		t.Fatalf("decode %s: %v", id, decodeErr)
	}

	if closeErr != nil {
		t.Fatalf("close %s: %v", id, closeErr)
	}

	return decoded
}

// TestImageRecipesFollowD2LegacyPresentationPalettes pins the manifest join that supplies portal recipes. The assertion
// ensures the portal follows authored palette changes and keeps the two button states mapped to their intended pages.
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

// TestRosterCacheSharesIdenticalResolvedCharacterVisuals proves that non-visual character data does not enter cache
// identity. Sharing these files is both a storage optimization and a contract of content-addressed roster URLs.
func TestRosterCacheSharesIdenticalResolvedCharacterVisuals(t *testing.T) {
	cache := newArchiveBackedCache(t)

	first, err := cache.PrepareRoster(
		d2save.Character{ID: "one", Name: "One", Class: "Amazon", Level: 1},
	)
	if err != nil {
		t.Fatal(err)
	}

	second, err := cache.PrepareRoster(
		d2save.Character{ID: "two", Name: "Two", Class: "Amazon", Level: 99},
	)
	if err != nil {
		t.Fatal(err)
	}

	if first.ID != second.ID || first.Image != second.Image || first.Metadata != second.Metadata {
		t.Fatalf("identical visuals did not share cache: first=%#v second=%#v", first, second)
	}

	if len(first.Frames) < 2 || first.FrameDurationMS <= 0 {
		t.Fatalf("roster metadata = %#v", first)
	}

	imagePath := filepath.Join(cache.directory, "roster-"+first.ID+".png")

	decoded := decodeCachedPNG(t, "roster image", imagePath)
	if decoded.Bounds().Empty() {
		t.Fatal("decoded roster image has empty bounds")
	}
}
