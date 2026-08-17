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

// New constructs the Ebitengine desktop backend selected by -tags ebitengine.
func New(options Options) (*Bundle, error) {
	if options.Logger == nil {
		options.Logger = slog.Default()
	}
	if options.WindowWidth <= 0 || options.WindowHeight <= 0 || options.LogicalWidth <= 0 || options.LogicalHeight <= 0 {
		return nil, errors.New("ebitengine backend: window and logical dimensions must be positive")
	}
	if options.ViewportFit != "" && options.ViewportFit != "contain" && options.ViewportFit != "stretch" {
		return nil, fmt.Errorf("ebitengine backend: viewport fit must be contain or stretch, got %q", options.ViewportFit)
	}
	if options.PalettePath != "" {
		return nil, errors.New("ebitengine backend: final display-palette quantization is not implemented")
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

func (c *callbacks) run() {
	c.mu.Lock()
	items := append([]func(){}, c.items...)
	c.mu.Unlock()
	for _, callback := range items {
		callback()
	}
}

func newEbitRenderer(options Options) *ebitRenderer {
	renderer := &ebitRenderer{options: options, logger: options.Logger}
	renderer.textureUploadBudget.Store(4 << 20)
	renderer.textureCacheBudget.Store(256 << 20)
	return renderer
}

func (*ebitRenderer) Name() string { return "ebitengine" }

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

func (r *ebitRenderer) Stop(context.Context) error {
	if !r.started.Swap(false) {
		return nil
	}
	if r.backend != nil {
		r.backend.close()
	}
	return nil
}

func (r *ebitRenderer) Run(ctx context.Context) error {
	if !r.started.Load() {
		return errors.New("ebitengine backend: not started")
	}
	run := &ebitGame{renderer: r, ctx: ctx, lastUpdate: time.Now()}
	return ebiten.RunGameWithOptions(run, &ebiten.RunGameOptions{SingleThread: true})
}

func (r *ebitRenderer) SubscribeFrame(callback func()) func() { return r.frames.subscribe(callback) }

func (r *ebitRenderer) SubscribePostFrame(callback func()) func() { return r.post.subscribe(callback) }

func (r *ebitRenderer) SubscribeOverlay(callback func()) func() { return r.overlay.subscribe(callback) }

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
	frame := &image.RGBA{Pix: pixels, Stride: 4 * bounds.Dx(), Rect: image.Rect(0, 0, bounds.Dx(), bounds.Dy())}
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

func (r *ebitRenderer) WindowSize() (int, int) {
	if !r.started.Load() {
		return r.options.WindowWidth, r.options.WindowHeight
	}
	return ebiten.WindowSize()
}

func (r *ebitRenderer) Resolution() (int, int) {
	return r.options.LogicalWidth, r.options.LogicalHeight
}

func (r *ebitRenderer) ScreenToGame(x, y int) (int, int, bool) {
	windowWidth, windowHeight := r.WindowSize()
	viewport, err := calculateViewport(windowWidth, windowHeight, r.options.LogicalWidth, r.options.LogicalHeight, r.options.ViewportFit)
	if err != nil || float64(x) < viewport.x || float64(y) < viewport.y || float64(x) >= viewport.x+viewport.width || float64(y) >= viewport.y+viewport.height {
		return 0, 0, false
	}
	gameX := int((float64(x) - viewport.x) * float64(r.options.LogicalWidth) / viewport.width)
	gameY := int((float64(y) - viewport.y) * float64(r.options.LogicalHeight) / viewport.height)
	return gameX, gameY, true
}

func (r *ebitRenderer) SetWindowTitle(title string) {
	r.options.WindowTitle = title
	if r.started.Load() {
		ebiten.SetWindowTitle(title)
	}
}

func (r *ebitRenderer) SetResidencyDebug(enabled bool) { r.residencyDebug.Store(enabled) }

func (r *ebitRenderer) SetTextureUploadBudget(bytes uint64) { r.textureUploadBudget.Store(bytes) }

func (r *ebitRenderer) SetTextureCacheBudget(bytes uint64) { r.textureCacheBudget.Store(bytes) }

func (r *ebitRenderer) BackendDiagnostics() any { return r.diagnostics.snapshot() }

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

func (g *ebitGame) Layout(_, _ int) (int, int) {
	return g.renderer.Resolution()
}

func (g *ebitGame) DrawFinalScreen(screen ebiten.FinalScreen, offscreen *ebiten.Image, geoM ebiten.GeoM) {
	if g.renderer.options.ViewportFit == "stretch" {
		width, height := screen.Bounds().Dx(), screen.Bounds().Dy()
		geoM.Reset()
		geoM.Scale(float64(width)/float64(offscreen.Bounds().Dx()), float64(height)/float64(offscreen.Bounds().Dy()))
	}
	options := &ebiten.DrawImageOptions{GeoM: geoM, Filter: ebiten.FilterNearest, DisableMipmaps: true}
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

func (d *ebitDiagnosticsCounters) snapshot() EbitengineDiagnostics {
	return EbitengineDiagnostics{
		Frames: d.frames.Load(), DrawCalls: d.drawCalls.Load(), NodesVisited: d.nodesVisited.Load(), SubtreesCulled: d.subtreesCulled.Load(), TextureUpdates: d.textureUpdates.Load(),
		TextureUploads: d.textureUploads.Load(), TextureUploadBytes: d.textureUploadBytes.Load(),
		LastFrameDrawCalls: d.lastDrawCalls.Load(), LastFrameNodesVisited: d.lastNodesVisited.Load(), LastFrameSubtreesCulled: d.lastCulled.Load(), LastFrameTextureUpdates: d.lastUpdates.Load(),
		LastFrameCompositionNS: d.lastCompositionNS.Load(), LastFrameRenderNS: d.lastRenderNS.Load(), LastFrameUploadNS: d.lastUploadNS.Load(), TextureUploadNS: d.textureUploadNS.Load(),
		ActualFPS: ebiten.ActualFPS(), ActualTPS: ebiten.ActualTPS(), Audio: "muted",
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

func newEbitComposition(renderer *ebitRenderer) *ebitComposition {
	return &ebitComposition{
		renderer:  renderer,
		resources: make(map[render.ResourceID]render.Resource), textures: make(map[render.ResourceID]*ebiten.Image), nodes: make(map[render.NodeID]render.Node),
		players: make(map[render.NodeID]*render.AnimationPlayer), seekRevisions: make(map[render.NodeID]uint64), paletteVariants: make(map[string]*ebiten.Image),
	}
}

func (b *ebitComposition) close() {
	b.mu.Lock()
	defer b.mu.Unlock()
	for _, texture := range b.textures {
		texture.Dispose()
	}
	for _, texture := range b.paletteVariants {
		texture.Dispose()
	}
	clear(b.resources)
	clear(b.textures)
	clear(b.nodes)
	clear(b.players)
	clear(b.paletteVariants)
}

func (b *ebitComposition) CanWarmTexture(_ string, weight uint64) bool {
	budget := b.renderer.textureCacheBudget.Load()
	return budget == 0 || b.textureBytes+weight <= budget
}

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
			texture.Dispose()
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

func (b *ebitComposition) uploadTexture(id render.ResourceID, source image.Image, update bool) error {
	started := time.Now()
	rgba := contiguousRGBA(source)
	width, height := rgba.Bounds().Dx(), rgba.Bounds().Dy()
	texture := b.textures[id]
	if texture == nil || texture.Bounds().Dx() != width || texture.Bounds().Dy() != height {
		if texture != nil {
			b.textureBytes -= uint64(4 * texture.Bounds().Dx() * texture.Bounds().Dy())
			texture.Dispose()
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

func (b *ebitComposition) clearPaletteVariants() {
	for key, texture := range b.paletteVariants {
		texture.Dispose()
		delete(b.paletteVariants, key)
	}
}

func (b *ebitComposition) advance(delta time.Duration) {
	b.mu.Lock()
	defer b.mu.Unlock()
	for _, player := range b.players {
		player.Advance(delta)
	}
}

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
		if transformedBounds(geometry, bounds.Dx(), bounds.Dy()).Intersect(screen.Bounds()).Empty() {
			b.renderer.diagnostics.subtreesCulled.Add(1)
			continue
		}
		options := &ebiten.DrawImageOptions{GeoM: geometry, Blend: ebitBlend(node.Blend), Filter: ebiten.FilterNearest, DisableMipmaps: true}
		tint := node.Tint
		if tint.A == 0 {
			tint = color.RGBA{255, 255, 255, 255}
		}
		options.ColorScale.Scale(float32(tint.R)/255, float32(tint.G)/255, float32(tint.B)/255, 1)
		target := screen
		if clip := b.effectiveClip(node.ID, make(map[render.NodeID]bool)); clip != nil {
			rect := image.Rect(int(math.Floor(clip.X)), int(math.Floor(clip.Y)), int(math.Ceil(clip.X+clip.Width)), int(math.Ceil(clip.Y+clip.Height))).Intersect(screen.Bounds())
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
	b.renderer.diagnostics.lastNodesVisited.Store(b.renderer.diagnostics.nodesVisited.Load() - startVisited)
	b.renderer.diagnostics.lastCulled.Store(b.renderer.diagnostics.subtreesCulled.Load() - startCulled)
	b.renderer.diagnostics.lastUpdates.Store(b.renderer.diagnostics.textureUpdates.Load() - startUpdates)
}

func (b *ebitComposition) visibleThroughParents(id render.NodeID, visiting map[render.NodeID]bool) bool {
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

func (b *ebitComposition) effectiveClip(id render.NodeID, visiting map[render.NodeID]bool) *render.Rect {
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
	x2, y2 := math.Min(parent.X+parent.Width, node.Clip.X+node.Clip.Width), math.Min(parent.Y+parent.Height, node.Clip.Y+node.Clip.Height)
	return &render.Rect{X: x1, Y: y1, Width: math.Max(0, x2-x1), Height: math.Max(0, y2-y1)}
}

func transformedBounds(matrix ebiten.GeoM, width, height int) image.Rectangle {
	x0, y0 := matrix.Apply(0, 0)
	x1, y1 := matrix.Apply(float64(width), 0)
	x2, y2 := matrix.Apply(0, float64(height))
	x3, y3 := matrix.Apply(float64(width), float64(height))
	left, right := math.Min(math.Min(x0, x1), math.Min(x2, x3)), math.Max(math.Max(x0, x1), math.Max(x2, x3))
	top, bottom := math.Min(math.Min(y0, y1), math.Min(y2, y3)), math.Max(math.Max(y0, y1), math.Max(y2, y3))
	return image.Rect(int(math.Floor(left)), int(math.Floor(top)), int(math.Ceil(right)), int(math.Ceil(bottom)))
}

func (b *ebitComposition) worldTransform(id render.NodeID, memo map[render.NodeID]ebiten.GeoM, visiting map[render.NodeID]bool) ebiten.GeoM {
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
	key := fmt.Sprintf("%d:%d/%d:%d", resourceID.Slot, resourceID.Generation, node.Palette.Slot, node.Palette.Generation)
	if variant := b.paletteVariants[key]; variant != nil {
		return variant, nil
	}
	palette := b.resources[node.Palette].Payload.(color.Palette)
	quantized := quantizeImage(resource.Payload.(image.Image), palette)
	variant := ebiten.NewImageFromImage(quantized)
	b.paletteVariants[key] = variant
	return variant, nil
}

func ebitBlend(name string) ebiten.Blend {
	switch name {
	case "additive", "add-colors":
		return ebiten.Blend{BlendFactorSourceRGB: ebiten.BlendFactorOne, BlendFactorSourceAlpha: ebiten.BlendFactorOne, BlendFactorDestinationRGB: ebiten.BlendFactorOne, BlendFactorDestinationAlpha: ebiten.BlendFactorOne, BlendOperationRGB: ebiten.BlendOperationAdd, BlendOperationAlpha: ebiten.BlendOperationAdd}
	case "screen":
		return ebiten.Blend{BlendFactorSourceRGB: ebiten.BlendFactorOne, BlendFactorSourceAlpha: ebiten.BlendFactorOne, BlendFactorDestinationRGB: ebiten.BlendFactorOneMinusSourceColor, BlendFactorDestinationAlpha: ebiten.BlendFactorOneMinusSourceAlpha, BlendOperationRGB: ebiten.BlendOperationAdd, BlendOperationAlpha: ebiten.BlendOperationAdd}
	case "multiply":
		// Match Raylib's RL_BLEND_MULTIPLIED exactly: GL_DST_COLOR,
		// GL_ONE_MINUS_SRC_ALPHA. Diablo II draw mode 4 relies on the
		// destination term to preserve the light glyph interior over button art.
		return ebiten.Blend{BlendFactorSourceRGB: ebiten.BlendFactorDestinationColor, BlendFactorSourceAlpha: ebiten.BlendFactorDestinationAlpha, BlendFactorDestinationRGB: ebiten.BlendFactorOneMinusSourceAlpha, BlendFactorDestinationAlpha: ebiten.BlendFactorOneMinusSourceAlpha, BlendOperationRGB: ebiten.BlendOperationAdd, BlendOperationAlpha: ebiten.BlendOperationAdd}
	case "subtract-colors":
		return ebiten.Blend{BlendFactorSourceRGB: ebiten.BlendFactorSourceAlpha, BlendFactorSourceAlpha: ebiten.BlendFactorOne, BlendFactorDestinationRGB: ebiten.BlendFactorOne, BlendFactorDestinationAlpha: ebiten.BlendFactorOne, BlendOperationRGB: ebiten.BlendOperationReverseSubtract, BlendOperationAlpha: ebiten.BlendOperationAdd}
	default:
		return ebiten.Blend{}
	}
}

func (b *ebitComposition) cacheDiagnostics(budget uint64) ebitCacheDiagnostics {
	b.mu.Lock()
	defer b.mu.Unlock()
	return ebitCacheDiagnostics{Entries: len(b.textures) + len(b.paletteVariants), Weight: b.textureBytes, Budget: budget}
}

var _ Renderer = (*ebitRenderer)(nil)
