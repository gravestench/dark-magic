package maprender

import (
	"image"
	"testing"

	"github.com/gravestench/dark-magic/internal/game/world"
)

func TestCollisionDiamondIncludesItsProjectedCenter(t *testing.T) {
	shade, visible := collisionColor(world.Flags{BlockWalk: true})
	if !visible {
		t.Fatal("blocked cell has no diagnostic color")
	}
	pixels := image.NewRGBA(image.Rect(0, 0, 64, 32))
	fillDiamond(pixels, 32, 16, 16, 8, shade)
	red, _, _, alpha := pixels.At(32, 16).RGBA()
	if red == 0 || alpha == 0 {
		t.Fatal("projected collision center is transparent")
	}
}
