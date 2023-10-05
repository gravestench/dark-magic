package guiManager

import (
	"image"
	"image/color"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/gravestench/runtime"
	"github.com/rs/zerolog"

	"github.com/gravestench/dark-magic/pkg/paths"
	"github.com/gravestench/dark-magic/pkg/services/dc6Loader"
	"github.com/gravestench/dark-magic/pkg/services/input"
	"github.com/gravestench/dark-magic/pkg/services/mpqLoader"
	"github.com/gravestench/dark-magic/pkg/services/raylibRenderer"
	"github.com/gravestench/dark-magic/pkg/services/spriteManager"
)

type Service struct {
	logger *zerolog.Logger

	dc6      dc6Loader.Dependency
	mpq      mpqLoader.Dependency
	sprite   spriteManager.Dependency
	renderer raylibRenderer.Dependency
	input    input.Dependency

	inputState struct {
		keys []int
	}

	texture *rl.Texture2D

	root Node
}

func (s *Service) Init(rt runtime.Runtime) {
	// This init method will be invoked by the runtime
	// as soon as the dependency resolution has finished.
	// If the service does not implement runtime.HasDependencies,
	// then this method is invoked immediately.

	s.root = s.NewTreeNode()

	paletteStream, err := s.mpq.Load(paths.PaletteTransformAct1)
	for err != nil {
		// TODO :: fix a race condition with the mpq loader
		time.Sleep(time.Millisecond * 5)
		paletteStream, err = s.mpq.Load(paths.PaletteTransformAct1)
	}

	const (
		numColors    = 256
		numBytesRGBA = numColors * 4
		opaque       = 255
	)

	paletteData := make([]byte, numBytesRGBA)
	numRead, err := paletteStream.Read(paletteData)
	if err != nil {
		s.logger.Fatal().Msgf("reading from PL2 stream: %v", err)
	} else if numRead != numBytesRGBA {
		s.logger.Fatal().Msgf("couldn't read all palette bytes")
	}

	act1 := make(color.Palette, numColors)
	for idx := range act1 {
		if idx == 0 {
			act1[idx] = image.Transparent

			continue
		}

		act1[idx] = color.RGBA{
			R: paletteData[(idx*4)+0],
			G: paletteData[(idx*4)+1],
			B: paletteData[(idx*4)+2],
			A: opaque,
		}
	}

	dc6Cursor, err := s.dc6.Load(paths.CursorDefault)
	for err != nil {
		// TODO :: fix a race condition with the mpq loader
		time.Sleep(time.Millisecond * 10)
		dc6Cursor, err = s.dc6.Load(paths.CursorDefault)
	}

	dc6Cursor.SetPalette(act1)

	cursor := s.NewNode()
	cursor.SetImage(dc6Cursor.Directions[0].Frames[0].ToImageRGBA())

	cursor.SetUpdateFunc(func() {
		x, y := s.input.MouseCursorState()
		cursor.SetPosition(float32(x), float32(y))
		r := cursor.Rotation()
		r += 0.01
		cursor.SetRotation(r)
	})

	ticker := time.NewTicker(time.Second / 60)

	// Create a Goroutine to handle the ticker
	go func() {
		for {
			select {
			case <-ticker.C:
				// Call your function here
				s.Update()
			}
		}
	}()
}

func (s *Service) Name() string {
	return "GUI"
}

// the following methods are boilerplate, but they are used
// by the runtime to enforce a standard logging format.

func (s *Service) BindLogger(logger *zerolog.Logger) {
	s.logger = logger
}

func (s *Service) Logger() *zerolog.Logger {
	return s.logger
}

func (s *Service) NewNode() Node {
	n := s.NewTreeNode()

	s.root.AddChild(n)

	return n
}

func (s *Service) Nodes() []Node {
	return s.root.Children()
}

func (s *Service) Update() {
	s.root.Update()
}
