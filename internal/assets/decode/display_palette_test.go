package assetdecode

import (
	"image/color"
	"testing"
	"testing/fstest"
)

func TestDisplayPaletteAcceptsTinyRawAndJSONPalettes(t *testing.T) {
	source := fstest.MapFS{
		"bw.pal":    {Data: []byte{0, 0, 0, 255, 255, 255}},
		"mono.pal":  {Data: []byte{3, 2, 1}},
		"warm.json": {Data: []byte(`{"colors":["#102030","#ffeedd"]}`)},
		"cool.gpl":  {Data: []byte("\ufeffGIMP Palette\nName: Cool\nColumns: 2\n# optional names follow RGB\n  1  2  3 Deep Blue\n255\t254\t253\tSnow\n")},
	}
	for name, want := range map[string][]color.RGBA{
		"bw.pal":    {{A: 255}, {R: 255, G: 255, B: 255, A: 255}},
		"mono.pal":  {{R: 1, G: 2, B: 3, A: 255}},
		"warm.json": {{R: 0x10, G: 0x20, B: 0x30, A: 255}, {R: 0xff, G: 0xee, B: 0xdd, A: 255}},
		"cool.gpl":  {{R: 1, G: 2, B: 3, A: 255}, {R: 255, G: 254, B: 253, A: 255}},
	} {
		palette, err := DisplayPalette(source, name)
		if err != nil {
			t.Fatal(err)
		}
		for index, expected := range want {
			if got := color.RGBAModel.Convert(palette[index]).(color.RGBA); got != expected {
				t.Errorf("%s[%d] = %#v, want %#v", name, index, got, expected)
			}
		}
	}
}

func TestDisplayPaletteRejectsMalformedGPL(t *testing.T) {
	t.Parallel()

	for name, data := range map[string]string{
		"header.gpl": "not a palette\n0 0 0\n",
		"empty.gpl":  "GIMP Palette\nName: Empty\n",
		"range.gpl":  "GIMP Palette\n256 0 0 Invalid\n",
		"row.gpl":    "GIMP Palette\n1 2\n",
	} {
		_, err := DisplayPalette(fstest.MapFS{name: {Data: []byte(data)}}, name)
		if err == nil {
			t.Errorf("%s was accepted", name)
		}
	}
}
