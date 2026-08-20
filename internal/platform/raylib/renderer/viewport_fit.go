package raylibRenderer

import "fmt"

// Viewport describes where the logical game surface is presented in window
// pixels. Console and other overlays deliberately do not use this transform.
type Viewport struct {
	X, Y, Width, Height float32
}

// gameViewport validates dimensions and computes either a stretching rectangle or centered aspect-preserving bounds.
// The same calculation drives rendering and input mapping, preventing disagreement in letterboxed regions.
func gameViewport(windowWidth, windowHeight, gameWidth, gameHeight int, fit string) (Viewport, error) {
	if windowWidth <= 0 || windowHeight <= 0 || gameWidth <= 0 || gameHeight <= 0 {
		return Viewport{}, fmt.Errorf("renderer: viewport dimensions must be positive")
	}

	if fit == "stretch" {
		return Viewport{Width: float32(windowWidth), Height: float32(windowHeight)}, nil
	}

	if fit != "contain" && fit != "" {
		return Viewport{}, fmt.Errorf("renderer: viewport fit must be contain or stretch, got %q", fit)
	}

	scale := min(float64(windowWidth)/float64(gameWidth), float64(windowHeight)/float64(gameHeight))
	width, height := float32(float64(gameWidth)*scale), float32(float64(gameHeight)*scale)

	return Viewport{
		X:      (float32(windowWidth) - width) / 2,
		Y:      (float32(windowHeight) - height) / 2,
		Width:  width,
		Height: height,
	}, nil
}

// ScreenToGame maps native window pixels into logical game coordinates. Pixels in letterboxing are rejected so clicks
// on decorative margins cannot affect the world.
func (s *Service) ScreenToGame(x, y int) (gameX, gameY int, inside bool) {
	windowWidth, windowHeight := s.WindowSize()

	viewport, err := gameViewport(
		windowWidth,
		windowHeight,
		s.config.Resolution.Width,
		s.config.Resolution.Height,
		s.config.Resolution.Fit,
	)
	if err != nil {
		return 0, 0, false
	}

	return mapScreenToGame(viewport, s.config.Resolution.Width, s.config.Resolution.Height, x, y)
}

// mapScreenToGame applies one precomputed viewport using half-open bounds. The right and bottom edges remain outside,
// ensuring mapped coordinates never equal the logical width or height.
func mapScreenToGame(viewport Viewport, gameWidth, gameHeight, x, y int) (gameX, gameY int, inside bool) {
	outside := float32(x) < viewport.X ||
		float32(y) < viewport.Y ||
		float32(x) >= viewport.X+viewport.Width ||
		float32(y) >= viewport.Y+viewport.Height
	if outside {
		return 0, 0, false
	}

	gameX = int((float32(x) - viewport.X) * float32(gameWidth) / viewport.Width)
	gameY = int((float32(y) - viewport.Y) * float32(gameHeight) / viewport.Height)

	return gameX, gameY, true
}
