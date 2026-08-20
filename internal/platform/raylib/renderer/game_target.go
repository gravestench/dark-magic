package raylibRenderer

import (
	"fmt"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// startGameTarget validates presentation geometry before allocating the render texture. In native mode this target is
// matched to the drawable window; otherwise it remains the manifest-owned logical resolution.
func (s *Service) startGameTarget() error {
	width, height := s.frameResolution()
	if width <= 0 || height <= 0 {
		return fmt.Errorf("renderer: logical resolution requires positive dimensions, got %dx%d", width, height)
	}

	_, err := gameViewport(
		s.config.Window.Width,
		s.config.Window.Height,
		width,
		height,
		s.config.Resolution.Fit,
	)
	if err != nil {
		return err
	}

	s.gameTarget = rl.LoadRenderTexture(int32(width), int32(height))
	if !rl.IsRenderTextureValid(s.gameTarget) {
		return fmt.Errorf("renderer: create logical game target %dx%d", width, height)
	}

	rl.SetTextureFilter(s.gameTarget.Texture, rl.FilterPoint)

	return nil
}

// resizeGameTargetForWindow replaces the offscreen target only in native mode and only after a real drawable-size
// change. It runs on the renderer owner thread before a frame begins, so no composition sees a half-resized target.
func (s *Service) resizeGameTargetForWindow() error {
	if s.config == nil || !s.config.Resolution.Native {
		return nil
	}

	width, height := s.frameResolution()
	if width <= 0 || height <= 0 {
		return fmt.Errorf("renderer: native resolution requires positive dimensions, got %dx%d", width, height)
	}

	if s.paletteQuantizer != nil {
		return s.resizePaletteTarget(width, height)
	}
	if int(s.gameTarget.Texture.Width) == width && int(s.gameTarget.Texture.Height) == height {
		return nil
	}
	s.stopGameTarget()
	return s.startGameTarget()
}

// frameResolution is safe both during startup (after the native window exists but before isInit is published) and
// during steady-state resize handling.
func (s *Service) frameResolution() (width, height int) {
	if s.config != nil && s.config.Resolution.Native {
		width, height = s.WindowSize()
		if width > 0 && height > 0 {
			return width, height
		}
	}
	return s.Resolution()
}

// stopGameTarget releases the logical render texture once and clears its handle so repeated shutdown remains safe.
func (s *Service) stopGameTarget() {
	if rl.IsRenderTextureValid(s.gameTarget) {
		rl.UnloadRenderTexture(s.gameTarget)
		s.gameTarget = rl.RenderTexture2D{}
	}
}

// renderGameTarget draws one complete logical frame into an offscreen texture. The camera and scene graph are confined
// to this target so later window-space overlays do not inherit their transform.
func (s *Service) renderGameTarget(target rl.RenderTexture2D) {
	rl.BeginTextureMode(target)
	rl.ClearBackground(rl.Black)
	rl.BeginMode2D(s.camera)
	s.update()
	s.render()
	rl.EndMode2D()
	rl.EndTextureMode()
}

// presentGameTarget maps the logical texture into the current drawable viewport, optionally through a final shader. A
// negative source height corrects render-texture orientation without copying pixels.
func (s *Service) presentGameTarget(target rl.RenderTexture2D, shader *rl.Shader) {
	viewport, err := gameViewport(
		rl.GetRenderWidth(),
		rl.GetRenderHeight(),
		int(target.Texture.Width),
		int(target.Texture.Height),
		s.config.Resolution.Fit,
	)
	if err != nil {
		return
	}

	if shader != nil {
		rl.BeginShaderMode(*shader)

		defer rl.EndShaderMode()
	}

	source := rl.NewRectangle(0, 0, float32(target.Texture.Width), -float32(target.Texture.Height))
	destination := rl.NewRectangle(viewport.X, viewport.Y, viewport.Width, viewport.Height)
	rl.DrawTexturePro(target.Texture, source, destination, rl.Vector2{}, 0, rl.White)
}
