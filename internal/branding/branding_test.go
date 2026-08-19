package branding

import (
	"bytes"
	"image/png"
	"testing"
)

// TestWindowIconPNG protects the minimum window-icon contract expected by rendering backends.
// Decoding the embedded bytes also catches packaging mistakes that would otherwise surface only at client startup.
func TestWindowIconPNG(t *testing.T) {
	decodedIcon, err := png.Decode(bytes.NewReader(WindowIconPNG()))
	if err != nil {
		t.Fatalf("decode window icon: %v", err)
	}

	// Preserve enough source resolution for each backend to downscale one square asset without distortion.
	iconBounds := decodedIcon.Bounds()
	if iconBounds.Dx() != iconBounds.Dy() || iconBounds.Dx() < 256 {
		t.Fatalf("window icon bounds = %v, want a square icon at least 256 pixels wide", iconBounds)
	}
}
