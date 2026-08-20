package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"

	assetdecode "github.com/gravestench/dark-magic/internal/assets/decode"
)

const fontPreviewWidth = 960

var fontPreviewSamples = map[string]string{
	"large":      "LARGE  MAP  TITLES",
	"medium":     "Medium tools, layers & metadata 0123456789",
	"small":      "Small labels: floor wall shadow object path warp",
	"very_small": "VERY SMALL  X:128 Y:064  TYPE:03 STYLE:12 SEQ:04  ← ↑ → ↓",
}

// writeFontPreview renders the encoded files back through the production decoder.
// The contact sheet is both visual documentation and an integration check of the final artifacts.
func writeFontPreview(outputRoot, previewPath string) error {
	rows := make([]image.Image, 0, len(editorFontSpecs))
	for _, spec := range editorFontSpecs {
		font, err := assetdecode.LoadBitmapFont(
			os.DirFS(outputRoot),
			"fonts/"+spec.Name+".tbl",
			"fonts/"+spec.Name+".dc6",
			"palette.dat",
		)
		if err != nil {
			return fmt.Errorf("preview %s font: %w", spec.Name, err)
		}

		row, err := font.Render(fontPreviewSamples[spec.Name], color.White, 0, "left")
		if err != nil {
			return fmt.Errorf("render %s font preview: %w", spec.Name, err)
		}
		rows = append(rows, row)
	}

	preview := composeFontPreview(rows)
	if err := os.MkdirAll(filepath.Dir(previewPath), 0o755); err != nil {
		return err
	}

	file, err := os.Create(previewPath)
	if err != nil {
		return err
	}
	defer file.Close()

	return png.Encode(file, preview)
}

// composeFontPreview places each size on a separate dark panel with generous inspection spacing.
func composeFontPreview(rows []image.Image) image.Image {
	const (
		outerPadding = 24
		rowGap       = 16
		rowPadding   = 12
	)

	height := outerPadding * 2
	for _, row := range rows {
		height += row.Bounds().Dy() + rowPadding*2 + rowGap
	}
	height -= rowGap

	preview := image.NewRGBA(image.Rect(0, 0, fontPreviewWidth, height))
	background := image.NewUniform(color.RGBA{R: 14, G: 18, B: 22, A: 255})
	draw.Draw(preview, preview.Bounds(), background, image.Point{}, draw.Src)

	y := outerPadding
	for _, row := range rows {
		panel := image.Rect(outerPadding, y, fontPreviewWidth-outerPadding, y+row.Bounds().Dy()+rowPadding*2)
		draw.Draw(preview, panel, image.NewUniform(color.RGBA{R: 31, G: 38, B: 43, A: 255}), image.Point{}, draw.Src)
		position := image.Pt(panel.Min.X+rowPadding, panel.Min.Y+rowPadding)
		draw.Draw(preview, row.Bounds().Add(position), row, row.Bounds().Min, draw.Over)
		y = panel.Max.Y + rowGap
	}

	return preview
}
