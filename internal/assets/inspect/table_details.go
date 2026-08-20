package assetinspect

import (
	"sort"
	"strings"

	"github.com/gravestench/dark-magic/internal/assets/decode"
	tbl "github.com/gravestench/tbl_text"
)

const fontTableSignature = "Woo!\x01"

// decodeTBLDetails distinguishes font metrics from localized string tables by
// signature because both formats use the same extension but incompatible decoders.
func decodeTBLDetails(data []byte) (any, error) {
	if len(data) >= len(fontTableSignature) && string(data[:len(fontTableSignature)]) == fontTableSignature {
		return decodeFontTableDetails(data)
	}

	return decodeStringTableDetails(data)
}

// decodeFontTableDetails reports the largest glyph bounds so diagnostics can
// size a preview without exposing the proprietary bitmap data.
func decodeFontTableDetails(data []byte) (any, error) {
	glyphs, err := assetdecode.FontTable(data)
	if err != nil {
		return nil, err
	}

	maxWidth, maxHeight := 0, 0
	for _, glyph := range glyphs {
		if glyph.Width > maxWidth {
			maxWidth = glyph.Width
		}

		if glyph.Height > maxHeight {
			maxHeight = glyph.Height
		}
	}

	return map[string]any{
		"format":     "font-table",
		"glyphs":     len(glyphs),
		"max_width":  maxWidth,
		"max_height": maxHeight,
	}, nil
}

// decodeStringTableDetails returns a stable, bounded sample so reports remain
// deterministic and compact regardless of source table size.
func decodeStringTableDetails(data []byte) (any, error) {
	table, err := tbl.Unmarshal(data)
	if err != nil {
		return nil, err
	}

	keys := make([]string, 0, len(table))
	for key := range table {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	if len(keys) > 5 {
		keys = keys[:5]
	}

	return map[string]any{
		"entries":     len(table),
		"sample_keys": keys,
	}, nil
}

// decodeTabularTextDetails normalizes line endings before counting rows while
// retaining the source header's full column count and only bounding its sample.
func decodeTabularTextDetails(data []byte) any {
	text := strings.TrimRight(string(data), "\x00\r\n")
	if text == "" {
		return map[string]any{"rows": 0, "columns": 0}
	}

	lines := strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
	headings := strings.Split(lines[0], "\t")

	columnCount := len(headings)
	if len(headings) > 10 {
		headings = headings[:10]
	}

	return map[string]any{
		"rows":           len(lines) - 1,
		"columns":        columnCount,
		"sample_columns": headings,
	}
}
