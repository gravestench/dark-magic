package modruntime

import (
	"image"
	"testing"

	"github.com/gravestench/dark-magic/internal/presentation/render"
)

func TestRenderResourcePoolSharesAndReferenceCountsTextures(t *testing.T) {
	composer := &render.Composer{}
	pool := newRenderResourcePool(composer)
	pixels := image.NewRGBA(image.Rect(0, 0, 16, 8))
	first, releaseFirst, err := pool.acquire("dt1:floor:7", pixels)
	if err != nil {
		t.Fatal(err)
	}
	second, releaseSecond, err := pool.acquire("dt1:floor:7", pixels)
	if err != nil {
		t.Fatal(err)
	}
	if first != second {
		t.Fatalf("same immutable tile received different resources: %v and %v", first, second)
	}
	if got := composer.Diagnostics(); got.ActiveResources != 1 || got.RetainedTextureBytes != 16*8*4 {
		t.Fatalf("shared resource diagnostics = %#v", got)
	}
	if err := releaseFirst(); err != nil {
		t.Fatal(err)
	}
	if got := composer.Diagnostics().ActiveResources; got != 1 {
		t.Fatalf("first placement destroyed shared resource: active=%d", got)
	}
	if err := releaseSecond(); err != nil {
		t.Fatal(err)
	}
	if got := composer.Diagnostics().ActiveResources; got != 0 {
		t.Fatalf("last placement did not release shared resource: active=%d", got)
	}
	// Placement cleanup is allowed to run more than once during parent teardown.
	if err := releaseSecond(); err != nil {
		t.Fatalf("repeated release = %v", err)
	}
}
