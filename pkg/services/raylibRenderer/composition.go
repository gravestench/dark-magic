package raylibRenderer

import (
	"fmt"
	"sync"

	"github.com/gravestench/dark-magic/internal/rendercore"
)

// AttachComposer drains backend-neutral render changes on the Raylib owner
// thread immediately before the legacy scene graph update.
func (s *Service) AttachComposer(composer *rendercore.Composer) error {
	if composer == nil {
		return fmt.Errorf("renderer: nil composition core")
	}
	s.compositionMu.Lock()
	defer s.compositionMu.Unlock()
	if s.composition != nil {
		return fmt.Errorf("renderer: composition core is already attached")
	}
	backend := &compositionBackend{renderer: s, nodes: make(map[rendercore.NodeID]Renderable)}
	s.composition = composer
	s.compositionBackend = backend
	s.OnFrame(func() {
		if err := composer.Drain(backend); err != nil && s.logger != nil {
			s.logger.Error("draining render composition", "error", err)
		}
	})
	return nil
}

type compositionBackend struct {
	renderer *Service
	mu       sync.Mutex
	nodes    map[rendercore.NodeID]Renderable
}

func (b *compositionBackend) Apply(change rendercore.Change) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	switch change.Kind {
	case "create":
		if _, exists := b.nodes[change.ID]; exists {
			return fmt.Errorf("node %v already exists", change.ID)
		}
		node := b.renderer.NewRenderable()
		b.nodes[change.ID] = node
		return b.applyNode(node, change.Node)
	case "update":
		node, exists := b.nodes[change.ID]
		if !exists {
			return fmt.Errorf("node %v does not exist", change.ID)
		}
		return b.applyNode(node, change.Node)
	case "destroy":
		node, exists := b.nodes[change.ID]
		if !exists {
			return fmt.Errorf("node %v does not exist", change.ID)
		}
		node.Disable()
		node.SetParent(nil)
		delete(b.nodes, change.ID)
		return nil
	default:
		return fmt.Errorf("unknown composition change %q", change.Kind)
	}
}

func (b *compositionBackend) applyNode(node Renderable, state rendercore.Node) error {
	if state.Parent != (rendercore.NodeID{}) {
		parent, exists := b.nodes[state.Parent]
		if !exists {
			return fmt.Errorf("parent node %v does not exist", state.Parent)
		}
		node.SetParent(parent)
	}
	node.SetPosition(float32(state.X), float32(state.Y))
	node.SetScale(float32(state.ScaleX))
	node.SetRotation(float32(state.Rotation))
	node.SetZIndex(float32(int(state.Layer)*1_000_000 + state.Z))
	if state.Visible {
		node.Enable()
	} else {
		node.Disable()
	}
	return nil
}
