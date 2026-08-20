package raylibRenderer

import (
	"fmt"
	"sync"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/gravestench/dark-magic/internal/presentation/render"
)

// AttachComposer installs one retained composition source whose native changes drain on the Raylib owner thread.
func (s *Service) AttachComposer(composer *render.Composer) error {
	if composer == nil {
		return fmt.Errorf("renderer: nil composition core")
	}

	s.compositionMu.Lock()
	defer s.compositionMu.Unlock()

	if s.composition != nil {
		return fmt.Errorf("renderer: composition core is already attached")
	}

	backend := newCompositionBackend(s)
	s.composition = composer
	s.compositionBackend = backend

	// These subscriptions preserve frame ordering: visible changes, drawing, optional warming, then diagnostics overlay.
	s.OnFrame(func() {
		started := time.Now()
		defer func() {
			s.lastFrameCompositionNS.Store(uint64(time.Since(started)))
		}()

		s.runCompositionFrame(composer, backend)
	})
	s.SubscribePostFrame(func() {
		s.warmCompositionTextures(composer, backend)
	})
	s.SubscribeOverlay(func() {
		s.drawResidencyDebug(composer)
	})

	return nil
}

// newCompositionBackend initializes all ownership maps together so command application has no partial state.
func newCompositionBackend(renderer *Service) *compositionBackend {
	return &compositionBackend{
		renderer:       renderer,
		nodes:          make(map[render.NodeID]*node),
		resources:      make(map[render.ResourceID]render.Resource),
		nodeResources:  make(map[render.NodeID]render.ResourceID),
		playbacks:      make(map[render.NodeID]*animationPlayback),
		paletteEffects: make(map[render.ResourceID]*gpuPaletteEffect),
	}
}

// runCompositionFrame applies budget changes and visible commands before advancing animation by the current frame time.
func (s *Service) runCompositionFrame(composer *render.Composer, backend *compositionBackend) {
	if err := s.applyTextureCacheBudget(); err != nil && s.logger != nil {
		s.logger.Error("applying texture cache budget", "error", err)
	}

	if err := composer.Drain(backend); err != nil && s.logger != nil {
		s.logger.Error("draining render composition", "error", err)
	}

	backend.advance(time.Duration(float64(time.Second) * float64(rl.GetFrameTime())))
}

// warmCompositionTextures runs after drawing so visible textures are newer in the LRU than speculative uploads.
func (s *Service) warmCompositionTextures(composer *render.Composer, backend *compositionBackend) {
	// Byte limits control total traffic; the short wall-clock limit protects frame pacing from one slow driver upload.
	err := composer.DrainWarmWithin(backend, s.textureUploadBudget.Load(), 2*time.Millisecond)
	if err != nil && s.logger != nil {
		s.logger.Error("warming texture residency", "error", err)
	}
}

// compositionBackend mirrors retained resources into native nodes and serializes mutation with animation updates.
type compositionBackend struct {
	renderer       *Service
	mu             sync.Mutex
	nodes          map[render.NodeID]*node
	resources      map[render.ResourceID]render.Resource
	nodeResources  map[render.NodeID]render.ResourceID
	playbacks      map[render.NodeID]*animationPlayback
	paletteEffects map[render.ResourceID]*gpuPaletteEffect
}

// CanWarmTexture rejects uploads that exceed the host integer range or would evict a currently resident texture.
func (b *compositionBackend) CanWarmTexture(key string, weight uint64) bool {
	if b == nil || b.renderer == nil || b.renderer.cache == nil || weight > uint64(^uint(0)>>1) {
		return false
	}

	return b.renderer.cache.CanInsertWithoutEviction(key, int(weight))
}

// close releases every backend-owned node while the graphics context is alive, acting as the shutdown safety net.
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

// closePaletteEffects unloads shaders before their lookup textures and before the native window disappears.
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

// advance updates every animation under the same lock used by composition changes, preventing frame/resource races.
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

// Apply dispatches one retained change while holding the backend lock across all related ownership mutations.
func (b *compositionBackend) Apply(change render.Change) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.ensureOwnershipMaps()

	switch change.Kind {
	case "texture-warm":
		return b.warmTexture(change.Resource)
	case "resource-create":
		return b.createResource(change.ResourceID, change.Resource)
	case "resource-update":
		return b.updateResource(change.ResourceID, change.Resource)
	case "resource-destroy":
		return b.destroyResource(change.ResourceID)
	case "create":
		return b.createNode(change.ID, change.Node)
	case "update":
		return b.updateNode(change.ID, change.Node)
	case "destroy":
		return b.destroyNode(change.ID)
	default:
		return fmt.Errorf("unknown composition change %q", change.Kind)
	}
}

// ensureOwnershipMaps preserves compatibility with zero-valued backends used by focused tests and tools.
func (b *compositionBackend) ensureOwnershipMaps() {
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
}
