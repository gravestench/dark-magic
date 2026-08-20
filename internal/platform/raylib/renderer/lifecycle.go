package raylibRenderer

import (
	"context"
	"errors"
	"log/slog"
	"math"
	"runtime"
	"strings"

	rl "github.com/gen2brain/raylib-go/raylib"

	"github.com/gravestench/dark-magic/internal/branding"
	"github.com/gravestench/dark-magic/internal/logging"
)

// Start initializes Raylib on the calling main thread and returns when native
// renderer capabilities are ready for dependents.
func (s *Service) Start(context.Context) error {
	if err := s.prepareStart(); err != nil {
		return err
	}

	s.installTraceLogging()
	s.openWindow()

	if err := s.startFrameTargets(); err != nil {
		return err
	}

	rl.InitAudioDevice()
	rl.SetTargetFPS(60)
	s.applyCursorVisibility()
	s.isInit.Store(true)

	return nil
}

// prepareStart validates mandatory dependencies before any native resource is allocated. It also creates the disabled
// scene root once so restart paths retain the same camera defaults without duplicating nodes.
func (s *Service) prepareStart() error {
	if s.config == nil {
		return errors.New("renderer: configuration is required")
	}

	if s.logger == nil {
		s.logger = slog.Default()
	}

	if s.cache == nil {
		return errors.New("renderer: cache is required")
	}

	if s.rootNode == nil {
		s.camera = rl.NewCamera2D(rl.Vector2{}, rl.Vector2{}, 0, 1)
		s.rootNode = s.newNode()
		s.rootNode.Disable()
	}

	return nil
}

// installTraceLogging maps raylib's native severity into structured logs. The sparse-font warning is expected for the
// Diablo bitmap font and remains trace-level so normal users do not see a false failure signal.
func (s *Service) installTraceLogging() {
	rl.SetTraceLogCallback(func(level int, message string) {
		switch {
		case level >= int(rl.LogError):
			s.logger.Error(message)
		case level == int(rl.LogWarning) && strings.HasPrefix(message, "FONT: Requested codepoints glyphs found:"):
			// Raylib reports the intentionally sparse Diablo II bitmap-font
			// codepoint set while successfully installing fallback glyphs.
			logging.Trace(s.logger, message)
		case level == int(rl.LogWarning):
			s.logger.Warn(message)
		default:
			logging.Trace(s.logger, message)
		}
	})
}

// openWindow applies flags before initialization, then configures platform-specific icon, close-key, and borderless
// behavior. These calls must remain on the owner thread because raylib delegates them directly to GLFW.
func (s *Service) openWindow() {
	if flags := windowConfigFlags(*s.config); flags != 0 {
		rl.SetConfigFlags(flags)
	}

	rl.InitWindow(int32(s.config.Window.Width), int32(s.config.Window.Height), s.config.Window.Title)

	if s.config.Window.Borderless {
		// Desktop fullscreen remains an ordinary window: it loses its frame and
		// fills the monitor without requesting an exclusive video mode.
		rl.MaximizeWindow()
	} else if !s.config.Window.Fullscreen {
		s.fitAndCenterWindow()
	}

	installWindowIcon(s.logger)

	// Escape belongs to scene and shell focus routing. WindowShouldClose still observes the native close control after
	// raylib's default Escape binding is disabled.
	rl.SetExitKey(rl.KeyNull)
}

// fitAndCenterWindow keeps an explicit development/tool window wholly visible
// on its selected monitor. Raylib/GLFW otherwise accepts oversized dimensions
// and may place the lower-right portion off screen on laptop displays.
func (s *Service) fitAndCenterWindow() {
	monitor := rl.GetCurrentMonitor()
	monitorWidth, monitorHeight := logicalMonitorSize(
		rl.GetMonitorWidth(monitor),
		rl.GetMonitorHeight(monitor),
		rl.GetWindowScaleDPI().X,
		rl.GetWindowScaleDPI().Y,
	)
	if monitorWidth <= 0 || monitorHeight <= 0 {
		return
	}
	position := rl.GetMonitorPosition(monitor)
	monitorX, monitorY := int(position.X), int(position.Y)
	if runtime.GOOS == "darwin" {
		monitorX, monitorY, monitorWidth, monitorHeight = macOSWindowWorkArea(
			monitorX,
			monitorY,
			monitorWidth,
			monitorHeight,
		)
	}
	width, height, x, y := fitWindowToMonitor(
		s.config.Window.Width,
		s.config.Window.Height,
		monitorWidth,
		monitorHeight,
		monitorX,
		monitorY,
	)
	if width <= 0 || height <= 0 {
		return
	}
	if width != s.config.Window.Width || height != s.config.Window.Height {
		rl.SetWindowSize(width, height)
		s.config.Window.Width, s.config.Window.Height = width, height
	}
	rl.SetWindowPosition(x, y)
}

// macOSWindowWorkArea reserves logical points for the menu bar, native title
// bar, and Dock because Raylib exposes the full monitor mode but not Cocoa's
// visible frame. Renderer coordinates still begin at the client area's origin;
// these insets affect only initial outer-window sizing and placement.
func macOSWindowWorkArea(x, y, width, height int) (int, int, int, int) {
	const (
		topInset    = 24
		bottomInset = 92
	)

	usableHeight := max(480, height-topInset-bottomInset)
	return x, y + topInset, width, usableHeight
}

// logicalMonitorSize converts a monitor video mode into the same logical-unit
// coordinate space used by GLFW window placement. On Retina displays, the
// monitor mode is commonly expressed in framebuffer pixels while window sizes
// and positions are expressed in points.
func logicalMonitorSize(width, height int, scaleX, scaleY float32) (int, int) {
	if scaleX > 1 {
		width = int(math.Round(float64(width) / float64(scaleX)))
	}
	if scaleY > 1 {
		height = int(math.Round(float64(height) / float64(scaleY)))
	}
	return width, height
}

// fitWindowToMonitor constrains a requested logical window size to its monitor
// and returns a centered global desktop position. Monitor origins matter on
// displays arranged to the left, right, above, or below the primary display.
func fitWindowToMonitor(
	requestedWidth int,
	requestedHeight int,
	monitorWidth int,
	monitorHeight int,
	monitorX int,
	monitorY int,
) (width, height, x, y int) {
	if requestedWidth <= 0 || requestedHeight <= 0 || monitorWidth <= 0 || monitorHeight <= 0 {
		return 0, 0, 0, 0
	}
	width = min(requestedWidth, monitorWidth)
	height = min(requestedHeight, monitorHeight)
	return width, height, monitorX + (monitorWidth-width)/2, monitorY + (monitorHeight-height)/2
}

// installWindowIcon decodes the embedded icon only on platforms where GLFW supports regular-window icons. Cocoa uses
// the application bundle icon, and attempting this native call there produces a misleading warning.
func installWindowIcon(logger *slog.Logger) {
	// Cocoa regular windows do not support GLFW window icons. macOS uses the
	// application-bundle icon instead, so avoid provoking a native warning.
	if runtime.GOOS == "darwin" {
		return
	}

	iconData := branding.WindowIconPNG()

	icon := rl.LoadImageFromMemory(".png", iconData, int32(len(iconData)))
	if icon.Width <= 0 || icon.Height <= 0 {
		logger.Warn("renderer: failed to decode embedded window icon")

		return
	}

	// GLFW accepts window icons only as RGBA pixels. The embedded PNG is intentionally stored as RGB; convert the
	// decoded image first.
	rl.ImageFormat(icon, rl.UncompressedR8g8b8a8)
	rl.SetWindowIcon(*icon)
	rl.UnloadImage(icon)
}

// startFrameTargets initializes logical rendering and optional palette quantization as one recoverable phase. Failure
// releases the window and any target allocated earlier so callers may safely retry Start.
func (s *Service) startFrameTargets() error {
	if s.paletteQuantizer == nil {
		if err := s.startGameTarget(); err != nil {
			rl.CloseWindow()

			return err
		}
	}

	if err := s.startPaletteQuantizer(); err != nil {
		s.stopGameTarget()
		rl.CloseWindow()

		return err
	}

	return nil
}

// applyCursorVisibility enforces the configured ownership of the system pointer after the native window exists.
func (s *Service) applyCursorVisibility() {
	if s.config.Window.ShowSystemCursor {
		rl.ShowCursor()
	} else {
		rl.HideCursor()
	}
}

// windowConfigFlags converts renderer configuration into raylib's pre-window bitmask. Borderless desktop mode includes
// maximization so startup uses the monitor work area without entering exclusive fullscreen.
func windowConfigFlags(config Config) uint32 {
	var flags uint32
	if config.Window.Resizable {
		flags |= rl.FlagWindowResizable
	}

	if config.Window.Fullscreen {
		flags |= rl.FlagFullscreenMode
	}

	if config.Window.Borderless {
		flags |= rl.FlagBorderlessWindowedMode | rl.FlagWindowMaximized
	}

	return flags
}

// Run owns the frame loop and must be invoked on the process main thread after
// all frame-producing components have started.
func (s *Service) Run(ctx context.Context) error {
	if !s.isInit.Load() {
		return errors.New("renderer: not started")
	}

	for !rl.WindowShouldClose() {
		select {
		case <-ctx.Done():
			return nil
		default:
		}
		if err := s.resizeGameTargetForWindow(); err != nil {
			return err
		}

		if s.paletteQuantizer != nil {
			if err := s.renderQuantizedFrame(); err != nil {
				return err
			}

			s.runPostFrame()

			continue
		}

		s.renderGameTarget(s.gameTarget)
		rl.BeginDrawing()
		rl.ClearBackground(rl.Black)
		s.presentGameTarget(s.gameTarget, nil)
		s.runOverlays()
		rl.EndDrawing()
		s.runPostFrame()
	}

	return nil
}

// Stop releases native renderer resources on the calling main thread.
func (s *Service) Stop(context.Context) error {
	if !s.isInit.Swap(false) {
		return nil
	}

	var result error

	if s.audioBackend != nil {
		s.audioBackend.Close()
	}

	if s.compositionBackend != nil {
		if s.composition != nil {
			if err := s.composition.Drain(s.compositionBackend); err != nil {
				result = errors.Join(result, err)
			}
		}

		s.compositionBackend.close()
		s.compositionBackend.closePaletteEffects()
	}

	if s.cache != nil {
		s.cache.Clear()
	}

	s.stopPaletteQuantizer()
	s.stopGameTarget()
	rl.CloseAudioDevice()
	rl.CloseWindow()

	s.rootNode = nil

	return result
}
