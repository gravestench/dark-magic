package raylibRenderer

import "testing"

func TestGameViewportContainsAndCentersAspectRatio(t *testing.T) {
	viewport, err := gameViewport(1920, 1080, 800, 600, "contain")
	if err != nil {
		t.Fatal(err)
	}
	if viewport.X != 240 || viewport.Y != 0 || viewport.Width != 1440 || viewport.Height != 1080 {
		t.Fatalf("viewport = %#v", viewport)
	}
}

func TestGameViewportCanStretch(t *testing.T) {
	viewport, err := gameViewport(1920, 1080, 800, 600, "stretch")
	if err != nil || viewport != (Viewport{Width: 1920, Height: 1080}) {
		t.Fatalf("viewport = %#v, %v", viewport, err)
	}
}

func TestScreenToGameRejectsLetterboxAndMapsViewport(t *testing.T) {
	viewport, _ := gameViewport(1920, 1080, 800, 600, "contain")
	if _, _, inside := mapScreenToGame(viewport, 800, 600, 100, 500); inside {
		t.Fatal("letterbox was treated as game input")
	}
	if x, y, inside := mapScreenToGame(viewport, 800, 600, 960, 540); !inside || x != 400 || y != 300 {
		t.Fatalf("center = %d,%d inside=%v", x, y, inside)
	}
}
