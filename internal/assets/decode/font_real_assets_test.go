package assetdecode_test

import (
	"bytes"
	"image"
	"image/color"
	"os"
	"testing"

	"github.com/gravestench/dark-magic/internal/assets/decode"
	"github.com/gravestench/dark-magic/internal/content"
)

// This optional test protects the indexed-font/PL2 boundary with legally
// supplied game data. Synthetic tests cover malformed inputs and exact slot
// selection; this test proves the real archive paths compose successfully.
func TestRealFontUsesContextualPL2TextTransforms(t *testing.T) {
	directory := os.Getenv("DARK_MAGIC_TEST_MPQ_DIRECTORY")
	if directory == "" {
		t.Skip("set DARK_MAGIC_TEST_MPQ_DIRECTORY to the Diablo II MPQ directory")
	}
	t.Setenv("MPQ_DIRECTORY", directory)
	assets, err := content.FromEnvironment(content.Layer{Name: "d2legacy", FS: content.D2Legacy()})
	if err != nil {
		t.Fatal(err)
	}
	font, err := assetdecode.LoadBitmapFontWithTransform(
		assets,
		"data/local/FONT/latin/font16.tbl",
		"data/local/FONT/latin/font16.dc6",
		"data/global/Palette/units/pal.dat",
		"data/global/Palette/Sky/Pal.pl2",
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(font.TextFrames) != 13 {
		t.Fatalf("PL2 text transforms = %d, want 13", len(font.TextFrames))
	}
	white, err := font.Render("[white]Hero", color.White, 0, "left")
	if err != nil {
		t.Fatal(err)
	}
	unshifted, err := font.Render("Hero", color.White, 0, "left")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(white.Pix, unshifted.Pix) {
		t.Fatal("real PL2 white run must preserve the palette-authored glyph")
	}
	gold, err := font.Render("[gold]Hero", color.White, 0, "left")
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Equal(white.Pix, gold.Pix) {
		t.Fatal("real PL2 white and gold runs produced identical pixels")
	}
	if opaquePixels(white) == 0 || opaquePixels(gold) == 0 {
		t.Fatal("real PL2 text transform produced an empty glyph image")
	}
}

func opaquePixels(source *image.RGBA) int {
	count := 0
	for offset := 3; offset < len(source.Pix); offset += 4 {
		if source.Pix[offset] != 0 {
			count++
		}
	}
	return count
}
