//go:build ebitengine

package desktop

import (
	"context"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log/slog"
	"math"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/gravestench/dark-magic/internal/audio"
	"github.com/gravestench/dark-magic/internal/presentation/render"
)

// New validates Ebitengine-specific limits before allocating renderer and input adapters; unsupported palette policy
// fails at composition time rather than after the native event loop starts.
func New(options Options) (*Bundle, error) {
	if options.Logger == nil {
		options.Logger = slog.Default()
	}

	if options.WindowWidth <= 0 || options.WindowHeight <= 0 || options.LogicalWidth <= 0 ||
		options.LogicalHeight <= 0 {
		return nil, errors.New("ebitengine backend: window and logical dimensions must be positive")
	}

	if options.ViewportFit != "" && options.ViewportFit != "contain" &&
		options.ViewportFit != "stretch" {
		return nil, fmt.Errorf(
			"ebitengine backend: viewport fit must be contain or stretch, got %q",
			options.ViewportFit,
		)
	}

	if options.PalettePath != "" {
		return nil, errors.New(
			"ebitengine backend: final display-palette quantization is not implemented",
		)
	}

	renderer := newEbitRenderer(options)

	return &Bundle{Renderer: renderer, Input: newEbitInput(renderer)}, nil
}

type ebitRenderer struct {
	options Options
	logger  *slog.Logger

	started atomic.Bool
	frames  callbacks
	post    callbacks
	overlay callbacks

	composer *render.Composer
	backend  *ebitComposition
	mixer    *audio.Mixer

	currentScreen *ebiten.Image

	textureUploadBudget atomic.Uint64
	textureCacheBudget  atomic.Uint64
	residencyDebug      atomic.Bool
	diagnostics         ebitDiagnosticsCounters
}

type callbacks struct {
	mu    sync.Mutex
	items []func()
}

// subscribe appends callbacks without removing entries during iteration; unsubscribe atomically disables its wrapper,
// so callbacks may safely unsubscribe themselves while a frame snapshot is running.
func (c *callbacks) subscribe(callback func()) func() {
	if callback == nil {
		return func() {}
	}

	var active atomic.Bool

	active.Store(true)
	c.mu.Lock()
	c.items = append(c.items, func() {
		if active.Load() {
			callback()
		}
	})
	c.mu.Unlock()

	return func() { active.Store(false) }
}

// run snapshots the callback list under lock and invokes it unlocked, preventing callbacks that subscribe more work
// from deadlocking or changing the current frame's iteration.
func (c *callbacks) run() {
	c.mu.Lock()
	items := append([]func(){}, c.items...)
	c.mu.Unlock()

	for _, callback := range items {
		callback()
	}
}

// newEbitRenderer applies the established upload and cache budgets before any composition can attach or render.
func newEbitRenderer(options Options) *ebitRenderer {
	renderer := &ebitRenderer{options: options, logger: options.Logger}
	renderer.textureUploadBudget.Store(4 << 20)
	renderer.textureCacheBudget.Store(256 << 20)

	return renderer
}

// Name identifies the selected renderer without exposing Ebitengine-native types through the desktop interface.
func (*ebitRenderer) Name() string { return "ebitengine" }

// Start configures native window policy once. Repeated starts are idempotent so lifecycle composition cannot apply
// fullscreen or cursor settings twice.
func (r *ebitRenderer) Start(context.Context) error {
	if !r.started.CompareAndSwap(false, true) {
		return nil
	}

	ebiten.SetWindowSize(r.options.WindowWidth, r.options.WindowHeight)
	ebiten.SetWindowTitle(r.options.WindowTitle)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetFullscreen(r.options.BorderlessFullscreen)

	if r.options.ShowSystemCursor {
		ebiten.SetCursorMode(ebiten.CursorModeVisible)
	} else {
		ebiten.SetCursorMode(ebiten.CursorModeHidden)
	}

	return nil
}

// Stop deallocates composition-owned native images once and leaves repeated shutdown calls as successful no-ops.
func (r *ebitRenderer) Stop(context.Context) error {
	if !r.started.Swap(false) {
		return nil
	}

	if r.backend != nil {
		r.backend.close()
	}

	return nil
}

// Run enters Ebitengine's single-thread event loop only after Start; single-thread mode keeps native presentation on
// the same owner thread as frame callbacks.
func (r *ebitRenderer) Run(ctx context.Context) error {
	if !r.started.Load() {
		return errors.New("ebitengine backend: not started")
	}

	run := &ebitGame{renderer: r, ctx: ctx, lastUpdate: time.Now()}

	return ebiten.RunGameWithOptions(run, &ebiten.RunGameOptions{SingleThread: true})
}

// SubscribeFrame registers work that must run before animation advancement for each update tick.
func (r *ebitRenderer) SubscribeFrame(callback func()) func() { return r.frames.subscribe(callback) }

// SubscribePostFrame registers work after composition and overlay drawing while the current screen is still available.
func (r *ebitRenderer) SubscribePostFrame(callback func()) func() { return r.post.subscribe(callback) }

// SubscribeOverlay registers drawing between retained composition and post-frame work, preserving overlay layering.
func (r *ebitRenderer) SubscribeOverlay(callback func()) func() { return r.overlay.subscribe(callback) }

// SubscribeViewport publishes the logical resolution on the first frame and only after later changes, avoiding
// redundant layout work for a stable surface.
func (r *ebitRenderer) SubscribeViewport(callback func(width, height int)) func() {
	lastWidth, lastHeight := -1, -1

	return r.SubscribeFrame(func() {
		width, height := r.Resolution()

		if width == lastWidth && height == lastHeight {
			return
		}

		lastWidth, lastHeight = width, height
		callback(width, height)
	})
}

// AttachComposer creates exactly one native composition adapter and drains retained changes before animation advances;
// logging drain failures keeps the frame loop alive for diagnostics.
func (r *ebitRenderer) AttachComposer(composer *render.Composer) error {
	if composer == nil {
		return errors.New("ebitengine backend: nil composition core")
	}

	if r.composer != nil {
		return errors.New("ebitengine backend: composition core is already attached")
	}

	r.composer = composer
	r.backend = newEbitComposition(r)
	r.SubscribeFrame(func() {
		started := time.Now()

		if err := composer.Drain(r.backend); err != nil {
			r.logger.Error("draining render composition", "error", err)
		}

		r.diagnostics.lastCompositionNS.Store(uint64(time.Since(started)))
	})

	return nil
}

// AttachAudio accepts one mixer and drains commands into the muted adapter every frame. Draining is still required to
// advance mixer ownership even though this backend currently emits no native sound.
func (r *ebitRenderer) AttachAudio(mixer *audio.Mixer) error {
	if mixer == nil {
		return errors.New("ebitengine backend: nil audio mixer")
	}

	if r.mixer != nil {
		return errors.New("ebitengine backend: audio mixer is already attached")
	}

	r.mixer = mixer
	r.SubscribeFrame(func() {
		if err := mixer.Drain(discardAudioBackend{}); err != nil {
			r.logger.Error("draining muted Ebitengine audio", "error", err)
		}
	})

	return nil
}

// CaptureScreenshot copies the active frame before encoding it, so PNG work never retains Ebitengine's transient screen
// image beyond Draw. Requests outside a draw callback fail instead of reading stale pixels.
func (r *ebitRenderer) CaptureScreenshot(name string) error {
	if name == "" {
		return errors.New("ebitengine backend: screenshot path is required")
	}

	if r.currentScreen == nil {
		return errors.New("ebitengine backend: screenshot requested before a frame was drawn")
	}

	bounds := r.currentScreen.Bounds()
	pixels := make([]byte, 4*bounds.Dx()*bounds.Dy())
	r.currentScreen.ReadPixels(pixels)
	frame := &image.RGBA{
		Pix:    pixels,
		Stride: 4 * bounds.Dx(),
		Rect:   image.Rect(0, 0, bounds.Dx(), bounds.Dy()),
	}

	if err := os.MkdirAll(filepath.Dir(name), 0o755); err != nil {
		return err
	}

	file, err := os.Create(name)
	if err != nil {
		return err
	}

	if err := png.Encode(file, frame); err != nil {
		_ = file.Close()
		return err
	}

	return file.Close()
}

// WindowSize returns configured dimensions before native startup and delegates to the resizable window afterward.
func (r *ebitRenderer) WindowSize() (int, int) {
	if !r.started.Load() {
		return r.options.WindowWidth, r.options.WindowHeight
	}

	return ebiten.WindowSize()
}

// Resolution returns the fixed logical surface independently of native window resizing and letterboxing.
func (r *ebitRenderer) Resolution() (int, int) {
	return r.options.LogicalWidth, r.options.LogicalHeight
}

// ScreenToGame rejects letterbox and out-of-window coordinates before mapping accepted native pixels into the logical
// surface, preventing edge clicks from becoming game actions.
func (r *ebitRenderer) ScreenToGame(x, y int) (int, int, bool) {
	windowWidth, windowHeight := r.WindowSize()

	viewport, err := calculateViewport(
		windowWidth,
		windowHeight,
		r.options.LogicalWidth,
		r.options.LogicalHeight,
		r.options.ViewportFit,
	)
	if err != nil || float64(x) < viewport.x || float64(y) < viewport.y ||
		float64(x) >= viewport.x+viewport.width ||
		float64(y) >= viewport.y+viewport.height {
		return 0, 0, false
	}

	gameX := int((float64(x) - viewport.x) * float64(r.options.LogicalWidth) / viewport.width)
	gameY := int((float64(y) - viewport.y) * float64(r.options.LogicalHeight) / viewport.height)

	return gameX, gameY, true
}

// SetWindowTitle stores pre-start changes in options and applies post-start changes immediately to the native window.
func (r *ebitRenderer) SetWindowTitle(title string) {
	r.options.WindowTitle = title
	if r.started.Load() {
		ebiten.SetWindowTitle(title)
	}
}

// SetResidencyDebug records the backend-neutral diagnostic control atomically; this backend currently retains the
// setting for interface parity but does not draw a residency overlay.
func (r *ebitRenderer) SetResidencyDebug(enabled bool) { r.residencyDebug.Store(enabled) }

// SetTextureUploadBudget records the backend-neutral upload control atomically; Ebitengine currently uploads drained
// changes immediately rather than enforcing this stored per-frame budget.
func (r *ebitRenderer) SetTextureUploadBudget(bytes uint64) { r.textureUploadBudget.Store(bytes) }

// SetTextureCacheBudget atomically updates the admission limit used by future texture warming decisions.
func (r *ebitRenderer) SetTextureCacheBudget(bytes uint64) { r.textureCacheBudget.Store(bytes) }

// BackendDiagnostics snapshots atomic renderer counters without exposing their mutable storage.
func (r *ebitRenderer) BackendDiagnostics() any { return r.diagnostics.snapshot() }

// CacheDiagnostics reports the configured budget even before a composition attaches, then includes live cache state.
func (r *ebitRenderer) CacheDiagnostics() any {
	if r.backend == nil {
		return ebitCacheDiagnostics{Budget: r.textureCacheBudget.Load()}
	}

	return r.backend.cacheDiagnostics(r.textureCacheBudget.Load())
}

type ebitGame struct {
	renderer   *ebitRenderer
	ctx        context.Context
	lastUpdate time.Time
}

// Update terminates the native loop when the caller's context ends, then runs frame subscribers before advancing
// retained animations by the elapsed wall-clock delta.
func (g *ebitGame) Update() error {
	select {
	case <-g.ctx.Done():
		return ebiten.Termination
	default:
	}

	now := time.Now()
	delta := now.Sub(g.lastUpdate)
	g.lastUpdate = now
	g.renderer.frames.run()

	if g.renderer.backend != nil {
		g.renderer.backend.advance(delta)
	}

	return nil
}

// Draw establishes the transient screenshot surface, then renders the black background, retained composition, overlay,
// and post-frame callbacks in their externally visible order.
func (g *ebitGame) Draw(screen *ebiten.Image) {
	r := g.renderer
	r.currentScreen = screen
	screen.Fill(color.Black)

	started := time.Now()

	if r.backend != nil {
		r.backend.draw(screen)
	}

	r.overlay.run()
	r.diagnostics.lastRenderNS.Store(uint64(time.Since(started)))
	r.diagnostics.frames.Add(1)

	r.post.run()
	r.currentScreen = nil
}

// Layout keeps Ebitengine's offscreen surface at the configured logical resolution regardless of native window size.
func (g *ebitGame) Layout(_, _ int) (int, int) {
	return g.renderer.Resolution()
}

// DrawFinalScreen applies stretch only when requested; contain uses Ebitengine's supplied letterbox transform. Nearest
// filtering and disabled mipmaps preserve pixel-art sampling in both modes.
func (g *ebitGame) DrawFinalScreen(
	screen ebiten.FinalScreen,
	offscreen *ebiten.Image,
	geoM ebiten.GeoM,
) {
	if g.renderer.options.ViewportFit == "stretch" {
		width, height := screen.Bounds().Dx(), screen.Bounds().Dy()

		geoM.Reset()
		geoM.Scale(
			float64(width)/float64(offscreen.Bounds().Dx()),
			float64(height)/float64(offscreen.Bounds().Dy()),
		)
	}

	options := &ebiten.DrawImageOptions{
		GeoM:           geoM,
		Filter:         ebiten.FilterNearest,
		DisableMipmaps: true,
	}
	screen.DrawImage(offscreen, options)
}

type ebitDiagnosticsCounters struct {
	frames             atomic.Uint64
	drawCalls          atomic.Uint64
	nodesVisited       atomic.Uint64
	subtreesCulled     atomic.Uint64
	textureUpdates     atomic.Uint64
	textureUploads     atomic.Uint64
	textureUploadBytes atomic.Uint64
	lastDrawCalls      atomic.Uint64
	lastNodesVisited   atomic.Uint64
	lastCulled         atomic.Uint64
	lastUpdates        atomic.Uint64
	lastCompositionNS  atomic.Uint64
	lastRenderNS       atomic.Uint64
	lastUploadNS       atomic.Uint64
	textureUploadNS    atomic.Uint64
}

// EbitengineDiagnostics is a detached snapshot of cumulative and most-recent-frame renderer work.
type EbitengineDiagnostics struct {
	Frames, DrawCalls, NodesVisited, SubtreesCulled, TextureUpdates uint64
	TextureUploads, TextureUploadBytes                              uint64
	LastFrameDrawCalls, LastFrameNodesVisited                       uint64
	LastFrameSubtreesCulled, LastFrameTextureUpdates                uint64
	LastFrameCompositionNS, LastFrameRenderNS, LastFrameUploadNS    uint64
	TextureUploadNS                                                 uint64
	ActualFPS, ActualTPS                                            float64
	Audio                                                           string
}

// snapshot loads every atomic counter independently and samples Ebitengine rates at call time, producing a detached
// diagnostics value suitable for logging without sharing counter ownership.
func (d *ebitDiagnosticsCounters) snapshot() EbitengineDiagnostics {
	return EbitengineDiagnostics{
		Frames:                  d.frames.Load(),
		DrawCalls:               d.drawCalls.Load(),
		NodesVisited:            d.nodesVisited.Load(),
		SubtreesCulled:          d.subtreesCulled.Load(),
		TextureUpdates:          d.textureUpdates.Load(),
		TextureUploads:          d.textureUploads.Load(),
		TextureUploadBytes:      d.textureUploadBytes.Load(),
		LastFrameDrawCalls:      d.lastDrawCalls.Load(),
		LastFrameNodesVisited:   d.lastNodesVisited.Load(),
		LastFrameSubtreesCulled: d.lastCulled.Load(),
		LastFrameTextureUpdates: d.lastUpdates.Load(),
		LastFrameCompositionNS:  d.lastCompositionNS.Load(),
		LastFrameRenderNS:       d.lastRenderNS.Load(),
		LastFrameUploadNS:       d.lastUploadNS.Load(),
		TextureUploadNS:         d.textureUploadNS.Load(),
		ActualFPS:               ebiten.ActualFPS(),
		ActualTPS:               ebiten.ActualTPS(),
		Audio:                   "muted",
	}
}

type ebitCacheDiagnostics struct {
	Entries int
	Weight  uint64
	Budget  uint64
}

type ebitComposition struct {
	renderer        *ebitRenderer
	mu              sync.Mutex
	resources       map[render.ResourceID]render.Resource
	textures        map[render.ResourceID]*ebiten.Image
	nodes           map[render.NodeID]render.Node
	players         map[render.NodeID]*render.AnimationPlayer
	seekRevisions   map[render.NodeID]uint64
	paletteVariants map[string]*ebiten.Image
	textureBytes    uint64
}

// newEbitComposition initializes every retained and native lookup before the composer can drain changes into it.
func newEbitComposition(renderer *ebitRenderer) *ebitComposition {
	return &ebitComposition{
		renderer:        renderer,
		resources:       make(map[render.ResourceID]render.Resource),
		textures:        make(map[render.ResourceID]*ebiten.Image),
		nodes:           make(map[render.NodeID]render.Node),
		players:         make(map[render.NodeID]*render.AnimationPlayer),
		seekRevisions:   make(map[render.NodeID]uint64),
		paletteVariants: make(map[string]*ebiten.Image),
	}
}

// close deallocates base textures and derived palette variants while holding composition ownership, then clears
// resource, texture, node, player, and variant maps so no later draw can reach released native state.
func (b *ebitComposition) close() {
	b.mu.Lock()
	defer b.mu.Unlock()

	for _, texture := range b.textures {
		texture.Deallocate()
	}

	for _, texture := range b.paletteVariants {
		texture.Deallocate()
	}

	clear(b.resources)
	clear(b.textures)
	clear(b.nodes)
	clear(b.players)
	clear(b.paletteVariants)
}

// CanWarmTexture admits a speculative texture only when its weight fits the current cache budget; a zero budget remains
// the established unlimited setting.
func (b *ebitComposition) CanWarmTexture(_ string, weight uint64) bool {
	budget := b.renderer.textureCacheBudget.Load()

	return budget == 0 || b.textureBytes+weight <= budget
}

// Apply serializes retained change ordering with native resource maps. It validates references before publication and
// invalidates palette variants whenever their source texture may have changed.
func (b *ebitComposition) Apply(change render.Change) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	switch change.Kind {
	case "texture-warm":
		return nil
	case "resource-create":
		b.resources[change.ResourceID] = change.Resource
		switch change.Resource.Kind {
		case render.ResourceTexture:
			return b.uploadTexture(change.ResourceID, change.Resource.Payload.(image.Image), false)
		case render.ResourceRenderTarget:
			target := change.Resource.Payload.(render.RenderTargetData)
			texture := ebiten.NewImage(target.Width, target.Height)

			b.textures[change.ResourceID] = texture
			b.textureBytes += uint64(4 * target.Width * target.Height)
		}

		return nil
	case "resource-update":
		resource, ok := b.resources[change.ResourceID]

		if !ok || resource.Kind != render.ResourceTexture {
			return fmt.Errorf("resource %v is not an updateable texture", change.ResourceID)
		}

		b.resources[change.ResourceID] = change.Resource
		b.clearPaletteVariants()

		return b.uploadTexture(change.ResourceID, change.Resource.Payload.(image.Image), true)
	case "resource-destroy":
		if _, ok := b.resources[change.ResourceID]; !ok {
			return fmt.Errorf("resource %v does not exist", change.ResourceID)
		}

		if texture := b.textures[change.ResourceID]; texture != nil {
			b.textureBytes -= uint64(4 * texture.Bounds().Dx() * texture.Bounds().Dy())
			texture.Deallocate()
		}

		delete(b.textures, change.ResourceID)
		delete(b.resources, change.ResourceID)
		b.clearPaletteVariants()

		return nil
	case "create":
		if _, ok := b.nodes[change.ID]; ok {
			return fmt.Errorf("node %v already exists", change.ID)
		}

		return b.applyNode(change.Node)
	case "update":
		if _, ok := b.nodes[change.ID]; !ok {
			return fmt.Errorf("node %v does not exist", change.ID)
		}

		return b.applyNode(change.Node)
	case "destroy":
		if _, ok := b.nodes[change.ID]; !ok {
			return fmt.Errorf("node %v does not exist", change.ID)
		}

		delete(b.nodes, change.ID)
		delete(b.players, change.ID)
		delete(b.seekRevisions, change.ID)

		return nil
	default:
		return fmt.Errorf("unknown composition change %q", change.Kind)
	}
}

// applyNode validates parent, resource, palette, and blend references before replacing visible node state. Animation
// players survive ordinary updates and seek only when the command revision changes.
func (b *ebitComposition) applyNode(node render.Node) error {
	if node.Parent != (render.NodeID{}) {
		if _, ok := b.nodes[node.Parent]; !ok {
			return fmt.Errorf("parent node %v does not exist", node.Parent)
		}
	}

	if node.Resource != (render.ResourceID{}) {
		resource, ok := b.resources[node.Resource]

		if !ok {
			return fmt.Errorf("resource %v does not exist", node.Resource)
		}

		if resource.Kind == render.ResourceAnimation {
			if _, ok := b.players[node.ID]; !ok || b.nodes[node.ID].Resource != node.Resource {
				animation := resource.Payload.(render.AnimationData)
				b.players[node.ID] = render.NewAnimationPlayer(animation.Durations, animation.Loop)
			}
		}
	}

	if node.Palette != (render.ResourceID{}) {
		resource, ok := b.resources[node.Palette]
		if !ok || resource.Kind != render.ResourcePalette {
			return fmt.Errorf("palette resource %v is unavailable", node.Palette)
		}
	}

	switch node.Blend {
	case "", "alpha", "additive", "screen", "multiply", "add-colors", "subtract-colors":
	default:
		return fmt.Errorf("unsupported blend mode %q", node.Blend)
	}

	if player := b.players[node.ID]; player != nil {
		player.SetPaused(node.AnimationPaused)

		if b.seekRevisions[node.ID] != node.AnimationSeekRevision {
			player.Seek(node.AnimationSeek)
			b.seekRevisions[node.ID] = node.AnimationSeekRevision
		}
	}

	b.nodes[node.ID] = node

	return nil
}

// uploadTexture reuses native images with matching dimensions and reallocates only on size changes. Upload timing and
// byte counters are recorded after the write so diagnostics describe completed owner-thread work.
func (b *ebitComposition) uploadTexture(
	id render.ResourceID,
	source image.Image,
	update bool,
) error {
	started := time.Now()
	rgba := contiguousRGBA(source)
	width, height := rgba.Bounds().Dx(), rgba.Bounds().Dy()

	texture := b.textures[id]

	if texture == nil || texture.Bounds().Dx() != width || texture.Bounds().Dy() != height {
		if texture != nil {
			b.textureBytes -= uint64(4 * texture.Bounds().Dx() * texture.Bounds().Dy())
			texture.Deallocate()
		}

		texture = ebiten.NewImage(width, height)
		b.textures[id] = texture
		b.textureBytes += uint64(len(rgba.Pix))
	}

	texture.WritePixels(rgba.Pix)

	elapsed := uint64(time.Since(started))

	b.renderer.diagnostics.textureUploads.Add(1)
	b.renderer.diagnostics.textureUploadBytes.Add(uint64(len(rgba.Pix)))
	b.renderer.diagnostics.textureUploadNS.Add(elapsed)
	b.renderer.diagnostics.lastUploadNS.Store(elapsed)

	if update {
		b.renderer.diagnostics.textureUpdates.Add(1)
	}

	return nil
}

// clearPaletteVariants deallocates every derived native image before deleting its cache entry, preventing stale
// palette results from surviving a source update or destruction.
func (b *ebitComposition) clearPaletteVariants() {
	for key, texture := range b.paletteVariants {
		texture.Deallocate()
		delete(b.paletteVariants, key)
	}
}

// advance moves every animation player by the same frame delta while excluding concurrent composition changes.
func (b *ebitComposition) advance(delta time.Duration) {
	b.mu.Lock()
	defer b.mu.Unlock()

	for _, player := range b.players {
		player.Advance(delta)
	}
}

// draw deterministically orders nodes by layer, Z, and slot, then applies hierarchy visibility, transforms, culling,
// clipping, tint, and blend policy while collecting per-frame diagnostics.
func (b *ebitComposition) draw(screen *ebiten.Image) {
	b.mu.Lock()
	defer b.mu.Unlock()

	startDraws := b.renderer.diagnostics.drawCalls.Load()
	startVisited := b.renderer.diagnostics.nodesVisited.Load()
	startCulled := b.renderer.diagnostics.subtreesCulled.Load()
	startUpdates := b.renderer.diagnostics.textureUpdates.Load()

	nodes := make([]render.Node, 0, len(b.nodes))
	for _, node := range b.nodes {
		nodes = append(nodes, node)
	}

	// Map iteration is nondeterministic, so rendering uses semantic layer/Z order
	// with the generation-checked slot as the stable final tie-breaker.
	sort.SliceStable(nodes, func(i, j int) bool {
		if nodes[i].Layer != nodes[j].Layer {
			return nodes[i].Layer < nodes[j].Layer
		}

		if nodes[i].Z != nodes[j].Z {
			return nodes[i].Z < nodes[j].Z
		}

		return nodes[i].ID.Slot < nodes[j].ID.Slot
	})

	transforms := make(map[render.NodeID]ebiten.GeoM, len(nodes))

	for _, node := range nodes {
		b.renderer.diagnostics.nodesVisited.Add(1)

		if !b.visibleThroughParents(node.ID, make(map[render.NodeID]bool)) {
			b.renderer.diagnostics.subtreesCulled.Add(1)
			continue
		}

		if node.Resource == (render.ResourceID{}) {
			continue
		}

		texture, err := b.nodeTexture(node)
		if err != nil || texture == nil {
			continue
		}

		bounds := texture.Bounds()

		var geometry ebiten.GeoM

		geometry.Translate(-node.OriginX*float64(bounds.Dx()), -node.OriginY*float64(bounds.Dy()))
		geometry.Concat(b.worldTransform(node.ID, transforms, make(map[render.NodeID]bool)))

		// Cull against transformed geometry before allocating draw options or clip
		// subimages; this is an optimization only and does not change node order.
		if transformedBounds(
			geometry,
			bounds.Dx(),
			bounds.Dy(),
		).Intersect(screen.Bounds()).
			Empty() {
			b.renderer.diagnostics.subtreesCulled.Add(1)
			continue
		}

		options := &ebiten.DrawImageOptions{
			GeoM:           geometry,
			Blend:          ebitBlend(node.Blend),
			Filter:         ebiten.FilterNearest,
			DisableMipmaps: true,
		}

		tint := node.Tint

		if tint.A == 0 {
			tint = color.RGBA{255, 255, 255, 255}
		}

		options.ColorScale.Scale(float32(tint.R)/255, float32(tint.G)/255, float32(tint.B)/255, 1)

		target := screen

		if clip := b.effectiveClip(node.ID, make(map[render.NodeID]bool)); clip != nil {
			rect := image.Rect(
				int(math.Floor(clip.X)),
				int(math.Floor(clip.Y)),
				int(math.Ceil(clip.X+clip.Width)),
				int(math.Ceil(clip.Y+clip.Height)),
			).
				Intersect(screen.Bounds())
			if rect.Empty() {
				b.renderer.diagnostics.subtreesCulled.Add(1)
				continue
			}

			target = screen.SubImage(rect).(*ebiten.Image)
		}

		target.DrawImage(texture, options)
		b.renderer.diagnostics.drawCalls.Add(1)
	}

	b.renderer.diagnostics.lastDrawCalls.Store(b.renderer.diagnostics.drawCalls.Load() - startDraws)
	b.renderer.diagnostics.lastNodesVisited.Store(
		b.renderer.diagnostics.nodesVisited.Load() - startVisited,
	)
	b.renderer.diagnostics.lastCulled.Store(
		b.renderer.diagnostics.subtreesCulled.Load() - startCulled,
	)
	b.renderer.diagnostics.lastUpdates.Store(
		b.renderer.diagnostics.textureUpdates.Load() - startUpdates,
	)
}

// visibleThroughParents requires every ancestor to exist and be visible; cycle detection fails closed so malformed
// retained graphs cannot recurse indefinitely or render a partially defined subtree.
func (b *ebitComposition) visibleThroughParents(
	id render.NodeID,
	visiting map[render.NodeID]bool,
) bool {
	if visiting[id] {
		return false
	}

	node, ok := b.nodes[id]
	if !ok || !node.Visible {
		return false
	}

	if node.Parent == (render.NodeID{}) {
		return true
	}

	visiting[id] = true
	visible := b.visibleThroughParents(node.Parent, visiting)
	delete(visiting, id)

	return visible
}

// effectiveClip intersects a node's clip with all ancestor clips. It copies a lone clip before returning so callers
// never receive a pointer into retained node storage.
func (b *ebitComposition) effectiveClip(
	id render.NodeID,
	visiting map[render.NodeID]bool,
) *render.Rect {
	if visiting[id] {
		return nil
	}

	node, ok := b.nodes[id]
	if !ok {
		return nil
	}

	var parent *render.Rect

	if node.Parent != (render.NodeID{}) {
		visiting[id] = true
		parent = b.effectiveClip(node.Parent, visiting)
		delete(visiting, id)
	}

	if node.Clip == nil {
		return parent
	}

	if parent == nil {
		copy := *node.Clip

		return &copy
	}

	x1, y1 := math.Max(parent.X, node.Clip.X), math.Max(parent.Y, node.Clip.Y)
	x2, y2 := math.Min(
		parent.X+parent.Width,
		node.Clip.X+node.Clip.Width,
	), math.Min(
		parent.Y+parent.Height,
		node.Clip.Y+node.Clip.Height,
	)

	return &render.Rect{X: x1, Y: y1, Width: math.Max(0, x2-x1), Height: math.Max(0, y2-y1)}
}

// transformedBounds converts all four transformed corners into a conservative integer rectangle for screen culling.
func transformedBounds(matrix ebiten.GeoM, width, height int) image.Rectangle {
	x0, y0 := matrix.Apply(0, 0)
	x1, y1 := matrix.Apply(float64(width), 0)
	x2, y2 := matrix.Apply(0, float64(height))
	x3, y3 := matrix.Apply(float64(width), float64(height))
	left, right := math.Min(
		math.Min(x0, x1),
		math.Min(x2, x3),
	), math.Max(
		math.Max(x0, x1),
		math.Max(x2, x3),
	)
	top, bottom := math.Min(
		math.Min(y0, y1),
		math.Min(y2, y3),
	), math.Max(
		math.Max(y0, y1),
		math.Max(y2, y3),
	)

	return image.Rect(
		int(math.Floor(left)),
		int(math.Floor(top)),
		int(math.Ceil(right)),
		int(math.Ceil(bottom)),
	)
}

// worldTransform composes local transforms through parents and memoizes completed matrices for sibling traversal.
// Missing nodes and cycles fail closed with the zero transform instead of recursing indefinitely.
func (b *ebitComposition) worldTransform(
	id render.NodeID,
	memo map[render.NodeID]ebiten.GeoM,
	visiting map[render.NodeID]bool,
) ebiten.GeoM {
	if matrix, ok := memo[id]; ok {
		return matrix
	}

	node, ok := b.nodes[id]
	if !ok || visiting[id] {
		return ebiten.GeoM{}
	}

	visiting[id] = true

	var matrix ebiten.GeoM

	matrix.Scale(node.ScaleX, node.ScaleY)
	matrix.Rotate(node.Rotation * math.Pi / 180)
	matrix.Translate(node.X, node.Y)

	if node.Parent != (render.NodeID{}) {
		parent := b.worldTransform(node.Parent, memo, visiting)
		matrix.Concat(parent)
	}

	delete(visiting, id)
	memo[id] = matrix

	return matrix
}

// nodeTexture resolves the current animation frame and lazily caches palette-derived native images by both resource
// generations, preventing stale cache hits after slot reuse.
func (b *ebitComposition) nodeTexture(node render.Node) (*ebiten.Image, error) {
	resource, ok := b.resources[node.Resource]
	if !ok {
		return nil, fmt.Errorf("resource %v unavailable", node.Resource)
	}

	resourceID := node.Resource

	if resource.Kind == render.ResourceAnimation {
		animation := resource.Payload.(render.AnimationData)

		frame := 0

		if player := b.players[node.ID]; player != nil {
			frame = player.Frame()
		}

		resourceID = animation.Frames[frame]
		resource = b.resources[resourceID]
	}

	texture := b.textures[resourceID]

	if node.Palette == (render.ResourceID{}) || texture == nil {
		return texture, nil
	}

	key := fmt.Sprintf(
		"%d:%d/%d:%d",
		resourceID.Slot,
		resourceID.Generation,
		node.Palette.Slot,
		node.Palette.Generation,
	)

	if variant := b.paletteVariants[key]; variant != nil {
		return variant, nil
	}

	palette := b.resources[node.Palette].Payload.(color.Palette)
	quantized := quantizeImage(resource.Payload.(image.Image), palette)
	variant := ebiten.NewImageFromImage(quantized)

	b.paletteVariants[key] = variant

	return variant, nil
}

// ebitBlend maps backend-neutral blend names to Ebitengine factors; the zero blend retains ordinary alpha behavior.
func ebitBlend(name string) ebiten.Blend {
	switch name {
	case "additive", "add-colors":
		return ebiten.Blend{
			BlendFactorSourceRGB:        ebiten.BlendFactorOne,
			BlendFactorSourceAlpha:      ebiten.BlendFactorOne,
			BlendFactorDestinationRGB:   ebiten.BlendFactorOne,
			BlendFactorDestinationAlpha: ebiten.BlendFactorOne,
			BlendOperationRGB:           ebiten.BlendOperationAdd,
			BlendOperationAlpha:         ebiten.BlendOperationAdd,
		}
	case "screen":
		return ebiten.Blend{
			BlendFactorSourceRGB:        ebiten.BlendFactorOne,
			BlendFactorSourceAlpha:      ebiten.BlendFactorOne,
			BlendFactorDestinationRGB:   ebiten.BlendFactorOneMinusSourceColor,
			BlendFactorDestinationAlpha: ebiten.BlendFactorOneMinusSourceAlpha,
			BlendOperationRGB:           ebiten.BlendOperationAdd,
			BlendOperationAlpha:         ebiten.BlendOperationAdd,
		}
	case "multiply":
		// Match Raylib's RL_BLEND_MULTIPLIED exactly: GL_DST_COLOR,
		// GL_ONE_MINUS_SRC_ALPHA. Diablo II draw mode 4 relies on the
		// destination term to preserve the light glyph interior over button art.
		return ebiten.Blend{
			BlendFactorSourceRGB:        ebiten.BlendFactorDestinationColor,
			BlendFactorSourceAlpha:      ebiten.BlendFactorDestinationAlpha,
			BlendFactorDestinationRGB:   ebiten.BlendFactorOneMinusSourceAlpha,
			BlendFactorDestinationAlpha: ebiten.BlendFactorOneMinusSourceAlpha,
			BlendOperationRGB:           ebiten.BlendOperationAdd,
			BlendOperationAlpha:         ebiten.BlendOperationAdd,
		}
	case "subtract-colors":
		return ebiten.Blend{
			BlendFactorSourceRGB:        ebiten.BlendFactorSourceAlpha,
			BlendFactorSourceAlpha:      ebiten.BlendFactorOne,
			BlendFactorDestinationRGB:   ebiten.BlendFactorOne,
			BlendFactorDestinationAlpha: ebiten.BlendFactorOne,
			BlendOperationRGB:           ebiten.BlendOperationReverseSubtract,
			BlendOperationAlpha:         ebiten.BlendOperationAdd,
		}
	default:
		return ebiten.Blend{}
	}
}

// cacheDiagnostics snapshots native base and palette-variant counts under the composition lock while pairing them with
// the caller's atomically loaded budget.
func (b *ebitComposition) cacheDiagnostics(budget uint64) ebitCacheDiagnostics {
	b.mu.Lock()
	defer b.mu.Unlock()

	return ebitCacheDiagnostics{
		Entries: len(b.textures) + len(b.paletteVariants),
		Weight:  b.textureBytes,
		Budget:  budget,
	}
}

var _ Renderer = (*ebitRenderer)(nil)
