package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/gravestench/dark-magic/internal/assets/inspect"
	"github.com/gravestench/dark-magic/internal/content"
)

const usageMessage = "usage: asset_inspect -source <directory-or-mpq> -asset <path>"

// commandOptions keeps command-line values together so execution phases do not depend on mutable flag pointers.
type commandOptions struct {
	sourcePath  string
	assetPath   string
	previewPath string
	direction   int
	frame       int
	dt1Paths    string
	palettePath string
}

// main validates required paths before any filesystem work, preserving the command's distinct usage exit code.
func main() {
	options := readCommandOptions()

	if !options.hasRequiredPaths() {
		fmt.Fprintln(os.Stderr, usageMessage)
		os.Exit(2)
	}

	if err := inspectAsset(options, os.Stdout); err != nil {
		fatal(err)
	}
}

// readCommandOptions uses the process flag set so help text and parse failures retain the standard flag behavior.
func readCommandOptions() commandOptions {
	var options commandOptions

	flag.StringVar(&options.sourcePath, "source", "", "directory or MPQ containing the asset")
	flag.StringVar(&options.assetPath, "asset", "", "asset path inside the source")
	flag.StringVar(&options.previewPath, "preview", "", "optional PNG output path for a DC6 frame")
	flag.IntVar(&options.direction, "direction", 0, "DC6 direction to preview")
	flag.IntVar(&options.frame, "frame", 0, "DC6 frame to preview")
	flag.StringVar(&options.dt1Paths, "dt1", "", "comma-separated DT1 paths used to texture a DS1 preview")
	flag.StringVar(&options.palettePath, "palette", "", "optional PL2 palette used for a textured DS1 preview")
	flag.Parse()

	return options
}

// hasRequiredPaths keeps missing-path validation independent from I/O so invalid invocations cannot open a source.
func (options commandOptions) hasRequiredPaths() bool {
	return options.sourcePath != "" && options.assetPath != ""
}

// inspectAsset writes the JSON report before any optional preview, preserving output and failure ordering for callers.
func inspectAsset(options commandOptions, reportOutput io.Writer) error {
	// Preserve command-lifetime ownership: report inspection and optional preview rendering share this source.
	filesystem, err := content.OpenSource(options.sourcePath)
	if err != nil {
		return err
	}

	report, err := assetinspect.Inspect(filesystem, options.assetPath)
	if err != nil {
		return err
	}

	// Publish metadata first so callers retain the successful report even when optional preview rendering fails.
	if err := json.NewEncoder(reportOutput).Encode(report); err != nil {
		return err
	}

	if options.previewPath == "" {
		return nil
	}

	return writePreview(filesystem, options)
}

// fatal reports one terminal error without decoration so existing scripts can continue matching decoder messages.
func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
