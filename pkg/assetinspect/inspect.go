// Package assetinspect provides a headless way to verify and describe assets
// from either a directory or an archive-backed fs.FS.
package assetinspect

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"

	"github.com/gravestench/dark-magic/pkg/assetdecode"
	dc6 "github.com/gravestench/dc6/pkg"
	"github.com/gravestench/dcc"
	"github.com/gravestench/ds1"
	"github.com/gravestench/dt1"
	tbl "github.com/gravestench/tbl_text"
)

type Report struct {
	Path    string `json:"path"`
	Type    string `json:"type"`
	Bytes   int64  `json:"bytes"`
	SHA256  string `json:"sha256"`
	Details any    `json:"details,omitempty"`
}

func Inspect(source fs.FS, path string) (Report, error) {
	file, err := source.Open(path)
	if err != nil {
		return Report{}, fmt.Errorf("opening asset %q: %w", path, err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return Report{}, fmt.Errorf("reading asset %q: %w", path, err)
	}
	return InspectData(path, data)
}

// InspectData describes already-loaded asset bytes. It is used by services
// that expose assets through their own composite loader.
func InspectData(path string, data []byte) (Report, error) {
	digest := sha256.Sum256(data)

	extension := strings.TrimPrefix(strings.ToLower(filepath.Ext(path)), ".")
	if extension == "" {
		extension = "unknown"
	}

	report := Report{
		Path:   path,
		Type:   extension,
		Bytes:  int64(len(data)),
		SHA256: hex.EncodeToString(digest[:]),
	}

	details, err := decodeDetails(extension, data)
	if err != nil {
		return Report{}, fmt.Errorf("decoding %s asset %q: %w", extension, path, err)
	}
	report.Details = details
	return report, nil
}

func decodeDetails(extension string, data []byte) (any, error) {
	switch extension {
	case "bik":
		return assetdecode.BIK(data)
	case "dc6":
		asset, err := dc6.FromBytes(data)
		if err != nil {
			return nil, err
		}
		frames := 0
		for _, direction := range asset.Directions {
			frames += len(direction.Frames)
		}
		return map[string]any{"version": asset.Version, "directions": len(asset.Directions), "frames": frames}, nil
	case "dcc":
		asset, err := dcc.FromBytes(data)
		if err != nil {
			return nil, err
		}
		return map[string]any{"version": asset.Version, "directions": len(asset.Directions()), "coded_bytes": asset.TotalSizeCoded}, nil
	case "ds1":
		asset, err := ds1.FromBytes(data)
		if err != nil {
			return nil, err
		}
		return map[string]any{"version": asset.Version, "width": asset.Width, "height": asset.Height, "act": asset.Act, "objects": len(asset.Objects)}, nil
	case "dt1":
		asset, err := dt1.FromBytes(data)
		if err != nil {
			return nil, err
		}
		types := make(map[int32]bool)
		styles := make(map[int32]bool)
		for _, tile := range asset.Tiles {
			types[tile.Type] = true
			styles[tile.Style] = true
		}
		return map[string]any{"tiles": len(asset.Tiles), "types": sortedInt32Keys(types), "styles": sortedInt32Keys(styles)}, nil
	case "tbl":
		if len(data) >= 5 && string(data[:5]) == "Woo!\x01" {
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
			return map[string]any{"format": "font-table", "glyphs": len(glyphs), "max_width": maxWidth, "max_height": maxHeight}, nil
		}
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
		return map[string]any{"entries": len(table), "sample_keys": keys}, nil
	case "txt", "tsv":
		text := strings.TrimRight(string(data), "\x00\r\n")
		if text == "" {
			return map[string]any{"rows": 0, "columns": 0}, nil
		}
		lines := strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
		headings := strings.Split(lines[0], "\t")
		if len(headings) > 10 {
			headings = headings[:10]
		}
		return map[string]any{"rows": len(lines) - 1, "columns": len(strings.Split(lines[0], "\t")), "sample_columns": headings}, nil
	default:
		return nil, nil
	}
}

func sortedInt32Keys(values map[int32]bool) []int32 {
	result := make([]int32, 0, len(values))
	for value := range values {
		result = append(result, value)
	}
	sort.Slice(result, func(i, j int) bool { return result[i] < result[j] })
	return result
}
