package spriteManager

import (
	"fmt"
	"image"

	"github.com/gravestench/dc6"
)

type LayeredAnimation struct {
	Layers []Layer
}

type Layer struct {
	Frames []Frame
}

type Frame struct {
	Image  image.Image
	Offset image.Point
}

func (s *Service) LoadDc6ToLayeredAnimation(pathDC6 string, pathPL2 string) (*LayeredAnimation, error) {
	cacheKey := fmt.Sprintf("layered animation: %s %s", pathDC6, pathPL2)

	if s.spriteCache != nil {
		cachedData, isCached := s.spriteCache.Retrieve(cacheKey)
		if isCached {
			return cachedData.(*LayeredAnimation), nil
		}
	}

	dc6Image, err := s.loadDc6WithPalette(pathDC6, pathPL2)
	if err != nil {
		return nil, fmt.Errorf("loading DC6 with palette: %v", err)
	}

	anim := &LayeredAnimation{}

	for idx, direction := range dc6Image.Directions {
		anim.Layers = append(anim.Layers, Layer{})
		anim.Layers[idx].Frames = adjustFramesDimensions(direction.Frames)
	}

	return anim, nil
}

// AdjustFramesDimensions adjusts the dimensions of frames to the maximum width and height.
func adjustFramesDimensions(frames []*dc6.Frame) []Frame {
	if len(frames) == 0 {
		return nil
	}

	// Find the maximum dimensions and calculate the total height
	maxWidth := 0
	maxHeight := 0
	for _, frame := range frames {
		rect := frame.Bounds()
		if rect.Dx() > maxWidth {
			maxWidth = rect.Dx()
		}

		if rect.Dy() > maxHeight {
			maxHeight = rect.Dy()
		}
	}

	// Create new frames with adjusted dimensions and aligned to the bottom-left
	adjustedFrames := make([]Frame, len(frames))
	yOffset := maxHeight // Initial Y offset at the bottom
	for i, frame := range frames {
		rect := frame.Bounds()
		offset := image.Point{
			X: int(frame.OffsetX),
			Y: int(frame.OffsetY),
		}

		// Create a new canvas image with the maximum dimensions
		canvas := image.NewRGBA(image.Rect(0, 0, maxWidth, maxHeight))

		// Calculate the Y offset for alignment to the bottom-left
		yOffset -= offset.Y

		// Paste the original image onto the canvas at the correct offset
		for y := 0; y < rect.Dy(); y++ {
			for x := 0; x < rect.Dx(); x++ {
				pixel := frame.At(x, y)
				canvas.Set(x, y, pixel)
			}
		}

		// Create an adjusted frame with the new canvas and original offset
		adjustedFrames[i] = Frame{
			Image:  canvas,
			Offset: offset,
		}
	}

	return adjustedFrames
}
