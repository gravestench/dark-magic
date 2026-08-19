package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/gravestench/dark-magic/internal/assets/inspect"
	darkpaths "github.com/gravestench/dark-magic/internal/paths"
)

// writePreview renders one diagnostic image and persists it with the command's stable file permissions.
func writePreview(filesystem fs.FS, options commandOptions) error {
	// Resolve the host destination first so an invalid alias takes precedence over any decoder failure.
	expandedPreview, err := darkpaths.ExpandHost(options.previewPath)
	if err != nil {
		return err
	}

	preview, err := renderPreview(filesystem, options)
	if err != nil {
		return err
	}

	if err := os.WriteFile(expandedPreview, preview, 0o644); err != nil {
		return fmt.Errorf("writing preview: %w", err)
	}

	return nil
}

// renderPreview routes an asset to the preview mode selected by the command-line contract.
func renderPreview(filesystem fs.FS, options commandOptions) ([]byte, error) {
	// Explicit DT1 paths are the compatibility opt-in for textured DS1 output; never infer them for the caller.
	if options.usesTexturedDS1Preview() {
		return assetinspect.TexturedDS1Preview(
			filesystem,
			options.assetPath,
			strings.Split(options.dt1Paths, ","),
			options.palettePath,
		)
	}

	return assetinspect.Preview(filesystem, options.assetPath, options.direction, options.frame)
}

// usesTexturedDS1Preview requires both a DS1 extension and DT1 paths so all other assets keep generic preview errors.
func (options commandOptions) usesTexturedDS1Preview() bool {
	return options.dt1Paths != "" && strings.EqualFold(filepath.Ext(options.assetPath), ".ds1")
}
