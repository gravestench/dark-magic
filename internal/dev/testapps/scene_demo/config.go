package main

import "flag"

// demoConfig retains raw flag values so path expansion and comma splitting occur at their historical use sites.
type demoConfig struct {
	seed        uint64
	savePath    string
	sourcePath  string
	mapPath     string
	dt1Paths    string
	palettePath string
}

// parseDemoConfig keeps flag registration in one place so tests and the command exercise identical defaults and names.
func parseDemoConfig(flagSet *flag.FlagSet, arguments []string) (demoConfig, error) {
	var config demoConfig

	flagSet.Uint64Var(&config.seed, "seed", 1, "deterministic scene seed")
	flagSet.StringVar(&config.savePath, "save", "dark-magic-scene.json", "scene save path")
	flagSet.StringVar(&config.sourcePath, "source", "", "optional directory or MPQ containing a DS1 map")
	flagSet.StringVar(&config.mapPath, "map", "", "optional DS1 path inside source")
	flagSet.StringVar(&config.dt1Paths, "dt1", "", "comma-separated DT1 paths used to texture the DS1 map")
	flagSet.StringVar(&config.palettePath, "palette", "", "optional PL2 palette for DT1 tiles")

	if err := flagSet.Parse(arguments); err != nil {
		return demoConfig{}, err
	}

	return config, nil
}

// requestsMapPreview detects either half of the paired map selection so incomplete input remains an explicit error.
func (config demoConfig) requestsMapPreview() bool {
	return config.sourcePath != "" || config.mapPath != ""
}

// hasCompleteMapSelection preserves the invariant that a DS1 path is interpreted only within an explicit content
// source.
func (config demoConfig) hasCompleteMapSelection() bool {
	return config.sourcePath != "" && config.mapPath != ""
}

// mapLabel returns the same fallback shown for the generated world when no DS1 map was requested.
func (config demoConfig) mapLabel() string {
	if config.mapPath != "" {
		return config.mapPath
	}

	return "generated grid"
}
