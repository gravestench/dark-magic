package raylibRenderer

import (
	"log/slog"
	"sync"
	"sync/atomic"

	rl "github.com/gen2brain/raylib-go/raylib"

	"github.com/gravestench/dark-magic/internal/cache"
	"github.com/gravestench/dark-magic/internal/presentation/render"
)

// Service owns the Raylib window, GPU resources, audio device, and native owner
// thread. Other packages submit backend-neutral state or subscribe callbacks;
// they never receive Raylib handles or call native APIs directly.
type Service struct {
	logger *slog.Logger

	config *Config
	cache  *cache.Cache

	camera   rl.Camera2D
	rootNode *node

	isInit             atomic.Bool
	frameMux           sync.Mutex
	frameCallbacks     []func()
	frameSnapshot      atomic.Value
	postFrameMux       sync.Mutex
	postFrameCallbacks []func()
	postFrameSnapshot  atomic.Value
	overlayMux         sync.Mutex
	overlayCallbacks   []func()
	overlaySnapshot    atomic.Value

	compositionMu      sync.Mutex
	composition        *render.Composer
	compositionBackend *compositionBackend

	audioMu      sync.Mutex
	audioBackend *raylibAudioBackend

	paletteQuantizer        *paletteQuantizer
	gameTarget              rl.RenderTexture2D
	frames                  atomic.Uint64
	drawCalls               atomic.Uint64
	nodesVisited            atomic.Uint64
	subtreesCulled          atomic.Uint64
	textureUpdates          atomic.Uint64
	textureUploads          atomic.Uint64
	textureUploadBytes      atomic.Uint64
	textureUploadBudget     atomic.Uint64
	textureCacheBudget      atomic.Uint64
	residencyDebug          atomic.Bool
	lastFrameDrawCalls      atomic.Uint64
	lastFrameNodesVisited   atomic.Uint64
	lastFrameSubtreesCulled atomic.Uint64
	lastFrameTextureUpdates atomic.Uint64
	lastFrameCompositionNS  atomic.Uint64
	lastFrameRenderNS       atomic.Uint64
	lastFrameUploadNS       atomic.Uint64
	textureUploadNS         atomic.Uint64
}

// BackendDiagnostics reports native-adapter work without exposing Raylib
// handles. Values are cumulative for the renderer lifetime.
type BackendDiagnostics struct {
	Frames, DrawCalls, NodesVisited, SubtreesCulled, TextureUpdates uint64
	LastFrameDrawCalls, LastFrameNodesVisited                       uint64
	LastFrameSubtreesCulled, LastFrameTextureUpdates                uint64
	LastFrameCompositionNS, LastFrameRenderNS, LastFrameUploadNS    uint64
	TextureUploadNS                                                 uint64
}

// BackendDiagnostics returns a lock-free snapshot suitable for overlays and profiling. Atomics let diagnostics sample
// the owner-thread renderer without stalling frame production.
func (s *Service) BackendDiagnostics() BackendDiagnostics {
	return BackendDiagnostics{
		Frames: s.frames.Load(), DrawCalls: s.drawCalls.Load(),
		NodesVisited: s.nodesVisited.Load(), SubtreesCulled: s.subtreesCulled.Load(),
		TextureUpdates:          s.textureUpdates.Load(),
		LastFrameDrawCalls:      s.lastFrameDrawCalls.Load(),
		LastFrameNodesVisited:   s.lastFrameNodesVisited.Load(),
		LastFrameSubtreesCulled: s.lastFrameSubtreesCulled.Load(),
		LastFrameTextureUpdates: s.lastFrameTextureUpdates.Load(),
		LastFrameCompositionNS:  s.lastFrameCompositionNS.Load(),
		LastFrameRenderNS:       s.lastFrameRenderNS.Load(),
		LastFrameUploadNS:       s.lastFrameUploadNS.Load(),
		TextureUploadNS:         s.textureUploadNS.Load(),
	}
}

// SubscribeOverlay registers owner-thread drawing after scene composition and
// display quantization but before the back buffer is presented.
func (s *Service) SubscribeOverlay(callback func()) func() {
	return subscribeCallback(callback, &s.overlayMux, &s.overlayCallbacks, &s.overlaySnapshot)
}

// runOverlays invokes the immutable callback snapshot on the renderer owner thread. Cancelled wrappers remain harmless
// in the snapshot, so frame rendering never contends with subscription mutation.
func (s *Service) runOverlays() {
	runCallbackSnapshot(s.overlaySnapshot.Load())
}

// SubscribePostFrame registers owner-thread work after the fully composed frame
// has been presented. Screenshot and visual-inspection tools belong here.
func (s *Service) SubscribePostFrame(callback func()) func() {
	return subscribeCallback(callback, &s.postFrameMux, &s.postFrameCallbacks, &s.postFrameSnapshot)
}

// runPostFrame invokes tools only after the presented frame is complete, ensuring captures never observe a partially
// composed back buffer.
func (s *Service) runPostFrame() {
	runCallbackSnapshot(s.postFrameSnapshot.Load())
}

// OnFrame registers work that must run on the renderer thread, immediately
// before scene graph updates. Raylib window and input calls belong here.
func (s *Service) OnFrame(callback func()) {
	_ = s.SubscribeFrame(callback)
}

// SubscribeFrame registers renderer-thread work and returns a safe idempotent
// cancellation function.
func (s *Service) SubscribeFrame(callback func()) func() {
	return subscribeCallback(callback, &s.frameMux, &s.frameCallbacks, &s.frameSnapshot)
}

// subscribeCallback appends an immutable wrapper snapshot while making cancellation lock-free and idempotent. Keeping
// old wrappers avoids modifying a callback slice concurrently with owner-thread iteration.
func subscribeCallback(
	callback func(),
	mux *sync.Mutex,
	callbacks *[]func(),
	snapshot *atomic.Value,
) func() {
	if callback == nil {
		return func() {}
	}

	var active atomic.Bool
	active.Store(true)

	wrapper := func() {
		if active.Load() {
			callback()
		}
	}

	mux.Lock()

	*callbacks = append(*callbacks, wrapper)
	snapshot.Store(append([]func(){}, (*callbacks)...))
	mux.Unlock()

	return func() { active.Store(false) }
}

// runCallbackSnapshot invokes one already-published callback list. Nil represents a channel with no subscriptions yet.
func runCallbackSnapshot(value any) {
	if value == nil {
		return
	}

	for _, callback := range value.([]func()) {
		callback()
	}
}

// SetLogger installs the renderer's structured diagnostic sink. Start supplies slog.Default only when this remains nil.
func (s *Service) SetLogger(logger *slog.Logger) {
	s.logger = logger
}

// Logger returns the currently configured logger without creating a fallback before lifecycle initialization.
func (s *Service) Logger() *slog.Logger {
	return s.logger
}
