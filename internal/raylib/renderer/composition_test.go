package raylibRenderer

import (
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/gravestench/dark-magic/internal/rendercore"
)

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
