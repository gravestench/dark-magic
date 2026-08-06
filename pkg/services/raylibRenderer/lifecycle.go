package raylibRenderer

import (
	"context"
	"errors"
	"log/slog"

	rl "github.com/gen2brain/raylib-go/raylib"
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
		if level >= 4 {
			s.logger.Error(message)
			return
		}
		s.logger.Debug(message)
	})
	rl.InitWindow(int32(s.config.Window.Width), int32(s.config.Window.Height), s.config.Window.Title)
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
		rl.BeginDrawing()
		rl.ClearBackground(rl.Black)
		rl.BeginMode2D(*s.GetDefaultCamera())
		s.update()
		s.render()
		rl.EndMode2D()
		rl.EndDrawing()
	}
	return nil
}

// Stop releases native renderer resources on the calling main thread.
func (s *Service) Stop(context.Context) error {
	if !s.isInit.Swap(false) {
		return nil
	}
	rl.CloseAudioDevice()
	rl.CloseWindow()
	return nil
}
