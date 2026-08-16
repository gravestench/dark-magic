package assets

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"image"
	"image/draw"
	"image/png"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	assetdecode "github.com/gravestench/dark-magic/internal/assets/decode"
)

type fontSpec struct {
	table     string
	sheet     string
	palette   string
	transform string
	color     int
}

var fontAllowlist = map[string]fontSpec{
	"font16-red":  {"data/local/FONT/latin/font16.tbl", "data/local/FONT/latin/font16.dc6", "data/global/Palette/units/pal.dat", "data/global/Palette/Act1/Pal.pl2", 1},
	"font16-gold": {"data/local/FONT/latin/font16.tbl", "data/local/FONT/latin/font16.dc6", "data/global/Palette/units/pal.dat", "data/global/Palette/Act1/Pal.pl2", 4},
	"font16-grey": {"data/local/FONT/latin/font16.tbl", "data/local/FONT/latin/font16.dc6", "data/global/Palette/units/pal.dat", "data/global/Palette/Act1/Pal.pl2", 5},
	"font16-blue": {"data/local/FONT/latin/font16.tbl", "data/local/FONT/latin/font16.dc6", "data/global/Palette/units/pal.dat", "data/global/Palette/Act1/Pal.pl2", 3},
	"exocet-grey": {"data/local/FONT/latin/fontexocet10.tbl", "data/local/FONT/latin/fontexocet10.dc6", "data/global/Palette/units/pal.dat", "", -1},
}

type fontGlyph struct {
	X       int `json:"x"`
	Y       int `json:"y"`
	Width   int `json:"width"`
	Height  int `json:"height"`
	Advance int `json:"advance"`
	OffsetX int `json:"offset_x"`
	OffsetY int `json:"offset_y"`
}

type fontMetadata struct {
	Image      string               `json:"image"`
	LineHeight int                  `json:"line_height"`
	Glyphs     map[string]fontGlyph `json:"glyphs"`
}

func (cache *Cache) serveFont(writer http.ResponseWriter, request *http.Request) {
	name := strings.TrimPrefix(request.URL.Path, "/account/fonts/")
	extension := filepath.Ext(name)
	id := strings.TrimSuffix(name, extension)
	pngPath, jsonPath, err := cache.renderFont(id)
	if err != nil {
		http.NotFound(writer, request)
		return
	}
	// Font aliases are stable while their generated atlases are versioned in the
	// private cache, so clients must revalidate the alias after renderer changes.
	writer.Header().Set("Cache-Control", "private, no-cache")
	switch extension {
	case ".png":
		writer.Header().Set("Content-Type", "image/png")
		http.ServeFile(writer, request, pngPath)
	case ".json":
		writer.Header().Set("Content-Type", "application/json")
		http.ServeFile(writer, request, jsonPath)
	default:
		http.NotFound(writer, request)
	}
}

func (cache *Cache) renderFont(id string) (string, string, error) {
	spec, found := fontAllowlist[id]
	if !found {
		return "", "", fs.ErrNotExist
	}

	cache.mu.Lock()
	defer cache.mu.Unlock()
	key := "font:" + id
	if prefix := cache.cached[key]; prefix != "" {
		return prefix + ".png", prefix + ".json", nil
	}

	hash := sha256.New()
	_, _ = io.WriteString(hash, rendererVersion+id)
	for _, path := range []string{spec.table, spec.sheet, spec.palette, spec.transform} {
		if path == "" {
			continue
		}
		data, err := fs.ReadFile(cache.source, path)
		if err != nil {
			return "", "", err
		}
		_, _ = hash.Write(data)
	}
	prefix := filepath.Join(cache.directory, id+"-"+hex.EncodeToString(hash.Sum(nil))[:20])
	if _, pngErr := os.Stat(prefix + ".png"); pngErr == nil {
		if _, jsonErr := os.Stat(prefix + ".json"); jsonErr == nil {
			cache.cached[key] = prefix
			return prefix + ".png", prefix + ".json", nil
		}
	}

	font, err := assetdecode.LoadBitmapFontWithTransform(
		cache.source,
		spec.table,
		spec.sheet,
		spec.palette,
		spec.transform,
	)
	if err != nil {
		return "", "", err
	}
	codes := make([]rune, 0, len(font.Glyphs))
	for code := range font.Glyphs {
		codes = append(codes, code)
	}
	sort.Slice(codes, func(i, j int) bool { return codes[i] < codes[j] })

	const atlasWidth = 1024
	x, y, rowHeight := 0, 0, 0
	metadata := fontMetadata{
		Image:      "/account/fonts/" + id + ".png?revision=3",
		LineHeight: font.LineHeight,
		Glyphs:     make(map[string]fontGlyph),
	}
	placements := make(map[rune]image.Rectangle, len(codes))
	for _, code := range codes {
		glyph := font.Glyphs[code]
		frame := font.Frames[glyph.Frame]
		if transformed := font.TextFrames[spec.color]; glyph.Frame < len(transformed) {
			frame = transformed[glyph.Frame]
		}
		width, height := frame.Bounds().Dx(), frame.Bounds().Dy()
		if x+width+1 > atlasWidth {
			x, y, rowHeight = 0, y+rowHeight+1, 0
		}
		placements[code] = image.Rect(x, y, x+width, y+height)
		offset := font.FrameOffsets[glyph.Frame]
		metadata.Glyphs[string(code)] = fontGlyph{
			X: x, Y: y, Width: width, Height: height,
			Advance: glyph.Width, OffsetX: offset.X, OffsetY: offset.Y,
		}
		x += width + 1
		if height > rowHeight {
			rowHeight = height
		}
	}

	atlas := image.NewRGBA(image.Rect(0, 0, atlasWidth, y+rowHeight))
	for _, code := range codes {
		glyph := font.Glyphs[code]
		frame := font.Frames[glyph.Frame]
		if transformed := font.TextFrames[spec.color]; glyph.Frame < len(transformed) {
			frame = transformed[glyph.Frame]
		}
		draw.Draw(atlas, placements[code], frame, frame.Bounds().Min, draw.Over)
	}
	if err := writePNG(prefix+".png", atlas); err != nil {
		return "", "", err
	}
	data, err := json.Marshal(metadata)
	if err != nil {
		return "", "", err
	}
	if err := writePrivateFile(prefix+".json", func(writer io.Writer) error {
		_, err := writer.Write(data)
		return err
	}); err != nil {
		return "", "", err
	}
	cache.cached[key] = prefix
	return prefix + ".png", prefix + ".json", nil
}

func writePNG(path string, source image.Image) error {
	return writePrivateFile(path, func(writer io.Writer) error {
		return png.Encode(writer, source)
	})
}
