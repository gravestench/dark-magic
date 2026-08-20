// DS1 Editor is a standalone content-authoring program. It mounts Diablo II
// data read-only and starts only the renderer, Lua authoring scene, and codecs.
package main

import (
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"runtime"
	"strings"

	"github.com/gravestench/dark-magic/internal/app/ds1editorapp"
	"github.com/gravestench/dark-magic/internal/app/envconfig"
	"github.com/gravestench/dark-magic/internal/content"
	"github.com/gravestench/dark-magic/internal/distribution"
	"github.com/gravestench/dark-magic/internal/modcache"
	darkpaths "github.com/gravestench/dark-magic/internal/paths"
)

// main pins the graphics process to its owner thread and returns the command's stable exit status.
func main() {
	runtime.LockOSThread()
	os.Exit(run())
}

// run resolves command policy, mounts read-only content, and transfers control to the standalone application.
// Every mounted extension remains owned here and is closed after the desktop process exits.
func run() int {
	environment, err := envconfig.Bootstrap("ds1-editor", os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	outputDefault, err := ds1editorapp.DefaultOutputDirectory()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	if value := strings.TrimSpace(os.Getenv("DARK_MAGIC_DS1_EDITOR_OUTPUT")); value != "" {
		outputDefault = value
	}
	windowDefault := environmentDefault("DARK_MAGIC_DS1_EDITOR_WINDOW_SIZE", "1280x760")
	// Migrate the original generated default. Existing users may already have
	// 1440x900 in ds1-editor.env even though they never chose it; on shorter Mac
	// displays Cocoa constrains that client area after Raylib allocates its first
	// target, which clips the top of the flipped frame until a manual resize.
	if windowDefault == "1440x900" {
		windowDefault = "1280x760"
	}
	fullscreenDefault := strings.EqualFold(environmentDefault("DARK_MAGIC_DS1_EDITOR_FULLSCREEN", "false"), "true")
	_ = flag.String(
		"env-file",
		environment.DefaultPath,
		"environment file (overrides the default ds1-editor.env selection)",
	)
	output := flag.String("output", outputDefault, "directory where edited DS1 copies are written")
	initialMap := flag.String("open", "", "mounted DS1 path to open immediately")
	windowSize := flag.String("window-size", windowDefault, "initial editor window size as WIDTHxHEIGHT")
	fullscreen := flag.Bool("fullscreen", fullscreenDefault, "use a maximized borderless editor window")
	mods := flag.String(
		"mods",
		environmentDefault("DARK_MAGIC_MODS", "none"),
		"content extension selection; none uses bundled d2legacy assets",
	)
	palette := flag.String(
		"output-palette",
		os.Getenv("DARK_MAGIC_OUTPUT_PALETTE"),
		"optional mounted palette for final display quantization",
	)
	nativeAudio := flag.Bool("native-audio", false, "enable native audio; disabled by default for this authoring tool")
	_ = environment
	flag.Parse()
	width, height, err := parseWindowSize(*windowSize)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	contentFS, contentOwner, err := mountEditorContent(*mods)
	if err != nil {
		slog.Error("mounting DS1 editor content", "error", err)
		return 1
	}
	defer func() { _ = contentOwner.Close() }()
	roots, err := ds1editorapp.MountedRoots(os.Getenv("MPQ_DIRECTORY"))
	if err != nil {
		slog.Error("resolving read-only game data", "error", err)
		return 1
	}
	outputPath, err := darkpaths.ExpandHost(*output)
	if err != nil {
		slog.Error("resolving DS1 editor output", "error", err)
		return 1
	}
	if err := ds1editorapp.Run(ds1editorapp.Options{
		Content:            contentFS,
		OutputDirectory:    outputPath,
		ReadOnlyRoots:      roots,
		WindowWidth:        width,
		WindowHeight:       height,
		InitialMap:         *initialMap,
		Borderless:         *fullscreen,
		OutputPalette:      *palette,
		DisableNativeAudio: !*nativeAudio,
	}); err != nil {
		slog.Error("running DS1 editor", "error", err)
		return 1
	}
	return 0
}

// mountEditorContent composes source assets and the editor package as read-only VFS layers.
// The returned owner keeps extension archives alive for exactly as long as the application uses them.
func mountEditorContent(modSelection string) (*content.FS, io.Closer, error) {
	modSet, err := distribution.PrepareMods(modSelection)
	if err != nil {
		return nil, nil, fmt.Errorf("prepare selected content: %w", err)
	}

	contentFS, err := content.FromEnvironment(modSet.Layers...)
	if err != nil {
		_ = modSet.Close()
		return nil, nil, err
	}

	editorManifest, err := modcache.ReadManifest(content.DS1Editor())
	if err != nil {
		_ = modSet.Close()
		return nil, nil, fmt.Errorf("describe editor package: %w", err)
	}
	editorFS, err := modcache.NewPackageFS(editorManifest, content.DS1Editor())
	if err != nil {
		_ = modSet.Close()
		return nil, nil, fmt.Errorf("mount editor package: %w", err)
	}
	if err := contentFS.MountFirst(content.Layer{Name: "builtin:ds1editor", FS: editorFS}); err != nil {
		_ = modSet.Close()
		return nil, nil, fmt.Errorf("layer editor package: %w", err)
	}

	return contentFS, modSet, nil
}

// environmentDefault returns a trimmed environment override or a stable command default.
func environmentDefault(name, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(name)); value != "" {
		return value
	}
	return fallback
}

// parseWindowSize validates the tool's minimum usable native-resolution workspace.
func parseWindowSize(value string) (int, int, error) {
	parts := strings.Split(strings.ToLower(strings.TrimSpace(value)), "x")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("window size %q must use WIDTHxHEIGHT", value)
	}
	var width, height int
	if _, err := fmt.Sscanf(parts[0], "%d", &width); err != nil {
		return 0, 0, fmt.Errorf("invalid window size %q", value)
	}
	if _, err := fmt.Sscanf(parts[1], "%d", &height); err != nil {
		return 0, 0, fmt.Errorf("invalid window size %q", value)
	}
	if width < 640 || height < 480 {
		return 0, 0, fmt.Errorf("window size %q must be at least 640x480", value)
	}
	return width, height, nil
}
