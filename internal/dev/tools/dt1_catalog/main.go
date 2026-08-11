// Command dt1_catalog prints the renderer-independent identities contained in
// one or more DT1 files. It is intentionally metadata-only: block pixels are
// never decoded, so it is safe to use while researching large production sets.
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	assetinspect "github.com/gravestench/dark-magic/internal/assets/inspect"
	"github.com/gravestench/dark-magic/internal/content"
	ds1 "github.com/gravestench/ds1/pkg"
	"github.com/gravestench/dt1"
)

func main() {
	paths := flag.String("assets", "", "comma-separated DT1 asset paths")
	stamps := flag.String("stamps", "", "comma-separated DS1 stamps whose tile identities should be printed")
	flag.Parse()
	if strings.TrimSpace(*paths) == "" && strings.TrimSpace(*stamps) == "" {
		fmt.Fprintln(os.Stderr, "dt1_catalog: -assets or -stamps is required")
		os.Exit(2)
	}
	source, err := content.FromEnvironment()
	if err != nil {
		fatal(err)
	}
	for _, path := range strings.Split(*paths, ",") {
		if strings.TrimSpace(path) == "" {
			continue
		}
		if err := inspect(source, strings.TrimSpace(path)); err != nil {
			fatal(err)
		}
	}
	for _, path := range strings.Split(*stamps, ",") {
		if strings.TrimSpace(path) == "" {
			continue
		}
		if err := inspectStamp(source, strings.TrimSpace(path)); err != nil {
			fatal(err)
		}
	}
}

func inspectStamp(source *content.FS, path string) error {
	file, err := source.Open(path)
	if err != nil {
		return fmt.Errorf("open %q: %w", path, err)
	}
	defer file.Close()
	stamp, err := ds1.FromReader(file)
	if err != nil {
		return fmt.Errorf("decode %q: %w", path, err)
	}
	fmt.Printf("%s (%dx%d)\n", path, stamp.Width, stamp.Height)
	dependencies, err := assetinspect.DS1TilePaths(source, path)
	if err != nil {
		return err
	}
	for _, dependency := range dependencies {
		fmt.Printf("  library %s\n", dependency)
	}
	for y, row := range stamp.Tiles {
		for x, record := range row {
			for layer, floor := range record.Floors {
				if !floor.Hidden && floor.Prop1 != 0 {
					fmt.Printf("  %3d,%3d floor[%d] main=%3d sub=%3d\n", x, y, layer, floor.Style, floor.Sequence)
				}
			}
			for layer, wall := range record.Walls {
				if !wall.Hidden && wall.Prop1 != 0 {
					fmt.Printf("  %3d,%3d wall[%d] orientation=%2d main=%3d sub=%3d\n", x, y, layer, wall.Type, wall.Style, wall.Sequence)
				}
			}
		}
	}
	return nil
}

func inspect(source *content.FS, path string) error {
	file, err := source.Open(path)
	if err != nil {
		return fmt.Errorf("open %q: %w", path, err)
	}
	defer file.Close()
	data, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("read %q: %w", path, err)
	}
	header, err := dt1.ProbeBytes(data)
	if err != nil {
		return fmt.Errorf("probe %q: %w", path, err)
	}
	fmt.Printf("%s (header %d.%d, %s)", path, header.Version, header.SubVersion, header.Layout)
	if header.Layout != dt1.LayoutModern {
		fmt.Println(" -- body intentionally not decoded")
		return nil
	}
	decoded, err := dt1.OpenBytes(data)
	if err != nil {
		return fmt.Errorf("decode %q: %w", path, err)
	}
	fmt.Printf(" (%d records)\n", decoded.NumTiles())
	for index := 0; index < decoded.NumTiles(); index++ {
		tile, err := decoded.TileMetadata(index)
		if err != nil {
			return err
		}
		fmt.Printf("%4d orientation=%2d main=%3d sub=%3d rarity=%3d size=%4dx%4d material=%v\n",
			index, tile.Type, tile.Style, tile.Sequence, tile.RarityFrameIndex,
			tile.Width, tile.Height, tile.MaterialFlags)
	}
	return nil
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "dt1_catalog:", err)
	os.Exit(1)
}
