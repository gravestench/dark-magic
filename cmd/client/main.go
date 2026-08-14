// Command darkmagic is the native client composition root. It parses process
// policy, constructs owned capabilities, and translates final errors to exit
// status; game rules and reusable behavior belong under internal packages.
package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/gravestench/dark-magic/internal/app/clientapp"
	"github.com/gravestench/dark-magic/internal/content"
	"github.com/gravestench/dark-magic/internal/dev/capture"
	"github.com/gravestench/dark-magic/internal/dev/profiling"
	"github.com/gravestench/dark-magic/internal/distribution"
	"github.com/gravestench/dark-magic/internal/logging"
	d2save "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/save"
	"github.com/gravestench/dark-magic/internal/modcache"
	darkpaths "github.com/gravestench/dark-magic/internal/paths"
	"github.com/gravestench/dark-magic/internal/shell"
)

func main() {
	exitCode := 0
	// Cleanups run before this final exit. Think of it like putting every toy
	// away before turning off the room light.
	defer func() { os.Exit(exitCode) }()

	// macOS requires the window to live on the first operating-system thread.
	// Locking here keeps that rule out of the rest of the application.
	runtime.LockOSThread()

	logLevelFlag := flag.String("log-level", environmentDefault("DARK_MAGIC_LOG_LEVEL", "info"), "log verbosity: trace, debug, info, warn, or error")
	profileDirectory := flag.String("profile-dir", os.Getenv("DARK_MAGIC_PROFILE_DIR"), "capture CPU and heap profiles plus PDF reports in this directory")
	profileScenes := flag.String("profile-scenes", os.Getenv("DARK_MAGIC_PROFILE_SCENES"), "comma-separated scene IDs (or all) for per-scene CPU and heap reports")
	captureDirectoryFlag := flag.String("capture-dir", os.Getenv("DARK_MAGIC_CAPTURE_DIR"), "write local scene screenshots and report.json to this directory")
	captureScenes := flag.String("capture-scenes", os.Getenv("DARK_MAGIC_CAPTURE_SCENES"), "comma-separated scene IDs to capture (defaults to loading,title)")
	captureSettle := flag.Int("capture-settle-frames", 10, "stable frames to wait before capturing a scene")
	startScene := flag.String("start-scene", os.Getenv("DARK_MAGIC_START_SCENE"), "development-only scene ID to enter after boot")
	startOverlays := flag.String("start-overlays", os.Getenv("DARK_MAGIC_START_OVERLAYS"), "development-only comma-separated overlays to open above the start scene")
	fixtureCharacters := flag.Int("fixture-characters", 0, "development-only number of in-memory characters to create")
	fixtureWorldLevel := flag.Int("fixture-world-level", 0, "development-only authoritative level for the selected fixture character (scene default when omitted)")
	fixtureWorldSpawn := flag.String("fixture-world-spawn", "entry", "development-only fixture spawn: entry or seam")
	fixturePointerMove := flag.Bool("fixture-pointer-move", false, "development-only click-to-move acceptance before capture")
	outputPalette := flag.String("output-palette", os.Getenv("DARK_MAGIC_OUTPUT_PALETTE"), "quantize the final display through this mounted pal.dat asset")
	viewportFit := flag.String("viewport-fit", environmentDefault("DARK_MAGIC_VIEWPORT_FIT", "contain"), "game viewport fit: contain or stretch")
	fullscreenDefault, _ := strconv.ParseBool(environmentDefault("DARK_MAGIC_FULLSCREEN", "false"))
	fullscreen := flag.Bool("fullscreen", fullscreenDefault, "use a maximized borderless window")
	presentationProfile := flag.String("presentation-profile", os.Getenv("DARK_MAGIC_PRESENTATION_PROFILE"), "manifest-owned presentation profile ID")
	modsFlag := flag.String("mods", os.Getenv("DARK_MAGIC_MODS"), "temporary comma-separated mod IDs, or 'none' for mod-neutral startup")
	flag.Parse()

	logLevel, err := parseLogLevel(*logLevelFlag)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		exitCode = 1
		return
	}
	logs := shell.NewLogBuffer(1000)
	handler := logging.NewHandlerWithObserver(&slog.HandlerOptions{Level: logLevel}, func(record logging.Record) {
		logs.Append(shell.LogEntry{At: record.At, Level: record.Level.String(), Message: record.Message, Attributes: record.Attributes})
	})
	slog.SetDefault(slog.New(handler))

	var profile *profiling.Session
	if *profileDirectory != "" {
		profile, err = profiling.Start(*profileDirectory, true)
		if err != nil {
			slog.Error("starting profiler", "error", err)
			exitCode = 1
			return
		}
		profile.ConfigureScenes(*profileScenes)
		defer func() {
			if err := profile.Stop(); err != nil {
				slog.Error("finishing profiler", "error", err)
			}
		}()
	}

	mods, err := distribution.PrepareMods(*modsFlag)
	if err != nil {
		slog.Error("preparing mod profile", "error", err)
		exitCode = 1
		return
	}
	defer func() {
		if err := mods.Close(); err != nil {
			slog.Error("closing mod packages", "error", err)
		}
	}()
	contentFS, err := content.FromEnvironment(mods.Layers...)
	if err == nil && len(mods.Lock.Packages) > 0 {
		err = content.ValidateClientAssets(contentFS)
	}
	if err != nil {
		slog.Error("preparing client content", "error", err)
		exitCode = 1
		return
	}

	*captureDirectoryFlag, *captureScenes = capture.Defaults(*captureDirectoryFlag, *captureScenes)
	captureDirectory, err := darkpaths.ExpandHost(*captureDirectoryFlag)
	if err != nil {
		slog.Error("expanding capture directory", "error", err)
		exitCode = 1
		return
	}
	if err := run(contentFS, &mods.Lock, profile, captureDirectory, *captureScenes, *captureSettle, *startScene, *startOverlays, *fixtureCharacters, *fixtureWorldLevel, *fixtureWorldSpawn, *fixturePointerMove, *outputPalette, *viewportFit, *fullscreen, *presentationProfile, logs); err != nil {
		slog.Error("running Dark Magic", "error", err)
		exitCode = 1
	}
}

func environmentDefault(name, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(name)); value != "" {
		return value
	}
	return fallback
}

func parseLogLevel(value string) (slog.Level, error) { return logging.ParseLevel(value) }

// run is intentionally boring. The command hands the pieces to the client
// application package, and that package explains how the pieces fit together.
func run(contentFS *content.FS, mods *modcache.Lock, profile *profiling.Session, captureDirectory, captureScenes string, captureSettle int, startScene, startOverlays string, fixtureCharacters, fixtureWorldLevel int, fixtureWorldSpawn string, fixturePointerMove bool, outputPalette, viewportFit string, fullscreen bool, presentationProfileID string, logs *shell.LogBuffer) error {
	playerProfilePath := strings.TrimSpace(os.Getenv("DARK_MAGIC_PLAYER_PROFILE"))
	if playerProfilePath == "" {
		configurationDirectory, err := os.UserConfigDir()
		if err != nil {
			return fmt.Errorf("resolve player profile directory: %w", err)
		}
		playerProfilePath = filepath.Join(configurationDirectory, "dark-magic", "player-profile.json")
	}
	options := clientapp.Options{
		Content: contentFS, Mods: mods, NewCapture: func(directory, scenes string, settle int, renderer clientapp.Screenshotter) (clientapp.Capture, error) {
			return capture.New(directory, scenes, settle, renderer)
		}, CaptureDirectory: captureDirectory,
		CaptureScenes: captureScenes, CaptureSettle: captureSettle, StartScene: startScene, StartOverlays: startOverlays,
		FixtureCharacters: fixtureCharacters, FixtureWorldLevel: fixtureWorldLevel, FixtureWorldSpawn: fixtureWorldSpawn, FixturePointerMove: fixturePointerMove, OutputPalette: outputPalette,
		PlayerProfilePath: playerProfilePath,
		ViewportFit:       viewportFit, BorderlessFullscreen: fullscreen, PresentationProfileID: presentationProfileID, Logs: logs,
	}
	// A nil pointer stored inside an interface looks non-nil. Only put the
	// profiler in the box when one was really started.
	if profile != nil {
		options.Profile = profile
	}
	return clientapp.Run(options)
}

// Keep this tiny forwarding function for the command's historical tests. The
// fixture recipe itself belongs beside the client application that consumes it.
func developmentCharacters(count int) []d2save.Character {
	return clientapp.DevelopmentCharacters(count)
}

func buildVersion() string { return clientapp.BuildVersion() }
