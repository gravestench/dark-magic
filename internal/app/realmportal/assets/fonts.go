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
	"slices"
	"strings"

	assetdecode "github.com/gravestench/dark-magic/internal/assets/decode"
)

const (
	font16Table     = "data/local/FONT/latin/font16.tbl"
	font16Sheet     = "data/local/FONT/latin/font16.dc6"
	exocetTable     = "data/local/FONT/latin/fontexocet10.tbl"
	exocetSheet     = "data/local/FONT/latin/fontexocet10.dc6"
	unitsPalette    = "data/global/Palette/units/pal.dat"
	actOneTransform = "data/global/Palette/Act1/Pal.pl2"
)

type fontSpec struct {
	table     string
	sheet     string
	palette   string
	transform string
	color     int
}

var fontAllowlist = map[string]fontSpec{
	"font16-red": {
		table: font16Table, sheet: font16Sheet, palette: unitsPalette, transform: actOneTransform, color: 1,
	},
	"font16-gold": {
		table: font16Table, sheet: font16Sheet, palette: unitsPalette, transform: actOneTransform, color: 4,
	},
	"font16-grey": {
		table: font16Table, sheet: font16Sheet, palette: unitsPalette, transform: actOneTransform, color: 5,
	},
	"font16-blue": {
		table: font16Table, sheet: font16Sheet, palette: unitsPalette, transform: actOneTransform, color: 3,
	},
	"exocet-grey": {
		table: exocetTable, sheet: exocetSheet, palette: unitsPalette, color: -1,
	},
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

// serveFont resolves only allowlisted font aliases and serves either an atlas or its matching metadata. Rendering
// errors intentionally remain indistinguishable from unknown IDs to avoid exposing archive details over HTTP.
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

// renderFont materializes a matched atlas and metadata pair under one lock. Holding the lock through both writes keeps
// concurrent requests from observing or memoizing a cache entry whose companion file has not been published yet.
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

	prefix, err := cache.fontCachePrefix(id, spec)
	if err != nil {
		return "", "", err
	}

	if fontCacheFilesExist(prefix) {
		cache.cached[key] = prefix
		return prefix + ".png", prefix + ".json", nil
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

	metadata, atlas := buildFontAtlas(id, spec, font)
	if err := writeFontFiles(prefix, metadata, atlas); err != nil {
		return "", "", err
	}

	cache.cached[key] = prefix

	return prefix + ".png", prefix + ".json", nil
}

// fontCachePrefix hashes the renderer and every authored input in a fixed order. Including absent transforms as an
// empty slot is unnecessary because the font ID and preceding paths already distinguish every allowlisted recipe.
func (cache *Cache) fontCachePrefix(id string, spec fontSpec) (string, error) {
	hash := sha256.New()
	_, _ = io.WriteString(hash, rendererVersion+id)

	for _, path := range []string{spec.table, spec.sheet, spec.palette, spec.transform} {
		if path == "" {
			continue
		}

		data, err := fs.ReadFile(cache.source, path)
		if err != nil {
			return "", err
		}

		_, _ = hash.Write(data)
	}

	name := id + "-" + hex.EncodeToString(hash.Sum(nil))[:20]

	return filepath.Join(cache.directory, name), nil
}

// fontCacheFilesExist treats an atlas as reusable only when both published files exist. A lone file can be left by an
// interrupted older render and must never be advertised as a complete font revision.
func fontCacheFilesExist(prefix string) bool {
	if _, err := os.Stat(prefix + ".png"); err != nil {
		return false
	}

	if _, err := os.Stat(prefix + ".json"); err != nil {
		return false
	}

	return true
}

// buildFontAtlas lays glyphs out in code-point order so identical inputs produce stable pixels and metadata even
// though the decoder stores glyphs in a map with nondeterministic iteration order.
func buildFontAtlas(id string, spec fontSpec, font *assetdecode.BitmapFont) (fontMetadata, image.Image) {
	codes := make([]rune, 0, len(font.Glyphs))
	for code := range font.Glyphs {
		codes = append(codes, code)
	}

	slices.Sort(codes)

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
		frame := fontGlyphFrame(font, spec.color, glyph.Frame)
		width, height := frame.Bounds().Dx(), frame.Bounds().Dy()
		// A one-pixel gutter prevents linear texture sampling from bleeding neighboring glyphs together in browsers.
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
		frame := fontGlyphFrame(font, spec.color, glyph.Frame)
		draw.Draw(atlas, placements[code], frame, frame.Bounds().Min, draw.Over)
	}

	return metadata, atlas
}

// fontGlyphFrame selects an authored color transform when it contains the requested frame and falls back to the base
// frame otherwise. The fallback preserves fonts whose transform banks are intentionally incomplete.
func fontGlyphFrame(font *assetdecode.BitmapFont, color, frameIndex int) image.Image {
	frame := font.Frames[frameIndex]

	if transformed := font.TextFrames[color]; frameIndex < len(transformed) {
		return transformed[frameIndex]
	}

	return frame
}

// writeFontFiles publishes atlas pixels before metadata. A client therefore cannot discover the new metadata before
// its atlas path exists, while the paired existence check rejects an interruption between the two publications.
func writeFontFiles(prefix string, metadata fontMetadata, atlas image.Image) error {
	if err := writePNG(prefix+".png", atlas); err != nil {
		return err
	}

	data, err := json.Marshal(metadata)
	if err != nil {
		return err
	}
	// writePrivateFile owns atomic publication and restrictive permissions for the encoded metadata.
	if err := writePrivateFile(prefix+".json", func(writer io.Writer) error {
		_, err := writer.Write(data)
		return err
	}); err != nil {
		return err
	}

	return nil
}

// writePNG delegates cache publication to writePrivateFile so images use the same best-effort private permissions and
// atomic path replacement as JSON metadata.
func writePNG(path string, source image.Image) error {
	// png.Encode writes through the temporary file supplied by writePrivateFile.
	return writePrivateFile(path, func(writer io.Writer) error {
		return png.Encode(writer, source)
	})
}
