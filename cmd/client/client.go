package main

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/gravestench/dark-magic/internal/app/clientapp"
	"github.com/gravestench/dark-magic/internal/content"
	"github.com/gravestench/dark-magic/internal/dev/capture"
	"github.com/gravestench/dark-magic/internal/dev/profiling"
	"github.com/gravestench/dark-magic/internal/distribution"
	d2save "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/save"
	"github.com/gravestench/dark-magic/internal/shell"
)

// run crosses the process/application ownership boundary. The command keeps
// responsibility for closing mounted content and profiling resources; the
// application receives only the capabilities and immutable policy it needs.
func run(
	contentFS *content.FS,
	mods *distribution.ModSet,
	profile *profiling.Session,
	config clientConfig,
	logs *shell.LogBuffer,
) error {
	assetSetID, err := content.AssetSetIdentityFromEnvironment()
	if err != nil {
		return fmt.Errorf("identify external game assets: %w", err)
	}

	slog.Debug("identified external game asset set", "asset_set_id", assetSetID)

	playerProfilePath, err := resolvePlayerProfilePath()
	if err != nil {
		return err
	}

	options := clientapp.Options{
		Content:               contentFS,
		Mods:                  &mods.Resolved,
		Packages:              mods.Packages,
		AssetSetID:            assetSetID,
		ModCache:              mods.Cache,
		NewCapture:            newCapture,
		CaptureDirectory:      config.captureDirectory,
		CaptureScenes:         config.captureScenes,
		CaptureSettle:         config.captureSettle,
		StartScene:            config.startScene,
		StartOverlays:         config.startOverlays,
		FixtureCharacters:     config.fixtureCharacters,
		FixtureWorldLevel:     config.fixtureWorldLevel,
		FixtureWorldSpawn:     config.fixtureWorldSpawn,
		FixturePointerMove:    config.fixturePointerMove,
		OutputPalette:         config.outputPalette,
		PlayerProfilePath:     playerProfilePath,
		ViewportFit:           config.viewportFit,
		BorderlessFullscreen:  config.fullscreen,
		DisableNativeAudio:    !config.nativeAudio,
		PresentationProfileID: config.presentationProfileID,
		Logs:                  logs,
	}
	// A nil pointer stored inside an interface looks non-nil. Only assign the
	// profiler when the command actually started a session.
	if profile != nil {
		options.Profile = profile
	}

	return clientapp.Run(options)
}

// resolvePlayerProfilePath locates durable offline character state. An explicit
// environment path supports portable/dev setups, while the platform default
// keeps ordinary players out of the working tree.
func resolvePlayerProfilePath() (string, error) {
	if configured := strings.TrimSpace(os.Getenv("DARK_MAGIC_PLAYER_PROFILE")); configured != "" {
		return configured, nil
	}

	configurationDirectory, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("resolve player profile directory: %w", err)
	}

	return filepath.Join(configurationDirectory, "dark-magic", "player-profile.json"), nil
}

// newCapture keeps the reusable clientapp package independent of the concrete
// developer capture implementation selected by this executable.
func newCapture(
	directory string,
	scenes string,
	settle int,
	renderer clientapp.Screenshotter,
) (clientapp.Capture, error) {
	return capture.New(directory, scenes, settle, renderer)
}

// developmentCharacters preserves the command-level fixture seam used by older
// tests and tools while clientapp remains the single owner of fixture contents.
func developmentCharacters(count int) []d2save.Character {
	return clientapp.DevelopmentCharacters(count)
}
