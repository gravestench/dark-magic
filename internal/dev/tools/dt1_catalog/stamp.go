package main

import (
	"fmt"
	"io"

	assetinspect "github.com/gravestench/dark-magic/internal/assets/inspect"
	"github.com/gravestench/dark-magic/internal/content"
	ds1 "github.com/gravestench/ds1/pkg"
)

// inspectStamp decodes one DS1 stamp and prints its library dependencies before its visible tile identities.
func inspectStamp(source *content.FS, output io.Writer, path string) error {
	file, err := source.Open(path)
	if err != nil {
		return fmt.Errorf("open %q: %w", path, err)
	}
	defer closeReadAsset(file)

	stamp, err := ds1.FromReader(file)
	if err != nil {
		return fmt.Errorf("decode %q: %w", path, err)
	}

	writeCatalogf(output, "%s (%dx%d)\n", path, stamp.Width, stamp.Height)

	dependencies, err := assetinspect.DS1TilePaths(source, path)
	if err != nil {
		return err
	}

	for _, dependency := range dependencies {
		writeCatalogf(output, "  library %s\n", dependency)
	}

	writeStampTileIdentities(output, stamp.Tiles)

	return nil
}

// writeStampTileIdentities walks rows and columns deterministically so printed coordinates retain DS1 storage order.
func writeStampTileIdentities(output io.Writer, tiles [][]ds1.TileRecord) {
	for y, row := range tiles {
		for x, record := range row {
			writeTileRecordIdentities(output, x, y, record)
		}
	}
}

// writeTileRecordIdentities prints only visible, populated floors and walls while preserving layer order.
func writeTileRecordIdentities(output io.Writer, x, y int, record ds1.TileRecord) {
	for layer, floor := range record.Floors {
		// Hidden and empty layers cannot resolve to visible catalog tiles, so keep them out of the research output.
		if floor.Hidden || floor.Prop1 == 0 {
			continue
		}

		writeCatalogf(
			output,
			"  %3d,%3d floor[%d] main=%3d sub=%3d\n",
			x,
			y,
			layer,
			floor.Style,
			floor.Sequence,
		)
	}

	for layer, wall := range record.Walls {
		// Apply the same visibility rule to walls without disturbing their independent layer indexes.
		if wall.Hidden || wall.Prop1 == 0 {
			continue
		}

		writeCatalogf(
			output,
			"  %3d,%3d wall[%d] orientation=%2d main=%3d sub=%3d\n",
			x,
			y,
			layer,
			wall.Type,
			wall.Style,
			wall.Sequence,
		)
	}
}
