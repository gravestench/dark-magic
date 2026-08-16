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
	backend := &compositionBackend{renderer: s, nodes: make(map[render.NodeID]*node), resources: make(map[render.ResourceID]render.Resource), nodeResources: make(map[render.NodeID]render.ResourceID), playbacks: make(map[render.NodeID]*animationPlayback), paletteEffects: make(map[render.ResourceID]*gpuPaletteEffect)}
	s.composition = composer
	s.compositionBackend = backend
	s.OnFrame(func() {
		started := time.Now()
		defer func() { s.lastFrameCompositionNS.Store(uint64(time.Since(started))) }()
		if err := s.applyTextureCacheBudget(); err != nil && s.logger != nil {
			s.logger.Error("applying texture cache budget", "error", err)
		}
		if err := composer.Drain(backend); err != nil && s.logger != nil {
			s.logger.Error("draining render composition", "error", err)
		}
		backend.advance(time.Duration(float64(time.Second) * float64(rl.GetFrameTime())))
	})
	// Warm optional textures only after the current frame has drawn. Every
	// texture the visible scene actually used is now newer in the LRU than
	// speculative work, so background warming cannot steal its priority.
	s.SubscribePostFrame(func() {
		// Byte limits control traffic; the short wall-clock limit controls visible
		// frame pacing on drivers where one upload is disproportionately slow.
		if err := composer.DrainWarmWithin(backend, s.textureUploadBudget.Load(), 2*time.Millisecond); err != nil && s.logger != nil {
			s.logger.Error("warming texture residency", "error", err)
		}
	})
	s.SubscribeOverlay(func() { s.drawResidencyDebug(composer) })
	return nil
}

type compositionBackend struct {
	renderer       *Service
	mu             sync.Mutex
	nodes          map[render.NodeID]*node
	resources      map[render.ResourceID]render.Resource
	nodeResources  map[render.NodeID]render.ResourceID
	playbacks      map[render.NodeID]*animationPlayback
	paletteEffects map[render.ResourceID]*gpuPaletteEffect
}

func (b *compositionBackend) CanWarmTexture(key string, weight uint64) bool {
	if b == nil || b.renderer == nil || b.renderer.cache == nil || weight > uint64(^uint(0)>>1) {
		return false
	}
	return b.renderer.cache.CanInsertWithoutEviction(key, int(weight))
}

// close releases every backend-owned node while the native graphics context is
// still alive. The composer owns semantic lifetimes; this method is the final
// adapter-side safety net during application shutdown.
func (b *compositionBackend) close() {
	if b == nil {
		return
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	for id, node := range b.nodes {
		node.Disable()
		node.setParent(nil)
		node.ClearTextures()
		delete(b.nodes, id)
	}
	clear(b.resources)
	clear(b.nodeResources)
	clear(b.playbacks)
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
	keys         []string
	seekRevision uint64
}

func (b *compositionBackend) advance(elapsed time.Duration) {
	b.mu.Lock()
	defer b.mu.Unlock()
	for id, playback := range b.playbacks {
		index, changed := playback.player.Advance(elapsed)
		if changed {
			b.setAnimationFrame(b.nodes[id], playback.frames[index], playback.keys[index], index)
		}
	}
}

func (b *compositionBackend) Apply(change render.Change) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.nodes == nil {
		b.nodes = make(map[render.NodeID]*node)
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
	case "texture-warm":
		if change.Resource.TextureKey == "" {
			return fmt.Errorf("warm texture key is empty")
		}
		b.renderer.getTexture(change.Resource.TextureKey, change.Resource.Payload.(image.Image))
		return nil
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
			node.UpdateImageResource(change.Resource.Payload.(image.Image), change.Resource.TextureKey)
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
		node := b.renderer.newNode()
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
		node.setParent(nil)
		node.ClearTextures()
		delete(b.nodes, change.ID)
		delete(b.nodeResources, change.ID)
		delete(b.playbacks, change.ID)
		return nil
	default:
		return fmt.Errorf("unknown composition change %q", change.Kind)
	}
}

func (b *compositionBackend) applyNode(node *node, state render.Node) error {
	if state.Parent != (render.NodeID{}) {
		parent, exists := b.nodes[state.Parent]
		if !exists {
			return fmt.Errorf("parent node %v does not exist", state.Parent)
		}
		node.setParent(parent)
	}
	node.SetPosition(float32(state.X), float32(state.Y))
	node.setScaleXY(float32(state.ScaleX), float32(state.ScaleY))
	node.SetRotation(float32(state.Rotation))
	node.SetOrigin(state.OriginX, state.OriginY)
	// RGB tint is deliberately independent of opacity. Disabled UI can be
	// visibly dark without becoming translucent over the scene behind it.
	tint := state.Tint
	if tint.A == 0 {
		tint = color.RGBA{R: 255, G: 255, B: 255, A: 255}
	}
	node.SetTint(rl.NewColor(tint.R, tint.G, tint.B, 255))
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
	case "screen":
		// Diablo II draw mode 3: GL_ONE, GL_ONE_MINUS_SRC_COLOR.
		// renderNode installs the custom factors immediately before drawing.
		node.SetBlendMode(rl.BlendCustom)
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
		node.SetShader(nil, nil, 0)
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
		node.SetShader(&effect.shader, &effect.texture, effect.textureLocation)
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
				node.SetImageResource(decoded, resource.TextureKey)
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
	node.setVisible(state.Visible)
	if playback := b.playbacks[state.ID]; playback != nil {
		playback.player.SetPaused(state.AnimationPaused)
		if playback.seekRevision != state.AnimationSeekRevision {
			frame, changed := playback.player.Seek(state.AnimationSeek)
			playback.seekRevision = state.AnimationSeekRevision
			if changed {
				b.setAnimationFrame(node, playback.frames[frame], playback.keys[frame], frame)
			}
		}
	}
	return nil
}

func (b *compositionBackend) attachAnimation(id render.NodeID, node *node, resource render.Resource) error {
	animation := resource.Payload.(render.AnimationData)
	frames := make([]image.Image, len(animation.Frames))
	keys := make([]string, len(animation.Frames))
	for index, id := range animation.Frames {
		frame, exists := b.resources[id]
		if !exists {
			return fmt.Errorf("animation %v frame %d is unavailable", resource.ID, index)
		}
		frames[index] = frame.Payload.(image.Image)
		keys[index] = frame.TextureKey
	}
	player := render.NewAnimationPlayer(animation.Durations, animation.Loop)
	b.playbacks[id] = &animationPlayback{player: player, frames: frames, keys: keys}
	node.SetAnimationFrame(frames[0], keys[0], 0)
	return nil
}

func (b *compositionBackend) setAnimationFrame(node *node, frame image.Image, key string, index int) {
	node.SetAnimationFrame(frame, key, index)
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
