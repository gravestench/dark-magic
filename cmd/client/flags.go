package main

import (
	"flag"
	"os"
	"strconv"
	"strings"
)

// clientFlags holds pointers populated by the standard flag package.
// It stays private because only this command translates flags into clientConfig.
type clientFlags struct {
	logLevel            *string
	profileDirectory    *string
	profileScenes       *string
	captureDirectory    *string
	captureScenes       *string
	captureSettle       *int
	startScene          *string
	startOverlays       *string
	fixtureCharacters   *int
	fixtureWorldLevel   *int
	fixtureWorldSpawn   *string
	fixturePointerMove  *bool
	outputPalette       *string
	viewportFit         *string
	fullscreen          *bool
	nativeAudio         *bool
	presentationProfile *string
	mods                *string
}

// parseClientConfig registers the public command-line interface and snapshots its values.
func parseClientConfig(defaultEnvironmentPath string) (clientConfig, error) {
	flags := registerClientFlags(defaultEnvironmentPath)
	flag.Parse()
	return flags.config()
}

// registerClientFlags groups related flag families so the complete interface is easy to scan.
func registerClientFlags(defaultEnvironmentPath string) clientFlags {
	_ = flag.String(
		"env-file",
		defaultEnvironmentPath,
		"environment file (overrides the default client.env selection)",
	)
	flags := clientFlags{}
	registerRuntimeFlags(&flags)
	registerCaptureFlags(&flags)
	registerFixtureFlags(&flags)
	registerDisplayFlags(&flags)
	return flags
}

// registerRuntimeFlags defines process-wide logging, profiling, and mod selection.
func registerRuntimeFlags(flags *clientFlags) {
	flags.logLevel = flag.String(
		"log-level",
		environmentDefault("DARK_MAGIC_LOG_LEVEL", "info"),
		"log verbosity: trace, debug, info, warn, or error",
	)
	flags.profileDirectory = flag.String(
		"profile-dir",
		os.Getenv("DARK_MAGIC_PROFILE_DIR"),
		"capture CPU and heap profiles plus PDF reports in this directory",
	)
	flags.profileScenes = flag.String(
		"profile-scenes",
		os.Getenv("DARK_MAGIC_PROFILE_SCENES"),
		"comma-separated scene IDs (or all) for per-scene CPU and heap reports",
	)
	flags.mods = flag.String(
		"mods",
		os.Getenv("DARK_MAGIC_MODS"),
		"temporary comma-separated extension IDs, or 'none' for vanilla d2legacy",
	)
}

// registerCaptureFlags defines screenshot capture and scene-selection controls.
func registerCaptureFlags(flags *clientFlags) {
	flags.captureDirectory = flag.String(
		"capture-dir",
		os.Getenv("DARK_MAGIC_CAPTURE_DIR"),
		"write local scene screenshots and report.json to this directory",
	)
	flags.captureScenes = flag.String(
		"capture-scenes",
		os.Getenv("DARK_MAGIC_CAPTURE_SCENES"),
		"comma-separated scene IDs to capture (defaults to loading,title)",
	)
	flags.captureSettle = flag.Int(
		"capture-settle-frames",
		10,
		"stable frames to wait before capturing a scene",
	)
	flags.startScene = flag.String(
		"start-scene",
		os.Getenv("DARK_MAGIC_START_SCENE"),
		"development-only scene ID to enter after boot",
	)
	flags.startOverlays = flag.String(
		"start-overlays",
		os.Getenv("DARK_MAGIC_START_OVERLAYS"),
		"development-only comma-separated overlays to open above the start scene",
	)
}

// registerFixtureFlags defines deterministic development-world controls.
func registerFixtureFlags(flags *clientFlags) {
	flags.fixtureCharacters = flag.Int(
		"fixture-characters",
		0,
		"development-only number of in-memory characters to create",
	)
	flags.fixtureWorldLevel = flag.Int(
		"fixture-world-level",
		0,
		"development-only authoritative level for the selected fixture character",
	)
	flags.fixtureWorldSpawn = flag.String(
		"fixture-world-spawn",
		"entry",
		"development-only fixture spawn: entry or seam",
	)
	flags.fixturePointerMove = flag.Bool(
		"fixture-pointer-move",
		false,
		"development-only click-to-move acceptance before capture",
	)
}

// registerDisplayFlags defines presentation, window, and audio controls.
func registerDisplayFlags(flags *clientFlags) {
	flags.outputPalette = flag.String(
		"output-palette",
		os.Getenv("DARK_MAGIC_OUTPUT_PALETTE"),
		"quantize the final display through this mounted pal.dat asset",
	)
	flags.viewportFit = flag.String(
		"viewport-fit",
		environmentDefault("DARK_MAGIC_VIEWPORT_FIT", "contain"),
		"game viewport fit: contain or stretch",
	)
	fullscreenDefault, _ := strconv.ParseBool(environmentDefault("DARK_MAGIC_FULLSCREEN", "false"))
	flags.fullscreen = flag.Bool("fullscreen", fullscreenDefault, "use a maximized borderless window")
	flags.nativeAudio = flag.Bool("native-audio", true, "enable the selected backend's native audio adapter")
	flags.presentationProfile = flag.String(
		"presentation-profile",
		os.Getenv("DARK_MAGIC_PRESENTATION_PROFILE"),
		"manifest-owned presentation profile ID",
	)
}

// config converts parsed flag pointers into an immutable command configuration.
func (flags clientFlags) config() (clientConfig, error) {
	logLevel, err := parseLogLevel(*flags.logLevel)
	if err != nil {
		return clientConfig{}, err
	}
	return clientConfig{
		profileDirectory:      *flags.profileDirectory,
		profileScenes:         *flags.profileScenes,
		captureDirectory:      *flags.captureDirectory,
		captureScenes:         *flags.captureScenes,
		captureSettle:         *flags.captureSettle,
		startScene:            *flags.startScene,
		startOverlays:         *flags.startOverlays,
		fixtureCharacters:     *flags.fixtureCharacters,
		fixtureWorldLevel:     *flags.fixtureWorldLevel,
		fixtureWorldSpawn:     *flags.fixtureWorldSpawn,
		fixturePointerMove:    *flags.fixturePointerMove,
		outputPalette:         *flags.outputPalette,
		viewportFit:           *flags.viewportFit,
		fullscreen:            *flags.fullscreen,
		nativeAudio:           *flags.nativeAudio,
		presentationProfileID: *flags.presentationProfile,
		mods:                  *flags.mods,
		logLevel:              logLevel,
	}, nil
}

// environmentDefault returns a non-empty environment value or the supplied fallback.
func environmentDefault(name, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(name)); value != "" {
		return value
	}
	return fallback
}
