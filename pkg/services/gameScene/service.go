package gameScene

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/gravestench/servicemesh"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"

	"github.com/gravestench/dark-magic/pkg/assetinspect"
	"github.com/gravestench/dark-magic/pkg/scene"
	"github.com/gravestench/dark-magic/pkg/services/common"
	"github.com/gravestench/dark-magic/pkg/services/fileLoader"
	"github.com/gravestench/dark-magic/pkg/services/input"
	"github.com/gravestench/dark-magic/pkg/services/locale"
	"github.com/gravestench/dark-magic/pkg/services/raylibRenderer"
)

const (
	movementSpeed      = 220.0
	hudRefreshInterval = 100 * time.Millisecond
	mapChunkSize       = 512
	mapCullMargin      = 64
)

type mapChunk struct {
	bounds image.Rectangle
	node   raylibRenderer.Renderable
}

type Service struct {
	common.Service
	renderer       raylibRenderer.Dependency
	input          input.Dependency
	files          fs.FS
	language       LanguageSource
	Config         Config
	state          *scene.State
	mapChunks      []mapChunk
	heroNode       raylibRenderer.Renderable
	hudNode        raylibRenderer.Renderable
	hudText        string
	mapLabel       string
	lastTick       time.Time
	lastHUDRefresh time.Time
}

// LanguageSource is the narrow localization seam needed by the compatibility
// HUD while the Lua-authored UI replaces it.
type LanguageSource interface {
	GetSupportedLanguages() []string
}

// New constructs the compatibility world scene with explicit dependencies.
func New(renderer raylibRenderer.Dependency, inputService input.Dependency, files fs.FS, language LanguageSource, config Config) *Service {
	return &Service{renderer: renderer, input: inputService, files: files, language: language, Config: config}
}

func (s *Service) Name() string { return "Game Scene" }

func (s *Service) Ready() bool {
	return s.renderer != nil && s.input != nil && s.files != nil && s.language != nil
}

func (s *Service) DependenciesResolved() bool { return s.Ready() && s.renderer.IsInit() }

func (s *Service) ResolveDependencies(services []servicemesh.Service) {
	for _, service := range services {
		switch candidate := service.(type) {
		case raylibRenderer.Dependency:
			s.renderer = candidate
		case input.Dependency:
			s.input = candidate
		case fileLoader.Dependency:
			if s.files == nil {
				s.files = candidate.FromGroups()
			}
		case locale.Dependency:
			if s.language == nil {
				s.language = candidate
			}
		}
	}
}

func (s *Service) Init(servicemesh.Mesh) {
	_ = s.Start(context.Background())
}

// Start creates the compatibility world renderables after its explicit native
// dependencies are ready.
func (s *Service) Start(context.Context) error {
	const fallbackWidth, fallbackHeight = 1280, 800
	mapImage, err := s.loadMapImage()
	if err != nil {
		s.Logger().Warn("using diagnostic scene grid", "error", err)
		mapImage = gridImage(fallbackWidth, fallbackHeight, 80)
		s.mapLabel = "Diagnostic grid"
	} else {
		s.mapLabel = filepath.Base(s.Config.Map)
	}
	worldWidth, worldHeight := mapImage.Bounds().Dx(), mapImage.Bounds().Dy()
	s.state = scene.New(1, float64(worldWidth), float64(worldHeight))
	for _, chunk := range splitMapImage(mapImage, mapChunkSize) {
		node := s.renderer.NewRenderable()
		node.SetOrigin(0, 0)
		node.SetPosition(float32(chunk.bounds.Min.X), float32(chunk.bounds.Min.Y))
		node.SetZIndex(0)
		node.SetImage(chunk.image)
		s.mapChunks = append(s.mapChunks, mapChunk{bounds: chunk.bounds, node: node})
	}

	s.heroNode = s.renderer.NewRenderable()
	s.heroNode.SetZIndex(10)
	s.heroNode.SetImage(heroImage(28))
	s.hudNode = s.renderer.NewRenderable()
	s.hudNode.SetOrigin(0, 0)
	s.hudNode.SetZIndex(100)
	now := time.Now()
	s.lastTick = now
	s.heroNode.OnUpdate(s.update)
	s.syncRenderState(now)
	return nil
}

// Stop detaches all compatibility scene nodes from the renderer graph.
func (s *Service) Stop(context.Context) error {
	for _, chunk := range s.mapChunks {
		chunk.node.Disable()
		chunk.node.SetParent(nil)
	}
	s.mapChunks = nil
	for _, node := range []raylibRenderer.Renderable{s.heroNode, s.hudNode} {
		if node != nil {
			node.Disable()
			node.SetParent(nil)
		}
	}
	s.heroNode = nil
	s.hudNode = nil
	return nil
}

func (s *Service) loadMapImage() (image.Image, error) {
	if !s.Config.Enabled {
		return nil, fmt.Errorf("DS1 map rendering is disabled")
	}
	var source fs.FS
	if s.Config.Source != "" {
		if strings.Contains(s.Config.Source, "$MPQ_DIRECTORY") && os.Getenv("MPQ_DIRECTORY") == "" {
			return nil, fmt.Errorf("MPQ_DIRECTORY is not configured")
		}
		candidate := fileLoader.NewSource(os.ExpandEnv(s.Config.Source))
		filesystem, err := candidate.Filesystem()
		if err != nil {
			return nil, fmt.Errorf("opening scene source: %w", err)
		}
		source = filesystem
	} else {
		source = s.files
	}
	preview, err := assetinspect.TexturedDS1Preview(source, s.Config.Map, s.Config.Tiles, s.Config.Palette)
	if err != nil {
		return nil, err
	}
	decoded, err := png.Decode(bytes.NewReader(preview))
	if err != nil {
		return nil, fmt.Errorf("decoding rendered scene: %w", err)
	}
	return decoded, nil
}

func (s *Service) update() {
	now := time.Now()
	delta := now.Sub(s.lastTick).Seconds()
	s.lastTick = now
	if delta > 0.1 {
		delta = 0.1
	}
	dx, dy := movementVectorFrom(s.input.KeyState)
	if dx != 0 || dy != 0 {
		s.state.MoveHero(dx*movementSpeed*delta, dy*movementSpeed*delta)
	}
	s.syncRenderState(now)
}

func (s *Service) syncRenderState(now time.Time) {
	s.heroNode.SetPosition(float32(s.state.Hero.X), float32(s.state.Hero.Y))
	camera := s.renderer.GetDefaultCamera()
	width, height := s.renderer.WindowSize()
	camera.Target = rl.Vector2{X: float32(s.state.Camera.X), Y: float32(s.state.Camera.Y)}
	camera.Offset = rl.Vector2{X: float32(width) / 2, Y: float32(height) / 2}
	s.syncMapChunks(camera, width, height)
	s.syncHUD(camera, width, height, now)
}

func (s *Service) syncMapChunks(camera *rl.Camera2D, width, height int) {
	zoom := camera.Zoom
	if zoom <= 0 {
		zoom = 1
	}
	halfWidth := float32(width)/(2*zoom) + mapCullMargin
	halfHeight := float32(height)/(2*zoom) + mapCullMargin
	viewport := floatBounds{
		minX: camera.Target.X - halfWidth,
		minY: camera.Target.Y - halfHeight,
		maxX: camera.Target.X + halfWidth,
		maxY: camera.Target.Y + halfHeight,
	}
	for _, chunk := range s.mapChunks {
		if viewport.intersects(chunk.bounds) {
			chunk.node.Enable()
		} else {
			chunk.node.Disable()
		}
	}
}

type chunkImage struct {
	bounds image.Rectangle
	image  image.Image
}

func splitMapImage(source image.Image, size int) []chunkImage {
	if source == nil || size <= 0 {
		return nil
	}
	bounds := source.Bounds()
	chunks := make([]chunkImage, 0, ((bounds.Dx()+size-1)/size)*((bounds.Dy()+size-1)/size))
	for y := bounds.Min.Y; y < bounds.Max.Y; y += size {
		for x := bounds.Min.X; x < bounds.Max.X; x += size {
			chunkBounds := image.Rect(x, y, min(x+size, bounds.Max.X), min(y+size, bounds.Max.Y))
			chunk := image.NewRGBA(image.Rect(0, 0, chunkBounds.Dx(), chunkBounds.Dy()))
			draw.Draw(chunk, chunk.Bounds(), source, chunkBounds.Min, draw.Src)
			chunks = append(chunks, chunkImage{bounds: chunkBounds.Sub(bounds.Min), image: chunk})
		}
	}
	return chunks
}

type floatBounds struct {
	minX, minY float32
	maxX, maxY float32
}

func (b floatBounds) intersects(rect image.Rectangle) bool {
	return b.maxX > float32(rect.Min.X) && b.minX < float32(rect.Max.X) &&
		b.maxY > float32(rect.Min.Y) && b.minY < float32(rect.Max.Y)
}

func (s *Service) syncHUD(camera *rl.Camera2D, width, height int, now time.Time) {
	s.hudNode.SetPosition(camera.Target.X-float32(width)/2+12, camera.Target.Y-float32(height)/2+12)
	if !hudRefreshDue(s.lastHUDRefresh, now, s.hudText == "") {
		return
	}
	s.lastHUDRefresh = now
	language := "Unknown"
	if languages := s.language.GetSupportedLanguages(); len(languages) != 0 {
		language = languages[0]
	}
	text := fmt.Sprintf("Dark Magic | %s | %s\nHero: %.0f, %.0f | WASD / arrows to move",
		s.mapLabel, language, s.state.Hero.X, s.state.Hero.Y)
	if text != s.hudText {
		s.hudText = text
		s.hudNode.SetImage(hudImage(text, 470, 44))
	}
}

func hudRefreshDue(last, now time.Time, uninitialized bool) bool {
	return uninitialized || last.IsZero() || now.Sub(last) >= hudRefreshInterval
}

func movementVector(states map[int32]input.InputState) (float64, float64) {
	return movementVectorFrom(func(key int32) input.InputState { return states[key] })
}

func movementVectorFrom(state func(int32) input.InputState) (float64, float64) {
	down := func(keys ...int32) bool {
		for _, key := range keys {
			keyState := state(key)
			if keyState == input.StateDown || keyState == input.StatePressed {
				return true
			}
		}
		return false
	}
	var x, y float64
	if down(rl.KeyA, rl.KeyLeft) {
		x--
	}
	if down(rl.KeyD, rl.KeyRight) {
		x++
	}
	if down(rl.KeyW, rl.KeyUp) {
		y--
	}
	if down(rl.KeyS, rl.KeyDown) {
		y++
	}
	if x != 0 && y != 0 {
		const diagonal = 0.7071067811865476
		x *= diagonal
		y *= diagonal
	}
	return x, y
}

func gridImage(width, height, spacing int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	background := color.RGBA{R: 22, G: 27, B: 30, A: 255}
	line := color.RGBA{R: 48, G: 58, B: 61, A: 255}
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.SetRGBA(x, y, background)
			if x%spacing == 0 || y%spacing == 0 {
				img.SetRGBA(x, y, line)
			}
		}
	}
	return img
}

func heroImage(size int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	center, radius := float64(size-1)/2, float64(size)/2
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			dx, dy := float64(x)-center, float64(y)-center
			if dx*dx+dy*dy <= radius*radius {
				img.SetRGBA(x, y, color.RGBA{R: 195, G: 52, B: 46, A: 255})
			}
		}
	}
	return img
}

func hudImage(text string, width, height int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	background := color.RGBA{R: 6, G: 8, B: 10, A: 215}
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.SetRGBA(x, y, background)
		}
	}
	drawer := font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(color.RGBA{R: 224, G: 218, B: 198, A: 255}),
		Face: basicfont.Face7x13,
	}
	for idx, line := range strings.Split(text, "\n") {
		drawer.Dot = fixed.P(10, 16+idx*17)
		drawer.DrawString(line)
	}
	return img
}
