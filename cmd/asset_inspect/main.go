package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gravestench/dark-magic/pkg/assetinspect"
	"github.com/gravestench/dark-magic/pkg/services/fileLoader"
)

func main() {
	sourcePath := flag.String("source", "", "directory or MPQ containing the asset")
	assetPath := flag.String("asset", "", "asset path inside the source")
	previewPath := flag.String("preview", "", "optional PNG output path for a DC6 frame")
	direction := flag.Int("direction", 0, "DC6 direction to preview")
	frame := flag.Int("frame", 0, "DC6 frame to preview")
	dt1Paths := flag.String("dt1", "", "comma-separated DT1 paths used to texture a DS1 preview")
	palettePath := flag.String("palette", "", "optional PL2 palette used for a textured DS1 preview")
	flag.Parse()

	if *sourcePath == "" || *assetPath == "" {
		fmt.Fprintln(os.Stderr, "usage: asset_inspect -source <directory-or-mpq> -asset <path>")
		os.Exit(2)
	}

	source := fileLoader.NewSource(*sourcePath)
	filesystem, err := source.Filesystem()
	if err != nil {
		fatal(err)
	}

	report, err := assetinspect.Inspect(filesystem, *assetPath)
	if err != nil {
		fatal(err)
	}

	if err := json.NewEncoder(os.Stdout).Encode(report); err != nil {
		fatal(err)
	}

	if *previewPath != "" {
		var preview []byte
		var err error
		if *dt1Paths != "" && strings.EqualFold(filepath.Ext(*assetPath), ".ds1") {
			preview, err = assetinspect.TexturedDS1Preview(filesystem, *assetPath, strings.Split(*dt1Paths, ","), *palettePath)
		} else {
			preview, err = assetinspect.Preview(filesystem, *assetPath, *direction, *frame)
		}
		if err != nil {
			fatal(err)
		}
		if err := os.WriteFile(*previewPath, preview, 0o644); err != nil {
			fatal(fmt.Errorf("writing preview: %w", err))
		}
	}
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
