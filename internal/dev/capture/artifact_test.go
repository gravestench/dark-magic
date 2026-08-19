package capture

import (
	"image"
	"image/color"
	"testing"
)

// TestDeathSceneAllowsSparseTextButRejectsBlankFrame protects the scene-specific threshold from leaking globally.
func TestDeathSceneAllowsSparseTextButRejectsBlankFrame(t *testing.T) {
	canvas := image.NewRGBA(image.Rect(0, 0, 800, 600))
	if hasVisiblePixels(canvas, "death") {
		t.Fatal("blank death frame was accepted")
	}

	for pixel := 0; pixel < 480; pixel++ {
		canvas.Set(pixel%800, pixel/800, color.RGBA{R: 180, A: 255})
	}

	if !hasVisiblePixels(canvas, "death") {
		t.Fatal("sparse death text was rejected")
	}

	if hasVisiblePixels(canvas, "title") {
		t.Fatal("sparse threshold leaked into ordinary scene capture")
	}
}

// TestScreenshotFilenamePreservesFormat makes sequence padding and scene normalization an explicit artifact contract.
func TestScreenshotFilenamePreservesFormat(t *testing.T) {
	name := screenshotFilename(3, "NPC Dialogue!")
	if name != "03-npc-dialogue.png" {
		t.Fatalf("screenshotFilename() = %q, want %q", name, "03-npc-dialogue.png")
	}
}
