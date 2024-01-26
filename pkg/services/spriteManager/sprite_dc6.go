package spriteManager

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/color/palette"
	"image/draw"
	"image/gif"
	"image/png"

	dc6 "github.com/gravestench/dc6/pkg"
)

type frameInfo struct {
	X, Y          int
	Width, Height int
}

func (s *Service) loadDc6WithPalette(pathDC6 string, pathPL2 string) (*dc6.DC6, error) {
	// the palette RGBA data is the first 256 x 4 bytes of the PL2 file
	pl2Palette, err := s.extractPaletteFromPl2(pathPL2)
	if err != nil {
		return nil, fmt.Errorf("extracting palette from pl2: %v", err)
	}

	dc6Image, err := s.dc6.Load(pathDC6)
	if err != nil {
		return nil, fmt.Errorf("loading dc6: %v", err)
	}

	dc6Image.SetPalette(pl2Palette)

	return dc6Image, nil
}

func (s *Service) LoadDc6ToPngSpriteAtlas(pathDC6 string, pathPL2 string) ([]byte, error) {
	cacheKey := fmt.Sprintf("png: %s %s", pathDC6, pathPL2)

	if s.spriteCache != nil {
		cachedData, isCached := s.spriteCache.Retrieve(cacheKey)
		if isCached {
			return cachedData.([]byte), nil
		}
	}

	dc6Image, err := s.loadDc6WithPalette(pathDC6, pathPL2)
	if err != nil {
		return nil, fmt.Errorf("loading DC6 with palette: %v", err)
	}

	frames := make([]*dc6.Frame, 0)
	for _, dir := range dc6Image.Directions {
		for _, frame := range dir.Frames {
			frames = append(frames, frame)
		}
	}

	pngData, err := generateDc6SpriteAtlasPng(frames)
	if err != nil {
		return nil, fmt.Errorf("generating sprite atlas")
	}

	if s.spriteCache != nil {
		if err = s.spriteCache.Insert(cacheKey, pngData, len(pngData)); err != nil {
			s.logger.Error("caching %s: %v", pathDC6, err)
		}
	}

	return pngData, nil
}

func (s *Service) LoadDc6ToAnimatedGif(pathDC6 string, pathPL2 string) ([]byte, error) {
	data, err := s.LoadDc6ToPngSpriteAtlas(pathDC6, pathPL2)
	if err != nil {
		return nil, fmt.Errorf("loading file: %v", err)
	}

	var gifImage []byte

	cacheKey := fmt.Sprintf("gif: %s %s", pathDC6, pathPL2)
	cachedData, isCached := s.spriteCache.Retrieve(cacheKey)
	if isCached {
		return cachedData.([]byte), nil
	}

	gifImage, err = generateAnimatedGifFromPngSpriteAtlasData(data)
	if err != nil {
		return nil, fmt.Errorf("creating animated GIF from png sprite atlas: %v", err)
	}

	if err = s.spriteCache.Insert(cacheKey, gifImage, len(gifImage)); err != nil {
		s.logger.Error("caching animated gif", "error", err)
	}

	return gifImage, nil
}

func (s *Service) extractPaletteFromPl2(pathPL2 string) (color.Palette, error) {
	cacheKey := fmt.Sprintf("pl2: %s", pathPL2)

	if s.spriteCache != nil {
		cachedData, isCached := s.spriteCache.Retrieve(cacheKey)
		if isCached {
			return cachedData.(color.Palette), nil
		}
	}

	paletteStream, err := s.mpq.Load(pathPL2)
	if err != nil {
		return nil, fmt.Errorf("loading pl2: %v", err)
	}

	const (
		numColors    = 256
		numBytesRGBA = numColors * 4
		opaque       = 255
	)

	paletteData := make([]byte, numBytesRGBA)
	numRead, err := paletteStream.Read(paletteData)
	if err != nil {
		return nil, fmt.Errorf("reading from PL2 stream: %v", err)
	} else if numRead != numBytesRGBA {
		return nil, fmt.Errorf("couldn't read all palette bytes")
	}

	p := make(color.Palette, numColors)
	for idx := range p {
		if idx == 0 {
			p[idx] = image.Transparent

			continue
		}

		p[idx] = color.RGBA{
			R: paletteData[(idx*4)+0],
			G: paletteData[(idx*4)+1],
			B: paletteData[(idx*4)+2],
			A: opaque,
		}
	}

	if s.spriteCache != nil {
		if err = s.spriteCache.Insert(cacheKey, p, numBytesRGBA); err != nil {
			s.logger.Error("caching %s: %v", pathPL2, err)
		}
	}

	return p, nil
}

// generateDc6SpriteAtlas generates a PNG sprite atlas from a slice of *image.RGBA.
func generateDc6SpriteAtlasPng(frames []*dc6.Frame) ([]byte, error) {
	// Calculate the dimensions of the sprite atlas.
	atlasWidth := 0
	atlasHeight := 0

	maxFrameWidth := 0
	maxFrameHeight := 0

	for _, frame := range frames {
		frameRight := frame.Bounds().Dx()
		frameBottom := frame.Bounds().Dy()

		if frameRight > atlasWidth {
			atlasWidth = frameRight
			maxFrameWidth = atlasWidth
		}

		if frameBottom > maxFrameHeight {
			maxFrameHeight = frameBottom
		}

		atlasHeight += frameBottom
	}

	// Create an empty RGBA image for the sprite atlas.
	atlas := image.NewRGBA(image.Rect(0, 0, atlasWidth, atlasHeight))
	atlasInfo := make([]frameInfo, 0)

	// Place each frame in the sprite atlas.
	frameY := 0
	for _, frame := range frames {
		frameBottom := int(frame.OffsetY) + frame.Bounds().Dy()
		dstRect := image.Rect(int(frame.OffsetX), int(frame.OffsetY), int(frame.OffsetX)+frame.Bounds().Dx(), int(frame.OffsetY)+frame.Bounds().Dy())

		atlasInfo = append(atlasInfo, frameInfo{
			X:      0,
			Y:      frameY,
			Width:  maxFrameWidth,
			Height: maxFrameHeight,
		})

		drawRect := image.Rect(0, 0, frame.Bounds().Dx(), frame.Bounds().Dy())
		drawRect.Max = drawRect.Max.Add(dstRect.Min.Sub(drawRect.Min))

		// Copy the image to the sprite atlas.
		for x := 0; x < frame.Bounds().Dx(); x++ {
			for y := 0; y < frame.Bounds().Dy(); y++ {
				atlas.Set(x, frameY+y, frame.At(drawRect.Min.X+x, drawRect.Min.Y+y))
			}
		}

		frameY += frameBottom
	}

	// Encode the sprite atlas as a PNG.
	pngData := bytes.NewBuffer([]byte{})
	if err := png.Encode(pngData, atlas); err != nil {
		return nil, err
	}

	atlasInfoData, err := json.Marshal(atlasInfo)
	if err != nil {
		return nil, fmt.Errorf("marshalling atlas frame info: %v", err)
	}

	pngDataWithExtras := append(pngData.Bytes(), atlasInfoData...)

	return pngDataWithExtras, nil
}

func generateAnimatedGifFromPngSpriteAtlasData(atlas []byte) ([]byte, error) {
	atlasImg, _, err := image.Decode(bytes.NewReader(atlas))
	if err != nil {
		return nil, fmt.Errorf("decoding sprite atlas: %v", err)
	}

	// Create an array of frame information.
	endOfPngData := findPngEnd(atlas)
	frameInfodata := atlas[endOfPngData+4:]

	var frameInfos []frameInfo
	if err = json.Unmarshal(frameInfodata, &frameInfos); err != nil {
		return nil, fmt.Errorf("getting frame info from sprite atlas: %v", err)
	}

	// Create a GIF writer.
	g := gif.GIF{}

	// Set the GIF loop count (0 means infinite loop).
	g.LoopCount = 0

	// Set the frame duration for 24 fps.
	frameDuration := 100 / 24 // in milliseconds

	// Loop through each frame and create GIF frames.
	for _, info := range frameInfos {
		// Create a new paletted image for the frame.
		frame := image.NewPaletted(image.Rect(0, 0, info.Width, info.Height), palette.WebSafe)

		// Draw the sprite from the atlas onto the frame.
		draw.Draw(frame, frame.Bounds(), atlasImg, image.Point{info.X, info.Y}, draw.Src)

		// Add the frame to the GIF with the specified duration.
		g.Image = append(g.Image, frame)
		g.Delay = append(g.Delay, frameDuration)
		g.Disposal = append(g.Disposal, uint8(0))
	}

	// Encode the GIF and save it to the output file.
	gifData := bytes.NewBuffer([]byte{})
	err = gif.EncodeAll(gifData, &g)
	if err != nil {
		return nil, fmt.Errorf("encoding GIF: %v", err)
	}

	return gifData.Bytes(), nil
}
