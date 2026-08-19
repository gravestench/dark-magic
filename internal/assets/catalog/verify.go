package assetcatalog

import (
	"crypto/sha256"
	"encoding/hex"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/gravestench/dark-magic/internal/assets/decode"
	dc6 "github.com/gravestench/dc6/pkg"
)

// Verify resolves and inspects every manifest entry in manifest order. Per-asset failures stay in the report so one
// missing or malformed archive entry cannot hide observations about the remaining assets.
func Verify(source fs.FS, manifest Manifest, options Options) Report {
	report := Report{
		ManifestVersion: manifest.Version,
		Results:         make([]Result, 0, len(manifest.Assets)),
	}

	for _, hypothesis := range manifest.Assets {
		report.Results = append(report.Results, inspect(source, hypothesis, options))
	}

	report.Aliases = findAliases(report.Results)

	return report
}

// findAliases groups successfully read assets by digest while retaining the manifest's first-seen digest order.
// Stable ordering keeps reports reviewable and prevents map iteration from changing serialized output.
func findAliases(results []Result) []Alias {
	byDigest := make(map[string][]string)
	digestOrder := make([]string, 0)

	for _, result := range results {
		if !result.Found || result.SHA256 == "" {
			continue
		}

		if _, exists := byDigest[result.SHA256]; !exists {
			digestOrder = append(digestOrder, result.SHA256)
		}

		byDigest[result.SHA256] = append(byDigest[result.SHA256], result.Path)
	}

	aliases := make([]Alias, 0)

	for _, digest := range digestOrder {
		paths := byDigest[digest]
		if len(paths) <= 1 {
			continue
		}

		aliases = append(aliases, Alias{SHA256: digest, Paths: paths})
	}

	return aliases
}

// inspect records all observations for one hypothesis in dependency order. Source resolution is advisory, whereas a
// filesystem read or DC6 decode failure determines whether later inspection phases can safely run.
func inspect(source fs.FS, hypothesis Hypothesis, options Options) Result {
	result := Result{Hypothesis: hypothesis}
	recordResolvedSource(&result, hypothesis.Path, options.Resolve)

	data, err := fs.ReadFile(source, hypothesis.Path)
	if err != nil {
		result.Error = err.Error()

		return result
	}

	recordContentIdentity(&result, data)

	if result.Type != "dc6" {
		return result
	}

	asset, err := dc6.FromBytes(data)
	if err != nil {
		result.Error = err.Error()

		return result
	}

	applyPalette(&result, source, asset, hypothesis.Palette)
	result.Directions = len(asset.Directions)
	result.Frames = describeFrames(asset)
	writeContactSheet(&result, asset, hypothesis.ID, options.SheetWriter)

	return result
}

// recordResolvedSource preserves provenance when resolution succeeds but deliberately ignores resolver failures.
// The content filesystem remains authoritative for availability even when layered metadata is unavailable.
func recordResolvedSource(result *Result, assetPath string, resolve func(string) (Source, error)) {
	if resolve == nil {
		return
	}

	resolved, err := resolve(assetPath)
	if err != nil {
		return
	}

	result.Source = &resolved
}

// recordContentIdentity captures byte-level facts before format decoding. Callers can therefore distinguish corrupt
// DC6 data from a missing archive entry and compare its exact source bytes.
func recordContentIdentity(result *Result, data []byte) {
	result.Found = true
	result.Bytes = len(data)

	digest := sha256.Sum256(data)
	result.SHA256 = hex.EncodeToString(digest[:])
	result.Type = strings.TrimPrefix(strings.ToLower(filepath.Ext(result.Path)), ".")
}

// applyPalette enriches DC6 rendering when palette data is available. Palette failures remain warnings because frame
// structure and source identity are still valid catalog observations.
func applyPalette(result *Result, source fs.FS, asset *dc6.DC6, palettePath string) {
	if palettePath == "" {
		return
	}

	palette, err := assetdecode.Palette(source, palettePath)
	if err != nil {
		result.Warnings = append(result.Warnings, err.Error())

		return
	}

	asset.SetPalette(palette)

	result.PaletteOK = true
}

// describeFrames flattens decoder-owned directions into stable direction/frame order. Copying scalar metadata keeps
// report ownership independent from the decoder object used later for optional rendering.
func describeFrames(asset *dc6.DC6) []Frame {
	var frames []Frame

	for directionIndex, direction := range asset.Directions {
		for frameIndex, frame := range direction.Frames {
			frames = append(frames, Frame{
				Direction: directionIndex,
				Frame:     frameIndex,
				Width:     int(frame.Width),
				Height:    int(frame.Height),
				OffsetX:   int(frame.OffsetX),
				OffsetY:   int(frame.OffsetY),
			})
		}
	}

	return frames
}

// writeContactSheet appends diagnostic failures as warnings so optional output cannot invalidate successful decoding.
// The writer-provided path is recorded only after the complete PNG has been accepted by the writer.
func writeContactSheet(
	result *Result,
	asset *dc6.DC6,
	hypothesisID string,
	writer func(name string, data []byte) (string, error),
) {
	if writer == nil {
		return
	}

	sheet, err := DC6ContactSheet(asset)
	if err != nil {
		result.Warnings = append(result.Warnings, err.Error())

		return
	}

	written, err := writer(safeName(hypothesisID)+".png", sheet)
	if err != nil {
		result.Warnings = append(result.Warnings, err.Error())

		return
	}

	result.Sheet = written
}

// safeName converts hypothesis IDs into portable contact-sheet basenames. Trimming replacement characters avoids
// leading or trailing separators while retaining existing ASCII letters, digits, dashes, and underscores.
func safeName(value string) string {
	value = strings.ToLower(value)
	// Map rune-by-rune so unsupported Unicode cannot leak platform-sensitive bytes into artifact filenames.
	value = strings.Map(func(r rune) rune {
		if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' || r == '-' || r == '_' {
			return r
		}

		return '-'
	}, value)

	return strings.Trim(value, "-")
}
