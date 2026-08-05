package gameScene

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
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

const movementSpeed = 220.0

type Service struct {
	common.Service
	renderer raylibRenderer.Dependency
	input    input.Dependency
	files    fileLoader.Dependency
	locale   locale.Dependency
	Config   Config
	state    *scene.State
	mapNode  raylibRenderer.Renderable
	heroNode raylibRenderer.Renderable
	hudNode  raylibRenderer.Renderable
	hudText  string
	mapLabel string
	lastTick time.Time
}

func (s *Service) Name() string { return "Game Scene" }

func (s *Service) Ready() bool {
	return s.renderer != nil && s.input != nil && s.files != nil && s.locale != nil
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
			s.files = candidate
		case locale.Dependency:
			s.locale = candidate
		}
	}
}

func (s *Service) Init(servicemesh.Mesh) {
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
	s.mapNode = s.renderer.NewRenderable()
	s.mapNode.SetOrigin(0, 0)
	s.mapNode.SetZIndex(0)
	s.mapNode.SetImage(mapImage)

	s.heroNode = s.renderer.NewRenderable()
	s.heroNode.SetZIndex(10)
	s.heroNode.SetImage(heroImage(28))
	s.hudNode = s.renderer.NewRenderable()
	s.hudNode.SetOrigin(0, 0)
	s.hudNode.SetZIndex(100)
	s.lastTick = time.Now()
	s.heroNode.OnUpdate(s.update)
	s.syncRenderState()
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
		source = s.files.FromGroups()
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
	dx, dy := movementVector(s.input.KeyboardState())
	if dx != 0 || dy != 0 {
		s.state.MoveHero(dx*movementSpeed*delta, dy*movementSpeed*delta)
	}
	s.syncRenderState()
}

func (s *Service) syncRenderState() {
	s.heroNode.SetPosition(float32(s.state.Hero.X), float32(s.state.Hero.Y))
	camera := s.renderer.GetDefaultCamera()
	width, height := s.renderer.WindowSize()
	camera.Target = rl.Vector2{X: float32(s.state.Camera.X), Y: float32(s.state.Camera.Y)}
	camera.Offset = rl.Vector2{X: float32(width) / 2, Y: float32(height) / 2}
	s.syncHUD(camera, width, height)
}

func (s *Service) syncHUD(camera *rl.Camera2D, width, height int) {
	language := "Unknown"
	if languages := s.locale.GetSupportedLanguages(); len(languages) != 0 {
		language = languages[0]
	}
	text := fmt.Sprintf("Dark Magic | %s | %s\nHero: %.0f, %.0f | WASD / arrows to move",
		s.mapLabel, language, s.state.Hero.X, s.state.Hero.Y)
	if text != s.hudText {
		s.hudText = text
		s.hudNode.SetImage(hudImage(text, 470, 44))
	}
	s.hudNode.SetPosition(camera.Target.X-float32(width)/2+12, camera.Target.Y-float32(height)/2+12)
}

func movementVector(states map[int32]input.InputState) (float64, float64) {
	down := func(keys ...int32) bool {
		for _, key := range keys {
			if states[key] == input.StateDown || states[key] == input.StatePressed {
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
