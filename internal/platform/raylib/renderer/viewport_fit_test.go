package raylibRenderer

import "testing"

// TestGameViewportContainsAndCentersAspectRatio verifies centered letterboxing when window and game aspect ratios
// differ.
func TestGameViewportContainsAndCentersAspectRatio(t *testing.T) {
	viewport, err := gameViewport(1920, 1080, 800, 600, "contain")
	if err != nil {
		t.Fatal(err)
	}

	if viewport.X != 240 || viewport.Y != 0 || viewport.Width != 1440 || viewport.Height != 1080 {
		t.Fatalf("viewport = %#v", viewport)
	}
}

// TestGameViewportCanStretch verifies the explicit compatibility mode that fills every native window pixel.
func TestGameViewportCanStretch(t *testing.T) {
	viewport, err := gameViewport(1920, 1080, 800, 600, "stretch")
	if err != nil || viewport != (Viewport{Width: 1920, Height: 1080}) {
		t.Fatalf("viewport = %#v, %v", viewport, err)
	}
}

// TestScreenToGameRejectsLetterboxAndMapsViewport verifies that input uses the same letterbox geometry as presentation.
func TestScreenToGameRejectsLetterboxAndMapsViewport(t *testing.T) {
	viewport, _ := gameViewport(1920, 1080, 800, 600, "contain")
	if _, _, inside := mapScreenToGame(viewport, 800, 600, 100, 500); inside {
		t.Fatal("letterbox was treated as game input")
	}

	x, y, inside := mapScreenToGame(viewport, 800, 600, 960, 540)
	if !inside || x != 400 || y != 300 {
		t.Fatalf("center = %d,%d inside=%v", x, y, inside)
	}
}

// TestNativeScreenToGameAccountsForHighDPIDrawableScale keeps pointer coordinates aligned on Retina displays.
func TestNativeScreenToGameAccountsForHighDPIDrawableScale(t *testing.T) {
	x, y, inside := mapNativeScreenToGame(3200, 2000, 1600, 1000, 800, 500)
	if !inside || x != 1600 || y != 1000 {
		t.Fatalf("native center = %d,%d inside=%v", x, y, inside)
	}
	if _, _, inside := mapNativeScreenToGame(3200, 2000, 1600, 1000, 1600, 20); inside {
		t.Fatal("native right edge was treated as in bounds")
	}
}

// TestLogicalMonitorSizeConvertsRetinaVideoModeToWindowPoints protects visible window placement on high-DPI monitors.
func TestLogicalMonitorSizeConvertsRetinaVideoModeToWindowPoints(t *testing.T) {
	width, height := logicalMonitorSize(3024, 1964, 2, 2)
	if width != 1512 || height != 982 {
		t.Fatalf("logical monitor = %dx%d", width, height)
	}
}

// TestMacOSWindowWorkAreaKeepsClientChromeClearOfSystemUI verifies that the
// initial window is centered within Cocoa's visible frame rather than the raw mode.
func TestMacOSWindowWorkAreaKeepsClientChromeClearOfSystemUI(t *testing.T) {
	x, y, width, height := macOSWindowWorkArea(1512, 0, 1512, 982)
	if x != 1512 || y != 24 || width != 1512 || height != 866 {
		t.Fatalf("macOS work area = %d,%d %dx%d", x, y, width, height)
	}

	windowWidth, windowHeight, windowX, windowY := fitWindowToMonitor(1440, 900, width, height, x, y)
	if windowWidth != 1440 || windowHeight != 866 || windowX != 1548 || windowY != 24 {
		t.Fatalf("fitted macOS window = %d,%d %dx%d", windowX, windowY, windowWidth, windowHeight)
	}
}

// TestFitWindowToMonitorCentersWithinSecondaryDisplay covers clamping and non-primary desktop origins.
func TestFitWindowToMonitorCentersWithinSecondaryDisplay(t *testing.T) {
	width, height, x, y := fitWindowToMonitor(1600, 1000, 1512, 982, 1512, 0)
	if width != 1512 || height != 982 || x != 1512 || y != 0 {
		t.Fatalf("fit = %dx%d at %d,%d", width, height, x, y)
	}

	width, height, x, y = fitWindowToMonitor(1000, 700, 1440, 900, -1440, 40)
	if width != 1000 || height != 700 || x != -1220 || y != 140 {
		t.Fatalf("secondary fit = %dx%d at %d,%d", width, height, x, y)
	}
}
