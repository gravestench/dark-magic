package main

import (
	"fmt"
	"io"

	"github.com/gravestench/dark-magic/internal/content"
	"github.com/gravestench/dt1"
)

// inspectDT1 probes every layout safely but decodes tile metadata only when the modern format is recognized.
func inspectDT1(source *content.FS, output io.Writer, path string) error {
	file, err := source.Open(path)
	if err != nil {
		return fmt.Errorf("open %q: %w", path, err)
	}
	defer closeReadAsset(file)

	data, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("read %q: %w", path, err)
	}

	header, err := dt1.ProbeBytes(data)
	if err != nil {
		return fmt.Errorf("probe %q: %w", path, err)
	}

	if !writeDT1Header(output, path, header) {
		return nil
	}

	decoded, err := dt1.OpenBytes(data)
	if err != nil {
		return fmt.Errorf("decode %q: %w", path, err)
	}

	writeCatalogf(output, " (%d records)\n", decoded.NumTiles())

	return writeDT1TileMetadata(output, decoded)
}

// writeDT1Header reports probe results and tells the caller whether decoding the body is safe for this layout.
func writeDT1Header(output io.Writer, path string, header dt1.Header) bool {
	writeCatalogf(output, "%s (header %d.%d, %s)", path, header.Version, header.SubVersion, header.Layout)

	if header.Layout == dt1.LayoutModern {
		return true
	}

	// Unknown layouts are still useful research evidence, but interpreting their offsets as modern would be unsafe.
	writeCatalogf(output, " -- body intentionally not decoded\n")

	return false
}

// writeDT1TileMetadata emits records in file order so indexes continue to match decoder lookups and research notes.
func writeDT1TileMetadata(output io.Writer, decoded *dt1.File) error {
	for index := 0; index < decoded.NumTiles(); index++ {
		tile, err := decoded.TileMetadata(index)
		if err != nil {
			return err
		}

		writeCatalogf(
			output,
			"%4d orientation=%2d main=%3d sub=%3d rarity=%3d size=%4dx%4d material=%v\n",
			index,
			tile.Type,
			tile.Style,
			tile.Sequence,
			tile.RarityFrameIndex,
			tile.Width,
			tile.Height,
			tile.MaterialFlags,
		)
	}

	return nil
}
