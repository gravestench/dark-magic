package raylibRenderer

import (
	"fmt"
	"image"
	"image/color"
	"sync"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/gravestench/dark-magic/internal/presentation/render"
)

// AttachComposer drains backend-neutral render changes on the Raylib owner
// thread immediately before the legacy scene graph update.
func (s *Service) AttachComposer(composer *render.Composer) error {
	if composer == nil {
		return fmt.Errorf("renderer: nil composition core")
	}
	s.compositionMu.Lock()
	defer s.compositionMu.Unlock()
	if s.composition != nil {
		return fmt.Errorf("renderer: composition core is already attached")
	}
	backend := &compositionBackend{renderer: s, nodes: make(map[render.NodeID]Renderable), resources: make(map[render.ResourceID]render.Resource), nodeResources: make(map[render.NodeID]render.ResourceID), playbacks: make(map[render.NodeID]*animationPlayback), paletteEffects: make(map[render.ResourceID]*gpuPaletteEffect)}
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
	renderer       *Service
	mu             sync.Mutex
	nodes          map[render.NodeID]Renderable
	resources      map[render.ResourceID]render.Resource
	nodeResources  map[render.NodeID]render.ResourceID
	playbacks      map[render.NodeID]*animationPlayback
	paletteEffects map[render.ResourceID]*gpuPaletteEffect
}

func (b *compositionBackend) closePaletteEffects() {
	if b == nil {
		return
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	for id, effect := range b.paletteEffects {
		effect.close()
		delete(b.paletteEffects, id)
	}
}

type animationPlayback struct {
	player       *render.AnimationPlayer
	frames       []image.Image
	seekRevision uint64
}

func (b *compositionBackend) Apply(change render.Change) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.nodes == nil {
		b.nodes = make(map[render.NodeID]Renderable)
	}
	if b.resources == nil {
		b.resources = make(map[render.ResourceID]render.Resource)
	}
	if b.nodeResources == nil {
		b.nodeResources = make(map[render.NodeID]render.ResourceID)
	}
	if b.playbacks == nil {
		b.playbacks = make(map[render.NodeID]*animationPlayback)
	}
	if b.paletteEffects == nil {
		b.paletteEffects = make(map[render.ResourceID]*gpuPaletteEffect)
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
		if resource.Kind != render.ResourceTexture || change.Resource.Kind != render.ResourceTexture {
			return fmt.Errorf("resource %v is not an updateable texture", change.ResourceID)
		}
		previous := resource.Payload.(image.Image).Bounds().Size()
		next := change.Resource.Payload.(image.Image).Bounds().Size()
		b.resources[change.ResourceID] = change.Resource
		for nodeID, resourceID := range b.nodeResources {
			if resourceID != change.ResourceID {
				continue
			}
			node := b.nodes[nodeID]
			// Streaming video frames retain one native texture. A resize is the
			// exceptional case that requires replacing its GPU allocation.
			if previous != next {
				node.ClearTextures()
			}
			node.SetImage(change.Resource.Payload.(image.Image))
		}
		return nil
	case "resource-destroy":
		if _, exists := b.resources[change.ResourceID]; !exists {
			return fmt.Errorf("resource %v does not exist", change.ResourceID)
		}
		if effect := b.paletteEffects[change.ResourceID]; effect != nil {
			effect.close()
			delete(b.paletteEffects, change.ResourceID)
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

func (b *compositionBackend) applyNode(node Renderable, state render.Node) error {
	if state.Parent != (render.NodeID{}) {
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
	if state.Palette == (render.ResourceID{}) {
		node.SetShader(nil)
	} else {
		effect := b.paletteEffects[state.Palette]
		if effect == nil {
			resource, exists := b.resources[state.Palette]
			if !exists || resource.Kind != render.ResourcePalette {
				return fmt.Errorf("palette resource %v is unavailable", state.Palette)
			}
			var err error
			effect, err = newGPUPaletteEffect(resource.Payload.(color.Palette))
			if err != nil {
				return err
			}
			b.paletteEffects[state.Palette] = effect
		}
		node.SetShader(&effect.shader)
	}
	if state.Resource != (render.ResourceID{}) {
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
			if resource.Kind == render.ResourceAnimation {
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
	// Resource-less retained nodes are grouping transforms, not drawable
	// surfaces. Enabling one would make raylib render its default 1x1 texture.
	if state.Visible && state.Resource != (render.ResourceID{}) {
		node.Enable()
	} else {
		node.Disable()
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

func (b *compositionBackend) attachAnimation(id render.NodeID, node Renderable, resource render.Resource) error {
	animation := resource.Payload.(render.AnimationData)
	frames := make([]image.Image, len(animation.Frames))
	for index, id := range animation.Frames {
		frame, exists := b.resources[id]
		if !exists {
			return fmt.Errorf("animation %v frame %d is unavailable", resource.ID, index)
		}
		frames[index] = frame.Payload.(image.Image)
	}
	player := render.NewAnimationPlayer(animation.Durations, animation.Loop)
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

func (b *compositionBackend) drawableImage(resource render.Resource) (image.Image, error) {
	switch resource.Kind {
	case render.ResourceTexture:
		return resource.Payload.(image.Image), nil
	case render.ResourceAnimation:
		animation := resource.Payload.(render.AnimationData)
		frame, exists := b.resources[animation.Frames[0]]
		if !exists {
			return nil, fmt.Errorf("animation %v frame is unavailable", resource.ID)
		}
		return frame.Payload.(image.Image), nil
	case render.ResourceRenderTarget:
		target := resource.Payload.(render.RenderTargetData)
		return image.NewRGBA(image.Rect(0, 0, target.Width, target.Height)), nil
	default:
		return nil, fmt.Errorf("resource kind %q is not drawable", resource.Kind)
	}
}
