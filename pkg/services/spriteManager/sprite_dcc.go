package spriteManager

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/png"

	"github.com/gravestench/dcc"
)

func (s *Service) LoadDccToPngSpriteAtlas(pathDCC string, pathPL2 string) ([]byte, error) {
	cacheKey := fmt.Sprintf("png: %s %s", pathDCC, pathPL2)

	if s.spriteCache != nil {
		cachedData, isCached := s.spriteCache.Retrieve(cacheKey)
		if isCached {
			return cachedData.([]byte), nil
		}
	}

	// the palette RGBA data is the first 256 x 4 bytes of the PL2 file
	pl2Palette, err := s.extractPaletteFromPl2(pathPL2)
	if err != nil {
		return nil, fmt.Errorf("extracting palette from pl2: %v", err)
	}

	dccImage, err := s.dcc.Load(pathDCC)
	if err != nil {
		return nil, fmt.Errorf("loading dc6: %v", err)
	}

	dccImage.SetPalette(pl2Palette)

	frames := make([]*dcc.Frame, 0)
	for _, dir := range dccImage.Directions() {
		for _, frame := range dir.Frames() {
			frames = append(frames, frame)
		}
	}

	pngData, err := generateDccSpriteAtlasPng(frames)
	if err != nil {
		return nil, fmt.Errorf("generating sprite atlas")
	}

	if s.spriteCache != nil {
		if err = s.spriteCache.Insert(cacheKey, pngData, len(pngData)); err != nil {
			s.logger.Error().Msgf("caching %s: %v", pathDCC, err)
		}
	}

	return pngData, nil
}

func (s *Service) LoadDccToAnimatedGif(pathDCC string, pathPL2 string) ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

// generateDccSpriteAtlas generates a PNG sprite atlas from a slice of *image.RGBA.
func generateDccSpriteAtlasPng(frames []*dcc.Frame) ([]byte, error) {
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
		frameBottom := frame.Bounds().Dy()
		dstRect := image.Rect(0, 0, frame.Bounds().Dx(), frame.Bounds().Dy())

		atlasInfo = append(atlasInfo, frameInfo{
			X:      0,
			Y:      frameY,
			Width:  maxFrameWidth,
			Height: maxFrameHeight,
		})

		drawRect := image.Rect(frame.Bounds().Min.X, frame.Bounds().Min.Y, frame.Bounds().Dx(), frame.Bounds().Dy())
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
