package branding

import (
	"bytes"
	"image/png"
	"testing"
)

func TestWindowIconPNG(t *testing.T) {
	icon, err := png.Decode(bytes.NewReader(WindowIconPNG()))
	if err != nil {
		t.Fatalf("decode window icon: %v", err)
	}
	if bounds := icon.Bounds(); bounds.Dx() != bounds.Dy() || bounds.Dx() < 256 {
		t.Fatalf("window icon bounds = %v, want a square icon at least 256 pixels wide", bounds)
	}
}
