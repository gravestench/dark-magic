package assetinspect

import (
	"fmt"
	"image"
	"image/color"
	"io"
	"io/fs"

	"github.com/gravestench/ds1"
	"github.com/gravestench/pl2"
)

// TexturedDS1Image composes a DS1 stamp using matching DT1 tile graphics.
// Missing definitions retain a structural diamond so incomplete tileset lists
// remain visible. Callers receive the image directly and own its later encoding.
func TexturedDS1Image(
	source fs.FS,
	ds1Path string,
	dt1Paths []string,
	palettePath string,
) (image.Image, error) {
	stamp, err := readTexturedDS1Stamp(source, ds1Path)
	if err != nil {
		return nil, err
	}

	palette, err := readTexturedDS1Palette(source, palettePath)
	if err != nil {
		return nil, err
	}

	needed := texturedDS1TileKeys(stamp)

	lookup, err := loadNeededDT1Tiles(source, dt1Paths, palette, needed)
	if err != nil {
		return nil, err
	}

	return renderTexturedDS1(stamp, lookup)
}

// TexturedDS1Preview encodes the composed image for command-line inspection
// tools while leaving runtime callers free to use TexturedDS1Image directly.
func TexturedDS1Preview(
	source fs.FS,
	ds1Path string,
	dt1Paths []string,
	palettePath string,
) ([]byte, error) {
	preview, err := TexturedDS1Image(source, ds1Path, dt1Paths, palettePath)
	if err != nil {
		return nil, err
	}

	return encodePreviewPNG(preview, "encoding textured DS1 preview")
}

// readTexturedDS1Stamp confines file ownership to loading and preserves the
// textured preview API's existing distinction between read and decode failures.
func readTexturedDS1Stamp(source fs.FS, ds1Path string) (*ds1.DS1, error) {
	stampFile, err := source.Open(ds1Path)
	if err != nil {
		return nil, fmt.Errorf("opening DS1 asset %q: %w", ds1Path, err)
	}

	stampData, readErr := io.ReadAll(stampFile)
	_ = stampFile.Close()

	if readErr != nil {
		return nil, readErr
	}

	stamp, err := ds1.FromBytes(stampData)
	if err != nil {
		return nil, fmt.Errorf("decoding DS1 asset %q: %w", ds1Path, err)
	}

	return stamp, nil
}

// readTexturedDS1Palette loads an optional palette before DT1 decoding because
// decoded tile images must receive their colors from the indexed source data.
func readTexturedDS1Palette(source fs.FS, palettePath string) (color.Palette, error) {
	if palettePath == "" {
		return nil, nil
	}

	file, err := source.Open(palettePath)
	if err != nil {
		return nil, fmt.Errorf("opening PL2 asset %q: %w", palettePath, err)
	}

	data, readErr := io.ReadAll(file)
	_ = file.Close()

	if readErr != nil {
		return nil, fmt.Errorf("reading PL2 asset %q: %w", palettePath, readErr)
	}

	paletteData, err := pl2.FromBytes(data)
	if err != nil {
		return nil, fmt.Errorf("decoding PL2 asset %q: %w", palettePath, err)
	}

	return paletteData.BasePalette, nil
}
