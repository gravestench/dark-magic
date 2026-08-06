package raylibRenderer

import (
	"image"
	"image/color"
	"testing"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/gravestench/dark-magic/internal/rendercore"
)

func TestCompositionBackendOwnsEveryManagedResourceKind(t *testing.T) {
	backend := &compositionBackend{resources: make(map[rendercore.ResourceID]rendercore.Resource)}
	texture := rendercore.Resource{ID: rendercore.ResourceID{Slot: 1, Generation: 1}, Kind: rendercore.ResourceTexture, Payload: image.NewRGBA(image.Rect(0, 0, 1, 1))}
	resources := []rendercore.Resource{
		texture,
		{ID: rendercore.ResourceID{Slot: 2, Generation: 1}, Kind: rendercore.ResourcePalette, Payload: color.Palette{color.Black}},
		{ID: rendercore.ResourceID{Slot: 3, Generation: 1}, Kind: rendercore.ResourceFont, Payload: rendercore.FontData{Bytes: []byte("font"), Size: 12}},
		{ID: rendercore.ResourceID{Slot: 4, Generation: 1}, Kind: rendercore.ResourceAnimation, Payload: rendercore.AnimationData{Frames: []rendercore.ResourceID{texture.ID}, Durations: []time.Duration{time.Second}}},
		{ID: rendercore.ResourceID{Slot: 5, Generation: 1}, Kind: rendercore.ResourceRenderTarget, Payload: rendercore.RenderTargetData{Width: 2, Height: 2}},
	}
	for _, resource := range resources {
		if err := backend.Apply(rendercore.Change{Kind: "resource-create", Resource: resource, ResourceID: resource.ID}); err != nil {
			t.Fatalf("create %s: %v", resource.Kind, err)
		}
	}
	for index := len(resources) - 1; index >= 0; index-- {
		if err := backend.Apply(rendercore.Change{Kind: "resource-destroy", ResourceID: resources[index].ID}); err != nil {
			t.Fatal(err)
		}
	}
}

func TestCompositionBackendMirrorsCheckedNodes(t *testing.T) {
	t.Parallel()

	renderer := &Service{}
	renderer.rootNode = renderer.NewRenderable()
	renderer.rootNode.Disable()
	backend := &compositionBackend{renderer: renderer, nodes: make(map[rendercore.NodeID]Renderable)}
	parent := rendercore.NodeID{Slot: 1, Generation: 1}
	child := rendercore.NodeID{Slot: 2, Generation: 1}
	if err := backend.Apply(rendercore.Change{Kind: "create", ID: parent, Node: rendercore.Node{ID: parent, Layer: rendercore.LayerHUD, ScaleX: 1, Visible: true}}); err != nil {
		t.Fatal(err)
	}
	if err := backend.Apply(rendercore.Change{Kind: "create", ID: child, Node: rendercore.Node{ID: child, Parent: parent, X: 12, Y: 34, ScaleX: 2, Visible: true}}); err != nil {
		t.Fatal(err)
	}
	renderer.rootNode.UpdateWorldMatrix(rl.MatrixIdentity())
	childNode := backend.nodes[child]
	if childNode.Parent() != backend.nodes[parent] || childNode.Scale() != 2 {
		t.Fatalf("child was not attached: parent=%v scale=%v", childNode.Parent(), childNode.Scale())
	}
	if err := backend.Apply(rendercore.Change{Kind: "destroy", ID: child}); err != nil {
		t.Fatal(err)
	}
	if _, exists := backend.nodes[child]; exists {
		t.Fatal("destroyed node remains in backend")
	}
}
