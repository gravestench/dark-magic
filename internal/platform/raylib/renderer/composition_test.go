package raylibRenderer

import (
	"image"
	"image/color"
	"testing"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/gravestench/dark-magic/internal/presentation/render"
)

func TestCompositionBackendOwnsEveryManagedResourceKind(t *testing.T) {
	backend := &compositionBackend{resources: make(map[render.ResourceID]render.Resource)}
	texture := render.Resource{ID: render.ResourceID{Slot: 1, Generation: 1}, Kind: render.ResourceTexture, Payload: image.NewRGBA(image.Rect(0, 0, 1, 1))}
	resources := []render.Resource{
		texture,
		{ID: render.ResourceID{Slot: 2, Generation: 1}, Kind: render.ResourcePalette, Payload: color.Palette{color.Black}},
		{ID: render.ResourceID{Slot: 3, Generation: 1}, Kind: render.ResourceFont, Payload: render.FontData{Bytes: []byte("font"), Size: 12}},
		{ID: render.ResourceID{Slot: 4, Generation: 1}, Kind: render.ResourceAnimation, Payload: render.AnimationData{Frames: []render.ResourceID{texture.ID}, Durations: []time.Duration{time.Second}}},
		{ID: render.ResourceID{Slot: 5, Generation: 1}, Kind: render.ResourceRenderTarget, Payload: render.RenderTargetData{Width: 2, Height: 2}},
	}
	for _, resource := range resources {
		if err := backend.Apply(render.Change{Kind: "resource-create", Resource: resource, ResourceID: resource.ID}); err != nil {
			t.Fatalf("create %s: %v", resource.Kind, err)
		}
	}
	for index := len(resources) - 1; index >= 0; index-- {
		if err := backend.Apply(render.Change{Kind: "resource-destroy", ResourceID: resources[index].ID}); err != nil {
			t.Fatal(err)
		}
	}
}

func TestCompositionBackendMirrorsCheckedNodes(t *testing.T) {
	t.Parallel()

	renderer := &Service{}
	renderer.rootNode = renderer.newNode()
	renderer.rootNode.Disable()
	backend := &compositionBackend{renderer: renderer, nodes: make(map[render.NodeID]*node)}
	parent := render.NodeID{Slot: 1, Generation: 1}
	child := render.NodeID{Slot: 2, Generation: 1}
	if err := backend.Apply(render.Change{Kind: "create", ID: parent, Node: render.Node{ID: parent, Layer: render.LayerHUD, ScaleX: 1, Visible: true}}); err != nil {
		t.Fatal(err)
	}
	if err := backend.Apply(render.Change{Kind: "create", ID: child, Node: render.Node{ID: child, Parent: parent, X: 12, Y: 34, ScaleX: 2, Visible: true}}); err != nil {
		t.Fatal(err)
	}
	renderer.rootNode.UpdateWorldMatrix(rl.MatrixIdentity(), true)
	childNode := backend.nodes[child]
	if childNode.parentNode() != backend.nodes[parent] || childNode.Scale() != 2 {
		t.Fatalf("child was not attached: parent=%v scale=%v", childNode.parentNode(), childNode.Scale())
	}
	if backend.nodes[parent].IsEnabled() || childNode.IsEnabled() {
		t.Fatal("resource-less grouping nodes were enabled for default-texture drawing")
	}
	if err := backend.Apply(render.Change{Kind: "destroy", ID: child}); err != nil {
		t.Fatal(err)
	}
	if _, exists := backend.nodes[child]; exists {
		t.Fatal("destroyed node remains in backend")
	}
}

func TestCompositionBackendMapsDiabloScreenBlend(t *testing.T) {
	renderer := &Service{}
	renderer.rootNode = renderer.newNode()
	renderer.rootNode.Disable()
	backend := &compositionBackend{renderer: renderer, nodes: make(map[render.NodeID]*node)}
	id := render.NodeID{Slot: 1, Generation: 1}
	if err := backend.Apply(render.Change{Kind: "create", ID: id, Node: render.Node{ID: id, Layer: render.LayerHUD, ScaleX: 1, Blend: "screen"}}); err != nil {
		t.Fatal(err)
	}
	if got := backend.nodes[id].BlendMode(); got != rl.BlendCustom {
		t.Fatalf("screen blend mapped to %v, want BlendCustom", got)
	}
}

func TestBackendUsesNonUniformScaleAndDirtyTransformPropagation(t *testing.T) {
	renderer := &Service{}
	renderer.rootNode = renderer.newNode()
	backend := &compositionBackend{renderer: renderer, nodes: make(map[render.NodeID]*node)}
	parent := render.NodeID{Slot: 1, Generation: 1}
	child := render.NodeID{Slot: 2, Generation: 1}
	if err := backend.Apply(render.Change{Kind: "create", ID: parent, Node: render.Node{ID: parent, ScaleX: 2, ScaleY: 3, Visible: true}}); err != nil {
		t.Fatal(err)
	}
	if err := backend.Apply(render.Change{Kind: "create", ID: child, Node: render.Node{ID: child, Parent: parent, X: 4, Y: 5, ScaleX: 1, ScaleY: 1, Visible: true}}); err != nil {
		t.Fatal(err)
	}
	renderer.rootNode.UpdateWorldMatrix(rl.MatrixIdentity(), false)
	x, y := backend.nodes[child].Position()
	if x != 8 || y != 15 {
		t.Fatalf("child world position = %v,%v, want 8,15", x, y)
	}
	if backend.nodes[child].transformDirty {
		t.Fatal("transform remained dirty after propagation")
	}
}

func TestFinalDrainLeavesNoPendingCompositionCommands(t *testing.T) {
	renderer := &Service{}
	renderer.rootNode = renderer.newNode()
	backend := &compositionBackend{renderer: renderer, nodes: make(map[render.NodeID]*node)}
	composer := &render.Composer{}
	id, err := composer.Create(render.NodeID{}, render.LayerHUD)
	if err != nil {
		t.Fatal(err)
	}
	if err := composer.Destroy(id); err != nil {
		t.Fatal(err)
	}
	if err := composer.Drain(backend); err != nil {
		t.Fatal(err)
	}
	backend.close()
	if diagnostics := composer.Diagnostics(); diagnostics.Pending != 0 {
		t.Fatalf("pending commands after final drain = %d", diagnostics.Pending)
	}
	if len(backend.nodes) != 0 {
		t.Fatalf("backend nodes after close = %d", len(backend.nodes))
	}
}
