package raylibRenderer

import (
	"image"
	"testing"

	"github.com/google/uuid"
)

func TestAnimationFramesUseStableTextureVariants(t *testing.T) {
	node := &node{uuid: uuid.New()}
	frame := image.NewRGBA(image.Rect(0, 0, 1, 1))
	node.SetAnimationFrame(frame, "", 0)
	first := node.textureVariant
	if node.dirty() {
		t.Fatal("animation frame requested an incremental texture upload")
	}
	node.SetAnimationFrame(frame, "", 1)
	node.SetAnimationFrame(frame, "", 0)
	if node.textureVariant != first || len(node.textureKeys) != 2 {
		t.Fatalf("variant = %q, keys = %d", node.textureVariant, len(node.textureKeys))
	}
	node.SetImage(frame)
	if node.textureVariant != "" || !node.dirty() {
		t.Fatal("static image did not restore the mutable base texture")
	}
}
