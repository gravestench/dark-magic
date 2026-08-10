package raylibRenderer

import (
	"fmt"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func (s *Service) startGameTarget() error {
	width, height := s.config.Resolution.Width, s.config.Resolution.Height
	if width <= 0 || height <= 0 {
		return fmt.Errorf("renderer: logical resolution requires positive dimensions, got %dx%d", width, height)
	}
	if _, err := gameViewport(s.config.Window.Width, s.config.Window.Height, width, height, s.config.Resolution.Fit); err != nil {
		return err
	}
	s.gameTarget = rl.LoadRenderTexture(int32(width), int32(height))
	if !rl.IsRenderTextureValid(s.gameTarget) {
		return fmt.Errorf("renderer: create logical game target %dx%d", width, height)
	}
	rl.SetTextureFilter(s.gameTarget.Texture, rl.FilterPoint)
	return nil
}

func (s *Service) stopGameTarget() {
	if rl.IsRenderTextureValid(s.gameTarget) {
		rl.UnloadRenderTexture(s.gameTarget)
		s.gameTarget = rl.RenderTexture2D{}
	}
}

func (s *Service) renderGameTarget(target rl.RenderTexture2D) {
	rl.BeginTextureMode(target)
	rl.ClearBackground(rl.Black)
	rl.BeginMode2D(s.camera)
	s.update()
	s.render()
	rl.EndMode2D()
	rl.EndTextureMode()
}

func (s *Service) presentGameTarget(target rl.RenderTexture2D, shader *rl.Shader) {
	viewport, err := gameViewport(rl.GetRenderWidth(), rl.GetRenderHeight(), int(target.Texture.Width), int(target.Texture.Height), s.config.Resolution.Fit)
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
