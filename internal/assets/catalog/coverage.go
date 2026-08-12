package assetcatalog

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/fs"
	"path"
	"regexp"
	"slices"
	"strings"
)

var staticAssetPath = regexp.MustCompile(`(?i)data/[a-z0-9_./ -]+\.(?:bik|cof|dat|dc6|dcc|ds1|dt1|pl2|tbl|txt|wav)`)

// Coverage classifies presentation paths without requiring proprietary assets.
// Paths are normalized case-insensitively because the layered VFS is likewise
// case-insensitive for Diablo archive names.
type Coverage struct {
	Version                int      `json:"version"`
	PresentationPaths      []string `json:"presentation_paths"`
	VerifiedPresentation   []string `json:"verified_presentation"`
	UnverifiedPresentation []string `json:"unverified_presentation"`
	CodeOwnedPaths         []string `json:"code_owned_paths"`
	DynamicPrefixes        []string `json:"dynamic_prefixes"`
	CatalogOnlyPaths       []string `json:"catalog_only_paths"`
	CatalogFixtureGaps     []string `json:"catalog_fixture_gaps"`
	Fingerprint            string   `json:"fingerprint"`
}

// BuildCoverage joins static paths in the presentation manifest and Lua mod
// with the curated catalog and redistributable structural fixture.
func BuildCoverage(source fs.FS, manifest Manifest, fixture Fixture) (Coverage, error) {
	presentation, err := fs.ReadFile(source, "manifests/presentation.v1.json")
	if err != nil {
		return Coverage{}, err
	}
	var document any
	if err := json.Unmarshal(presentation, &document); err != nil {
		return Coverage{}, fmt.Errorf("asset catalog: decode presentation manifest: %w", err)
	}
	presentationPaths := make(map[string]bool)
	collectJSONPaths(document, presentationPaths)

	codePaths, dynamic := make(map[string]bool), make(map[string]bool)
	for _, root := range []string{"boot.lua", "lua"} {
		err := fs.WalkDir(source, root, func(name string, entry fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if entry.IsDir() || path.Ext(name) != ".lua" {
				return nil
			}
			data, err := fs.ReadFile(source, name)
			if err != nil {
				return err
			}
			for _, match := range staticAssetPath.FindAllString(string(data), -1) {
				codePaths[normalizeAssetPath(match)] = true
			}
			for _, match := range regexp.MustCompile(`(?i)data/[a-z0-9_./ -]+/`).FindAllString(string(data), -1) {
				if strings.Contains(string(data), match+`" ..`) {
					dynamic[normalizeAssetPath(match)] = true
				}
			}
			return nil
		})
		if err != nil {
			return Coverage{}, err
		}
	}

	catalogPaths := make(map[string]bool, len(manifest.Assets))
	catalogIDs := make(map[string]string, len(manifest.Assets))
	for _, asset := range manifest.Assets {
		normalized := normalizeAssetPath(asset.Path)
		catalogPaths[normalized] = true
		catalogIDs[asset.ID] = normalized
	}
	fixtureIDs := make(map[string]string, len(fixture.Assets))
	for _, asset := range fixture.Assets {
		fixtureIDs[asset.ID] = normalizeAssetPath(asset.Path)
	}

	result := Coverage{Version: 1}
	for name := range presentationPaths {
		result.PresentationPaths = append(result.PresentationPaths, name)
		if catalogPaths[name] {
			result.VerifiedPresentation = append(result.VerifiedPresentation, name)
		} else {
			result.UnverifiedPresentation = append(result.UnverifiedPresentation, name)
		}
	}
	for name := range codePaths {
		if !presentationPaths[name] {
			result.CodeOwnedPaths = append(result.CodeOwnedPaths, name)
		}
	}
	for name := range dynamic {
		result.DynamicPrefixes = append(result.DynamicPrefixes, name)
	}
	for name := range catalogPaths {
		if !presentationPaths[name] {
			result.CatalogOnlyPaths = append(result.CatalogOnlyPaths, name)
		}
	}
	for id, catalogPath := range catalogIDs {
		if fixturePath, ok := fixtureIDs[id]; !ok || fixturePath != catalogPath {
			result.CatalogFixtureGaps = append(result.CatalogFixtureGaps, id)
		}
	}
	for id := range fixtureIDs {
		if _, ok := catalogIDs[id]; !ok {
			result.CatalogFixtureGaps = append(result.CatalogFixtureGaps, id)
		}
	}
	result.sort()
	result.Fingerprint = result.fingerprint()
	return result, nil
}

func collectJSONPaths(value any, paths map[string]bool) {
	switch current := value.(type) {
	case []any:
		for _, item := range current {
			collectJSONPaths(item, paths)
		}
	case map[string]any:
		for _, item := range current {
			collectJSONPaths(item, paths)
		}
	case string:
		if staticAssetPath.MatchString(current) {
			paths[normalizeAssetPath(current)] = true
		}
	}
}

func normalizeAssetPath(value string) string {
	return strings.ToLower(strings.ReplaceAll(strings.TrimSpace(value), `\`, "/"))
}

func (c *Coverage) sort() {
	for _, values := range [][]string{c.PresentationPaths, c.VerifiedPresentation, c.UnverifiedPresentation, c.CodeOwnedPaths, c.DynamicPrefixes, c.CatalogOnlyPaths, c.CatalogFixtureGaps} {
		slices.Sort(values)
	}
}

func (c Coverage) fingerprint() string {
	hash := sha256.New()
	for _, group := range [][]string{c.PresentationPaths, c.VerifiedPresentation, c.UnverifiedPresentation, c.CodeOwnedPaths, c.DynamicPrefixes, c.CatalogOnlyPaths, c.CatalogFixtureGaps} {
		for _, value := range group {
			_, _ = hash.Write([]byte(value))
			_, _ = hash.Write([]byte{0})
		}
		_, _ = hash.Write([]byte{0xff})
	}
	return hex.EncodeToString(hash.Sum(nil))
}
