// Package assetcatalog verifies documented Diablo II asset knowledge against a
// content filesystem. It never modifies the source archives.
package assetcatalog

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/gravestench/dark-magic/internal/assets/decode"
	dc6 "github.com/gravestench/dc6/pkg"
)

// Source is the provenance needed by the catalog without coupling this public
// package to the internal content package.
type Source struct {
	Layer string `json:"layer"`
	Path  string `json:"path"`
}

// Hypothesis is one piece of community-derived knowledge to verify.
type Hypothesis struct {
	ID         string   `json:"id"`
	Screen     string   `json:"screen"`
	Path       string   `json:"path"`
	Palette    string   `json:"palette,omitempty"`
	Meaning    string   `json:"meaning"`
	References []string `json:"references,omitempty"`
}

// Manifest is a versioned set of hypotheses.
type Manifest struct {
	Version int          `json:"version"`
	Assets  []Hypothesis `json:"assets"`
}

// Frame describes stored DC6 placement metadata.
type Frame struct {
	Direction int `json:"direction"`
	Frame     int `json:"frame"`
	Width     int `json:"width"`
	Height    int `json:"height"`
	OffsetX   int `json:"offset_x"`
	OffsetY   int `json:"offset_y"`
}

// Result records whether a hypothesis resolved and what was observed.
type Result struct {
	Hypothesis
	Found      bool     `json:"found"`
	Source     *Source  `json:"source,omitempty"`
	Bytes      int      `json:"bytes,omitempty"`
	SHA256     string   `json:"sha256,omitempty"`
	Type       string   `json:"type,omitempty"`
	Directions int      `json:"directions,omitempty"`
	Frames     []Frame  `json:"frames,omitempty"`
	PaletteOK  bool     `json:"palette_found,omitempty"`
	Sheet      string   `json:"contact_sheet,omitempty"`
	Warnings   []string `json:"warnings,omitempty"`
	Error      string   `json:"error,omitempty"`
}

// Report is deterministic output suitable for review or shim generation.
type Report struct {
	ManifestVersion int      `json:"manifest_version"`
	Results         []Result `json:"results"`
	Aliases         []Alias  `json:"aliases,omitempty"`
}

// Alias groups different documented paths whose resolved bytes are identical.
type Alias struct {
	SHA256 string   `json:"sha256"`
	Paths  []string `json:"paths"`
}

// Options controls optional diagnostic artifacts.
type Options struct {
	SheetWriter func(name string, data []byte) (string, error)
	Resolve     func(name string) (Source, error)
}

// Verify resolves and inspects every manifest entry. Missing or malformed
// assets are represented in the report and do not abort the remaining scan.
func Verify(source fs.FS, manifest Manifest, options Options) Report {
	report := Report{ManifestVersion: manifest.Version, Results: make([]Result, 0, len(manifest.Assets))}
	for _, hypothesis := range manifest.Assets {
		result := inspect(source, hypothesis, options)
		report.Results = append(report.Results, result)
	}
	report.Aliases = findAliases(report.Results)
	return report
}

func findAliases(results []Result) []Alias {
	byDigest := make(map[string][]string)
	order := make([]string, 0)
	for _, result := range results {
		if !result.Found || result.SHA256 == "" {
			continue
		}
		if _, exists := byDigest[result.SHA256]; !exists {
			order = append(order, result.SHA256)
		}
		byDigest[result.SHA256] = append(byDigest[result.SHA256], result.Path)
	}
	aliases := make([]Alias, 0)
	for _, digest := range order {
		paths := byDigest[digest]
		if len(paths) > 1 {
			aliases = append(aliases, Alias{SHA256: digest, Paths: paths})
		}
	}
	return aliases
}

func inspect(source fs.FS, hypothesis Hypothesis, options Options) Result {
	result := Result{Hypothesis: hypothesis}
	if options.Resolve != nil {
		resolved, err := options.Resolve(hypothesis.Path)
		if err == nil {
			result.Source = &resolved
		}
	}
	data, err := fs.ReadFile(source, hypothesis.Path)
	if err != nil {
		result.Error = err.Error()
		return result
	}
	result.Found = true
	result.Bytes = len(data)
	digest := sha256.Sum256(data)
	result.SHA256 = hex.EncodeToString(digest[:])
	result.Type = strings.TrimPrefix(strings.ToLower(filepath.Ext(hypothesis.Path)), ".")
	if result.Type != "dc6" {
		return result
	}

	asset, err := dc6.FromBytes(data)
	if err != nil {
		result.Error = err.Error()
		return result
	}
	if hypothesis.Palette != "" {
		palette, paletteErr := assetdecode.Palette(source, hypothesis.Palette)
		if paletteErr != nil {
			result.Warnings = append(result.Warnings, paletteErr.Error())
		} else {
			asset.SetPalette(palette)
			result.PaletteOK = true
		}
	}
	result.Directions = len(asset.Directions)
	for directionIndex, direction := range asset.Directions {
		for frameIndex, frame := range direction.Frames {
			result.Frames = append(result.Frames, Frame{
				Direction: directionIndex,
				Frame:     frameIndex,
				Width:     int(frame.Width),
				Height:    int(frame.Height),
				OffsetX:   int(frame.OffsetX),
				OffsetY:   int(frame.OffsetY),
			})
		}
	}
	if options.SheetWriter != nil {
		sheet, sheetErr := DC6ContactSheet(asset)
		if sheetErr != nil {
			result.Warnings = append(result.Warnings, sheetErr.Error())
		} else {
			name := safeName(hypothesis.ID) + ".png"
			written, writeErr := options.SheetWriter(name, sheet)
			if writeErr != nil {
				result.Warnings = append(result.Warnings, writeErr.Error())
			} else {
				result.Sheet = written
			}
		}
	}
	return result
}

func safeName(value string) string {
	value = strings.ToLower(value)
	value = strings.Map(func(r rune) rune {
		if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' || r == '-' || r == '_' {
			return r
		}
		return '-'
	}, value)
	return strings.Trim(value, "-")
}

// Validate checks manifest structure before touching any game data.
func (m Manifest) Validate() error {
	if m.Version < 1 {
		return errors.New("asset catalog: manifest version must be positive")
	}
	seen := make(map[string]struct{}, len(m.Assets))
	for index, asset := range m.Assets {
		if asset.ID == "" || asset.Screen == "" || asset.Path == "" || asset.Meaning == "" {
			return fmt.Errorf("asset catalog: asset %d requires id, screen, path, and meaning", index)
		}
		if _, exists := seen[asset.ID]; exists {
			return fmt.Errorf("asset catalog: duplicate id %q", asset.ID)
		}
		seen[asset.ID] = struct{}{}
	}
	return nil
}
