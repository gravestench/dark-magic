package spriteManager

import (
	"bytes"
	"fmt"
	"image"
	"image/draw"
	"image/png"
)

func (s *Service) LoadDt1ToPngSpriteAtlas(pathDT1 string, pathPL2 string) ([]byte, error) {
	cacheKey := fmt.Sprintf("dt1: %s %s", pathDT1, pathPL2)

	if s.spriteCache != nil {
		cachedData, isCached := s.spriteCache.Retrieve(cacheKey)
		if isCached {
			return cachedData.([]byte), nil
		}
	}

	// the palette RGBA data is the first 256 x 4 bytes of the PL2 file
	pl2Palette, err := s.extractPaletteFromPl2(pathPL2)
	if err != nil {
		return nil, fmt.Errorf("extracting palette from pl2", "error", err)
	}

	dt1, err := s.dt1.Load(pathDT1)
	if err != nil {
		return nil, fmt.Errorf("loading DT1 tileset", "error", err)
	}

	dt1.SetPalette(pl2Palette)

	floorImages := make([]image.Image, 0)
	wallImages := make([]image.Image, 0)

	for _, tile := range dt1.Tiles {
		floorImages = append(floorImages, tile.FloorImage())
		wallImages = append(wallImages, tile.WallImage())
	}

	floorAtlas, _ := generateVerticalComposite(floorImages)
	wallAtlas, _ := generateVerticalComposite(wallImages)

	atlas := generateHorizontalComposite([]image.Image{floorAtlas, wallAtlas})

	// Encode the sprite atlas as a PNG.
	pngData := bytes.NewBuffer([]byte{})
	if err := png.Encode(pngData, atlas); err != nil {
		return nil, err
	}

	//atlasInfoData, err := json.Marshal(atlasInfo)
	//if err != nil {
	//	return nil, fmt.Errorf("marshalling atlas frame info", "error", err)
	//}
	//
	//pngDataWithExtras := append(pngData.Bytes(), atlasInfoData...)

	if s.spriteCache != nil {
		if err = s.spriteCache.Insert(cacheKey, pngData.Bytes(), len(pngData.Bytes())); err != nil {
			s.logger.Error("caching %s: %v", pathPL2, err)
		}
	}

	return pngData.Bytes(), nil
}

// generateVerticalComposite creates a sprite atlas and returns the atlas image
// along with a slice of LayeredAnimation structs for each image.
func generateVerticalComposite(images []image.Image) (*image.RGBA, []frameInfo) {
	// Step 1: Calculate total height and maximum width of the atlas
	maxWidth, totalHeight := 0, 0
	for _, img := range images {
		if img.Bounds().Dx() > maxWidth {
			maxWidth = img.Bounds().Dx()
		}
		totalHeight += img.Bounds().Dy()
	}

	// Step 2: Create a new image with the calculated dimensions
	atlas := image.NewRGBA(image.Rect(0, 0, maxWidth, totalHeight))

	// Step 3 & 4: Draw each image onto the sprite atlas and save its position and dimensions
	sprites := make([]frameInfo, len(images))
	yOffset := 0
	for i, img := range images {
		dstRect := image.Rect(0, yOffset, img.Bounds().Dx(), yOffset+img.Bounds().Dy())
		draw.Draw(atlas, dstRect, img, img.Bounds().Min, draw.Over)

		sprites[i] = frameInfo{
			X:      0,
			Y:      yOffset,
			Width:  img.Bounds().Dx(),
			Height: img.Bounds().Dy(),
		}
		yOffset += img.Bounds().Dy()
	}

	return atlas, sprites
}

func generateHorizontalComposite(images []image.Image) *image.RGBA {
	// Step 1: Calculate total width and maximum height of the composite image
	totalWidth, maxHeight := 0, 0
	for _, img := range images {
		totalWidth += img.Bounds().Dx()
		if img.Bounds().Dy() > maxHeight {
			maxHeight = img.Bounds().Dy()
		}
	}

	// Step 2: Create a new image with the calculated dimensions
	composite := image.NewRGBA(image.Rect(0, 0, totalWidth, maxHeight))

	// Step 3: Draw each image onto the composite image
	xOffset := 0
	for _, img := range images {
		dstRect := image.Rect(xOffset, 0, xOffset+img.Bounds().Dx(), img.Bounds().Dy())
		draw.Draw(composite, dstRect, img, img.Bounds().Min, draw.Src)
		xOffset += img.Bounds().Dx()
	}

	return composite
}
