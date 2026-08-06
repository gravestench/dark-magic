package videocore

import (
	"image"
	"math"
	"testing"

	"github.com/gravestench/dark-magic/internal/rendercore"
)

func TestPresenterLetterboxesAndReusesOneTexture(t *testing.T) {
	var composer rendercore.Composer
	presenter, err := NewPresenter(&composer, image.Pt(640, 292), image.Pt(640, 480))
	if err != nil {
		t.Fatal(err)
	}
	nodes := composer.Snapshot()
	if len(nodes) != 1 || nodes[0].Layer != rendercore.LayerTransition {
		t.Fatalf("nodes = %#v", nodes)
	}
	if nodes[0].X != 0 || math.Abs(nodes[0].Y-94) > 0.001 || nodes[0].ScaleX != 1 {
		t.Fatalf("initial fit = x %.2f y %.2f scale %.2f", nodes[0].X, nodes[0].Y, nodes[0].ScaleX)
	}
	before := composer.Diagnostics()
	frame := image.NewRGBA(image.Rect(0, 0, 640, 292))
	if err := presenter.Present(frame); err != nil {
		t.Fatal(err)
	}
	after := composer.Diagnostics()
	if before.ActiveNodes != after.ActiveNodes || before.ActiveResources != after.ActiveResources {
		t.Fatalf("presentation allocated retained objects: before %#v after %#v", before, after)
	}
	if err := presenter.Resize(image.Pt(800, 600)); err != nil {
		t.Fatal(err)
	}
	node := composer.Snapshot()[0]
	if math.Abs(node.X) > 0.001 || math.Abs(node.Y-117.5) > 0.001 || math.Abs(node.ScaleX-1.25) > 0.001 {
		t.Fatalf("resized fit = x %.2f y %.2f scale %.2f", node.X, node.Y, node.ScaleX)
	}
	if err := presenter.Close(); err != nil {
		t.Fatal(err)
	}
	if got := composer.Diagnostics(); got.ActiveNodes != 0 || got.ActiveResources != 0 {
		t.Fatalf("presenter leaked retained objects: %#v", got)
	}
	if err := presenter.Present(frame); err == nil {
		t.Fatal("closed presenter accepted a frame")
	}
}
