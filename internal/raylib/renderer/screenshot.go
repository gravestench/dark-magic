package raylibRenderer

import (
	"fmt"
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
	rl.TakeScreenshot(name)
	info, err := os.Stat(name)
	if err != nil {
		return fmt.Errorf("renderer: verify screenshot: %w", err)
	}
	if info.Size() == 0 {
		return fmt.Errorf("renderer: screenshot %q is empty", name)
	}
	return nil
}
