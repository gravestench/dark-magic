package raylibRenderer

import (
	"fmt"
	"image"
	"sync"

	rl "github.com/gen2brain/raylib-go/raylib"
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
	backend := &compositionBackend{renderer: s, nodes: make(map[rendercore.NodeID]Renderable), resources: make(map[rendercore.ResourceID]rendercore.Resource), nodeResources: make(map[rendercore.NodeID]rendercore.ResourceID)}
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
	renderer      *Service
	mu            sync.Mutex
	nodes         map[rendercore.NodeID]Renderable
	resources     map[rendercore.ResourceID]rendercore.Resource
	nodeResources map[rendercore.NodeID]rendercore.ResourceID
}

func (b *compositionBackend) Apply(change rendercore.Change) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.nodes == nil {
		b.nodes = make(map[rendercore.NodeID]Renderable)
	}
	if b.resources == nil {
		b.resources = make(map[rendercore.ResourceID]rendercore.Resource)
	}
	if b.nodeResources == nil {
		b.nodeResources = make(map[rendercore.NodeID]rendercore.ResourceID)
	}
	switch change.Kind {
	case "resource-create":
		b.resources[change.ResourceID] = change.Resource
		return nil
	case "resource-destroy":
		if _, exists := b.resources[change.ResourceID]; !exists {
			return fmt.Errorf("resource %v does not exist", change.ResourceID)
		}
		delete(b.resources, change.ResourceID)
		return nil
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
		if b.renderer.cache != nil {
			b.renderer.cache.Remove(node.UUID().String())
		}
		delete(b.nodes, change.ID)
		delete(b.nodeResources, change.ID)
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
	switch state.Blend {
	case "", "alpha":
		node.SetBlendMode(rl.BlendAlpha)
	case "additive":
		node.SetBlendMode(rl.BlendAdditive)
	case "multiply":
		node.SetBlendMode(rl.BlendMultiplied)
	case "add-colors":
		node.SetBlendMode(rl.BlendAddColors)
	case "subtract-colors":
		node.SetBlendMode(rl.BlendSubtractColors)
	default:
		return fmt.Errorf("unsupported blend mode %q", state.Blend)
	}
	node.SetZIndex(float32(int(state.Layer)*1_000_000 + state.Z))
	if state.Visible {
		node.Enable()
	} else {
		node.Disable()
	}
	if state.Resource != (rendercore.ResourceID{}) {
		resource, exists := b.resources[state.Resource]
		if !exists {
			return fmt.Errorf("resource %v does not exist", state.Resource)
		}
		decoded, err := b.drawableImage(resource)
		if err != nil {
			return err
		}
		if b.nodeResources[state.ID] != state.Resource {
			if b.renderer.cache != nil {
				b.renderer.cache.Remove(node.UUID().String())
			}
			node.SetImage(decoded)
			b.nodeResources[state.ID] = state.Resource
		}
	}
	return nil
}

func (b *compositionBackend) drawableImage(resource rendercore.Resource) (image.Image, error) {
	switch resource.Kind {
	case rendercore.ResourceTexture:
		return resource.Payload.(image.Image), nil
	case rendercore.ResourceAnimation:
		animation := resource.Payload.(rendercore.AnimationData)
		frame, exists := b.resources[animation.Frames[0]]
		if !exists {
			return nil, fmt.Errorf("animation %v frame is unavailable", resource.ID)
		}
		return frame.Payload.(image.Image), nil
	case rendercore.ResourceRenderTarget:
		target := resource.Payload.(rendercore.RenderTargetData)
		return image.NewRGBA(image.Rect(0, 0, target.Width, target.Height)), nil
	default:
		return nil, fmt.Errorf("resource kind %q is not drawable", resource.Kind)
	}
}
