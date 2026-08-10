package raylibRenderer

import (
	"context"
	"errors"
	"log/slog"
	"runtime"
	"strings"

	rl "github.com/gen2brain/raylib-go/raylib"

	"github.com/gravestench/dark-magic/internal/branding"
)

// Start initializes Raylib on the calling main thread and returns when native
// renderer capabilities are ready for dependents.
func (s *Service) Start(context.Context) error {
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
		s.cameras = make(map[string]*rl.Camera2D)
		s.rootNode = s.NewRenderable()
		s.rootNode.Disable()
	}
	rl.SetTraceLogCallback(func(level int, message string) {
		switch {
		case level >= int(rl.LogError):
			s.logger.Error(message)
		case level == int(rl.LogWarning) && strings.HasPrefix(message, "FONT: Requested codepoints glyphs found:"):
			// Raylib reports the intentionally sparse Diablo II bitmap-font
			// codepoint set while successfully installing fallback glyphs.
			s.logger.Debug(message)
		case level == int(rl.LogWarning):
			s.logger.Warn(message)
		default:
			s.logger.Debug(message)
		}
	})
	if s.config.Window.Resizable {
		rl.SetConfigFlags(rl.FlagWindowResizable)
	}
	rl.InitWindow(int32(s.config.Window.Width), int32(s.config.Window.Height), s.config.Window.Title)
	// Cocoa regular windows do not support GLFW window icons. macOS uses the
	// application-bundle icon instead, so avoid provoking a native warning.
	if runtime.GOOS != "darwin" {
		iconData := branding.WindowIconPNG()
		icon := rl.LoadImageFromMemory(".png", iconData, int32(len(iconData)))
		if icon.Width > 0 && icon.Height > 0 {
			// GLFW accepts window icons only as RGBA pixels. The embedded PNG is
			// intentionally stored as RGB; convert the decoded image first.
			rl.ImageFormat(icon, rl.UncompressedR8g8b8a8)
			rl.SetWindowIcon(*icon)
			rl.UnloadImage(icon)
		} else {
			s.logger.Warn("renderer: failed to decode embedded window icon")
		}
	}
	// Escape belongs to scene and shell focus routing. WindowShouldClose still
	// observes the native close control after Raylib's default Escape binding is
	// disabled.
	rl.SetExitKey(rl.KeyNull)
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
	rl.InitAudioDevice()
	rl.SetTargetFPS(60)
	rl.HideCursor()
	s.isInit.Store(true)
	return nil
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
	if s.cache != nil {
		s.cache.Clear()
	}
	if s.audioBackend != nil {
		s.audioBackend.Close()
	}
	if s.compositionBackend != nil {
		s.compositionBackend.closePaletteEffects()
	}
	s.stopPaletteQuantizer()
	s.stopGameTarget()
	rl.CloseAudioDevice()
	rl.CloseWindow()
	return nil
}
