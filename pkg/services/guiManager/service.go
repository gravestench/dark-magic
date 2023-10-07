package guiManager

import (
	"image"
	"image/color"
	"image/draw"
	"time"

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
}

func (s *Service) Init(rt runtime.Runtime) {
	// This init method will be invoked by the runtime
	// as soon as the dependency resolution has finished.
	// If the service does not implement runtime.HasDependencies,
	// then this method is invoked immediately.

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

	test := s.renderer.NewRenderable()
	// Create a new 10x10 image with RGBA color model
	img := image.NewRGBA(image.Rect(0, 0, 3, 3))

	// Fill the image with red color
	red := color.RGBA{255, 0, 0, 255}
	draw.Draw(img, img.Bounds(), &image.Uniform{red}, image.Point{}, draw.Src)

	test.SetImage(img)

	cursor := s.renderer.NewRenderable()
	test.SetParent(cursor)
	test.SetPosition(-3, -3)
	cursor.SetScale(3)
	d1 := dc6Cursor.Directions[0]
	frames := d1.Frames
	frame := 0
	forward := true
	cursor.SetImage(frames[frame].ToImageRGBA())

	t := time.Now()

	sampler := s.logger.Sample(&zerolog.BasicSampler{N: 24})

	cursor.OnUpdate(func() {
		if time.Since(t) < time.Second/24 {
			return
		}

		t = time.Now()

		if frame == len(frames)-1 {
			forward = false
		} else if frame == 0 {
			forward = true
		}

		if forward {
			frame++
		} else {
			frame--
		}

		r := cursor.Rotation()
		r += 2
		cursor.SetRotation(r)

		tx, ty := test.Position()
		sampler.Info().Msgf("test: %v, %v", tx, ty)

		x, y := s.input.MouseCursorState()
		cursor.SetPosition(float32(x), float32(y))
		cursor.SetImage(frames[frame%len(frames)].ToImageRGBA())
	})
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
