package raylibRenderer

import (
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// CaptureScreenshot writes the currently presented framebuffer. It must be
// called from renderer owner-thread work, normally a post-frame callback.
func (s *Service) CaptureScreenshot(name string) error {
	if !s.isInit.Load() {
		return fmt.Errorf("renderer: screenshot requested before initialization")
	}
	if name == "" {
		return fmt.Errorf("renderer: screenshot path is required")
	}
	if err := os.MkdirAll(filepath.Dir(name), 0o755); err != nil {
		return err
	}
	screen := rl.LoadImageFromScreen()
	if screen == nil || !rl.IsImageValid(screen) {
		return fmt.Errorf("renderer: read screenshot framebuffer")
	}
	defer rl.UnloadImage(screen)

	colors := rl.LoadImageColors(screen)
	if len(colors) == 0 {
		return fmt.Errorf("renderer: screenshot framebuffer is empty")
	}
	defer rl.UnloadImageColors(colors)

	frame := image.NewRGBA(image.Rect(0, 0, int(screen.Width), int(screen.Height)))
	for index, pixel := range colors {
		x := index % int(screen.Width)
		y := index / int(screen.Width)
		frame.SetRGBA(x, y, pixel)
	}

	file, err := os.Create(name)
	if err != nil {
		return fmt.Errorf("renderer: create screenshot: %w", err)
	}
	if err := png.Encode(file, frame); err != nil {
		_ = file.Close()
		return fmt.Errorf("renderer: encode screenshot: %w", err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("renderer: close screenshot: %w", err)
	}
	return nil
}
