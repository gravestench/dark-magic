package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gravestench/dark-magic/internal/content"
	"github.com/gravestench/dark-magic/pkg/assetinspect"
	darkpaths "github.com/gravestench/dark-magic/pkg/paths"
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

	filesystem, err := content.OpenSource(*sourcePath)
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
		expandedPreview, err := darkpaths.ExpandHost(*previewPath)
		if err != nil {
			fatal(err)
		}
		var preview []byte
		var previewErr error
		if *dt1Paths != "" && strings.EqualFold(filepath.Ext(*assetPath), ".ds1") {
			preview, previewErr = assetinspect.TexturedDS1Preview(filesystem, *assetPath, strings.Split(*dt1Paths, ","), *palettePath)
		} else {
			preview, previewErr = assetinspect.Preview(filesystem, *assetPath, *direction, *frame)
		}
		if previewErr != nil {
			fatal(previewErr)
		}
		if err := os.WriteFile(expandedPreview, preview, 0o644); err != nil {
			fatal(fmt.Errorf("writing preview: %w", err))
		}
	}
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
