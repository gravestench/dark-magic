// Command ds1_editor_assets converts generated source sheets into compact,
// palette-indexed DC6 frames owned by the standalone DS1 editor package.
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"image"
	"image/color"
	_ "image/png"
	"os"
	"path/filepath"

	"github.com/gravestench/dc6"
	xdraw "golang.org/x/image/draw"
)

type sheetSpec struct {
	Source      string
	Output      string
	FrameSize   int
	VisibleSize int
	Names       []string
}

var sheets = []sheetSpec{
	{Source: "utility.png", Output: "utility.dc6", FrameSize: 32, VisibleSize: 28, Names: []string{
		"open", "save", "undo", "redo", "zoom_in", "zoom_out", "fit", "grid",
		"auto_draw", "inspector", "collision", "warning", "visible", "hidden", "lock", "unlock",
	}},
	{Source: "authoring.png", Output: "authoring.dc6", FrameSize: 32, VisibleSize: 28, Names: []string{
		"paint", "pick", "erase", "pan", "rectangle_select", "lasso_select", "fill", "stamp",
		"floor", "wall", "shadow", "object", "path", "warp", "rotate", "mirror",
	}},
	{Source: "cursors-markers.png", Output: "cursors-markers.dc6", FrameSize: 32, VisibleSize: 28, Names: []string{
		"cursor_pointer", "cursor_pressed", "cursor_open_hand", "cursor_closed_hand",
		"cursor_paint", "cursor_pick", "cursor_erase", "cursor_forbidden",
		"selected_tile", "hover_tile", "object_anchor", "path_node", "path_arrow", "warp_marker",
		"collision_warning", "empty_tile",
	}},
	// Chrome is authored edge-to-edge because these frames are repeated at native
	// resolution. Icon sheets retain breathing room, but padding here would become
	// a visible seam at every repeated tile boundary.
	{Source: "chrome.png", Output: "chrome.dc6", FrameSize: 16, VisibleSize: 16, Names: []string{
		"panel_top_left", "panel_top", "panel_top_right", "panel_left",
		"panel_fill", "panel_right", "panel_bottom_left", "panel_bottom",
		"panel_bottom_right", "button_idle", "button_hover", "button_pressed",
		"tab_idle", "tab_selected", "divider", "scroll_thumb",
	}},
}

type manifest struct {
	Schema  string                   `json:"schema"`
	Version int                      `json:"version"`
	Palette string                   `json:"palette"`
	Sheets  map[string]sheetManifest `json:"sheets"`
}

type sheetManifest struct {
	Path     string                 `json:"path"`
	TileSize int                    `json:"tile_size"`
	Frames   map[string]interface{} `json:"frames"`
	Sizes    map[string][2]int      `json:"sizes,omitempty"`
}

// main resolves repository-relative source and output locations and reports reproducible build failures.
func main() {
	sourceRoot := flag.String("source", "tools/ds1-editor/assets/source", "generated PNG source directory")
	outputRoot := flag.String(
		"output",
		"internal/content/ds1editor/darkmagic/ds1-editor/ui",
		"embedded editor UI output directory",
	)
	flag.Parse()
	if err := build(*sourceRoot, *outputRoot); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// build publishes the shared palette, bitmap fonts, and icon sheets into the editor's embedded content root.
// Every output is derived from checked-in sources, making asset changes reviewable and reproducible.
func build(sourceRoot, outputRoot string) error {
	if err := os.MkdirAll(outputRoot, 0o755); err != nil {
		return err
	}
	palette := editorPalette()
	if err := os.WriteFile(filepath.Join(outputRoot, "palette.dat"), encodeBGRPalette(palette), 0o644); err != nil {
		return err
	}
	if err := buildFonts(outputRoot, palette); err != nil {
		return err
	}
	previewPath := filepath.Join(filepath.Dir(sourceRoot), "previews", "fonts.png")
	if err := writeFontPreview(outputRoot, previewPath); err != nil {
		return err
	}
	document := manifest{Schema: "ds1editor.ui/v1", Version: 1,
		Palette: "darkmagic/ds1-editor/ui/palette.dat", Sheets: make(map[string]sheetManifest)}
	for _, spec := range sheets {
		asset, err := openImage(filepath.Join(sourceRoot, spec.Source))
		if err != nil {
			return err
		}
		animation, err := dc6.New(1, len(spec.Names))
		if err != nil {
			return err
		}
		frames := make(map[string]interface{}, len(spec.Names))
		for index, name := range spec.Names {
			frame := extractFrame(asset, index%4, index/4, 4, 4, spec.FrameSize, spec.VisibleSize)
			if spec.Output == "chrome.dc6" {
				frame = extractChromeFrame(asset, index, name, spec.FrameSize)
			}
			encoded := animation.Directions[0].Frames[index]
			encoded.Width, encoded.Height = uint32(spec.FrameSize), uint32(spec.FrameSize)
			encoded.IndexData = quantize(frame, palette)
			frames[name] = index
		}
		data, err := dc6.Encode(animation)
		if err != nil {
			return fmt.Errorf("encode %s: %w", spec.Output, err)
		}
		if err := os.WriteFile(filepath.Join(outputRoot, spec.Output), data, 0o644); err != nil {
			return err
		}
		document.Sheets[spec.Output[:len(spec.Output)-len(filepath.Ext(spec.Output))]] = sheetManifest{
			Path: "darkmagic/ds1-editor/ui/" + spec.Output, TileSize: spec.FrameSize, Frames: frames,
		}
	}
	compositionSheet, err := buildCompositionSheet(sourceRoot, outputRoot, palette)
	if err != nil {
		return err
	}
	document.Sheets["composition"] = compositionSheet
	data, err := json.MarshalIndent(document, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(filepath.Join(outputRoot, "assets.json"), data, 0o644)
}

// extractChromeFrame converts the generated ornamental sheet into one compact
// pixel-art tile. State centers use the source's inset material instead of its
// standalone outer frame because nine-slice edges supply the actual border.
func extractChromeFrame(source image.Image, index int, name string, size int) *image.NRGBA {
	bounds := source.Bounds()
	column, row := index%4, index/4
	cell := image.Rect(
		bounds.Min.X+column*bounds.Dx()/4,
		bounds.Min.Y+row*bounds.Dy()/4,
		bounds.Min.X+(column+1)*bounds.Dx()/4,
		bounds.Min.Y+(row+1)*bounds.Dy()/4,
	)
	content := chromeArtBounds(source, cell)
	if index >= 9 && index <= 13 {
		insetX, insetY := content.Dx()/4, content.Dy()/4
		content = image.Rect(
			content.Min.X+insetX,
			content.Min.Y+insetY,
			content.Max.X-insetX,
			content.Max.Y-insetY,
		)
	}

	result := image.NewNRGBA(image.Rect(0, 0, size, size))
	if !content.Empty() {
		xdraw.NearestNeighbor.Scale(result, result.Bounds(), source, content, xdraw.Src, nil)
	}
	normalizeChromeSeams(result, name)
	return result
}

// chromeArtBounds ignores the generated sheet's opaque black gutters while
// retaining near-black stone and metal pixels inside each semantic cell.
func chromeArtBounds(source image.Image, cell image.Rectangle) image.Rectangle {
	result := image.Rectangle{}
	for y := cell.Min.Y; y < cell.Max.Y; y++ {
		for x := cell.Min.X; x < cell.Max.X; x++ {
			red, green, blue, alpha := source.At(x, y).RGBA()
			if alpha < 0x1000 || red+green+blue < 0x0800 {
				continue
			}
			point := image.Rect(x, y, x+1, y+1)
			if result.Empty() {
				result = point
			} else {
				result = result.Union(point)
			}
		}
	}
	return result
}

// normalizeChromeSeams duplicates opposing boundary samples only along axes
// that repeat at runtime. Corners retain their authored asymmetry and detail.
func normalizeChromeSeams(target *image.NRGBA, name string) {
	horizontal := name == "panel_top" || name == "panel_fill" || name == "panel_bottom" ||
		name == "button_idle" || name == "button_hover" || name == "button_pressed" ||
		name == "tab_idle" || name == "tab_selected"
	vertical := name == "panel_left" || name == "panel_fill" || name == "panel_right" ||
		name == "button_idle" || name == "button_hover" || name == "button_pressed" ||
		name == "tab_idle" || name == "tab_selected"
	bounds := target.Bounds()
	if horizontal {
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			target.Set(bounds.Max.X-1, y, target.At(bounds.Min.X, y))
		}
	}
	if vertical {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			target.Set(x, bounds.Max.Y-1, target.At(x, bounds.Min.Y))
		}
	}
}

// openImage decodes one generated PNG while keeping file lifetime local to source ingestion.
func openImage(path string) (image.Image, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	asset, _, err := image.Decode(file)
	return asset, err
}

// extractFrame crops one logical source cell, removes transparent margins, and fits it into the 32×32 contract.
func extractFrame(
	source image.Image,
	column int,
	row int,
	columns int,
	rows int,
	outputSize int,
	contentSize int,
) *image.NRGBA {
	bounds := source.Bounds()
	cell := image.Rect(
		bounds.Min.X+column*bounds.Dx()/columns,
		bounds.Min.Y+row*bounds.Dy()/rows,
		bounds.Min.X+(column+1)*bounds.Dx()/columns,
		bounds.Min.Y+(row+1)*bounds.Dy()/rows,
	)
	content := alphaBounds(source, cell)
	result := image.NewNRGBA(image.Rect(0, 0, outputSize, outputSize))
	if content.Empty() {
		return result
	}
	scale := min(float64(contentSize)/float64(content.Dx()), float64(contentSize)/float64(content.Dy()))
	width := max(1, int(float64(content.Dx())*scale+0.5))
	height := max(1, int(float64(content.Dy())*scale+0.5))
	destination := image.Rect(
		(outputSize-width)/2,
		(outputSize-height)/2,
		(outputSize+width)/2,
		(outputSize+height)/2,
	)
	xdraw.CatmullRom.Scale(result, destination, source, content, xdraw.Src, nil)
	return result
}

// alphaBounds finds visible source pixels so inconsistent generation margins cannot shrink the final icon.
func alphaBounds(source image.Image, cell image.Rectangle) image.Rectangle {
	result := image.Rectangle{}
	for y := cell.Min.Y; y < cell.Max.Y; y++ {
		for x := cell.Min.X; x < cell.Max.X; x++ {
			_, _, _, alpha := source.At(x, y).RGBA()
			if alpha < 0x1000 {
				continue
			}
			point := image.Rect(x, y, x+1, y+1)
			if result.Empty() {
				result = point
			} else {
				result = result.Union(point)
			}
		}
	}
	return result
}

// editorPalette builds a deterministic 256-color cube plus grayscale ramp with transparent index zero.
func editorPalette() color.Palette {
	result := color.Palette{color.NRGBA{}}
	levels := [...]uint8{0, 51, 102, 153, 204, 255}
	for _, red := range levels {
		for _, green := range levels {
			for _, blue := range levels {
				result = append(result, color.NRGBA{R: red, G: green, B: blue, A: 255})
			}
		}
	}
	for index := 0; index < 39; index++ {
		value := uint8(index * 255 / 38)
		result = append(result, color.NRGBA{R: value, G: value, B: value, A: 255})
	}
	return result
}

// quantize converts one 32×32 RGBA frame to palette indices and reserves index zero for transparency.
func quantize(source image.Image, palette color.Palette) []byte {
	width, height := source.Bounds().Dx(), source.Bounds().Dy()
	result := make([]byte, width*height)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			value := color.NRGBAModel.Convert(source.At(x, y)).(color.NRGBA)
			if value.A < 96 {
				continue
			}
			value.A = 255
			result[y*width+x] = uint8(palette.Index(value))
		}
	}
	return result
}

// encodeBGRPalette writes Diablo's packed BGR palette representation without an alpha channel.
func encodeBGRPalette(palette color.Palette) []byte {
	var output bytes.Buffer
	for _, value := range palette {
		red, green, blue, _ := value.RGBA()
		output.WriteByte(byte(blue >> 8))
		output.WriteByte(byte(green >> 8))
		output.WriteByte(byte(red >> 8))
	}
	return output.Bytes()
}
