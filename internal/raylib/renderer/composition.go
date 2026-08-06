package raylibRenderer

import (
	"fmt"
	"image"
	"sync"
	"time"

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
	backend := &compositionBackend{renderer: s, nodes: make(map[rendercore.NodeID]Renderable), resources: make(map[rendercore.ResourceID]rendercore.Resource), nodeResources: make(map[rendercore.NodeID]rendercore.ResourceID), playbacks: make(map[rendercore.NodeID]*animationPlayback)}
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
	playbacks     map[rendercore.NodeID]*animationPlayback
}

type animationPlayback struct {
	player       *rendercore.AnimationPlayer
	frames       []image.Image
	seekRevision uint64
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
	if b.playbacks == nil {
		b.playbacks = make(map[rendercore.NodeID]*animationPlayback)
	}
	switch change.Kind {
	case "resource-create":
		b.resources[change.ResourceID] = change.Resource
		return nil
	case "resource-update":
		resource, exists := b.resources[change.ResourceID]
		if !exists {
			return fmt.Errorf("resource %v does not exist", change.ResourceID)
		}
		if resource.Kind != rendercore.ResourceTexture || change.Resource.Kind != rendercore.ResourceTexture {
			return fmt.Errorf("resource %v is not an updateable texture", change.ResourceID)
		}
		b.resources[change.ResourceID] = change.Resource
		for nodeID, resourceID := range b.nodeResources {
			if resourceID != change.ResourceID {
				continue
			}
			node := b.nodes[nodeID]
			node.ClearTextures()
			node.SetImage(change.Resource.Payload.(image.Image))
		}
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
		node.ClearTextures()
		delete(b.nodes, change.ID)
		delete(b.nodeResources, change.ID)
		delete(b.playbacks, change.ID)
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
	if state.Clip == nil {
		node.SetClip(nil)
	} else {
		clip := rl.NewRectangle(float32(state.Clip.X), float32(state.Clip.Y), float32(state.Clip.Width), float32(state.Clip.Height))
		node.SetClip(&clip)
	}
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
			node.ClearTextures()
			if resource.Kind == rendercore.ResourceAnimation {
				if err := b.attachAnimation(state.ID, node, resource); err != nil {
					return err
				}
			} else {
				node.SetImage(decoded)
				node.OnUpdate(nil)
				delete(b.playbacks, state.ID)
			}
			b.nodeResources[state.ID] = state.Resource
		}
	}
	if playback := b.playbacks[state.ID]; playback != nil {
		playback.player.SetPaused(state.AnimationPaused)
		if playback.seekRevision != state.AnimationSeekRevision {
			frame, changed := playback.player.Seek(state.AnimationSeek)
			playback.seekRevision = state.AnimationSeekRevision
			if changed {
				b.setAnimationFrame(node, playback.frames[frame], frame)
			}
		}
	}
	return nil
}

func (b *compositionBackend) attachAnimation(id rendercore.NodeID, node Renderable, resource rendercore.Resource) error {
	animation := resource.Payload.(rendercore.AnimationData)
	frames := make([]image.Image, len(animation.Frames))
	for index, id := range animation.Frames {
		frame, exists := b.resources[id]
		if !exists {
			return fmt.Errorf("animation %v frame %d is unavailable", resource.ID, index)
		}
		frames[index] = frame.Payload.(image.Image)
	}
	player := rendercore.NewAnimationPlayer(animation.Durations, animation.Loop)
	b.playbacks[id] = &animationPlayback{player: player, frames: frames}
	node.SetAnimationFrame(frames[0], 0)
	node.OnUpdate(func() {
		index, changed := player.Advance(time.Duration(float64(time.Second) * float64(rl.GetFrameTime())))
		if changed {
			b.setAnimationFrame(node, frames[index], index)
		}
	})
	return nil
}

func (b *compositionBackend) setAnimationFrame(node Renderable, frame image.Image, index int) {
	node.SetAnimationFrame(frame, index)
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
