package main

import (
	"errors"
	"io/fs"
	"strings"

	"github.com/gravestench/dark-magic/internal/assets/inspect"
	"github.com/gravestench/dark-magic/internal/content"
)

// loadMapPreview validates paired map inputs before opening content, preventing an ambiguous source or DS1 lookup.
func loadMapPreview(config demoConfig) ([]byte, error) {
	if !config.requestsMapPreview() {
		return nil, nil
	}

	if !config.hasCompleteMapSelection() {
		return nil, errors.New("both -source and -map are required")
	}

	filesystem, err := content.OpenSource(config.sourcePath)
	if err != nil {
		return nil, err
	}

	// The short-lived demo has historically left archive lifetime to process teardown; keep that ownership contract
	// intact.
	return renderMapPreview(filesystem, config)
}

// renderMapPreview selects structural or textured rendering without changing the comma-separated DT1 path semantics.
func renderMapPreview(filesystem fs.FS, config demoConfig) ([]byte, error) {
	if config.dt1Paths == "" {
		return assetinspect.DS1Preview(filesystem, config.mapPath)
	}

	return assetinspect.TexturedDS1Preview(
		filesystem,
		config.mapPath,
		strings.Split(config.dt1Paths, ","),
		config.palettePath,
	)
}
