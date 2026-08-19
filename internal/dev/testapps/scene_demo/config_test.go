package main

import (
	"flag"
	"io"
	"testing"
)

// TestParseDemoConfigDefaults protects the command-line contract used when the demo starts without explicit options.
func TestParseDemoConfigDefaults(t *testing.T) {
	flagSet := newTestFlagSet()

	actual, err := parseDemoConfig(flagSet, nil)
	if err != nil {
		t.Fatalf("parse default config: %v", err)
	}

	expected := demoConfig{
		seed:     1,
		savePath: "dark-magic-scene.json",
	}
	if actual != expected {
		t.Fatalf("default config = %#v, want %#v", actual, expected)
	}
}

// TestParseDemoConfigExplicitValues ensures every supported flag reaches the domain configuration without
// normalization.
func TestParseDemoConfigExplicitValues(t *testing.T) {
	flagSet := newTestFlagSet()
	arguments := []string{
		"-seed", "42",
		"-save", "save.json",
		"-source", "maps.mpq",
		"-map", "data/global/tiles/map.ds1",
		"-dt1", "floor.dt1,walls.dt1",
		"-palette", "palette.pl2",
	}

	actual, err := parseDemoConfig(flagSet, arguments)
	if err != nil {
		t.Fatalf("parse explicit config: %v", err)
	}

	expected := demoConfig{
		seed:        42,
		savePath:    "save.json",
		sourcePath:  "maps.mpq",
		mapPath:     "data/global/tiles/map.ds1",
		dt1Paths:    "floor.dt1,walls.dt1",
		palettePath: "palette.pl2",
	}
	if actual != expected {
		t.Fatalf("explicit config = %#v, want %#v", actual, expected)
	}
}

// newTestFlagSet isolates flag registration between tests and suppresses expected parser diagnostics from test output.
func newTestFlagSet() *flag.FlagSet {
	flagSet := flag.NewFlagSet("scene_demo", flag.ContinueOnError)
	flagSet.SetOutput(io.Discard)

	return flagSet
}
